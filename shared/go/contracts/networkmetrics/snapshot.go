package networkmetrics

import "lite-nas/shared/metrics"

// GetSnapshotRequest requests the latest processed network snapshot.
type GetSnapshotRequest struct{}

// GetSnapshotResponse returns the latest processed network snapshot when
// available.
type GetSnapshotResponse struct {
	Available bool                           `json:"available"`
	Snapshot  metrics.NetworkMetricsSnapshot `json:"snapshot,omitempty"`
}

// SnapshotUpdatedEvent publishes the latest processed network snapshot.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.NetworkMetricsSnapshot `json:"snapshot"`
}
