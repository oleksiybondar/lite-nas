package workers

import (
	"errors"
	"testing"
)

// Requirements: zfs-metrics-cli/FR-005
func TestProcessDefaultsToCurrentMode(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/zfs-metrics-cli.conf")
	invocation, err := processor.Process([]string{})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Mode != ModeCurrent {
		t.Fatalf("invocation.Mode = %s, want %s", invocation.Mode, ModeCurrent)
	}
}

// Requirements: zfs-metrics-cli/FR-005
func TestProcessSetsHistoryMode(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/zfs-metrics-cli.conf")
	invocation, err := processor.Process([]string{"--history"})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Mode != ModeHistory {
		t.Fatalf("invocation.Mode = %s, want %s", invocation.Mode, ModeHistory)
	}
}

// Requirements: zfs-metrics-cli/FR-005
func TestProcessReturnsHelpRequested(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/zfs-metrics-cli.conf")
	_, err := processor.Process([]string{"--help"})
	if !errors.Is(err, ErrHelpRequested) {
		t.Fatalf("Process() error = %v, want ErrHelpRequested", err)
	}
}
