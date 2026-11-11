// Package sensormonitor defines the NATS contracts owned by Sensor Monitor.
package sensormonitor

import "lite-nas/shared/metrics"

const (
	// SnapshotEventSubject publishes each completed sensor snapshot.
	SnapshotEventSubject = "sensor.monitor.events.snapshot"
	// SnapshotRPCSubject serves the latest sensor snapshot.
	SnapshotRPCSubject = "sensor.monitor.rpc.snapshot.get"
	// HistoryRPCSubject serves retained sensor snapshots.
	HistoryRPCSubject = "sensor.monitor.rpc.history.get"
)

// GetSnapshotRequest requests the latest snapshot without parameters.
type GetSnapshotRequest struct{}

// GetSnapshotResponse reports the latest snapshot when available.
type GetSnapshotResponse struct {
	Available bool                   `json:"available"`
	Snapshot  metrics.SensorSnapshot `json:"snapshot,omitempty"`
}

// GetHistoryRequest requests the bounded history without parameters.
type GetHistoryRequest struct{}

// GetHistoryResponse returns retained snapshots in chronological order.
type GetHistoryResponse struct {
	Items []metrics.SensorSnapshot `json:"items"`
}

// SnapshotUpdatedEvent announces one completed collection cycle.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.SensorSnapshot `json:"snapshot" validate:"required"`
}
