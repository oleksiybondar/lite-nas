package state

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/historytest"
)

// Requirements: disk-metrics/FR-010, disk-metrics/FR-012
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	historytest.RunDropsOldestAtCapacityCase(t, newHistoryStoreAdapter, []metrics.DiskMetricsSnapshot{
		{Timestamp: time.Unix(1, 0)},
		{Timestamp: time.Unix(2, 0)},
		{Timestamp: time.Unix(3, 0)},
	})
}

// Requirements: disk-metrics/FR-011
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	snapshot := metrics.DiskMetricsSnapshot{Timestamp: time.Unix(4, 0)}
	historytest.RunLatestReturnsMostRecentCase(t, newHistoryStoreAdapter, snapshot)
}

// Requirements: disk-metrics/FR-011, disk-metrics/FR-012
func TestHistoryStoreLatestRemainsAvailableWhenHistoryRetentionIsDisabled(t *testing.T) {
	t.Parallel()

	snapshot := metrics.DiskMetricsSnapshot{Timestamp: time.Unix(5, 0)}
	historytest.RunLatestRemainsAvailableWhenHistoryRetentionIsDisabledCase(t, newHistoryStoreAdapter, snapshot)
}

func TestHistoryStoreLatestReturnsFalseWhenEmpty(t *testing.T) {
	t.Parallel()

	historytest.RunLatestReturnsFalseWhenEmptyCase(t, newHistoryStoreAdapter, metrics.DiskMetricsSnapshot{})
}

func TestHistoryStoreListReturnsCopy(t *testing.T) {
	t.Parallel()

	historytest.RunListReturnsCopyCase(
		t,
		newHistoryStoreAdapter,
		metrics.DiskMetricsSnapshot{Timestamp: time.Unix(6, 0)},
		metrics.DiskMetricsSnapshot{Timestamp: time.Unix(7, 0)},
	)
}

// newHistoryStoreAdapter exposes the concrete history store through the shared
// generic test helper contract.
func newHistoryStoreAdapter(capacity int) historytest.Store[metrics.DiskMetricsSnapshot] {
	return NewHistoryStore(capacity)
}
