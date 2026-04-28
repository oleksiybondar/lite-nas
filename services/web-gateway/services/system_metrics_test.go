package services

import (
	"context"
	"reflect"
	"testing"
	"time"

	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/systemmetricstest"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestSystemMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := metrics.SystemSnapshot{Timestamp: time.Unix(100, 0)}
	client := systemmetricstest.NewSnapshotClient(want)
	service := NewSystemMetricsService(client)

	got := mustGetSnapshot(t, service)

	if client.Subject != systemmetricscontract.SnapshotRPCSubject {
		t.Fatalf("subject = %q, want %q", client.Subject, systemmetricscontract.SnapshotRPCSubject)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetSnapshot() = %#v, want %#v", got, want)
	}
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestSystemMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.SystemSnapshot{
		{Timestamp: time.Unix(100, 0)},
		{Timestamp: time.Unix(101, 0)},
	}
	client := systemmetricstest.NewHistoryClient(want)
	service := NewSystemMetricsService(client)

	got := mustGetHistory(t, service)

	if client.Subject != systemmetricscontract.HistoryRPCSubject {
		t.Fatalf("subject = %q, want %q", client.Subject, systemmetricscontract.HistoryRPCSubject)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetHistory() = %#v, want %#v", got, want)
	}
}

func mustGetSnapshot(t *testing.T, service SystemMetricsService) metrics.SystemSnapshot {
	t.Helper()

	got, err := service.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	return got
}

func mustGetHistory(t *testing.T, service SystemMetricsService) []metrics.SystemSnapshot {
	t.Helper()

	got, err := service.GetHistory(context.Background())
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	return got
}
