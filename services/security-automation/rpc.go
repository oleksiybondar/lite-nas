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
	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"
)

// register binds every published operator subject to a validated service handler.
func (runtime *automationRuntime) register(server messaging.Server) error {
	for _, operation := range []struct {
		subject string
		handler messaging.RPCHandler
	}{
		{contract.RulesListSubject, runtime.listRules},
		{contract.ActiveListSubject, runtime.listActive},
		{contract.RollbackSubject, runtime.rollback},
		{contract.ForceNormalizeSubject, runtime.forceNormalize},
		{contract.ForceMitigateSubject, runtime.forceMitigate},
		{contract.CooldownResetSubject, runtime.resetCooldown},
		{contract.RuleSetEnabledSubject, runtime.setEnabled},
	} {
		if err := server.RegisterRPC(operation.subject, operation.handler); err != nil {
			return err
		}
	}
	return nil
}

// decodeAutomationRequest rejects unknown JSON, malformed DTOs, and inactive or unauthorized tokens.
func decodeAutomationRequest[T any](ctx context.Context, runtime *automationRuntime, envelope messaging.Envelope, tokenOf func(T) string) (T, authtoken.AccessClaims, error) {
	var request T
	if len(envelope.Payload) == 0 || len(envelope.Payload) > 16*1024 {
		return request, authtoken.AccessClaims{}, errors.New("invalid automation RPC payload size")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, authtoken.AccessClaims{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return request, authtoken.AccessClaims{}, errors.New("trailing automation RPC JSON")
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

// listRules returns all configured definitions and their temporary effective state.
func (runtime *automationRuntime) listRules(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, _, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.ListRequest) string { return value.AccessToken }); err != nil {
		return nil, err
	}
	runtime.mutex.RLock()
	defer runtime.mutex.RUnlock()
	counts := make(map[string]int)
	states := make(map[string]contract.MitigationState)
	if runtime.incidents != nil {
		for _, record := range runtime.incidents.list() {
			counts[record.RuleID]++
			if incidentStatePriority(record.State) > incidentStatePriority(states[record.RuleID]) {
				states[record.RuleID] = record.State
			}
		}
	}
	response := contract.RulesListResponse{Rules: make([]contract.Rule, 0, len(runtime.order))}
	for _, id := range runtime.order {
		view := runtime.rules[id].view
		view.Enabled = runtime.rules[id].enabled.Load()
		view.ActiveCount = counts[id]
		if states[id] != "" {
			view.State = states[id]
		}
		response.Rules = append(response.Rules, view)
	}
	return response, nil
}

// incidentStatePriority selects the most urgent active state for one rule row.
func incidentStatePriority(state contract.MitigationState) int {
	switch state {
	case contract.StateFailed:
		return 4
	case contract.StateRestoring:
		return 3
	case contract.StatePending:
		return 2
	case contract.StateMitigating:
		return 1
	default:
		return 0
	}
}

// listActive returns durable incidents whose owned effects may need attention.
func (runtime *automationRuntime) listActive(ctx context.Context, envelope messaging.Envelope) (any, error) {
	if _, _, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.ListRequest) string { return value.AccessToken }); err != nil {
		return nil, err
	}
	runtime.mutex.RLock()
	defer runtime.mutex.RUnlock()
	response := contract.ActiveListResponse{Items: []contract.ActiveAutomation{}}
	if runtime.incidents == nil {
		return response, nil
	}
	for _, record := range runtime.incidents.list() {
		rule := runtime.rules[record.RuleID]
		if rule == nil {
			continue
		}
		view := rule.view
		view.Enabled = rule.enabled.Load()
		target := record.SourceIP
		mitigation := "block_source_ip"
		if record.ActionName == "block_scan_source_ip" {
			mitigation = record.ActionName
		}
		if record.ActionName == "revoke_web_sessions" {
			target = record.Username + " from " + record.SourceIP
			mitigation = record.ActionName
		}
		if record.ActionName == "login_terminate_session" {
			target = record.Username + " tty " + record.TTY
			mitigation = record.ActionName
		}
		if record.ActionName == "usb_block_device" {
			target = record.DevicePort + " " + record.DeviceProduct
			mitigation = record.ActionName
		}
		response.Items = append(response.Items, contract.ActiveAutomation{
			IncidentID: record.ID, Rule: view, Target: target, Mitigation: mitigation,
			StartedAt: record.StartedAt.UTC().Format(time.RFC3339), ExpiresAt: record.ExpiresAt.UTC().Format(time.RFC3339),
			State: record.State, LastOutcome: record.Outcome,
		})
	}
	return response, nil
}

// setEnabled applies only temporary changes to a compiled startup instance.
func (runtime *automationRuntime) setEnabled(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.RuleSetEnabledRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	rule := runtime.rules[request.RuleID]
	if rule == nil {
		return nil, fmt.Errorf("unknown automation rule %q", request.RuleID)
	}
	if request.Enabled && rule.plugin == nil {
		return nil, errors.New("rule has no initialized FSM; enable it in configuration and restart after enrollment")
	}
	if request.Enabled && !rule.enabled.Load() {
		if err := rule.plugin.Reset(); err != nil {
			return nil, fmt.Errorf("reset rule %q before enable: %w", request.RuleID, err)
		}
	}
	rule.enabled.Store(request.Enabled)
	rule.view.LastOutcome = fmt.Sprintf("temporarily set enabled=%t by %s", request.Enabled, claims.Login)
	pendingRestores := 0
	if !request.Enabled && runtime.incidents != nil && runtime.operatorActions != nil {
		for _, record := range runtime.incidents.list() {
			if record.RuleID != request.RuleID {
				continue
			}
			if record.ActionName == "usb_block_device" || record.ActionName == "revoke_web_sessions" || record.ActionName == "login_terminate_session" {
				continue // USB requires explicit rollback; one-way actions have no inverse.
			}
			if _, err := runtime.queueRestore(ctx, record.ID, claims.Login); err != nil {
				return nil, err
			}
			pendingRestores++
		}
	}
	runtime.log.Warn("automation rule state changed", "rule", request.RuleID, "enabled", request.Enabled, "actor", claims.Login, "reason", request.Reason)
	return contract.OperationResponse{Completed: pendingRestores == 0, PendingEffects: pendingRestores, Outcome: rule.view.LastOutcome}, nil
}

// rollback queues restoration of one service-owned incident.
func (runtime *automationRuntime) rollback(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.RollbackRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	return runtime.queueRestore(ctx, request.IncidentID, claims.Login)
}

// forceNormalize requests the same verified restore path as rollback.
func (runtime *automationRuntime) forceNormalize(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.RollbackRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	return runtime.queueRestore(ctx, request.IncidentID, claims.Login)
}

// queueRestore records the operator intent before its effect worker can act.
func (runtime *automationRuntime) queueRestore(ctx context.Context, id, actor string) (any, error) {
	if runtime.incidents == nil || runtime.operatorActions == nil {
		return nil, errors.New("no enrolled incident effect store is available for restoration")
	}
	for {
		record, ok := runtime.incidents.get(id)
		if !ok {
			return nil, errors.New("unknown automation incident")
		}
		if record.State == contract.StateRestoring {
			return contract.OperationResponse{Completed: false, PendingEffects: 1, Outcome: "restoration already pending"}, nil
		}
		if record.ActionName == "revoke_web_sessions" || record.ActionName == "login_terminate_session" {
			return nil, errors.New("session action is one-way and has no restore operation")
		}
		previous := record.State
		record.State = contract.StateRestoring
		record.Outcome = "operator restoration queued"
		updated, err := runtime.incidents.transition(record, previous)
		if err != nil {
			return nil, err
		}
		if updated {
			break
		}
	}
	select {
	case runtime.operatorActions <- effectCommand{kind: "restore", id: id}:
		runtime.log.Warn("automation operator restoration queued", "incident", id, "actor", actor)
		return contract.OperationResponse{Completed: false, PendingEffects: 1, Outcome: "restoration queued"}, nil
	case <-ctx.Done():
		return nil, ctx.Err() // The timer sees the durable restoring record.
	default:
		return nil, errors.New("effect queue is full; durable restoration will be retried")
	}
}

// forceMitigate queues an operator-confirmed, scoped direct HTTPS source block.
func (runtime *automationRuntime) forceMitigate(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.ForceMitigateRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	runtime.mutex.RLock()
	defer runtime.mutex.RUnlock()
	rule := runtime.rules[request.RuleID]
	if rule == nil {
		return nil, fmt.Errorf("unknown automation rule %q", request.RuleID)
	}
	if runtime.incidents == nil || runtime.operatorActions == nil || runtime.physicalInterfaces == nil || runtime.runUFW == nil || runtime.audit == nil ||
		rule.spec.Action.Kind != "builtin" || rule.spec.Action.Name != "block_source_ip" || !rule.enabled.Load() {
		return nil, errors.New("rule action is not enrolled or enabled")
	}
	if len(request.Operands) != 1 {
		return nil, errors.New("exactly one source IP operand is required")
	}
	source := request.Operands["requester_ip_address"]
	if _, err := validateUFWSource(source); err != nil {
		return nil, err
	}
	return runtime.queueSourceBlock(ctx, request.RuleID, source, "block_source_ip", claims.Login)
}

// queueSourceBlock journals one action-scoped UFW effect before the outbound
// worker touches the host, whether requested by an operator or verified alert.
func (runtime *automationRuntime) queueSourceBlock(ctx context.Context, ruleID, source, actionName, actor string) (contract.OperationResponse, error) {
	if actionName != "block_source_ip" && actionName != "block_scan_source_ip" || runtime.rules[ruleID] == nil || runtime.rules[ruleID].spec.Action.Name != actionName {
		return contract.OperationResponse{}, errors.New("invalid source-block action")
	}
	if _, err := validateUFWSource(source); err != nil {
		return contract.OperationResponse{}, err
	}
	for _, existing := range runtime.incidents.list() {
		if (existing.ActionName == "" || existing.ActionName == "block_source_ip" || existing.ActionName == "block_scan_source_ip") && existing.SourceIP == source {
			return contract.OperationResponse{}, errors.New("source already has an active automation UFW incident")
		}
	}
	interfaces, err := runtime.physicalInterfaces()
	if err != nil {
		return contract.OperationResponse{}, err
	}
	id, err := newIncidentID()
	if err != nil {
		return contract.OperationResponse{}, err
	}
	now := time.Now().UTC()
	record := incidentRecord{
		ID: id, RuleID: ruleID, ActionName: actionName, SourceIP: source, Interfaces: interfaces,
		StartedAt: now, ExpiresAt: now.Add(30 * time.Minute), State: contract.StatePending, Outcome: "operator block queued",
	}
	if actor == "automatic web alert" {
		record.Outcome = "verified direct HTTPS alert block queued"
	}
	if actor == "automatic scan sequence" {
		record.Outcome = "verified scan-sequence source block queued"
	}
	if err := runtime.incidents.save(record); err != nil {
		return contract.OperationResponse{}, err
	}
	select {
	case runtime.operatorActions <- effectCommand{kind: "apply", id: id}:
		runtime.log.Warn("automation mitigation queued", "incident", id, "rule", ruleID, "actor", actor)
		return contract.OperationResponse{Completed: false, PendingEffects: len(interfaces), Outcome: actionName + " queued as incident " + id}, nil
	case <-ctx.Done():
		return contract.OperationResponse{}, ctx.Err() // Pending record is reconciled on restart.
	default:
		record.State = contract.StateFailed
		record.Outcome = "effect queue is full; no UFW command was queued"
		if err := runtime.incidents.save(record); err != nil {
			return contract.OperationResponse{}, err
		}
		return contract.OperationResponse{}, errors.New("effect queue is full")
	}
}

// resetCooldown clears one keyed detector's volatile state; it never changes
// a durable incident or reverses an already applied effect.
func (runtime *automationRuntime) resetCooldown(ctx context.Context, envelope messaging.Envelope) (any, error) {
	request, claims, err := decodeAutomationRequest(ctx, runtime, envelope, func(value contract.CooldownResetRequest) string { return value.AccessToken })
	if err != nil {
		return nil, err
	}
	runtime.mutex.RLock()
	rule := runtime.rules[request.RuleID]
	runtime.mutex.RUnlock()
	if rule == nil {
		return nil, fmt.Errorf("unknown automation rule %q", request.RuleID)
	}
	if rule.plugin == nil {
		return nil, errors.New("rule has no initialized FSM")
	}
	reset, err := rule.plugin.ResetKey(ctx, request.CorrelationKey)
	if err != nil {
		return nil, err
	}
	outcome := "no detection state for key"
	if reset {
		outcome = "detection state reset for key"
	}
	runtime.setRuleOutcome(request.RuleID, outcome)
	runtime.log.Warn("automation detection state reset", "rule", request.RuleID, "key", request.CorrelationKey, "actor", claims.Login, "reset", reset, "reason", request.Reason)
	return contract.OperationResponse{Completed: true, Outcome: outcome}, nil
}
