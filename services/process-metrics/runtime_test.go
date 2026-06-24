package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	processconfig "lite-nas/services/process-metrics/config"
	"lite-nas/services/process-metrics/modules"
	processstate "lite-nas/services/process-metrics/state"
	processmetricscontract "lite-nas/shared/contracts/processmetrics"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/metricsruntimetest"
)

func TestRegisterRPCHandlers(t *testing.T) {
	t.Parallel()

	t.Run("empty store returns unavailable snapshot", testRegisterRPCHandlersEmptyStore)
	t.Run("registration failure is returned", testRegisterRPCHandlersRegistrationFailure)
	t.Run("stored snapshot returns available response", testRegisterRPCHandlersStoredSnapshot)
}

func testRegisterRPCHandlersEmptyStore(t *testing.T) {
	t.Parallel()

	store := processstate.NewSnapshotStore()
	server := mustRegisterTestHandlers(t, store)
	snapshotResponse := invokeTestRPC[processmetricscontract.GetSnapshotResponse](
		t,
		server,
		processmetricscontract.SnapshotRPCSubject,
	)
	if snapshotResponse.Available {
		t.Fatal("snapshot response Available = true, want false")
	}
}

func testRegisterRPCHandlersRegistrationFailure(t *testing.T) {
	t.Parallel()

	server := &metricsruntimetest.RecordingServer{RegisterErr: errors.New("register failed")}
	if err := registerRPCHandlers(server, processstate.NewSnapshotStore()); !errors.Is(err, server.RegisterErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, server.RegisterErr)
	}
}

func testRegisterRPCHandlersStoredSnapshot(t *testing.T) {
	t.Parallel()

	store := processstate.NewSnapshotStore()
	store.Add(metrics.ProcessMetricsSnapshot{Processes: []metrics.ProcessMetric{{PID: 7}}})
	server := mustRegisterTestHandlers(t, store)
	snapshotResponse := invokeTestRPC[processmetricscontract.GetSnapshotResponse](
		t,
		server,
		processmetricscontract.SnapshotRPCSubject,
	)
	if !snapshotResponse.Available {
		t.Fatal("snapshot response Available = false, want true")
	}
	if got := len(snapshotResponse.Snapshot.Processes); got != 1 {
		t.Fatalf("len(Snapshot.Processes) = %d, want 1", got)
	}
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

	t.Run("stores and publishes snapshots", testHandleSnapshotStoresAndPublishes)
	t.Run("closed input stops processing", testHandleSnapshotStopsOnClosedInput)
	t.Run("publish failure is logged", testHandleSnapshotLogsPublishFailure)
}

func testHandleSnapshotStoresAndPublishes(t *testing.T) {
	t.Parallel()

	store := processstate.NewSnapshotStore()
	client := &metricsruntimetest.RecordingClient{}
	stopped := handleSnapshot(
		context.Background(),
		store,
		client,
		&metricsruntimetest.RecordingLogger{},
		metrics.ProcessMetricsSnapshot{},
		true,
	)
	if stopped {
		t.Fatal("handleSnapshot() stopped = true, want false")
	}
	if client.PublishCalls[0].Subject != processmetricscontract.SnapshotEventSubject {
		t.Fatalf("Publish() subject = %q, want %q", client.PublishCalls[0].Subject, processmetricscontract.SnapshotEventSubject)
	}
	if _, ok := store.Latest(); !ok {
		t.Fatal("Latest() ok = false, want stored snapshot")
	}
}

func testHandleSnapshotStopsOnClosedInput(t *testing.T) {
	t.Parallel()

	stopped := handleSnapshot(
		context.Background(),
		processstate.NewSnapshotStore(),
		&metricsruntimetest.RecordingClient{},
		&metricsruntimetest.RecordingLogger{},
		metrics.ProcessMetricsSnapshot{},
		false,
	)
	if !stopped {
		t.Fatal("handleSnapshot() stopped = false, want true")
	}
}

func testHandleSnapshotLogsPublishFailure(t *testing.T) {
	t.Parallel()

	log := &metricsruntimetest.RecordingLogger{}
	stopped := handleSnapshot(
		context.Background(),
		processstate.NewSnapshotStore(),
		&metricsruntimetest.RecordingClient{PublishErr: errors.New("publish failed")},
		log,
		metrics.ProcessMetricsSnapshot{},
		true,
	)
	if stopped {
		t.Fatal("handleSnapshot() stopped = true, want false")
	}
	if got := len(log.Errors); got != 1 {
		t.Fatalf("Error logs = %d, want 1", got)
	}
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

	t.Run("returns when context is canceled", testServeSnapshotsReturnsContextError)
	t.Run("logs poll errors while channel remains open", testServeSnapshotsLogsPollErrors)
}

func testServeSnapshotsReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := serveSnapshots(
		ctx,
		make(chan metrics.ProcessMetricsSnapshot),
		make(chan error),
		processstate.NewSnapshotStore(),
		&metricsruntimetest.RecordingClient{},
		&metricsruntimetest.RecordingLogger{},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
	}
}

func testServeSnapshotsLogsPollErrors(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan metrics.ProcessMetricsSnapshot)
	pollErrors := make(chan error, 1)
	log := &metricsruntimetest.RecordingLogger{}
	done := make(chan error, 1)
	go func() {
		done <- serveSnapshots(
			ctx,
			input,
			pollErrors,
			processstate.NewSnapshotStore(),
			&metricsruntimetest.RecordingClient{},
			log,
		)
	}()

	pollErrors <- errors.New("poll failed")
	waitForRecordedErrorLog(t, log)
	cancel()

	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
	}
}

func TestStartWorkers(t *testing.T) {
	t.Parallel()

	procRoot := createMinimalProcRoot(t)
	channels := modules.NewChannelsModule(1)
	workerModule, err := modules.NewWorkersModule(
		processconfig.ProcessMetricsConfig{
			Enabled:      true,
			PollInterval: time.Millisecond,
		},
		channels,
		modules.SourcePaths{ProcRoot: procRoot},
	)
	if err != nil {
		t.Fatalf("NewWorkersModule() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startWorkers(ctx, workerModule)

	select {
	case <-channels.ProcessSnapshots:
	case <-time.After(time.Second):
		t.Fatal("process snapshot was not emitted")
	}
}

func mustRegisterTestHandlers(t *testing.T, store *processstate.SnapshotStore) *metricsruntimetest.RecordingServer {
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

	handler, ok := server.RPCHandlers[subject]
	if !ok {
		t.Fatalf("RPC handler for %q is not registered", subject)
	}

	response, err := handler(context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("RPC handler %q error = %v", subject, err)
	}

	typed, ok := response.(Response)
	if !ok {
		t.Fatalf("RPC response type = %T, want %T", response, *new(Response))
	}

	return typed
}

func createMinimalProcRoot(t *testing.T) string {
	t.Helper()

	procRoot := filepath.Join(t.TempDir(), "proc")
	if err := os.MkdirAll(procRoot, 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", procRoot, err)
	}
	if err := os.WriteFile(filepath.Join(procRoot, "stat"), []byte("btime 100\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(stat) error = %v", err)
	}

	return procRoot
}

func waitForRecordedErrorLog(t *testing.T, log *metricsruntimetest.RecordingLogger) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(log.Errors) == 1 {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatalf("Error logs = %d, want 1", len(log.Errors))
}
