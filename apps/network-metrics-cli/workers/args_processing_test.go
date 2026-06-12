package workers

import (
	"errors"
	"testing"
)

// Requirements: network-metrics-cli/FR-001, network-metrics-cli/FR-003, network-metrics-cli/IR-001
func TestArgsProcessorDefaultsToCurrentMode(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/network-metrics-cli.conf")

	invocation, err := processor.Process(nil)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Mode != ModeCurrent {
		t.Fatalf("Mode = %q, want %q", invocation.Mode, ModeCurrent)
	}
}

// Requirements: network-metrics-cli/FR-001, network-metrics-cli/FR-003, network-metrics-cli/IR-001
func TestArgsProcessorDefaultsToAllSections(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/network-metrics-cli.conf")

	invocation, err := processor.Process(nil)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !invocation.CurrentSelection.Interfaces || !invocation.CurrentSelection.Protocols ||
		!invocation.CurrentSelection.Sockets || !invocation.CurrentSelection.Pressure {
		t.Fatalf("CurrentSelection = %#v, want all sections selected", invocation.CurrentSelection)
	}
}

// Requirements: network-metrics-cli/FR-004, network-metrics-cli/IR-001
func TestArgsProcessorParsesHistoryMode(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/network-metrics-cli.conf")

	invocation, err := processor.Process([]string{"--history"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Mode != ModeHistory {
		t.Fatalf("Mode = %q, want %q", invocation.Mode, ModeHistory)
	}
}

// Requirements: network-metrics-cli/IR-001
func TestArgsProcessorParsesExplicitConfigPath(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/network-metrics-cli.conf")

	invocation, err := processor.Process([]string{"--config=/tmp/network-metrics-cli.conf"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.ConfigPath != "/tmp/network-metrics-cli.conf" {
		t.Fatalf("ConfigPath = %q, want /tmp/network-metrics-cli.conf", invocation.ConfigPath)
	}
}

// Requirements: network-metrics-cli/IR-001
func TestArgsProcessorRejectsHistorySectionCombination(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/network-metrics-cli.conf")

	_, err := processor.Process([]string{"--history", "--interfaces"})
	if err == nil {
		t.Fatal("Process() error = nil, want rejection for incompatible flags")
	}
}

// Requirements: network-metrics-cli/IR-001
func TestArgsProcessorReturnsHelpRequested(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/network-metrics-cli.conf")

	_, err := processor.Process([]string{"--help"})
	if !errors.Is(err, ErrHelpRequested) {
		t.Fatalf("Process() error = %v, want %v", err, ErrHelpRequested)
	}
}
