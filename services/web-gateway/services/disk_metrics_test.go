package services

import (
	"testing"
	"time"

	diskmetricscontract "lite-nas/shared/contracts/diskmetrics"
	"lite-nas/shared/metrics"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestDiskMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := metrics.DiskMetricsSnapshot{Timestamp: time.Unix(100, 0)}
	client := newSnapshotClientStub(t, want, func(response *diskmetricscontract.GetSnapshotResponse, snapshot metrics.DiskMetricsSnapshot) {
		response.Snapshot = snapshot
	})
	service := NewDiskMetricsService(client)

	got := mustGetSnapshot(t, service)

	assertMetricsSubject(t, client.subject, diskmetricscontract.SnapshotRPCSubject)
	assertMetricsResult(t, "GetSnapshot()", got, want)
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestDiskMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.DiskMetricsSnapshot{
		{Timestamp: time.Unix(100, 0)},
		{Timestamp: time.Unix(101, 0)},
	}
	client := newHistoryClientStub(t, want, func(response *diskmetricscontract.GetHistoryResponse, history []metrics.DiskMetricsSnapshot) {
		response.Items = history
	})
	service := NewDiskMetricsService(client)

	got := mustGetHistory(t, service)

	assertMetricsSubject(t, client.subject, diskmetricscontract.HistoryRPCSubject)
	assertMetricsResult(t, "GetHistory()", got, want)
}
