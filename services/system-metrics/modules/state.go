package modules

import (
	"sync"

	"lite-nas/services/system-metrics/state"
	"lite-nas/shared/metrics"
)

// State groups runtime-owned in-memory service state.
type State struct {
	snapshotStore *SnapshotStore
}

// SnapshotStore wraps the bounded history store with concurrency protection.
type SnapshotStore struct {
	mu      sync.RWMutex
	history state.HistoryStore
}

// NewStateModule creates the runtime state module.
func NewStateModule(historySize int) State {
	return State{
		snapshotStore: &SnapshotStore{
			history: state.NewHistoryStore(historySize),
		},
	}
}

// SnapshotStore returns the snapshot store.
func (m State) SnapshotStore() *SnapshotStore {
	return m.snapshotStore
}

// Add stores the next snapshot.
func (s *SnapshotStore) Add(snapshot metrics.SystemSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history.Add(snapshot)
}

// Latest returns the latest stored snapshot.
func (s *SnapshotStore) Latest() (metrics.SystemSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.history.Latest()
}

// List returns the full stored history.
func (s *SnapshotStore) List() []metrics.SystemSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.history.List()
}
