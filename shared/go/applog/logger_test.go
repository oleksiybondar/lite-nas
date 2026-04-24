package applog

import (
	"os"
	"path/filepath"
	"testing"

	sharedconfig "lite-nas/shared/config"
)

func TestBuildMessageIDFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normalizes hyphen", input: "system-metrics", want: "SYSTEM_METRICS"},
		{name: "trims whitespace", input: " app ", want: "APP"},
		{name: "falls back for empty", input: "", want: "APP"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := buildMessageID(testCase.input); got != testCase.want {
				t.Fatalf("buildMessageID() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestMapLogLevelFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "debug", input: "debug", want: "DEBUG"},
		{name: "warn", input: "warn", want: "WARN"},
		{name: "error", input: "error", want: "ERROR"},
		{name: "default info", input: "info", want: "INFO"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := mapLogLevel(testCase.input).String(); got != testCase.want {
				t.Fatalf("mapLogLevel() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestBuildLogWriterStdoutOutput(t *testing.T) {
	t.Parallel()

	writer, cleanup, err := buildLogWriter(sharedconfig.LoggingConfig{Output: "stdout"})
	if err != nil {
		t.Fatalf("buildLogWriter() error = %v", err)
	}
	defer cleanup()

	if writer != os.Stdout {
		t.Fatalf("buildLogWriter() = %#v, want %#v", writer, os.Stdout)
	}
}

func TestBuildLogWriterStderrOutput(t *testing.T) {
	t.Parallel()

	writer, cleanup, err := buildLogWriter(sharedconfig.LoggingConfig{Output: "stderr"})
	if err != nil {
		t.Fatalf("buildLogWriter() error = %v", err)
	}
	defer cleanup()

	if writer != os.Stderr {
		t.Fatalf("buildLogWriter() = %#v, want %#v", writer, os.Stderr)
	}
}

func TestBuildLogWriterFileOutput(t *testing.T) {
	baseDir := t.TempDir()
	restoreAllowedLogDir := allowedLogDir
	allowedLogDir = baseDir
	defer func() {
		allowedLogDir = restoreAllowedLogDir
	}()

	logPath := filepath.Join(baseDir, "app.log")
	writer, cleanup, err := buildLogWriter(sharedconfig.LoggingConfig{
		Output:   "file",
		FilePath: logPath,
	})
	if err != nil {
		t.Fatalf("buildLogWriter() error = %v", err)
	}
	defer cleanup()

	if writer == nil {
		t.Fatal("expected file writer")
	}
}

func TestBuildLogWriterRejectsPathOutsideAllowedDirectory(t *testing.T) {
	baseDir := t.TempDir()
	restoreAllowedLogDir := allowedLogDir
	allowedLogDir = baseDir
	defer func() {
		allowedLogDir = restoreAllowedLogDir
	}()

	_, cleanup, err := buildLogWriter(sharedconfig.LoggingConfig{
		Output:   "file",
		FilePath: filepath.Join(filepath.Dir(baseDir), "app.log"),
	})
	if cleanup != nil {
		defer cleanup()
	}

	if err == nil {
		t.Fatal("expected path validation error")
	}
}

func TestNewAppLoggerCreatesLogger(t *testing.T) {
	t.Parallel()

	log, cleanup, err := NewAppLogger("system-metrics", sharedconfig.LoggingConfig{
		Level:  "info",
		Format: "rfc5424",
		Output: "stdout",
	})
	if err != nil {
		t.Fatalf("NewAppLogger() error = %v", err)
	}
	defer cleanup()

	if log == nil {
		t.Fatal("expected logger")
	}
}
