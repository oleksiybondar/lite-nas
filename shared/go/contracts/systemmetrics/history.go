package systemmetrics

import "lite-nas/shared/metrics"

// GetHistoryRequest requests the retained system metrics history.
type GetHistoryRequest struct{}

// GetHistoryResponse returns the retained system metrics history in
// chronological order.
type GetHistoryResponse struct {
	Items []metrics.SystemSnapshot `json:"items"`
}
