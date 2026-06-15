package state

import (
	"sync"

	"lite-nas/shared/metrics"
)

// HistoryStore keeps bounded chronological service snapshot history.
type HistoryStore struct {
	mu        sync.RWMutex
	snapshots []metrics.ServiceMetricsSnapshot
	capacity  int
	latest    metrics.ServiceMetricsSnapshot
	hasLatest bool
}

// NewHistoryStore creates a service history store with the specified capacity.
func NewHistoryStore(capacity int) *HistoryStore {
	initialCapacity := max(capacity, 0)

	return &HistoryStore{
		snapshots: make([]metrics.ServiceMetricsSnapshot, 0, initialCapacity),
		capacity:  capacity,
	}
}

// Add appends a snapshot, keeps the latest snapshot available, and bounds
// retained history by capacity.
func (s *HistoryStore) Add(snapshot metrics.ServiceMetricsSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.latest = snapshot
	s.hasLatest = true
	s.snapshots = s.appendSnapshot(snapshot)
}

// Latest returns the most recent snapshot and whether any snapshot exists.
func (s *HistoryStore) Latest() (metrics.ServiceMetricsSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.hasLatest {
		return metrics.ServiceMetricsSnapshot{}, false
	}

	return s.latest, true
}

// List returns a copy of the stored snapshots in chronological order.
func (s *HistoryStore) List() []metrics.ServiceMetricsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]metrics.ServiceMetricsSnapshot(nil), s.snapshots...)
}

// appendSnapshot applies the configured retention policy to one snapshot write.
func (s *HistoryStore) appendSnapshot(snapshot metrics.ServiceMetricsSnapshot) []metrics.ServiceMetricsSnapshot {
	switch {
	case s.capacity <= 0:
		return s.snapshots
	case len(s.snapshots) < s.capacity:
		return append(s.snapshots, snapshot)
	default:
		return append(s.snapshots[1:], snapshot)
	}
}
