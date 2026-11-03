package modules

import (
	"testing"
	"time"

	serviceconfig "lite-nas/services/disk-metrics/config"
)

func TestNewWorkersModuleBuildsTimerAndPollingWorker(t *testing.T) {
	t.Parallel()

	workers, err := NewWorkersModule(
		serviceconfig.MetricsConfig{PollInterval: time.Second},
		NewChannelsModule(1),
		SourcePaths{
			SysBlock:          "/sys/block",
			ProcDiskStats:     "/proc/diskstats",
			ProcSelfMountInfo: "/proc/self/mountinfo",
			ProcMounts:        "/proc/mounts",
		},
	)
	if err != nil {
		t.Fatalf("NewWorkersModule() error = %v", err)
	}
	_ = workers
}

func TestNewWorkersModuleRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	_, err := NewWorkersModule(serviceconfig.MetricsConfig{}, NewChannelsModule(1), SourcePaths{})
	if err == nil {
		t.Fatal("NewWorkersModule() error = nil, want invalid poll interval error")
	}
}
