package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	alertcontract "lite-nas/shared/contracts/loggingmanager"
	contract "lite-nas/shared/contracts/securityautomation"
)

const repositoryRules = "../../configs/etc/lite-nas/security-automation/rules.json"

// mustRuleCopy writes one mutated copy of packaged rule definitions.
func mustRuleCopy(t *testing.T, old, replacement string) string {
	t.Helper()
	data, err := os.ReadFile(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "rules.json")
	if err := os.WriteFile(path, []byte(strings.Replace(string(data), old, replacement, 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// Requirements: security-automation/FR-003.
// TestPackagedRuleEnablement checks startup disables the uncorrelatable integrity rule.
func TestPackagedRuleEnablement(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]bool, len(rules))
	for id, rule := range rules {
		got[id] = *rule.spec.Enabled
	}
	want := map[string]bool{
		"account_enumeration": true, "account_under_attack": true,
		"file_access_denial": true, "integrity_service_disruption": false,
		"network_reconnaissance": true, "network_scan_sequence": true, "privilege_persistence": true,
		"recon_then_ssh": true, "repeated_sudo_denial": true,
		"service_recovery_failure": true, "ssh_brute_force": true,
		"ssh_privilege_probing": true, "usb_suspicious_activity": true,
		"ssh_probe_then_root_login": true, "sudo_denial_then_group_addition": true,
		"web_password_guessing":  true,
		"web_https_source_block": true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("packaged rule enablement = %+v, want %+v", got, want)
	}
}

// Requirements: security-automation/FR-003.
// TestEnabledRuleCreatesConfiguredFSM checks startup constructs the selected compiled plugin.
func TestEnabledRuleCreatesConfiguredFSM(t *testing.T) {
	path := repositoryRules
	rules, err := loadAutomationRules(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
		t.Fatal(err)
	}
	if rules["ssh_brute_force"].plugin == nil || !rules["ssh_brute_force"].enabled.Load() {
		t.Fatal("enabled rule has no initialized detector")
	}
}

// Requirements: security-automation/FR-003.
// TestRuleLoaderRejectsMalformedDefinitions checks startup rejects unsafe or ambiguous rule files.
func TestRuleLoaderRejectsMalformedDefinitions(t *testing.T) {
	cases := []struct{ name, old, replacement string }{
		{"unknown field", `"id": "ssh_brute_force"`, `"id": "ssh_brute_force", "unknown": true`},
		{"missing enabled", `, "enabled": true`, ``},
		{"duplicate ID", `"id": "account_enumeration"`, `"id": "ssh_brute_force"`},
		{"invalid regex", `"regex": "from (?P<requester_ip_address>[0-9a-fA-F:.]+)"`, `"regex": "["`},
		{"missing capture", `"regex": "from (?P<requester_ip_address>[0-9a-fA-F:.]+)"`, `"regex": "from ([0-9.]+)"`},
		{"invalid action", `"kind": "builtin"`, `"kind": "shell"`},
		{"unknown profile", `"name": "block_source_ip"`, `"name": "arbitrary_command"`},
		{"wrong operand", `"required_operands": ["requester_ip_address"]`, `"required_operands": ["unit_name"]`},
		{"invalid FSM", `"type": "count_within_window"`, `"type": "unknown"`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := mustRuleCopy(t, test.old, test.replacement)
			if _, err := loadAutomationRules(path); err == nil {
				t.Fatal("invalid automation rules accepted")
			}
		})
	}
}

// Requirements: security-automation/FR-002.
// TestRuleMatcherUsesValidatedCapture checks one alert yields only its named correlation key.
func TestRuleMatcherUsesValidatedCapture(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	key, matched := rules["ssh_brute_force"].matches(0, alertObservation{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "secssh_1", Source: "log-analysis", Message: "Security rule triggered: ssh_preauth_probing from 192.0.2.1"}})
	if !matched || key != "192.0.2.1" {
		t.Fatalf("match = %t, key = %q", matched, key)
	}
}

// Requirements: security-automation/FR-002.
// TestRuleMatcherUsesAlertFieldWithContains checks non-regex selectors can correlate on a validated field.
func TestRuleMatcherUsesAlertFieldWithContains(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	rule := rules["ssh_brute_force"]
	rule.spec.CorrelateBy = "source"
	rule.spec.Steps[0].Message.Regex = ""
	rule.spec.Steps[0].Message.Contains = "probe"
	rule.regexes[0] = nil
	key, matched := rule.matches(0, alertObservation{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "secssh_1", Source: "log-analysis", Message: "failed probe"}})
	if !matched || key != "log-analysis" {
		t.Fatalf("field correlation = %t, %q", matched, key)
	}
}

// Requirements: security-automation/FR-002.
// TestRuleMatcherRejectsOtherAlert checks unrelated alert types and IDs leave the FSM unchanged.
func TestRuleMatcherRejectsOtherAlert(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	cases := []alertObservation{
		{Kind: "system", Alert: alertcontract.AlertPayload{EventID: "secssh_1", Source: "log-analysis", Message: "Security rule triggered: ssh_preauth_probing from 192.0.2.1"}},
		{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "other_1", Source: "log-analysis", Message: "Security rule triggered: ssh_preauth_probing from 192.0.2.1"}},
		{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "secssh_2", Source: "log-analysis", Message: "Security rule triggered: ssh_root_login from 192.0.2.1"}},
		{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "secssh_1", Message: "no address"}},
	}
	for _, input := range cases {
		if _, matched := rules["ssh_brute_force"].matches(0, input); matched {
			t.Fatalf("unrelated alert matched: %+v", input)
		}
	}
}

// Requirements: security-automation/FR-002.
// TestAccountAttackStepsRequireDistinctAlertKinds checks that the packaged
// sequence cannot treat two password alerts as a password-and-lockout pair.
func TestAccountAttackStepsRequireDistinctAlertKinds(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		step    int
		alert   alertcontract.AlertPayload
		matched bool
	}{
		{"password first", 0, alertcontract.AlertPayload{EventID: "secauth_1", Category: "password_guessing", Source: "log-analysis", Message: "Security rule triggered: pam_password_guessing via Web/API for account alice"}, true},
		{"password cannot finish", 1, alertcontract.AlertPayload{EventID: "secauth_2", Category: "password_guessing", Source: "log-analysis", Message: "Security rule triggered: pam_password_guessing via Web/API for account alice"}, false},
		{"lockout finishes", 1, alertcontract.AlertPayload{EventID: "secauth_3", Category: "account_lockout", Source: "log-analysis", Message: "Security rule triggered: pam_account_lockout for account alice"}, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, matched := rules["account_under_attack"].matches(test.step, alertObservation{Kind: "security", Alert: test.alert})
			if matched != test.matched {
				t.Fatalf("step %d match = %t, want %t", test.step, matched, test.matched)
			}
		})
	}
}

// TestCorrelationAlertSelectors verifies each new sequence accepts only its
// intended Log Analysis rule while preserving the shared correlation key.
// Requirements: security-automation/FR-002.
func TestCorrelationAlertSelectors(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, ruleID, key string
		step              int
		alert             alertcontract.AlertPayload
	}{
		{"SSH probing", "ssh_probe_then_root_login", "198.51.100.10", 0, alertcontract.AlertPayload{EventID: "secssh_1", Source: "log-analysis", Category: "password_guessing", Message: "Security rule triggered: ssh_preauth_probing from 198.51.100.10"}},
		{"SSH root login", "ssh_probe_then_root_login", "198.51.100.10", 1, alertcontract.AlertPayload{EventID: "secssh_2", Source: "log-analysis", Category: "privilege_escalation", Message: "Security rule triggered: ssh_root_login from 198.51.100.10"}},
		{"sudo denial", "sudo_denial_then_group_addition", "alice", 0, alertcontract.AlertPayload{EventID: "secpriv_1", Source: "log-analysis", Category: "privilege_escalation", Message: "Security rule triggered: repeated_sudo_policy_violation for account alice"}},
		{"group addition", "sudo_denial_then_group_addition", "alice", 1, alertcontract.AlertPayload{EventID: "secpriv_2", Source: "log-analysis", Category: "privilege_escalation", Message: "Security rule triggered: privileged_group_membership for account alice"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			key, matched := rules[test.ruleID].matches(test.step, alertObservation{Kind: "security", Alert: test.alert})
			if !matched || key != test.key {
				t.Fatalf("correlation match = %t, %q", matched, key)
			}
		})
	}
}

// TestNetworkScanSequenceSelectors accepts the published horizontal and
// vertical scan summaries for the same validated source address.
// Requirements: security-automation/FR-002.
func TestNetworkScanSequenceSelectors(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, category, message string
		step                    int
	}{
		{"horizontal", "horizontal_scan", "Blocked TCP horizontal scan from 198.51.100.10 on port 443: 12 distinct destinations in 2m0s", 0},
		{"vertical", "port_scan", "Blocked TCP port scan from 198.51.100.10 to 192.0.2.10: 12 distinct ports in 2m0s", 1},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			key, matched := rules["network_scan_sequence"].matches(test.step, alertObservation{Kind: "security", Alert: alertcontract.AlertPayload{EventID: "secnet_1", Source: "network-security", Category: test.category, Message: test.message}})
			if !matched || key != "198.51.100.10" {
				t.Fatalf("scan match = %t, %q", matched, key)
			}
		})
	}
}

// TestNetworkScanSequenceRejectsRepeatedAttempts keeps a connection-pressure
// alert from satisfying the vertical port-scan step.
// Requirements: security-automation/FR-002.
func TestNetworkScanSequenceRejectsRepeatedAttempts(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	_, matched := rules["network_scan_sequence"].matches(1, alertObservation{Kind: "security", Alert: alertcontract.AlertPayload{
		EventID: "secnet_1", Source: "network-security", Category: "repeated_attempts",
		Message: "Repeated blocked TCP attempts from 198.51.100.10 to 192.0.2.10 port 443: 12 attempts in 2m0s",
	}})
	if matched {
		t.Fatal("repeated attempts matched vertical scan")
	}
}

// Requirements: security-automation/FR-003.
// TestRuleViewStartsNormal checks operator state describes an unmitigated definition.
func TestRuleViewStartsNormal(t *testing.T) {
	rules, err := loadAutomationRules(repositoryRules)
	if err != nil {
		t.Fatal(err)
	}
	if rules["ssh_brute_force"].view.State != contract.StateNormal || !rules["ssh_brute_force"].view.ConfiguredEnabled {
		t.Fatalf("rule view = %+v", rules["ssh_brute_force"].view)
	}
}
