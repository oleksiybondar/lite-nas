package controllers

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/services/web-gateway/dto/journal"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
	contract "lite-nas/shared/contracts/journalmanager"
)

// JournalController adapts security-gated HTTP requests to journal RPCs.
type JournalController struct{ service services.JournalService }

// NewJournalController constructs the browser-facing journal controller.
func NewJournalController(service services.JournalService) JournalController {
	return JournalController{service: service}
}

// Query returns a chunk of the whole host journal.
func (c JournalController) Query(ctx context.Context, input *journal.QueryInput) (*journal.QueryOutput, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing or invalid access token")
	}
	if input.Since != "" && input.Until != "" {
		since, err1 := time.Parse(time.RFC3339, input.Since)
		until, err2 := time.Parse(time.RFC3339, input.Until)
		if err1 != nil || err2 != nil || since.After(until) {
			return nil, huma.Error422UnprocessableEntity("invalid journal time range")
		}
	}
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}
	var priority *int
	if input.Priority != "" {
		value, err := strconv.Atoi(input.Priority)
		if err != nil || value < 0 || value > 7 {
			return nil, huma.Error422UnprocessableEntity("invalid journal priority")
		}
		priority = &value
	}
	response, err := c.service.Query(ctx, contract.QueryRequest{AccessToken: token, Since: input.Since, Until: input.Until, Message: input.Message, Unit: input.Unit, Priority: priority, Cursor: input.Cursor, Limit: limit})
	if err != nil {
		slog.Error("journal query failed", "error", err)
		return nil, huma.Error502BadGateway("failed to query journal")
	}
	return &journal.QueryOutput{Body: journal.NewQueryBody(response)}, nil
}

// Units returns service units with entries matching the current search filters.
func (c JournalController) Units(ctx context.Context, input *journal.UnitsInput) (*journal.UnitsOutput, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing or invalid access token")
	}
	if input.Since != "" && input.Until != "" {
		since, err1 := time.Parse(time.RFC3339, input.Since)
		until, err2 := time.Parse(time.RFC3339, input.Until)
		if err1 != nil || err2 != nil || since.After(until) {
			return nil, huma.Error422UnprocessableEntity("invalid journal time range")
		}
	}
	var priority *int
	if input.Priority != "" {
		value, err := strconv.Atoi(input.Priority)
		if err != nil || value < 0 || value > 7 {
			return nil, huma.Error422UnprocessableEntity("invalid journal priority")
		}
		priority = &value
	}
	response, err := c.service.Units(ctx, contract.UnitsRequest{AccessToken: token, Since: input.Since, Until: input.Until, Message: input.Message, Priority: priority})
	if err != nil {
		slog.Error("journal unit query failed", "error", err)
		return nil, huma.Error502BadGateway("failed to query journal units")
	}
	return &journal.UnitsOutput{Body: journal.NewUnitsBody(response)}, nil
}

// Vacuum applies the requested retention and reports completion.
func (c JournalController) Vacuum(ctx context.Context, input *journal.VacuumInput) (*journal.VacuumOutput, error) {
	token, ok := middlewares.AccessTokenFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("missing or invalid access token")
	}
	if err := c.service.Vacuum(ctx, contract.VacuumRequest{AccessToken: token, Days: input.Body.Days}); err != nil {
		slog.Error("journal vacuum failed", "error", err)
		return nil, huma.Error502BadGateway("journal vacuum failed")
	}
	return &journal.VacuumOutput{Body: journal.NewVacuumBody()}, nil
}
