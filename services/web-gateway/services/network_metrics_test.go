package services

import (
	"testing"

	"lite-nas/services/web-gateway/testutil/networkmetricstest"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	"lite-nas/shared/metrics"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestNetworkMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := networkmetricstest.Snapshot(100)
	client := newSnapshotClientStub(t, want, func(response *networkmetricscontract.GetSnapshotResponse, snapshot metrics.NetworkMetricsSnapshot) {
		response.Snapshot = snapshot
	})
	service := NewNetworkMetricsService(client)

	got := mustGetSnapshot(t, service)

	assertMetricsSubject(t, client.subject, networkmetricscontract.SnapshotRPCSubject)
	assertMetricsResult(t, "GetSnapshot()", got, want)
}

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestNetworkMetricsServiceRequestsHistorySubject(t *testing.T) {
	t.Parallel()

	want := []metrics.NetworkMetricsSnapshot{
		networkmetricstest.Snapshot(100),
		networkmetricstest.Snapshot(101),
	}
	client := newHistoryClientStub(t, want, func(response *networkmetricscontract.GetHistoryResponse, history []metrics.NetworkMetricsSnapshot) {
		response.Items = history
	})
	service := NewNetworkMetricsService(client)

	got := mustGetHistory(t, service)

	assertMetricsSubject(t, client.subject, networkmetricscontract.HistoryRPCSubject)
	assertMetricsResult(t, "GetHistory()", got, want)
}
