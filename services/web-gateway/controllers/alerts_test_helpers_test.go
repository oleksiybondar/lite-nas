package controllers

import (
	"context"
	"errors"
	"testing"

	alertsdto "lite-nas/services/web-gateway/dto/alerts"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/services"
	"lite-nas/shared/authtoken"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"

	"github.com/danielgtaylor/huma/v2"
)

func authenticatedAlertsContext() context.Context {
	return middlewares.NewAuthenticatedContext(
		context.Background(),
		"AT",
		authtoken.AccessClaims{Login: "john.doe"},
	)
}

func assertSuccessfulAlertResponse(t *testing.T, success bool, timestampIsZero bool) {
	t.Helper()

	if !success {
		t.Fatal("expected success=true")
	}
	if timestampIsZero {
		t.Fatal("expected timestamp to be populated")
	}
}

func assertUnacknowledgedListInput(t *testing.T, input services.AlertListInput) {
	t.Helper()

	if input.AccessToken != "AT" || input.Page != alertsdto.DefaultPage || input.Size != alertsdto.DefaultSize {
		t.Fatalf("list input = %#v, want normalized auth and pagination", input)
	}
}

func assertActiveListInput(t *testing.T, input services.AlertListInput) {
	t.Helper()

	if input.AccessToken != "AT" || input.Page != alertsdto.DefaultPage || input.Size != alertsdto.DefaultSize {
		t.Fatalf("active input = %#v, want normalized auth and pagination", input)
	}
}

func assertActionPrincipal(t *testing.T, input services.AlertActionInput) {
	t.Helper()

	if input.AccessToken != "AT" || input.ID != "evt-1" || input.ActorLogin != "john.doe" {
		t.Fatalf("action input = %#v, want forwarded token, id, and login", input)
	}
}

func assertFoundAlertOutput(t *testing.T, output *alertsdto.GetOutput) {
	t.Helper()

	assertSuccessfulAlertResponse(t, output.Body.Success, output.Body.Timestamp.IsZero())
	if output.Body.Data.EventID != "evt-1" {
		t.Fatalf("event id = %q, want evt-1", output.Body.Data.EventID)
	}
}

func humaStatus(err error) int {
	var typed *huma.ErrorModel
	if !errors.As(err, &typed) {
		return 0
	}
	return typed.Status
}

func runAlertsListFlowTest(
	t *testing.T,
	name string,
	invoke func(AlertsController, context.Context) (*alertsdto.ListOutput, error),
	assertInput func(*testing.T, *stubAlertsService),
) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		service, controller, ctx := newAlertsControllerTestContext(services.AlertListPage{
			Items:      []loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
			TotalCount: 21,
		})

		output, err := invoke(controller, ctx)
		if err != nil {
			t.Fatalf("%s error = %v", name, err)
		}

		assertSuccessfulAlertResponse(t, output.Body.Success, output.Body.Timestamp.IsZero())
		assertInput(t, service)
		if output.Body.Data.Metadata.TotalPages != 2 {
			t.Fatalf("total_pages = %d, want 2", output.Body.Data.Metadata.TotalPages)
		}
	})
}

func runAlertsCountFlowTest(
	t *testing.T,
	name string,
	invoke func(AlertsController, context.Context) (*alertsdto.CountOutput, error),
	assertInput func(*testing.T, *stubAlertsService),
) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		service, controller, ctx := newAlertsControllerTestContext(services.AlertListPage{TotalCount: 9})

		output, err := invoke(controller, ctx)
		if err != nil {
			t.Fatalf("%s error = %v", name, err)
		}
		if output.Body.Data.Count != 9 {
			t.Fatalf("count = %d, want 9", output.Body.Data.Count)
		}
		assertInput(t, service)
	})
}

func newAlertsControllerTestContext(
	listPage services.AlertListPage,
) (*stubAlertsService, AlertsController, context.Context) {
	service := &stubAlertsService{listPage: listPage}
	return service, NewSystemAlertsController(service), authenticatedAlertsContext()
}
