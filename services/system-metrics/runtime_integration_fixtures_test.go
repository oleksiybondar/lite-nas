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
	"lite-nas/shared/metrics"
)

type serviceCycleResult struct {
	client *recordingClient
	server *recordingServer
	store  *modules.SnapshotStore
}

func runServiceCycleFixture(t *testing.T) serviceCycleResult {
	t.Helper()

	cpuPath, memPath := createMetricsFilesFixture(t)
	writeMetricsFixtureFiles(t, cpuPath, memPath, baselineCPUFixture(), baselineMemFixture())

	channels := modules.NewChannelsModule(4)
	ioModule, err := modules.NewIOModule(cpuPath, memPath)
	if err != nil {
		t.Fatalf("NewIOModule() error = %v", err)
	}

	workerModule := modules.NewWorkersModule(
		serviceconfig.MetricsConfig{PollInterval: 5 * time.Millisecond, HistorySize: 4},
		channels,
		ioModule,
	)
	stateModule := modules.NewStateModule(4)
	server := &recordingServer{}
	if err := registerRPCHandlers(server, stateModule.SnapshotStore()); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &recordingClient{publishHook: cancel}
	startWorkers(ctx, workerModule)
	go updateMetricsFixtureFilesAfterDelay(t, 10*time.Millisecond, cpuPath, memPath, updatedCPUFixture(), updatedMemFixture())

	err = serveSnapshots(ctx, channels.SystemSnapshots(), stateModule.SnapshotStore(), client, &recordingLogger{})
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v", err)
	}

	return serviceCycleResult{
		client: client,
		server: server,
		store:  stateModule.SnapshotStore(),
	}
}

func createMetricsFilesFixture(t *testing.T) (string, string) {
	t.Helper()

	baseDir := t.TempDir()
	return filepath.Join(baseDir, "stat"), filepath.Join(baseDir, "meminfo")
}

func writeMetricsFixtureFiles(t *testing.T, cpuPath string, memPath string, cpuData string, memData string) {
	t.Helper()

	if err := os.WriteFile(cpuPath, []byte(cpuData), 0o600); err != nil {
		t.Fatalf("WriteFile(cpu) error = %v", err)
	}

	if err := os.WriteFile(memPath, []byte(memData), 0o600); err != nil {
		t.Fatalf("WriteFile(mem) error = %v", err)
	}
}

func updateMetricsFixtureFilesAfterDelay(
	t *testing.T,
	delay time.Duration,
	cpuPath string,
	memPath string,
	cpuData string,
	memData string,
) {
	t.Helper()

	time.Sleep(delay)
	writeMetricsFixtureFiles(t, cpuPath, memPath, cpuData, memData)
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

	snapshot, ok := client.publishCalls[0].payload.(metrics.SystemSnapshot)
	if !ok {
		t.Fatalf("publish payload type = %T, want metrics.SystemSnapshot", client.publishCalls[0].payload)
	}

	return snapshot
}
