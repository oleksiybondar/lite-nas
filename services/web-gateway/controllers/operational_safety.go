package controllers

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	dto "lite-nas/services/web-gateway/dto/operational_safety"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
	contract "lite-nas/shared/contracts/operationalsafety"
)

// OperationalSafetyController adapts authenticated operator HTTP requests to RPCs.
type OperationalSafetyController struct {
	service services.OperationalSafetyService
}

// NewOperationalSafetyController binds the backend operator service.
func NewOperationalSafetyController(service services.OperationalSafetyService) OperationalSafetyController {
	return OperationalSafetyController{service: service}
}

// token retrieves the already validated access token from the HTTP context.
func (c OperationalSafetyController) token(ctx context.Context) (string, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return "", huma.Error401Unauthorized("missing or invalid access token")
	}
	return token, nil
}

// operationalSafetyRPCError maps backend transport failures to one gateway response.
func operationalSafetyRPCError(err error) error {
	slog.Error("operational safety RPC failed", "error", err)
	return huma.Error502BadGateway("operational safety service unavailable")
}

// Rules returns all configured rules and their temporary enabled state.
func (c OperationalSafetyController) Rules(ctx context.Context, _ *dto.EmptyInput) (*dto.RulesOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.Rules(ctx, token)
	if err != nil {
		return nil, operationalSafetyRPCError(err)
	}
	return dto.NewRulesOutput(result), nil
}

// Tasks returns active in-memory tasks.
func (c OperationalSafetyController) Tasks(ctx context.Context, _ *dto.EmptyInput) (*dto.TasksOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.Tasks(ctx, token)
	if err != nil {
		return nil, operationalSafetyRPCError(err)
	}
	return dto.NewTasksOutput(result), nil
}

// SetEnabled changes a rule's effective enabled state until restart.
func (c OperationalSafetyController) SetEnabled(ctx context.Context, input *dto.RuleStateInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.SetEnabled(ctx, contract.OperationalSafetyRuleSetEnabledRequest{AccessToken: token, RuleID: input.RuleID, Enabled: input.Body.Enabled, Reason: input.Body.Reason})
	if err != nil {
		return nil, operationalSafetyRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}

// ForceSet requests one rule-declared effect level on an existing task.
func (c OperationalSafetyController) ForceSet(ctx context.Context, input *dto.ForceSetInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.ForceSet(ctx, contract.OperationalSafetyTaskForceSetRequest{AccessToken: token, TaskID: input.TaskID, Level: input.Body.Level, Reason: input.Body.Reason})
	if err != nil {
		return nil, operationalSafetyRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}

// ForceStop starts staged recovery for an existing task.
func (c OperationalSafetyController) ForceStop(ctx context.Context, input *dto.ForceStopInput) (*dto.OperationOutput, error) {
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.service.ForceStop(ctx, contract.OperationalSafetyTaskForceStopRequest{AccessToken: token, TaskID: input.TaskID, Reason: input.Body.Reason})
	if err != nil {
		return nil, operationalSafetyRPCError(err)
	}
	return dto.NewOperationOutput(result), nil
}
