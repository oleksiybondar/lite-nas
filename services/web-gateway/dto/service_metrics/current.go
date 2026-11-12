package servicemetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// ServiceSnapshotOutput returns the latest service metrics snapshot with the
// common browser-facing response envelope.
type ServiceSnapshotOutput struct {
	Body ServiceSnapshotBody
}

// ServiceSnapshotBody defines the browser-facing snapshot response body.
type ServiceSnapshotBody struct {
	dto.ResponseMeta
	Data metrics.ServiceMetricsSnapshot `json:"data"`
}

// NewSnapshotBody creates the snapshot response body with common metadata set.
func NewSnapshotBody(now time.Time, data metrics.ServiceMetricsSnapshot) ServiceSnapshotBody {
	return ServiceSnapshotBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
