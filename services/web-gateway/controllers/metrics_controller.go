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

// newMetricsController wires the shared controller behavior for snapshot and
// history endpoints while letting callers provide DTO mappers and endpoint-
// specific error messages.
func newMetricsController[T any, SnapshotOutput any, HistoryOutput any](
	getSnapshot func(context.Context) (T, error),
	getHistory func(context.Context) ([]T, error),
	buildSnapshotOutput func(time.Time, T) SnapshotOutput,
	buildHistoryOutput func(time.Time, []T) HistoryOutput,
	snapshotErrorMessage string,
	historyErrorMessage string,
) metricsController[T, SnapshotOutput, HistoryOutput] {
	return metricsController[T, SnapshotOutput, HistoryOutput]{
		getSnapshot:          getSnapshot,
		getHistory:           getHistory,
		buildSnapshotOutput:  buildSnapshotOutput,
		buildHistoryOutput:   buildHistoryOutput,
		snapshotErrorMessage: snapshotErrorMessage,
		historyErrorMessage:  historyErrorMessage,
	}
}

// NewSystemMetricsController creates a SystemMetricsController.
//
// Parameters:
//   - service: backend-facing system metrics service used by the controller
func NewSystemMetricsController(service SystemMetricsService) SystemMetricsController {
	return newMetricsController[
		metrics.SystemSnapshot,
		systemmetricsdto.SnapshotOutput,
		systemmetricsdto.HistoryOutput,
	](
		service.GetSnapshot,
		service.GetHistory,
		func(now time.Time, snapshot metrics.SystemSnapshot) systemmetricsdto.SnapshotOutput {
			return systemmetricsdto.SnapshotOutput{Body: systemmetricsdto.NewSnapshotBody(now, snapshot)}
		},
		func(now time.Time, history []metrics.SystemSnapshot) systemmetricsdto.HistoryOutput {
			return systemmetricsdto.HistoryOutput{Body: systemmetricsdto.NewHistoryBody(now, history)}
		},
		"failed to fetch latest system metrics snapshot",
		"failed to fetch system metrics history",
	)
}

// NewZFSMetricsController creates a ZFSMetricsController.
//
// Parameters:
//   - service: backend-facing ZFS metrics service used by the controller
func NewZFSMetricsController(service ZFSMetricsService) ZFSMetricsController {
	return newMetricsController[
		metrics.ZFSSnapshot,
		zfsmetricsdto.ZFSSnapshotOutput,
		zfsmetricsdto.ZFSHistoryOutput,
	](
		service.GetSnapshot,
		service.GetHistory,
		func(now time.Time, snapshot metrics.ZFSSnapshot) zfsmetricsdto.ZFSSnapshotOutput {
			return zfsmetricsdto.ZFSSnapshotOutput{Body: zfsmetricsdto.NewSnapshotBody(now, snapshot)}
		},
		func(now time.Time, history []metrics.ZFSSnapshot) zfsmetricsdto.ZFSHistoryOutput {
			return zfsmetricsdto.ZFSHistoryOutput{Body: zfsmetricsdto.NewHistoryBody(now, history)}
		},
		"failed to fetch latest ZFS metrics snapshot",
		"failed to fetch ZFS metrics history",
	)
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
