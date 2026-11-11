package automationdetection

import (
	"context"
	"errors"
	"time"

	"lite-nas/shared/fsm"
)

// ChainStageState identifies a protective stage's verified lifecycle phase.
type ChainStageState string

const (
	// ChainStageIdle has no active effect or owned alert.
	ChainStageIdle ChainStageState = "idle"
	// ChainStageActive owns a verified protective cap and alert.
	ChainStageActive ChainStageState = "active"
	// ChainStageCooling keeps its cap and alert while waiting for a clear condition.
	ChainStageCooling ChainStageState = "cooling"
	// ChainStageRestoring has verified the parent cap and must normalize its alert.
	ChainStageRestoring ChainStageState = "restoring"
)

// ChainStageEffect names the host or alert operation required before committing a transition.
type ChainStageEffect string

const (
	// ChainStageApply verifies the stage cap and creates its alert.
	ChainStageApply ChainStageEffect = "apply"
	// ChainStageRestore verifies the parent cap before normalizing the stage alert.
	ChainStageRestore ChainStageEffect = "restore"
)

type chainStageInput string

const (
	chainStageEvaluate chainStageInput = "evaluate"
	chainStageFinish   chainStageInput = "finish"
)

// ChainStageSnapshot is a fully evaluated local view of source, parent, and child alerts.
type ChainStageSnapshot struct {
	At           time.Time
	SourceActive bool
	ParentActive bool
	ChildActive  bool
}

// ChainStageData retains only the recovery deadline and current evaluated inputs.
type ChainStageData struct {
	Deadline time.Time
	snapshot ChainStageSnapshot
}

// ChainStageResult is an uncommitted transition whose effect must be verified first.
type ChainStageResult = fsm.Result[ChainStageState, chainStageInput, ChainStageData, ChainStageEffect]

// ChainStage holds the table for one rule's active cap, cooldown, and parent restoration.
type ChainStage struct {
	machine  *fsm.Machine[ChainStageState, chainStageInput, ChainStageData, ChainStageEffect]
	cooldown time.Duration
}

// NewChainStage compiles one stage with a positive recovery cooldown.
func NewChainStage(cooldown time.Duration) (*ChainStage, error) {
	if cooldown <= 0 {
		return nil, errors.New("chain stage cooldown must be positive")
	}
	stage := &ChainStage{cooldown: cooldown}
	machine, err := fsm.New(fsm.Definition[ChainStageState, chainStageInput, ChainStageData, ChainStageEffect]{
		Initial: ChainStageIdle,
		States:  []ChainStageState{ChainStageIdle, ChainStageActive, ChainStageCooling, ChainStageRestoring},
		Inputs:  []chainStageInput{chainStageEvaluate, chainStageFinish},
		Rows: []fsm.Row[ChainStageState, chainStageInput]{
			{From: ChainStageIdle, On: chainStageEvaluate, Guard: "ready", To: ChainStageActive, Action: "apply"},
			{From: ChainStageActive, On: chainStageEvaluate, Guard: "clear", To: ChainStageCooling, Action: "start_cooldown"},
			{From: ChainStageCooling, On: chainStageEvaluate, Guard: "blocked", To: ChainStageActive, Action: "cancel_cooldown"},
			{From: ChainStageCooling, On: chainStageEvaluate, Guard: "due", To: ChainStageRestoring, Action: "restore"},
			{From: ChainStageRestoring, On: chainStageEvaluate, Guard: "ready", To: ChainStageActive, Action: "apply"},
			{From: ChainStageRestoring, On: chainStageFinish, To: ChainStageIdle, Action: "finish"},
		},
		Guards: map[string]fsm.Guard[ChainStageData, chainStageInput]{
			"ready": func(_ context.Context, data ChainStageData, _ chainStageInput) (bool, error) {
				return data.snapshot.SourceActive && data.snapshot.ParentActive, nil
			},
			"clear": func(_ context.Context, data ChainStageData, _ chainStageInput) (bool, error) {
				return !data.snapshot.SourceActive && !data.snapshot.ChildActive, nil
			},
			"blocked": func(_ context.Context, data ChainStageData, _ chainStageInput) (bool, error) {
				return data.snapshot.SourceActive || data.snapshot.ChildActive, nil
			},
			"due": func(_ context.Context, data ChainStageData, _ chainStageInput) (bool, error) {
				return !data.Deadline.IsZero() && !data.snapshot.At.Before(data.Deadline) && !data.snapshot.SourceActive && !data.snapshot.ChildActive, nil
			},
		},
		Actions: map[string]fsm.Action[ChainStageData, chainStageInput, ChainStageEffect]{
			"apply": func(_ context.Context, data ChainStageData, _ chainStageInput) (ChainStageData, []ChainStageEffect, error) {
				data.Deadline = time.Time{}
				return data, []ChainStageEffect{ChainStageApply}, nil
			},
			"start_cooldown": func(_ context.Context, data ChainStageData, _ chainStageInput) (ChainStageData, []ChainStageEffect, error) {
				data.Deadline = data.snapshot.At.Add(stage.cooldown)
				return data, nil, nil
			},
			"cancel_cooldown": func(_ context.Context, data ChainStageData, _ chainStageInput) (ChainStageData, []ChainStageEffect, error) {
				data.Deadline = time.Time{}
				return data, nil, nil
			},
			"restore": func(_ context.Context, data ChainStageData, _ chainStageInput) (ChainStageData, []ChainStageEffect, error) {
				data.Deadline = time.Time{}
				return data, []ChainStageEffect{ChainStageRestore}, nil
			},
			"finish": func(_ context.Context, data ChainStageData, _ chainStageInput) (ChainStageData, []ChainStageEffect, error) {
				return ChainStageData{}, nil, nil
			},
		},
	})
	if err != nil {
		return nil, err
	}
	stage.machine = machine
	return stage, nil
}

// Initial creates a clean stage instance without changing the host.
func (stage *ChainStage) Initial() fsm.Instance[ChainStageState, ChainStageData] {
	return stage.machine.Initial(ChainStageData{})
}

// Evaluate computes the next stage state from a caller-validated alert snapshot.
func (stage *ChainStage) Evaluate(ctx context.Context, current fsm.Instance[ChainStageState, ChainStageData], snapshot ChainStageSnapshot) (ChainStageResult, error) {
	if snapshot.At.IsZero() {
		return ChainStageResult{}, errors.New("stage snapshot needs a time")
	}
	current.Data.snapshot = snapshot
	return stage.machine.Step(ctx, current, chainStageEvaluate)
}

// Finish records that the stage's owned alert was successfully normalized.
func (stage *ChainStage) Finish(ctx context.Context, current fsm.Instance[ChainStageState, ChainStageData]) (ChainStageResult, error) {
	if current.State != ChainStageRestoring {
		return ChainStageResult{}, errors.New("only a restoring stage can finish")
	}
	return stage.machine.Step(ctx, current, chainStageFinish)
}
