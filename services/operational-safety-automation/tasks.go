package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"

	detection "lite-nas/shared/automationdetection"
	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/fsm"
)

// safetyTask is one in-memory CPU effect state for a rule and correlation key.
type safetyTask struct {
	id, ruleID, key string
	plugin          *detection.StagedRecovery
	instance        fsm.Instance[detection.RecoveryLevel, detection.RecoveryData]
	level           contract.OperationalSafetyLevel
	pending         bool
	lastOutcome     string
}

// view snapshots one task for an authenticated operator read.
func (task *safetyTask) view() contract.OperationalSafetyTask {
	ids := make([]string, 0, len(task.instance.Data.Active))
	for id := range task.instance.Data.Active {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	view := contract.OperationalSafetyTask{ID: task.id, RuleID: task.ruleID, CorrelationKey: task.key, Level: task.level, ActiveAlertIDs: ids, LastOutcome: task.lastOutcome}
	if !task.instance.Data.Deadline.IsZero() {
		view.CooldownUntil = task.instance.Data.Deadline.UTC().Format(time.RFC3339)
	}
	return view
}

// newTaskID produces a short unpredictable identifier accepted by the RPC boundary.
func newTaskID() (string, error) {
	var raw [12]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// levelNumber maps the external level contract to the staged recovery FSM order.
func levelNumber(level contract.OperationalSafetyLevel) (detection.RecoveryLevel, error) {
	switch level {
	case contract.LevelNormal:
		return detection.LevelNormal, nil
	case contract.LevelWarning:
		return detection.LevelWarning, nil
	case contract.LevelError:
		return detection.LevelError, nil
	case contract.LevelCritical:
		return detection.LevelCritical, nil
	default:
		return 0, errors.New("unknown CPU effect level")
	}
}

// levelName maps a verified FSM level back to its operator response.
func levelName(level detection.RecoveryLevel) contract.OperationalSafetyLevel {
	switch level {
	case detection.LevelWarning:
		return contract.LevelWarning
	case detection.LevelError:
		return contract.LevelError
	case detection.LevelCritical:
		return contract.LevelCritical
	default:
		return contract.LevelNormal
	}
}

// activateTask steps the rule's FSM and commits only after a verified effect.
func (runtime *safetyRuntime) activateTask(ctx context.Context, ruleID, key, eventID string, level contract.OperationalSafetyLevel) error {
	runtime.mu.Lock()
	rule, err := runtime.mustRule(ruleID)
	if err != nil {
		runtime.mu.Unlock()
		return err
	}
	var task *safetyTask
	for _, existing := range runtime.tasks {
		if existing.ruleID == ruleID && existing.key == key {
			task = existing
			break
		}
	}
	if task == nil {
		id, idErr := newTaskID()
		if idErr != nil {
			runtime.mu.Unlock()
			return idErr
		}
		plugin := rule.recovery
		task = &safetyTask{id: id, ruleID: ruleID, key: key, plugin: plugin, instance: plugin.Initial(), level: contract.LevelNormal}
		runtime.tasks[id] = task
	}
	if task.pending {
		runtime.mu.Unlock()
		return errors.New("task effect is pending")
	}
	desired, err := levelNumber(level)
	if err != nil {
		runtime.mu.Unlock()
		return err
	}
	result, err := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.TriggerActive, AlertID: eventID, Level: desired, At: runtime.now()})
	if err != nil {
		runtime.mu.Unlock()
		return err
	}
	if len(result.Effects) == 0 {
		task.instance = result.Next
		runtime.mu.Unlock()
		return nil
	}
	task.pending = true
	id := task.id
	runtime.mu.Unlock()
	return runtime.submitTransition(ctx, id, result)
}

// normalizeTaskAlert marks one verified normal alert and starts the rule cooldown.
func (runtime *safetyRuntime) normalizeTaskAlert(ctx context.Context, eventID string) error {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	for _, task := range runtime.tasks {
		if _, tracked := task.instance.Data.Active[eventID]; !tracked {
			continue
		}
		if task.pending {
			return errors.New("task effect is pending during normalization")
		}
		result, err := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.TriggerNormalized, AlertID: eventID, At: runtime.now()})
		if err != nil {
			return err
		}
		task.instance = result.Next
		task.lastOutcome = "trigger normalized; recovery cooldown started"
	}
	return nil
}

// advanceCooldowns requests at most one verified relaxation for each due task.
func (runtime *safetyRuntime) advanceCooldowns(ctx context.Context) {
	runtime.mu.Lock()
	ids := make([]string, 0, len(runtime.tasks))
	for id, task := range runtime.tasks {
		if !task.pending && !task.instance.Data.Deadline.IsZero() && !runtime.now().Before(task.instance.Data.Deadline) {
			ids = append(ids, id)
		}
	}
	runtime.mu.Unlock()
	for _, id := range ids {
		if err := runtime.advanceCooldown(ctx, id); err != nil {
			runtime.log.Error("operational safety cooldown failed", "task_id", id, "error", err)
		}
	}
	for _, ruleID := range runtime.order {
		if runtime.rules[ruleID].stage != nil {
			if err := runtime.evaluateStage(ctx, ruleID); err != nil {
				runtime.log.Error("operational safety stage cooldown failed", "rule_id", ruleID, "error", err)
			}
		}
	}
	runtime.retryNormalRecords(ctx)
}

// retryNormalRecords requeues failed informational-event normalization without dropping it.
func (runtime *safetyRuntime) retryNormalRecords(ctx context.Context) {
	runtime.mu.Lock()
	ids := make([]string, 0, len(runtime.pendingNormal))
	for id, pending := range runtime.pendingNormal {
		if !pending {
			runtime.pendingNormal[id] = true
			ids = append(ids, id)
		}
	}
	runtime.mu.Unlock()
	for _, id := range ids {
		if err := runtime.enqueueEffect(ctx, effectIntent{kind: "normalize_record", eventID: id}); err != nil {
			runtime.mu.Lock()
			runtime.pendingNormal[id] = false
			runtime.mu.Unlock()
			runtime.log.Error("normal event retry enqueue failed", "event_id", id, "error", err)
		}
	}
}

// advanceCooldown evaluates one table-driven recovery step without bypassing active alerts.
func (runtime *safetyRuntime) advanceCooldown(ctx context.Context, id string) error {
	runtime.mu.Lock()
	task := runtime.tasks[id]
	if task == nil || task.pending {
		runtime.mu.Unlock()
		return nil
	}
	result, err := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.CooldownTick, At: runtime.now()})
	if err != nil {
		runtime.mu.Unlock()
		return err
	}
	if len(result.Effects) == 0 {
		task.instance = result.Next
		runtime.mu.Unlock()
		return nil
	}
	task.pending = true
	runtime.mu.Unlock()
	return runtime.submitTransition(ctx, id, result)
}

// submitTransition hands one proposed FSM state to the outbound worker and waits for verification.
func (runtime *safetyRuntime) submitTransition(ctx context.Context, id string, result detection.RecoveryResult) error {
	if len(result.Effects) != 1 {
		return fmt.Errorf("unexpected effect count %d", len(result.Effects))
	}
	done := make(chan error, 1)
	intent := effectIntent{kind: "cpu", taskID: id, level: levelName(result.Effects[0].Level), next: result.Next, done: done}
	if err := runtime.enqueueEffect(ctx, intent); err != nil {
		runtime.mu.Lock()
		if task := runtime.tasks[id]; task != nil {
			task.pending = false
		}
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
