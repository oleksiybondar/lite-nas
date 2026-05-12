package modules

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	serviceconfig "lite-nas/services/system-metrics/config"
)

func TestNewWorkersModuleBuildsPollingAndProcessingWorkers(t *testing.T) {
	t.Parallel()

	cpuPath, memPath := createWorkerFixtureFiles(t)
	channels := NewChannelsModule(1)
	ioModule, err := NewIOModule(cpuPath, memPath)
	if err != nil {
		t.Fatalf("NewIOModule() error = %v", err)
	}

	module, err := NewWorkersModule(serviceconfig.MetricsConfig{PollInterval: time.Second}, channels, ioModule)
	if err != nil {
		t.Fatalf("NewWorkersModule() error = %v", err)
	}

	assertWorkerModule(t, module)
}

func createWorkerFixtureFiles(t *testing.T) (string, string) {
	t.Helper()

	baseDir := t.TempDir()
	cpuPath := filepath.Join(baseDir, "stat")
	memPath := filepath.Join(baseDir, "meminfo")

	if err := os.WriteFile(cpuPath, []byte("cpu  1 1 1 1 0 0 0 0 0 0\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(cpu) error = %v", err)
	}

	if err := os.WriteFile(memPath, []byte("MemTotal: 1 kB\nMemAvailable: 1 kB\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(mem) error = %v", err)
	}

	return cpuPath, memPath
}

func assertWorkerModule(t *testing.T, module Workers) {
	t.Helper()

	if reflect.ValueOf(module.Polling).IsZero() {
		t.Fatal("expected polling worker")
	}
	if reflect.ValueOf(module.Processing).IsZero() {
		t.Fatal("expected processing worker")
	}
	if reflect.ValueOf(module.Timer).IsZero() {
		t.Fatal("expected timer worker")
	}
}
