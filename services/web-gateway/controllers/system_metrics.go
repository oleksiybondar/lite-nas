package controllers

import (
	"context"
	"time"

	systemmetricsdto "lite-nas/services/web-gateway/dto/system_metrics"
	"lite-nas/shared/metrics"

	"github.com/danielgtaylor/huma/v2"
)

// SystemMetricsService defines the system metrics behavior required by the
// browser-facing controller.
type SystemMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.SystemSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.SystemSnapshot, error)
}

// SystemMetricsController exposes browser-facing system metrics endpoints.
type SystemMetricsController struct {
	service SystemMetricsService
}

// NewSystemMetricsController creates a SystemMetricsController.
//
// Parameters:
//   - service: backend-facing system metrics service used by the controller
func NewSystemMetricsController(service SystemMetricsService) SystemMetricsController {
	return SystemMetricsController{service: service}
}

// GetSnapshot returns the latest system metrics snapshot as a browser-facing
// DTO payload.
func (c SystemMetricsController) GetSnapshot(
	ctx context.Context,
	_ *struct{},
) (*systemmetricsdto.SnapshotOutput, error) {
	now := time.Now()
	snapshot, err := c.service.GetSnapshot(ctx)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to fetch latest system metrics snapshot")
	}

	return &systemmetricsdto.SnapshotOutput{Body: systemmetricsdto.NewSnapshotBody(now, snapshot)}, nil
}

// GetHistory returns the stored system metrics history as a browser-facing DTO
// payload.
func (c SystemMetricsController) GetHistory(
	ctx context.Context,
	_ *struct{},
) (*systemmetricsdto.HistoryOutput, error) {
	now := time.Now()
	history, err := c.service.GetHistory(ctx)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to fetch system metrics history")
	}

	return &systemmetricsdto.HistoryOutput{Body: systemmetricsdto.NewHistoryBody(now, history)}, nil
}
