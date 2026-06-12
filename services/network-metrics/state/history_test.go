package state

import (
	"reflect"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: network-metrics-svc/FR-007, network-metrics-svc/FR-009
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	first := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(1, 0)}
	second := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(2, 0)}
	third := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(3, 0)}

	store.Add(first)
	store.Add(second)
	store.Add(third)

	wantHistory := []metrics.NetworkMetricsSnapshot{second, third}
	if got := store.List(); !reflect.DeepEqual(got, wantHistory) {
		t.Fatalf("List() = %#v, want %#v", got, wantHistory)
	}
}

// Requirements: network-metrics-svc/FR-008
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	snapshot := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(4, 0)}
	store.Add(snapshot)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", got, snapshot)
	}
}

// Requirements: network-metrics-svc/FR-008, network-metrics-svc/FR-009
func TestHistoryStoreLatestRemainsAvailableWhenHistoryRetentionIsDisabled(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(0)
	snapshot := metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(5, 0)}
	store.Add(snapshot)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", got, snapshot)
	}
	if history := store.List(); len(history) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(history))
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

	store := NewHistoryStore(1)
	store.Add(metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(6, 0)})

	history := store.List()
	history[0] = metrics.NetworkMetricsSnapshot{Timestamp: time.Unix(7, 0)}

	got, _ := store.Latest()
	if got.Timestamp != time.Unix(6, 0) {
		t.Fatalf("Latest().Timestamp = %v, want %v", got.Timestamp, time.Unix(6, 0))
	}
}
