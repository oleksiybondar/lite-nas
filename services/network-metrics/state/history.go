package state

import (
	"sync"

	"lite-nas/shared/metrics"
)

// HistoryStore keeps bounded chronological network snapshot history.
type HistoryStore struct {
	mu        sync.RWMutex
	snapshots []metrics.NetworkMetricsSnapshot
	capacity  int
	latest    metrics.NetworkMetricsSnapshot
	hasLatest bool
}

// NewHistoryStore creates a network history store with the specified capacity.
func NewHistoryStore(capacity int) *HistoryStore {
	initialCapacity := max(capacity, 0)

	return &HistoryStore{
		snapshots: make([]metrics.NetworkMetricsSnapshot, 0, initialCapacity),
		capacity:  capacity,
	}
}

// Add appends a snapshot, keeps the latest snapshot available, and bounds
// retained history by capacity.
func (s *HistoryStore) Add(snapshot metrics.NetworkMetricsSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.latest = snapshot
	s.hasLatest = true

	if s.capacity <= 0 {
		return
	}

	if len(s.snapshots) == s.capacity {
		s.snapshots = append(s.snapshots[1:], snapshot)
		return
	}

	s.snapshots = append(s.snapshots, snapshot)
}

// Latest returns the most recent snapshot and whether any snapshot exists.
func (s *HistoryStore) Latest() (metrics.NetworkMetricsSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.hasLatest {
		return metrics.NetworkMetricsSnapshot{}, false
	}

	return s.latest, true
}

// List returns a copy of the stored snapshots in chronological order.
func (s *HistoryStore) List() []metrics.NetworkMetricsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]metrics.NetworkMetricsSnapshot, len(s.snapshots))
	copy(result, s.snapshots)

	return result
}
