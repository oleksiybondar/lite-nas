package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"lite-nas/services/system-metrics/modules"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestServeSnapshotsStoresAndPublishesSnapshot(t *testing.T) {
	t.Parallel()

	store, client, _, snapshot := runServeSnapshotsFixture(t, nil, time.Unix(100, 0))

	latest, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(latest, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", latest, snapshot)
	}

	if len(client.publishCalls) != 1 {
		t.Fatalf("publishCalls = %d, want 1", len(client.publishCalls))
	}

	if client.publishCalls[0].subject != systemmetricscontract.SnapshotEventSubject {
		t.Fatalf("publish subject = %q, want %q", client.publishCalls[0].subject, systemmetricscontract.SnapshotEventSubject)
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestServeSnapshotsStoresSnapshotWhenPublishFails(t *testing.T) {
	t.Parallel()

	store, _, log, snapshot := runServeSnapshotsFixture(t, errors.New("publish failed"), time.Unix(101, 0))

	latest, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(latest, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", latest, snapshot)
	}

	if len(log.warns) != 1 {
		t.Fatalf("warn count = %d, want 1", len(log.warns))
	}
}

func TestServeSnapshotsReturnsContextErrorOnCancellation(t *testing.T) {
	t.Parallel()

	store := newSnapshotStore(1)
	client := &recordingClient{}
	log := &recordingLogger{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := serveSnapshots(ctx, make(chan metrics.SystemSnapshot), store, client, log)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want %v", err, context.Canceled)
	}

	if len(log.infos) != 1 {
		t.Fatalf("info count = %d, want 1", len(log.infos))
	}
}

func TestGracefulExitReturnsContextError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := gracefulExit(ctx, log)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("gracefulExit() error = %v, want %v", err, context.Canceled)
	}

	if len(log.infos) != 1 {
		t.Fatalf("info count = %d, want 1", len(log.infos))
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/IR-001
func TestRegisterRPCHandlersStatsReturnsEmptySnapshotBeforeData(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := newSnapshotStore(2)
	mustRegisterRPCHandlers(t, server, store)
	snapshotResponse := mustInvokeSnapshotRPC(t, server)

	if snapshotResponse.Available {
		t.Fatal("stats response Available = true, want false")
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/IR-001
func TestRegisterRPCHandlersStatsReturnsLatestSnapshot(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := newSnapshotStore(2)
	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(102, 0)}
	store.Add(snapshot)
	mustRegisterRPCHandlers(t, server, store)
	snapshotResponse := mustInvokeSnapshotRPC(t, server)

	if !snapshotResponse.Available {
		t.Fatal("stats response Available = false, want true")
	}

	if !reflect.DeepEqual(snapshotResponse.Snapshot, snapshot) {
		t.Fatalf("stats response = %#v, want %#v", snapshotResponse.Snapshot, snapshot)
	}
}

// Requirements: system-metrics-svc/FR-004, system-metrics-svc/IR-001
func TestRegisterRPCHandlersHistoryReturnsEmptyListBeforeData(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := newSnapshotStore(2)
	mustRegisterRPCHandlers(t, server, store)
	historyResponse := mustInvokeHistoryRPC(t, server)

	if len(historyResponse.Items) != 0 {
		t.Fatalf("history len = %d, want 0", len(historyResponse.Items))
	}
}

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-004, system-metrics-svc/IR-001
func TestRegisterRPCHandlersHistoryReturnsChronologicalHistory(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := newSnapshotStore(3)
	first := metrics.SystemSnapshot{Timestamp: time.Unix(103, 0)}
	second := metrics.SystemSnapshot{Timestamp: time.Unix(104, 0)}
	store.Add(first)
	store.Add(second)

	mustRegisterRPCHandlers(t, server, store)
	historyResponse := mustInvokeHistoryRPC(t, server)

	wantHistory := []metrics.SystemSnapshot{first, second}
	if !reflect.DeepEqual(historyResponse.Items, wantHistory) {
		t.Fatalf("history = %#v, want %#v", historyResponse.Items, wantHistory)
	}
}

func TestRegisterRPCHandlersReturnsStatsRegistrationError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("register failed")
	server := &recordingServer{
		registerRPCErrors: map[string]error{systemmetricscontract.SnapshotRPCSubject: expectedErr},
	}

	err := registerRPCHandlers(server, newSnapshotStore(1))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, expectedErr)
	}
}

func TestRegisterRPCHandlersReturnsHistoryRegistrationError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("register failed")
	server := &recordingServer{
		registerRPCErrors: map[string]error{systemmetricscontract.HistoryRPCSubject: expectedErr},
	}

	err := registerRPCHandlers(server, newSnapshotStore(1))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, expectedErr)
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestStoreAndPublishSnapshotStoresLatestSnapshot(t *testing.T) {
	t.Parallel()

	store := newSnapshotStore(1)
	client := &recordingClient{}
	log := &recordingLogger{}
	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(105, 0)}

	storeAndPublishSnapshot(context.Background(), snapshot, store, client, log)

	latest, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(latest, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", latest, snapshot)
	}
}

func runServeSnapshotsFixture(
	t *testing.T,
	publishErr error,
	timestamp time.Time,
) (*modules.SnapshotStore, *recordingClient, *recordingLogger, metrics.SystemSnapshot) {
	t.Helper()

	store := newSnapshotStore(2)
	client := &recordingClient{publishErr: publishErr}
	log := &recordingLogger{}
	input := make(chan metrics.SystemSnapshot, 1)
	snapshot := metrics.SystemSnapshot{Timestamp: timestamp}

	input <- snapshot
	close(input)

	if err := serveSnapshots(context.Background(), input, store, client, log); err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}

	return store, client, log, snapshot
}
