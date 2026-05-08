package systemmetrics

import "lite-nas/shared/metrics"

// GetSnapshotRequest requests the latest processed system metrics snapshot.
type GetSnapshotRequest struct{}

// GetSnapshotResponse returns the latest processed system metrics snapshot when
// one is available.
type GetSnapshotResponse struct {
	Available bool                   `json:"available"`
	Snapshot  metrics.SystemSnapshot `json:"snapshot,omitempty"`
}

// SnapshotUpdatedEvent publishes the latest processed system metrics snapshot
// to subscribers.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.SystemSnapshot `json:"snapshot"`
}
