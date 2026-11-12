package historytest

import (
	"reflect"
	"testing"
)

// Store describes the bounded-history behavior shared by metrics history
// stores across services.
type Store[T any] interface {
	Add(T)
	Latest() (T, bool)
	List() []T
}

// RunDropsOldestAtCapacityCase verifies that a store evicts the oldest
// snapshot once it reaches capacity.
func RunDropsOldestAtCapacityCase[T any](
	t *testing.T,
	newStore func(int) Store[T],
	snapshots []T,
) {
	t.Helper()

	store := newStore(2)
	for _, snapshot := range snapshots {
		store.Add(snapshot)
	}

	want := snapshots[1:]
	if got := store.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
}

// RunLatestReturnsMostRecentCase verifies that Latest returns the last added
// snapshot.
func RunLatestReturnsMostRecentCase[T any](
	t *testing.T,
	newStore func(int) Store[T],
	snapshot T,
) {
	t.Helper()

	store := newStore(2)
	store.Add(snapshot)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", got, snapshot)
	}
}

// RunLatestRemainsAvailableWhenHistoryRetentionIsDisabledCase verifies that
// Latest still returns the most recent snapshot even when retained history is
// disabled.
func RunLatestRemainsAvailableWhenHistoryRetentionIsDisabledCase[T any](
	t *testing.T,
	newStore func(int) Store[T],
	snapshot T,
) {
	t.Helper()

	store := newStore(0)
	store.Add(snapshot)

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", got, snapshot)
	}
	if history := store.List(); len(history) != 0 {
		t.Fatalf("len(List()) = %d, want 0", len(history))
	}
}

// RunLatestReturnsFalseWhenEmptyCase verifies that Latest reports an empty
// store when no snapshots have been added.
func RunLatestReturnsFalseWhenEmptyCase[T any](
	t *testing.T,
	newStore func(int) Store[T],
	zero T,
) {
	t.Helper()

	store := newStore(1)
	got, ok := store.Latest()
	if ok {
		t.Fatalf("Latest() = %#v, want empty result", got)
	}
}

// RunListReturnsCopyCase verifies that List returns a detached slice copy.
func RunListReturnsCopyCase[T any](
	t *testing.T,
	newStore func(int) Store[T],
	snapshot T,
	mutated T,
) {
	t.Helper()

	store := newStore(1)
	store.Add(snapshot)

	history := store.List()
	history[0] = mutated

	got, ok := store.Latest()
	if !ok {
		t.Fatal("expected latest snapshot")
	}
	if !reflect.DeepEqual(got, snapshot) {
		t.Fatalf("Latest() = %#v, want %#v", got, snapshot)
	}
}
