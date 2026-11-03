package networkmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// NetworkHistoryOutput returns the stored network metrics history with the common
// browser-facing response envelope.
type NetworkHistoryOutput struct {
	Body NetworkHistoryBody
}

// NetworkHistoryBody defines the browser-facing history response body.
type NetworkHistoryBody struct {
	dto.ResponseMeta
	Data []metrics.NetworkMetricsSnapshot `json:"data"`
}

// NewHistoryBody creates the history response body with common metadata set.
func NewHistoryBody(now time.Time, data []metrics.NetworkMetricsSnapshot) NetworkHistoryBody {
	return NetworkHistoryBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
