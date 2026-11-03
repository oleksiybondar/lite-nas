package processmetrics

import "lite-nas/shared/metrics"

// GetSnapshotRequest requests the latest collected process metrics snapshot.
type GetSnapshotRequest struct{}

// GetSnapshotResponse returns the latest collected process metrics snapshot
// when one is available.
type GetSnapshotResponse struct {
	Available bool                           `json:"available"`
	Snapshot  metrics.ProcessMetricsSnapshot `json:"snapshot,omitempty"`
}

// SnapshotUpdatedEvent publishes the latest collected process metrics
// snapshot to subscribers.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.ProcessMetricsSnapshot `json:"snapshot"`
}
