package workers

import (
	"errors"
	"testing"
)

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/FR-003, system-metrics-cli/IR-001
func TestArgsProcessorDefaultsToCurrentMode(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	invocation, err := processor.Process(nil)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Mode != ModeCurrent {
		t.Fatalf("Mode = %q, want %q", invocation.Mode, ModeCurrent)
	}
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/FR-003, system-metrics-cli/IR-001
func TestArgsProcessorDefaultsToCPUSelection(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	invocation, err := processor.Process(nil)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !invocation.CurrentSelection.CPU {
		t.Fatalf("CurrentSelection.CPU = %t, want true", invocation.CurrentSelection.CPU)
	}
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/FR-003, system-metrics-cli/IR-001
func TestArgsProcessorDefaultsToRAMSelection(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	invocation, err := processor.Process(nil)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if !invocation.CurrentSelection.RAM {
		t.Fatalf("CurrentSelection.RAM = %t, want true", invocation.CurrentSelection.RAM)
	}
}

// Requirements: system-metrics-cli/FR-004, system-metrics-cli/IR-001
func TestArgsProcessorParsesHistoryMode(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	invocation, err := processor.Process([]string{"--history"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Mode != ModeHistory {
		t.Fatalf("Mode = %q, want %q", invocation.Mode, ModeHistory)
	}
}

// Requirements: system-metrics-cli/IR-001
func TestArgsProcessorParsesExplicitConfigPath(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	invocation, err := processor.Process([]string{"--config=/tmp/system-metrics-cli.conf"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.ConfigPath != "/tmp/system-metrics-cli.conf" {
		t.Fatalf("ConfigPath = %q, want /tmp/system-metrics-cli.conf", invocation.ConfigPath)
	}
}

// Requirements: system-metrics-cli/IR-001
func TestArgsProcessorRejectsHistorySectionCombination(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	_, err := processor.Process([]string{"--history", "--cpu"})
	if err == nil {
		t.Fatal("Process() error = nil, want rejection for incompatible flags")
	}
}

// Requirements: system-metrics-cli/IR-001
func TestArgsProcessorReturnsHelpRequested(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/liteNAS/system-metrics-cli.conf")

	_, err := processor.Process([]string{"--help"})
	if !errors.Is(err, ErrHelpRequested) {
		t.Fatalf("Process() error = %v, want %v", err, ErrHelpRequested)
	}
}
