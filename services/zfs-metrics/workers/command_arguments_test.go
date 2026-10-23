package workers

import "testing"

// TestValidateZpoolArgumentsRejectsUnexpectedCommands covers the polling command allowlist.
func TestValidateZpoolArgumentsRejectsUnexpectedCommands(t *testing.T) {
	for _, args := range [][]string{
		{"status", "-P", "-L", "other"},
		{"destroy", "pool"},
		{"list", "-H", "-p", "-o", "name", ";"},
	} {
		if err := validateZpoolArguments(args); err == nil {
			t.Errorf("validateZpoolArguments(%q) accepted unexpected command", args)
		}
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
