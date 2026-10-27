package main

import (
	"context"
	"errors"
	"testing"

	"lite-nas/services/log-analysis/plugins"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/messaging"
)

// TestImmediateCreateRetriesResolutionWithoutRepublishing verifies that a
// manager persistence delay cannot turn resolution retries into repeated
// active creates.
func TestImmediateCreateRetriesResolutionWithoutRepublishing(t *testing.T) {
	client := &deliveryTestClient{
		alertExists: []bool{false, true, true},
		stateOK:     []bool{false, true},
	}
	alertSent := false
	effect := pluginsDetectionForDeliveryTest()

	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err == nil {
		t.Fatal("resolution proceeded before the manager persisted the alert")
	}
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err == nil {
		t.Fatal("first state update unexpectedly succeeded")
	}
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err != nil {
		t.Fatalf("retry resolution: %v", err)
	}
	if client.publishCount != 1 || client.getAlertCount != 3 || client.stateCount != 2 {
		t.Fatalf("publish/get-alert/state counts = %d/%d/%d, want 1/3/2", client.publishCount, client.getAlertCount, client.stateCount)
	}
}

// pluginsDetectionForDeliveryTest returns the immediate effect used by the
// delivery retry test.
func pluginsDetectionForDeliveryTest() plugins.Detection {
	return plugins.Detection{
		Operation: "create", ResolveImmediately: true, EventID: "SEC123456_12345678",
		RuleID: "pam_auth_failure", Category: "auth_failure", Severity: "warning",
	}
}

// deliveryTestClient records publishes and provides scripted state RPC replies.
type deliveryTestClient struct {
	publishCount  int
	getAlertCount int
	stateCount    int
	alertExists   []bool
	stateOK       []bool
}

// Publish records one event publication for the retry assertion.
func (client *deliveryTestClient) Publish(_ context.Context, _ string, _ any) error {
	client.publishCount++
	return nil
}

// Request returns the next scripted manager response for an update request.
func (client *deliveryTestClient) Request(_ context.Context, subject string, _ any, response any) error {
	switch subject {
	case securitycontract.GetAlertRPCSubject:
		if client.getAlertCount >= len(client.alertExists) {
			return errors.New("unscripted get-alert request")
		}
		if client.alertExists[client.getAlertCount] {
			response.(*loggingcontract.GetAlertResponse).Item = &loggingcontract.ListAlertItem{}
		}
		client.getAlertCount++
	case securitycontract.UpdateAlertStateRPCSubject:
		if client.stateCount >= len(client.stateOK) {
			return errors.New("unscripted state request")
		}
		response.(*loggingcontract.OKResponse).OK = client.stateOK[client.stateCount]
		client.stateCount++
	default:
		return errors.New("unexpected request subject")
	}
	return nil
}

// Drain satisfies messaging.Client without owning transport resources.
func (*deliveryTestClient) Drain() error { return nil }

// Close satisfies messaging.Client without owning transport resources.
func (*deliveryTestClient) Close() {}

var _ messaging.Client = (*deliveryTestClient)(nil)
