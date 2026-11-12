package diskmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// DiskHistoryOutput returns the stored disk metrics history with the common
// browser-facing response envelope.
type DiskHistoryOutput struct {
	Body DiskHistoryBody
}

// DiskHistoryBody defines the browser-facing history response body.
type DiskHistoryBody struct {
	dto.ResponseMeta
	Data []metrics.DiskMetricsSnapshot `json:"data"`
}

// NewHistoryBody creates the history response body with common metadata set.
func NewHistoryBody(now time.Time, data []metrics.DiskMetricsSnapshot) DiskHistoryBody {
	return DiskHistoryBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
