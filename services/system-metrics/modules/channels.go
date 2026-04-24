package modules

import "lite-nas/shared/metrics"

// Channels groups the runtime channels used by the metrics pipeline.
type Channels struct {
	rawSnapshots    chan metrics.RawSystemSnapshot
	systemSnapshots chan metrics.SystemSnapshot
}

// NewChannelsModule creates the service channels with the provided buffer size.
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		rawSnapshots:    make(chan metrics.RawSystemSnapshot, bufferSize),
		systemSnapshots: make(chan metrics.SystemSnapshot, bufferSize),
	}
}

// RawSnapshots returns the raw snapshot channel.
func (m Channels) RawSnapshots() <-chan metrics.RawSystemSnapshot {
	return m.rawSnapshots
}

// SystemSnapshots returns the processed snapshot channel.
func (m Channels) SystemSnapshots() <-chan metrics.SystemSnapshot {
	return m.systemSnapshots
}
