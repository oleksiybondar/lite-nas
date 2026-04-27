package systemmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// HistoryOutput returns the stored system metrics history with the common
// browser-facing response envelope.
type HistoryOutput struct {
	Body HistoryBody
}

// HistoryBody defines the browser-facing history response body.
type HistoryBody struct {
	dto.ResponseMeta
	Data []metrics.SystemSnapshot `json:"data"`
}

// NewHistoryBody creates the history response body with common metadata set.
func NewHistoryBody(now time.Time, data []metrics.SystemSnapshot) HistoryBody {
	return HistoryBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
