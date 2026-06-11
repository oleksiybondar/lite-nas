package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	serviceconfig "lite-nas/services/system-metrics/config"
	"lite-nas/services/system-metrics/modules"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/metrics"
)

type serviceCycleResult struct {
	client *recordingClient
	server *recordingServer
	store  *modules.SnapshotStore
}

func runServiceCycleFixture(t *testing.T) serviceCycleResult {
	t.Helper()

	channels, workerModule, stateModule, cpuPath, memPath := prepareServiceCycleModulesFixture(t)
	server := &recordingServer{}
	if err := registerRPCHandlers(server, stateModule.SnapshotStore); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &recordingClient{publishHook: cancel}
	startWorkers(ctx, workerModule)
	updateDone := scheduleMetricsUpdateFixture(cpuPath, memPath)

	err := serveSnapshots(ctx, channels.SystemSnapshots, stateModule.SnapshotStore, client, &recordingLogger{})
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v", err)
	}
	mustCompleteMetricsUpdateFixture(t, updateDone)

	return serviceCycleResult{
		client: client,
		server: server,
		store:  stateModule.SnapshotStore,
	}
}

func prepareServiceCycleModulesFixture(
	t *testing.T,
) (modules.Channels, modules.Workers, modules.State, string, string) {
	t.Helper()

	cpuPath, memPath := createMetricsFilesFixture(t)
	if err := writeMetricsFixtureFiles(cpuPath, memPath, baselineCPUFixture(), baselineMemFixture()); err != nil {
		t.Fatalf("writeMetricsFixtureFiles() error = %v", err)
	}

	channels := modules.NewChannelsModule(4)
	ioModule, err := modules.NewIOModule(cpuPath, memPath)
	if err != nil {
		t.Fatalf("NewIOModule() error = %v", err)
	}

	workerModule, err := modules.NewWorkersModule(
		serviceconfig.MetricsConfig{PollInterval: 5 * time.Millisecond, HistorySize: 4},
		channels,
		ioModule,
	)
	if err != nil {
		t.Fatalf("NewWorkersModule() error = %v", err)
	}

	return channels, workerModule, modules.NewStateModule(4), cpuPath, memPath
}

func scheduleMetricsUpdateFixture(cpuPath string, memPath string) <-chan error {
	updateDone := make(chan error, 1)

	go updateMetricsFixtureFilesAfterDelay(
		updateDone,
		10*time.Millisecond,
		cpuPath,
		memPath,
		updatedCPUFixture(),
		updatedMemFixture(),
	)

	return updateDone
}

func createMetricsFilesFixture(t *testing.T) (string, string) {
	t.Helper()

	baseDir := t.TempDir()
	return filepath.Join(baseDir, "stat"), filepath.Join(baseDir, "meminfo")
}

func writeMetricsFixtureFiles(cpuPath string, memPath string, cpuData string, memData string) error {
	if err := os.WriteFile(cpuPath, []byte(cpuData), 0o600); err != nil {
		return err
	}

	if err := os.WriteFile(memPath, []byte(memData), 0o600); err != nil {
		return err
	}

	return nil
}

func updateMetricsFixtureFilesAfterDelay(
	updateDone chan<- error,
	delay time.Duration,
	cpuPath string,
	memPath string,
	cpuData string,
	memData string,
) {
	time.Sleep(delay)
	updateDone <- writeMetricsFixtureFiles(cpuPath, memPath, cpuData, memData)
}

func mustCompleteMetricsUpdateFixture(t *testing.T, updateDone <-chan error) {
	t.Helper()

	if err := <-updateDone; err != nil {
		t.Fatalf("writeMetricsFixtureFiles() error = %v", err)
	}
}

func baselineCPUFixture() string {
	return "cpu  10 20 30 40 5 6 7 8 0 0\ncpu0 1 2 3 4 1 0 0 0 0 0\n"
}

func updatedCPUFixture() string {
	return "cpu  20 30 40 50 5 6 7 8 0 0\ncpu0 2 4 6 5 1 0 0 0 0 0\n"
}

func baselineMemFixture() string {
	return "MemTotal: 1000 kB\nMemAvailable: 250 kB\n"
}

func updatedMemFixture() string {
	return "MemTotal: 1000 kB\nMemAvailable: 200 kB\n"
}

func extractPublishedSnapshot(t *testing.T, client *recordingClient) metrics.SystemSnapshot {
	t.Helper()

	if len(client.publishCalls) != 1 {
		t.Fatalf("publishCalls = %d, want 1", len(client.publishCalls))
	}

	event, ok := client.publishCalls[0].payload.(systemmetricscontract.SnapshotUpdatedEvent)
	if !ok {
		t.Fatalf("publish payload type = %T, want systemmetrics.SnapshotUpdatedEvent", client.publishCalls[0].payload)
	}

	return event.Snapshot
}
