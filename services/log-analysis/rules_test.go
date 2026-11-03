package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// sampleRuleCase pairs one packaged rule with a normalized log message and
// whether that rule must select it. Messages are real host records unless a
// comment states their origin.
type sampleRuleCase struct {
	name    string
	rule    string
	message string
	want    bool
}

// Requirements: security-monitoring/FR-007, security-monitoring/FR-009.
func TestSampleRulesSelectSecurityRecords(t *testing.T) {
	t.Parallel()
	cases := []sampleRuleCase{
		{
			name: "sudo password failure", rule: "pam_password_guessing", want: true,
			message: "pam_unix(sudo:auth): authentication failure; logname=alice uid=1001 euid=0 tty=/dev/pts/1 ruser=alice rhost=  user=alice",
		},
		{
			name: "web API password failure with source", rule: "pam_web_api_password_guessing", want: true,
			message: "pam_unix(litenas-auth:auth): authentication failure; logname= uid=0 rhost=198.51.100.10 user=alice",
		},
		{
			name: "web API password failure without source", rule: "pam_web_api_password_guessing", want: false,
			message: "pam_unix(litenas-auth:auth): authentication failure; logname= uid=0 rhost= user=alice",
		},
		{
			name: "sudo session does not reset guessing", rule: "pam_password_guessing", want: false,
			message: "pam_unix(sudo:session): session opened for user alice(uid=1001) by (uid=969)",
		},
		{
			name: "PAM lockout creates an alert", rule: "pam_account_lockout", want: true,
			message: "pam_faillock(sshd:auth): Consecutive login failures for user alice account temporarily locked",
		},
		{
			name: "PAM lockout contributes to the fifth-lockout alert", rule: "pam_repeated_account_lockout", want: true,
			message: "pam_faillock(litenas-auth:auth): Consecutive login failures for user alice account temporarily locked",
		},
		{
			name: "blocked retry is not a new lockout", rule: "pam_account_lockout", want: false,
			message: "pam_faillock(sshd:auth): User alice is temporarily locked out due to 5 consecutive failed login attempts",
		},
		// Classic sudo denial format; sudoers.so contains these reasons.
		{
			name: "user not in sudoers", rule: "sudo_policy_violation", want: true,
			message: "mallory : user NOT in sudoers ; TTY=pts/0 ; PWD=/home/mallory ; USER=root ; COMMAND=/usr/bin/id",
		},
		{
			name: "service command not allowed", rule: "sudo_policy_violation", want: true,
			message: "lite-nas-file-service : command not allowed ; PWD=/ ; USER=root ; COMMAND=/bin/sh",
		},
		{
			name: "allowed service sudo", rule: "sudo_policy_violation", want: false,
			message: "lite-nas-file-service :  PWD=/ ; USER=alice ; COMMAND=/usr/bin/cat -- /home/alice/.lite-nas/ui.json",
		},
		{
			name: "added to sudo group", rule: "privileged_group_membership", want: true,
			message: "add 'mallory' to group 'sudo'",
		},
		{
			name: "shadow group duplicate", rule: "privileged_group_membership", want: false,
			message: "add 'mallory' to shadow group 'sudo'",
		},
		// shadow-utils gpasswd format.
		{
			name: "gpasswd added to adm", rule: "privileged_group_membership", want: true,
			message: "user mallory added by root to group adm",
		},
		{
			name: "service group membership", rule: "privileged_group_membership", want: false,
			message: "add 'lite-nas-log-analysis' to group 'lite-nas-security'",
		},
		{
			name: "login account created", rule: "login_account_created", want: true,
			message: "new user: name=mallory, UID=1006, GID=1006, home=/home/mallory, shell=/bin/bash, from=/dev/pts/2",
		},
		{
			name: "service account created", rule: "login_account_created", want: false,
			message: "new user: name=lite-nas-log-analysis, UID=962, GID=944, home=/nonexistent, shell=/usr/sbin/nologin, from=/dev/pts/2",
		},
		{
			name: "root password changed", rule: "root_password_changed", want: true,
			message: "pam_unix(passwd:chauthtok): password changed for root",
		},
		{
			name: "other password changed", rule: "root_password_changed", want: false,
			message: "pam_unix(chpasswd:chauthtok): password changed for testsudouser",
		},
		{
			name: "ssh invalid user", rule: "ssh_preauth_probing", want: true,
			message: "Invalid user sec-ssh-7fecf0af4ace from 127.0.0.1 port 45776",
		},
		{
			name: "ssh failed password", rule: "ssh_preauth_probing", want: true,
			message: "Failed password for alice from 192.0.2.7 port 45776 ssh2",
		},
		{
			name: "ssh invalid user counted once", rule: "ssh_preauth_probing", want: false,
			message: "Failed password for invalid user sec-ssh-7fecf0af4ace from 127.0.0.1 port 45776 ssh2",
		},
		{
			name: "ssh preauth close counted once", rule: "ssh_preauth_probing", want: false,
			message: "Connection closed by invalid user sec-ssh-7fecf0af4ace 127.0.0.1 port 45776 [preauth]",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := mustSampleRule(t, testCase.rule).matches(testCase.message); got != testCase.want {
				t.Fatalf("%s matches(%q) = %v, want %v", testCase.rule, testCase.message, got, testCase.want)
			}
		})
	}
}

// Requirements: security-monitoring/FR-007, security-monitoring/FR-009.
func TestAuthFormatExtractsCorrelationAccount(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"mallory : user NOT in sudoers ; TTY=pts/0 ; PWD=/ ; USER=root ; COMMAND=/usr/bin/id": "mallory",
		"add 'mallory' to group 'sudo'":           "mallory",
		"user mallory added by root to group adm": "mallory",
		"new user: name=mallory, UID=1006, GID=1006, home=/home/mallory, shell=/bin/bash, from=/dev/pts/2":        "mallory",
		"pam_unix(passwd:chauthtok): password changed for root":                                                   "root",
		"pam_unix(su-l:session): session opened for user testsudouser(uid=1005) by oleksiybondar(uid=1001)":       "testsudouser",
		"pam_unix(sudo:auth): authentication failure; logname=alice uid=1001 euid=0 ruser=alice rhost=  user=bob": "bob",
	}
	for message, want := range cases {
		t.Run(want+"/"+message[:12], func(t *testing.T) {
			t.Parallel()
			record := logRecord{Message: message}
			extractAuthFields(&record)
			if record.Account != want {
				t.Fatalf("account = %q, want %q", record.Account, want)
			}
		})
	}
}

// Requirements: security-monitoring/FR-009.
func TestSampleLockoutRulesCaptureAccount(t *testing.T) {
	t.Parallel()
	message := "pam_faillock(sshd:auth): Consecutive login failures for user alice account temporarily locked"
	for _, id := range []string{"pam_account_lockout", "pam_repeated_account_lockout"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			match := mustSampleRule(t, id).account.FindStringSubmatch(message)
			if len(match) != 2 || match[1] != "alice" {
				t.Fatalf("account match = %q, want alice", match)
			}
		})
	}
}

// Requirements: security-monitoring/FR-009.
func TestAccountPatternRejectsMissingCapture(t *testing.T) {
	t.Parallel()
	_, err := compileRulePatterns(ruleSpec{Pattern: "locked", AccountPattern: "user \\S+"})
	if err == nil {
		t.Fatal("account pattern without a capture was accepted")
	}
}

// Requirements: security-monitoring/FR-001.
func TestRulesFileRejectsInvalidConfiguredBoundaries(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		mutate func(*rulesFile)
	}{
		{name: "duplicate rule", mutate: func(spec *rulesFile) { spec.Rules[1].ID = spec.Rules[0].ID }},
		{name: "duplicate source", mutate: func(spec *rulesFile) { spec.Sources[1].ID = spec.Sources[0].ID }},
		{name: "no sources", mutate: func(spec *rulesFile) { spec.Sources = nil }},
		{name: "duplicate instance", mutate: func(spec *rulesFile) { spec.Instances[1].ID = spec.Instances[0].ID }},
		{name: "bad instance event prefix", mutate: func(spec *rulesFile) { spec.Instances[0].EventIDPrefix = "bad-prefix" }},
		{name: "bad event prefix", mutate: func(spec *rulesFile) { spec.EventIDPrefix = "bad-prefix" }},
		{name: "missing source", mutate: func(spec *rulesFile) { spec.Rules[0].Source = "missing" }},
		{name: "missing instance", mutate: func(spec *rulesFile) { spec.Rules[0].Instance = "missing" }},
		{name: "bad severity", mutate: func(spec *rulesFile) { spec.Rules[0].Severity = "urgent" }},
		{name: "partial metadata", mutate: func(spec *rulesFile) { spec.Rules[0].Explanation = "" }},
		{name: "control character in guidance", mutate: func(spec *rulesFile) { spec.Rules[0].PossibleReactions[0] = "review\nlogs" }},
		{name: "missing category", mutate: func(spec *rulesFile) { spec.Rules[0].Category = "" }},
		{name: "noncritical lockdown", mutate: func(spec *rulesFile) { spec.Rules[0].Lockdown = true }},
		{name: "noncritical email", mutate: func(spec *rulesFile) { spec.Rules[0].EmailOnCritical = true }},
		{name: "bad rule pattern", mutate: func(spec *rulesFile) { spec.Rules[0].Pattern = "[" }},
		{name: "bad success pattern", mutate: func(spec *rulesFile) { spec.Rules[0].SuccessPattern = "[" }},
		{name: "bad account capture", mutate: func(spec *rulesFile) { spec.Rules[len(spec.Rules)-1].AccountPattern = "no capture" }},
		{name: "too many requirements", mutate: func(spec *rulesFile) { spec.Rules[0].Require = make([]string, 9) }},
		{name: "too many exclusions", mutate: func(spec *rulesFile) { spec.Rules[0].Exclude = make([]string, 9) }},
		{name: "bad cooldown", mutate: func(spec *rulesFile) { spec.Instances[0].Cooldown = "never" }},
		{name: "disabled instance", mutate: func(spec *rulesFile) { spec.Instances[0].Enabled = false }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			spec, err := loadRulesFile("../../configs/etc/lite-nas/log-analysis/rules.json")
			if err != nil {
				t.Fatalf("load packaged rules: %v", err)
			}
			testCase.mutate(&spec)
			data, err := json.Marshal(spec)
			if err != nil {
				t.Fatalf("encode rules fixture: %v", err)
			}
			path := filepath.Join(t.TempDir(), "rules.json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatalf("write rules fixture: %v", err)
			}
			if _, err := loadRulesFile(path); err == nil {
				t.Fatal("invalid rules file accepted")
			}
		})
	}
}

// Requirements: security-monitoring/FR-001.
func TestRulesFileRejectsMalformedOrOversizedInput(t *testing.T) {
	t.Parallel()
	cases := map[string][]byte{
		"invalid JSON":  []byte("{"),
		"trailing JSON": []byte(`{"event_id_prefix":"SEC"}{}`),
		"oversized":     bytes.Repeat([]byte("x"), 1024*1024+1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "rules.json")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatalf("write rules fixture: %v", err)
			}
			if _, err := loadRulesFile(path); err == nil {
				t.Fatal("invalid rule input accepted")
			}
		})
	}
}

// Requirements: security-monitoring/FR-001.
func TestRulesFileReportsMissingPath(t *testing.T) {
	t.Parallel()
	if _, err := loadRulesFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Fatal("missing rules file was accepted")
	}
}

// Requirements: security-monitoring/FR-007.
func TestAuthFormatExtractsSSHSource(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"Invalid user admin from 203.0.113.9 port 40022":                                     "203.0.113.9",
		"Failed password for alice from 2001:db8::7 port 22 ssh2":                            "2001:db8::7",
		"Connection closed by invalid user admin 127.0.0.1 port 45776 [preauth]":             "127.0.0.1",
		"pam_unix(sshd:auth): authentication failure; logname= uid=0 rhost=192.0.2.4 user=x": "192.0.2.4",
	}
	for message, want := range cases {
		t.Run(want, func(t *testing.T) {
			t.Parallel()
			record := logRecord{Message: message}
			extractAuthFields(&record)
			if record.SourceAddress != want {
				t.Fatalf("source = %q, want %q", record.SourceAddress, want)
			}
		})
	}
}
