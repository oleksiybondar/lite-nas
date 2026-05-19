package modules

import "lite-nas/shared/metrics"

// Channels groups runtime channels used by the zfs-metrics pipeline.
type Channels struct {
	ZFSSnapshots chan metrics.ZFSSnapshot
	PollErrors   chan error
}

// NewChannelsModule allocates pipeline channels.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		ZFSSnapshots: make(chan metrics.ZFSSnapshot, bufferSize),
		PollErrors:   make(chan error, bufferSize),
	}
}
