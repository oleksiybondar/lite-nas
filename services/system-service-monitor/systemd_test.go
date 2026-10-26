package main

import (
	"testing"

	servicecontract "lite-nas/shared/contracts/servicemonitor"
)

// Requirements: task.md/SVC-004.
func TestValidUnitNameRejectsOptionOperand(t *testing.T) {
	t.Parallel()
	if validUnitName("--system.service") {
		t.Fatal("option-shaped unit accepted")
	}
}

// Requirements: task.md/SVC-004.
func TestValidUnitNameRejectsPathOperand(t *testing.T) {
	t.Parallel()
	if validUnitName("/etc/systemd/system/a.service") {
		t.Fatal("path-shaped unit accepted")
	}
}

// Requirements: task.md/SVC-004.
func TestValidUnitNameAcceptsService(t *testing.T) {
	t.Parallel()
	if !validUnitName("lite-nas-auth.service") {
		t.Fatal("valid service rejected")
	}
}

// Requirements: task.md/SVC-004.
func TestParseUnitStateRejectsMalformedRestartCount(t *testing.T) {
	t.Parallel()
	_, err := parseUnitState("lite-nas-auth.service", "ActiveState=failed\nSubState=failed\nResult=exit-code\nNRestarts=bad\n")
	if err == nil {
		t.Fatal("malformed restart count accepted")
	}
}

// Requirements: task.md/SVC-004.
func TestClassifyUnitRecognizesStartLimit(t *testing.T) {
	t.Parallel()
	condition := classifyUnit(nil, unitState{ActiveState: "failed", SubState: "failed", Result: "start-limit-hit"})
	if condition != servicecontract.PersistentFailure {
		t.Fatalf("condition = %s", condition)
	}
}

// Requirements: task.md/SVC-004.
func TestClassifyUnitDistinguishesAutoRestart(t *testing.T) {
	t.Parallel()
	condition := classifyUnit(nil, unitState{ActiveState: "activating", SubState: "auto-restart"})
	if condition != servicecontract.Restarting {
		t.Fatalf("condition = %s", condition)
	}
}

// Requirements: task.md/SVC-004.
func TestClassifyUnitRecognizesRecovery(t *testing.T) {
	t.Parallel()
	previous := &unitState{ActiveState: "failed"}
	condition := classifyUnit(previous, unitState{ActiveState: "active", SubState: "running"})
	if condition != servicecontract.Restored {
		t.Fatalf("condition = %s", condition)
	}
}
