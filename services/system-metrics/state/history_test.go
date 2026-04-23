package state

import (
	"reflect"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-004
func TestHistoryStoreDropsOldestSnapshotAtCapacity(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	first := metrics.SystemSnapshot{Timestamp: time.Unix(1, 0)}
	second := metrics.SystemSnapshot{Timestamp: time.Unix(2, 0)}
	third := metrics.SystemSnapshot{Timestamp: time.Unix(3, 0)}

	store.Add(first)
	store.Add(second)
	store.Add(third)

	wantHistory := []metrics.SystemSnapshot{second, third}
	if got := store.List(); !reflect.DeepEqual(got, wantHistory) {
		t.Fatalf("List() = %#v, want %#v", got, wantHistory)
	}
}

// Requirements: system-metrics-svc/FR-003
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	snapshot := metrics.SystemSnapshot{Timestamp: time.Unix(4, 0)}
	store.Add(snapshot)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", got, snapshot)
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
	store.Add(metrics.SystemSnapshot{Timestamp: time.Unix(5, 0)})

	history := store.List()
	history[0] = metrics.SystemSnapshot{Timestamp: time.Unix(6, 0)}

	got, _ := store.Latest()
	if got.Timestamp != time.Unix(5, 0) {
		t.Fatalf("Latest().Timestamp = %v, want %v", got.Timestamp, time.Unix(5, 0))
	}
}
