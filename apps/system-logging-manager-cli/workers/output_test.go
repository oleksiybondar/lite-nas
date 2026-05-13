package workers

import (
	"bytes"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	"lite-nas/shared/loggingmanager/enum"
)

func TestOutputWriterWritesEventsAsTable(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var out bytes.Buffer

	err := writer.WriteEvents(
		&out,
		[]loggingmanagercontract.ListAlertItem{
			{
				EventID:      "event_1",
				Category:     "disk_health",
				Severity:     enum.SeverityWarning,
				Status:       enum.StatusActive,
				Acknowledged: false,
				Muted:        false,
				Source:       "raid-monitor",
				CreatedAt:    "2026-05-12T14:30:00Z",
			},
		},
		false,
	)
	if err != nil {
		t.Fatalf("WriteEvents() error = %v", err)
	}

	got := out.String()
	if !bytes.Contains([]byte(got), []byte("EVENT_ID")) {
		t.Fatalf("WriteEvents() output missing header: %q", got)
	}
	if !bytes.Contains([]byte(got), []byte("event_1")) {
		t.Fatalf("WriteEvents() output missing row: %q", got)
	}
}

func TestOutputWriterWritesEventsAsJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var out bytes.Buffer

	err := writer.WriteEvents(&out, []loggingmanagercontract.ListAlertItem{}, true)
	if err != nil {
		t.Fatalf("WriteEvents() error = %v", err)
	}

	want := "[]\n"
	if out.String() != want {
		t.Fatalf("WriteEvents() output = %q, want %q", out.String(), want)
	}
}

func TestOutputWriterWritesOKInHumanReadableFormat(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var out bytes.Buffer

	err := writer.WriteOK(&out, loggingmanagercontract.OKResponse{OK: true}, false)
	if err != nil {
		t.Fatalf("WriteOK() error = %v", err)
	}

	want := "OK=TRUE\n"
	if out.String() != want {
		t.Fatalf("WriteOK() output = %q, want %q", out.String(), want)
	}
}

func TestOutputWriterWritesOKAsJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var out bytes.Buffer

	err := writer.WriteOK(&out, loggingmanagercontract.OKResponse{OK: true}, true)
	if err != nil {
		t.Fatalf("WriteOK() error = %v", err)
	}

	want := "{\n  \"ok\": true\n}\n"
	if out.String() != want {
		t.Fatalf("WriteOK() output = %q, want %q", out.String(), want)
	}
}
