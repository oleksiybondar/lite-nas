package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"lite-nas/apps/zfs-metrics-cli/workers"
	zfsmetricscontract "lite-nas/shared/contracts/zfsmetrics"
	"lite-nas/shared/metrics"
)

type requestClientStub struct {
	subject    string
	requestErr error

	snapshotResponse zfsmetricscontract.GetSnapshotResponse
	historyResponse  zfsmetricscontract.GetHistoryResponse
}

func (c *requestClientStub) Request(_ context.Context, subject string, _ any, response any) error {
	c.subject = subject
	if c.requestErr != nil {
		return c.requestErr
	}

	switch res := response.(type) {
	case *zfsmetricscontract.GetSnapshotResponse:
		*res = c.snapshotResponse
	case *zfsmetricscontract.GetHistoryResponse:
		*res = c.historyResponse
	}

	return nil
}

type outputWriterStub struct {
	currentWritten bool
	historyWritten bool
	writeErr       error
}

func (o *outputWriterStub) WriteCurrent(_ io.Writer, _ metrics.ZFSSnapshot) error {
	if o.writeErr != nil {
		return o.writeErr
	}
	o.currentWritten = true
	return nil
}

func (o *outputWriterStub) WriteHistory(_ io.Writer, _ []metrics.ZFSSnapshot) error {
	if o.writeErr != nil {
		return o.writeErr
	}
	o.historyWritten = true
	return nil
}

// Requirements: zfs-metrics-cli/FR-001
func TestExecuteCommandRequestsCurrentSnapshot(t *testing.T) {
	t.Parallel()

	client := &requestClientStub{
		snapshotResponse: zfsmetricscontract.GetSnapshotResponse{Available: true},
	}
	output := &outputWriterStub{}

	err := executeCommand(
		context.Background(),
		workers.Invocation{Mode: workers.ModeCurrent},
		client,
		output,
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("executeCommand() error = %v", err)
	}
	if client.subject != zfsmetricscontract.SnapshotRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.subject, zfsmetricscontract.SnapshotRPCSubject)
	}
	if !output.currentWritten {
		t.Fatal("WriteCurrent() was not called")
	}
}

// Requirements: zfs-metrics-cli/FR-002
func TestExecuteCurrentCommandHandlesUnavailableSnapshot(t *testing.T) {
	t.Parallel()

	client := &requestClientStub{
		snapshotResponse: zfsmetricscontract.GetSnapshotResponse{Available: false},
	}
	output := &outputWriterStub{}
	var stdout bytes.Buffer

	err := executeCurrentCommand(context.Background(), client, output, &stdout)
	if err != nil {
		t.Fatalf("executeCurrentCommand() error = %v", err)
	}
	if output.currentWritten {
		t.Fatal("WriteCurrent() called for unavailable snapshot")
	}
	if stdout.String() != "No ZFS snapshot available yet.\n" {
		t.Fatalf("stdout = %q, want unavailable message", stdout.String())
	}
}

// Requirements: zfs-metrics-cli/FR-003
func TestExecuteHistoryCommandRequestsHistory(t *testing.T) {
	t.Parallel()

	client := &requestClientStub{
		historyResponse: zfsmetricscontract.GetHistoryResponse{Items: []metrics.ZFSSnapshot{{}}},
	}
	output := &outputWriterStub{}

	err := executeHistoryCommand(context.Background(), client, output, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("executeHistoryCommand() error = %v", err)
	}
	if client.subject != zfsmetricscontract.HistoryRPCSubject {
		t.Fatalf("Request() subject = %q, want %q", client.subject, zfsmetricscontract.HistoryRPCSubject)
	}
	if !output.historyWritten {
		t.Fatal("WriteHistory() was not called")
	}
}

// Requirements: zfs-metrics-cli/FR-006
func TestExecuteCurrentCommandReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")
	err := executeCurrentCommand(
		context.Background(),
		&requestClientStub{requestErr: wantErr},
		&outputWriterStub{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCurrentCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: zfs-metrics-cli/FR-006
func TestExecuteCurrentCommandReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")
	err := executeCurrentCommand(
		context.Background(),
		&requestClientStub{snapshotResponse: zfsmetricscontract.GetSnapshotResponse{Available: true}},
		&outputWriterStub{writeErr: wantErr},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeCurrentCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: zfs-metrics-cli/FR-006
func TestExecuteHistoryCommandReturnsRequestError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("request failed")
	err := executeHistoryCommand(
		context.Background(),
		&requestClientStub{requestErr: wantErr},
		&outputWriterStub{},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeHistoryCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: zfs-metrics-cli/FR-006
func TestExecuteHistoryCommandReturnsOutputError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")
	err := executeHistoryCommand(
		context.Background(),
		&requestClientStub{historyResponse: zfsmetricscontract.GetHistoryResponse{Items: []metrics.ZFSSnapshot{}}},
		&outputWriterStub{writeErr: wantErr},
		&bytes.Buffer{},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("executeHistoryCommand() error = %v, want %v", err, wantErr)
	}
}

// Requirements: zfs-metrics-cli/FR-005
func TestExecuteCommandRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()

	err := executeCommand(
		context.Background(),
		workers.Invocation{Mode: workers.Mode("invalid")},
		&requestClientStub{},
		&outputWriterStub{},
		&bytes.Buffer{},
	)
	if err == nil {
		t.Fatal("executeCommand() error = nil, want unsupported mode error")
	}
}

func TestPrintUsageWritesCLIUsage(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	printUsage(&output)

	want := "Usage: zfs-metrics-cli [--config=/etc/lite-nas/zfs-metrics-cli.conf] [--history]\n"
	if output.String() != want {
		t.Fatalf("printUsage() output = %q, want %q", output.String(), want)
	}
}

func TestRunReturnsCanceledOnHelp(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"--help"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v, want context.Canceled", err)
	}
}

func TestRunReturnsArgumentError(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"--unknown"})
	if err == nil {
		t.Fatal("run() error = nil, want argument error")
	}
	if !strings.Contains(err.Error(), "unknown argument") {
		t.Fatalf("run() error = %v, want unknown argument error", err)
	}
}
