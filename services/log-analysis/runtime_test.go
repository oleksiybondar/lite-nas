package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lite-nas/services/log-analysis/plugins"
	sharedlogger "lite-nas/shared/logger"
)

// Requirements: task.md/SVC-001.
func TestSampleRulesLoad(t *testing.T) {
	t.Parallel()
	path := "../../configs/etc/lite-nas/log-analysis/rules.json"
	if _, err := loadRulesFile(path); err != nil {
		t.Fatalf("sample rules: %v", err)
	}
}

// Requirements: task.md/SVC-001.
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

// Requirements: task.md/SVC-001.
func TestParserExtractsPAMAccount(t *testing.T) {
	t.Parallel()
	line := "2026-10-25T08:26:06.830681+01:00 host auth-service: pam_unix(litenas-auth:auth): authentication failure; rhost=192.0.2.4 user=alice"
	record, err := parseAuthRecord("auth", line)
	if err != nil {
		t.Fatalf("parse auth: %v", err)
	}
	if record.Account != "alice" {
		t.Fatalf("account = %q", record.Account)
	}
}

// Requirements: task.md/SVC-001.
func TestParserRejectsMalformedTimestamp(t *testing.T) {
	t.Parallel()
	if _, err := parseAuthRecord("auth", "bad host auth-service: failure"); err == nil {
		t.Fatal("invalid timestamp accepted")
	}
}

// Requirements: task.md/SVC-001.
func TestTailerBuffersPartialLine(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "auth.log")
	if err := os.WriteFile(path, []byte("first"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	tailer, err := newTailer(path, false)
	if err != nil {
		t.Fatalf("tailer: %v", err)
	}
	defer tailer.close()
	var lines []string
	if err := tailer.readLines(func(line string) error { lines = append(lines, line); return nil }); err != nil {
		t.Fatalf("read: %v", err)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if _, err := file.WriteString(" line\n"); err != nil {
		t.Fatalf("write append: %v", err)
	}
	file.Close()
	if err := tailer.readLines(func(line string) error { lines = append(lines, line); return nil }); err != nil {
		t.Fatalf("read append: %v", err)
	}
	if len(lines) != 1 || lines[0] != "first line" {
		t.Fatalf("lines = %v", lines)
	}
}

// TestTailerReadsFirstCreatedLog verifies that a source created after startup
// is read from its beginning rather than silently made the new baseline.
// Requirements: task.md/SVC-001.
func TestTailerReadsFirstCreatedLog(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "auth.log")
	tailer, err := newTailer(path, true)
	if err != nil {
		t.Fatalf("tailer: %v", err)
	}
	defer tailer.close()
	if err := os.WriteFile(path, []byte("first line\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	var lines []string
	if err := tailer.readLines(func(line string) error { lines = append(lines, line); return nil }); err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(lines) != 1 || lines[0] != "first line" {
		t.Fatalf("lines = %v", lines)
	}
}

// Requirements: task.md/SVC-001.
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
	dispatcher := newDispatchWorker(input, routes, instances, sharedlogger.NewNop())
	dispatcher.run(nil)
	if len(instances["account"].Input()) != 1 || len(instances["source"].Input()) != 1 {
		t.Fatal("fanout missed a plugin instance")
	}
}
