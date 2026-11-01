package controllers

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	securitydto "lite-nas/services/web-gateway/dto/security_tasks"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
)

// SecurityTasksController adapts authorized HTTP calls to security task services.
type SecurityTasksController struct{ service services.SecurityTasksService }

// NewSecurityTasksController binds the gateway security task service.
func NewSecurityTasksController(service services.SecurityTasksService) SecurityTasksController {
	return SecurityTasksController{service: service}
}

// StartAntivirusScan queues a scan for the validated folder request.
func (controller SecurityTasksController) StartAntivirusScan(ctx context.Context, input *securitydto.StartAntivirusInput) (*securitydto.StartOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.StartAntivirusScan(ctx, token, input.Folder)
	if err != nil {
		slog.Error("antivirus scan start RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to start antivirus scan: " + err.Error())
	}
	return securitydto.NewStartOutput(result), nil
}

// StartIntegrityCheck queues an AIDE check.
func (controller SecurityTasksController) StartIntegrityCheck(ctx context.Context, _ *securitydto.EmptyInput) (*securitydto.StartOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.StartIntegrityCheck(ctx, token)
	if err != nil {
		slog.Error("integrity check start RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to start integrity check")
	}
	return securitydto.NewStartOutput(result), nil
}

// StartIntegrityBaseline queues an AIDE baseline generation task.
func (controller SecurityTasksController) StartIntegrityBaseline(ctx context.Context, _ *securitydto.EmptyInput) (*securitydto.StartOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.StartIntegrityBaseline(ctx, token)
	if err != nil {
		slog.Error("integrity baseline start RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to start integrity baseline generation")
	}
	return securitydto.NewStartOutput(result), nil
}

// AntivirusTaskStatus returns the current scan attempt snapshot.
func (controller SecurityTasksController) AntivirusTaskStatus(ctx context.Context, input *securitydto.TaskIDInput) (*securitydto.TaskOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.AntivirusTaskStatus(ctx, token, input.TaskID)
	if err != nil {
		slog.Error("antivirus task status RPC failed", "task_id", input.TaskID, "error", err)
		return nil, huma.Error502BadGateway("failed to read antivirus task status")
	}
	return securitydto.NewTaskOutput(result), nil
}

// IntegrityTaskStatus returns the current AIDE attempt snapshot.
func (controller SecurityTasksController) IntegrityTaskStatus(ctx context.Context, input *securitydto.TaskIDInput) (*securitydto.TaskOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.IntegrityTaskStatus(ctx, token, input.TaskID)
	if err != nil {
		slog.Error("integrity task status RPC failed", "task_id", input.TaskID, "error", err)
		return nil, huma.Error502BadGateway("failed to read integrity task status")
	}
	return securitydto.NewTaskOutput(result), nil
}

// AntivirusInfo returns compact ClamAV signature state and optional diagnostic details.
func (controller SecurityTasksController) AntivirusInfo(ctx context.Context, _ *securitydto.EmptyInput) (*securitydto.AntivirusInfoOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.AntivirusInfo(ctx, token)
	if err != nil {
		slog.Error("antivirus info RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to read antivirus information")
	}
	return securitydto.NewAntivirusInfoOutput(result), nil
}

// UpdateAntivirusSignatures queues a ClamAV database update for the authorized user.
func (controller SecurityTasksController) UpdateAntivirusSignatures(ctx context.Context, _ *securitydto.EmptyInput) (*securitydto.AntivirusUpdateOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.UpdateAntivirusSignatures(ctx, token)
	if err != nil {
		slog.Error("antivirus signature update RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to update antivirus signatures")
	}
	return securitydto.NewAntivirusUpdateOutput(result), nil
}

// IntegrityInfo returns compact AIDE baseline availability and optional metadata.
func (controller SecurityTasksController) IntegrityInfo(ctx context.Context, _ *securitydto.EmptyInput) (*securitydto.IntegrityInfoOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.IntegrityInfo(ctx, token)
	if err != nil {
		slog.Error("integrity info RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to read integrity information")
	}
	return securitydto.NewIntegrityInfoOutput(result), nil
}

// RetryAntivirusTask starts another attempt of a completed scan.
func (controller SecurityTasksController) RetryAntivirusTask(ctx context.Context, input *securitydto.TaskIDInput) (*securitydto.RetryOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.RetryAntivirusTask(ctx, token, input.TaskID)
	if err != nil {
		slog.Error("antivirus task retry RPC failed", "task_id", input.TaskID, "error", err)
		return nil, huma.Error502BadGateway("failed to retry antivirus scan")
	}
	return securitydto.NewRetryOutput(result), nil
}

// RetryIntegrityTask starts another attempt of a completed AIDE check.
func (controller SecurityTasksController) RetryIntegrityTask(ctx context.Context, input *securitydto.TaskIDInput) (*securitydto.RetryOutput, error) {
	token, err := controller.token(ctx)
	if err != nil {
		return nil, err
	}
	result, err := controller.service.RetryIntegrityTask(ctx, token, input.TaskID)
	if err != nil {
		slog.Error("integrity task retry RPC failed", "task_id", input.TaskID, "error", err)
		return nil, huma.Error502BadGateway("failed to retry integrity check")
	}
	return securitydto.NewRetryOutput(result), nil
}

// token extracts the authenticated access token added by gateway middleware.
func (controller SecurityTasksController) token(ctx context.Context) (string, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return "", huma.Error401Unauthorized("missing or invalid access token")
	}
	return token, nil
}
