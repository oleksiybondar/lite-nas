package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"lite-nas/services/system-metrics/modules"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestServeSnapshotsStoresAndPublishesSnapshot(t *testing.T) {
	t.Parallel()

	store := modules.NewStateModule(2).SnapshotStore()
	client := &recordingClient{}
	log := &recordingLogger{}
	input := make(chan metrics.SystemSnapshot, 1)
	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(100, 0)}

	input <- snapshot
	close(input)

	if err := serveSnapshots(context.Background(), input, store, client, log); err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}

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

	if client.publishCalls[0].subject != statsEventSubject {
		t.Fatalf("publish subject = %q, want %q", client.publishCalls[0].subject, statsEventSubject)
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestServeSnapshotsStoresSnapshotWhenPublishFails(t *testing.T) {
	t.Parallel()

	store := modules.NewStateModule(2).SnapshotStore()
	client := &recordingClient{publishErr: errors.New("publish failed")}
	log := &recordingLogger{}
	input := make(chan metrics.SystemSnapshot, 1)
	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(101, 0)}

	input <- snapshot
	close(input)

	if err := serveSnapshots(context.Background(), input, store, client, log); err != nil {
		t.Fatalf("serveSnapshots() error = %v", err)
	}

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

	store := modules.NewStateModule(1).SnapshotStore()
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
	store := modules.NewStateModule(2).SnapshotStore()

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	response, err := server.rpcHandlers[statsRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("stats handler error = %v", err)
	}

	emptySnapshot, ok := response.(map[string]any)
	if !ok {
		t.Fatalf("stats response type = %T, want map[string]any", response)
	}

	if len(emptySnapshot) != 0 {
		t.Fatalf("stats response len = %d, want 0", len(emptySnapshot))
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/IR-001
func TestRegisterRPCHandlersStatsReturnsLatestSnapshot(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := modules.NewStateModule(2).SnapshotStore()
	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(102, 0)}
	store.Add(snapshot)

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	response, err := server.rpcHandlers[statsRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("stats handler error = %v", err)
	}

	gotSnapshot, ok := response.(metrics.SystemSnapshot)
	if !ok {
		t.Fatalf("stats response type = %T, want metrics.SystemSnapshot", response)
	}

	if !reflect.DeepEqual(gotSnapshot, snapshot) {
		t.Fatalf("stats response = %#v, want %#v", gotSnapshot, snapshot)
	}
}

// Requirements: system-metrics-svc/FR-004, system-metrics-svc/IR-001
func TestRegisterRPCHandlersHistoryReturnsEmptyListBeforeData(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := modules.NewStateModule(2).SnapshotStore()

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	response, err := server.rpcHandlers[historyRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("history handler error = %v", err)
	}

	history, ok := response.([]metrics.SystemSnapshot)
	if !ok {
		t.Fatalf("history response type = %T, want []metrics.SystemSnapshot", response)
	}

	if len(history) != 0 {
		t.Fatalf("history len = %d, want 0", len(history))
	}
}

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-004, system-metrics-svc/IR-001
func TestRegisterRPCHandlersHistoryReturnsChronologicalHistory(t *testing.T) {
	t.Parallel()

	server := &recordingServer{}
	store := modules.NewStateModule(3).SnapshotStore()
	first := metrics.SystemSnapshot{Timestamp: time.Unix(103, 0)}
	second := metrics.SystemSnapshot{Timestamp: time.Unix(104, 0)}
	store.Add(first)
	store.Add(second)

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	response, err := server.rpcHandlers[historyRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("history handler error = %v", err)
	}

	history, ok := response.([]metrics.SystemSnapshot)
	if !ok {
		t.Fatalf("history response type = %T, want []metrics.SystemSnapshot", response)
	}

	wantHistory := []metrics.SystemSnapshot{first, second}
	if !reflect.DeepEqual(history, wantHistory) {
		t.Fatalf("history = %#v, want %#v", history, wantHistory)
	}
}

func TestRegisterRPCHandlersReturnsStatsRegistrationError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("register failed")
	server := &recordingServer{
		registerRPCErrors: map[string]error{statsRPCSubject: expectedErr},
	}

	err := registerRPCHandlers(server, modules.NewStateModule(1).SnapshotStore())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, expectedErr)
	}
}

func TestRegisterRPCHandlersReturnsHistoryRegistrationError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("register failed")
	server := &recordingServer{
		registerRPCErrors: map[string]error{historyRPCSubject: expectedErr},
	}

	err := registerRPCHandlers(server, modules.NewStateModule(1).SnapshotStore())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, expectedErr)
	}
}

// Requirements: system-metrics-svc/FR-003, system-metrics-svc/FR-005, system-metrics-svc/IR-002
func TestStoreAndPublishSnapshotStoresLatestSnapshot(t *testing.T) {
	t.Parallel()

	store := modules.NewStateModule(1).SnapshotStore()
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
