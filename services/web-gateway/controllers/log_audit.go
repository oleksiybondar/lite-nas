package controllers

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	dto "lite-nas/services/web-gateway/dto/log_audit"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
)

// LogAuditController presents the configured Log Analysis rules to operators.
type LogAuditController struct{ service services.LogAuditService }

// NewLogAuditController binds the rule catalog service.
func NewLogAuditController(service services.LogAuditService) LogAuditController {
	return LogAuditController{service: service}
}

// Rules returns the configured rule catalog for an authenticated security operator.
func (c LogAuditController) Rules(ctx context.Context, _ *dto.EmptyInput) (*dto.LogAuditRulesOutput, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing or invalid access token")
	}
	result, err := c.service.Rules(ctx, token)
	if err != nil {
		slog.Error("log analysis rules RPC failed", "error", err)
		return nil, huma.Error502BadGateway("log analysis service unavailable")
	}
	return dto.NewRulesOutput(result), nil
}
