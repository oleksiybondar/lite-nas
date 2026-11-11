package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/authtoken"
	detection "lite-nas/shared/automationdetection"
	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"
)

// register binds each operator subject to a validated RPC handler.
func (runtime *safetyRuntime) register(server messaging.Server) error {
	for _, entry := range []struct {
		subject string
		handler messaging.RPCHandler
	}{
		{contract.RulesListSubject, runtime.listRules},
		{contract.TasksListSubject, runtime.listTasks},
		{contract.RuleSetEnabledSubject, runtime.setEnabled},
		{contract.TaskForceSetSubject, runtime.forceSet},
		{contract.TaskForceStopSubject, runtime.forceStop},
		{contract.CPUSetCapSubject, runtime.setCPUCap},
	} {
		if err := server.RegisterRPC(entry.subject, entry.handler); err != nil {
			return err
		}
	}
	return nil
}

// setCPUCap verifies a direct operator cap only when this process has no automation tasks.
func (runtime *safetyRuntime) setCPUCap(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeOperatorRequest(ctx, runtime, envelope, func(value contract.OperationalSafetyCPUSetCapRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	runtime.mu.Lock()
	busy := len(runtime.tasks) != 0 || len(runtime.stages) != 0
	runtime.mu.Unlock()
	if busy {
		return contract.OperationalSafetyCPUSetCapResponse{State: contract.OutcomeBlocked, Outcome: "manual cap requires no in-memory automation tasks"}, nil
	}
	done := make(chan error, 1)
	if err := runtime.enqueueEffect(ctx, effectIntent{kind: "manual_cap", capKHz: request.CapKHz, key: claims.Login, message: request.Reason, done: done}); err != nil {
		return nil, err
	}
	select {
	case err := <-done:
		if err != nil {
			return contract.OperationalSafetyCPUSetCapResponse{State: contract.OutcomeFailed, Outcome: err.Error()}, nil
		}
		return contract.OperationalSafetyCPUSetCapResponse{State: contract.OutcomeCompleted, CapKHz: request.CapKHz, Outcome: "CPU cap verified"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// decodeOperatorRequest rejects malformed input and verifies a live security-role token.
func decodeOperatorRequest[T any](ctx context.Context, runtime *safetyRuntime, envelope messaging.Envelope, tokenOf func(T) string) (T, authtoken.AccessClaims, error) {
	var request T
	if len(envelope.Payload) == 0 || len(envelope.Payload) > 16*1024 {
		return request, authtoken.AccessClaims{}, errors.New("invalid operational safety RPC payload size")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, authtoken.AccessClaims{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return request, authtoken.AccessClaims{}, errors.New("trailing operational safety RPC JSON")
	}
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(request); err != nil {
		return request, authtoken.AccessClaims{}, err
	}
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	claims, err := authtoken.VerifyLive(checkCtx, runtime.client, runtime.verifier, tokenOf(request))
	if err != nil || !roleauth.MatchesRequirement(claims.Roles, roleauth.RequirementSecurity) {
		return request, authtoken.AccessClaims{}, errors.New("security operator authorization required")
	}
	return request, claims, nil
}

// listRules reports every configured rule and effective enabled state.
func (runtime *safetyRuntime) listRules(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, _, err := decodeOperatorRequest(ctx, runtime, envelope, func(value contract.OperationalSafetyListRequest) string { return value.AccessToken }); err != nil {
		return nil, err
	}
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	out := contract.OperationalSafetyRulesResponse{Rules: make([]contract.OperationalSafetyRule, 0, len(runtime.order))}
	for _, id := range runtime.order {
		rule := runtime.rules[id]
		out.Rules = append(out.Rules, contract.OperationalSafetyRule{ID: id, Name: rule.spec.Name, Enabled: rule.enabled, ConfiguredEnabled: *rule.spec.Enabled, OutputIDPrefix: rule.spec.Output.IDPrefix, Cooldown: rule.spec.Action.Cooldown})
	}
	return out, nil
}

// listTasks reports active in-memory effect tasks.
func (runtime *safetyRuntime) listTasks(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, _, err := decodeOperatorRequest(ctx, runtime, envelope, func(value contract.OperationalSafetyListRequest) string { return value.AccessToken }); err != nil {
		return nil, err
	}
	return contract.OperationalSafetyTasksResponse{Items: runtime.taskViews()}, nil
}

// setEnabled applies a temporary rule state and publishes an operator audit alert.
func (runtime *safetyRuntime) setEnabled(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeOperatorRequest(ctx, runtime, envelope, func(value contract.OperationalSafetyRuleSetEnabledRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	if !ruleIDPattern.MatchString(request.RuleID) {
		return nil, errors.New("invalid rule ID")
	}
	runtime.mu.Lock()
	rule, err := runtime.mustRule(request.RuleID)
	if err != nil {
		runtime.mu.Unlock()
		return nil, err
	}
	old := rule.enabled
	enabled := *request.Enabled
	if enabled && !old && rule.detector != nil {
		if err := rule.detector.Reset(); err != nil {
			runtime.mu.Unlock()
			return nil, err
		}
	}
	rule.enabled = enabled
	runtime.mu.Unlock()
	if old != enabled {
		message := fmt.Sprintf("Operational safety rule %s temporarily enabled=%t by %s", request.RuleID, enabled, claims.Login)
		done := make(chan error, 1)
		if err := runtime.enqueueEffect(ctx, effectIntent{kind: "rule_state", ruleID: request.RuleID, message: message, done: done}); err != nil {
			return nil, err
		}
		select {
		case err := <-done:
			if err != nil {
				runtime.mu.Lock()
				rule.enabled = old
				runtime.mu.Unlock()
				return contract.OperationalSafetyOperationResponse{State: contract.OutcomeFailed, Outcome: err.Error()}, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	runtime.log.Warn("operational safety rule state changed", "rule", request.RuleID, "enabled", enabled, "actor", claims.Login, "reason", request.Reason)
	return contract.OperationalSafetyOperationResponse{State: contract.OutcomeCompleted, Outcome: "temporary rule state updated"}, nil
}

// forceSet requests a rule-declared level on an existing task through the normal effect path.
func (runtime *safetyRuntime) forceSet(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeOperatorRequest(ctx, runtime, envelope, func(value contract.OperationalSafetyTaskForceSetRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	runtime.mu.Lock()
	task := runtime.tasks[request.TaskID]
	if task == nil {
		stageRuleID, stage := runtime.stageByTaskIDLocked(request.TaskID)
		runtime.mu.Unlock()
		if stage != nil {
			runtime.log.Warn("operational safety force set", "task_id", request.TaskID, "level", request.Level, "actor", claims.Login, "reason", request.Reason)
			return runtime.forceStageSet(ctx, stageRuleID, request.Level)
		}
		return nil, errors.New("unknown operational safety task")
	}
	if task.pending {
		response := taskResponse(task, contract.OutcomePending, "effect already pending")
		runtime.mu.Unlock()
		return response, nil
	}
	desired, _ := levelNumber(request.Level)
	current, _ := levelNumber(task.level)
	if desired < current {
		manualID := "manual" + task.id
		for id, active := range task.instance.Data.Active {
			if id == manualID {
				continue
			}
			if active > desired {
				response := taskResponse(task, contract.OutcomeBlocked, "an active alert still requires a stronger cap")
				runtime.mu.Unlock()
				return response, nil
			}
		}
		if _, manual := task.instance.Data.Active[manualID]; manual {
			result, stepErr := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.TriggerNormalized, AlertID: manualID, At: runtime.now()})
			if stepErr != nil {
				runtime.mu.Unlock()
				return nil, stepErr
			}
			task.instance = result.Next
		}
		if desired > detection.LevelNormal {
			result, stepErr := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.TriggerActive, AlertID: manualID, Level: desired, At: runtime.now()})
			if stepErr != nil {
				runtime.mu.Unlock()
				return nil, stepErr
			}
			task.instance = result.Next
		}
		task.instance.Data.Deadline = runtime.now().Add(runtime.rules[task.ruleID].cooldown)
		response := taskResponse(task, contract.OutcomePending, "staged recovery cooldown started")
		runtime.mu.Unlock()
		return response, nil
	}
	if desired == current {
		response := taskResponse(task, contract.OutcomeCompleted, "requested level already verified")
		runtime.mu.Unlock()
		return response, nil
	}
	result, err := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.TriggerActive, AlertID: "manual" + task.id, Level: desired, At: runtime.now()})
	if err != nil {
		runtime.mu.Unlock()
		return nil, err
	}
	task.pending = true
	runtime.mu.Unlock()
	runtime.log.Warn("operational safety force set", "task_id", request.TaskID, "level", request.Level, "actor", claims.Login, "reason", request.Reason)
	if err := runtime.submitTransition(ctx, request.TaskID, result); err != nil {
		runtime.mu.Lock()
		level := task.level
		runtime.mu.Unlock()
		return contract.OperationalSafetyOperationResponse{TaskID: request.TaskID, State: contract.OutcomeFailed, Level: level, Outcome: err.Error()}, nil
	}
	return contract.OperationalSafetyOperationResponse{TaskID: request.TaskID, State: contract.OutcomeCompleted, Level: request.Level, Outcome: "CPU cap verified"}, nil
}

// forceStop starts the same one-level cooldown path used for automatic recovery.
func (runtime *safetyRuntime) forceStop(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeOperatorRequest(ctx, runtime, envelope, func(value contract.OperationalSafetyTaskForceStopRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	runtime.mu.Lock()
	task := runtime.tasks[request.TaskID]
	if task == nil {
		stageRuleID, stage := runtime.stageByTaskIDLocked(request.TaskID)
		runtime.mu.Unlock()
		if stage != nil {
			runtime.log.Warn("operational safety force stop", "task_id", request.TaskID, "actor", claims.Login, "reason", request.Reason)
			return runtime.forceStageStop(ctx, stageRuleID)
		}
		return nil, errors.New("unknown operational safety task")
	}
	for id := range task.instance.Data.Active {
		if id != "manual"+task.id {
			response := taskResponse(task, contract.OutcomeBlocked, "active alert prevents recovery")
			runtime.mu.Unlock()
			return response, nil
		}
	}
	if _, manual := task.instance.Data.Active["manual"+task.id]; manual {
		result, stepErr := task.plugin.Step(ctx, task.instance, detection.RecoveryInput{Kind: detection.TriggerNormalized, AlertID: "manual" + task.id, At: runtime.now()})
		if stepErr != nil {
			runtime.mu.Unlock()
			return nil, stepErr
		}
		task.instance = result.Next
	}
	if task.level == contract.LevelNormal {
		response := taskResponse(task, contract.OutcomeCompleted, "task already at normal cap")
		runtime.mu.Unlock()
		return response, nil
	}
	if task.instance.Data.Deadline.IsZero() {
		task.instance.Data.Deadline = runtime.now().Add(runtime.rules[task.ruleID].cooldown)
	}
	response := taskResponse(task, contract.OutcomePending, "staged recovery cooldown started")
	runtime.mu.Unlock()
	runtime.log.Warn("operational safety force stop", "task_id", request.TaskID, "actor", claims.Login, "reason", request.Reason)
	return response, nil
}

// taskResponse includes the required task ID and current verified level.
func taskResponse(task *safetyTask, state contract.OperationalSafetyOutcomeState, outcome string) contract.OperationalSafetyOperationResponse {
	return contract.OperationalSafetyOperationResponse{TaskID: task.id, State: state, Level: task.level, Outcome: outcome}
}
