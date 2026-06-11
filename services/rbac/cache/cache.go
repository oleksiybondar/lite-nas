package cache

import (
	"context"
	"sync"
	"time"
)

// Store keeps sudo decision cache entries indexed by UID and normalized command key.
type Store struct {
	mutex        sync.RWMutex
	entries      map[string]map[string]Entry
	invalidateCh <-chan struct{}
}

// NewStore constructs a cache store with a channel used to trigger TTL invalidation passes.
func NewStore(invalidateCh <-chan struct{}) *Store {
	return &Store{
		entries:      make(map[string]map[string]Entry),
		invalidateCh: invalidateCh,
	}
}

// Get returns the cached entry for UID and command key when present.
func (store *Store) Get(uid string, commandKey string) (Entry, bool) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	byCommand, ok := store.entries[uid]
	if !ok {
		return Entry{}, false
	}

	entry, found := byCommand[commandKey]
	if !found {
		return Entry{}, false
	}

	return entry, true
}

// Set stores or replaces a cached entry for UID and command key.
func (store *Store) Set(uid string, commandKey string, entry Entry) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	byCommand, ok := store.entries[uid]
	if !ok {
		byCommand = make(map[string]Entry)
		store.entries[uid] = byCommand
	}

	byCommand[commandKey] = entry
}

// InvalidateUID removes all cached decisions for a single UID.
func (store *Store) InvalidateUID(uid string) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	delete(store.entries, uid)
}

// InvalidateAll removes all cached decisions from the store.
func (store *Store) InvalidateAll() {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.entries = make(map[string]map[string]Entry)
}

// InvalidateExpired removes entries that are no longer valid at the provided time.
func (store *Store) InvalidateExpired(now time.Time) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	for uid, byCommand := range store.entries {
		invalidateExpiredByCommand(byCommand, now)
		deleteUIDIfEmpty(store.entries, uid, byCommand)
	}
}

func invalidateExpiredByCommand(byCommand map[string]Entry, now time.Time) {
	for commandKey, entry := range byCommand {
		if entry.IsValid(now) {
			continue
		}
		delete(byCommand, commandKey)
	}
}

func deleteUIDIfEmpty(entries map[string]map[string]Entry, uid string, byCommand map[string]Entry) {
	if len(byCommand) != 0 {
		return
	}
	delete(entries, uid)
}

// RunInvalidationWorker listens for invalidation signals and removes expired entries.
func (store *Store) RunInvalidationWorker(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-store.invalidateCh:
			if !ok {
				return nil
			}

			store.InvalidateExpired(time.Now())
		}
	}
}
