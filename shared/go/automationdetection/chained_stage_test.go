package automationdetection

import (
	"context"
	"testing"
	"time"
)

// mustChainStage constructs the table used by a deterministic scenario.
func mustChainStage(t *testing.T) *ChainStage {
	t.Helper()
	stage, err := NewChainStage(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return stage
}

// mustStageStep evaluates and commits one simulated verified transition.
func mustStageStep(t *testing.T, stage *ChainStage, state *ChainStageResult, snapshot ChainStageSnapshot) ChainStageResult {
	t.Helper()
	result, err := stage.Evaluate(context.Background(), state.Next, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	*state = result
	return result
}

// Requirements: operational-safety-automation/FR-CHAIN-001.
func TestChainStageWaitsForParentAlert(t *testing.T) {
	stage := mustChainStage(t)
	state := ChainStageResult{Next: stage.Initial()}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, SourceActive: true})
	if state.Next.State != ChainStageIdle || len(state.Effects) != 0 {
		t.Fatalf("stage without parent = %q, effects=%v", state.Next.State, state.Effects)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-002.
func TestChainStageRestoresOnlyAfterClearAndCooldown(t *testing.T) {
	stage := mustChainStage(t)
	state := ChainStageResult{Next: stage.Initial()}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, SourceActive: true, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now.Add(59 * time.Second), ParentActive: true})
	if got := mustStageStep(t, stage, &state, ChainStageSnapshot{At: now.Add(time.Minute), ParentActive: true}); got.Next.State != ChainStageRestoring || len(got.Effects) != 1 || got.Effects[0] != ChainStageRestore {
		t.Fatalf("due cooldown transition = %#v", got)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-003.
func TestChainStageRenewedBadValueResetsCooldown(t *testing.T) {
	stage := mustChainStage(t)
	state := ChainStageResult{Next: stage.Initial()}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, SourceActive: true, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now.Add(30 * time.Second), SourceActive: true, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now.Add(40 * time.Second), ParentActive: true})
	got := mustStageStep(t, stage, &state, ChainStageSnapshot{At: now.Add(time.Minute), ParentActive: true})
	if got.Next.State != ChainStageCooling || len(got.Effects) != 0 || !got.Next.Data.Deadline.Equal(now.Add(100*time.Second)) {
		t.Fatalf("renewed-value cooldown = %#v", got)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-004.
func TestChainStageChildAlertBlocksParentRecovery(t *testing.T) {
	stage := mustChainStage(t)
	state := ChainStageResult{Next: stage.Initial()}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, SourceActive: true, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, ParentActive: true, ChildActive: true})
	got := mustStageStep(t, stage, &state, ChainStageSnapshot{At: now.Add(2 * time.Minute), ParentActive: true, ChildActive: true})
	if got.Next.State != ChainStageActive || len(got.Effects) != 0 {
		t.Fatalf("parent with child active = %#v", got)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-005.
func TestChainStageNormalizesOnlyAfterRestoreVerification(t *testing.T) {
	stage := mustChainStage(t)
	state := ChainStageResult{Next: stage.Initial()}
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, SourceActive: true, ParentActive: true})
	mustStageStep(t, stage, &state, ChainStageSnapshot{At: now, ParentActive: true})
	result, err := stage.Evaluate(context.Background(), state.Next, ChainStageSnapshot{At: now.Add(time.Minute), ParentActive: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stage.Finish(context.Background(), state.Next); err == nil || result.Next.State != ChainStageRestoring {
		t.Fatalf("unverified stage finished; result=%#v error=%v", result, err)
	}
}
