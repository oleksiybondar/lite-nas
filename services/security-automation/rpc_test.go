package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/authtoken"
	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/messagingtest"
)

// mustOperatorRuntime constructs a live-token-compatible service fixture.
func mustOperatorRuntime(t *testing.T, roles []string) (*automationRuntime, string) {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := authtoken.NewIssuer(authtoken.IssuerOptions{Issuer: "test", Audience: "test", AccessLifetime: time.Minute}, private)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: "test", Audience: "test"}, public)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := issuer.Issue(authtoken.Principal{Subject: "123", Login: "alice", Roles: roles})
	if err != nil {
		t.Fatal(err)
	}
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	return &automationRuntime{rules: rules, order: []string{"ssh_brute_force"}, verifier: verifier, client: &messagingtest.RecordingClient{AuthUsername: "alice"}, log: logger.NopLogger{}}, token
}

// mustRPCEnvelope encodes an operator DTO for a handler fixture.
func mustRPCEnvelope(t *testing.T, request any) messaging.Envelope {
	t.Helper()
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return messaging.Envelope{Payload: payload}
}

// Requirements: security-automation/IR-003.
// TestRegisterAutomationRPCs checks every gateway and CLI subject has a service handler.
func TestRegisterAutomationRPCs(t *testing.T) {
	server := &messagingtest.RecordingServer{}
	if err := new(automationRuntime).register(server); err != nil {
		t.Fatal(err)
	}
	for _, subject := range []string{contract.RulesListSubject, contract.ActiveListSubject, contract.RollbackSubject, contract.ForceNormalizeSubject, contract.ForceMitigateSubject, contract.CooldownResetSubject, contract.RuleSetEnabledSubject} {
		if server.Handlers[subject] == nil {
			t.Fatalf("missing RPC %q", subject)
		}
	}
}

// Requirements: security-automation/IR-003.
// TestRegisterAutomationRPCFailure checks startup stops when a subject cannot be registered.
func TestRegisterAutomationRPCFailure(t *testing.T) {
	server := &messagingtest.RecordingServer{RegisterErr: errors.New("registration failed")}
	if err := new(automationRuntime).register(server); err == nil {
		t.Fatal("registration failure ignored")
	}
}

// Requirements: security-automation/SR-001.
// TestAutomationRPCRejectsMalformedPayload checks input validation precedes authorization and business logic.
func TestAutomationRPCRejectsMalformedPayload(t *testing.T) {
	runtime, _ := mustOperatorRuntime(t, []string{"lite-nas-security"})
	cases := []string{"", "{}", `{"access_token":"x","extra":1}`, `{"access_token":"x"} {}`}
	for _, payload := range cases {
		_, err := runtime.listRules(context.Background(), messaging.Envelope{Payload: []byte(payload)})
		if err == nil {
			t.Fatalf("malformed payload %q accepted", payload)
		}
	}
}

// Requirements: security-automation/SR-001.
// TestAutomationRPCRequiresSecurityRole checks a valid operator token cannot mutate security rules.
func TestAutomationRPCRequiresSecurityRole(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-operator"})
	_, err := runtime.setEnabled(context.Background(), mustRPCEnvelope(t, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: "ssh_brute_force", Enabled: true}))
	if err == nil {
		t.Fatal("operator role changed security rule")
	}
}

// Requirements: security-automation/IR-003.
// TestAutomationRPCListsConfiguredRules checks enabled definitions remain visible.
func TestAutomationRPCListsConfiguredRules(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	response, err := runtime.listRules(context.Background(), mustRPCEnvelope(t, contract.ListRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if rules := response.(contract.RulesListResponse).Rules; len(rules) != 1 || rules[0].ID != "ssh_brute_force" || !rules[0].Enabled {
		t.Fatalf("listed rules = %+v", rules)
	}
}

// Requirements: security-automation/IR-003.
// TestAutomationRPCListsNoUnverifiedEffects checks the service does not invent active incidents.
func TestAutomationRPCListsNoUnverifiedEffects(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	response, err := runtime.listActive(context.Background(), mustRPCEnvelope(t, contract.ListRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.(contract.ActiveListResponse).Items) != 0 {
		t.Fatalf("unverified active incidents = %+v", response)
	}
}

// Requirements: security-automation/SR-002.
// TestAutomationRPCRejectsUnenrolledEffects checks operators cannot claim an unverified mitigation or restore.
func TestAutomationRPCRejectsUnenrolledEffects(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	cases := []struct {
		name    string
		handler messaging.RPCHandler
		request any
	}{
		{"rollback", runtime.rollback, contract.RollbackRequest{AccessToken: token, IncidentID: "incident1"}},
		{"normalize", runtime.forceNormalize, contract.RollbackRequest{AccessToken: token, IncidentID: "incident1"}},
		{"mitigate", runtime.forceMitigate, contract.ForceMitigateRequest{AccessToken: token, RuleID: "ssh_brute_force", Operands: map[string]string{"requester_ip_address": "192.0.2.1"}}},
		{"cooldown", runtime.resetCooldown, contract.CooldownResetRequest{AccessToken: token, RuleID: "ssh_brute_force", CorrelationKey: "alice"}},
		{"enable", runtime.setEnabled, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: "ssh_brute_force", Enabled: true}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := test.handler(context.Background(), mustRPCEnvelope(t, test.request)); err == nil {
				t.Fatal("unenrolled action accepted")
			}
		})
	}
}

// TestForceMitigatePersistsBeforeQueue verifies an operator request is durable
// before its worker can apply a direct HTTPS block.
// Requirements: security-automation/SR-002.
func TestForceMitigatePersistsBeforeQueue(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	runtime.incidents = mustIncidentStore(t)
	runtime.operatorActions = make(chan effectCommand, 1)
	runtime.physicalInterfaces = func() ([]string, error) { return []string{"physical0"}, nil }
	runtime.runUFW = func(context.Context, string, string, string, string, string) error { return nil }
	runtime.audit = func(context.Context, incidentRecord, bool) error { return nil }
	request := contract.ForceMitigateRequest{AccessToken: token, RuleID: "ssh_brute_force", Operands: map[string]string{"requester_ip_address": "198.51.100.10"}}
	response, err := runtime.forceMitigate(context.Background(), mustRPCEnvelope(t, request))
	if err != nil {
		t.Fatal(err)
	}
	if result := response.(contract.OperationResponse); result.Completed || len(runtime.incidents.list()) != 1 {
		t.Fatalf("queued mitigation = %+v, incidents = %+v", result, runtime.incidents.list())
	}
}

// Requirements: security-automation/IR-003.
// TestAutomationRPCDisablesRuleTemporarily checks a configured instance can be stopped in memory.
func TestAutomationRPCDisablesRuleTemporarily(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	response, err := runtime.setEnabled(context.Background(), mustRPCEnvelope(t, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: "ssh_brute_force", Enabled: false}))
	if err != nil {
		t.Fatal(err)
	}
	if !response.(contract.OperationResponse).Completed || runtime.rules["ssh_brute_force"].enabled.Load() {
		t.Fatalf("disable response = %+v", response)
	}
}

// Requirements: security-automation/FR-003.
// TestDisableLeavesOneWayIncidentOpen checks a rule can stop taking new
// observations without attempting to reverse a completed Auth action.
func TestDisableLeavesOneWayIncidentOpen(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	runtime.incidents = mustIncidentStore(t)
	runtime.operatorActions = make(chan effectCommand, 1)
	now := time.Now().UTC()
	record := incidentRecord{
		ID: "A1B2C3D4E5F6G7H8", RuleID: "web_password_guessing", ActionName: "revoke_web_sessions",
		Username: "alice", SourceIP: "198.51.100.10", StartedAt: now, ExpiresAt: now.Add(5 * time.Minute),
		State: contract.StatePending, Outcome: "revocation pending",
	}
	if err := runtime.incidents.save(record); err != nil {
		t.Fatal(err)
	}
	request := contract.RuleSetEnabledRequest{AccessToken: token, RuleID: "web_password_guessing", Enabled: false}
	response, err := runtime.setEnabled(context.Background(), mustRPCEnvelope(t, request))
	if err != nil {
		t.Fatal(err)
	}
	if !response.(contract.OperationResponse).Completed {
		t.Fatal("one-way incident prevented rule disablement")
	}
}

// Requirements: security-automation/FR-003.
// TestTemporaryRuleReenableStartsFreshFSM checks an RPC override clears incomplete detection state.
func TestTemporaryRuleReenableStartsFreshFSM(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	path := mustRuleCopy(t, `"id": "account_enumeration", "name": "Account enumeration", "description": "Distinct account probes from one source address.", "enabled": false`, `"id": "account_enumeration", "name": "Account enumeration", "description": "Distinct account probes from one source address.", "enabled": true`)
	rules, err := loadAutomationRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
		t.Fatal(err)
	}
	runtime.rules = rules
	plugin := rules["account_enumeration"].plugin
	now := time.Now()
	for count := 0; count < 4; count++ {
		if _, err := plugin.Observe(context.Background(), detection.Event{ID: fmt.Sprintf("alert%d", count), Key: "alice", At: now}); err != nil {
			t.Fatal(err)
		}
	}
	for _, enabled := range []bool{false, true} {
		if _, err := runtime.setEnabled(context.Background(), mustRPCEnvelope(t, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: "account_enumeration", Enabled: enabled})); err != nil {
			t.Fatal(err)
		}
	}
	matches, err := plugin.Observe(context.Background(), detection.Event{ID: "fifth", Key: "alice", At: now.Add(time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("re-enabled detector completed with %d matches", len(matches))
	}
}

// Requirements: security-automation/IR-003.
// TestCooldownResetClearsOnlyDetectionState checks the authenticated command
// resets a configured FSM key without needing an incident effect store.
func TestCooldownResetClearsOnlyDetectionState(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	if err := constructAutomationInstances(runtime.rules, time.Now); err != nil {
		t.Fatal(err)
	}
	plugin := runtime.rules["account_enumeration"].plugin
	now := time.Now()
	for count := 0; count < 4; count++ {
		if _, err := plugin.Observe(context.Background(), detection.Event{ID: fmt.Sprintf("alert%d", count), Key: "alice", At: now}); err != nil {
			t.Fatal(err)
		}
	}
	request := contract.CooldownResetRequest{AccessToken: token, RuleID: "account_enumeration", CorrelationKey: "alice"}
	response, err := runtime.resetCooldown(context.Background(), mustRPCEnvelope(t, request))
	if err != nil {
		t.Fatal(err)
	}
	if !response.(contract.OperationResponse).Completed {
		t.Fatal("detection reset was not acknowledged")
	}
}

// Requirements: security-automation/FR-003.
// TestConfiguredEnablementRestoresOnReload checks temporary state never changes rules.json.
func TestConfiguredEnablementRestoresOnReload(t *testing.T) {
	runtime, token := mustOperatorRuntime(t, []string{"lite-nas-security"})
	path := mustRuleCopy(t, `"id": "ssh_brute_force", "name": "SSH brute force", "description": "Aggregated SSH failures from one source address.", "enabled": false`, `"id": "ssh_brute_force", "name": "SSH brute force", "description": "Aggregated SSH failures from one source address.", "enabled": true`)
	rules, err := loadAutomationRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
		t.Fatal(err)
	}
	runtime.rules = rules
	if _, err := runtime.setEnabled(context.Background(), mustRPCEnvelope(t, contract.RuleSetEnabledRequest{AccessToken: token, RuleID: "ssh_brute_force", Enabled: false})); err != nil {
		t.Fatal(err)
	}
	reloaded, err := loadAutomationRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if !*reloaded["ssh_brute_force"].spec.Enabled {
		t.Fatal("configured enablement was changed by temporary RPC override")
	}
}
