package main

import (
	"context"
	"errors"
	"testing"
	"time"

	rbaccache "lite-nas/services/rbac/cache"
)

type failRunner struct{}

func (failRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return nil, errors.New("runner should not be called")
}

func TestInvalidateCacheByUIDAndAll(t *testing.T) {
	t.Parallel()

	store := rbaccache.NewStore(make(chan struct{}, 1))
	store.Set("1001", "/usr/bin/zfs", rbaccache.Entry{Allowed: true, ExpiresAt: time.Now().Add(time.Hour)})
	store.Set("1002", "/usr/bin/zfs", rbaccache.Entry{Allowed: true, ExpiresAt: time.Now().Add(time.Hour)})

	service := newDecisionService(store, failRunner{}, time.Hour, time.Hour)
	service.InvalidateCache("1001")
	if _, ok := store.Get("1001", "/usr/bin/zfs"); ok {
		t.Fatalf("expected UID cache entry to be invalidated")
	}
	if _, ok := store.Get("1002", "/usr/bin/zfs"); !ok {
		t.Fatalf("expected other UID cache entry to remain")
	}

	service.InvalidateCache("")
	if _, ok := store.Get("1002", "/usr/bin/zfs"); ok {
		t.Fatalf("expected all cache entries to be invalidated")
	}
}

func TestCanSudoExecUsesValidCacheEntry(t *testing.T) {
	t.Parallel()

	store := rbaccache.NewStore(make(chan struct{}, 1))
	store.Set("1001", "/usr/bin/zfs", rbaccache.Entry{Allowed: true, ExpiresAt: time.Now().Add(time.Hour)})
	service := newDecisionService(store, failRunner{}, time.Hour, time.Hour)

	if !service.CanSudoExec(t.Context(), "1001", "/usr/bin/zfs") {
		t.Fatalf("CanSudoExec() expected cached allow decision")
	}
}
