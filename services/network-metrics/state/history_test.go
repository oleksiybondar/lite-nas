package state

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/historytest"
)

// Requirements: network-metrics-svc/FR-007, network-metrics-svc/FR-009
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	historytest.RunDropsOldestAtCapacityCase(t, newHistoryStoreAdapter, []metrics.NetworkMetricsSnapshot{
		{Timestamp: time.Unix(1, 0)},
		{Timestamp: time.Unix(2, 0)},
		{Timestamp: time.Unix(3, 0)},
	})
}

// Requirements: network-metrics-svc/FR-008
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	snapshot := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(4, 0)}
	historytest.RunLatestReturnsMostRecentCase(t, newHistoryStoreAdapter, snapshot)
}

// Requirements: network-metrics-svc/FR-008, network-metrics-svc/FR-009
func TestHistoryStoreLatestRemainsAvailableWhenHistoryRetentionIsDisabled(t *testing.T) {
	t.Parallel()

	snapshot := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(5, 0)}
	historytest.RunLatestRemainsAvailableWhenHistoryRetentionIsDisabledCase(t, newHistoryStoreAdapter, snapshot)
}

func TestHistoryStoreLatestReturnsFalseWhenEmpty(t *testing.T) {
	t.Parallel()

	historytest.RunLatestReturnsFalseWhenEmptyCase(t, newHistoryStoreAdapter, metrics.NetworkMetricsSnapshot{})
}

func TestHistoryStoreListReturnsCopy(t *testing.T) {
	t.Parallel()

	historytest.RunListReturnsCopyCase(
		t,
		newHistoryStoreAdapter,
		metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(6, 0)},
		metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(7, 0)},
	)
}

// newHistoryStoreAdapter exposes the concrete history store through the shared
// generic test helper contract.
func newHistoryStoreAdapter(capacity int) historytest.Store[metrics.NetworkMetricsSnapshot] {
	return NewHistoryStore(capacity)
}
