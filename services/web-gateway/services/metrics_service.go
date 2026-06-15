package services

import (
	"context"

	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
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

// ServiceMetricsService defines the backend-facing service metrics flows used by
// the gateway service layer.
type ServiceMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.ServiceMetricsSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.ServiceMetricsSnapshot, error)
}

// ZFSMetricsService defines the backend-facing ZFS metrics flows used by the
// gateway service layer.
type ZFSMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.ZFSSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.ZFSSnapshot, error)
}

// DiskMetricsService defines the backend-facing disk metrics flows used by the
// gateway service layer.
type DiskMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.DiskMetricsSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.DiskMetricsSnapshot, error)
}

// NetworkMetricsService defines the backend-facing network metrics flows used by
// the gateway service layer.
type NetworkMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.NetworkMetricsSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.NetworkMetricsSnapshot, error)
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

// NewServiceMetricsService creates a service that fetches service metrics over
// the shared messaging transport.
func NewServiceMetricsService(client messaging.Client) ServiceMetricsService {
	return metricsRPCService[
		metrics.ServiceMetricsSnapshot,
		servicemetricscontract.GetSnapshotResponse,
		servicemetricscontract.GetHistoryResponse,
	]{
		client:          client,
		snapshotSubject: servicemetricscontract.SnapshotRPCSubject,
		historySubject:  servicemetricscontract.HistoryRPCSubject,
		snapshotRequest: servicemetricscontract.GetSnapshotRequest{},
		historyRequest:  servicemetricscontract.GetHistoryRequest{},
		selectSnapshot: func(response servicemetricscontract.GetSnapshotResponse) metrics.ServiceMetricsSnapshot {
			return response.Snapshot
		},
		selectHistoryItems: func(response servicemetricscontract.GetHistoryResponse) []metrics.ServiceMetricsSnapshot {
			return response.Items
		},
	}
}

// NewZFSMetricsService creates a service that fetches ZFS metrics over the
// shared messaging transport.
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

// NewDiskMetricsService creates a service that fetches disk metrics over the
// shared messaging transport.
func NewDiskMetricsService(client messaging.Client) DiskMetricsService {
	return metricsRPCService[
		metrics.DiskMetricsSnapshot,
		diskmetricscontract.GetSnapshotResponse,
		diskmetricscontract.GetHistoryResponse,
	]{
		client:          client,
		snapshotSubject: diskmetricscontract.SnapshotRPCSubject,
		historySubject:  diskmetricscontract.HistoryRPCSubject,
		snapshotRequest: diskmetricscontract.GetSnapshotRequest{},
		historyRequest:  diskmetricscontract.GetHistoryRequest{},
		selectSnapshot: func(response diskmetricscontract.GetSnapshotResponse) metrics.DiskMetricsSnapshot {
			return response.Snapshot
		},
		selectHistoryItems: func(response diskmetricscontract.GetHistoryResponse) []metrics.DiskMetricsSnapshot {
			return response.Items
		},
	}
}

// NewNetworkMetricsService creates a service that fetches network metrics over
// the shared messaging transport.
func NewNetworkMetricsService(client messaging.Client) NetworkMetricsService {
	return metricsRPCService[
		metrics.NetworkMetricsSnapshot,
		networkmetricscontract.GetSnapshotResponse,
		networkmetricscontract.GetHistoryResponse,
	]{
		client:          client,
		snapshotSubject: networkmetricscontract.SnapshotRPCSubject,
		historySubject:  networkmetricscontract.HistoryRPCSubject,
		snapshotRequest: networkmetricscontract.GetSnapshotRequest{},
		historyRequest:  networkmetricscontract.GetHistoryRequest{},
		selectSnapshot: func(response networkmetricscontract.GetSnapshotResponse) metrics.NetworkMetricsSnapshot {
			return response.Snapshot
		},
		selectHistoryItems: func(response networkmetricscontract.GetHistoryResponse) []metrics.NetworkMetricsSnapshot {
			return response.Items
		},
	}
}

// GetSnapshot requests the latest metrics snapshot over messaging.
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
func (s metricsRPCService[T, SnapshotResponse, HistoryResponse]) GetHistory(ctx context.Context) ([]T, error) {
	return requestHistory(
		ctx,
		s.client,
		s.historySubject,
		s.historyRequest,
		s.selectHistoryItems,
	)
}
