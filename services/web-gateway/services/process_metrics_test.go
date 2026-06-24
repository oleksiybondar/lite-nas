package services

import (
	"testing"
	"time"

	processmetricscontract "lite-nas/shared/contracts/processmetrics"
	"lite-nas/shared/metrics"
)

// Requirements: web-gateway/FR-003, web-gateway/IR-002
func TestProcessMetricsServiceRequestsSnapshotSubject(t *testing.T) {
	t.Parallel()

	want := metrics.ProcessMetricsSnapshot{Timestamp: time.Unix(100, 0).UTC()}
	client := newSnapshotClientStub(t, want, func(response *processmetricscontract.GetSnapshotResponse, snapshot metrics.ProcessMetricsSnapshot) {
		response.Available = true
		response.Snapshot = snapshot
	})
	service := NewProcessMetricsService(client)

	got := mustGetSnapshot(t, service)

	assertMetricsSubject(t, client.subject, processmetricscontract.SnapshotRPCSubject)
	assertMetricsResult(t, "GetSnapshot()", got, want)
}
