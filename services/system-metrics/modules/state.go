package modules

import (
	"sync"

	"lite-nas/services/system-metrics/state"
	"lite-nas/shared/metrics"
)

// State groups runtime-owned in-memory service state.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only handles to mutable runtime state.
type State struct {
	SnapshotStore *SnapshotStore
}

// SnapshotStore wraps the bounded history store with concurrency protection.
type SnapshotStore struct {
	mu      sync.RWMutex
	history state.HistoryStore
}

// NewStateModule creates the in-memory state owned by the service runtime.
//
// Parameters:
//   - historySize: maximum number of snapshots retained in memory
func NewStateModule(historySize int) State {
	return State{
		SnapshotStore: &SnapshotStore{
			history: state.NewHistoryStore(historySize),
		},
	}
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
