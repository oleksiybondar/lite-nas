package modules

import "lite-nas/shared/metrics"

// Channels groups runtime channels used by the service-metrics pipeline.
type Channels struct {
	ServiceSnapshots chan metrics.ServiceMetricsSnapshot
	PollErrors       chan error
}

// NewChannelsModule allocates pipeline channels.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		ServiceSnapshots: make(chan metrics.ServiceMetricsSnapshot, bufferSize),
		PollErrors:       make(chan error, bufferSize),
	}
}
