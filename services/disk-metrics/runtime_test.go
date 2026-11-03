package main

import (
	"context"
	"errors"
	"testing"

	diskstate "lite-nas/services/disk-metrics/state"
	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/metricsruntimetest"
)

func TestRegisterRPCHandlers(t *testing.T) {
	t.Parallel()

	runEmptyStoreSnapshotCase(t)
	runRegisterRPCHandlersErrorCase(t)
	runLatestSnapshotAndHistoryCase(t)
}

func TestHandlePollError(t *testing.T) {
	t.Parallel()

	log := &metricsruntimetest.RecordingLogger{}
	for _, open := range []bool{true, false} {
		handlePollError(log, errors.New("poll failed"), open)
	}

	if got := len(log.Errors); got != 1 {
		t.Fatalf("Error logs = %d, want 1", got)
	}
}

func TestHandleSnapshot(t *testing.T) {
	t.Parallel()

	runStoreAndPublishSnapshotCase(t)
	runClosedInputSnapshotCase(t)
	runPublishFailureSnapshotCase(t)
}

func TestHandleShutdownReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := handleShutdown(ctx, &metricsruntimetest.RecordingLogger{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("handleShutdown() error = %v, want context.Canceled", err)
	}
}

func TestServeSnapshots(t *testing.T) {
	t.Parallel()

	runServeSnapshotsCanceledCase(t)
	runServeSnapshotsProcessesDataCase(t)
}

func runEmptyStoreSnapshotCase(t *testing.T) {
	t.Helper()

	t.Run("empty store returns unavailable snapshot", func(t *testing.T) {
		t.Parallel()

		store := diskstate.NewHistoryStore(2)
		server := mustRegisterTestHandlers(t, store)
		snapshotResponse := invokeTestRPC[diskmetricscontract.GetSnapshotResponse](
			t,
			server,
			diskmetricscontract.SnapshotRPCSubject,
		)
		if snapshotResponse.Available {
			t.Fatal("snapshot response Available = true, want false")
		}
	})
}

func runRegisterRPCHandlersErrorCase(t *testing.T) {
	t.Helper()

	t.Run("registration failure is returned", func(t *testing.T) {
		t.Parallel()

		server := &metricsruntimetest.RecordingServer{RegisterErr: errors.New("register failed")}
		if err := registerRPCHandlers(server, diskstate.NewHistoryStore(1)); !errors.Is(err, server.RegisterErr) {
			t.Fatalf("registerRPCHandlers() error = %v, want %v", err, server.RegisterErr)
		}
	})
}

func runLatestSnapshotAndHistoryCase(t *testing.T) {
	t.Helper()

	t.Run("latest snapshot and history are exposed", func(t *testing.T) {
		t.Parallel()

		store := diskstate.NewHistoryStore(2)
		store.Add(metrics.DiskMetricsSnapshot{})
		server := mustRegisterTestHandlers(t, store)

		current := invokeTestRPC[diskmetricscontract.GetSnapshotResponse](
			t,
			server,
			diskmetricscontract.SnapshotRPCSubject,
		)
		if !current.Available {
			t.Fatal("snapshot response Available = false, want true")
		}

		history := invokeTestRPC[diskmetricscontract.GetHistoryResponse](
			t,
			server,
			diskmetricscontract.HistoryRPCSubject,
		)
		if got := len(history.Items); got != 1 {
			t.Fatalf("history length = %d, want 1", got)
		}
	})
}

func runStoreAndPublishSnapshotCase(t *testing.T) {
	t.Helper()

	t.Run("stores and publishes snapshots", func(t *testing.T) {
		t.Parallel()

		store := diskstate.NewHistoryStore(2)
		client := &metricsruntimetest.RecordingClient{}
		stopped := handleSnapshot(
			context.Background(),
			store,
			client,
			&metricsruntimetest.RecordingLogger{},
			metrics.DiskMetricsSnapshot{},
			true,
		)
		if stopped {
			t.Fatal("handleSnapshot() stopped = true, want false")
		}
		if client.PublishCalls[0].Subject != diskmetricscontract.SnapshotEventSubject {
			t.Fatalf("Publish() subject = %q, want %q", client.PublishCalls[0].Subject, diskmetricscontract.SnapshotEventSubject)
		}
		if _, ok := store.Latest(); !ok {
			t.Fatal("Latest() ok = false, want stored snapshot")
		}
	})
}

func runClosedInputSnapshotCase(t *testing.T) {
	t.Helper()

	t.Run("closed input stops processing", func(t *testing.T) {
		t.Parallel()

		stopped := handleSnapshot(
			context.Background(),
			diskstate.NewHistoryStore(1),
			&metricsruntimetest.RecordingClient{},
			&metricsruntimetest.RecordingLogger{},
			metrics.DiskMetricsSnapshot{},
			false,
		)
		if !stopped {
			t.Fatal("handleSnapshot() stopped = false, want true")
		}
	})
}

func runPublishFailureSnapshotCase(t *testing.T) {
	t.Helper()

	t.Run("publish failure is logged", func(t *testing.T) {
		t.Parallel()

		log := &metricsruntimetest.RecordingLogger{}
		client := &metricsruntimetest.RecordingClient{PublishErr: errors.New("publish failed")}

		handleSnapshot(context.Background(), diskstate.NewHistoryStore(1), client, log, metrics.DiskMetricsSnapshot{}, true)

		if got := len(log.Errors); got != 1 {
			t.Fatalf("Error logs = %d, want 1", got)
		}
	})
}

func runServeSnapshotsCanceledCase(t *testing.T) {
	t.Helper()

	t.Run("returns when context is canceled", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := serveSnapshots(
			ctx,
			make(chan metrics.DiskMetricsSnapshot),
			make(chan error),
			diskstate.NewHistoryStore(1),
			&metricsruntimetest.RecordingClient{},
			&metricsruntimetest.RecordingLogger{},
		)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
		}
	})
}

func runServeSnapshotsProcessesDataCase(t *testing.T) {
	t.Helper()

	t.Run("processes poll errors and snapshots", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		input := make(chan metrics.DiskMetricsSnapshot, 1)
		pollErrors := make(chan error, 1)
		client := &metricsruntimetest.RecordingClient{PublishSignal: make(chan struct{}, 1)}
		log := &metricsruntimetest.RecordingLogger{}

		pollErrors <- errors.New("poll failed")
		input <- metrics.DiskMetricsSnapshot{}

		done := make(chan error, 1)
		go func() {
			done <- serveSnapshots(ctx, input, pollErrors, diskstate.NewHistoryStore(1), client, log)
		}()

		metricsruntimetest.WaitForPublish(t, client)
		cancel()

		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
		}
		if len(log.Errors) == 0 {
			t.Fatal("expected poll error to be logged")
		}
	})
}

func mustRegisterTestHandlers(t *testing.T, store *diskstate.HistoryStore) *metricsruntimetest.RecordingServer {
	t.Helper()

	server := &metricsruntimetest.RecordingServer{}
	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	return server
}

func invokeTestRPC[Response any](
	t *testing.T,
	server *metricsruntimetest.RecordingServer,
	subject string,
) Response {
	t.Helper()

	return metricsruntimetest.MustInvokeRPCHandler[Response](t, server, subject)
}
