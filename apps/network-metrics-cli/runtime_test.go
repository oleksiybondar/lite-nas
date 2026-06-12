package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"lite-nas/apps/network-metrics-cli/workers"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	"lite-nas/shared/metrics"
	sharedmetricscli "lite-nas/shared/metricscli"
)

type requestClientStub struct {
	subject  string
	response any
	err      error
}

func (c *requestClientStub) Request(_ context.Context, subject string, _ any, response any) error {
	c.subject = subject
	if c.err != nil {
		return c.err
	}

	switch payload := c.response.(type) {
	case networkmetricscontract.GetSnapshotResponse:
		typed := response.(*networkmetricscontract.GetSnapshotResponse)
		*typed = payload
	case networkmetricscontract.GetHistoryResponse:
		typed := response.(*networkmetricscontract.GetHistoryResponse)
		*typed = payload
	}

	return nil
}

type stubOutputWriter struct {
	currentSnapshot  metrics.NetworkMetricsSnapshot
	currentSelection workers.CurrentSelection
	history          []metrics.NetworkMetricsSnapshot
	err              error
}

func (w *stubOutputWriter) WriteCurrent(_ io.Writer, snapshot metrics.NetworkMetricsSnapshot, selection workers.CurrentSelection) error {
	if w.err != nil {
		return w.err
	}

	w.currentSnapshot = snapshot
	w.currentSelection = selection
	return nil
}

func (w *stubOutputWriter) WriteHistory(_ io.Writer, history []metrics.NetworkMetricsSnapshot) error {
	if w.err != nil {
		return w.err
	}

	w.history = history
	return nil
}

// Requirements: network-metrics-cli/FR-001, network-metrics-cli/IR-002
func TestExecuteCommandRequestsCurrentSnapshot(t *testing.T) {
	t.Parallel()

	client, _ := executeCurrentCommandFixture(t, metrics.NetworkMetricsSnapshot{
		Timestamp: time.Unix(1700000000, 0).UTC(),
	})

	if client.subject != networkmetricscontract.SnapshotRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.subject, networkmetricscontract.SnapshotRPCSubject)
	}
}

// Requirements: network-metrics-cli/FR-001, network-metrics-cli/IR-002
func TestExecuteCommandPassesCurrentSelection(t *testing.T) {
	t.Parallel()

	_, output := executeCurrentCommandFixture(t, metrics.NetworkMetricsSnapshot{})

	if !output.currentSelection.Interfaces || !output.currentSelection.Protocols ||
		!output.currentSelection.Sockets || !output.currentSelection.Pressure {
		t.Fatalf("current selection = %#v, want all sections true", output.currentSelection)
	}
}

// Requirements: network-metrics-cli/FR-004, network-metrics-cli/IR-002
func TestExecuteCommandRequestsHistory(t *testing.T) {
	t.Parallel()

	client := &requestClientStub{
		response: networkmetricscontract.GetHistoryResponse{
			Items: []metrics.NetworkMetricsSnapshot{{Timestamp: time.Unix(1700000000, 0).UTC()}},
		},
	}
	output := &stubOutputWriter{}

	mustExecuteCommand(
		t,
		context.Background(),
		workers.Invocation{Mode: workers.ModeHistory},
		client,
		output,
		&bytes.Buffer{},
	)

	if client.subject != networkmetricscontract.HistoryRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.subject, networkmetricscontract.HistoryRPCSubject)
	}

	if len(output.history) != 1 {
		t.Fatalf("history length = %d, want 1", len(output.history))
	}
}

// Requirements: network-metrics-cli/IR-001
func TestExecuteCommandRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	err := executeCommand(
		context.Background(),
		workers.Invocation{Mode: workers.Mode("invalid")},
		&requestClientStub{},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCommand() error = nil, want unsupported mode error")
	}
}

// Requirements: network-metrics-cli/RR-001
func TestExecuteCurrentCommandReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")

	err := executeCurrentCommand(
		context.Background(),
		workers.Invocation{CurrentSelection: workers.CurrentSelection{Interfaces: true}},
		&requestClientStub{err: wantErr},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCurrentCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: network-metrics-cli/RR-001
func TestExecuteCurrentCommandReturnsUnavailableError(t *testing.T) {
	t.Parallel()

	err := executeCurrentCommand(
		context.Background(),
		workers.Invocation{CurrentSelection: workers.CurrentSelection{Interfaces: true}},
		&requestClientStub{response: networkmetricscontract.GetSnapshotResponse{Available: false}},
		&stubOutputWriter{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCurrentCommand() error = nil, want unavailable snapshot error")
	}
}

// Requirements: network-metrics-cli/RR-001
func TestExecuteHistoryCommandReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")

	err := executeHistoryCommand(
		context.Background(),
		&requestClientStub{response: networkmetricscontract.GetHistoryResponse{}},
		&stubOutputWriter{err: wantErr},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeHistoryCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: network-metrics-cli/IR-001
func TestPrintUsageWritesCLIUsage(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer

	printUsage(&output)

	want := "Usage: network-metrics-cli [--config=/etc/lite-nas/network-metrics-cli.conf] [--interfaces] [--protocols] [--sockets] [--pressure] [--history]\n"
	if output.String() != want {
		t.Fatalf("printUsage() output = %q, want %q", output.String(), want)
	}
}

func mustExecuteCommand(
	t *testing.T,
	ctx context.Context,
	invocation workers.Invocation,
	client sharedmetricscli.RequestClient,
	output workers.OutputWriter,
	stdout io.Writer,
) {
	t.Helper()

	if err := executeCommand(ctx, invocation, client, output, stdout); err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}
}

func executeCurrentCommandFixture(
	t *testing.T,
	snapshot metrics.NetworkMetricsSnapshot,
) (*requestClientStub, *stubOutputWriter) {
	t.Helper()

	client := &requestClientStub{
		response: networkmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		},
	}
	output := &stubOutputWriter{}

	mustExecuteCommand(
		t,
		context.Background(),
		workers.Invocation{
			Mode: workers.ModeCurrent,
			CurrentSelection: workers.CurrentSelection{
				Interfaces: true,
				Protocols:  true,
				Sockets:    true,
				Pressure:   true,
			},
		},
		client,
		output,
		&bytes.Buffer{},
	)

	return client, output
}
