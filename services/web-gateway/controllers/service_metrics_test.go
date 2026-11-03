package controllers

import (
	"context"
	"testing"
	"time"

	servicemetricsdto "lite-nas/services/web-gateway/dto/service_metrics"
	"lite-nas/shared/metrics"
)

type stubServiceMetricsService struct {
	snapshot metrics.ServiceMetricsSnapshot
	history  []metrics.ServiceMetricsSnapshot
	err      error
}

func (s stubServiceMetricsService) GetSnapshot(context.Context) (metrics.ServiceMetricsSnapshot, error) {
	if s.err != nil {
		return metrics.ServiceMetricsSnapshot{}, s.err
	}
	return s.snapshot, nil
}

func (s stubServiceMetricsService) GetHistory(context.Context) ([]metrics.ServiceMetricsSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.history, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestServiceMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	snapshot := serviceMetricsControllerFixture(123)
	controller := NewServiceMetricsController(stubServiceMetricsService{snapshot: snapshot})
	assertSnapshotWrapped(
		t,
		snapshot,
		controller.GetSnapshot,
		func(output *servicemetricsdto.ServiceSnapshotOutput) bool { return output.Body.Success },
		func(output *servicemetricsdto.ServiceSnapshotOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *servicemetricsdto.ServiceSnapshotOutput) metrics.ServiceMetricsSnapshot {
			return output.Body.Data
		},
	)
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestServiceMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	history := []metrics.ServiceMetricsSnapshot{
		serviceMetricsControllerFixture(123),
		serviceMetricsControllerFixture(124),
	}
	controller := NewServiceMetricsController(stubServiceMetricsService{history: history})
	assertHistoryWrapped(
		t,
		history,
		controller.GetHistory,
		func(output *servicemetricsdto.ServiceHistoryOutput) bool { return output.Body.Success },
		func(output *servicemetricsdto.ServiceHistoryOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *servicemetricsdto.ServiceHistoryOutput) []metrics.ServiceMetricsSnapshot {
			return output.Body.Data
		},
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestServiceMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewServiceMetricsController(stubServiceMetricsService{err: backendFailure()})
	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}

func serviceMetricsControllerFixture(unixSeconds int64) metrics.ServiceMetricsSnapshot {
	enabled := "enabled"
	active := "active"
	startedAt := time.Unix(unixSeconds-10, 0).UTC()
	return metrics.ServiceMetricsSnapshot{
		Timestamp: time.Unix(unixSeconds, 0).UTC(),
		Services: []metrics.ServiceUnitSnapshot{{
			Name:         "nats-server.service",
			Manager:      "systemd",
			UnitType:     "service",
			EnabledState: &enabled,
			ActiveState:  &active,
			StartedAt:    &startedAt,
		}},
	}
}
