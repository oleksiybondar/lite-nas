package zfsmetrics

import "lite-nas/shared/metrics"

// GetSnapshotRequest requests the latest processed ZFS snapshot.
type GetSnapshotRequest struct{}

// GetSnapshotResponse returns the latest processed ZFS snapshot when available.
type GetSnapshotResponse struct {
	Available bool                `json:"available"`
	Snapshot  metrics.ZFSSnapshot `json:"snapshot,omitempty"`
}

// SnapshotUpdatedEvent publishes the latest processed ZFS snapshot.
type SnapshotUpdatedEvent struct {
	Snapshot metrics.ZFSSnapshot `json:"snapshot"`
}
