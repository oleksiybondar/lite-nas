package processor

import (
	"context"
	"testing"

	"lite-nas/shared/fsm"
)

// Requirements: security-monitoring/IR-002.
func TestLifecycleMachineRoutesNormalization(t *testing.T) {
	t.Parallel()
	result, err := lifecycleMachine.Step(context.Background(), fsm.Instance[lifecycleState, struct{}]{State: lifecycleActive}, lifecycleCleared)
	if err != nil {
		t.Fatalf("step: %v", err)
	}
	if len(result.Effects) != 1 || result.Effects[0] != effectNormalize {
		t.Fatalf("effects = %v", result.Effects)
	}
}
