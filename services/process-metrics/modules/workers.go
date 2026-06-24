package modules

import (
	processconfig "lite-nas/services/process-metrics/config"
	"lite-nas/services/process-metrics/workers"
	sharedworkers "lite-nas/shared/workers"
)

// SourcePaths groups filesystem roots used by the polling worker.
type SourcePaths struct {
	ProcRoot string
}

// Workers groups worker instances used by the service runtime.
type Workers struct {
	Timer   sharedworkers.TimerWorker
	Polling workers.PollingWorker
}

// NewWorkersModule assembles workers required by the process-metrics runtime.
func NewWorkersModule(
	cfg processconfig.ProcessMetricsConfig,
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
			paths.ProcRoot,
			pollTicks,
			channels.ProcessSnapshots,
			channels.PollErrors,
		),
	}, nil
}
