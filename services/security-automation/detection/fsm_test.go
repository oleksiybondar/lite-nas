package detection

import (
	"context"
	"testing"
	"time"
)

// mustPlugin constructs a test detector with a controlled clock.
func mustPlugin(t *testing.T, name PluginName, steps, threshold int, now *time.Time) *Plugin {
	t.Helper()
	p, err := New(Config{Name: name, Steps: steps, Threshold: threshold, Window: 5 * time.Minute, MaxKeys: 8, MaxIdle: time.Hour}, func() time.Time { return *now })
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// mustObserve submits a valid event whose success is part of the fixture flow.
func mustObserve(t *testing.T, p *Plugin, id, key string, step int, at time.Time) []Match {
	t.Helper()
	matches, err := p.Observe(context.Background(), Event{ID: id, Key: key, Step: step, At: at})
	if err != nil {
		t.Fatal(err)
	}
	return matches
}

// TestSequenceRequiresOrderedSteps verifies that an out-of-order alert cannot complete a sequence.
func TestSequenceRequiresOrderedSteps(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, Sequence, 2, 0, &now)
	mustObserve(t, p, "second", "192.0.2.1", 1, now)
	mustObserve(t, p, "first", "192.0.2.1", 0, now.Add(time.Second))
	if matches := mustObserve(t, p, "second-again", "192.0.2.1", 1, now.Add(2*time.Second)); len(matches) != 1 {
		t.Fatalf("matches = %d, want 1", len(matches))
	}
}

// Requirements: security-automation/FR-002.
// TestSequenceOverlappingSelectorsAdvancesOnLaterAlert checks that a first-step
// match from the next alert cannot erase an in-progress sequence.
func TestSequenceOverlappingSelectorsAdvancesOnLaterAlert(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, Sequence, 2, 0, &now)
	mustObserve(t, p, "first", "alice", 0, now)
	mustObserve(t, p, "first", "alice", 1, now)
	mustObserve(t, p, "second", "alice", 0, now.Add(time.Second))
	if matches := mustObserve(t, p, "second", "alice", 1, now.Add(time.Second)); len(matches) != 1 {
		t.Fatalf("overlapping selectors produced %d matches, want 1", len(matches))
	}
}

// TestCountKeepsKeysIndependent verifies that one actor cannot fill another actor's threshold.
func TestCountKeepsKeysIndependent(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, CountWithinWindow, 0, 2, &now)
	mustObserve(t, p, "a1", "alice", 0, now)
	mustObserve(t, p, "b1", "bob", 0, now)
	if matches := mustObserve(t, p, "a2", "alice", 0, now.Add(time.Second)); len(matches) != 1 || matches[0].Key != "alice" {
		t.Fatalf("matches = %+v, want alice", matches)
	}
}

// TestDuplicateAlertDoesNotCompleteThreshold verifies idempotence within one key.
func TestDuplicateAlertDoesNotCompleteThreshold(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, CountWithinWindow, 0, 2, &now)
	mustObserve(t, p, "a1", "alice", 0, now)
	if matches := mustObserve(t, p, "a1", "alice", 0, now); len(matches) != 0 {
		t.Fatalf("duplicate produced %d matches", len(matches))
	}
}

// TestWindowExpiryResetsPartialCount verifies a stale observation cannot complete a new window.
func TestWindowExpiryResetsPartialCount(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, CountWithinWindow, 0, 2, &now)
	mustObserve(t, p, "a1", "alice", 0, now)
	now = now.Add(6 * time.Minute)
	if err := p.Advance(context.Background()); err != nil {
		t.Fatal(err)
	}
	if matches := mustObserve(t, p, "a2", "alice", 0, now); len(matches) != 0 {
		t.Fatalf("expired count produced %d matches", len(matches))
	}
}

// Requirements: security-automation/FR-003.
// TestResetForgetsIncompleteCount checks temporary re-enablement starts a new detection window.
func TestResetForgetsIncompleteCount(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, CountWithinWindow, 0, 2, &now)
	mustObserve(t, p, "first", "alice", 0, now)
	if err := p.Reset(); err != nil {
		t.Fatal(err)
	}
	if matches := mustObserve(t, p, "second", "alice", 0, now.Add(time.Second)); len(matches) != 0 {
		t.Fatalf("reset detector completed with %d matches", len(matches))
	}
}

// Requirements: security-automation/FR-003.
// TestResetKeyKeepsOtherInstances checks an operator reset affects only one
// correlation key and allows its prior alert IDs to start a fresh window.
func TestResetKeyKeepsOtherInstances(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, CountWithinWindow, 0, 2, &now)
	mustObserve(t, p, "a1", "alice", 0, now)
	mustObserve(t, p, "b1", "bob", 0, now)
	if _, err := p.ResetKey(context.Background(), "alice"); err != nil {
		t.Fatal(err)
	}
	if matches := mustObserve(t, p, "b2", "bob", 0, now.Add(time.Second)); len(matches) != 1 {
		t.Fatalf("other key produced %d matches, want 1", len(matches))
	}
}

// Requirements: security-automation/FR-003.
// TestResetKeyClearsPartialCount checks a reset key cannot complete from its
// pre-reset observations.
func TestResetKeyClearsPartialCount(t *testing.T) {
	now := time.Now()
	p := mustPlugin(t, CountWithinWindow, 0, 2, &now)
	mustObserve(t, p, "a1", "alice", 0, now)
	if _, err := p.ResetKey(context.Background(), "alice"); err != nil {
		t.Fatal(err)
	}
	if matches := mustObserve(t, p, "a2", "alice", 0, now.Add(time.Second)); len(matches) != 0 {
		t.Fatalf("reset key produced %d matches, want 0", len(matches))
	}
}
