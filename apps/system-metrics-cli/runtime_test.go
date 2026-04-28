package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"lite-nas/apps/system-metrics-cli/workers"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/systemmetricstest"
)

type stubOutputWriter struct {
	currentSnapshot  metrics.SystemSnapshot
	currentSelection workers.CurrentSelection
	history          []metrics.SystemSnapshot
	err              error
}

func (w *stubOutputWriter) WriteCurrent(_ io.Writer, snapshot metrics.SystemSnapshot, selection workers.CurrentSelection) error {
	if w.err != nil {
		return w.err
	}

	w.currentSnapshot = snapshot
	w.currentSelection = selection
	return nil
}

func (w *stubOutputWriter) WriteHistory(_ io.Writer, history []metrics.SystemSnapshot) error {
	if w.err != nil {
		return w.err
	}

	w.history = history
	return nil
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/IR-002
func TestExecuteCommandRequestsCurrentSnapshot(t *testing.T) {
	t.Parallel()

	client, _ := executeCurrentCommandFixture(t, metrics.SystemSnapshot{
		Timestamp: time.Unix(1700000000, 0).UTC(),
	})

	if client.Subject != systemmetricscontract.SnapshotRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.Subject, systemmetricscontract.SnapshotRPCSubject)
	}
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/IR-002
func TestExecuteCommandPassesCPUSelection(t *testing.T) {
	t.Parallel()

	_, output := executeCurrentCommandFixture(t, metrics.SystemSnapshot{})

	if !output.currentSelection.CPU {
		t.Fatalf("current selection CPU = %t, want true", output.currentSelection.CPU)
	}
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/IR-002
func TestExecuteCommandPassesRAMSelection(t *testing.T) {
	t.Parallel()

	_, output := executeCurrentCommandFixture(t, metrics.SystemSnapshot{})

	if !output.currentSelection.RAM {
		t.Fatalf("current selection RAM = %t, want true", output.currentSelection.RAM)
	}
}

// Requirements: system-metrics-cli/FR-004, system-metrics-cli/IR-002
func TestExecuteCommandRequestsHistory(t *testing.T) {
	t.Parallel()

	client := systemmetricstest.NewHistoryClient([]metrics.SystemSnapshot{{Timestamp: time.Unix(1700000000, 0).UTC()}})
	output := &stubOutputWriter{}

	mustExecuteCommand(
		t,
		context.Background(),
		workers.Invocation{Mode: workers.ModeHistory},
		client,
		output,
		&bytes.Buffer{},
	)

	if client.Subject != systemmetricscontract.HistoryRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.Subject, systemmetricscontract.HistoryRPCSubject)
	}

	if len(output.history) != 1 {
		t.Fatalf("history length = %d, want 1", len(output.history))
	}
}

// Requirements: system-metrics-cli/IR-001
func TestExecuteCommandRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	err := executeCommand(
		context.Background(),
		workers.Invocation{Mode: workers.Mode("invalid")},
		&systemmetricstest.RequestClient{},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCommand() error = nil, want unsupported mode error")
	}
}

// Requirements: system-metrics-cli/RR-001
func TestExecuteCurrentCommandReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")

	err := executeCurrentCommand(
		context.Background(),
		workers.Invocation{CurrentSelection: workers.CurrentSelection{CPU: true}},
		systemmetricstest.NewRequestErrorClient(wantErr),
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCurrentCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: system-metrics-cli/RR-001
func TestExecuteCurrentCommandReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")

	err := executeCurrentCommand(
		context.Background(),
		workers.Invocation{CurrentSelection: workers.CurrentSelection{CPU: true}},
		systemmetricstest.NewSnapshotClient(metrics.SystemSnapshot{}),
		&stubOutputWriter{err: wantErr},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCurrentCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: system-metrics-cli/RR-001
func TestExecuteHistoryCommandReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")

	err := executeHistoryCommand(
		context.Background(),
		systemmetricstest.NewRequestErrorClient(wantErr),
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeHistoryCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: system-metrics-cli/RR-001
func TestExecuteHistoryCommandReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")

	err := executeHistoryCommand(
		context.Background(),
		systemmetricstest.NewHistoryClient([]metrics.SystemSnapshot{}),
		&stubOutputWriter{err: wantErr},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeHistoryCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: system-metrics-cli/IR-001
func TestPrintUsageWritesCLIUsage(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	printUsage(&output)

	want := "Usage: system-metrics-cli [--config=/etc/lite-nas/system-metrics-cli.conf] [--cpu] [--ram] [--history]\n"
	if output.String() != want {
		t.Fatalf("printUsage() output = %q, want %q", output.String(), want)
	}
}

func mustExecuteCommand(
	t *testing.T,
	ctx context.Context,
	invocation workers.Invocation,
	client *systemmetricstest.RequestClient,
	output *stubOutputWriter,
	stdout io.Writer,
) {
	t.Helper()

	if err := executeCommand(ctx, invocation, client, output, stdout); err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}
}

func executeCurrentCommandFixture(
	t *testing.T,
	snapshot metrics.SystemSnapshot,
) (*systemmetricstest.RequestClient, *stubOutputWriter) {
	t.Helper()

	client := systemmetricstest.NewSnapshotClient(snapshot)
	output := &stubOutputWriter{}

	mustExecuteCommand(
		t,
		context.Background(),
		workers.Invocation{
			Mode:             workers.ModeCurrent,
			CurrentSelection: workers.CurrentSelection{CPU: true, RAM: true},
		},
		client,
		output,
		&bytes.Buffer{},
	)

	return client, output
}
