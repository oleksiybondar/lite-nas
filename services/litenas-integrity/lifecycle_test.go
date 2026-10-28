package main

import (
	"context"
	"testing"
)

// Requirements: security-monitoring/IR-002.
func TestIntegrityLifecycleRoutesRepeatedFailureToOccurrence(t *testing.T) {
	t.Parallel()
	effect, err := integrityTransition(context.Background(), true, integrityFailure)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if effect != integrityOccurrence {
		t.Fatalf("effect = %s", effect)
	}
}
