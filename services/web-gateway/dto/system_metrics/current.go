package systemmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// SnapshotOutput returns the latest system metrics snapshot with the common
// browser-facing response envelope.
type SnapshotOutput struct {
	Body SnapshotBody
}

// SnapshotBody defines the browser-facing snapshot response body.
type SnapshotBody struct {
	dto.ResponseMeta
	Data metrics.SystemSnapshot `json:"data"`
}

// NewSnapshotBody creates the snapshot response body with common metadata set.
func NewSnapshotBody(now time.Time, data metrics.SystemSnapshot) SnapshotBody {
	return SnapshotBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
