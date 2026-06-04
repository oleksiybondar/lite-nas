package services

import (
	"context"
	"errors"
	"reflect"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
)

// Requirements: web-gateway/FR-005, web-gateway/IR-002
func TestSystemAlertsServiceRequestsListSubject(t *testing.T) {
	t.Parallel()

	wantItems := []loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}}
	client := newAlertsListClientStub(t, wantItems, 7)
	service := NewSystemAlertsService(client)

	got, err := service.List(context.Background(), AlertListInput{AccessToken: "AT", Page: 2, Size: 5})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if client.subject != systemloggingmanagercontract.GetAlertsRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, systemloggingmanagercontract.GetAlertsRPCSubject)
	}

	assertAlertsListRequest(t, client.request, "AT", 2, 5)

	if !reflect.DeepEqual(got.Items, wantItems) || got.TotalCount != 7 {
		t.Fatalf("List() = %#v, want items=%#v total_count=7", got, wantItems)
	}
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
