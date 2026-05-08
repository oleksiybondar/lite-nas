package controllers

import (
	"context"
	"errors"
	"reflect"
	"testing"

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

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	assertSuccessfulSystemMetricsEnvelope(t, got.Body.Success, got.Body.Timestamp.IsZero())

	if !reflect.DeepEqual(got.Body.Data, snapshot) {
		t.Fatalf("Data = %#v, want %#v", got.Body.Data, snapshot)
	}
}

// Requirements: web-gateway/FR-002, web-gateway/FR-003, web-gateway/TR-001
func TestSystemMetricsControllerGetHistoryWrapsResponseInEnvelope(t *testing.T) {
	t.Parallel()

	history := []metrics.SystemSnapshot{
		systemSnapshotFixture(123, 12.5, []float64{10, 15}, 1024, 512, 50),
		systemSnapshotFixture(124, 20, []float64{18, 22}, 1024, 640, 62.5),
	}
	controller := NewSystemMetricsController(stubSystemMetricsService{history: history})

	got, err := controller.GetHistory(context.Background(), &struct{}{})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	assertSuccessfulSystemMetricsEnvelope(t, got.Body.Success, got.Body.Timestamp.IsZero())

	if len(got.Body.Data) != len(history) {
		t.Fatalf("len(Data) = %d, want %d", len(got.Body.Data), len(history))
	}

	for i := range history {
		if !reflect.DeepEqual(got.Body.Data[i], history[i]) {
			t.Fatalf("Data[%d] = %#v, want %#v", i, got.Body.Data[i], history[i])
		}
	}
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestSystemMetricsControllerGetSnapshotMapsBackendFailure(t *testing.T) {
	t.Parallel()

	controller := NewSystemMetricsController(stubSystemMetricsService{err: errors.New("backend failed")})

	got, err := controller.GetSnapshot(context.Background(), &struct{}{})
	if err == nil {
		t.Fatal("GetSnapshot() error = nil, want error")
	}

	if got != nil {
		t.Fatalf("GetSnapshot() result = %#v, want nil", got)
	}
}
