package servicemetrics

import "lite-nas/shared/metrics"

// GetHistoryRequest requests retained service snapshot history.
type GetHistoryRequest struct{}

// GetHistoryResponse returns retained service snapshot history in
// chronological order.
type GetHistoryResponse struct {
	Items []metrics.ServiceMetricsSnapshot `json:"items"`
}
