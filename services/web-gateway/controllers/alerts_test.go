package controllers

import (
	"context"
	"testing"

	alertsdto "lite-nas/services/web-gateway/dto/alerts"
	"lite-nas/services/web-gateway/services"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

type stubAlertsService struct {
	listPage            services.AlertListPage
	listErr             error
	item                loggingmanagercontract.ListAlertItem
	found               bool
	getErr              error
	actionErr           error
	listInput           services.AlertListInput
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

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestAlertsControllerListUnacknowledgedWrapsMetadataAndDefaultsPagination(t *testing.T) {
	t.Parallel()

	service := &stubAlertsService{listPage: alertListPageFixture()}
	controller := NewSystemAlertsController(service)
	ctx := authenticatedAlertsContext()

	output, err := controller.ListUnacknowledged(ctx, &alertsdto.ListInput{})
	if err != nil {
		t.Fatalf("ListUnacknowledged() error = %v", err)
	}

	assertSuccessfulAlertResponse(t, output.Body.Success, output.Body.Timestamp.IsZero())
	assertUnacknowledgedListInput(t, service.unacknowledgedInput)
	if output.Body.Data.Metadata.TotalPages != 2 {
		t.Fatalf("total_pages = %d, want 2", output.Body.Data.Metadata.TotalPages)
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

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestAlertsControllerCountUsesTotalCount(t *testing.T) {
	t.Parallel()

	service := &stubAlertsService{listPage: services.AlertListPage{TotalCount: 9}}
	controller := NewSystemAlertsController(service)
	ctx := authenticatedAlertsContext()

	output, err := controller.Count(ctx, &alertsdto.CountInput{Page: 3, Size: 7})
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if output.Body.Data.Count != 9 {
		t.Fatalf("count = %d, want 9", output.Body.Data.Count)
	}
	if service.listInput.Page != 3 || service.listInput.Size != 7 {
		t.Fatalf("list input = %#v, want forwarded page and size", service.listInput)
	}
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
