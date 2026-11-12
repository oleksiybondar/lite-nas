package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/services/log-analysis/plugins"
	authcontract "lite-nas/shared/contracts/auth"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	networksecuritycontract "lite-nas/shared/contracts/networksecurity"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

// TestCreateDoesNotResolveSecurityAlert verifies that one-shot security alerts
// remain active for human acknowledgment after publication.
func TestCreateDoesNotResolveSecurityAlert(t *testing.T) {
	client := &deliveryTestClient{}
	alertSent := false
	effect := pluginsDetectionForDeliveryTest()

	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if client.publishCount != 1 || client.stateCount != 0 {
		t.Fatalf("publish/state counts = %d/%d, want 1/0", client.publishCount, client.stateCount)
	}
}

// TestCreateIncludesOperatorGuidance verifies rule reactions reach the alert payload.
func TestCreateIncludesOperatorGuidance(t *testing.T) {
	client := &deliveryTestClient{}
	effect := pluginsDetectionForDeliveryTest()
	effect.PossibleReactions = []string{"Review the source.", "Restrict it if hostile."}
	alertSent := false
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err != nil {
		t.Fatal(err)
	}
	if client.published.Mitigate != "Review the source.\nRestrict it if hostile." {
		t.Fatalf("mitigation guidance = %q", client.published.Mitigate)
	}
}

// Requirements: security-monitoring/FR-001.
func TestOccurrenceUsesExistingAlertSubject(t *testing.T) {
	client := &deliveryTestClient{}
	effect := pluginsDetectionForDeliveryTest()
	effect.Operation = "occurrence"
	alertSent := false
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err != nil {
		t.Fatalf("publish occurrence: %v", err)
	}
	if client.publishSubject != securitycontract.AlertOccurrenceSubject {
		t.Fatalf("subject = %q", client.publishSubject)
	}
}

// Requirements: security-monitoring/FR-001.
func TestCriticalCreateRequestsConfiguredLockdown(t *testing.T) {
	client := &deliveryTestClient{lockdown: true}
	effect := pluginsDetectionForDeliveryTest()
	effect.Severity = "critical"
	effect.Lockdown = true
	alertSent := false
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err != nil {
		t.Fatalf("deliver critical alert: %v", err)
	}
	if client.lockdownCount != 1 {
		t.Fatalf("lockdown requests = %d", client.lockdownCount)
	}
}

// Requirements: security-monitoring/FR-001.
func TestLockdownRejectsNoncriticalAlert(t *testing.T) {
	client := &deliveryTestClient{}
	effect := pluginsDetectionForDeliveryTest()
	effect.Lockdown = true
	alertSent := false
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err == nil {
		t.Fatal("noncritical lockdown accepted")
	}
}

// Requirements: security-monitoring/FR-001.
func TestLockdownRejectsUnconfirmedResponse(t *testing.T) {
	client := &deliveryTestClient{}
	effect := pluginsDetectionForDeliveryTest()
	effect.Severity = "critical"
	effect.Lockdown = true
	alertSent := false
	if err := deliverDetection(context.Background(), effect, &alertSent, client, client); err == nil {
		t.Fatal("unconfirmed lockdown accepted")
	}
}

// Requirements: security-monitoring/FR-001.
func TestReportDetectionsPublishesQueuedAlert(t *testing.T) {
	client := &deliveryTestClient{}
	effects := make(chan plugins.Detection, 1)
	effects <- pluginsDetectionForDeliveryTest()
	close(effects)
	reportDetections(context.Background(), effects, client, client, sharedlogger.NewNop())
	if client.publishCount != 1 {
		t.Fatalf("alert publications = %d", client.publishCount)
	}
}

// Requirements: security-monitoring/FR-001.
func TestReportDetectionsAccountsForFailureAtShutdown(t *testing.T) {
	client := &deliveryTestClient{publishErr: errors.New("offline")}
	effects := make(chan plugins.Detection, 1)
	effects <- pluginsDetectionForDeliveryTest()
	close(effects)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reportDetections(ctx, effects, client, client, sharedlogger.NewNop())
	if client.publishCount != 1 {
		t.Fatalf("alert attempts at shutdown = %d", client.publishCount)
	}
}

// Requirements: security-monitoring/FR-001.
func TestPublishFirewallAttemptsForwardsQueuedEvent(t *testing.T) {
	client := &deliveryTestClient{}
	attempts := make(chan networksecuritycontract.FirewallAttempt, 1)
	attempts <- networksecuritycontract.FirewallAttempt{SourceIP: "192.0.2.7", DestinationPort: 22, Timestamp: time.Now().UTC()}
	close(attempts)
	publishFirewallAttempts(context.Background(), attempts, client, sharedlogger.NewNop())
	if client.publishSubject != networksecuritycontract.FirewallAttemptSubject {
		t.Fatalf("firewall subject = %q", client.publishSubject)
	}
}

// Requirements: security-monitoring/FR-001.
func TestPublishFirewallAttemptsAccountsForFailureAtShutdown(t *testing.T) {
	client := &deliveryTestClient{publishErr: errors.New("offline")}
	attempts := make(chan networksecuritycontract.FirewallAttempt, 1)
	attempts <- networksecuritycontract.FirewallAttempt{SourceIP: "192.0.2.7", DestinationPort: 22, Timestamp: time.Now().UTC()}
	close(attempts)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	publishFirewallAttempts(ctx, attempts, client, sharedlogger.NewNop())
	if client.publishCount != 1 {
		t.Fatalf("firewall attempts at shutdown = %d", client.publishCount)
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
	published      loggingcontract.AlertPayload
	publishCount   int
	stateCount     int
	publishSubject string
	publishErr     error
	requestErr     error
	lockdown       bool
	lockdownCount  int
}

// Publish records one event publication for the retry assertion.
func (client *deliveryTestClient) Publish(_ context.Context, subject string, payload any) error {
	client.publishCount++
	client.publishSubject = subject
	if alert, ok := payload.(loggingcontract.AlertPayload); ok {
		client.published = alert
	}
	return client.publishErr
}

// Request returns the next scripted manager response for an update request.
func (client *deliveryTestClient) Request(_ context.Context, subject string, _ any, response any) error {
	if client.requestErr != nil {
		return client.requestErr
	}
	if subject == authcontract.SetLockdownRPCSubject {
		client.lockdownCount++
		response.(*authcontract.SetLockdownResponse).Lockdown = client.lockdown
		return nil
	}
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
