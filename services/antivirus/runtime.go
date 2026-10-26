package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	sharedcontracts "lite-nas/shared/contracts"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	mountcontract "lite-nas/shared/contracts/mountmonitor"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	systemcontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanagerauth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/mountinfo"
	"lite-nas/shared/workers"
)

const (
	packagedAntivirusConfig = "/etc/lite-nas/antivirus.conf"
	maxMountEventBytes      = 16 * 1024
	maxScanQueue            = 64
)

// scanJob carries an independently cancelable scan for one verified mount.
type scanJob struct {
	mount mountinfo.Record
	ctx   context.Context
}

// run composes the validated NATS bridge, coordinator, scanner and outbound
// reporting worker within one service process.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedAntivirusConfig)
	if err != nil {
		return err
	}
	cfg, err := loadAntivirusConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceAntivirus, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	server, err := messaging.NewServer(cfg.Messaging, infra.Logger, messaging.NewJSONCodec())
	if err != nil {
		return err
	}
	defer server.Close()
	incoming := make(chan mountcontract.LifecycleEvent, maxScanQueue)
	jobs := make(chan scanJob, maxScanQueue)
	results := make(chan scanResult, maxScanQueue)
	effects := make(chan scanResult, maxScanQueue)
	validator := validator.New(validator.WithRequiredStructEnabled())
	if err := server.Subscribe(mountcontract.LifecycleSubject, func(handlerCtx context.Context, envelope messaging.Envelope) error {
		event, err := decodeMountEvent(envelope, validator)
		if err != nil {
			infra.Logger.Warn("invalid mount notification", "error", err)
			return err
		}
		select {
		case incoming <- event:
			return nil
		case <-handlerCtx.Done():
			return handlerCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return err
	}
	operatorClient := loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken)
	coordinatorDone := make(chan struct{})
	reporterDone := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) { runScanWorker(workerCtx, jobs, results, cfg.ScanTimeout) }).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		coordinateScans(workerCtx, cfg, incoming, jobs, results, effects, infra.Logger)
		close(coordinatorDone)
	}).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		reportScans(workerCtx, effects, operatorClient, infra.Logger)
		close(reporterDone)
	}).Start(ctx)
	<-ctx.Done()
	if err := server.Drain(); err != nil {
		infra.Logger.Error("NATS subscription drain failed", "error", err)
	}
	close(incoming) // The drained NATS server can no longer send.
	<-coordinatorDone
	<-reporterDone
	return ctx.Err()
}

// decodeMountEvent enforces the versioned transport schema and bounded payload
// before the coordinator performs any eligibility or filesystem work.
func decodeMountEvent(envelope messaging.Envelope, validate *validator.Validate) (mountcontract.LifecycleEvent, error) {
	var event mountcontract.LifecycleEvent
	if len(envelope.Payload) == 0 || len(envelope.Payload) > maxMountEventBytes {
		return event, errors.New("mount event payload size invalid")
	}
	decoder := json.NewDecoder(strings.NewReader(string(envelope.Payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		return event, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return event, errors.New("mount event contains trailing JSON")
	}
	if err := validate.Struct(event); err != nil {
		return event, err
	}
	if event.MountPoint == "" || len(event.MountPoint) > 4096 {
		return event, errors.New("invalid mount point")
	}
	return event, nil
}

// coordinateScans reconciles existing mounts once, then consumes lifecycle
// events. It keeps at most one active scan per mount ID and cancels old scans
// on unmount or identity change. Queue saturation is logged explicitly.
func coordinateScans(ctx context.Context, cfg antivirusConfig, incoming <-chan mountcontract.LifecycleEvent,
	jobs chan<- scanJob, results <-chan scanResult, effects chan<- scanResult, log sharedlogger.Logger,
) {
	defer close(effects)
	defer close(jobs)
	active := make(map[uint64]context.CancelFunc)
	activeMounts := make(map[uint64]mountinfo.Record)
	completed := make(map[uint64]mountinfo.Record)
	pending := make(map[uint64]mountinfo.Record)
	schedule := func(mount mountinfo.Record) {
		if !eligibleMount(cfg, mount) || activeMounts[mount.ID] == mount || completed[mount.ID] == mount {
			return
		}
		if cancel := active[mount.ID]; cancel != nil {
			cancel()
			pending[mount.ID] = mount
			return
		}
		jobCtx, cancel := context.WithCancel(ctx)
		select {
		case jobs <- scanJob{mount: mount, ctx: jobCtx}:
			active[mount.ID] = cancel
			activeMounts[mount.ID] = mount
		case <-ctx.Done():
			cancel()
			log.Warn("scan not queued during shutdown", "mount_id", mount.ID)
		default:
			cancel()
			log.Error("antivirus scan queue full", "mount_id", mount.ID)
		}
	}
	if mounts, err := mountinfo.ReadCurrent(); err != nil {
		log.Error("startup mount reconciliation failed", "error", err)
	} else {
		for _, mount := range mounts {
			schedule(mount)
		}
	}
	for incoming != nil || len(active) > 0 {
		select {
		case event, ok := <-incoming:
			if !ok {
				incoming = nil
				continue
			}
			if event.Action == mountcontract.Unmounted {
				if cancel := active[event.MountID]; cancel != nil {
					cancel()
				}
				delete(pending, event.MountID)
				delete(completed, event.MountID)
				continue
			}
			mounts, err := mountinfo.ReadCurrent()
			if err != nil {
				log.Error("mount recheck failed", "error", err)
				continue
			}
			mount, exists := mounts[event.MountID]
			if !exists || mount.Point != event.MountPoint || mount.Source != event.Source || mount.FSType != event.FSType {
				log.Warn("stale mount notification", "event_id", event.EventID)
				continue
			}
			schedule(mount)
		case result, ok := <-results:
			if !ok {
				results = nil
				continue
			}
			if cancel := active[result.mount.ID]; cancel != nil {
				cancel()
			}
			delete(active, result.mount.ID)
			delete(activeMounts, result.mount.ID)
			if result.status == scanClean || result.status == scanInfected {
				completed[result.mount.ID] = result.mount
			}
			effects <- result
			if next, exists := pending[result.mount.ID]; exists {
				delete(pending, result.mount.ID)
				schedule(next)
			}
		}
	}
}

// runScanWorker serially executes bounded scans and drains queued jobs as
// cancelled results after process cancellation.
func runScanWorker(_ context.Context, jobs <-chan scanJob, results chan<- scanResult, timeout time.Duration) {
	defer close(results)
	for job := range jobs {
		results <- scanMount(job.ctx, job.mount, timeout)
	}
}

// reportScans emits malware through the security logging manager's existing
// alert path, which the security email notifier already consumes.
func reportScans(ctx context.Context, effects <-chan scanResult, client messaging.Client, log sharedlogger.Logger) {
	for result := range effects {
		if result.status == scanClean {
			log.Info("antivirus scan clean", "mount_id", result.mount.ID)
			continue
		}
		if result.status == scanCancelled {
			log.Warn("antivirus scan cancelled", "mount_id", result.mount.ID)
			continue
		}
		if result.eventID == "" {
			id, err := newScanEventID()
			if err != nil {
				log.Error("antivirus scan event ID generation failed", "mount_id", result.mount.ID, "error", err)
				continue
			}
			result.eventID = id
		}
		for {
			attemptCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := publishScanResult(attemptCtx, result, client)
			if cancel != nil {
				cancel()
			}
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				log.Error("scan result undelivered at shutdown", "mount_id", result.mount.ID, "error", err)
				break
			}
			log.Error("scan result publication failed; retrying", "mount_id", result.mount.ID, "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// publishScanResult distinguishes malware from an incomplete scan. The scan
// result owns a unique event ID so retries update only that scan's event.
func publishScanResult(ctx context.Context, result scanResult, client messaging.Client) error {
	if result.status == scanInfected {
		message := "Malware detected on " + result.mount.Point
		if len(message) > 256 {
			message = message[:256]
		}
		evidence := result.evidence
		if len(evidence) > 512 {
			evidence = evidence[:512]
		}
		if err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
			EventID: result.eventID, Category: "Malware detection", Severity: enum.SeverityCritical,
			CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
			Message: message, TriggerValue: evidence,
		}); err != nil {
			return err
		}
		response := loggingcontract.OKResponse{}
		resolved := "Malware detection recorded for scan on " + result.mount.Point
		if len(resolved) > 256 {
			resolved = resolved[:256]
		}
		if err := client.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: result.eventID, Status: enum.StatusNormal, Message: &resolved}, &response); err != nil {
			return err
		}
		if !response.OK {
			return errors.New("security logging manager rejected malware event resolution")
		}
		return nil
	}
	message := "Antivirus scan incomplete on " + result.mount.Point
	if len(message) > 256 {
		message = message[:256]
	}
	return client.Publish(ctx, systemcontract.AlertSubject, loggingcontract.AlertPayload{
		EventID: result.eventID, Category: "Antivirus scan", Severity: enum.SeverityWarning,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
		Message: message,
	})
}

// newScanEventID creates a bounded identifier unique to one scan result.
func newScanEventID() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "A" + hex.EncodeToString(bytes[:]), nil
}
