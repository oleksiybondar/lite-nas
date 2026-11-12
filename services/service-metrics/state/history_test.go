package state

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/historytest"
)

// Requirements: service-metrics-svc/FR-010, service-metrics-svc/FR-012
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	snapshots := []metrics.ServiceMetricsSnapshot{
		{Timestamp: time.Unix(1, 0)},
		{Timestamp: time.Unix(2, 0)},
		{Timestamp: time.Unix(3, 0)},
	}

	historytest.RunDropsOldestAtCapacityCase(t, newHistoryStoreAdapter, snapshots)
}

// Requirements: service-metrics-svc/FR-011
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	want := metrics.ServiceMetricsSnapshot{Timestamp: time.Unix(4, 0)}
	historytest.RunLatestReturnsMostRecentCase(t, newHistoryStoreAdapter, want)
}

// Requirements: service-metrics-svc/FR-011, service-metrics-svc/FR-012
func TestHistoryStoreLatestRemainsAvailableWhenHistoryRetentionIsDisabled(t *testing.T) {
	t.Parallel()

	want := metrics.ServiceMetricsSnapshot{Timestamp: time.Unix(5, 0)}
	historytest.RunLatestRemainsAvailableWhenHistoryRetentionIsDisabledCase(t, newHistoryStoreAdapter, want)
}

// newHistoryStoreAdapter exposes the concrete history store through the shared
// generic test helper contract.
func newHistoryStoreAdapter(capacity int) historytest.Store[metrics.ServiceMetricsSnapshot] {
	return NewHistoryStore(capacity)
}
