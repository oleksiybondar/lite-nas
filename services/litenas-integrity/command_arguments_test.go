package main

import "testing"

// TestRunAIDECommandRejectsUnexpectedSubcommand covers the fixed AIDE invocation.
func TestRunAIDECommandRejectsUnexpectedSubcommand(t *testing.T) {
	service := &integrityService{}
	if _, err := service.runAIDECommand(t.Context(), "--config=/tmp/other"); err == nil {
		t.Fatal("runAIDECommand() accepted unexpected subcommand")
	}
}
