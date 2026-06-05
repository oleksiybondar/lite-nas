package services

import (
	"context"
	"errors"
	"reflect"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
)

func TestSystemAlertsServiceRequestsListSubject(t *testing.T) {
	runAlertsServiceListTest(
		t,
		"List",
		func(service AlertsService) (AlertListPage, error) {
			return service.List(context.Background(), AlertListInput{AccessToken: "AT", Page: 2, Size: 5})
		},
		systemloggingmanagercontract.GetAlertsRPCSubject,
		2,
		5,
		nil,
		[]loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
		7,
	)
}

func TestSystemAlertsServiceRequestsActiveSubject(t *testing.T) {
	runAlertsServiceListTest(
		t,
		"ListActive",
		func(service AlertsService) (AlertListPage, error) {
			return service.ListActive(context.Background(), AlertListInput{AccessToken: "AT", Page: 1, Size: 10})
		},
		systemloggingmanagercontract.GetActiveAlertsRPCSubject,
		1,
		10,
		nil,
		[]loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
		3,
	)
}

func TestSystemAlertsServiceRequestsUnacknowledgedSubject(t *testing.T) {
	runAlertsServiceListTest(
		t,
		"ListUnacknowledged",
		func(service AlertsService) (AlertListPage, error) {
			return service.ListUnacknowledged(context.Background(), AlertListInput{AccessToken: "AT", Page: 1, Size: 10})
		},
		systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
		1,
		10,
		nil,
		[]loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
		4,
	)
}

func TestSystemAlertsServiceForwardsFilters(t *testing.T) {
	runAlertsServiceListTest(
		t,
		"ListWithFilters",
		func(service AlertsService) (AlertListPage, error) {
			return service.List(context.Background(), AlertListInput{
				AccessToken: "AT",
				Page:        2,
				Size:        5,
				Filters: []loggingmanagerdto.Filter{{
					Key:       loggingmanagerdto.FilterKeyCategory,
					Condition: loggingmanagerdto.FilterConditionEQ,
					Values:    []string{"system.metrics.mem.used"},
				}},
			})
		},
		systemloggingmanagercontract.GetAlertsRPCSubject,
		2,
		5,
		[]loggingmanagerdto.Filter{{
			Key:       loggingmanagerdto.FilterKeyCategory,
			Condition: loggingmanagerdto.FilterConditionEQ,
			Values:    []string{"system.metrics.mem.used"},
		}},
		nil,
		0,
	)
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestSystemAlertsServiceGetReturnsNotFoundWhenRPCItemMissing(t *testing.T) {
	t.Parallel()

	client := &alertsClientStub{}
	service := NewSystemAlertsService(client)

	got, found, err := service.Get(context.Background(), AlertGetInput{AccessToken: "AT", ID: "evt-1"})
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if found {
		t.Fatalf("found = true, want false")
	}
	if !reflect.DeepEqual(got, loggingmanagercontract.ListAlertItem{}) {
		t.Fatalf("Get() item = %#v, want zero value", got)
	}
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestSystemAlertsServiceAcknowledgeMapsNegativeBackendReply(t *testing.T) {
	t.Parallel()

	client := newAlertsActionClientStub(t, false)
	service := NewSystemAlertsService(client)

	err := service.Acknowledge(context.Background(), AlertActionInput{AccessToken: "AT", ID: "evt-1", ActorLogin: "john.doe"})
	if !errors.Is(err, ErrAlertActionFailed) {
		t.Fatalf("Acknowledge() error = %v, want ErrAlertActionFailed", err)
	}

	if client.subject != systemloggingmanagercontract.AcknowledgeAlertRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, systemloggingmanagercontract.AcknowledgeAlertRPCSubject)
	}

	assertAcknowledgeRequest(t, client.request, "AT", "evt-1", "john.doe")
}

// Requirements: web-gateway/FR-005, web-gateway/IR-002
func TestSecurityAlertsServiceRequestsActiveSubject(t *testing.T) {
	t.Parallel()

	client := newAlertsListClientStub(t, nil, 0)
	service := NewSecurityAlertsService(client)

	if _, err := service.ListActive(context.Background(), AlertListInput{AccessToken: "AT", Page: 1, Size: 10}); err != nil {
		t.Fatalf("ListActive() error = %v", err)
	}

	if client.subject != securityloggingmanagercontract.GetActiveAlertsRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, securityloggingmanagercontract.GetActiveAlertsRPCSubject)
	}
}

type alertsClientStub struct {
	subject     string
	request     any
	requestFunc func(context.Context, string, any, any) error
}

func (c *alertsClientStub) Publish(context.Context, string, any) error { return nil }

func (c *alertsClientStub) Request(ctx context.Context, subject string, request any, response any) error {
	c.subject = subject
	c.request = request
	if c.requestFunc == nil {
		return nil
	}
	return c.requestFunc(ctx, subject, request, response)
}

func (c *alertsClientStub) Drain() error { return nil }
func (c *alertsClientStub) Close()       {}
