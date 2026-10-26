package main

import (
	"context"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/messaging"
	"lite-nas/shared/systemd"
)

// Requirements: task.md/SVC-005.
func TestWatchdogDefersToSystemdAutoRestart(t *testing.T) {
	t.Parallel()
	now := time.Now()
	cfg := watchdogConfig{RecoveryDelay: time.Second, MaxAttempts: 3}
	progress := unitProgress{FailureSince: now.Add(-time.Minute)}
	unit := protectedUnit{Name: "sample.service", Recover: true}
	if shouldAttempt(now, cfg, unit, systemd.UnitState{ActiveState: "activating", SubState: "auto-restart"}, progress) {
		t.Fatal("attempted during systemd auto-restart")
	}
}

// Requirements: task.md/SVC-005.
func TestWatchdogRecoversManualStopAfterDelay(t *testing.T) {
	t.Parallel()
	now := time.Now()
	cfg := watchdogConfig{RecoveryDelay: time.Second, MaxAttempts: 3}
	progress := unitProgress{FailureSince: now.Add(-time.Minute)}
	unit := protectedUnit{Name: "sample.service", Recover: true}
	if !shouldAttempt(now, cfg, unit, systemd.UnitState{ActiveState: "inactive", SubState: "dead"}, progress) {
		t.Fatal("manual stop was not eligible")
	}
}

// Requirements: task.md/SVC-005.
func TestWatchdogHonorsAttemptLimit(t *testing.T) {
	t.Parallel()
	now := time.Now()
	cfg := watchdogConfig{RecoveryDelay: time.Second, MaxAttempts: 3}
	progress := unitProgress{FailureSince: now.Add(-time.Minute), Attempts: 3}
	unit := protectedUnit{Name: "sample.service", Recover: true}
	if shouldAttempt(now, cfg, unit, systemd.UnitState{ActiveState: "failed"}, progress) {
		t.Fatal("attempt limit bypassed")
	}
}

// Requirements: task.md/SVC-005.
func TestWatchdogRejectsMalformedNATSUnit(t *testing.T) {
	t.Parallel()
	envelope := messaging.Envelope{Payload: []byte(`{"version":1,"event_id":"one","unit":"--bad.service","condition":"failed","active_state":"failed","sub_state":"failed","observed_at":"2026-10-26T18:00:00Z"}`)}
	_, err := decodeServiceEvent(envelope, validator.New())
	if err == nil {
		t.Fatal("option-shaped NATS unit accepted")
	}
}

// Requirements: task.md/SVC-005.
func TestWatchdogRejectsUnallowlistedRecoveryJob(t *testing.T) {
	t.Parallel()
	jobs := make(chan recoveryIntent, 1)
	results := make(chan recoveryResult, 1)
	jobs <- recoveryIntent{unit: "unlisted.service"}
	close(jobs)
	recoverServices(context.Background(), watchdogConfig{Units: map[string]protectedUnit{}}, jobs, results)
	if result := <-results; result.err == nil {
		t.Fatal("unallowlisted recovery accepted")
	}
}
