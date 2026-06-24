package controllers

import (
	"context"
	"time"

	processmetricsdto "lite-nas/services/web-gateway/dto/process_metrics"
	"lite-nas/shared/metrics"
)

// ProcessMetricsService defines the process metrics behavior required by the
// browser-facing controller.
type ProcessMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.ProcessMetricsSnapshot, error)
}

// ProcessMetricsController exposes browser-facing process metrics endpoints.
type ProcessMetricsController struct {
	getSnapshot func(context.Context) (metrics.ProcessMetricsSnapshot, error)
}

// NewProcessMetricsController creates a ProcessMetricsController.
func NewProcessMetricsController(service ProcessMetricsService) ProcessMetricsController {
	return ProcessMetricsController{
		getSnapshot: service.GetSnapshot,
	}
}

// GetSnapshot returns the latest process metrics snapshot as a browser-facing
// DTO payload.
func (c ProcessMetricsController) GetSnapshot(
	ctx context.Context,
	_ *struct{},
) (*processmetricsdto.ProcessSnapshotOutput, error) {
	return fetchSnapshotOutput(
		ctx,
		c.getSnapshot,
		func(now time.Time, snapshot metrics.ProcessMetricsSnapshot) processmetricsdto.ProcessSnapshotOutput {
			return processmetricsdto.ProcessSnapshotOutput{Body: processmetricsdto.NewSnapshotBody(now, snapshot)}
		},
		"failed to fetch latest process metrics snapshot",
	)
}
