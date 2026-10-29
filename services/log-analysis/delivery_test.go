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

// TestCreateDoesNotResolveSecurityAlert verifies that one-shot security alerts
// remain active for human acknowledgment after publication.
func TestCreateDoesNotResolveSecurityAlert(t *testing.T) {
	client := &deliveryTestClient{}
	alertSent := false
	effect := pluginsDetectionForDeliveryTest()

	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err == nil {
		t.Fatal("create failed")
	}
	if client.publishCount != 1 || client.stateCount != 0 {
		t.Fatalf("publish/state counts = %d/%d, want 1/0", client.publishCount, client.stateCount)
	}
}

// pluginsDetectionForDeliveryTest returns the immediate effect used by the
// delivery retry test.
func pluginsDetectionForDeliveryTest() plugins.Detection {
	return plugins.Detection{
		Operation: "create", EventID: "SEC123456_12345678",
		RuleID: "pam_auth_failure", Category: "auth_failure", Severity: "warning",
	}
}

// deliveryTestClient records alert publications and any attempted state update.
type deliveryTestClient struct {
	publishCount int
	stateCount   int
}

// Publish records one event publication for the retry assertion.
func (client *deliveryTestClient) Publish(_ context.Context, _ string, _ any) error {
	client.publishCount++
	return nil
}

// Request returns the next scripted manager response for an update request.
func (client *deliveryTestClient) Request(_ context.Context, subject string, _ any, response any) error {
	if subject == securitycontract.UpdateAlertStateRPCSubject {
		client.stateCount++
		response.(*loggingcontract.OKResponse).OK = true
		return nil
	}
	return errors.New("unexpected request subject")
}

// Drain satisfies messaging.Client without owning transport resources.
func (*deliveryTestClient) Drain() error { return nil }

// Close satisfies messaging.Client without owning transport resources.
func (*deliveryTestClient) Close() {}

var _ messaging.Client = (*deliveryTestClient)(nil)
