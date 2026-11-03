package servicemetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// ServiceHistoryOutput returns the stored service metrics history with the
// common browser-facing response envelope.
type ServiceHistoryOutput struct {
	Body ServiceHistoryBody
}

// ServiceHistoryBody defines the browser-facing history response body.
type ServiceHistoryBody struct {
	dto.ResponseMeta
	Data []metrics.ServiceMetricsSnapshot `json:"data"`
}

// NewHistoryBody creates the history response body with common metadata set.
func NewHistoryBody(now time.Time, data []metrics.ServiceMetricsSnapshot) ServiceHistoryBody {
	return ServiceHistoryBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
