package sensormonitor

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// SensorHistoryOutput wraps retained sensor snapshots.
type SensorHistoryOutput struct{ Body SensorHistoryBody }

// SensorHistoryBody is the browser-facing sensor history envelope.
type SensorHistoryBody struct {
	dto.ResponseMeta
	Data []metrics.SensorSnapshot `json:"data"`
}

// NewSensorHistoryBody builds a successful history response.
func NewSensorHistoryBody(now time.Time, data []metrics.SensorSnapshot) SensorHistoryBody {
	return SensorHistoryBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: now}, Data: data}
}
