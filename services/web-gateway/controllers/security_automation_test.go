package controllers

import (
	"context"
	"errors"
	"testing"

	dto "lite-nas/services/web-gateway/dto/security_automation"
	contract "lite-nas/shared/contracts/securityautomation"
)

// automationControllerStub returns controlled RPC outcomes to the HTTP adapter.
type automationControllerStub struct{ err error }

// Rules returns a configured rule or the selected backend error.
func (s automationControllerStub) Rules(context.Context, string) (contract.RulesListResponse, error) {
	return contract.RulesListResponse{Rules: []contract.Rule{{ID: "rule1"}}}, s.err
}

// Active returns one incident or the selected backend error.
func (s automationControllerStub) Active(context.Context, string) (contract.ActiveListResponse, error) {
	return contract.ActiveListResponse{Items: []contract.ActiveAutomation{{IncidentID: "incident1"}}}, s.err
}

// Rollback returns a controlled operation result.
func (s automationControllerStub) Rollback(context.Context, contract.RollbackRequest) (contract.OperationResponse, error) {
	return contract.OperationResponse{Completed: true}, s.err
}

// ForceNormalize returns a controlled operation result.
func (s automationControllerStub) ForceNormalize(context.Context, contract.RollbackRequest) (contract.OperationResponse, error) {
	return contract.OperationResponse{Completed: true}, s.err
}

// ForceMitigate returns a controlled operation result.
func (s automationControllerStub) ForceMitigate(context.Context, contract.ForceMitigateRequest) (contract.OperationResponse, error) {
	return contract.OperationResponse{Completed: true}, s.err
}

// ResetCooldown returns a controlled operation result.
func (s automationControllerStub) ResetCooldown(context.Context, contract.CooldownResetRequest) (contract.OperationResponse, error) {
	return contract.OperationResponse{Completed: true}, s.err
}

// SetEnabled returns a controlled operation result.
func (s automationControllerStub) SetEnabled(context.Context, contract.RuleSetEnabledRequest) (contract.OperationResponse, error) {
	return contract.OperationResponse{Completed: true}, s.err
}

// automationControllerCases provide one real DTO for each gateway operation.
func automationControllerCases() []struct {
	name string
	run  func(SecurityAutomationController, context.Context) error
} {
	incident := &dto.IncidentInput{IncidentID: "incident1"}
	incident.Body.Reason = "operator request"
	instance := &dto.InstanceInput{RuleID: "rule1"}
	instance.Body.CorrelationKey = "host1"
	state := &dto.RuleStateInput{RuleID: "rule1"}
	state.Body.Enabled = true
	mitigate := &dto.ForceMitigateInput{RuleID: "rule1"}
	mitigate.Body.Operands = map[string]string{"target": "host1"}
	return []struct {
		name string
		run  func(SecurityAutomationController, context.Context) error
	}{
		{"rules", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.Rules(ctx, &dto.EmptyInput{})
			return err
		}},
		{"active", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.Active(ctx, &dto.EmptyInput{})
			return err
		}},
		{"rollback", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.Rollback(ctx, incident)
			return err
		}},
		{"force-normalize", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.ForceNormalize(ctx, incident)
			return err
		}},
		{"force-mitigate", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.ForceMitigate(ctx, mitigate)
			return err
		}},
		{"reset-cooldown", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.ResetCooldown(ctx, instance)
			return err
		}},
		{"set-enabled", func(c SecurityAutomationController, ctx context.Context) error {
			_, err := c.SetEnabled(ctx, state)
			return err
		}},
	}
}

// Requirements: web-gateway/FR-003.
// TestSecurityAutomationControllerOperationsReachBackend checks every HTTP adapter path.
func TestSecurityAutomationControllerOperationsReachBackend(t *testing.T) {
	for _, testCase := range automationControllerCases() {
		t.Run(testCase.name, func(t *testing.T) {
			controller := NewSecurityAutomationController(automationControllerStub{})
			if err := testCase.run(controller, backupContext()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Requirements: web-gateway/FR-003.
// TestSecurityAutomationControllerMapsBackendFailure checks RPC errors remain HTTP errors.
func TestSecurityAutomationControllerMapsBackendFailure(t *testing.T) {
	for _, testCase := range automationControllerCases() {
		t.Run(testCase.name, func(t *testing.T) {
			controller := NewSecurityAutomationController(automationControllerStub{err: errors.New("backend unavailable")})
			if err := testCase.run(controller, backupContext()); err == nil {
				t.Fatal("backend failure was not mapped")
			}
		})
	}
}

// Requirements: web-gateway/FR-004.
// TestSecurityAutomationControllerRequiresToken rejects unauthenticated reads.
func TestSecurityAutomationControllerRequiresToken(t *testing.T) {
	controller := NewSecurityAutomationController(automationControllerStub{})
	_, err := controller.Rules(context.Background(), &dto.EmptyInput{})
	if err == nil {
		t.Fatal("unauthenticated request accepted")
	}
}

// Requirements: web-gateway/FR-003.
// TestSecurityAutomationControllerRejectsUnsafeOperands prevents an invalid force action.
func TestSecurityAutomationControllerRejectsUnsafeOperands(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		operands map[string]string
	}{
		{"option", map[string]string{"target": "--help"}},
		{"invalid name", map[string]string{"bad-name": "host1"}},
		{"control character", map[string]string{"target": "host\n1"}},
		{"empty value", map[string]string{"target": ""}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := &dto.ForceMitigateInput{RuleID: "rule1"}
			input.Body.Operands = testCase.operands
			controller := NewSecurityAutomationController(automationControllerStub{})
			_, err := controller.ForceMitigate(backupContext(), input)
			if err == nil {
				t.Fatal("unsafe operand accepted")
			}
		})
	}
}
