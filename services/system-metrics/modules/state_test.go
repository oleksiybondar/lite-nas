package modules

import (
	"reflect"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: system-metrics-svc/FR-003
func TestNewStateModuleLatestReturnsFalseWhenEmpty(t *testing.T) {
	t.Parallel()

	_, ok := NewStateModule(2).SnapshotStore().Latest()
	if ok {
		t.Fatal("expected empty latest result")
	}
}

// Requirements: system-metrics-svc/FR-002, system-metrics-svc/FR-003, system-metrics-svc/FR-004
func TestSnapshotStoreRetainsChronologicalHistoryAndLatestSnapshot(t *testing.T) {
	t.Parallel()

	store := NewStateModule(2).SnapshotStore()
	first := metrics.SystemSnapshot{Timestamp: time.Unix(10, 0)}
	second := metrics.SystemSnapshot{Timestamp: time.Unix(11, 0)}

	store.Add(first)
	store.Add(second)

	latest, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}

	if !reflect.DeepEqual(latest, second) {
		t.Fatalf("Latest() = %#v, want %#v", latest, second)
	}

	wantHistory := []metrics.SystemSnapshot{first, second}
	if got := store.List(); !reflect.DeepEqual(got, wantHistory) {
		t.Fatalf("List() = %#v, want %#v", got, wantHistory)
	}
}
