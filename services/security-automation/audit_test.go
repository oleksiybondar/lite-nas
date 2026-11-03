package main

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"lite-nas/services/security-automation/detection"
	authcontract "lite-nas/shared/contracts/auth"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager"
)

// auditClientStub captures the existing Auth and Logging Manager wire shapes.
type auditClientStub struct {
	published  loggingcontract.AlertPayload
	normalized loggingcontract.UpdateAlertStateInput
}

// Publish captures one security alert after Auth token injection.
func (client *auditClientStub) Publish(_ context.Context, subject string, payload any) error {
	if subject == securitycontract.AlertSubject {
		client.published = payload.(loggingcontract.AlertPayload)
	}
	return nil
}

// Request issues a test service token and acknowledges alert normalization.
func (client *auditClientStub) Request(_ context.Context, subject string, request any, response any) error {
	switch subject {
	case authcontract.ServiceTokenLoginRPCSubject:
		*response.(*authcontract.ServiceTokenLoginResponse) = authcontract.ServiceTokenLoginResponse{AccessToken: "service-token", ExpiresAt: time.Now().Add(time.Hour)}
	case securitycontract.UpdateAlertStateRPCSubject:
		client.normalized = request.(loggingcontract.UpdateAlertStateInput)
		*response.(*loggingcontract.OKResponse) = loggingcontract.OKResponse{OK: true}
	}
	return nil
}

// Drain satisfies the messaging client lifecycle without a transport.
func (client *auditClientStub) Drain() error { return nil }

// Close satisfies the messaging client lifecycle without a transport.
func (client *auditClientStub) Close() {}

// TestAuditEventIDUsesLoggingManagerShape verifies stable incident IDs fit the
// established alert contract without introducing a new subject or DTO.
// Requirements: security-automation/FR-007.
func TestAuditEventIDUsesLoggingManagerShape(t *testing.T) {
	id, err := auditEventID("A1B2C3D4E5F6G7H8")
	if err != nil || !regexp.MustCompile(`^t[0-9]{19}$`).MatchString(id) {
		t.Fatalf("audit event ID = %q, %v", id, err)
	}
}

// TestCorrelationAlertPublishesContributingIDs checks one completed sequence
// becomes a security alert with a stable ID and the configured email policy.
// Requirements: security-automation/FR-007.
func TestCorrelationAlertPublishesContributingIDs(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	client := &auditClientStub{}
	publish, err := newCorrelationAlert(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	match := detection.Match{Key: "198.51.100.10", AlertIDs: []string{"secssh_1", "secssh_2"}}
	if err := publish(context.Background(), rules["ssh_probe_then_root_login"].spec, match); err != nil {
		t.Fatal(err)
	}
	wantID, err := correlationEventID("ssh_probe_then_root_login", match.AlertIDs)
	if err != nil {
		t.Fatal(err)
	}
	got := client.published
	if got.EventID != wantID || got.Category != "automation_correlation" || got.Severity != "critical" || got.Email == nil || !*got.Email || !strings.Contains(got.Message, "secssh_1, secssh_2") {
		t.Fatalf("correlation alert = %+v", got)
	}
}

// TestCorrelationAlertIncludesOperatorGuidance verifies alert rules publish their configured guidance.
func TestCorrelationAlertIncludesOperatorGuidance(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	client := &auditClientStub{}
	publish, err := newCorrelationAlert(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	rule := rules["ssh_probe_then_root_login"].spec
	if err := publish(context.Background(), rule, detection.Match{Key: "198.51.100.10", AlertIDs: []string{"secssh_1", "secssh_2"}}); err != nil {
		t.Fatal(err)
	}
	if client.published.Mitigate != rule.Action.Mitigate {
		t.Fatalf("mitigation guidance = %q", client.published.Mitigate)
	}
}

// TestCorrelationAlertUsesLoggingManagerContract checks the published report
// satisfies the transport validator, including its compact event ID policy.
// Requirements: security-automation/FR-007.
func TestCorrelationAlertUsesLoggingManagerContract(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	client := &auditClientStub{}
	publish, err := newCorrelationAlert(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	if err := publish(context.Background(), rules["ssh_probe_then_root_login"].spec, detection.Match{Key: "198.51.100.10", AlertIDs: []string{"secssh_1", "secssh_2"}}); err != nil {
		t.Fatal(err)
	}
	validate, err := loggingmanager.NewInputValidator()
	if err != nil {
		t.Fatal(err)
	}
	if err := validate.Struct(client.published); err != nil {
		t.Fatalf("invalid correlation alert payload: %v", err)
	}
}

// TestApplyIncidentJournalsAlertPublication checks the alert send is recorded
// so expiry and restart can distinguish completed from pending publication.
// Requirements: security-automation/FR-007.
func TestApplyIncidentJournalsAlertPublication(t *testing.T) {
	store := mustIncidentStore(t)
	record := mustPendingIncident(t, store)
	runtime := &automationRuntime{
		incidents: store, log: logger.NopLogger{},
		runUFW: func(context.Context, string, string, string, string, string) error { return nil },
		audit:  func(context.Context, incidentRecord, bool) error { return nil },
	}
	runtime.applyIncident(context.Background(), record.ID)
	got, _ := store.get(record.ID)
	if !got.AlertPublished {
		t.Fatal("successful alert publication was not journaled")
	}
}

// TestIncidentAuditPublishesAuthenticatedAlert verifies the existing service
// token is attached to the Security Logging Manager create contract.
// Requirements: security-automation/FR-007.
func TestIncidentAuditPublishesAuthenticatedAlert(t *testing.T) {
	client := &auditClientStub{}
	audit, err := newIncidentAudit(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	record := mustPendingIncident(t, mustIncidentStore(t))
	if err := audit(context.Background(), record, false); err != nil {
		t.Fatal(err)
	}
	if client.published.AccessToken != "service-token" || client.published.Source != "security-automation" {
		t.Fatalf("security alert payload = %+v", client.published)
	}
}

// TestIncidentAuditNormalizesWithServiceToken verifies restored incidents use
// the existing authenticated Logging Manager state-update RPC.
// Requirements: security-automation/FR-007.
func TestIncidentAuditNormalizesWithServiceToken(t *testing.T) {
	client := &auditClientStub{}
	audit, err := newIncidentAudit(client, "lite-nas-security-automation")
	if err != nil {
		t.Fatal(err)
	}
	record := mustPendingIncident(t, mustIncidentStore(t))
	if err := audit(context.Background(), record, true); err != nil {
		t.Fatal(err)
	}
	if client.normalized.AccessToken != "service-token" || client.normalized.EventID == "" {
		t.Fatalf("security alert normalization = %+v", client.normalized)
	}
}
