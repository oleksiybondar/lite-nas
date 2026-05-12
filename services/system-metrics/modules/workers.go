package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	"lite-nas/services/system-metrics/workers"
	sharedworkers "lite-nas/shared/workers"
)

// Workers groups the worker instances used by the service runtime.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only after construction.
type Workers struct {
	Timer      sharedworkers.TimerWorker
	Polling    workers.PollingWorker
	Processing workers.ProcessingWorker
}

// NewWorkersModule assembles the workers required by the metrics pipeline.
//
// Parameters:
//   - cfg: polling and retention settings used by the workers
//   - channels: runtime pipeline channels shared between workers
//   - io: procfs readers consumed by the polling worker
func NewWorkersModule(
	cfg serviceconfig.MetricsConfig,
	channels Channels,
	io IO,
) (Workers, error) {
	pollTicks := make(chan struct{}, 1)
	timerWorker, err := sharedworkers.NewTimerWorker(
		sharedworkers.TimerConfig{
			Interval:    cfg.PollInterval,
			EmitOnStart: true,
		},
		pollTicks,
	)
	if err != nil {
		return Workers{}, err
	}

	return Workers{
		Timer:   timerWorker,
		Polling: workers.NewPollingWorker(io.CPUReader, io.MemReader, pollTicks, channels.RawSnapshots),
		Processing: workers.NewProcessingWorker(
			channels.RawSnapshots,
			channels.SystemSnapshots,
		),
	}, nil
}
