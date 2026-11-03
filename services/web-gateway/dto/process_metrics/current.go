package processmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// SnapshotOutput returns the latest process metrics snapshot with the common
// browser-facing response envelope.
type ProcessSnapshotOutput struct {
	Body ProcessSnapshotBody
}

// ProcessSnapshotBody defines the browser-facing snapshot response body.
type ProcessSnapshotBody struct {
	dto.ResponseMeta
	Data metrics.ProcessMetricsSnapshot `json:"data"`
}

// NewSnapshotBody creates the snapshot response body with common metadata set.
func NewSnapshotBody(now time.Time, data metrics.ProcessMetricsSnapshot) ProcessSnapshotBody {
	return ProcessSnapshotBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
