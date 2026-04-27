package services

import (
	"context"

	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

const (
	statsRPCSubject   = "system.metrics.rpc.stats.get"
	historyRPCSubject = "system.metrics.rpc.history.get"
)

// SystemMetricsService defines the backend-facing system metrics flows used by
// the gateway service layer.
type SystemMetricsService interface {
	GetSnapshot(ctx context.Context) (metrics.SystemSnapshot, error)
	GetHistory(ctx context.Context) ([]metrics.SystemSnapshot, error)
}

type systemMetricsService struct {
	client messaging.Client
}

// NewSystemMetricsService creates a service that fetches system metrics over
// the shared messaging transport.
//
// Parameters:
//   - client: messaging client used for request/reply RPC calls
func NewSystemMetricsService(client messaging.Client) SystemMetricsService {
	return systemMetricsService{client: client}
}

// GetSnapshot requests the latest system metrics snapshot over messaging.
//
// Parameters:
//   - ctx: request-scoped context used for cancellation and deadlines
func (s systemMetricsService) GetSnapshot(ctx context.Context) (metrics.SystemSnapshot, error) {
	var snapshot metrics.SystemSnapshot
	if err := s.client.Request(ctx, statsRPCSubject, map[string]any{}, &snapshot); err != nil {
		return metrics.SystemSnapshot{}, err
	}

	return snapshot, nil
}

// GetHistory requests the system metrics history over messaging.
//
// Parameters:
//   - ctx: request-scoped context used for cancellation and deadlines
func (s systemMetricsService) GetHistory(ctx context.Context) ([]metrics.SystemSnapshot, error) {
	var history []metrics.SystemSnapshot
	if err := s.client.Request(ctx, historyRPCSubject, map[string]any{}, &history); err != nil {
		return nil, err
	}

	return history, nil
}
