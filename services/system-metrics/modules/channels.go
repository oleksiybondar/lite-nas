package modules

import "lite-nas/shared/metrics"

// Channels groups the runtime channels used by the metrics pipeline.
//
// The fields are populated once during startup and are expected to be treated
// as logical wiring owned by the runtime after construction.
type Channels struct {
	RawSnapshots    chan metrics.RawSystemSnapshot
	SystemSnapshots chan metrics.SystemSnapshot
}

// NewChannelsModule allocates the channels used between polling, processing,
// and publishing stages.
//
// Parameters:
//   - bufferSize: channel capacity used by both pipeline stages
func NewChannelsModule(bufferSize int) Channels {
	return Channels{
		RawSnapshots:    make(chan metrics.RawSystemSnapshot, bufferSize),
		SystemSnapshots: make(chan metrics.SystemSnapshot, bufferSize),
	}
}
