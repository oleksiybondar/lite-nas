package main

import (
	"strings"
	"testing"
)

// Requirements: security-monitoring/FR-001.
func TestParserExtractsPAMAccount(t *testing.T) {
	t.Parallel()
	record := mustNormalize(t, testSources[0], authDatagram)
	if record.Account != "alice" {
		t.Fatalf("account = %q", record.Account)
	}
}

// Requirements: security-monitoring/FR-001.
func TestParserTrimsSyslogTagSeparator(t *testing.T) {
	t.Parallel()
	record, _ := mustDecodeRecord(t, authDatagram)
	if !strings.HasPrefix(record.Message, "pam_unix(") {
		t.Fatalf("message = %q", record.Message)
	}
}

// Requirements: security-monitoring/FR-006.
func TestParserExtractsUSBDeviceIdentity(t *testing.T) {
	t.Parallel()
	record := mustNormalize(t, testSources[1], kernelDatagram)
	if record.Device != "3-5 203a:fff9" {
		t.Fatalf("device = %q", record.Device)
	}
}

// Requirements: security-monitoring/FR-001.
func TestParserRejectsInvalidDatagram(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"unknown field":     `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"imuxsock","program":"su","message":"x","host":"h"}`,
		"trailing json":     `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"imuxsock","program":"su","message":"x"}{}`,
		"bad timestamp":     `{"timestamp":"yesterday","facility":"auth","input":"imuxsock","program":"su","message":"x"}`,
		"missing message":   `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"imuxsock","program":"su"}`,
		"blank message":     `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"imuxsock","program":"su","message":"   "}`,
		"nul in message":    `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"imuxsock","program":"su","message":"a\u0000b"}`,
		"non-alnum input":   `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"im klog","program":"su","message":"x"}`,
		"oversized message": `{"timestamp":"2026-10-25T08:26:06+01:00","facility":"auth","input":"imuxsock","program":"su","message":"` + strings.Repeat("é", maxLogLineBytes/2+1) + `"}`,
		"not json":          `2026-10-25T08:26:06+01:00 host sshd: failure`,
	}
	for name, datagram := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, _, err := decodeRsyslogRecord([]byte(datagram), testValidator()); err == nil {
				t.Fatal("invalid datagram accepted")
			}
		})
	}
}

// Requirements: security-monitoring/FR-001.
func TestSourceSelectsRecord(t *testing.T) {
	t.Parallel()
	kernelRecord, _ := mustDecodeRecord(t, kernelDatagram)
	authRecord, _ := mustDecodeRecord(t, authDatagram)
	// A local process can claim the kern facility, but not the imklog input.
	forgedKernel := kernelRecord
	forgedKernel.Input = "imuxsock"
	cases := []struct {
		name   string
		source sourceSpec
		record rsyslogRecord
		want   bool
	}{
		{name: "auth facility", source: testSources[0], record: authRecord, want: true},
		{name: "kernel input", source: testSources[1], record: kernelRecord, want: true},
		{name: "kernel source ignores auth", source: testSources[1], record: authRecord, want: false},
		{name: "kernel source ignores forged kernel", source: testSources[1], record: forgedKernel, want: false},
		{name: "program selector", source: sourceSpec{ID: "ssh", Format: formatAuth, Programs: []string{"su"}}, record: authRecord, want: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := testCase.source.selects(testCase.record); got != testCase.want {
				t.Fatalf("selects() = %v, want %v", got, testCase.want)
			}
		})
	}
}

// Requirements: security-monitoring/FR-001.
func TestSourceSpecRejectsInvalidSelection(t *testing.T) {
	t.Parallel()
	cases := map[string]sourceSpec{
		"no selector":      {ID: "auth", Format: formatAuth},
		"unknown format":   {ID: "auth", Format: "rsyslog_auth", Facilities: []string{"auth"}},
		"invalid selector": {ID: "auth", Format: formatAuth, Facilities: []string{"auth priv"}},
		"invalid id":       {ID: "auth/log", Format: formatAuth, Facilities: []string{"auth"}},
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if validSourceSpec(source) {
				t.Fatal("invalid source accepted")
			}
		})
	}
}
