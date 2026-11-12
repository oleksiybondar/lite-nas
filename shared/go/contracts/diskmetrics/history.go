package diskmetrics

import "lite-nas/shared/metrics"

// GetHistoryRequest requests retained disk snapshot history.
type GetHistoryRequest struct{}

// GetHistoryResponse returns retained disk snapshot history in chronological
// order.
type GetHistoryResponse struct {
	Items []metrics.DiskMetricsSnapshot `json:"items"`
}
