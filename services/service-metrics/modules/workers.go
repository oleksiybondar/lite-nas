package modules

import (
	serviceconfig "lite-nas/services/service-metrics/config"
	"lite-nas/services/service-metrics/workers"
	sharedworkers "lite-nas/shared/workers"
)

// SourcePaths groups filesystem roots used by the polling worker.
type SourcePaths struct {
	UnitRoots      []string
	SystemSliceDir string
}

// Workers groups worker instances used by the service runtime.
type Workers struct {
	Timer   sharedworkers.TimerWorker
	Polling workers.PollingWorker
}

// NewWorkersModule assembles workers required by the service-metrics runtime.
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
			paths.UnitRoots,
			paths.SystemSliceDir,
			pollTicks,
			channels.ServiceSnapshots,
			channels.PollErrors,
		),
	}, nil
}
