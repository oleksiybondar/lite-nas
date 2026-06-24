package modules

import "lite-nas/shared/metrics"

// Channels groups runtime channels used by the process-metrics pipeline.
type Channels struct {
	ProcessSnapshots chan metrics.ProcessMetricsSnapshot
	PollErrors       chan error
}

// NewChannelsModule allocates pipeline channels.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		ProcessSnapshots: make(chan metrics.ProcessMetricsSnapshot, bufferSize),
		PollErrors:       make(chan error, bufferSize),
	}
}
