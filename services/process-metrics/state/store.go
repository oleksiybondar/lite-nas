package state

import (
	"sync"

	"lite-nas/shared/metrics"
)

// SnapshotStore keeps the latest collected process snapshot.
type SnapshotStore struct {
	mu      sync.RWMutex
	latest  metrics.ProcessMetricsSnapshot
	hasData bool
}

// NewSnapshotStore creates an empty live-snapshot store.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{}
}

// Add stores the next latest snapshot.
func (s *SnapshotStore) Add(snapshot metrics.ProcessMetricsSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.latest = snapshot
	s.hasData = true
}

// Latest returns the most recent snapshot and whether one is available.
func (s *SnapshotStore) Latest() (metrics.ProcessMetricsSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.hasData {
		return metrics.ProcessMetricsSnapshot{}, false
	}

	return s.latest, true
}
