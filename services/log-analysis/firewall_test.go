package main

import (
	"testing"
	"time"
)

// TestParseFirewallAttemptAcceptsTrustedBlockedTCP verifies factual extraction.
// Requirements: network-security/FR-001.
func TestParseFirewallAttemptAcceptsTrustedBlockedTCP(t *testing.T) {
	record := logRecord{Format: formatKernel, KernelTrusted: true, At: time.Now(), Message: "[UFW BLOCK] IN=eth0 OUT= MAC= SRC=192.0.2.9 DST=192.0.2.10 LEN=60 PROTO=TCP SPT=48123 DPT=443"}
	attempt, ok := parseFirewallAttempt(record)
	if !ok || attempt.SourceIP != "192.0.2.9" || attempt.DestinationPort != 443 {
		t.Fatalf("unexpected firewall attempt: %+v, %v", attempt, ok)
	}
}

// TestParseFirewallAttemptRejectsUntrustedKernelClaim protects the source boundary.
// Requirements: network-security/FR-001.
func TestParseFirewallAttemptRejectsUntrustedKernelClaim(t *testing.T) {
	record := logRecord{Format: formatKernel, Message: "[UFW BLOCK] IN=eth0 SRC=192.0.2.9 DST=192.0.2.10 PROTO=TCP DPT=443"}
	_, ok := parseFirewallAttempt(record)
	if ok {
		t.Fatal("untrusted kernel claim was accepted")
	}
}

// TestParseFirewallAttemptRejectsMalformedPort verifies port bounds.
// Requirements: network-security/FR-001.
func TestParseFirewallAttemptRejectsMalformedPort(t *testing.T) {
	record := logRecord{Format: formatKernel, KernelTrusted: true, Message: "[UFW BLOCK] IN=eth0 SRC=192.0.2.9 DST=192.0.2.10 PROTO=TCP DPT=-1"}
	_, ok := parseFirewallAttempt(record)
	if ok {
		t.Fatal("malformed destination port was accepted")
	}
}
