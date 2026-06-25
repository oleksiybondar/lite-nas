package modules

import "lite-nas/shared/metrics"

// Channels groups runtime channels used by the process-metrics pipeline.
type Channels struct {
	ProcessedProcessSnapshots chan metrics.ProcessMetricsSnapshot
	RawProcessSnapshots       chan metrics.ProcessMetricsSnapshot
	PollErrors                chan error
}

// NewChannelsModule allocates pipeline channels.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		ProcessedProcessSnapshots: make(chan metrics.ProcessMetricsSnapshot, bufferSize),
		RawProcessSnapshots:       make(chan metrics.ProcessMetricsSnapshot, bufferSize),
		PollErrors:                make(chan error, bufferSize),
	}
}
