package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	detection "lite-nas/shared/automationdetection"
	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/fsm"
)

// stageTask owns one host-local protective rule and its published event.
type stageTask struct {
	id, ruleID, eventID string
	instance            fsm.Instance[detection.ChainStageState, detection.ChainStageData]
	sources             []string
	capKHz              uint64
	manual              bool
	pending             bool
	lastOutcome         string
}

// view exposes the verified stage and current source alerts to an operator.
func (task *stageTask) view(rule *compiledRule) contract.OperationalSafetyTask {
	level := contract.LevelNormal
	if task.instance.State == detection.ChainStageActive || task.instance.State == detection.ChainStageCooling {
		level = rule.spec.Action.Level
	}
	view := contract.OperationalSafetyTask{ID: task.id, RuleID: task.ruleID, CorrelationKey: "local-host", Level: level, ActiveAlertIDs: append([]string(nil), task.sources...), LastOutcome: task.lastOutcome}
	if !task.instance.Data.Deadline.IsZero() {
		view.CooldownUntil = task.instance.Data.Deadline.UTC().Format(time.RFC3339)
	}
	return view
}

// stageSnapshotLocked joins source alerts and the adjacent rule events for one local host.
func (runtime *safetyRuntime) stageSnapshotLocked(ruleID string) (detection.ChainStageSnapshot, []string) {
	rule := runtime.rules[ruleID]
	snapshot := detection.ChainStageSnapshot{At: runtime.now(), ParentActive: rule.spec.Action.ParentRule == ""}
	sources := make([]string, 0)
	for id, event := range runtime.events {
		if event.state == "normal" || event.alert.EventID == "" {
			continue
		}
		for step := range rule.spec.Steps {
			if _, matched := rule.matches(step, "stage", id, *event); matched {
				sources = append(sources, id)
				break
			}
		}
	}
	sort.Strings(sources)
	snapshot.SourceActive = len(sources) > 0
	if task := runtime.stages[ruleID]; task != nil && task.manual {
		snapshot.SourceActive = true
	}
	if parent := runtime.stages[rule.spec.Action.ParentRule]; parent != nil && parent.eventID != "" {
		parentEvent := runtime.events[parent.eventID]
		snapshot.ParentActive = (parent.instance.State == detection.ChainStageActive || parent.instance.State == detection.ChainStageCooling) && parentEvent != nil && parentEvent.state != "normal"
	}
	for childID, child := range runtime.stages {
		if childID != ruleID && runtime.rules[childID].spec.Action.ParentRule == ruleID && child.instance.State != detection.ChainStageIdle {
			snapshot.ChildActive = true
			break
		}
	}
	return snapshot, sources
}

// evaluateStage advances one rule only after an input or its cooldown is observed.
func (runtime *safetyRuntime) evaluateStage(ctx context.Context, ruleID string) error {
	runtime.mu.Lock()
	rule := runtime.rules[ruleID]
	if rule == nil || rule.stage == nil {
		runtime.mu.Unlock()
		return nil
	}
	snapshot, sources := runtime.stageSnapshotLocked(ruleID)
	task := runtime.stages[ruleID]
	if !rule.enabled && task == nil {
		runtime.mu.Unlock()
		return nil
	}
	if task == nil {
		if !snapshot.SourceActive || !snapshot.ParentActive {
			runtime.mu.Unlock()
			return nil
		}
		id, err := newTaskID()
		if err != nil {
			runtime.mu.Unlock()
			return err
		}
		task = &stageTask{id: id, ruleID: ruleID, instance: rule.stage.Initial()}
		if runtime.stages == nil {
			runtime.stages = make(map[string]*stageTask)
		}
		runtime.stages[ruleID] = task
	}
	if task.pending {
		runtime.mu.Unlock()
		return nil
	}
	task.sources = sources
	result, err := rule.stage.Evaluate(ctx, task.instance, snapshot)
	if err != nil {
		runtime.mu.Unlock()
		return err
	}
	if len(result.Effects) == 0 {
		task.instance = result.Next
		kind := ""
		if task.instance.State == detection.ChainStageRestoring && task.eventID != "" {
			kind = "stage_normalize"
		}
		if kind == "" {
			runtime.mu.Unlock()
			return nil
		}
		task.pending = true
		runtime.mu.Unlock()
		return runtime.submitStage(ctx, ruleID, kind, result.Next)
	}
	kind := "stage_apply"
	if result.Effects[0] == detection.ChainStageRestore {
		kind = "stage_restore"
	}
	task.pending = true
	runtime.mu.Unlock()
	return runtime.submitStage(ctx, ruleID, kind, result.Next)
}

// submitStage sends one proposed stage effect to the outbound worker and waits for verification.
func (runtime *safetyRuntime) submitStage(ctx context.Context, ruleID, kind string, next fsm.Instance[detection.ChainStageState, detection.ChainStageData]) error {
	done := make(chan error, 1)
	if err := runtime.enqueueEffect(ctx, effectIntent{kind: kind, ruleID: ruleID, stageNext: next, done: done}); err != nil {
		runtime.mu.Lock()
		runtime.stages[ruleID].pending = false
		runtime.mu.Unlock()
		return err
	}
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// applyStageIntent verifies a host cap, then creates or normalizes the owned event.
func (runtime *safetyRuntime) applyStageIntent(ctx context.Context, intent effectIntent) error {
	runtime.mu.Lock()
	task := runtime.stages[intent.ruleID]
	if task == nil {
		runtime.mu.Unlock()
		return errors.New("stage task disappeared")
	}
	rule := runtime.rules[intent.ruleID]
	target := rule.spec.Action.ActiveKHz
	if intent.kind == "stage_restore" {
		target = rule.spec.Action.NormalKHz
	}
	for id, other := range runtime.stages {
		if id != intent.ruleID && (other.instance.State == detection.ChainStageActive || other.instance.State == detection.ChainStageCooling) && other.capKHz > 0 && other.capKHz < target {
			target = other.capKHz
		}
	}
	for _, other := range runtime.tasks {
		candidate := runtime.rules[other.ruleID].spec.Action.CapsKHz[other.level]
		if candidate > 0 && candidate < target {
			target = candidate
		}
	}
	expected := runtime.lastCap
	previousEventID := task.eventID
	runtime.mu.Unlock()
	if err := runtime.apply(ctx, target, expected); err != nil {
		runtime.mu.Lock()
		task.pending = false
		task.lastOutcome = "CPU cap failed: " + err.Error()
		runtime.mu.Unlock()
		return err
	}
	runtime.mu.Lock()
	runtime.lastCap = target
	task.capKHz = target
	runtime.mu.Unlock()
	if intent.kind == "stage_apply" && previousEventID == "" {
		id, err := runtime.publishOwned(ctx, intent.ruleID, "local-host")
		if err != nil {
			runtime.mu.Lock()
			task.pending = false
			task.lastOutcome = "CPU cap verified; event publication failed: " + err.Error()
			runtime.mu.Unlock()
			return err
		}
		runtime.mu.Lock()
		task.eventID = id
		runtime.mu.Unlock()
	}
	runtime.mu.Lock()
	task.instance = intent.stageNext
	task.pending = intent.kind == "stage_restore"
	task.lastOutcome = fmt.Sprintf("verified CPU cap %d kHz", target)
	runtime.mu.Unlock()
	if intent.kind == "stage_restore" {
		return runtime.normalizeStageTask(ctx, intent.ruleID)
	}
	return nil
}

// normalizeStageTask retries a failed event normalization without reapplying the cap.
func (runtime *safetyRuntime) normalizeStageTask(ctx context.Context, ruleID string) error {
	runtime.mu.Lock()
	task := runtime.stages[ruleID]
	if task == nil || task.instance.State != detection.ChainStageRestoring || task.eventID == "" {
		runtime.mu.Unlock()
		return errors.New("stage has no event to normalize")
	}
	id := task.eventID
	runtime.mu.Unlock()
	if err := runtime.normalizeOwned(ctx, id); err != nil {
		runtime.mu.Lock()
		task.pending = false
		task.lastOutcome = "cap verified; event normalization failed: " + err.Error()
		runtime.mu.Unlock()
		return err
	}
	runtime.mu.Lock()
	result, err := runtime.rules[ruleID].stage.Finish(ctx, task.instance)
	if err == nil {
		task.instance = result.Next
		task.eventID = ""
		task.pending = false
		task.lastOutcome = "stage event normalized after verified recovery"
	}
	runtime.mu.Unlock()
	return err
}

// stageByTaskIDLocked locates one operator-visible linked stage task.
func (runtime *safetyRuntime) stageByTaskIDLocked(id string) (string, *stageTask) {
	for ruleID, task := range runtime.stages {
		if task.id == id {
			return ruleID, task
		}
	}
	return "", nil
}

// stageResponse reports the task's currently verified stage level.
func (runtime *safetyRuntime) stageResponse(ruleID string, task *stageTask, state contract.OperationalSafetyOutcomeState, outcome string) contract.OperationalSafetyOperationResponse {
	return contract.OperationalSafetyOperationResponse{TaskID: task.id, State: state, Level: task.view(runtime.rules[ruleID]).Level, Outcome: outcome}
}

// forceStageSet applies only the active or normal cap declared by the existing stage task.
func (runtime *safetyRuntime) forceStageSet(ctx context.Context, ruleID string, level contract.OperationalSafetyLevel) (contract.OperationalSafetyOperationResponse, error) {
	if level == contract.LevelNormal {
		return runtime.forceStageStop(ctx, ruleID)
	}
	runtime.mu.Lock()
	task := runtime.stages[ruleID]
	rule := runtime.rules[ruleID]
	if level != rule.spec.Action.Level {
		runtime.mu.Unlock()
		return contract.OperationalSafetyOperationResponse{}, errors.New("level is not declared by this stage task")
	}
	if task.pending {
		response := runtime.stageResponse(ruleID, task, contract.OutcomePending, "stage effect is pending")
		runtime.mu.Unlock()
		return response, nil
	}
	snapshot, _ := runtime.stageSnapshotLocked(ruleID)
	if !snapshot.ParentActive {
		response := runtime.stageResponse(ruleID, task, contract.OutcomeBlocked, "parent stage event is not active")
		runtime.mu.Unlock()
		return response, nil
	}
	previousManual := task.manual
	task.manual = true
	runtime.mu.Unlock()
	if err := runtime.evaluateStage(ctx, ruleID); err != nil {
		runtime.mu.Lock()
		task.manual = previousManual
		response := runtime.stageResponse(ruleID, task, contract.OutcomeFailed, err.Error())
		runtime.mu.Unlock()
		return response, nil
	}
	runtime.mu.Lock()
	response := runtime.stageResponse(ruleID, task, contract.OutcomeCompleted, "configured stage cap verified")
	runtime.mu.Unlock()
	return response, nil
}

// forceStageStop releases a manual hold and starts cooldown only after source and child alerts clear.
func (runtime *safetyRuntime) forceStageStop(ctx context.Context, ruleID string) (contract.OperationalSafetyOperationResponse, error) {
	runtime.mu.Lock()
	task := runtime.stages[ruleID]
	if task.pending {
		response := runtime.stageResponse(ruleID, task, contract.OutcomePending, "stage effect is pending")
		runtime.mu.Unlock()
		return response, nil
	}
	snapshot, sources := runtime.stageSnapshotLocked(ruleID)
	if len(sources) != 0 || snapshot.ChildActive {
		response := runtime.stageResponse(ruleID, task, contract.OutcomeBlocked, "source or child alert still requires this cap")
		runtime.mu.Unlock()
		return response, nil
	}
	if task.instance.State == detection.ChainStageIdle {
		response := runtime.stageResponse(ruleID, task, contract.OutcomeCompleted, "stage is already normal")
		runtime.mu.Unlock()
		return response, nil
	}
	task.manual = false
	runtime.mu.Unlock()
	if err := runtime.evaluateStage(ctx, ruleID); err != nil {
		return contract.OperationalSafetyOperationResponse{}, err
	}
	runtime.mu.Lock()
	response := runtime.stageResponse(ruleID, task, contract.OutcomePending, "stage recovery cooldown started")
	runtime.mu.Unlock()
	return response, nil
}
