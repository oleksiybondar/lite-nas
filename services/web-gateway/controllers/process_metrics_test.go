package controllers

import (
	"context"
	"testing"
	"time"

	processmetricsdto "lite-nas/services/web-gateway/dto/process_metrics"
	"lite-nas/shared/metrics"
)

type stubProcessMetricsService struct {
	snapshot metrics.ProcessMetricsSnapshot
	err      error
}

func (s stubProcessMetricsService) GetSnapshot(context.Context) (metrics.ProcessMetricsSnapshot, error) {
	if s.err != nil {
		return metrics.ProcessMetricsSnapshot{}, s.err
	}

	return s.snapshot, nil
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestProcessMetricsControllerGetSnapshotWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	snapshot := metrics.ProcessMetricsSnapshot{
		Timestamp: time.Unix(123, 0).UTC(),
		Processes: []metrics.ProcessMetric{{PID: 101, Name: "alpha"}},
	}
	controller := NewProcessMetricsController(stubProcessMetricsService{snapshot: snapshot})
	assertSnapshotWrapped(
		t,
		snapshot,
		controller.GetSnapshot,
		func(output *processmetricsdto.ProcessSnapshotOutput) bool { return output.Body.Success },
		func(output *processmetricsdto.ProcessSnapshotOutput) bool { return output.Body.Timestamp.IsZero() },
		func(output *processmetricsdto.ProcessSnapshotOutput) metrics.ProcessMetricsSnapshot {
			return output.Body.Data
		},
	)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestProcessMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewProcessMetricsController(stubProcessMetricsService{err: backendFailure()})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	assertBackendFailureMapped(t, got, err, "GetSnapshot()")
}
