package zfsmetrics

import "lite-nas/shared/metrics"

// GetHistoryRequest requests retained ZFS snapshot history.
type GetHistoryRequest struct{}

// GetHistoryResponse returns retained ZFS snapshot history in chronological order.
type GetHistoryResponse struct {
	Items []metrics.ZFSSnapshot `json:"items"`
}
