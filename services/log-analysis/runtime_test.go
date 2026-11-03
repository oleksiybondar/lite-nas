package main

import (
	"context"
	"strings"
	"testing"

	"lite-nas/services/log-analysis/plugins"
	sharedlogger "lite-nas/shared/logger"
)

// Requirements: security-monitoring/FR-001.
func TestSampleRulesLoad(t *testing.T) {
	t.Parallel()
	path := "../../configs/etc/lite-nas/log-analysis/rules.json"
	if _, err := loadRulesFile(path); err != nil {
		t.Fatalf("sample rules: %v", err)
	}
}

// Requirements: security-monitoring/FR-010.
func TestFileAccessRuleMatchesDenial(t *testing.T) {
	t.Parallel()
	rule := mustSampleRule(t, "file_access_denied")
	if !rule.matches("<12>1 now nas file-service 1 FILE_SERVICE - access_denied user=alice") {
		t.Fatal("file access denial did not match")
	}
}

// Requirements: security-monitoring/FR-001.
func TestDetectionMessageIdentifiesAuthenticationChannel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		service string
		channel string
	}{
		{service: "litenas-auth", channel: "Web/API"},
		{service: "sshd", channel: "SSH/CLI"},
		{service: "su-l", channel: "local CLI"},
	}
	for _, testCase := range cases {
		t.Run(testCase.service, func(t *testing.T) {
			t.Parallel()
			message := detectionMessage("Security detection recorded", plugins.Detection{
				RuleID: "pam_password_guessing", Service: testCase.service,
			})
			if !strings.Contains(message, "via "+testCase.channel) {
				t.Fatalf("message %q does not identify channel %q", message, testCase.channel)
			}
		})
	}
}

// Requirements: security-monitoring/FR-006.
func TestDetectionMessageIdentifiesUSBDevice(t *testing.T) {
	t.Parallel()
	message := detectionMessage("Security rule triggered", plugins.Detection{
		RuleID: "usb_device_connected", Device: "3-5 203a:fff9",
	})
	if message != "Security rule triggered: usb_device_connected device 3-5 203a:fff9" {
		t.Fatalf("message = %q", message)
	}
}

// Requirements: security-monitoring/FR-006.
func TestSampleUSBRuleIgnoresRootHub(t *testing.T) {
	t.Parallel()
	rule := mustSampleRule(t, "usb_device_connected")
	if rule.matches("usb 1-0: New USB device found, idVendor=1d6b, idProduct=0002, bcdDevice= 6.08") {
		t.Fatal("root hub enumeration matched the USB device rule")
	}
}

// Requirements: security-monitoring/FR-006.
func TestSampleUSBRuleMatchesConnectedDevice(t *testing.T) {
	t.Parallel()
	rule := mustSampleRule(t, "usb_device_connected")
	if !rule.matches("usb 3-5: New USB device found, idVendor=203a, idProduct=fff9, bcdDevice= 1.00") {
		t.Fatal("connected USB device did not match the USB device rule")
	}
}

// mustSampleRule compiles one enabled rule from the packaged rules file.
func mustSampleRule(t *testing.T, id string) compiledRule {
	t.Helper()
	spec, err := loadRulesFile("../../configs/etc/lite-nas/log-analysis/rules.json")
	if err != nil {
		t.Fatalf("sample rules: %v", err)
	}
	for _, rule := range spec.Rules {
		if rule.ID == id {
			compiled, err := compileRulePatterns(rule)
			if err != nil {
				t.Fatalf("compile %s: %v", id, err)
			}
			return compiled
		}
	}
	t.Fatalf("sample rule %s not found", id)
	return compiledRule{}
}

// Requirements: security-monitoring/FR-001.
func TestFanoutReachesTwoPluginInstancesOnOneSource(t *testing.T) {
	t.Parallel()
	output := make(chan plugins.Detection, 2)
	instances, err := constructInstances([]instanceSpec{
		{ID: "account", Enabled: true, Type: "keyed_threshold", Threshold: 3, Window: "2m", Cooldown: "10m", Idle: "30m", KeyBy: "account", MaxKeys: 16},
		{ID: "source", Enabled: true, Type: "keyed_threshold", Threshold: 12, Window: "5m", Cooldown: "30m", Idle: "1h", KeyBy: "source", MaxKeys: 16},
	}, "SEC", output, sharedlogger.NewNop())
	if err != nil {
		t.Fatalf("construct instances: %v", err)
	}
	spec := rulesFile{Rules: []ruleSpec{
		{ID: "account_rule", Enabled: true, Source: "auth", Category: "account", Pattern: "authentication failure", Instance: "account", Severity: "warning"},
		{ID: "source_rule", Enabled: true, Source: "auth", Category: "probing", Pattern: "authentication failure", Instance: "source", Severity: "critical"},
	}}
	routes, err := buildRoutes(spec, instances)
	if err != nil {
		t.Fatalf("build routes: %v", err)
	}
	input := make(chan logRecord, 1)
	input <- logRecord{SourceID: "auth", Message: "authentication failure", Account: "alice", SourceAddress: "192.0.2.4"}
	close(input)
	dispatcher := newDispatchWorker(input, nil, routes, instances, sharedlogger.NewNop())
	dispatcher.run(nil)
	if len(instances["account"].Input()) != 1 || len(instances["source"].Input()) != 1 {
		t.Fatal("fanout missed a plugin instance")
	}
}

// Requirements: security-monitoring/FR-009, security-monitoring/IR-004.
func TestLockoutRuleCapturesAccountBeforeKeyedPlugin(t *testing.T) {
	t.Parallel()
	output := make(chan plugins.Detection, 1)
	instances, err := constructInstances([]instanceSpec{{
		ID: "lockout", Enabled: true, Type: "keyed_threshold", Threshold: 1,
		Window: "75m", Cooldown: "75m", Idle: "2h", KeyBy: "account", MaxKeys: 16,
	}}, "SEC", output, sharedlogger.NewNop())
	if err != nil {
		t.Fatalf("construct lockout instance: %v", err)
	}
	rule := mustSampleRule(t, "pam_account_lockout").spec
	rule.Instance = "lockout"
	routes, err := buildRoutes(rulesFile{Rules: []ruleSpec{rule}}, instances)
	if err != nil {
		t.Fatalf("build lockout route: %v", err)
	}
	input := make(chan logRecord, 1)
	input <- logRecord{SourceID: "auth", Message: "pam_faillock(sshd:auth): Consecutive login failures for user alice account temporarily locked"}
	close(input)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	instances["lockout"].Start(ctx)
	newDispatchWorker(input, nil, routes, instances, sharedlogger.NewNop()).run(nil)
	<-instances["lockout"].Done()
	effect := <-output
	if effect.Account != "alice" {
		t.Fatalf("lockout account = %q, want alice", effect.Account)
	}
}

// Requirements: security-monitoring/IR-004.
func TestConstructInstancesRejectsInvalidPluginDefinitions(t *testing.T) {
	t.Parallel()
	valid := instanceSpec{
		ID: "known", Enabled: true, Type: "threshold", Threshold: 2,
		Window: "1m", Cooldown: "1m", Idle: "2m", KeyBy: "global", MaxKeys: 16,
	}
	cases := []struct {
		name  string
		specs []instanceSpec
	}{
		{name: "unknown type", specs: []instanceSpec{{ID: "unknown", Enabled: true, Type: "absent"}}},
		{name: "bad threshold", specs: []instanceSpec{{ID: "bad", Enabled: true, Type: "threshold", Threshold: 0, Window: "1m", Cooldown: "1m", Idle: "2m", KeyBy: "global", MaxKeys: 16}}},
		{name: "duplicate ID", specs: []instanceSpec{valid, valid}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if _, err := constructInstances(testCase.specs, "SEC", make(chan plugins.Detection, 1), sharedlogger.NewNop()); err == nil {
				t.Fatal("invalid plugin instances accepted")
			}
		})
	}
}
