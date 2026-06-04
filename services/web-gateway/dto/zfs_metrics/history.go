package zfsmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// ZFSHistoryOutput returns the stored ZFS metrics history with the common
// browser-facing response envelope.
type ZFSHistoryOutput struct {
	Body ZFSHistoryBody
}

// ZFSHistoryBody defines the browser-facing history response body.
type ZFSHistoryBody struct {
	dto.ResponseMeta
	Data []metrics.ZFSSnapshot `json:"data"`
}

// NewHistoryBody creates the history response body with common metadata set.
func NewHistoryBody(now time.Time, data []metrics.ZFSSnapshot) ZFSHistoryBody {
	return ZFSHistoryBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
