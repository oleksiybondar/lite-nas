package services

import (
	"context"

	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

// SystemMetricsService defines the backend-facing system metrics flows used by
// the gateway service layer.
type SystemMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.SystemSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.SystemSnapshot, error)
}

// ZFSMetricsService defines the backend-facing ZFS metrics flows used by the
// gateway service layer.
type ZFSMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.ZFSSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.ZFSSnapshot, error)
}

type metricsRPCService[T any, SnapshotResponse any, HistoryResponse any] struct {
	client             messaging.Client
	snapshotSubject    string
	historySubject     string
	snapshotRequest    any
	historyRequest     any
	selectSnapshot     func(SnapshotResponse) T
	selectHistoryItems func(HistoryResponse) []T
}

// NewSystemMetricsService creates a service that fetches system metrics over
// the shared messaging transport.
//
// Parameters:
//   - client: messaging client used for request/reply RPC calls
func NewSystemMetricsService(client messaging.Client) SystemMetricsService {
	return metricsRPCService[
		metrics.SystemSnapshot,
		systemmetricscontract.GetSnapshotResponse,
		systemmetricscontract.GetHistoryResponse,
	]{
		client:          client,
		snapshotSubject: systemmetricscontract.SnapshotRPCSubject,
		historySubject:  systemmetricscontract.HistoryRPCSubject,
		snapshotRequest: systemmetricscontract.GetSnapshotRequest{},
		historyRequest:  systemmetricscontract.GetHistoryRequest{},
		selectSnapshot: func(response systemmetricscontract.GetSnapshotResponse) metrics.SystemSnapshot {
			return response.Snapshot
		},
		selectHistoryItems: func(response systemmetricscontract.GetHistoryResponse) []metrics.SystemSnapshot {
			return response.Items
		},
	}
}

// NewZFSMetricsService creates a service that fetches ZFS metrics over the
// shared messaging transport.
//
// Parameters:
//   - client: messaging client used for request/reply RPC calls
func NewZFSMetricsService(client messaging.Client) ZFSMetricsService {
	return metricsRPCService[
		metrics.ZFSSnapshot,
		zfsmetricscontract.GetSnapshotResponse,
		zfsmetricscontract.GetHistoryResponse,
	]{
		client:          client,
		snapshotSubject: zfsmetricscontract.SnapshotRPCSubject,
		historySubject:  zfsmetricscontract.HistoryRPCSubject,
		snapshotRequest: zfsmetricscontract.GetSnapshotRequest{},
		historyRequest:  zfsmetricscontract.GetHistoryRequest{},
		selectSnapshot: func(response zfsmetricscontract.GetSnapshotResponse) metrics.ZFSSnapshot {
			return response.Snapshot
		},
		selectHistoryItems: func(response zfsmetricscontract.GetHistoryResponse) []metrics.ZFSSnapshot {
			return response.Items
		},
	}
}

// GetSnapshot requests the latest metrics snapshot over messaging.
//
// Parameters:
//   - ctx: request-scoped context used for cancellation and deadlines
func (s metricsRPCService[T, SnapshotResponse, HistoryResponse]) GetSnapshot(ctx context.Context) (T, error) {
	return requestSnapshot(
		ctx,
		s.client,
		s.snapshotSubject,
		s.snapshotRequest,
		s.selectSnapshot,
	)
}

// GetHistory requests the metrics history over messaging.
//
// Parameters:
//   - ctx: request-scoped context used for cancellation and deadlines
func (s metricsRPCService[T, SnapshotResponse, HistoryResponse]) GetHistory(ctx context.Context) ([]T, error) {
	return requestHistory(
		ctx,
		s.client,
		s.historySubject,
		s.historyRequest,
		s.selectHistoryItems,
	)
}
