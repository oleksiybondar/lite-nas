package modules

import (
	"testing"
	"time"

	processconfig "lite-nas/services/process-metrics/config"
)

func TestNewChannelsModuleAllocatesBufferedChannels(t *testing.T) {
	t.Parallel()

	channels := NewChannelsModule(2)

	assertChannelAllocatedWithCapacity(t, "ProcessedProcessSnapshots", channels.ProcessedProcessSnapshots, 2)
	assertChannelAllocatedWithCapacity(t, "RawProcessSnapshots", channels.RawProcessSnapshots, 2)
	assertChannelAllocatedWithCapacity(t, "PollErrors", channels.PollErrors, 2)
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

	if got := cap(channels.ProcessedProcessSnapshots); got != 3 {
		t.Fatalf("cap(ProcessedProcessSnapshots) = %d, want 3", got)
	}
	if got := cap(channels.RawProcessSnapshots); got != 3 {
		t.Fatalf("cap(RawProcessSnapshots) = %d, want 3", got)
	}
	if got := cap(channels.PollErrors); got != 3 {
		t.Fatalf("cap(PollErrors) = %d, want 3", got)
	}
	_ = workerModule
}

func assertChannelAllocatedWithCapacity[T any](
	t *testing.T,
	name string,
	channel chan T,
	wantCapacity int,
) {
	t.Helper()

	if channel == nil {
		t.Fatalf("%s = nil, want allocated channel", name)
	}
	if got := cap(channel); got != wantCapacity {
		t.Fatalf("cap(%s) = %d, want %d", name, got, wantCapacity)
	}
}
