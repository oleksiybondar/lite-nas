package modules

import (
	serviceconfig "lite-nas/services/disk-metrics/config"
	"lite-nas/services/disk-metrics/workers"
	sharedworkers "lite-nas/shared/workers"
)

// SourcePaths groups filesystem paths used by the polling worker.
type SourcePaths struct {
	SysBlock          string
	ProcDiskStats     string
	ProcSelfMountInfo string
	ProcMounts        string
}

// Workers groups worker instances used by the service runtime.
type Workers struct {
	Timer   sharedworkers.TimerWorker
	Polling workers.PollingWorker
}

// NewWorkersModule assembles workers required by the disk-metrics runtime.
func NewWorkersModule(
	cfg serviceconfig.MetricsConfig,
	channels Channels,
	paths SourcePaths,
) (Workers, error) {
	timerWorker, pollTicks, err := sharedworkers.NewPollingTimerWorker(cfg.PollInterval, 1)
	if err != nil {
		return Workers{}, err
	}

	return Workers{
		Timer: sharedworkers.TimerWorker(timerWorker),
		Polling: workers.NewPollingWorker(
			paths.SysBlock,
			paths.ProcDiskStats,
			paths.ProcSelfMountInfo,
			paths.ProcMounts,
			pollTicks,
			channels.DiskSnapshots,
			channels.PollErrors,
		),
	}, nil
}
