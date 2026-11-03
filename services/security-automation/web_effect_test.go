package main

import (
	"context"
	"testing"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/logger"
)

// TestVerifiedWebAlertQueuesDurableRevocation checks the exact account and
// source are persisted before the Auth request worker can run.
// Requirements: security-automation/FR-007.
func TestVerifiedWebAlertQueuesDurableRevocation(t *testing.T) {
	store := mustIncidentStore(t)
	runtime := &automationRuntime{incidents: store, operatorActions: make(chan effectCommand, 1)}
	intent := detectionIntent{RuleID: "web_password_guessing", Match: detection.Match{Key: "secweb123", AlertIDs: []string{"secweb123"}, LastMessage: "Security rule triggered: pam_web_api_password_guessing via Web/API for account alice from 198.51.100.10"}}
	if err := runtime.queueWebRevocation(intent); err != nil {
		t.Fatal(err)
	}
	if got := store.list(); len(got) != 1 || got[0].Username != "alice" || got[0].SourceIP != "198.51.100.10" {
		t.Fatalf("journaled web target = %+v", got)
	}
}

// TestVerifiedWebAlertRejectsMalformedSource keeps an invalid alert from
// reaching the Auth effect queue.
// Requirements: security-automation/FR-007.
func TestVerifiedWebAlertRejectsMalformedSource(t *testing.T) {
	runtime := &automationRuntime{incidents: mustIncidentStore(t), operatorActions: make(chan effectCommand, 1)}
	intent := detectionIntent{RuleID: "web_password_guessing", Match: detection.Match{Key: "secweb123", AlertIDs: []string{"secweb123"}, LastMessage: "Security rule triggered: pam_web_api_password_guessing via Web/API for account alice from --all"}}
	if err := runtime.queueWebRevocation(intent); err == nil {
		t.Fatal("invalid source queued")
	}
}

// TestWebRevocationCompletesAfterAudit verifies one-way effects leave no
// restoreable active incident after their security alert is accepted.
// Requirements: security-automation/FR-007.
func TestWebRevocationCompletesAfterAudit(t *testing.T) {
	store := mustIncidentStore(t)
	runtime := &automationRuntime{
		incidents: store, operatorActions: make(chan effectCommand, 1), log: logger.NopLogger{},
		runWebRevocation: func(context.Context, string, string) (int, error) { return 1, nil },
		audit:            func(context.Context, incidentRecord, bool) error { return nil },
	}
	intent := detectionIntent{RuleID: "web_password_guessing", Match: detection.Match{Key: "secweb123", AlertIDs: []string{"secweb123"}, LastMessage: "Security rule triggered: pam_web_api_password_guessing via Web/API for account alice from 198.51.100.10"}}
	if err := runtime.queueWebRevocation(intent); err != nil {
		t.Fatal(err)
	}
	command := <-runtime.operatorActions
	runtime.applyOneWayIncident(context.Background(), command.id)
	if got := store.list(); len(got) != 0 {
		t.Fatalf("completed web incident retained = %+v", got)
	}
}

// TestDirectWebAlertQueuesScopedUFW checks automatic blocking requires the
// matched source from the dedicated direct HTTPS alert.
// Requirements: security-automation/FR-007.
func TestDirectWebAlertQueuesScopedUFW(t *testing.T) {
	store := mustIncidentStore(t)
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	runtime := &automationRuntime{
		incidents: store, operatorActions: make(chan effectCommand, 1), log: logger.NopLogger{},
		rules: rules, physicalInterfaces: func() ([]string, error) { return []string{"physical0"}, nil },
	}
	intent := detectionIntent{RuleID: "web_https_source_block", Match: detection.Match{Key: "198.51.100.10", AlertIDs: []string{"secweb123"}, LastMessage: "Security rule triggered: pam_web_api_password_guessing via Web/API for account alice from 198.51.100.10"}}
	if err := runtime.queueAutomaticUFW(intent); err != nil {
		t.Fatal(err)
	}
	if got := store.list(); len(got) != 1 || got[0].SourceIP != "198.51.100.10" || len(got[0].Interfaces) != 1 {
		t.Fatalf("automatic direct HTTPS incident = %+v", got)
	}
}
