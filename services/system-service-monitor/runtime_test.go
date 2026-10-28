package main

import (
	"regexp"
	"testing"
)

// TestReadableEventID verifies the producer's domain prefix and decimal counter.
func TestReadableEventID(t *testing.T) {
	t.Parallel()
	if got := newStateEventID(); !regexp.MustCompile(`^syssvc_[0-9]{8}$`).MatchString(got) {
		t.Fatalf("unexpected event ID: %q", got)
	}
}
