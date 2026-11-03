package main

import (
	"context"
	"errors"
	"testing"
	"time"

	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/logger"
)

// mustIncidentStore creates an empty durable store for effect tests.
func mustIncidentStore(t *testing.T) *incidentStore {
	t.Helper()
	store, err := openIncidentStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return store
}

// mustPendingIncident persists an effect intent before a host command.
func mustPendingIncident(t *testing.T, store *incidentStore) incidentRecord {
	t.Helper()
	record := incidentRecord{
		ID: "A1B2C3D4E5F6G7H8", RuleID: "network_reconnaissance", SourceIP: "198.51.100.10",
		Interfaces: []string{"physical0"}, StartedAt: time.Now().UTC(), ExpiresAt: time.Now().UTC().Add(30 * time.Minute),
		State: contract.StatePending, Outcome: "operator block queued",
	}
	if err := store.save(record); err != nil {
		t.Fatal(err)
	}
	return record
}

// TestIncidentStoreReloadsPendingEffect verifies crash recovery sees durable intent.
// Requirements: security-automation/SR-002.
func TestIncidentStoreReloadsPendingEffect(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	reloaded, err := openIncidentStore(store.directory)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := reloaded.get(record.ID)
	if !ok || got.State != contract.StatePending {
		t.Fatalf("reloaded pending incident = %+v, %t", got, ok)
	}
}

// TestApplyIncidentPersistsVerifiedState checks the executor result is journaled.
// Requirements: security-automation/SR-002.
func TestApplyIncidentPersistsVerifiedState(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	runtime := &automationRuntime{incidents: store, log: logger.NopLogger{}, runUFW: func(context.Context, string, string, string, string, string) error { return nil }}
	runtime.applyIncident(context.Background(), record.ID)
	got, _ := store.get(record.ID)
	if got.State != contract.StateMitigating {
		t.Fatalf("incident state = %q", got.State)
	}
}

// TestRestoreIncidentRemovesOnlyOwnedRecord checks successful restore retires it.
// Requirements: security-automation/SR-002.
func TestRestoreIncidentRemovesOnlyOwnedRecord(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	runtime := &automationRuntime{incidents: store, log: logger.NopLogger{}, runUFW: func(context.Context, string, string, string, string, string) error { return nil }}
	if err := runtime.restoreIncident(context.Background(), record.ID); err != nil {
		t.Fatal(err)
	}
	_, exists := store.get(record.ID)
	if exists {
		t.Fatal("restored incident retained")
	}
}

// TestOperatorWorkerRestoresExpiredRecordOnStartup checks a restart does not
// leave a 30-minute UFW rule active past its recorded expiry.
// Requirements: security-automation/SR-002.
func TestOperatorWorkerRestoresExpiredRecordOnStartup(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	record.StartedAt = time.Now().UTC().Add(-time.Hour)
	record.ExpiresAt = record.StartedAt.Add(30 * time.Minute)
	record.State = contract.StateMitigating
	if err := store.save(record); err != nil {
		t.Fatal(err)
	}
	runtime := &automationRuntime{
		incidents: store, log: logger.NopLogger{}, operatorActions: make(chan effectCommand),
		runUFW: func(context.Context, string, string, string, string, string) error { return nil },
	}
	done := startOperatorEffectsWorker(context.Background(), runtime)
	close(runtime.operatorActions)
	<-done
	if _, exists := store.get(record.ID); exists {
		t.Fatal("expired incident survived startup reconciliation")
	}
}

// TestRollbackIntentSurvivesConcurrentApply verifies a finishing UFW command
// cannot overwrite a durable operator restoration request.
// Requirements: security-automation/SR-002.
func TestRollbackIntentSurvivesConcurrentApply(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	started := make(chan struct{})
	release := make(chan struct{})
	runtime := &automationRuntime{
		incidents: store, log: logger.NopLogger{}, operatorActions: make(chan effectCommand, 1),
		runUFW: func(context.Context, string, string, string, string, string) error {
			close(started)
			<-release
			return nil
		},
	}
	finished := make(chan struct{})
	go func() {
		runtime.applyIncident(context.Background(), record.ID)
		close(finished)
	}()
	<-started
	if _, err := runtime.queueRestore(context.Background(), record.ID, "operator"); err != nil {
		t.Fatal(err)
	}
	close(release)
	<-finished
	got, _ := store.get(record.ID)
	if got.State != contract.StateRestoring {
		t.Fatalf("rollback intent overwritten by apply: %q", got.State)
	}
}

// TestRestoreIncidentRetainsPendingAlert verifies a successful UFW deletion
// does not discard the incident while alert normalization is unacknowledged.
// Requirements: security-automation/FR-007.
func TestRestoreIncidentRetainsPendingAlert(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	runtime := &automationRuntime{
		incidents: store, log: logger.NopLogger{},
		runUFW: func(context.Context, string, string, string, string, string) error { return nil },
		audit: func(_ context.Context, _ incidentRecord, normalize bool) error {
			if normalize {
				return errors.New("not acknowledged")
			}
			return nil
		},
	}
	_ = runtime.restoreIncident(context.Background(), record.ID)
	got, exists := store.get(record.ID)
	if !exists || got.State != contract.StateRestoring {
		t.Fatalf("unacknowledged incident = %+v, %t", got, exists)
	}
}
