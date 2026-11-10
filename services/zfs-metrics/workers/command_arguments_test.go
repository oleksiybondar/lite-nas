package workers

import (
	"slices"
	"testing"
)

// TestValidateZpoolArgumentsRejectsUnexpectedCommands covers the polling command allowlist.
func TestValidateZpoolArgumentsRejectsUnexpectedCommands(t *testing.T) {
	for _, args := range [][]string{
		{"status", "-P", "-L", "other"},
		{"destroy", "pool"},
		{"list", "-H", "-p", "-o", "name", ";"},
		{"iostat", "-H", "-p"},
	} {
		if err := validateZpoolArguments(args); err == nil {
			t.Errorf("validateZpoolArguments(%q) accepted unexpected command", args)
		}
	}
}

// TestZpoolIostatArgsUseIntervalSample covers the live-sampling command contract.
func TestZpoolIostatArgsUseIntervalSample(t *testing.T) {
	expected := []string{"iostat", "-H", "-p", "-y", "1", "1"}
	if !slices.Equal(zpoolIostatArgs, expected) {
		t.Fatalf("zpoolIostatArgs = %q, want %q", zpoolIostatArgs, expected)
	}
}

// TestValidateZpoolArgumentsAcceptsPollingCommands covers each configured read-only invocation.
func TestValidateZpoolArgumentsAcceptsPollingCommands(t *testing.T) {
	for _, args := range [][]string{zpoolStatusArgs, zpoolListArgs, zpoolIostatArgs} {
		if err := validateZpoolArguments(args); err != nil {
			t.Errorf("validateZpoolArguments(%q) rejected polling command: %v", args, err)
		}
	}
}
