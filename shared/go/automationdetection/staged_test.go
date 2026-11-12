package automationdetection

import (
	"context"
	"testing"
	"time"
)

// mustRecovery creates the staged machine for one controlled timeline.
func mustRecovery(t *testing.T) *StagedRecovery {
	t.Helper()
	machine, err := NewStagedRecovery(time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return machine
}

// mustTransition evaluates and commits one verified transition.
func mustTransition(t *testing.T, machine *StagedRecovery, state *RecoveryResult, input RecoveryInput) []RecoveryEffect {
	t.Helper()
	result, err := machine.Step(context.Background(), state.Next, input)
	if err != nil {
		t.Fatal(err)
	}
	*state = result
	return result.Effects
}

// TestStagedRecoveryNeedsThreeCooldowns verifies critical recovery crosses each declared level.
func TestStagedRecoveryNeedsThreeCooldowns(t *testing.T) {
	machine := mustRecovery(t)
	state := RecoveryResult{Next: machine.Initial()}
	start := time.Now()
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerActive, AlertID: "cpu", Level: LevelCritical, At: start})
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerNormalized, AlertID: "cpu", At: start.Add(time.Second)})
	for step, want := range []RecoveryLevel{LevelError, LevelWarning, LevelNormal} {
		effects := mustTransition(t, machine, &state, RecoveryInput{Kind: CooldownTick, At: start.Add(time.Duration(step+1)*time.Minute + time.Second)})
		if len(effects) != 1 || effects[0].Level != want {
			t.Fatalf("step %d effects = %#v, want level %d", step, effects, want)
		}
	}
}

// TestActiveTriggerBlocksRecovery verifies a still-active alert prevents relaxation.
func TestActiveTriggerBlocksRecovery(t *testing.T) {
	machine := mustRecovery(t)
	state := RecoveryResult{Next: machine.Initial()}
	start := time.Now()
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerActive, AlertID: "power", Level: LevelCritical, At: start})
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerActive, AlertID: "heat", Level: LevelCritical, At: start})
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerNormalized, AlertID: "power", At: start.Add(time.Second)})
	if effects := mustTransition(t, machine, &state, RecoveryInput{Kind: CooldownTick, At: start.Add(2 * time.Minute)}); len(effects) != 0 {
		t.Fatalf("active critical trigger allowed relaxation: %#v", effects)
	}
}

// TestLowerSeverityDuplicateCannotRelax verifies repeated alert creation cannot reduce its tracked severity.
func TestLowerSeverityDuplicateCannotRelax(t *testing.T) {
	machine := mustRecovery(t)
	state := RecoveryResult{Next: machine.Initial()}
	start := time.Now()
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerActive, AlertID: "power", Level: LevelCritical, At: start})
	mustTransition(t, machine, &state, RecoveryInput{Kind: TriggerActive, AlertID: "power", Level: LevelWarning, At: start.Add(time.Second)})
	if state.Next.Data.Active["power"] != LevelCritical {
		t.Fatalf("tracked severity = %d, want critical", state.Next.Data.Active["power"])
	}
}
