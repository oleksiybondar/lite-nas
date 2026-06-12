package modules

import "lite-nas/shared/metrics"

// Channels groups runtime channels used by the network-metrics pipeline.
type Channels struct {
	NetworkSnapshots chan metrics.NetworkMetricsSnapshot
	PollErrors       chan error
}

// NewChannelsModule allocates pipeline channels.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		NetworkSnapshots: make(chan metrics.NetworkMetricsSnapshot, bufferSize),
		PollErrors:       make(chan error, bufferSize),
	}
}
