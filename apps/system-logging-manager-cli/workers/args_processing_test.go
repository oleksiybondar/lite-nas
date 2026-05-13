package workers

import (
	"errors"
	"testing"
)

func TestArgsProcessorParsesCreateEventInvocation(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "createEvent",
		"--data", `{"category":"cpu"}`,
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Command != CommandCreateEvent {
		t.Fatalf("Command = %q, want %q", invocation.Command, CommandCreateEvent)
	}
	if invocation.Data == "" {
		t.Fatal("Data is empty, want parsed JSON payload")
	}
}

func TestArgsProcessorParsesCreateOccurrenceAlias(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "createOccurence",
		"--eventID", "event_1",
		"--data", `{"timestamp":"2026-05-12T10:00:00Z","value_type":"text","value_text":"x"}`,
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Command != CommandCreateOccurrence {
		t.Fatalf("Command = %q, want %q", invocation.Command, CommandCreateOccurrence)
	}
}

func TestArgsProcessorParsesGetActiveUnacknowledgedAlias(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "getActiveUnacknowladgedEvents",
		"--page", "2",
		"--json",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Command != CommandGetActiveUnacknowledgedEvents {
		t.Fatalf("Command = %q, want %q", invocation.Command, CommandGetActiveUnacknowledgedEvents)
	}
	if invocation.Page != 2 {
		t.Fatalf("Page = %d, want 2", invocation.Page)
	}
	if !invocation.JSONOutput {
		t.Fatal("JSONOutput = false, want true")
	}
}

func TestArgsProcessorParsesGetEventsAlias(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "getEvents",
		"--page", "1",
		"--json",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Command != CommandGetAlerts {
		t.Fatalf("Command = %q, want %q", invocation.Command, CommandGetAlerts)
	}
}

func TestArgsProcessorParsesGetEventInvocation(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	invocation, err := processor.Process([]string{
		"--cmd", "getEvent",
		"--eventID", "event_1",
		"--json",
	})
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}

	if invocation.Command != CommandGetEvent {
		t.Fatalf("Command = %q, want %q", invocation.Command, CommandGetEvent)
	}
	if invocation.EventID != "event_1" {
		t.Fatalf("EventID = %q, want %q", invocation.EventID, "event_1")
	}
}

func TestArgsProcessorRequiresCommand(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process(nil)
	if err == nil {
		t.Fatal("Process() error = nil, want missing command error")
	}
}

func TestArgsProcessorReturnsHelpRequested(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--help"})
	if !errors.Is(err, ErrHelpRequested) {
		t.Fatalf("Process() error = %v, want %v", err, ErrHelpRequested)
	}
}

func TestArgsProcessorRejectsUnknownArgument(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--unknown"})
	if err == nil {
		t.Fatal("Process() error = nil, want unknown argument error")
	}
}

func TestArgsProcessorRejectsMissingFlagValue(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd"})
	if err == nil {
		t.Fatal("Process() error = nil, want missing value error")
	}
}

func TestArgsProcessorRejectsCreateOccurrenceWithoutEventID(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd", "createOccurrence", "--data", `{"value_type":"text"}`})
	if err == nil {
		t.Fatal("Process() error = nil, want eventID validation error")
	}
}

func TestArgsProcessorRejectsGetEventWithoutEventID(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd", "getEvent"})
	if err == nil {
		t.Fatal("Process() error = nil, want eventID validation error")
	}
}

func TestArgsProcessorRejectsListPageLowerThanOne(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd", "getActiveEvents", "--page", "0"})
	if err == nil {
		t.Fatal("Process() error = nil, want page validation error")
	}
}

func TestArgsProcessorRejectsNegativePageSize(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd", "getActiveEvents", "--pageSize", "-1"})
	if err == nil {
		t.Fatal("Process() error = nil, want pageSize validation error")
	}
}

func TestArgsProcessorRejectsInvalidPageValue(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd", "getActiveEvents", "--page", "x"})
	if err == nil {
		t.Fatal("Process() error = nil, want parse error")
	}
}

func TestArgsProcessorRejectsInvalidPageSizeValue(t *testing.T) {
	t.Parallel()

	processor := NewArgsProcessor("/etc/lite-nas/system-logging-manager-cli.conf")
	_, err := processor.Process([]string{"--cmd", "getActiveEvents", "--pageSize", "x"})
	if err == nil {
		t.Fatal("Process() error = nil, want parse error")
	}
}
