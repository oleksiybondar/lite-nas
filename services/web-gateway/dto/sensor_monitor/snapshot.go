// Package sensormonitor defines browser-facing sensor monitor responses.
package sensormonitor

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	"lite-nas/shared/metrics"
)

// SensorSnapshotOutput wraps the latest sensor snapshot.
type SensorSnapshotOutput struct{ Body SensorSnapshotBody }

// SensorSnapshotBody is the browser-facing sensor snapshot envelope.
type SensorSnapshotBody struct {
	dto.ResponseMeta
	Data metrics.SensorSnapshot `json:"data"`
}

// NewSensorSnapshotBody builds a successful snapshot response.
func NewSensorSnapshotBody(now time.Time, data metrics.SensorSnapshot) SensorSnapshotBody {
	return SensorSnapshotBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: now}, Data: data}
}
