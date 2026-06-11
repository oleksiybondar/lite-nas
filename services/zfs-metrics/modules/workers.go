package modules

import (
	serviceconfig "lite-nas/services/zfs-metrics/config"
	"lite-nas/services/zfs-metrics/workers"
	sharedworkers "lite-nas/shared/workers"
)

// Workers groups worker instances used by the service runtime.
type Workers struct {
	Timer   sharedworkers.TimerWorker
	Polling workers.PollingWorker
}

// NewWorkersModule assembles workers required by the zfs-metrics pipeline.
func NewWorkersModule(
	cfg serviceconfig.MetricsConfig,
	channels Channels,
) (Workers, error) {
	timerWorker, pollTicks, err := sharedworkers.NewPollingTimerWorker(cfg.PollInterval, 1)
	if err != nil {
		return Workers{}, err
	}

	return Workers{
		Timer:   timerWorker,
		Polling: workers.NewPollingWorker(cfg.ZpoolPath, pollTicks, channels.ZFSSnapshots, channels.PollErrors),
	}, nil
}
