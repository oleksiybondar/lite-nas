package state

import (
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: process-metrics-svc/IR-001
func TestSnapshotStoreReturnsLatestSnapshot(t *testing.T) {
	t.Parallel()

	store := NewSnapshotStore()
	snapshot := metrics.ProcessMetricsSnapshot{Timestamp: time.Unix(123, 0).UTC()}
	store.Add(snapshot)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("Latest() ok = false, want true")
	}
	if !got.Timestamp.Equal(snapshot.Timestamp) {
		t.Fatalf("Latest().Timestamp = %v, want %v", got.Timestamp, snapshot.Timestamp)
	}
}

// Requirements: process-metrics-svc/IR-001
func TestSnapshotStoreReportsEmptyState(t *testing.T) {
	t.Parallel()

	if _, ok := NewSnapshotStore().Latest(); ok {
		t.Fatal("Latest() ok = true, want false")
	}
}
