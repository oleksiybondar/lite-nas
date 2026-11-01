package controllers

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/dto/management"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
)

// ServicesManagementController adapts authenticated HTTP requests to system service commands.
type ServicesManagementController struct {
	service services.ServicesManagementService
}

// SessionsManagementController adapts authenticated HTTP requests to host session commands.
type SessionsManagementController struct {
	service services.SessionsManagementService
}

// USBManagerController adapts authenticated HTTP requests to USBGuard commands.
type USBManagerController struct {
	service services.USBManagementService
}

// NewServicesManagementController constructs the system service controller.
func NewServicesManagementController(service services.ServicesManagementService) ServicesManagementController {
	return ServicesManagementController{service}
}

// NewSessionsManagementController constructs the host session controller.
func NewSessionsManagementController(service services.SessionsManagementService) SessionsManagementController {
	return SessionsManagementController{service}
}

// NewUSBManagerController constructs the USBGuard management controller.
func NewUSBManagerController(service services.USBManagementService) USBManagerController {
	return USBManagerController{service}
}

func managementToken(ctx context.Context) (string, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return "", huma.Error401Unauthorized("missing or invalid access token")
	}
	return token, nil
}

// List returns loaded and installed systemd units.
func (c ServicesManagementController) List(ctx context.Context, _ *management.ListInput) (*management.ServicesOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.service.List(ctx, token)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to list services")
	}
	return &management.ServicesOutput{Body: management.NewServicesBody(items)}, nil
}

// List returns the current host login sessions.
func (c SessionsManagementController) List(ctx context.Context, _ *management.ListInput) (*management.SessionsOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.service.List(ctx, token)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to list sessions")
	}
	return &management.SessionsOutput{Body: management.NewSessionsBody(items)}, nil
}

func (c ServicesManagementController) action(ctx context.Context, input *management.ServiceActionInput, action string) (*management.ManagementActionOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.service.Action(ctx, token, input.Unit, action); err != nil {
		return nil, huma.Error502BadGateway("failed to " + action + " service")
	}
	return &management.ManagementActionOutput{Body: management.NewManagementActionBody()}, nil
}

func (c SessionsManagementController) action(ctx context.Context, input *management.SessionActionInput, action string) (*management.ManagementActionOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.service.Action(ctx, token, input.Session, action); err != nil {
		return nil, huma.Error502BadGateway("failed to " + action + " session")
	}
	return &management.ManagementActionOutput{Body: management.NewManagementActionBody()}, nil
}

// List returns the current USBGuard device inventory.
func (c USBManagerController) List(ctx context.Context, _ *management.ListInput) (*management.USBDevicesOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.service.List(ctx, token)
	if err != nil {
		slog.Error("USBGuard inventory RPC failed", "error", err)
		return nil, huma.Error502BadGateway("failed to list USBGuard devices")
	}
	return &management.USBDevicesOutput{Body: management.NewUSBDevicesBody(items)}, nil
}

// action authorizes the caller and forwards one validated USBGuard decision.
func (c USBManagerController) action(ctx context.Context, input *management.USBDeviceActionInput, action string) (*management.ManagementActionOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.service.Action(ctx, token, input.DeviceID, action, input.Permanent); err != nil {
		slog.Error("USBGuard device action RPC failed", "action", action, "device_id", input.DeviceID, "permanent", input.Permanent, "error", err)
		return nil, huma.Error502BadGateway("failed to " + action + " USBGuard device")
	}
	return &management.ManagementActionOutput{Body: management.NewManagementActionBody()}, nil
}

// Start starts one system service unit.
func (c ServicesManagementController) Start(ctx context.Context, input *management.ServiceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "start")
}

// Stop stops one system service unit.
func (c ServicesManagementController) Stop(ctx context.Context, input *management.ServiceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "stop")
}

// Restart restarts one system service unit.
func (c ServicesManagementController) Restart(ctx context.Context, input *management.ServiceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "restart")
}

// Enable enables one system service unit at boot.
func (c ServicesManagementController) Enable(ctx context.Context, input *management.ServiceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "enable")
}

// Disable disables one system service unit at boot.
func (c ServicesManagementController) Disable(ctx context.Context, input *management.ServiceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "disable")
}

// Terminate ends one host login session.
func (c SessionsManagementController) Terminate(ctx context.Context, input *management.SessionActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "terminate")
}

// Lock locks one host login session.
func (c SessionsManagementController) Lock(ctx context.Context, input *management.SessionActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "lock")
}

// Unlock unlocks one host login session.
func (c SessionsManagementController) Unlock(ctx context.Context, input *management.SessionActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "unlock")
}

// Allow authorizes one USBGuard device, optionally persisting the decision.
func (c USBManagerController) Allow(ctx context.Context, input *management.USBDeviceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "allow")
}

// Block deauthorizes one USBGuard device, optionally persisting the decision.
func (c USBManagerController) Block(ctx context.Context, input *management.USBDeviceActionInput) (*management.ManagementActionOutput, error) {
	return c.action(ctx, input, "block")
}
