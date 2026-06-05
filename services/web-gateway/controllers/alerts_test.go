package controllers

import (
	"context"
	"testing"

	alertsdto "lite-nas/services/web-gateway/dto/alerts"
	"lite-nas/services/web-gateway/services"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
)

type stubAlertsService struct {
	listPage            services.AlertListPage
	listErr             error
	item                loggingmanagercontract.ListAlertItem
	found               bool
	getErr              error
	actionErr           error
	listInput           services.AlertListInput
	activeInput         services.AlertListInput
	unacknowledgedInput services.AlertListInput
	getInput            services.AlertGetInput
	actionInput         services.AlertActionInput
}

func (s *stubAlertsService) List(_ context.Context, input services.AlertListInput) (services.AlertListPage, error) {
	s.listInput = input
	if s.listErr != nil {
		return services.AlertListPage{}, s.listErr
	}
	return s.listPage, nil
}

func (s *stubAlertsService) ListActive(_ context.Context, input services.AlertListInput) (services.AlertListPage, error) {
	s.activeInput = input
	if s.listErr != nil {
		return services.AlertListPage{}, s.listErr
	}
	return s.listPage, nil
}

func (s *stubAlertsService) ListUnacknowledged(_ context.Context, input services.AlertListInput) (services.AlertListPage, error) {
	s.unacknowledgedInput = input
	if s.listErr != nil {
		return services.AlertListPage{}, s.listErr
	}
	return s.listPage, nil
}

func (s *stubAlertsService) Get(_ context.Context, input services.AlertGetInput) (loggingmanagercontract.ListAlertItem, bool, error) {
	s.getInput = input
	if s.getErr != nil {
		return loggingmanagercontract.ListAlertItem{}, false, s.getErr
	}
	return s.item, s.found, nil
}

func (s *stubAlertsService) Acknowledge(_ context.Context, input services.AlertActionInput) error {
	s.actionInput = input
	return s.actionErr
}

func (s *stubAlertsService) Mute(_ context.Context, input services.AlertActionInput) error {
	s.actionInput = input
	return s.actionErr
}

func TestAlertsControllerListActiveWrapsMetadataAndDefaultsPagination(t *testing.T) {
	runAlertsListFlowTest(
		t,
		"ListActive",
		func(controller AlertsController, ctx context.Context) (*alertsdto.ListOutput, error) {
			return controller.ListActive(ctx, &alertsdto.ListInput{})
		},
		func(t *testing.T, service *stubAlertsService) {
			assertActiveListInput(t, service.activeInput)
		},
	)
}

func TestAlertsControllerListUnacknowledgedWrapsMetadataAndDefaultsPagination(t *testing.T) {
	runAlertsListFlowTest(
		t,
		"ListUnacknowledged",
		func(controller AlertsController, ctx context.Context) (*alertsdto.ListOutput, error) {
			return controller.ListUnacknowledged(ctx, &alertsdto.ListInput{})
		},
		func(t *testing.T, service *stubAlertsService) {
			assertUnacknowledgedListInput(t, service.unacknowledgedInput)
		},
	)
}

func TestAlertsControllerListForwardsParsedFilters(t *testing.T) {
	t.Parallel()

	service, controller, ctx := newAlertsControllerTestContext(services.AlertListPage{})
	output, err := controller.List(ctx, &alertsdto.ListInput{
		Filters: []string{
			`{"key":"category","condition":"eq","values":["system.metrics.mem.used"]}`,
			`{"key":"created_at","condition":"between","values":["2026-05-12T10:00:00Z","2026-05-12T11:00:00Z"]}`,
		},
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if output == nil {
		t.Fatal("List() output = nil, want response")
	}

	assertListFilters(t, service.listInput, []loggingmanagerdto.Filter{
		{
			Key:       loggingmanagerdto.FilterKeyCategory,
			Condition: loggingmanagerdto.FilterConditionEQ,
			Values:    []string{"system.metrics.mem.used"},
		},
		{
			Key:       loggingmanagerdto.FilterKeyCreatedAt,
			Condition: loggingmanagerdto.FilterConditionBetween,
			Values:    []string{"2026-05-12T10:00:00Z", "2026-05-12T11:00:00Z"},
		},
	})
}

func TestAlertsControllerListRejectsInvalidFilters(t *testing.T) {
	t.Parallel()

	controller := NewSystemAlertsController(&stubAlertsService{})
	ctx := authenticatedAlertsContext()

	output, err := controller.List(ctx, &alertsdto.ListInput{
		Filters: []string{`{"key":"created_at","condition":"between","values":["2026-05-12T10:00:00Z"]}`},
	})
	if output != nil {
		t.Fatalf("List() output = %#v, want nil", output)
	}
	if err == nil || humaStatus(err) != 422 {
		t.Fatalf("List() error = %v, want status 422", err)
	}
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestAlertsControllerGetMapsMissingAlertToNotFound(t *testing.T) {
	t.Parallel()

	controller := NewSystemAlertsController(&stubAlertsService{found: false})
	ctx := authenticatedAlertsContext()

	output, err := controller.Get(ctx, &alertsdto.GetInput{ID: "evt-1"})
	if output != nil {
		t.Fatalf("Get() output = %#v, want nil", output)
	}
	if err == nil || humaStatus(err) != 404 {
		t.Fatalf("Get() error = %v, want status 404", err)
	}
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestAlertsControllerGetWrapsFoundAlert(t *testing.T) {
	t.Parallel()

	service := &stubAlertsService{
		item:  loggingmanagercontract.ListAlertItem{EventID: "evt-1"},
		found: true,
	}
	controller := NewSystemAlertsController(service)
	ctx := authenticatedAlertsContext()

	output, err := controller.Get(ctx, &alertsdto.GetInput{ID: "evt-1"})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	assertFoundAlertOutput(t, output)
	if service.getInput.AccessToken != "AT" || service.getInput.ID != "evt-1" {
		t.Fatalf("get input = %#v, want forwarded token and id", service.getInput)
	}
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestAlertsControllerAcknowledgeUsesAuthenticatedPrincipal(t *testing.T) {
	t.Parallel()

	service := &stubAlertsService{}
	controller := NewSystemAlertsController(service)
	ctx := authenticatedAlertsContext()

	output, err := controller.Acknowledge(ctx, &alertsdto.ActionInput{ID: "evt-1"})
	if err != nil {
		t.Fatalf("Acknowledge() error = %v", err)
	}
	assertSuccessfulAlertResponse(t, output.Body.Success, output.Body.Timestamp.IsZero())
	if output.Body.Message != "alert acknowledged" {
		t.Fatalf("message = %q, want alert acknowledged", output.Body.Message)
	}
	assertActionPrincipal(t, service.actionInput)
}

func TestAlertsControllerCountUsesTotalCount(t *testing.T) {
	runAlertsCountFlowTest(
		t,
		"Count",
		func(controller AlertsController, ctx context.Context) (*alertsdto.CountOutput, error) {
			return controller.Count(ctx, &alertsdto.CountInput{Page: 3, Size: 7})
		},
		func(t *testing.T, service *stubAlertsService) {
			if service.listInput.Page != 3 || service.listInput.Size != 7 {
				t.Fatalf("list input = %#v, want forwarded page and size", service.listInput)
			}
		},
	)
}

func TestAlertsControllerCountActiveUsesTotalCount(t *testing.T) {
	runAlertsCountFlowTest(
		t,
		"CountActive",
		func(controller AlertsController, ctx context.Context) (*alertsdto.CountOutput, error) {
			return controller.CountActive(ctx, &alertsdto.CountInput{Page: 3, Size: 7})
		},
		func(t *testing.T, service *stubAlertsService) {
			if service.activeInput.Page != 3 || service.activeInput.Size != 7 {
				t.Fatalf("active input = %#v, want forwarded page and size", service.activeInput)
			}
		},
	)
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestAlertsControllerRejectsMissingAuthenticatedPrincipal(t *testing.T) {
	t.Parallel()

	controller := NewSystemAlertsController(&stubAlertsService{})

	output, err := controller.Mute(context.Background(), &alertsdto.ActionInput{ID: "evt-1"})
	if output != nil {
		t.Fatalf("Mute() output = %#v, want nil", output)
	}
	if err == nil || humaStatus(err) != 401 {
		t.Fatalf("Mute() error = %v, want status 401", err)
	}
}
