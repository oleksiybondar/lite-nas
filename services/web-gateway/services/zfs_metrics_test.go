package services

import (
	"testing"

	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	"lite-nas/shared/metrics"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestZFSMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := zfsServiceSnapshotFixture(100)
	client := newSnapshotClientStub(t, want, func(response *zfsmetricscontract.GetSnapshotResponse, snapshot metrics.ZFSSnapshot) {
		response.Snapshot = snapshot
	})
	service := NewZFSMetricsService(client)

	got := mustGetSnapshot(t, service)

	assertMetricsSubject(t, client.subject, zfsmetricscontract.SnapshotRPCSubject)
	assertMetricsResult(t, "GetSnapshot()", got, want)
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestZFSMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.ZFSSnapshot{
		zfsServiceSnapshotFixture(100),
		zfsServiceSnapshotFixture(101),
	}
	client := newHistoryClientStub(t, want, func(response *zfsmetricscontract.GetHistoryResponse, history []metrics.ZFSSnapshot) {
		response.Items = history
	})
	service := NewZFSMetricsService(client)

	got := mustGetHistory(t, service)

	assertMetricsSubject(t, client.subject, zfsmetricscontract.HistoryRPCSubject)
	assertMetricsResult(t, "GetHistory()", got, want)
}

// zfsServiceSnapshotFixture returns one representative ZFS snapshot for service tests.
func zfsServiceSnapshotFixture(unixSeconds int64) metrics.ZFSSnapshot {
	return metrics.ZFSSnapshot{Timestamp: unixFromSeconds(unixSeconds)}
}
