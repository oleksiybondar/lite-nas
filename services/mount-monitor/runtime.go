package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	sharedcontracts "lite-nas/shared/contracts"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	mountcontract "lite-nas/shared/contracts/mountmonitor"
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
	packagedMountConfig = "/etc/lite-nas/mount-monitor.conf"
	mountQueueCapacity  = 8
)

// mountSnapshot is one complete read; a failed read never reaches the diff
// worker, preventing transient parse errors from looking like unmounts.
type mountSnapshot struct {
	mounts map[uint64]mountRecord
	at     time.Time
}

// mountIntent holds one lifecycle event until both NATS and the operational
// alert path have been attempted in order.
type mountIntent struct {
	event mountcontract.LifecycleEvent
}

// run assembles one process with polling, diffing, and outbound workers.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedMountConfig)
	if err != nil {
		return err
	}
	cfg, err := loadMountConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceMountMonitor, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	return runMountPipeline(ctx, cfg.PollInterval, infra.Client,
		loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken),
		infra.Logger, readCurrentMounts)
}

// readCurrentMounts reads the process mount namespace without privileges.
func readCurrentMounts() (map[uint64]mountRecord, error) {
	return mountinfo.ReadCurrent()
}

// runMountPipeline owns both bounded channels. The poller alone closes
// snapshots; the diff worker alone closes intents after consuming snapshots.
// The publisher finishes queued intents before its done channel closes.
func runMountPipeline(ctx context.Context, interval time.Duration, eventClient, alertClient messaging.Client,
	log sharedlogger.Logger, read func() (map[uint64]mountRecord, error),
) error {
	timer, ticks, err := workers.NewPollingTimerWorker(interval, 1)
	if err != nil {
		return err
	}
	snapshots := make(chan mountSnapshot, mountQueueCapacity)
	intents := make(chan mountIntent, mountQueueCapacity)
	done := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) { pollMounts(workerCtx, ticks, snapshots, read, log) }).Start(ctx)
	workers.NewBaseWorker(func(_ context.Context) { diffMounts(snapshots, intents, log) }).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		publishMounts(workerCtx, intents, eventClient, alertClient, log)
		close(done)
	}).Start(ctx)
	timer.Start(ctx)
	<-done
	return ctx.Err()
}

// pollMounts sends complete snapshots with backpressure. Cancellation is
// observable and closes its single-producer channel after the final poll.
func pollMounts(ctx context.Context, ticks <-chan struct{}, out chan<- mountSnapshot,
	read func() (map[uint64]mountRecord, error), log sharedlogger.Logger,
) {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticks:
			mounts, err := read()
			if err != nil {
				log.Error("mountinfo read failed", "error", err)
				continue
			}
			select {
			case out <- mountSnapshot{mounts: mounts, at: time.Now().UTC()}:
			case <-ctx.Done():
				log.Warn("mount snapshot pending at shutdown")
				return
			}
		}
	}
}

// diffMounts uses the first complete snapshot as a baseline and fans every
// later state change into the outbound queue; it never drops a queued change.
func diffMounts(snapshots <-chan mountSnapshot, out chan<- mountIntent, log sharedlogger.Logger) {
	defer close(out)
	var previous map[uint64]mountRecord
	for snapshot := range snapshots {
		if previous == nil {
			previous = snapshot.mounts
			continue
		}
		for _, change := range changedMounts(previous, snapshot.mounts) {
			id, err := newMountEventID()
			if err != nil {
				log.Error("mount event ID generation failed", "error", err)
				continue
			}
			event := mountcontract.LifecycleEvent{
				Version: mountcontract.CurrentVersion, EventID: id,
				Action: mountcontract.Action(change.action), MountID: change.mount.ID,
				MountPoint: change.mount.Point, Source: change.mount.Source,
				FSType: change.mount.FSType, ObservedAt: snapshot.at.Format(time.RFC3339),
			}
			out <- mountIntent{event: event}
		}
		previous = snapshot.mounts
	}
}

// newMountEventID uses randomness so a restart cannot reuse a logging-manager
// event identifier. The 17-character ID fits its 20-character contract.
func newMountEventID() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return "M" + hex.EncodeToString(bytes[:]), nil
}

// publishMounts retries a pending intent during normal operation. At process
// shutdown it makes one bounded final attempt and logs any undelivered intent.
func publishMounts(ctx context.Context, intents <-chan mountIntent, eventClient, alertClient messaging.Client, log sharedlogger.Logger) {
	for intent := range intents {
		for {
			attemptCtx := ctx
			if ctx.Err() != nil {
				var cancel context.CancelFunc
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
				err := publishMountIntent(attemptCtx, intent, eventClient, alertClient)
				cancel()
				if err != nil {
					log.Error("mount intent undelivered at shutdown", "event_id", intent.event.EventID, "error", err)
				}
				break
			}
			if err := publishMountIntent(attemptCtx, intent, eventClient, alertClient); err == nil {
				break
			} else {
				log.Error("mount intent publication failed; retrying", "event_id", intent.event.EventID, "error", err)
			}
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// publishMountIntent emits the NATS lifecycle event and the informational
// system alert for mounts and unmounts. Duplicate event IDs let consumers
// deduplicate retries if the first publication succeeded remotely.
func publishMountIntent(ctx context.Context, intent mountIntent, eventClient, alertClient messaging.Client) error {
	if err := eventClient.Publish(ctx, mountcontract.LifecycleSubject, intent.event); err != nil {
		return err
	}
	if intent.event.Action == mountcontract.Changed {
		return nil
	}
	message := fmt.Sprintf("Mount %s: %s", intent.event.Action, intent.event.MountPoint)
	if len(message) > 256 {
		message = message[:256]
	}
	alert := loggingmanagercontract.AlertPayload{
		EventID: intent.event.EventID, Category: "Mount lifecycle", Severity: enum.SeverityInfo,
		CreatedAt: intent.event.ObservedAt, Source: sharedcontracts.ServiceMountMonitor,
		Message: message, TriggerValue: intent.event.MountPoint,
	}
	if len(alert.TriggerValue) > 512 {
		alert.TriggerValue = alert.TriggerValue[:512]
	}
	return alertClient.Publish(ctx, systemcontract.AlertSubject, alert)
}
