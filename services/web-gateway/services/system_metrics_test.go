package services

import (
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
