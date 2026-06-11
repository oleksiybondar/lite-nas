package zfsmetrics

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// ZFSSnapshotOutput returns the latest ZFS metrics snapshot with the common
// browser-facing response envelope.
type ZFSSnapshotOutput struct {
	Body ZFSSnapshotBody
}

// ZFSSnapshotBody defines the browser-facing snapshot response body.
type ZFSSnapshotBody struct {
	dto.ResponseMeta
	Data metrics.ZFSSnapshot `json:"data"`
}

// NewSnapshotBody creates the snapshot response body with common metadata set.
func NewSnapshotBody(now time.Time, data metrics.ZFSSnapshot) ZFSSnapshotBody {
	return ZFSSnapshotBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
