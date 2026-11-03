package diskmetrics

import "lite-nas/shared/metrics"

// GetSnapshotRequest requests the latest processed disk snapshot.
type GetSnapshotRequest struct{}

// GetSnapshotResponse returns the latest processed disk snapshot when
// available.
type GetSnapshotResponse struct {
	Available bool                        `json:"available"`
	Snapshot  metrics.DiskMetricsSnapshot `json:"snapshot,omitempty"`
}

// SnapshotUpdatedEvent publishes the latest processed disk snapshot.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.DiskMetricsSnapshot `json:"snapshot"`
}
