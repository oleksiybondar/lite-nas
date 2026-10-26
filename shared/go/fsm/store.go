package fsm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DueInput identifies a keyed timeout to feed through Apply like any other
// input. Callers normally enumerate these on ticks from workers.TimerWorker.
type DueInput[I comparable] struct {
	Key   string
	Input I
}

// Deadline returns the next timeout input for an instance, if any.
type Deadline[S comparable, I comparable, D any] func(Instance[S, D]) (time.Time, I, bool)

// MemoryStore holds bounded volatile instances. State is lost on restart;
// callers with durable incident or command requirements need another store.
// Effect intents returned from Apply must be accounted for by the caller.
type MemoryStore[S comparable, I comparable, D any, E any] struct {
	mu       sync.RWMutex
	entries  map[string]*memoryEntry[S, D]
	machine  *Machine[S, I, D, E]
	newData  func(string) D
	clock    func() time.Time
	deadline Deadline[S, I, D]
	maxKeys  int
	maxIdle  time.Duration
}

type memoryEntry[S comparable, D any] struct {
	mu       sync.Mutex
	instance Instance[S, D]
	lastUsed time.Time
}

// NewMemoryStore requires explicit capacity and idle expiry. The clock is
// injectable so timeout boundaries can be tested without sleeping.
func NewMemoryStore[S comparable, I comparable, D any, E any](
	machine *Machine[S, I, D, E], maxKeys int, maxIdle time.Duration,
	newData func(string) D, clock func() time.Time, deadline Deadline[S, I, D],
) (*MemoryStore[S, I, D, E], error) {
	if machine == nil || maxKeys < 1 || maxIdle <= 0 || newData == nil || clock == nil {
		return nil, errors.New("fsm memory store requires machine, capacity, idle expiry, data factory and clock")
	}
	return &MemoryStore[S, I, D, E]{
		entries: make(map[string]*memoryEntry[S, D]), machine: machine,
		maxKeys: maxKeys, maxIdle: maxIdle, newData: newData, clock: clock, deadline: deadline,
	}, nil
}

// Apply serializes one key's transition, but independent keys may advance
// concurrently. An empty key is rejected to prevent accidental collisions.
func (s *MemoryStore[S, I, D, E]) Apply(ctx context.Context, key string, input I) (Result[S, I, D, E], error) {
	return s.ApplyWith(ctx, key, input, nil)
}

// ApplyWith adds validated observation data under the same key lock used for
// transition evaluation. The prepare function must not perform side effects;
// its data is committed only if Step succeeds.
func (s *MemoryStore[S, I, D, E]) ApplyWith(ctx context.Context, key string, input I, prepare func(D) D) (Result[S, I, D, E], error) {
	var empty Result[S, I, D, E]
	if key == "" {
		return empty, errors.New("fsm instance key is required")
	}
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	var entry *memoryEntry[S, D]
	for {
		var err error
		entry, err = s.getOrCreate(key)
		if err != nil {
			return empty, err
		}
		s.mu.RLock()
		if s.entries[key] == entry {
			break
		}
		s.mu.RUnlock()
	}
	defer s.mu.RUnlock()
	entry.mu.Lock()
	defer entry.mu.Unlock()
	current := entry.instance
	if prepare != nil {
		current.Data = prepare(current.Data)
	}
	result, err := s.machine.Step(ctx, current, input)
	if err != nil {
		return result, fmt.Errorf("fsm instance %q: %w", key, err)
	}
	entry.instance = result.Next
	entry.lastUsed = s.clock()
	return result, nil
}

// Due enumerates timeout inputs for instances whose deadlines are at or
// before the injected clock. It does not alter state: callers pass each item
// to Apply, which re-evaluates the transition under that key's lock.
func (s *MemoryStore[S, I, D, E]) Due() []DueInput[I] {
	if s.deadline == nil {
		return nil
	}
	now := s.clock()
	s.mu.RLock()
	defer s.mu.RUnlock()
	var due []DueInput[I]
	for key, entry := range s.entries {
		entry.mu.Lock()
		deadline, input, exists := s.deadline(entry.instance)
		entry.mu.Unlock()
		if exists && !now.Before(deadline) {
			due = append(due, DueInput[I]{Key: key, Input: input})
		}
	}
	return due
}

// ExpireIdle removes inactive instances to release cardinality capacity.
// Callers must ensure expired instances have no unhandled external effects.
func (s *MemoryStore[S, I, D, E]) ExpireIdle() int {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for key, entry := range s.entries {
		entry.mu.Lock()
		idle := now.Sub(entry.lastUsed) >= s.maxIdle
		entry.mu.Unlock()
		if idle {
			delete(s.entries, key)
			removed++
		}
	}
	return removed
}

func (s *MemoryStore[S, I, D, E]) getOrCreate(key string) (*memoryEntry[S, D], error) {
	s.mu.RLock()
	entry := s.entries[key]
	s.mu.RUnlock()
	if entry != nil {
		return entry, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry = s.entries[key]; entry != nil {
		return entry, nil
	}
	if len(s.entries) >= s.maxKeys {
		return nil, fmt.Errorf("fsm instance capacity %d reached", s.maxKeys)
	}
	entry = &memoryEntry[S, D]{instance: s.machine.Initial(s.newData(key)), lastUsed: s.clock()}
	s.entries[key] = entry
	return entry, nil
}
