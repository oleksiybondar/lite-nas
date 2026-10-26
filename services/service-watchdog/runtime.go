package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	sharedcontracts "lite-nas/shared/contracts"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	servicecontract "lite-nas/shared/contracts/servicemonitor"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanagerauth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/systemd"
	"lite-nas/shared/workers"
)

const (
	packagedWatchdogConfig = "/etc/lite-nas/service-watchdog.conf"
	maxServiceEventBytes   = 8 * 1024
)

// recoveryIntent is sent only after an authoritative probe and a configured
// unit allowlist check. It cannot be constructed from NATS payload alone.
type recoveryIntent struct{ unit string }

// recoveryResult reports the command and verification outcome to the owner of
// Watchdog's per-unit backoff and incident state.
type recoveryResult struct {
	unit  string
	state systemd.UnitState
	err   error
}

// securityEffect is a typed logging-manager lifecycle operation.
type securityEffect struct {
	unit      protectedUnit
	kind      string
	operation string
	message   string
	attempts  int
}

// run starts validated NATS, reconciliation, recovery and reporter workers.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedWatchdogConfig)
	if err != nil {
		return err
	}
	cfg, err := loadWatchdogConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceWatchdog, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	server, err := messaging.NewServer(cfg.Messaging, infra.Logger, messaging.NewJSONCodec())
	if err != nil {
		return err
	}
	defer server.Close()
	notifications := make(chan servicecontract.StateEvent, 32)
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := server.Subscribe(servicecontract.StateSubject, func(handlerCtx context.Context, envelope messaging.Envelope) error {
		event, err := decodeServiceEvent(envelope, validate)
		if err != nil {
			infra.Logger.Warn("invalid service notification", "error", err)
			return err
		}
		if _, protected := cfg.Units[event.Unit]; !protected {
			return nil
		}
		select {
		case notifications <- event:
			return nil
		case <-handlerCtx.Done():
			return handlerCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return err
	}
	timer, ticks, err := workers.NewPollingTimerWorker(cfg.PollInterval, 1)
	if err != nil {
		return err
	}
	recoveryJobs := make(chan recoveryIntent, 16)
	recoveryResults := make(chan recoveryResult, 16)
	effects := make(chan securityEffect, 32)
	coordinatorDone := make(chan struct{})
	reporterDone := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) { recoverServices(workerCtx, cfg, recoveryJobs, recoveryResults) }).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		coordinateWatchdog(workerCtx, cfg, ticks, notifications, recoveryJobs, recoveryResults, effects, infra.Logger)
		close(coordinatorDone)
	}).Start(ctx)
	operatorClient := loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		reportSecurityEffects(workerCtx, effects, operatorClient, infra.Logger)
		close(reporterDone)
	}).Start(ctx)
	timer.Start(ctx)
	<-ctx.Done()
	if err := server.Drain(); err != nil {
		infra.Logger.Error("watchdog NATS drain failed", "error", err)
	}
	close(notifications)
	<-coordinatorDone
	<-reporterDone
	return ctx.Err()
}

// decodeServiceEvent rejects oversized, unknown, or invalid transport data.
func decodeServiceEvent(envelope messaging.Envelope, validate *validator.Validate) (servicecontract.StateEvent, error) {
	var event servicecontract.StateEvent
	if len(envelope.Payload) == 0 || len(envelope.Payload) > maxServiceEventBytes {
		return event, errors.New("service event payload size invalid")
	}
	decoder := json.NewDecoder(strings.NewReader(string(envelope.Payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return event, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return event, errors.New("service event contains trailing JSON")
	}
	if err := validate.Struct(event); err != nil {
		return event, err
	}
	if !systemd.ValidUnitName(event.Unit) {
		return event, errors.New("invalid service unit name")
	}
	return event, nil
}

// coordinateWatchdog owns all mutable recovery and incident state. NATS
// messages only request an early authoritative check; ticks reconcile all
// protected units even if notifications were missed.
func coordinateWatchdog(ctx context.Context, cfg watchdogConfig, ticks <-chan struct{}, notifications <-chan servicecontract.StateEvent,
	recoveryJobs chan<- recoveryIntent, recoveryResults <-chan recoveryResult, effects chan<- securityEffect, log sharedlogger.Logger,
) {
	defer close(effects)
	progress := make(map[string]unitProgress, len(cfg.Units))
	for {
		select {
		case <-ctx.Done():
			close(recoveryJobs)
			for result := range recoveryResults {
				if result.err != nil {
					log.Warn("recovery ended during shutdown", "unit", result.unit, "error", result.err)
				}
			}
			return
		case <-ticks:
			for _, unit := range cfg.Units {
				checkProtectedUnit(ctx, cfg, unit, progress, recoveryJobs, effects, log)
			}
		case event, ok := <-notifications:
			if !ok {
				notifications = nil
				continue
			}
			if unit, protected := cfg.Units[event.Unit]; protected {
				checkProtectedUnit(ctx, cfg, unit, progress, recoveryJobs, effects, log)
			}
		case result, ok := <-recoveryResults:
			if !ok {
				recoveryResults = nil
				continue
			}
			unit := cfg.Units[result.unit]
			state := progress[result.unit]
			state.Recovering = false
			if result.err != nil {
				operation := "occurrence"
				if !state.RecoveryIncident {
					operation = "create"
					state.RecoveryIncident = true
				}
				effects <- securityEffect{
					unit: unit, kind: "recovery", operation: operation,
					message: fmt.Sprintf("Watchdog recovery failed for %s after %d attempts", result.unit, state.Attempts), attempts: state.Attempts,
				}
			}
			progress[result.unit] = state
		}
	}
}

// checkProtectedUnit verifies enablement and availability independently of
// the notification body before considering a narrowly authorized start.
func checkProtectedUnit(ctx context.Context, cfg watchdogConfig, unit protectedUnit, progress map[string]unitProgress,
	recoveryJobs chan<- recoveryIntent, effects chan<- securityEffect, log sharedlogger.Logger,
) {
	observed, err := systemd.ProbeUnit(ctx, unit.Name)
	if err != nil {
		log.Error("watchdog status inspection failed", "unit", unit.Name, "error", err)
		return
	}
	state := progress[unit.Name]
	enablement, err := systemd.EnablementState(ctx, unit.Name)
	if err != nil {
		log.Error("watchdog enablement inspection failed", "unit", unit.Name, "error", err)
	} else if enablement != unit.ExpectedEnablement && !state.EnablementIncident {
		effects <- securityEffect{
			unit: unit, kind: "enablement", operation: "create",
			message: fmt.Sprintf("Unexpected enablement of %s: observed %s, expected %s", unit.Name, enablement, unit.ExpectedEnablement),
		}
		state.EnablementIncident = true
	} else if enablement == unit.ExpectedEnablement && state.EnablementIncident {
		effects <- securityEffect{unit: unit, kind: "enablement", operation: "normalize", message: "Expected service enablement restored: " + unit.Name}
		state.EnablementIncident = false
	}
	if observed.ActiveState == "active" {
		if state.RecoveryIncident {
			effects <- securityEffect{unit: unit, kind: "recovery", operation: "normalize", message: "Protected service recovered: " + unit.Name}
			state.RecoveryIncident = false
		}
		state.FailureSince = time.Time{}
		state.NextAttempt = time.Time{}
		state.Attempts = 0
		progress[unit.Name] = state
		return
	}
	if state.FailureSince.IsZero() {
		state.FailureSince = time.Now()
	}
	if shouldAttempt(time.Now(), cfg, unit, observed, state) {
		state.Recovering = true
		state.Attempts++
		state.NextAttempt = time.Now().Add(nextBackoff(cfg.Backoff, state.Attempts))
		select {
		case recoveryJobs <- recoveryIntent{unit: unit.Name}:
		case <-ctx.Done():
			state.Recovering = false
			state.Attempts--
		}
	}
	progress[unit.Name] = state
}

// recoverServices runs one fixed sudo/systemctl start command at a time. The
// coordinator has already checked the unit allowlist, and the worker checks it
// again before invoking the exact sudoers operation.
func recoverServices(ctx context.Context, cfg watchdogConfig, jobs <-chan recoveryIntent, results chan<- recoveryResult) {
	defer close(results)
	for job := range jobs {
		unit, allowed := cfg.Units[job.unit]
		if !allowed || !unit.Recover || !systemd.ValidUnitName(job.unit) {
			results <- recoveryResult{unit: job.unit, err: errors.New("recovery target not allowlisted")}
			continue
		}
		commandCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		command := exec.CommandContext(commandCtx, "/usr/bin/sudo", "-n", systemd.SystemctlBinary, "start", job.unit)
		command.Env = []string{"PATH=/usr/bin:/bin", "LANG=C"}
		err := command.Run()
		cancel()
		if err != nil {
			results <- recoveryResult{unit: job.unit, err: fmt.Errorf("systemctl start: %w", err)}
			continue
		}
		observed, err := systemd.ProbeUnit(ctx, job.unit)
		if err == nil && observed.ActiveState != "active" && observed.ActiveState != "activating" {
			err = fmt.Errorf("unit remains %s/%s after start", observed.ActiveState, observed.SubState)
		}
		results <- recoveryResult{unit: job.unit, state: observed, err: err}
	}
}

// reportSecurityEffects retries every security incident and lifecycle update
// while the service is running, then logs any effect still undelivered at stop.
func reportSecurityEffects(ctx context.Context, effects <-chan securityEffect, client messaging.Client, log sharedlogger.Logger) {
	for effect := range effects {
		for {
			attemptCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := publishSecurityEffect(attemptCtx, effect, client)
			if cancel != nil {
				cancel()
			}
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				log.Error("watchdog effect undelivered at shutdown", "unit", effect.unit.Name, "error", err)
				break
			}
			log.Error("watchdog effect publication failed; retrying", "unit", effect.unit.Name, "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// publishSecurityEffect uses the existing logging-manager create, occurrence
// and normalize contracts. Critical create alerts follow the email notifier's
// existing subscription path.
func publishSecurityEffect(ctx context.Context, effect securityEffect, client messaging.Client) error {
	digest := sha256.Sum256([]byte(effect.kind + ":" + effect.unit.Name))
	id := "W" + hex.EncodeToString(digest[:8])
	if len(effect.message) > 256 {
		effect.message = effect.message[:256]
	}
	switch effect.operation {
	case "create":
		severity := enum.SeverityWarning
		if effect.unit.Critical {
			severity = enum.SeverityCritical
		}
		return client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
			EventID: id, Category: "Service protection", Severity: severity,
			CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceWatchdog,
			Message: effect.message,
		})
	case "occurrence":
		count := float64(effect.attempts)
		return client.Publish(ctx, securitycontract.AlertOccurrenceSubject, loggingcontract.AlertOccurrencePayload{
			EventID: id, Timestamp: time.Now().UTC().Format(time.RFC3339),
			ValueType: enum.ValueTypeInt, ValueNum: &count,
		})
	case "normalize":
		response := loggingcontract.OKResponse{}
		err := client.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: id, Status: enum.StatusNormal, Message: &effect.message}, &response)
		if err != nil {
			return err
		}
		if !response.OK {
			return errors.New("security logging manager rejected normalization")
		}
		return nil
	default:
		return errors.New("unknown watchdog effect")
	}
}
