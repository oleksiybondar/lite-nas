package networkmetrics

import "lite-nas/shared/metrics"

// GetHistoryRequest requests retained network snapshot history.
type GetHistoryRequest struct{}

// GetHistoryResponse returns retained network snapshot history in
// chronological order.
type GetHistoryResponse struct {
	Items []metrics.NetworkMetricsSnapshot `json:"items"`
}
