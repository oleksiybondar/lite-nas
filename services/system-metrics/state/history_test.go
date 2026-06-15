package state

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/historytest"
)

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-004
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	historytest.RunDropsOldestAtCapacityCase(t, newHistoryStoreAdapter, []metrics.SystemSnapshot{
		{Timestamp: time.Unix(1, 0)},
		{Timestamp: time.Unix(2, 0)},
		{Timestamp: time.Unix(3, 0)},
	})
}

// Requirements: system-metrics-svc/FR-003
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(4, 0)}
	historytest.RunLatestReturnsMostRecentCase(t, newHistoryStoreAdapter, snapshot)
}

func TestHistoryStoreLatestReturnsFalseWhenEmpty(t *testing.T) {
	t.Parallel()

	historytest.RunLatestReturnsFalseWhenEmptyCase(t, newHistoryStoreAdapter, metrics.SystemSnapshot{})
}

func TestHistoryStoreListReturnsCopy(t *testing.T) {
	t.Parallel()

	historytest.RunListReturnsCopyCase(
		t,
		newHistoryStoreAdapter,
		metrics.SystemSnapshot{Timestamp: time.Unix(5, 0)},
		metrics.SystemSnapshot{Timestamp: time.Unix(6, 0)},
	)
}

// newHistoryStoreAdapter exposes the concrete history store through the shared
// generic test helper contract.
func newHistoryStoreAdapter(capacity int) historytest.Store[metrics.SystemSnapshot] {
	store := NewHistoryStore(capacity)
	return &store
}
