package state

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: zfs-metrics-svc/FR-006
func TestHistoryStoreLatestReturnsMostRecentSnapshot(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(3)
	older := metrics.ZFSSnapshot{Timestamp: time.Unix(100, 0)}
	newer := metrics.ZFSSnapshot{Timestamp: time.Unix(200, 0)}

	store.Add(older)
	store.Add(newer)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("Latest() ok = false, want true")
	}
	if !got.Timestamp.Equal(newer.Timestamp) {
		t.Fatalf("Latest() timestamp = %v, want %v", got.Timestamp, newer.Timestamp)
	}
}

// Requirements: zfs-metrics-svc/FR-006
func TestHistoryStoreRetentionDropsOldestWhenFull(t *testing.T) {
	t.Parallel()

	store := NewHistoryStore(2)
	first := metrics.ZFSSnapshot{Timestamp: time.Unix(100, 0)}
	second := metrics.ZFSSnapshot{Timestamp: time.Unix(200, 0)}
	third := metrics.ZFSSnapshot{Timestamp: time.Unix(300, 0)}

	store.Add(first)
	store.Add(second)
	store.Add(third)

	history := store.List()
	if len(history) != 2 {
		t.Fatalf("len(List()) = %d, want 2", len(history))
	}
	if !history[0].Timestamp.Equal(second.Timestamp) {
		t.Fatalf("history[0] timestamp = %v, want %v", history[0].Timestamp, second.Timestamp)
	}
	if !history[1].Timestamp.Equal(third.Timestamp) {
		t.Fatalf("history[1] timestamp = %v, want %v", history[1].Timestamp, third.Timestamp)
	}
}
