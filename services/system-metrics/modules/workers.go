package modules

import (
	serviceconfig "lite-nas/services/system-metrics/config"
	"lite-nas/services/system-metrics/workers"
)

// Workers groups the worker instances used by the service runtime.
type Workers struct {
	polling    workers.PollingWorker
	processing workers.ProcessingWorker
}

// NewWorkersModule creates the workers required by the metrics pipeline.
func NewWorkersModule(
	cfg serviceconfig.MetricsConfig,
	channels Channels,
	io IO,
) Workers {
	return Workers{
		polling: workers.NewPollingWorker(cfg, io.CPUReader(), io.MemReader(), channels.rawSnapshots),
		processing: workers.NewProcessingWorker(
			channels.rawSnapshots,
			channels.systemSnapshots,
		),
	}
}

// Polling returns the polling worker.
func (m Workers) Polling() workers.PollingWorker {
	return m.polling
}

// Processing returns the processing worker.
func (m Workers) Processing() workers.ProcessingWorker {
	return m.processing
}
