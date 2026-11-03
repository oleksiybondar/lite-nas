package main

import (
	"context"
	"errors"
	"testing"

	networkstate "lite-nas/services/network-metrics/state"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/metricsruntimetest"
)

func TestRegisterRPCHandlersReturnsUnavailableSnapshotWhenStoreEmpty(t *testing.T) {
	t.Parallel()

	store := networkstate.NewHistoryStore(2)
	server := &metricsruntimetest.RecordingServer{}

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	typed := metricsruntimetest.MustInvokeRPCHandler[networkmetricscontract.GetSnapshotResponse](
		t,
		server,
		networkmetricscontract.SnapshotRPCSubject,
	)
	if typed.Available {
		t.Fatal("snapshot response Available = true, want false")
	}
}

func TestRegisterRPCHandlersReturnsServerError(t *testing.T) {
	t.Parallel()

	server := &metricsruntimetest.RecordingServer{RegisterErr: errors.New("register failed")}
	err := registerRPCHandlers(server, networkstate.NewHistoryStore(1))
	if !errors.Is(err, server.RegisterErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, server.RegisterErr)
	}
}

func TestRegisterRPCHandlersReturnsLatestSnapshotAndHistory(t *testing.T) {
	t.Parallel()

	store := networkstate.NewHistoryStore(2)
	snapshot := metrics.NetworkMetricsSnapshot{}
	store.Add(snapshot)
	server := &metricsruntimetest.RecordingServer{}

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	current := metricsruntimetest.MustInvokeRPCHandler[networkmetricscontract.GetSnapshotResponse](
		t,
		server,
		networkmetricscontract.SnapshotRPCSubject,
	)
	if !current.Available {
		t.Fatal("snapshot response Available = false, want true")
	}
	history := metricsruntimetest.MustInvokeRPCHandler[networkmetricscontract.GetHistoryResponse](
		t,
		server,
		networkmetricscontract.HistoryRPCSubject,
	)
	if len(history.Items) != 1 {
		t.Fatalf("history length = %d, want 1", len(history.Items))
	}
}

func TestHandlePollErrorLogsOnlyWhenChannelOpen(t *testing.T) {
	t.Parallel()

	log := &metricsruntimetest.RecordingLogger{}
	handlePollError(log, errors.New("boom"), true)
	handlePollError(log, errors.New("ignored"), false)

	if len(log.Errors) != 1 {
		t.Fatalf("Error logs = %d, want 1", len(log.Errors))
	}
}

func TestHandleSnapshotStoresAndPublishesSnapshot(t *testing.T) {
	t.Parallel()

	store := networkstate.NewHistoryStore(2)
	client := &metricsruntimetest.RecordingClient{}
	log := &metricsruntimetest.RecordingLogger{}
	snapshot := metrics.NetworkMetricsSnapshot{}

	stopped := handleSnapshot(context.Background(), store, client, log, snapshot, true)
	if stopped {
		t.Fatal("handleSnapshot() stopped = true, want false")
	}

	if client.PublishCalls[0].Subject != networkmetricscontract.SnapshotEventSubject {
		t.Fatalf("Publish() subject = %q, want %q", client.PublishCalls[0].Subject, networkmetricscontract.SnapshotEventSubject)
	}
	if _, ok := store.Latest(); !ok {
		t.Fatal("Latest() ok = false, want stored snapshot")
	}
}

func TestHandleSnapshotReturnsTrueWhenInputChannelClosed(t *testing.T) {
	t.Parallel()

	stopped := handleSnapshot(context.Background(), networkstate.NewHistoryStore(1), &metricsruntimetest.RecordingClient{}, &metricsruntimetest.RecordingLogger{}, metrics.NetworkMetricsSnapshot{}, false)
	if !stopped {
		t.Fatal("handleSnapshot() stopped = false, want true")
	}
}

func TestHandleSnapshotLogsPublishError(t *testing.T) {
	t.Parallel()

	log := &metricsruntimetest.RecordingLogger{}
	client := &metricsruntimetest.RecordingClient{PublishErr: errors.New("publish failed")}

	handleSnapshot(context.Background(), networkstate.NewHistoryStore(1), client, log, metrics.NetworkMetricsSnapshot{}, true)

	if len(log.Errors) != 1 {
		t.Fatalf("Error logs = %d, want 1", len(log.Errors))
	}
}

func TestHandleShutdownReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleShutdown(ctx, &metricsruntimetest.RecordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("handleShutdown() error = %v, want context.Canceled", err)
	}
}

func TestServeSnapshotsReturnsOnContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := serveSnapshots(ctx, make(chan metrics.NetworkMetricsSnapshot), make(chan error), networkstate.NewHistoryStore(1), &metricsruntimetest.RecordingClient{}, &metricsruntimetest.RecordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
	}
}

func TestServeSnapshotsProcessesPollErrorsAndSnapshots(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan metrics.NetworkMetricsSnapshot, 1)
	pollErrors := make(chan error, 1)
	store := networkstate.NewHistoryStore(1)
	client := &metricsruntimetest.RecordingClient{PublishSignal: make(chan struct{}, 1)}
	log := &metricsruntimetest.RecordingLogger{}
	snapshot := metrics.NetworkMetricsSnapshot{}

	pollErrors <- errors.New("poll failed")
	input <- snapshot

	done := make(chan error, 1)
	go func() {
		done <- serveSnapshots(ctx, input, pollErrors, store, client, log)
	}()

	metricsruntimetest.WaitForPublish(t, client)
	cancel()

	err := <-done
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
	}
	if len(log.Errors) == 0 {
		t.Fatal("expected poll error to be logged")
	}
}
