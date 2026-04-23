package state

import "lite-nas/shared/metrics"

// HistoryStore keeps a bounded chronological history of system snapshots.
//
// The store maintains snapshots in FIFO order. When capacity is reached,
// the oldest snapshot is discarded to make room for a new one.
//
// This object is intentionally simple and does not manage concurrency.
// Synchronization should be handled by the caller if needed.
type HistoryStore struct {
	snapshots []metrics.SystemSnapshot
	capacity  int
}

// NewHistoryStore creates a HistoryStore with the specified capacity.
//
// Capacity defines the maximum number of snapshots retained in memory.
// If capacity is zero or negative, the store will not retain any data.
func NewHistoryStore(capacity int) HistoryStore {
	return HistoryStore{
		snapshots: make([]metrics.SystemSnapshot, 0, capacity),
		capacity:  capacity,
	}
}

// Add appends a new system snapshot to the history.
//
// If the store is full, the oldest snapshot is removed before appending
// the new one.
func (s *HistoryStore) Add(snapshot metrics.SystemSnapshot) {
	if s.capacity <= 0 {
		return
	}

	if len(s.snapshots) == s.capacity {
		// Drop oldest
		s.snapshots = append(s.snapshots[1:], snapshot)
		return
	}

	s.snapshots = append(s.snapshots, snapshot)
}

// List returns the stored snapshot history in chronological order.
//
// The returned slice is a copy, so callers cannot modify the internal
// state of the store.
func (s *HistoryStore) List() []metrics.SystemSnapshot {
	result := make([]metrics.SystemSnapshot, len(s.snapshots))
	copy(result, s.snapshots)

	return result
}

// Latest returns the most recent snapshot.
//
// The second return value is false if the store is empty.
func (s *HistoryStore) Latest() (metrics.SystemSnapshot, bool) {
	if len(s.snapshots) == 0 {
		return metrics.SystemSnapshot{}, false
	}

	return s.snapshots[len(s.snapshots)-1], true
}
