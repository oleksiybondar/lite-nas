package networkmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// NetworkSnapshotOutput returns the latest network metrics snapshot with the common
// browser-facing response envelope.
type NetworkSnapshotOutput struct {
	Body NetworkSnapshotBody
}

// NetworkSnapshotBody defines the browser-facing snapshot response body.
type NetworkSnapshotBody struct {
	dto.ResponseMeta
	Data metrics.NetworkMetricsSnapshot `json:"data"`
}

// NewSnapshotBody creates the snapshot response body with common metadata set.
func NewSnapshotBody(now time.Time, data metrics.NetworkMetricsSnapshot) NetworkSnapshotBody {
	return NetworkSnapshotBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
