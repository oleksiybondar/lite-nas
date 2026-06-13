package state

import (
	"reflect"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: disk-metrics/FR-010, disk-metrics/FR-012
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	snapshots := []metrics.DiskMetricsSnapshot{
		{Timestamp: time.Unix(1, 0)},
		{Timestamp: time.Unix(2, 0)},
		{Timestamp: time.Unix(3, 0)},
	}

	for _, snapshot := range snapshots {
		store.Add(snapshot)
	}

	want := snapshots[1:]
	if got := store.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
}

// Requirements: disk-metrics/FR-011
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	want := metrics.DiskMetricsSnapshot{Timestamp: time.Unix(4, 0)}
	store.Add(want)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Latest() = %#v, want %#v", got, want)
	}
}

// Requirements: disk-metrics/FR-011, disk-metrics/FR-012
func TestHistoryStoreLatestRemainsAvailableWhenHistoryRetentionIsDisabled(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(0)
	want := metrics.DiskMetricsSnapshot{Timestamp: time.Unix(5, 0)}
	store.Add(want)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Latest() = %#v, want %#v", got, want)
	}
	if gotHistory := store.List(); len(gotHistory) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(gotHistory))
	}
}

func TestHistoryStoreLatestReturnsFalseWhenEmpty(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(1)
	_, ok := store.Latest()
	if ok {
		t.Fatal("expected empty latest result")
	}
}

func TestHistoryStoreListReturnsCopy(t *testing.T) {
	t.Parallel()

	const originalTimestamp = 6
	store := NewHistoryStore(1)
	store.Add(metrics.DiskMetricsSnapshot{Timestamp: time.Unix(originalTimestamp, 0)})

	history := store.List()
	history[0] = metrics.DiskMetricsSnapshot{Timestamp: time.Unix(7, 0)}

	got, _ := store.Latest()
	wantTimestamp := time.Unix(originalTimestamp, 0)
	if got.Timestamp != wantTimestamp {
		t.Fatalf("Latest().Timestamp = %v, want %v", got.Timestamp, wantTimestamp)
	}
}
