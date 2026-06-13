package diskmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// DiskSnapshotOutput returns the latest disk metrics snapshot with the common
// browser-facing response envelope.
type DiskSnapshotOutput struct {
	Body DiskSnapshotBody
}

// DiskSnapshotBody defines the browser-facing snapshot response body.
type DiskSnapshotBody struct {
	dto.ResponseMeta
	Data metrics.DiskMetricsSnapshot `json:"data"`
}

// NewSnapshotBody creates the snapshot response body with common metadata set.
func NewSnapshotBody(now time.Time, data metrics.DiskMetricsSnapshot) DiskSnapshotBody {
	return DiskSnapshotBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
