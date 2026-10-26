package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	sharedcontracts "lite-nas/shared/contracts"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	servicecontract "lite-nas/shared/contracts/servicemonitor"
	systemcontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanagerauth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/workers"
)

const packagedMonitorConfig = "/etc/lite-nas/system-service-monitor.conf"

// unitSnapshot contains successful observations only. A failed systemctl
// query never masquerades as an inactive unit.
type unitSnapshot struct {
	states []unitState
	at     time.Time
}

// stateIntent moves one factual transition through the outbound worker.
type stateIntent struct {
	event servicecontract.StateEvent
}

// run loads the configured allowlist and runs the polling worker pipeline.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedMonitorConfig)
	if err != nil {
		return err
	}
	cfg, err := loadMonitorConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceSystemServiceMonitor, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	return runServicePipeline(ctx, cfg, infra.Client,
		loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken),
		infra.Logger, probeUnit)
}

// runServicePipeline owns the bounded channels and waits for the publisher to
// account for queued effects after polling stops.
func runServicePipeline(ctx context.Context, cfg monitorConfig, eventClient, alertClient messaging.Client,
	log sharedlogger.Logger, probe func(context.Context, string) (unitState, error),
) error {
	timer, ticks, err := workers.NewPollingTimerWorker(cfg.PollInterval, 1)
	if err != nil {
		return err
	}
	snapshots := make(chan unitSnapshot, 2)
	intents := make(chan stateIntent, 32)
	done := make(chan struct{})
	poll := func(ctx context.Context) (unitSnapshot, error) {
		snapshot := unitSnapshot{at: time.Now().UTC(), states: make([]unitState, 0, len(cfg.Units))}
		for _, unit := range cfg.Units {
			if ctx.Err() != nil {
				break
			}
			state, err := probe(ctx, unit)
			if err != nil {
				log.Error("systemd unit inspection failed", "unit", unit, "error", err)
				continue
			}
			snapshot.states = append(snapshot.states, state)
		}
		return snapshot, nil
	}
	poller := workers.NewPollingWorker(ticks, snapshots, make(chan error, 1), poll)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(snapshots)
		for poller.WaitNextPoll(workerCtx) {
			poller.PollAndSend(workerCtx)
		}
	}).Start(ctx)
	workers.NewBaseWorker(func(_ context.Context) { evaluateUnits(snapshots, intents, log) }).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		publishStates(workerCtx, intents, eventClient, alertClient, log)
		close(done)
	}).Start(ctx)
	timer.Start(ctx)
	<-done
	return ctx.Err()
}

// evaluateUnits emits only baseline unavailable conditions and later factual
// condition changes. A transient missing observation does not clear state.
func evaluateUnits(snapshots <-chan unitSnapshot, out chan<- stateIntent, log sharedlogger.Logger) {
	defer close(out)
	previous := make(map[string]unitState)
	for snapshot := range snapshots {
		for _, current := range snapshot.states {
			prior, exists := previous[current.Unit]
			condition := classifyUnit(nil, current)
			if exists {
				priorCondition := classifyUnit(nil, prior)
				if condition == priorCondition {
					previous[current.Unit] = current
					continue
				}
				condition = classifyUnit(&prior, current)
			} else if condition == servicecontract.Active {
				previous[current.Unit] = current
				continue
			}
			id, err := newStateEventID()
			if err != nil {
				log.Error("service event ID generation failed", "unit", current.Unit, "error", err)
				continue
			}
			out <- stateIntent{event: servicecontract.StateEvent{
				Version: servicecontract.CurrentVersion, EventID: id, Unit: current.Unit,
				Condition: condition, ActiveState: current.ActiveState, SubState: current.SubState,
				Result: current.Result, RestartCount: current.RestartCount,
				ObservedAt: snapshot.at.Format(time.RFC3339),
			}}
			previous[current.Unit] = current
		}
	}
}

// newStateEventID fits the existing logging-manager event ID limit.
func newStateEventID() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "S" + hex.EncodeToString(bytes[:]), nil
}

// publishStates retries intents with backpressure. Alert IDs stay correlated
// per unit for this process lifetime and normalize only after RPC success.
func publishStates(ctx context.Context, intents <-chan stateIntent, eventClient, alertClient messaging.Client, log sharedlogger.Logger) {
	alerts := make(map[string]string)
	for intent := range intents {
		stateEventSent := false
		for {
			attemptCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := publishStateIntent(attemptCtx, intent.event, &stateEventSent, alerts, eventClient, alertClient)
			if cancel != nil {
				cancel()
			}
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				log.Error("service state intent undelivered at shutdown", "event_id", intent.event.EventID, "error", err)
				break
			}
			log.Error("service state intent publication failed; retrying", "event_id", intent.event.EventID, "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// publishStateIntent sends a state event once per local retry sequence and
// follows the system alert lifecycle for down and restored conditions.
func publishStateIntent(ctx context.Context, event servicecontract.StateEvent, eventSent *bool, alerts map[string]string,
	eventClient, alertClient messaging.Client,
) error {
	if !*eventSent {
		if err := eventClient.Publish(ctx, servicecontract.StateSubject, event); err != nil {
			return err
		}
		*eventSent = true
	}
	if event.Condition == servicecontract.Restored || event.Condition == servicecontract.Active {
		id := alerts[event.Unit]
		if id == "" {
			return nil
		}
		message := "Service restored: " + event.Unit
		response := loggingcontract.OKResponse{}
		err := alertClient.Request(ctx, systemcontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: id, Status: enum.StatusNormal, Message: &message}, &response)
		if err != nil {
			return err
		}
		if !response.OK {
			return fmt.Errorf("system logging manager rejected restoration")
		}
		delete(alerts, event.Unit)
		return nil
	}
	if event.Condition != servicecontract.Failed && event.Condition != servicecontract.PersistentFailure && event.Condition != servicecontract.Inactive {
		return nil
	}
	if alerts[event.Unit] != "" {
		return nil
	}
	message := fmt.Sprintf("Service unavailable: %s (%s/%s, result %s)", event.Unit, event.ActiveState, event.SubState, event.Result)
	if len(message) > 256 {
		message = message[:256]
	}
	alert := loggingcontract.AlertPayload{
		EventID: event.EventID, Category: "Service availability", Severity: enum.SeverityError,
		CreatedAt: event.ObservedAt, Source: sharedcontracts.ServiceSystemServiceMonitor, Message: message,
	}
	if err := alertClient.Publish(ctx, systemcontract.AlertSubject, alert); err != nil {
		return err
	}
	alerts[event.Unit] = event.EventID
	return nil
}
