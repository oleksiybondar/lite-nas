package rfc5424_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/logger/formatters/rfc5424"
)

func TestFormatterFormatsRFC5424Line(t *testing.T) {
	t.Parallel()

	output := formatRFC5424Fixture(t)
	got := string(output)
	want := "<14>1 2026-04-23T12:34:56.123Z rpi-srv-box system-metrics 1823 CPU_SAMPLE - cpu usage calculated\n"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestFormatterSanitizesNewlinesToSingleLine(t *testing.T) {
	t.Parallel()

	got := string(formatSanitizedRFC5424Fixture(t))
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("expected single line output, got %q", got)
	}
}

func TestFormatterIncludesSanitizedMetadataAndAttrs(t *testing.T) {
	t.Parallel()

	got := string(formatSanitizedRFC5424Fixture(t))
	if !strings.Contains(got, "host_name system_metrics 123 - - line one line two detail=first second") {
		t.Fatalf("unexpected sanitized output: %q", got)
	}
}

func TestSeverityMapsSlogLevelsToSyslogSeverity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		level slog.Level
		want  int
	}{
		{level: slog.LevelDebug, want: 7},
		{level: slog.LevelInfo, want: 6},
		{level: slog.LevelWarn, want: 4},
		{level: slog.LevelError, want: 3},
	}

	for _, testCase := range cases {
		if got := rfc5424.Severity(testCase.level); got != testCase.want {
			t.Fatalf("Severity(%s) = %d, want %d", testCase.level, got, testCase.want)
		}
	}
}

func formatRFC5424Fixture(t *testing.T) []byte {
	t.Helper()

	formatter := rfc5424.New(rfc5424.Config{
		Facility:  1,
		ProcessID: "1823",
		MessageID: "CPU_SAMPLE",
	})

	record := slog.NewRecord(
		time.Date(2026, 4, 23, 12, 34, 56, 123000000, time.UTC),
		slog.LevelInfo,
		"cpu usage calculated",
		0,
	)
	record.AddAttrs(
		slog.Int("core", 0),
		slog.String("usage", "12.4"),
	)

	output, err := formatter.Format(record, nil, sharedlogger.RecordMetadata{
		ServiceName: "system-metrics",
		Hostname:    "rpi-srv-box",
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	return output
}

func formatSanitizedRFC5424Fixture(t *testing.T) []byte {
	t.Helper()

	formatter := rfc5424.New(rfc5424.Config{ProcessID: "123"})
	record := slog.NewRecord(
		time.Date(2026, 4, 23, 12, 34, 56, 0, time.UTC),
		slog.LevelWarn,
		"line one\nline two",
		0,
	)

	output, err := formatter.Format(record, []slog.Attr{
		slog.String("detail", "first\nsecond"),
	}, sharedlogger.RecordMetadata{
		ServiceName: "system\nmetrics",
		Hostname:    "host name",
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	return output
}
