package controllers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/dto/management"
	"lite-nas/services/web-gateway/services"
)

// AuthSessionsController adapts administrator HTTP requests to Auth Service RPCs.
type AuthSessionsController struct{ service services.AuthSessionsService }

// NewAuthSessionsController constructs the authentication session controller.
func NewAuthSessionsController(service services.AuthSessionsService) AuthSessionsController {
	return AuthSessionsController{service: service}
}

// List returns active LiteNAS authentication sessions.
func (c AuthSessionsController) List(ctx context.Context, _ *management.ListInput) (*management.AuthSessionsOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.service.List(ctx, token)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to list authentication sessions")
	}
	return &management.AuthSessionsOutput{Body: management.NewAuthSessionsBody(items)}, nil
}

// Revoke invalidates the selected LiteNAS authentication session.
func (c AuthSessionsController) Revoke(ctx context.Context, input *management.AuthSessionActionInput) (*management.AuthSessionRevokeOutput, error) {
	token, err := managementToken(ctx)
	if err != nil {
		return nil, err
	}
	revoked, err := c.service.Revoke(ctx, token, input.SessionID)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to revoke authentication session")
	}
	return &management.AuthSessionRevokeOutput{Body: management.NewAuthSessionRevokeBody(revoked)}, nil
}
