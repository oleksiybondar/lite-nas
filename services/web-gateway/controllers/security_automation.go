package controllers

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/danielgtaylor/huma/v2"

	dto "lite-nas/services/web-gateway/dto/security_automation"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
	contract "lite-nas/shared/contracts/securityautomation"
)

// SecurityAutomationController adapts authenticated HTTP operations to NATS RPCs.
type SecurityAutomationController struct {
	service services.SecurityAutomationService
}

var operandName = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// validOperands bounds dynamic names and values before the rule validates each operand.
func validOperands(values map[string]string) bool {
	if len(values) > 16 {
		return false
	}
	for name, value := range values {
		if !operandName.MatchString(name) || len(value) < 1 || len(value) > 256 || value[0] == '-' {
			return false
		}
		for _, char := range value {
			if char < 32 || char == 127 {
				return false
			}
		}
	}
	return true
}

// NewSecurityAutomationController binds the operator RPC service.
func NewSecurityAutomationController(service services.SecurityAutomationService) SecurityAutomationController {
	return SecurityAutomationController{service: service}
}

func (c SecurityAutomationController) token(ctx context.Context) (string, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return "", huma.Error401Unauthorized("missing or invalid access token")
	}
	return token, nil
}

func automationRPCError(err error) error {
	slog.Error("security automation RPC failed", "error", err)
	return huma.Error502BadGateway("security automation service unavailable")
}

// Rules returns configured rules, including temporarily disabled rules.
func (c SecurityAutomationController) Rules(ctx context.Context, _ *dto.EmptyInput) (*dto.RulesOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.Rules(ctx, token)
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewRulesOutput(result), nil
}

// Active returns incidents with active or pending effects.
func (c SecurityAutomationController) Active(ctx context.Context, _ *dto.EmptyInput) (*dto.ActiveOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.Active(ctx, token)
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewActiveOutput(result), nil
}

// Rollback requests restoration of the named incident.
func (c SecurityAutomationController) Rollback(ctx context.Context, input *dto.IncidentInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.Rollback(ctx, contract.RollbackRequest{AccessToken: token, IncidentID: input.IncidentID, Reason: input.Body.Reason})
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}

// ForceNormalize requests restoration and closes the incident only after verification.
func (c SecurityAutomationController) ForceNormalize(ctx context.Context, input *dto.IncidentInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.ForceNormalize(ctx, contract.RollbackRequest{AccessToken: token, IncidentID: input.IncidentID, Reason: input.Body.Reason})
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}

// ForceMitigate requests an incident through the rule's validated action path.
func (c SecurityAutomationController) ForceMitigate(ctx context.Context, input *dto.ForceMitigateInput) (*dto.OperationOutput, error) {
	if !validOperands(input.Body.Operands) {
		return nil, huma.Error422UnprocessableEntity("invalid mitigation operands")
	}
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.ForceMitigate(ctx, contract.ForceMitigateRequest{
		AccessToken: token, RuleID: input.RuleID, Reason: input.Body.Reason,
		Operands: input.Body.Operands,
	})
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}

// ResetCooldown clears one named rule's volatile detection state for a key.
func (c SecurityAutomationController) ResetCooldown(ctx context.Context, input *dto.InstanceInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.ResetCooldown(ctx, contract.CooldownResetRequest{AccessToken: token, RuleID: input.RuleID, CorrelationKey: input.Body.CorrelationKey, Reason: input.Body.Reason})
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}

// SetEnabled changes one rule's effective enabled state until restart.
func (c SecurityAutomationController) SetEnabled(ctx context.Context, input *dto.RuleStateInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.SetEnabled(ctx, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: input.RuleID, Enabled: input.Body.Enabled, Reason: input.Body.Reason})
	if err != nil {
		return nil, automationRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}
