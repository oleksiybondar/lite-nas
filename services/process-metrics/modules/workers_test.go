package modules

import (
	"testing"
	"time"

	processconfig "lite-nas/services/process-metrics/config"
)

func TestNewChannelsModuleAllocatesBufferedChannels(t *testing.T) {
	t.Parallel()

	channels := NewChannelsModule(2)
	if channels.ProcessSnapshots == nil {
		t.Fatal("ProcessSnapshots = nil, want allocated channel")
	}
	if channels.PollErrors == nil {
		t.Fatal("PollErrors = nil, want allocated channel")
	}
	if got := cap(channels.ProcessSnapshots); got != 2 {
		t.Fatalf("cap(ProcessSnapshots) = %d, want 2", got)
	}
	if got := cap(channels.PollErrors); got != 2 {
		t.Fatalf("cap(PollErrors) = %d, want 2", got)
	}
}

func TestNewWorkersModuleReturnsTimerValidationError(t *testing.T) {
	t.Parallel()

	_, err := NewWorkersModule(
		processconfig.ProcessMetricsConfig{
			Enabled:      true,
			PollInterval: 0,
		},
		NewChannelsModule(1),
		SourcePaths{ProcRoot: "/proc"},
	)
	if err == nil {
		t.Fatal("NewWorkersModule() error = nil, want timer validation error")
	}
}

func TestNewWorkersModuleBuildsTimerAndPollingWorkers(t *testing.T) {
	t.Parallel()

	channels := NewChannelsModule(3)
	workerModule, err := NewWorkersModule(
		processconfig.ProcessMetricsConfig{
			Enabled:      true,
			PollInterval: 5 * time.Second,
		},
		channels,
		SourcePaths{ProcRoot: "/proc"},
	)
	if err != nil {
		t.Fatalf("NewWorkersModule() error = %v", err)
	}

	if got := cap(channels.ProcessSnapshots); got != 3 {
		t.Fatalf("cap(ProcessSnapshots) = %d, want 3", got)
	}
	if got := cap(channels.PollErrors); got != 3 {
		t.Fatalf("cap(PollErrors) = %d, want 3", got)
	}
	_ = workerModule
}
