package modules

import "lite-nas/shared/metrics"

// Channels groups runtime channels used by the disk-metrics pipeline.
type Channels struct {
	DiskSnapshots chan metrics.DiskMetricsSnapshot
	PollErrors    chan error
}

// NewChannelsModule allocates pipeline channels.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		DiskSnapshots: make(chan metrics.DiskMetricsSnapshot, bufferSize),
		PollErrors:    make(chan error, bufferSize),
	}
}
