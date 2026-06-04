package controllers

import (
	"context"
	"errors"
	"time"

	alertsdto "lite-nas/services/web-gateway/dto/alerts"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
	"lite-nas/shared/authtoken"

	"github.com/danielgtaylor/huma/v2"
)

// AlertsController translates one alert-domain HTTP surface into backend service calls.
type AlertsController struct {
	service services.AlertsService
}

// NewSystemAlertsController creates the browser-facing controller for system alerts.
func NewSystemAlertsController(service services.AlertsService) AlertsController {
	return AlertsController{service: service}
}

// NewSecurityAlertsController creates the browser-facing controller for security alerts.
func NewSecurityAlertsController(service services.AlertsService) AlertsController {
	return AlertsController{service: service}
}

// List returns one browser-facing page of alerts from the configured domain.
func (c AlertsController) List(ctx context.Context, input *alertsdto.ListInput) (*alertsdto.ListOutput, error) {
	now := time.Now()
	request, err := extractAlertListInput(ctx, input)
	if err != nil {
		return nil, err
	}

	page, err := c.service.List(ctx, request)
	if err != nil {
		return nil, mapAlertBackendError(err, "failed to fetch alerts")
	}

	return &alertsdto.ListOutput{
		Body: alertsdto.NewListBody(now, page.Items, newListMetadata(request.Page, request.Size, page.TotalCount)),
	}, nil
}

// ListUnacknowledged returns one browser-facing page of unacknowledged alerts from the configured domain.
func (c AlertsController) ListUnacknowledged(ctx context.Context, input *alertsdto.ListInput) (*alertsdto.ListOutput, error) {
	now := time.Now()
	request, err := extractAlertListInput(ctx, input)
	if err != nil {
		return nil, err
	}

	page, err := c.service.ListUnacknowledged(ctx, request)
	if err != nil {
		return nil, mapAlertBackendError(err, "failed to fetch unacknowledged alerts")
	}

	return &alertsdto.ListOutput{
		Body: alertsdto.NewListBody(now, page.Items, newListMetadata(request.Page, request.Size, page.TotalCount)),
	}, nil
}

// Count returns the simplified count of all alerts in the configured domain.
func (c AlertsController) Count(ctx context.Context, input *alertsdto.CountInput) (*alertsdto.CountOutput, error) {
	now := time.Now()
	request, err := extractAlertCountInput(ctx, input)
	if err != nil {
		return nil, err
	}

	page, err := c.service.List(ctx, request)
	if err != nil {
		return nil, mapAlertBackendError(err, "failed to fetch alerts count")
	}

	return &alertsdto.CountOutput{Body: alertsdto.NewCountBody(now, page.TotalCount)}, nil
}

// CountUnacknowledged returns the simplified count of unacknowledged alerts in the configured domain.
func (c AlertsController) CountUnacknowledged(ctx context.Context, input *alertsdto.CountInput) (*alertsdto.CountOutput, error) {
	now := time.Now()
	request, err := extractAlertCountInput(ctx, input)
	if err != nil {
		return nil, err
	}

	page, err := c.service.ListUnacknowledged(ctx, request)
	if err != nil {
		return nil, mapAlertBackendError(err, "failed to fetch unacknowledged alerts count")
	}

	return &alertsdto.CountOutput{Body: alertsdto.NewCountBody(now, page.TotalCount)}, nil
}

// Get returns one browser-facing alert detail from the configured domain.
func (c AlertsController) Get(ctx context.Context, input *alertsdto.GetInput) (*alertsdto.GetOutput, error) {
	now := time.Now()
	request, err := extractAlertGetInput(ctx, input)
	if err != nil {
		return nil, err
	}

	item, found, err := c.service.Get(ctx, request)
	if err != nil {
		return nil, mapAlertBackendError(err, "failed to fetch alert")
	}
	if !found {
		return nil, huma.Error404NotFound("alert not found")
	}

	return &alertsdto.GetOutput{Body: alertsdto.NewGetBody(now, item)}, nil
}

// Acknowledge acknowledges one alert in the configured domain.
func (c AlertsController) Acknowledge(ctx context.Context, input *alertsdto.ActionInput) (*alertsdto.ActionOutput, error) {
	now := time.Now()
	request, err := extractAlertActionInput(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := c.service.Acknowledge(ctx, request); err != nil {
		return nil, mapAlertBackendError(err, "failed to acknowledge alert")
	}

	return &alertsdto.ActionOutput{Body: alertsdto.NewActionBody(now, "alert acknowledged")}, nil
}

// Mute mutes one alert in the configured domain.
func (c AlertsController) Mute(ctx context.Context, input *alertsdto.ActionInput) (*alertsdto.ActionOutput, error) {
	now := time.Now()
	request, err := extractAlertActionInput(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := c.service.Mute(ctx, request); err != nil {
		return nil, mapAlertBackendError(err, "failed to mute alert")
	}

	return &alertsdto.ActionOutput{Body: alertsdto.NewActionBody(now, "alert muted")}, nil
}

func extractAlertListInput(ctx context.Context, input *alertsdto.ListInput) (services.AlertListInput, error) {
	accessToken, _, err := extractAuthenticatedAlertPrincipal(ctx)
	if err != nil {
		return services.AlertListInput{}, err
	}

	page := alertsdto.DefaultPage
	size := alertsdto.DefaultSize
	if input != nil && input.Page > 0 {
		page = input.Page
	}
	if input != nil && input.Size > 0 {
		size = input.Size
	}

	return services.AlertListInput{
		AccessToken: accessToken,
		Page:        page,
		Size:        size,
	}, nil
}

func extractAlertCountInput(ctx context.Context, input *alertsdto.CountInput) (services.AlertListInput, error) {
	accessToken, _, err := extractAuthenticatedAlertPrincipal(ctx)
	if err != nil {
		return services.AlertListInput{}, err
	}

	page := alertsdto.DefaultPage
	size := alertsdto.DefaultSize
	if input != nil && input.Page > 0 {
		page = input.Page
	}
	if input != nil && input.Size > 0 {
		size = input.Size
	}

	return services.AlertListInput{AccessToken: accessToken, Page: page, Size: size}, nil
}

func extractAlertGetInput(ctx context.Context, input *alertsdto.GetInput) (services.AlertGetInput, error) {
	accessToken, _, err := extractAuthenticatedAlertPrincipal(ctx)
	if err != nil {
		return services.AlertGetInput{}, err
	}

	return services.AlertGetInput{AccessToken: accessToken, ID: input.ID}, nil
}

func extractAlertActionInput(ctx context.Context, input *alertsdto.ActionInput) (services.AlertActionInput, error) {
	accessToken, claims, err := extractAuthenticatedAlertPrincipal(ctx)
	if err != nil {
		return services.AlertActionInput{}, err
	}

	return services.AlertActionInput{
		AccessToken: accessToken,
		ID:          input.ID,
		ActorLogin:  claims.Login,
	}, nil
}

func extractAuthenticatedAlertPrincipal(ctx context.Context) (string, authtoken.AccessClaims, error) {
	accessToken, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return "", authtoken.AccessClaims{}, huma.Error401Unauthorized("missing or invalid access token")
	}

	claims, ok := middlewares.AccessClaimsFromContext(ctx)
	if !ok {
		return "", authtoken.AccessClaims{}, huma.Error401Unauthorized("missing or invalid access token")
	}
	if claims.Login == "" {
		return "", authtoken.AccessClaims{}, huma.Error401Unauthorized("missing or invalid access token")
	}

	return accessToken, claims, nil
}

func newListMetadata(page int, size int, totalCount int) alertsdto.ListMetadata {
	totalPages := 0
	if totalCount > 0 && size > 0 {
		totalPages = (totalCount + size - 1) / size
	}

	return alertsdto.ListMetadata{
		Page:       page,
		Size:       size,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

func mapAlertBackendError(err error, message string) error {
	if errors.Is(err, services.ErrAlertActionFailed) {
		return huma.Error502BadGateway(message)
	}

	return huma.Error502BadGateway(message)
}
