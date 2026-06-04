package controllers

import (
	"context"
	"testing"

	systemmetricsdto "lite-nas/services/web-gateway/dto/system_metrics"
	"lite-nas/shared/metrics"
)

type stubSystemMetricsService struct {
	snapshot metrics.SystemSnapshot
	history  []metrics.SystemSnapshot
	err      error
}

func (s stubSystemMetricsService) GetSnapshot(context.Context) (metrics.SystemSnapshot, error) {
	if s.err != nil {
		return metrics.SystemSnapshot{}, s.err
	}

	return s.snapshot, nil
}

func (s stubSystemMetricsService) GetHistory(context.Context) ([]metrics.SystemSnapshot, error) {
	if s.err != nil {
		return nil, s.err
	}

	return s.history, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestSystemMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	snapshot := systemSnapshotFixture(123, 12.5, []float64{10, 15}, 1024, 512, 50)
	controller := NewSystemMetricsController(stubSystemMetricsService{snapshot: snapshot})
	assertSnapshotWrapped(
		t,
		snapshot,
		controller.GetSnapshot,
		func(output *systemmetricsdto.SnapshotOutput) bool { return output.Body.Success },
		func(output *systemmetricsdto.SnapshotOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *systemmetricsdto.SnapshotOutput) metrics.SystemSnapshot { return output.Body.Data },
	)
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestSystemMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	history := []metrics.SystemSnapshot{
		systemSnapshotFixture(123, 12.5, []float64{10, 15}, 1024, 512, 50),
		systemSnapshotFixture(124, 20, []float64{18, 22}, 1024, 640, 62.5),
	}
	controller := NewSystemMetricsController(stubSystemMetricsService{history: history})
	assertHistoryWrapped(
		t,
		history,
		controller.GetHistory,
		func(output *systemmetricsdto.HistoryOutput) bool { return output.Body.Success },
		func(output *systemmetricsdto.HistoryOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *systemmetricsdto.HistoryOutput) []metrics.SystemSnapshot { return output.Body.Data },
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestSystemMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewSystemMetricsController(stubSystemMetricsService{err: backendFailure()})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}
