package checks

import "testing"

// TestValidateIdentityOperandRejectsOptionAndControlInputs covers RBAC id operands.
func TestValidateIdentityOperandRejectsOptionAndControlInputs(t *testing.T) {
	for _, value := range []string{"--help", "-u", "name\nother", "name other", ""} {
		if err := ValidateIdentityOperand(value); err == nil {
			t.Errorf("ValidateIdentityOperand(%q) accepted invalid operand", value)
		}
	}
}

// TestValidateSudoArgumentsRejectsInvalidInputs covers RBAC sudo policy operands.
func TestValidateSudoArgumentsRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		uid     string
		command string
	}{
		{uid: "-1", command: "/usr/bin/zfs"},
		{uid: "0", command: "/usr/bin/zfs"},
		{uid: "1002", command: "--help"},
		{uid: "1002", command: "usr/bin/zfs"},
		{uid: "1002", command: "/usr/bin/../bin/zfs"},
		{uid: "1002", command: "/usr/bin/zfs\nother"},
	}
	for _, test := range tests {
		if err := ValidateSudoArguments(test.uid, test.command); err == nil {
			t.Errorf("ValidateSudoArguments(%q, %q) accepted invalid input", test.uid, test.command)
		}
	}
}

// TestValidateSudoArgumentsAcceptsLiteralPathCharacters keeps valid filenames usable.
func TestValidateSudoArgumentsAcceptsLiteralPathCharacters(t *testing.T) {
	if err := ValidateSudoArguments("1002", "/opt/tools/report;archive"); err != nil {
		t.Fatalf("ValidateSudoArguments() rejected literal path characters: %v", err)
	}
}

// TestCanSudoExecRejectsInvalidCommand confirms command validation.
func TestCanSudoExecRejectsInvalidCommand(t *testing.T) {
	runner := newScriptedRunner(nil)
	if _, err := CanSudoExec(t.Context(), runner, "1002", "--help"); err == nil {
		t.Fatal("CanSudoExec() accepted invalid command")
	}
}

// TestCanSudoExecSkipsRunnerForInvalidCommand confirms validation precedes execution.
func TestCanSudoExecSkipsRunnerForInvalidCommand(t *testing.T) {
	runner := newScriptedRunner(nil)
	_, _ = CanSudoExec(t.Context(), runner, "1002", "--help")
	if runner.index != 0 {
		t.Fatalf("runner made %d calls for invalid command", runner.index)
	}
}
