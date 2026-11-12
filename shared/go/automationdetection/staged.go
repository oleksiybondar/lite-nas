package automationdetection

import (
	"context"
	"errors"
	"time"

	"lite-nas/shared/fsm"
)

// RecoveryLevel orders an automation's configured effect levels from normal to critical.
type RecoveryLevel uint8

const (
	LevelNormal RecoveryLevel = iota
	LevelWarning
	LevelError
	LevelCritical
)

// RecoveryInputKind identifies a trigger, its normalization, or a cooldown tick.
type RecoveryInputKind string

const (
	TriggerActive     RecoveryInputKind = "active"
	TriggerNormalized RecoveryInputKind = "normalized"
	CooldownTick      RecoveryInputKind = "cooldown"
)

// RecoveryInput is one validated alert lifecycle observation or timer tick.
type RecoveryInput struct {
	Kind    RecoveryInputKind
	AlertID string
	Level   RecoveryLevel
	At      time.Time
}

// RecoveryData records active alert levels and the next allowed relaxation time.
// The caller must persist this data with its verified effect state.
type RecoveryData struct {
	Active   map[string]RecoveryLevel
	Deadline time.Time
	input    RecoveryInput
	applied  RecoveryLevel
}

// RecoveryEffect asks the owner to apply and verify one effect level.
type RecoveryEffect struct {
	Level RecoveryLevel
}

// RecoveryResult is an uncommitted transition. Commit Next only after any
// returned effect is durably recorded, applied, and verified.
type RecoveryResult = fsm.Result[RecoveryLevel, RecoveryInputKind, RecoveryData, RecoveryEffect]

// StagedRecovery evaluates severity escalation and one-level cooldown recovery.
// One instance belongs to one configured rule and correlation key.
type StagedRecovery struct {
	machine  *fsm.Machine[RecoveryLevel, RecoveryInputKind, RecoveryData, RecoveryEffect]
	cooldown time.Duration
}

// NewStagedRecovery compiles the table for one positive cooldown interval.
func NewStagedRecovery(cooldown time.Duration) (*StagedRecovery, error) {
	if cooldown <= 0 {
		return nil, errors.New("recovery cooldown must be positive")
	}
	plugin := &StagedRecovery{cooldown: cooldown}
	states := []RecoveryLevel{LevelNormal, LevelWarning, LevelError, LevelCritical}
	rows := make([]fsm.Row[RecoveryLevel, RecoveryInputKind], 0, 35)
	guards := make(map[string]fsm.Guard[RecoveryData, RecoveryInputKind])
	for _, from := range states {
		for _, to := range states {
			name := guardName(from, to)
			target := to
			current := from
			guards[name] = func(_ context.Context, data RecoveryData, _ RecoveryInputKind) (bool, error) {
				needed := maximumActive(withActive(data.Active, data.input))
				if needed < current {
					needed = current
				}
				return needed == target, nil
			}
			rows = append(rows, fsm.Row[RecoveryLevel, RecoveryInputKind]{From: from, On: TriggerActive, Guard: name, To: to, Action: "active"})
		}
		rows = append(rows, fsm.Row[RecoveryLevel, RecoveryInputKind]{From: from, On: TriggerNormalized, To: from, Action: "normalize"})
		if from > LevelNormal {
			rows = append(rows, fsm.Row[RecoveryLevel, RecoveryInputKind]{From: from, On: CooldownTick, Guard: cooldownGuardName(from), To: from - 1, Action: "relax"})
		}
	}
	for _, from := range states[1:] {
		current := from
		guards[cooldownGuardName(from)] = func(_ context.Context, data RecoveryData, _ RecoveryInputKind) (bool, error) {
			return !data.Deadline.IsZero() && !data.input.At.Before(data.Deadline) && maximumActive(data.Active) < current, nil
		}
	}
	machine, err := fsm.New(fsm.Definition[RecoveryLevel, RecoveryInputKind, RecoveryData, RecoveryEffect]{
		Initial: LevelNormal, States: states,
		Inputs: []RecoveryInputKind{TriggerActive, TriggerNormalized, CooldownTick},
		Rows:   rows, Guards: guards,
		Actions: map[string]fsm.Action[RecoveryData, RecoveryInputKind, RecoveryEffect]{
			"active":    plugin.active,
			"normalize": plugin.normalize,
			"relax":     plugin.relax,
		},
	})
	if err != nil {
		return nil, err
	}
	plugin.machine = machine
	return plugin, nil
}

// Initial returns an empty instance for a newly configured task.
func (plugin *StagedRecovery) Initial() fsm.Instance[RecoveryLevel, RecoveryData] {
	return plugin.machine.Initial(RecoveryData{Active: make(map[string]RecoveryLevel)})
}

// Step evaluates one input without committing it, allowing the caller to
// verify an effect before persisting Next.
func (plugin *StagedRecovery) Step(ctx context.Context, instance fsm.Instance[RecoveryLevel, RecoveryData], input RecoveryInput) (RecoveryResult, error) {
	if input.At.IsZero() || input.Kind != TriggerActive && input.Kind != TriggerNormalized && input.Kind != CooldownTick ||
		input.Kind == TriggerActive && (input.AlertID == "" || input.Level < LevelWarning || input.Level > LevelCritical) ||
		input.Kind == TriggerNormalized && input.AlertID == "" || input.Kind == CooldownTick && input.AlertID != "" ||
		input.Kind != TriggerActive && input.Level != LevelNormal {
		return RecoveryResult{}, errors.New("invalid recovery input")
	}
	instance.Data.input = input
	instance.Data.applied = instance.State
	return plugin.machine.Step(ctx, instance, input.Kind)
}

// active tracks a validated alert and requests an immediate tighter verified effect.
func (plugin *StagedRecovery) active(_ context.Context, data RecoveryData, _ RecoveryInputKind) (RecoveryData, []RecoveryEffect, error) {
	data.Active = withActive(data.Active, data.input)
	level := maximumActive(data.Active)
	if level >= data.applied {
		data.Deadline = time.Time{}
	}
	if level > data.applied {
		return data, []RecoveryEffect{{Level: level}}, nil
	}
	return data, nil, nil
}

// normalize removes one tracked alert and starts a fresh cooldown.
func (plugin *StagedRecovery) normalize(_ context.Context, data RecoveryData, _ RecoveryInputKind) (RecoveryData, []RecoveryEffect, error) {
	if _, tracked := data.Active[data.input.AlertID]; !tracked {
		return data, nil, nil
	}
	active := make(map[string]RecoveryLevel, len(data.Active))
	for id, level := range data.Active {
		if id != data.input.AlertID {
			active[id] = level
		}
	}
	data.Active = active
	data.Deadline = data.input.At.Add(plugin.cooldown)
	return data, nil, nil
}

// relax requests exactly one less restrictive level after the deadline.
func (plugin *StagedRecovery) relax(_ context.Context, data RecoveryData, _ RecoveryInputKind) (RecoveryData, []RecoveryEffect, error) {
	data.Deadline = data.input.At.Add(plugin.cooldown)
	return data, []RecoveryEffect{{Level: data.applied - 1}}, nil
}

// withActive copies alert state before adding one observation.
func withActive(active map[string]RecoveryLevel, input RecoveryInput) map[string]RecoveryLevel {
	result := make(map[string]RecoveryLevel, len(active)+1)
	for id, level := range active {
		result[id] = level
	}
	if previous := result[input.AlertID]; previous > input.Level {
		result[input.AlertID] = previous
	} else {
		result[input.AlertID] = input.Level
	}
	return result
}

// maximumActive finds the strongest remaining trigger.
func maximumActive(active map[string]RecoveryLevel) RecoveryLevel {
	var maximum RecoveryLevel
	for _, level := range active {
		if level > maximum {
			maximum = level
		}
	}
	return maximum
}

// guardName identifies one compiled escalation row.
func guardName(from, to RecoveryLevel) string {
	return string(rune('0'+from)) + "_" + string(rune('0'+to))
}

// cooldownGuardName identifies one compiled recovery row.
func cooldownGuardName(from RecoveryLevel) string {
	return "cooldown_" + string(rune('0'+from))
}
