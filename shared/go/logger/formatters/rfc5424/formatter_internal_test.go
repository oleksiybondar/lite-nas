package rfc5424

import (
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	sharedlogger "lite-nas/shared/logger"
)

func TestNewAppliesDefaultValues(t *testing.T) {
	t.Parallel()

	formatter := New(Config{})
	if formatter.facility != defaultFacility || formatter.processID != "-" || formatter.messageID != "-" {
		t.Fatalf("unexpected formatter defaults: %#v", formatter)
	}
}

func TestFormatHandlesZeroTimestampAndEmptyMetadata(t *testing.T) {
	t.Parallel()

	formatter := New(Config{})
	record := slog.NewRecord(time.Time{}, slog.LevelInfo, "sample", 0)

	output, err := formatter.Format(record, nil, sharedlogger.RecordMetadata{})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	got := string(output)
	if !strings.Contains(got, " - - - - - sample\n") {
		t.Fatalf("unexpected zero-value output: %q", got)
	}
}

func TestFormatCoversPrimitiveAndFallbackAttrValues(t *testing.T) {
	t.Parallel()

	formatter := New(Config{ProcessID: "7"})
	record := slog.NewRecord(time.Date(2026, 4, 23, 12, 34, 56, 0, time.UTC), slog.LevelInfo, "sample", 0)

	output, err := formatter.Format(record, []slog.Attr{
		slog.Bool("ok", true),
		slog.Int64("count", 4),
		slog.Uint64("size", 5),
		slog.Float64("ratio", 1.5),
		slog.Duration("elapsed", 2*time.Second),
		slog.Time("at", time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)),
		slog.Any("err", errors.New("boom")),
		slog.Any("", struct{ Value string }{Value: "x"}),
	}, sharedlogger.RecordMetadata{
		ServiceName: "",
		Hostname:    "",
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	got := string(output)
	for _, expected := range []string{
		"ok=true",
		"count=4",
		"size=5",
		"ratio=1.5",
		"elapsed=2s",
		"at=2026-04-23T12:00:00Z",
		"err=boom",
		"attr={x}",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected %q in output %q", expected, got)
		}
	}
}
