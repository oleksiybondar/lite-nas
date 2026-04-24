package logger_test

import (
	"bytes"
	"testing"

	sharedlogger "lite-nas/shared/logger"
)

func TestNewNopReturnsUsableLogger(t *testing.T) {
	t.Parallel()

	logger := sharedlogger.NewNop()
	logger.Debug("debug", "key", "value")
	logger.Info("info", "key", "value")
	logger.Warn("warn", "key", "value")
	logger.Error("error", "key", "value")

	child := logger.With("component", "poller")
	if _, ok := child.(sharedlogger.NopLogger); !ok {
		t.Fatalf("With() returned %T, want logger.NopLogger", child)
	}
}

func TestNewRejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cfg  sharedlogger.Config
	}{
		{name: "missing writer", cfg: sharedlogger.Config{}},
		{name: "missing formatter", cfg: sharedlogger.Config{Writer: &bytes.Buffer{}}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if _, err := sharedlogger.New(testCase.cfg); err == nil {
				t.Fatal("expected constructor error")
			}
		})
	}
}

func TestLoggerEmitsLevels(t *testing.T) {
	t.Parallel()

	assertLoggerOutputContains(t, logDebug, "DEBUG|debug")
	assertLoggerOutputContains(t, logWarn, "WARN|warn")
	assertLoggerOutputContains(t, logError, "ERROR|error")
}

func assertLoggerOutputContains(t *testing.T, logFn func(sharedlogger.Logger), expected string) {
	t.Helper()

	var buffer bytes.Buffer
	logger, err := sharedlogger.New(sharedlogger.Config{
		ServiceName:  "system-metrics",
		Hostname:     "rpi",
		Writer:       &buffer,
		Formatter:    stubFormatter{},
		MinimumLevel: -8,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logFn(logger)
	if !bytes.Contains(buffer.Bytes(), []byte(expected)) {
		t.Fatalf("expected %q in output %q", expected, buffer.String())
	}
}

func logDebug(log sharedlogger.Logger) {
	log.Debug("debug")
}

func logWarn(log sharedlogger.Logger) {
	log.Warn("warn")
}

func logError(log sharedlogger.Logger) {
	log.Error("error")
}
