package state

import (
	"sync"

	"lite-nas/shared/metrics"
)

// HistoryStore keeps bounded chronological ZFS snapshot history.
type HistoryStore struct {
	mu        sync.RWMutex
	snapshots []metrics.ZFSSnapshot
	capacity  int
}

// NewHistoryStore creates a ZFS history store with the specified capacity.
func NewHistoryStore(capacity int) *HistoryStore {
	return &HistoryStore{
		snapshots: make([]metrics.ZFSSnapshot, 0, max(capacity, 0)),
		capacity:  capacity,
	}
}

// Add appends a snapshot and keeps retention bounded by capacity.
func (s *HistoryStore) Add(snapshot metrics.ZFSSnapshot) {
	if s.capacity <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.snapshots) == s.capacity {
		s.snapshots = append(s.snapshots[1:], snapshot)
		return
	}

	s.snapshots = append(s.snapshots, snapshot)
}

// Latest returns the most recent snapshot and whether any snapshot exists.
func (s *HistoryStore) Latest() (metrics.ZFSSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.snapshots) == 0 {
		return metrics.ZFSSnapshot{}, false
	}

	return s.snapshots[len(s.snapshots)-1], true
}

// List returns a copy of the stored snapshots in chronological order.
func (s *HistoryStore) List() []metrics.ZFSSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]metrics.ZFSSnapshot, len(s.snapshots))
	copy(result, s.snapshots)
	return result
}
