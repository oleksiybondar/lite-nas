package controllers

import (
	"context"
	"time"

	systemmetricsdto "lite-nas/services/web-gateway/dto/system_metrics"
	zfsmetricsdto "lite-nas/services/web-gateway/dto/zfs_metrics"
	"lite-nas/shared/metrics"
)

// SystemMetricsService defines the system metrics behavior required by the
// browser-facing controller.
type SystemMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.SystemSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.SystemSnapshot, error)
}

// ZFSMetricsService defines the ZFS metrics behavior required by the
// browser-facing controller.
type ZFSMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.ZFSSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.ZFSSnapshot, error)
}

type metricsController[T any, SnapshotOutput any, HistoryOutput any] struct {
	getSnapshot          func(context.Context) (T, error)
	getHistory           func(context.Context) ([]T, error)
	buildSnapshotOutput  func(time.Time, T) SnapshotOutput
	buildHistoryOutput   func(time.Time, []T) HistoryOutput
	snapshotErrorMessage string
	historyErrorMessage  string
}

// SystemMetricsController exposes browser-facing system metrics endpoints.
type SystemMetricsController = metricsController[
	metrics.SystemSnapshot,
	systemmetricsdto.SnapshotOutput,
	systemmetricsdto.HistoryOutput,
]

// ZFSMetricsController exposes browser-facing ZFS metrics endpoints.
type ZFSMetricsController = metricsController[
	metrics.ZFSSnapshot,
	zfsmetricsdto.ZFSSnapshotOutput,
	zfsmetricsdto.ZFSHistoryOutput,
]

// NewSystemMetricsController creates a SystemMetricsController.
//
// Parameters:
//   - service: backend-facing system metrics service used by the controller
func NewSystemMetricsController(service SystemMetricsService) SystemMetricsController {
	return metricsController[
		metrics.SystemSnapshot,
		systemmetricsdto.SnapshotOutput,
		systemmetricsdto.HistoryOutput,
	]{
		getSnapshot: service.GetSnapshot,
		getHistory:  service.GetHistory,
		buildSnapshotOutput: func(now time.Time, snapshot metrics.SystemSnapshot) systemmetricsdto.SnapshotOutput {
			return systemmetricsdto.SnapshotOutput{Body: systemmetricsdto.NewSnapshotBody(now, snapshot)}
		},
		buildHistoryOutput: func(now time.Time, history []metrics.SystemSnapshot) systemmetricsdto.HistoryOutput {
			return systemmetricsdto.HistoryOutput{Body: systemmetricsdto.NewHistoryBody(now, history)}
		},
		snapshotErrorMessage: "failed to fetch latest system metrics snapshot",
		historyErrorMessage:  "failed to fetch system metrics history",
	}
}

// NewZFSMetricsController creates a ZFSMetricsController.
//
// Parameters:
//   - service: backend-facing ZFS metrics service used by the controller
func NewZFSMetricsController(service ZFSMetricsService) ZFSMetricsController {
	return metricsController[
		metrics.ZFSSnapshot,
		zfsmetricsdto.ZFSSnapshotOutput,
		zfsmetricsdto.ZFSHistoryOutput,
	]{
		getSnapshot: service.GetSnapshot,
		getHistory:  service.GetHistory,
		buildSnapshotOutput: func(now time.Time, snapshot metrics.ZFSSnapshot) zfsmetricsdto.ZFSSnapshotOutput {
			return zfsmetricsdto.ZFSSnapshotOutput{Body: zfsmetricsdto.NewSnapshotBody(now, snapshot)}
		},
		buildHistoryOutput: func(now time.Time, history []metrics.ZFSSnapshot) zfsmetricsdto.ZFSHistoryOutput {
			return zfsmetricsdto.ZFSHistoryOutput{Body: zfsmetricsdto.NewHistoryBody(now, history)}
		},
		snapshotErrorMessage: "failed to fetch latest ZFS metrics snapshot",
		historyErrorMessage:  "failed to fetch ZFS metrics history",
	}
}

// GetSnapshot returns the latest metrics snapshot as a browser-facing DTO
// payload.
func (c metricsController[T, SnapshotOutput, HistoryOutput]) GetSnapshot(
	ctx context.Context,
	_ *struct{},
) (*SnapshotOutput, error) {
	return fetchSnapshotOutput(
		ctx,
		c.getSnapshot,
		c.buildSnapshotOutput,
		c.snapshotErrorMessage,
	)
}

// GetHistory returns the stored metrics history as a browser-facing DTO
// payload.
func (c metricsController[T, SnapshotOutput, HistoryOutput]) GetHistory(
	ctx context.Context,
	_ *struct{},
) (*HistoryOutput, error) {
	return fetchHistoryOutput(
		ctx,
		c.getHistory,
		c.buildHistoryOutput,
		c.historyErrorMessage,
	)
}
