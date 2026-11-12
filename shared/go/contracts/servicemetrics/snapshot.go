package servicemetrics

import "lite-nas/shared/metrics"

// GetSnapshotRequest requests the latest processed service snapshot.
type GetSnapshotRequest struct{}

// GetSnapshotResponse returns the latest processed service snapshot when
// available.
type GetSnapshotResponse struct {
	Available bool                           `json:"available"`
	Snapshot  metrics.ServiceMetricsSnapshot `json:"snapshot,omitempty"`
}

// SnapshotUpdatedEvent publishes the latest processed service snapshot.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.ServiceMetricsSnapshot `json:"snapshot"`
}
