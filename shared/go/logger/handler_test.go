package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	sharedlogger "lite-nas/shared/logger"
)

type stubFormatter struct{}

func (stubFormatter) Format(
	record slog.Record,
	attrs []slog.Attr,
	_ sharedlogger.RecordMetadata,
) ([]byte, error) {
	var builder strings.Builder
	builder.WriteString(record.Level.String())
	builder.WriteString("|")
	builder.WriteString(record.Message)

	for _, attr := range attrs {
		builder.WriteString("|")
		builder.WriteString(attr.Key)
		builder.WriteString("=")
		builder.WriteString(attr.Value.Resolve().String())
	}

	builder.WriteString("\n")
	return []byte(builder.String()), nil
}

func TestHandlerEnabled(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		level slog.Level
		want  bool
	}{
		{name: "filters info", level: slog.LevelInfo, want: false},
		{name: "accepts error", level: slog.LevelError, want: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			handler := newHandlerFixture(t, slog.LevelWarn)
			if got := handler.Enabled(context.Background(), testCase.level); got != testCase.want {
				t.Fatalf("Enabled(%v) = %v, want %v", testCase.level, got, testCase.want)
			}
		})
	}
}

func TestLoggerWithInheritsAttributes(t *testing.T) {
	t.Parallel()

	base, buffer := newLoggerFixture(t)
	base.With("component", "poller").Info("sample", "step", "collect")

	output := buffer.String()
	if !strings.Contains(output, "sample|component=poller|step=collect") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestLoggerWithGroupPrefixesAttributes(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	handler := newHandlerFixtureWithBuffer(t, slog.LevelDebug, buffer)
	slog.New(handler).WithGroup("metrics").Info("sample", "cpu", "ok")

	output := buffer.String()
	if !strings.Contains(output, "metrics.cpu=ok") {
		t.Fatalf("unexpected grouped output: %q", output)
	}
}

func newHandlerFixture(t *testing.T, minimumLevel slog.Level) *sharedlogger.Handler {
	t.Helper()

	return newHandlerFixtureWithBuffer(t, minimumLevel, &bytes.Buffer{})
}

func newHandlerFixtureWithBuffer(t *testing.T, minimumLevel slog.Level, buffer *bytes.Buffer) *sharedlogger.Handler {
	t.Helper()

	handler, err := sharedlogger.NewHandler(sharedlogger.HandlerConfig{
		Writer:       buffer,
		Formatter:    stubFormatter{},
		MinimumLevel: minimumLevel,
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	return handler
}

func newLoggerFixture(t *testing.T) (sharedlogger.Logger, *bytes.Buffer) {
	t.Helper()

	buffer := &bytes.Buffer{}
	base, err := sharedlogger.New(sharedlogger.Config{
		ServiceName:  "system-metrics",
		Hostname:     "rpi",
		Writer:       buffer,
		Formatter:    stubFormatter{},
		MinimumLevel: slog.LevelDebug,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return base, buffer
}
