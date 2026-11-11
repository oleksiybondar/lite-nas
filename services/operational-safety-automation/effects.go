package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	detection "lite-nas/shared/automationdetection"
	alertcontract "lite-nas/shared/contracts/loggingmanager"
	contract "lite-nas/shared/contracts/operationalsafety"
	systemcontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/fsm"
	"lite-nas/shared/loggingmanager/enum"
)

// effectIntent is one typed outbound alert or verified CPU policy transition.
type effectIntent struct {
	kind      string
	ruleID    string
	key       string
	taskID    string
	level     contract.OperationalSafetyLevel
	next      fsm.Instance[detection.RecoveryLevel, detection.RecoveryData]
	stageNext fsm.Instance[detection.ChainStageState, detection.ChainStageData]
	done      chan error
	message   string
	capKHz    uint64
	eventID   string
}

// applyIntent serializes effects and publications on the outbound worker.
func (runtime *safetyRuntime) applyIntent(ctx context.Context, intent effectIntent) {
	var err error
	switch intent.kind {
	case "cpu":
		err = runtime.applyCPUIntent(ctx, intent)
	case "emit", "rule_state":
		err = runtime.publish(ctx, intent)
	case "emit_normal":
		err = runtime.emitNormalIntent(ctx, intent)
	case "normalize_record":
		err = runtime.normalizeRecordIntent(ctx, intent.eventID)
	case "stage_apply", "stage_restore":
		err = runtime.applyStageIntent(ctx, intent)
	case "stage_normalize":
		err = runtime.normalizeStageTask(ctx, intent.ruleID)
	case "manual_cap":
		err = runtime.applyManualIntent(ctx, intent)
	default:
		err = errors.New("unknown outbound intent")
	}
	if err != nil {
		runtime.log.Error("operational safety effect failed", "kind", intent.kind, "rule", intent.ruleID, "task_id", intent.taskID, "error", err)
	}
	if intent.done != nil {
		intent.done <- err
	}
}

// emitNormalIntent records the verified baseline and tracks normalization for retry.
func (runtime *safetyRuntime) emitNormalIntent(ctx context.Context, intent effectIntent) error {
	runtime.mu.Lock()
	full := len(runtime.pendingNormal) >= 128
	runtime.mu.Unlock()
	if full {
		return errors.New("pending normal-event queue is full")
	}
	id, err := runtime.publishOwned(ctx, intent.ruleID, intent.key)
	if err != nil {
		return err
	}
	runtime.mu.Lock()
	if runtime.pendingNormal == nil {
		runtime.pendingNormal = make(map[string]bool)
	}
	runtime.pendingNormal[id] = true
	runtime.mu.Unlock()
	return runtime.normalizeRecordIntent(ctx, id)
}

// normalizeRecordIntent retries one normal record until the Logging Manager accepts it.
func (runtime *safetyRuntime) normalizeRecordIntent(ctx context.Context, id string) error {
	err := runtime.normalizeOwned(ctx, id)
	runtime.mu.Lock()
	if err == nil {
		delete(runtime.pendingNormal, id)
	} else {
		runtime.pendingNormal[id] = false
	}
	runtime.mu.Unlock()
	return err
}

// applyManualIntent changes a fresh runtime's host cap under the same effect serialization.
func (runtime *safetyRuntime) applyManualIntent(ctx context.Context, intent effectIntent) error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	if len(runtime.tasks) != 0 || len(runtime.stages) != 0 {
		return errors.New("manual cap requires no in-memory automation tasks")
	}
	if err := runtime.applyManual(ctx, intent.capKHz); err != nil {
		return err
	}
	runtime.lastCap = intent.capKHz
	runtime.log.Warn("operational safety manual CPU cap verified", "cap_khz", intent.capKHz, "actor", intent.key, "reason", intent.message)
	return nil
}

// applyCPUIntent applies and verifies a configured cap before committing the FSM state.
func (runtime *safetyRuntime) applyCPUIntent(ctx context.Context, intent effectIntent) error {
	runtime.mu.Lock()
	task := runtime.tasks[intent.taskID]
	if task == nil {
		runtime.mu.Unlock()
		return errors.New("task disappeared before effect")
	}
	rule := runtime.rules[task.ruleID]
	capKHz := rule.spec.Action.CapsKHz[intent.level]
	for _, other := range runtime.tasks {
		if other.id == task.id {
			continue
		}
		candidate := runtime.rules[other.ruleID].spec.Action.CapsKHz[other.level]
		if candidate > 0 && candidate < capKHz {
			capKHz = candidate
		}
	}
	for _, stage := range runtime.stages {
		if (stage.instance.State == detection.ChainStageActive || stage.instance.State == detection.ChainStageCooling) && stage.capKHz > 0 && stage.capKHz < capKHz {
			capKHz = stage.capKHz
		}
	}
	expectedKHz := runtime.lastCap
	runtime.mu.Unlock()
	if capKHz == 0 {
		return errors.New("effect level has no configured cap")
	}
	if err := runtime.apply(ctx, capKHz, expectedKHz); err != nil {
		runtime.mu.Lock()
		task.pending = false
		task.lastOutcome = "CPU cap failed: " + err.Error()
		runtime.mu.Unlock()
		return err
	}
	runtime.mu.Lock()
	task.instance = intent.next
	runtime.lastCap = capKHz
	task.level = intent.level
	task.pending = false
	task.lastOutcome = fmt.Sprintf("verified CPU cap %d kHz", capKHz)
	runtime.mu.Unlock()
	intent.ruleID = rule.spec.ID
	intent.key = task.key
	intent.message = fmt.Sprintf("%s: CPU cap %d kHz at %s", strings.ReplaceAll(rule.spec.Output.Message, "{key}", task.key), capKHz, intent.level)
	if err := runtime.publish(ctx, intent); err != nil {
		runtime.mu.Lock()
		task.lastOutcome = "CPU cap verified; alert publication failed: " + err.Error()
		runtime.mu.Unlock()
		return err
	}
	return nil
}

// publishIntent creates one normalizable system alert with a rule-owned ID prefix.
func (runtime *safetyRuntime) publishIntent(ctx context.Context, intent effectIntent) error {
	_, err := runtime.publishAlert(ctx, intent.ruleID, intent.key, intent.message)
	return err
}

// publishStageAlert creates an owned event whose ID is retained until verified recovery.
func (runtime *safetyRuntime) publishStageAlert(ctx context.Context, ruleID, key string) (string, error) {
	return runtime.publishAlert(ctx, ruleID, key, "")
}

// publishAlert reserves an ID and sends one configured system alert.
func (runtime *safetyRuntime) publishAlert(ctx context.Context, ruleID, key, message string) (string, error) {
	rule := runtime.rules[ruleID]
	if rule == nil {
		return "", errors.New("output rule unavailable")
	}
	token, err := runtime.serviceToken(ctx)
	if err != nil {
		return "", err
	}
	id, err := runtime.ids.Next(rule.spec.Output.IDPrefix)
	if err != nil {
		return "", err
	}
	if message == "" {
		message = strings.ReplaceAll(rule.spec.Output.Message, "{key}", key)
	}
	if len(message) > 256 {
		message = message[:256]
	}
	if err := runtime.client.Publish(ctx, systemcontract.AlertSubject, alertcontract.AlertPayload{
		AccessToken: token, EventID: id, Category: rule.spec.Output.Category,
		Severity: enum.Severity(rule.spec.Output.Severity), CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Source: "operational-safety-automation", Message: message,
	}); err != nil {
		return "", err
	}
	return id, nil
}

// normalizeStageAlert closes one owned event after its parent cap is verified.
func (runtime *safetyRuntime) normalizeStageAlert(ctx context.Context, eventID string) error {
	token, err := runtime.serviceToken(ctx)
	if err != nil {
		return err
	}
	message := "CPU frequency stage recovered after cooldown"
	var response alertcontract.OKResponse
	if err := runtime.client.Request(ctx, systemcontract.UpdateAlertStateRPCSubject, alertcontract.UpdateAlertStateInput{AccessToken: token, EventID: eventID, Status: enum.StatusNormal, Message: &message}, &response); err != nil {
		return err
	}
	if !response.OK {
		return errors.New("system alert normalization was rejected")
	}
	return nil
}
