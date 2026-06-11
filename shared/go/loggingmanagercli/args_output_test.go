package loggingmanagercli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

func TestArgsProcessorParsesListCommand(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/default.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "getAlerts",
		"--page", "2",
		"--pageSize", "20",
		"--json",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if invocation.Command != CommandGetAlerts || invocation.Page != 2 || invocation.PageSize != 20 || !invocation.JSONOutput {
		t.Fatalf("invocation = %#v", invocation)
	}
}

func TestArgsProcessorNormalizesAliases(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/default.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "getEvents",
		"--page", "1",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if invocation.Command != CommandGetAlerts {
		t.Fatalf("Command = %q, want %q", invocation.Command, CommandGetAlerts)
	}
}

func TestArgsProcessorRejectsUnknownArgument(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/default.conf")
	_, err := processor.Process([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestArgsProcessorReturnsHelpRequested(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/default.conf")
	_, err := processor.Process([]string{"--help"})
	if !errors.Is(err, ErrHelpRequested) {
		t.Fatalf("Process() error = %v, want ErrHelpRequested", err)
	}
}

func TestOutputWriterWriteEventsTextAndJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	buffer := &bytes.Buffer{}
	events := []loggingmanagercontract.ListAlertItem{{
		EventID:      "event_1",
		Category:     "disk",
		Severity:     "warning",
		Status:       "active",
		Acknowledged: true,
		Muted:        false,
		Source:       "system",
		CreatedAt:    "2026-05-13T10:00:00Z",
	}}
	if err := writer.WriteEvents(buffer, events, false); err != nil {
		t.Fatalf("WriteEvents(text) error = %v", err)
	}
	if !strings.Contains(buffer.String(), "EVENT_ID") || !strings.Contains(buffer.String(), "event_1") {
		t.Fatalf("text output = %q", buffer.String())
	}

	buffer.Reset()
	if err := writer.WriteEvents(buffer, events, true); err != nil {
		t.Fatalf("WriteEvents(json) error = %v", err)
	}
	if !strings.Contains(buffer.String(), "\"EventID\": \"event_1\"") {
		t.Fatalf("json output = %q", buffer.String())
	}
}

func TestOutputWriterWriteOKTextAndJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	buffer := &bytes.Buffer{}
	if err := writer.WriteOK(buffer, loggingmanagercontract.OKResponse{OK: true}, false); err != nil {
		t.Fatalf("WriteOK(text) error = %v", err)
	}
	if !strings.Contains(buffer.String(), "OK=TRUE") {
		t.Fatalf("text output = %q", buffer.String())
	}

	buffer.Reset()
	if err := writer.WriteOK(buffer, loggingmanagercontract.OKResponse{OK: false}, true); err != nil {
		t.Fatalf("WriteOK(json) error = %v", err)
	}
	if !strings.Contains(buffer.String(), "\"ok\": false") {
		t.Fatalf("json output = %q", buffer.String())
	}
}

func TestRunPrintsUsageOnHelp(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	err := Run(
		context.Background(),
		[]string{"--help"},
		"/etc/default.conf",
		"test-cli",
		Subjects{},
		func(string, string) (func(), MessagingClient, error) {
			t.Fatal("loadInfra should not be called")
			return nil, nil, nil
		},
		buffer,
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if !strings.Contains(buffer.String(), "Usage: test-cli") {
		t.Fatalf("usage output = %q", buffer.String())
	}
}
