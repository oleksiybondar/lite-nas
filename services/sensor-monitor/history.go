package main

import (
	"sync"

	"lite-nas/shared/metrics"
)

// HistoryStore retains the latest snapshot and a bounded chronological history.
type HistoryStore struct {
	mu        sync.RWMutex
	latest    metrics.SensorSnapshot
	available bool
	items     []metrics.SensorSnapshot
	capacity  int
}

// NewHistoryStore allocates a bounded in-memory snapshot store.
func NewHistoryStore(capacity int) *HistoryStore {
	return &HistoryStore{capacity: capacity, items: make([]metrics.SensorSnapshot, 0, max(capacity, 0))}
}

// Add records a snapshot as latest and retains it if history is enabled.
func (store *HistoryStore) Add(snapshot metrics.SensorSnapshot) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.latest, store.available = snapshot, true
	if store.capacity <= 0 {
		return
	}
	if len(store.items) == store.capacity {
		store.items = append(store.items[1:], snapshot)
		return
	}
	store.items = append(store.items, snapshot)
}

// Latest returns the most recently collected snapshot, independent of retention.
func (store *HistoryStore) Latest() (metrics.SensorSnapshot, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.latest, store.available
}

// List returns retained snapshots in chronological order.
func (store *HistoryStore) List() []metrics.SensorSnapshot {
	store.mu.RLock()
	defer store.mu.RUnlock()
	items := make([]metrics.SensorSnapshot, len(store.items))
	copy(items, store.items)
	return items
}
