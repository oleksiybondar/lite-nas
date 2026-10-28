package fsm

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type (
	testState string
	testInput string
)

const (
	idle     testState = "idle"
	counting testState = "counting"
	cooldown testState = "cooldown"
	observe  testInput = "observe"
	expire   testInput = "expire"
)

type testData struct {
	count    int
	deadline time.Time
}

func thresholdDefinition(now func() time.Time) Definition[testState, testInput, testData, string] {
	return Definition[testState, testInput, testData, string]{
		Initial: idle,
		States:  []testState{idle, counting, cooldown},
		Inputs:  []testInput{observe, expire},
		Rows: []Row[testState, testInput]{
			{From: idle, On: observe, To: counting, Action: "first"},
			{From: counting, On: observe, Guard: "below", To: counting, Action: "increment"},
			{From: counting, On: observe, Guard: "reached", To: cooldown, Action: "detect"},
			{From: counting, On: expire, To: idle, Action: "clear"},
			{From: cooldown, On: observe, To: cooldown},
			{From: cooldown, On: expire, To: idle, Action: "clear"},
		},
		Guards: map[string]Guard[testData, testInput]{
			"below":   func(_ context.Context, data testData, _ testInput) (bool, error) { return data.count < 2, nil },
			"reached": func(_ context.Context, data testData, _ testInput) (bool, error) { return data.count >= 2, nil },
		},
		Actions: map[string]Action[testData, testInput, string]{
			"first": func(_ context.Context, data testData, _ testInput) (testData, []string, error) {
				data.count = 1
				data.deadline = now().Add(time.Minute)
				return data, nil, nil
			},
			"increment": func(_ context.Context, data testData, _ testInput) (testData, []string, error) {
				data.count++
				return data, nil, nil
			},
			"detect": func(_ context.Context, data testData, _ testInput) (testData, []string, error) {
				data.count++
				data.deadline = now().Add(time.Minute)
				return data, []string{"alert"}, nil
			},
			"clear": func(_ context.Context, _ testData, _ testInput) (testData, []string, error) {
				return testData{}, nil, nil
			},
		},
	}
}

func mustThresholdMachine(t *testing.T, now func() time.Time) *Machine[testState, testInput, testData, string] {
	t.Helper()
	machine, err := New(thresholdDefinition(now))
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}
	return machine
}

func mustApply(t *testing.T, store *MemoryStore[testState, testInput, testData, string], key string, input testInput) Result[testState, testInput, testData, string] {
	t.Helper()
	result, err := store.Apply(context.Background(), key, input)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	return result
}

func mustThresholdStore(t *testing.T, now func() time.Time, maxKeys int) *MemoryStore[testState, testInput, testData, string] {
	t.Helper()
	store, err := NewMemoryStore(mustThresholdMachine(t, now), maxKeys, time.Hour,
		func(string) testData { return testData{} }, now,
		func(instance Instance[testState, testData]) (time.Time, testInput, bool) {
			return instance.Data.deadline, expire, !instance.Data.deadline.IsZero()
		})
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	return store
}

// Requirements: security-monitoring/IR-001.
func TestThresholdCountsCrossingObservationOnce(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 2)
	mustApply(t, store, "rule/a", observe)
	mustApply(t, store, "rule/a", observe)
	result := mustApply(t, store, "rule/a", observe)
	if result.Next.Data.count != 3 {
		t.Fatalf("count = %d", result.Next.Data.count)
	}
}

// Requirements: security-monitoring/IR-001.
func TestThresholdEmitsOneAlert(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 2)
	mustApply(t, store, "rule/a", observe)
	mustApply(t, store, "rule/a", observe)
	result := mustApply(t, store, "rule/a", observe)
	if len(result.Effects) != 1 {
		t.Fatalf("effects = %v", result.Effects)
	}
}

// Requirements: security-monitoring/IR-001.
func TestThresholdCooldownSuppressesRepeat(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 2)
	for range 3 {
		mustApply(t, store, "rule/a", observe)
	}
	result := mustApply(t, store, "rule/a", observe)
	if len(result.Effects) != 0 {
		t.Fatalf("effects = %v", result.Effects)
	}
}

// Requirements: security-monitoring/IR-001.
func TestThresholdKeysRemainIndependent(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 2)
	for range 3 {
		mustApply(t, store, "rule/a", observe)
	}
	result := mustApply(t, store, "rule/b", observe)
	if result.Next.State != counting {
		t.Fatalf("state = %s", result.Next.State)
	}
}

// Requirements: security-monitoring/IR-001.
func TestDeadlineIsOrdinaryInput(t *testing.T) {
	t.Parallel()
	now := time.Now()
	clock := func() time.Time { return now }
	store := mustThresholdStore(t, clock, 2)
	mustApply(t, store, "rule/a", observe)
	now = now.Add(time.Minute)
	due := store.Due()
	if len(due) != 1 || due[0].Input != expire {
		t.Fatalf("due = %v", due)
	}
}

// Requirements: security-monitoring/IR-001.
func TestExpiryResetsCounter(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 2)
	mustApply(t, store, "rule/a", observe)
	result := mustApply(t, store, "rule/a", expire)
	if result.Next.Data.count != 0 {
		t.Fatalf("count = %d", result.Next.Data.count)
	}
}

// Requirements: security-monitoring/IR-001.
func TestCapacityRejectsExtraKey(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 1)
	mustApply(t, store, "rule/a", observe)
	if _, err := store.Apply(context.Background(), "rule/b", observe); err == nil {
		t.Fatal("capacity not enforced")
	}
}

// Requirements: security-monitoring/IR-001.
func TestConcurrentSameKeySerializes(t *testing.T) {
	t.Parallel()
	store := mustThresholdStore(t, time.Now, 1)
	var group sync.WaitGroup
	for range 20 {
		group.Add(1)
		go func() { defer group.Done(); _, _ = store.Apply(context.Background(), "rule/a", observe) }()
	}
	group.Wait()
	result := mustApply(t, store, "rule/a", observe)
	if result.Next.State != cooldown {
		t.Fatalf("state = %s", result.Next.State)
	}
}

// Requirements: security-monitoring/IR-001.
func TestAmbiguousUnguardedRowsRejected(t *testing.T) {
	t.Parallel()
	definition := thresholdDefinition(time.Now)
	definition.Rows = append(definition.Rows, Row[testState, testInput]{From: idle, On: observe, To: idle, Guard: "below"})
	if _, err := New(definition); err == nil {
		t.Fatal("ambiguous rows accepted")
	}
}

// Requirements: security-monitoring/IR-001.
func TestMissingGuardRejected(t *testing.T) {
	t.Parallel()
	definition := thresholdDefinition(time.Now)
	definition.Rows[0].Guard = "missing"
	if _, err := New(definition); err == nil {
		t.Fatal("missing guard accepted")
	}
}

// Requirements: security-monitoring/IR-001.
func TestOverlappingGuardsFailAtRuntime(t *testing.T) {
	t.Parallel()
	definition := thresholdDefinition(time.Now)
	definition.Guards["below"] = func(context.Context, testData, testInput) (bool, error) { return true, nil }
	definition.Guards["reached"] = func(context.Context, testData, testInput) (bool, error) { return true, nil }
	machine, err := New(definition)
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}
	_, err = machine.Step(context.Background(), Instance[testState, testData]{State: counting}, observe)
	if err == nil {
		t.Fatal("overlapping guards accepted")
	}
}

// Requirements: security-monitoring/IR-001.
func TestGuardErrorPreservesState(t *testing.T) {
	t.Parallel()
	definition := thresholdDefinition(time.Now)
	definition.Guards["below"] = func(context.Context, testData, testInput) (bool, error) { return false, errors.New("guard failed") }
	machine, err := New(definition)
	if err != nil {
		t.Fatalf("new machine: %v", err)
	}
	current := Instance[testState, testData]{State: counting, Data: testData{count: 1}}
	result, _ := machine.Step(context.Background(), current, observe)
	if result.Next != current {
		t.Fatalf("state changed after guard error: %+v", result.Next)
	}
}
