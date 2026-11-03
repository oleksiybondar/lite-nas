package services

import (
	"reflect"
	"testing"
	"time"

	servicemetricscontract "lite-nas/shared/contracts/servicemetrics"
	"lite-nas/shared/metrics"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestServiceMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := serviceMetricsSnapshotFixture(100)
	client := newSnapshotClientStub(t, want, func(response *servicemetricscontract.GetSnapshotResponse, snapshot metrics.ServiceMetricsSnapshot) {
		response.Available = true
		response.Snapshot = snapshot
	})
	service := NewServiceMetricsService(client)

	got := mustGetSnapshot(t, service)
	assertMetricsSubject(t, client.subject, servicemetricscontract.SnapshotRPCSubject)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetSnapshot() = %#v, want %#v", got, want)
	}
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestServiceMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.ServiceMetricsSnapshot{
		serviceMetricsSnapshotFixture(100),
		serviceMetricsSnapshotFixture(101),
	}
	client := newHistoryClientStub(t, want, func(response *servicemetricscontract.GetHistoryResponse, history []metrics.ServiceMetricsSnapshot) {
		response.Items = history
	})
	service := NewServiceMetricsService(client)

	got := mustGetHistory(t, service)
	assertMetricsSubject(t, client.subject, servicemetricscontract.HistoryRPCSubject)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetHistory() = %#v, want %#v", got, want)
	}
}

func serviceMetricsSnapshotFixture(unixSeconds int64) metrics.ServiceMetricsSnapshot {
	enabled := "enabled"
	active := "active"
	return metrics.ServiceMetricsSnapshot{
		Timestamp: time.Unix(unixSeconds, 0),
		Services: []metrics.ServiceUnitSnapshot{{
			Name:         "lite-nas-system-metrics.service",
			Manager:      "systemd",
			UnitType:     "service",
			EnabledState: &enabled,
			ActiveState:  &active,
		}},
	}
}
