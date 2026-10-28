package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	sharedcontracts "lite-nas/shared/contracts"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	mountcontract "lite-nas/shared/contracts/mountmonitor"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/eventmanager"
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
		if prefix := excludedMountPrefix(cfg, mount.Point); prefix != "" {
			log.Warn("mount scan skipped by configured exclusion", "mount_id", mount.ID,
				"mount_point", mount.Point, "excluded_prefix", prefix, "filesystem_type", mount.FSType)
			return
		}
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
				for localID, mount := range activeMounts {
					if sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
						if cancel := active[localID]; cancel != nil {
							cancel()
						}
						delete(pending, localID)
						delete(completed, localID)
					}
				}
				for localID, mount := range pending {
					if sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
						delete(pending, localID)
					}
				}
				for localID, mount := range completed {
					if sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
						delete(completed, localID)
					}
				}
				continue
			}
			mounts, err := mountinfo.ReadCurrent()
			if err != nil {
				log.Error("mount recheck failed", "error", err)
				continue
			}
			mount, exists := findMountByIdentity(mounts, event)
			if !exists {
				log.Warn("stale mount notification", "event_id", event.EventID,
					"mount_id", event.MountID, "mount_point", event.MountPoint,
					"source", event.Source, "filesystem_type", event.FSType)
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

// findMountByIdentity resolves a lifecycle notification in the receiver's
// mount namespace, where kernel mount IDs may differ from the producer's IDs.
func findMountByIdentity(mounts map[uint64]mountinfo.Record, event mountcontract.LifecycleEvent) (mountinfo.Record, bool) {
	if mount, exists := mounts[event.MountID]; exists && sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
		return mount, true
	}
	var match mountinfo.Record
	found := false
	for _, mount := range mounts {
		if !sameMountIdentity(mount, event.MountPoint, event.Source, event.FSType) {
			continue
		}
		if found {
			return mountinfo.Record{}, false
		}
		match = mount
		found = true
	}
	return match, found
}

// sameMountIdentity compares mount metadata shared across mount namespaces.
func sameMountIdentity(mount mountinfo.Record, point, source, fsType string) bool {
	return mount.Point == point && mount.Source == source && mount.FSType == fsType
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
		if result.status == scanCancelled {
			log.Warn("antivirus scan cancelled", "mount_id", result.mount.ID)
			continue
		}
		if result.status == scanIncomplete {
			log.Warn("antivirus scan incomplete", "mount_id", result.mount.ID,
				"mount_point", result.mount.Point, "error", result.err)
		}
		if result.eventID == "" {
			id := newScanEventID()
			result.eventID = id
		}
		alertSent := false
		for {
			attemptCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := publishScanResult(attemptCtx, result, &alertSent, client)
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

// publishScanResult distinguishes clean scans, malware and incomplete scans. The scan
// result owns a unique event ID so retries update only that scan's event.
func publishScanResult(ctx context.Context, result scanResult, alertSent *bool, client messaging.Client) error {
	if result.status == scanInfected {
		message := "Malware detected on " + result.mount.Point
		if len(message) > 256 {
			message = message[:256]
		}
		evidence := result.evidence
		if len(evidence) > 512 {
			evidence = evidence[:512]
		}
		email := true
		if !*alertSent {
			if err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
				EventID: result.eventID, Category: "Malware detection", Severity: enum.SeverityError,
				CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
				Message: message, TriggerValue: evidence, Email: &email,
			}); err != nil {
				return err
			}
			// The manager handles event creation asynchronously; do not publish
			// another active create while retrying its immediate resolution.
			*alertSent = true
		}
		response := loggingcontract.OKResponse{}
		findingCount := strings.Count(result.evidence, " FOUND")
		resolved := fmt.Sprintf("Malware detection recorded for scan on %s (%d findings)", result.mount.Point, findingCount)
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
	if result.status == scanClean {
		message := "Antivirus scan clean on " + result.mount.Point
		if len(message) > 256 {
			message = message[:256]
		}
		if !*alertSent {
			if err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
				EventID: result.eventID, Category: "Antivirus scan", Severity: enum.SeverityInfo,
				CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
				Message: message,
			}); err != nil {
				return err
			}
			*alertSent = true
		}
		response := loggingcontract.OKResponse{}
		if err := client.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: result.eventID, Status: enum.StatusNormal}, &response); err != nil {
			return err
		}
		if !response.OK {
			return errors.New("security logging manager rejected clean scan resolution")
		}
		return nil
	}
	message := "Antivirus scan incomplete on " + result.mount.Point
	if result.err != nil {
		message += ": " + result.err.Error()
	}
	if len(message) > 256 {
		message = message[:256]
	}
	if *alertSent {
		return nil
	}
	evidence := result.evidence
	if len(evidence) > 512 {
		evidence = evidence[:512]
	}
	err := client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
		EventID: result.eventID, Category: "Antivirus scan", Severity: enum.SeverityWarning,
		CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceAntivirus,
		Message: message, TriggerValue: evidence,
	})
	if err != nil {
		return err
	}
	*alertSent = true
	return nil
}

// monitorEventIDs shares the resources monitor's time-seeded counter convention.
var monitorEventIDs = eventmanager.NewManager(uint64(time.Now().Unix()) % 99_999_999)

// newScanEventID allocates a readable ID for one new event.
func newScanEventID() string {
	return monitorEventIDs.NextEventID("secav")
}
