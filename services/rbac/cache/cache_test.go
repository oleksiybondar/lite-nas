package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEntryIsValid(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_700_000_000, 0)
	valid := Entry{Allowed: true, ExpiresAt: now.Add(10 * time.Second)}
	expired := Entry{Allowed: false, ExpiresAt: now}

	if !valid.IsValid(now) {
		t.Fatalf("valid entry marked as invalid")
	}

	if expired.IsValid(now) {
		t.Fatalf("expired entry marked as valid")
	}
}

func TestStoreSetAndGet(t *testing.T) {
	t.Parallel()

	store := NewStore(make(chan struct{}))
	expected := Entry{
		Allowed:   true,
		ExpiresAt: time.Unix(1_700_000_100, 0),
	}

	store.Set(1000, "/usr/bin/zfs", expected)

	got, ok := store.Get(1000, "/usr/bin/zfs")
	if !ok {
		t.Fatalf("Get() did not find expected entry")
	}

	if got != expected {
		t.Fatalf("Get() entry = %#v, want %#v", got, expected)
	}

	if _, ok = store.Get(1000, "/usr/bin/missing"); ok {
		t.Fatalf("Get() unexpectedly found missing command key")
	}
}

func TestInvalidateExpiredRemovesOnlyExpiredEntries(t *testing.T) {
	t.Parallel()

	store := NewStore(make(chan struct{}))
	now := time.Unix(1_700_000_000, 0)

	store.Set(1000, "expired", Entry{Allowed: false, ExpiresAt: now})
	store.Set(1000, "valid", Entry{Allowed: true, ExpiresAt: now.Add(time.Minute)})
	store.Set(1001, "valid-other", Entry{Allowed: true, ExpiresAt: now.Add(time.Minute)})

	store.InvalidateExpired(now)

	if _, ok := store.Get(1000, "expired"); ok {
		t.Fatalf("expired entry was not removed")
	}

	if _, ok := store.Get(1000, "valid"); !ok {
		t.Fatalf("valid entry was removed")
	}

	if _, ok := store.Get(1001, "valid-other"); !ok {
		t.Fatalf("valid entry for other UID was removed")
	}
}

func TestInvalidateUIDRemovesOnlySingleUID(t *testing.T) {
	t.Parallel()

	store := storeWithTwoUIDEntries()

	store.InvalidateUID(1000)

	if _, ok := store.Get(1000, "command"); ok {
		t.Fatalf("uid-specific invalidation did not remove UID 1000 entries")
	}

	if _, ok := store.Get(1001, "command"); !ok {
		t.Fatalf("uid-specific invalidation removed other UID entries")
	}
}

func TestInvalidateAllRemovesEverything(t *testing.T) {
	t.Parallel()

	store := storeWithTwoUIDEntries()

	store.InvalidateAll()

	if _, ok := store.Get(1000, "command"); ok {
		t.Fatalf("InvalidateAll() did not remove UID 1000 entry")
	}

	if _, ok := store.Get(1001, "command"); ok {
		t.Fatalf("InvalidateAll() did not remove UID 1001 entry")
	}
}

func TestRunInvalidationWorkerRemovesExpiredOnSignal(t *testing.T) {
	t.Parallel()

	invalidateCh := make(chan struct{}, 1)
	store := NewStore(invalidateCh)

	now := time.Now()
	store.Set(1000, "expired", Entry{Allowed: false, ExpiresAt: now.Add(-time.Second)})
	store.Set(1000, "valid", Entry{Allowed: true, ExpiresAt: now.Add(time.Minute)})

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- store.RunInvalidationWorker(ctx)
	}()

	invalidateCh <- struct{}{}

	requireEntryMissing(t, store, 1000, "expired")
	requireEntryPresent(t, store, 1000, "valid")

	cancel()
	err := <-done
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunInvalidationWorker() error = %v, want %v", err, context.Canceled)
	}
}

func requireEntryPresent(t *testing.T, store *Store, uid uint32, commandKey string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := store.Get(uid, commandKey); ok {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("entry %q for uid %d was not present before timeout", commandKey, uid)
}

func requireEntryMissing(t *testing.T, store *Store, uid uint32, commandKey string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := store.Get(uid, commandKey); !ok {
			return
		}

		time.Sleep(5 * time.Millisecond)
	}

	t.Fatalf("entry %q for uid %d was still present before timeout", commandKey, uid)
}

func storeWithTwoUIDEntries() *Store {
	store := NewStore(make(chan struct{}))
	entry := Entry{Allowed: true, ExpiresAt: time.Unix(1_700_000_100, 0)}
	store.Set(1000, "command", entry)
	store.Set(1001, "command", entry)
	return store
}
