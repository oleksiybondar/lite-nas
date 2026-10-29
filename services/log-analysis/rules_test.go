package main

import "testing"

// sampleRuleCase pairs one packaged rule with a normalized log message and
// whether that rule must select it. Messages are real host records unless a
// comment states their origin.
type sampleRuleCase struct {
	name    string
	rule    string
	message string
	want    bool
}

// Requirements: security-monitoring/FR-007.
func TestSampleRulesSelectSecurityRecords(t *testing.T) {
	t.Parallel()
	cases := []sampleRuleCase{
		{
			name: "sudo password failure", rule: "pam_password_guessing", want: true,
			message: "pam_unix(sudo:auth): authentication failure; logname=alice uid=1001 euid=0 tty=/dev/pts/1 ruser=alice rhost=  user=alice",
		},
		{
			name: "sudo session does not reset guessing", rule: "pam_password_guessing", want: false,
			message: "pam_unix(sudo:session): session opened for user alice(uid=1001) by (uid=969)",
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

// Requirements: security-monitoring/FR-007.
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
