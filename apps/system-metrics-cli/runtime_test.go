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
)

type stubRequestClient struct {
	subject  string
	request  any
	response any
	err      error
}

func (c *stubRequestClient) Request(_ context.Context, subject string, request any, response any) error {
	c.subject = subject
	c.request = request

	if c.err != nil {
		return c.err
	}

	switch payload := c.response.(type) {
	case systemmetricscontract.GetSnapshotResponse:
		target, ok := response.(*systemmetricscontract.GetSnapshotResponse)
		if !ok {
			return errors.New("unexpected current response target")
		}

		*target = payload
	case systemmetricscontract.GetHistoryResponse:
		target, ok := response.(*systemmetricscontract.GetHistoryResponse)
		if !ok {
			return errors.New("unexpected history response target")
		}

		*target = payload
	default:
		return errors.New("unexpected stub response type")
	}

	return nil
}

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

	client := &stubRequestClient{
		response: systemmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot: metrics.SystemSnapshot{
				Timestamp: time.Unix(1700000000, 0).UTC(),
			},
		},
	}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Mode:             workers.ModeCurrent,
			CurrentSelection: workers.CurrentSelection{CPU: true, RAM: true},
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if client.subject != systemmetricscontract.SnapshotRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.subject, systemmetricscontract.SnapshotRPCSubject)
	}
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/IR-002
func TestExecuteCommandPassesCPUSelection(t *testing.T) {
	t.Parallel()

	client := &stubRequestClient{
		response: systemmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  metrics.SystemSnapshot{},
		},
	}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Mode:             workers.ModeCurrent,
			CurrentSelection: workers.CurrentSelection{CPU: true, RAM: true},
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if !output.currentSelection.CPU {
		t.Fatalf("current selection CPU = %t, want true", output.currentSelection.CPU)
	}
}

// Requirements: system-metrics-cli/FR-001, system-metrics-cli/IR-002
func TestExecuteCommandPassesRAMSelection(t *testing.T) {
	t.Parallel()

	client := &stubRequestClient{
		response: systemmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  metrics.SystemSnapshot{},
		},
	}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{
			Mode:             workers.ModeCurrent,
			CurrentSelection: workers.CurrentSelection{CPU: true, RAM: true},
		},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if !output.currentSelection.RAM {
		t.Fatalf("current selection RAM = %t, want true", output.currentSelection.RAM)
	}
}

// Requirements: system-metrics-cli/FR-004, system-metrics-cli/IR-002
func TestExecuteCommandRequestsHistory(t *testing.T) {
	t.Parallel()

	client := &stubRequestClient{
		response: systemmetricscontract.GetHistoryResponse{
			Items: []metrics.SystemSnapshot{{Timestamp: time.Unix(1700000000, 0).UTC()}},
		},
	}
	output := &stubOutputWriter{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{Mode: workers.ModeHistory},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}

	if client.subject != systemmetricscontract.HistoryRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.subject, systemmetricscontract.HistoryRPCSubject)
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
		&stubRequestClient{},
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
		&stubRequestClient{err: wantErr},
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
		&stubRequestClient{
			response: systemmetricscontract.GetSnapshotResponse{
				Available: true,
				Snapshot:  metrics.SystemSnapshot{},
			},
		},
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
		&stubRequestClient{err: wantErr},
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
		&stubRequestClient{
			response: systemmetricscontract.GetHistoryResponse{Items: []metrics.SystemSnapshot{}},
		},
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
