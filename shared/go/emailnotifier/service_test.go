package emailnotifier

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedlogger "lite-nas/shared/logger"
)

func TestNewInfraModuleRejectsMissingConfigFile(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/nonexistent/email-notifier.conf", "email-notifier")
	if err == nil {
		t.Fatal("expected config file error")
	}
}

func TestNewInfraModuleRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	configPath := writeConfigFixture(t, "[messaging\n")

	_, err := NewInfraModule(configPath, "email-notifier")
	if err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestBuildWorkerRuntimeBuildsWorkerWithValidatedDependencies(t *testing.T) {
	t.Parallel()

	config := buildWorkerConfig(t)
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatalf("os.Hostname() error = %v", err)
	}

	validate, input, worker, err := BuildWorkerRuntime(
		config.TemplatesPath,
		config.Email,
		config.SMTP,
		3,
	)
	if err != nil {
		t.Fatalf("BuildWorkerRuntime() error = %v", err)
	}

	assertBuiltWorkerRuntime(t, validate, input, worker, hostname, config.TemplatesPath)
}

func TestRunWorkerLogsShutdownOnCleanExit(t *testing.T) {
	t.Parallel()

	input := make(chan loggingmanagercontract.AlertPayload)
	close(input)
	worker := mustNewWorker(t, buildWorkerConfig(t), input)
	logger := &recordingLogger{}

	if err := RunWorker(context.Background(), worker, logger, "service stopping"); err != nil {
		t.Fatalf("RunWorker() error = %v", err)
	}
	if len(logger.infoMessages) != 1 || logger.infoMessages[0] != "service stopping" {
		t.Fatalf("shutdown log = %#v, want [\"service stopping\"]", logger.infoMessages)
	}
}

func TestRunWorkerLogsShutdownOnCanceledContext(t *testing.T) {
	t.Parallel()

	worker := mustNewWorker(t, buildWorkerConfig(t), make(chan loggingmanagercontract.AlertPayload))
	logger := &recordingLogger{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RunWorker(ctx, worker, logger, "service stopping")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunWorker() error = %v, want %v", err, context.Canceled)
	}
	if len(logger.infoMessages) != 1 || logger.infoMessages[0] != "service stopping" {
		t.Fatalf("shutdown log = %#v, want [\"service stopping\"]", logger.infoMessages)
	}
}

func TestRunWorkerDoesNotLogShutdownOnWorkerError(t *testing.T) {
	t.Parallel()

	input := make(chan loggingmanagercontract.AlertPayload, 1)
	worker := mustNewWorker(t, buildWorkerConfig(t), input)
	expectedErr := errors.New("smtp failed")
	worker.send = func(context.Context, smtpRequest) error {
		return expectedErr
	}
	logger := &recordingLogger{}
	input <- buildAlertPayload()

	err := RunWorker(context.Background(), worker, logger, "service stopping")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("RunWorker() error = %v, want %v", err, expectedErr)
	}
	if len(logger.infoMessages) != 0 {
		t.Fatalf("shutdown log = %#v, want none", logger.infoMessages)
	}
}

type recordingLogger struct {
	infoMessages []string
}

func (l *recordingLogger) Debug(string, ...any) {}

func (l *recordingLogger) Info(msg string, _ ...any) {
	l.infoMessages = append(l.infoMessages, msg)
}

func (l *recordingLogger) Warn(string, ...any) {}

func (l *recordingLogger) Error(string, ...any) {}

func (l *recordingLogger) With(...any) sharedlogger.Logger {
	return l
}

func assertBuiltWorkerRuntime(
	t *testing.T,
	validate any,
	input chan loggingmanagercontract.AlertPayload,
	worker Worker,
	wantHostname string,
	wantTemplatesPath string,
) {
	t.Helper()

	if validate == nil {
		t.Fatal("expected validator")
	}
	if cap(input) != 3 {
		t.Fatalf("cap(input) = %d, want 3", cap(input))
	}
	if worker.config.Hostname != wantHostname {
		t.Fatalf("worker hostname = %q, want %q", worker.config.Hostname, wantHostname)
	}
	if worker.config.TemplatesPath != wantTemplatesPath {
		t.Fatalf("worker templates path = %q, want %q", worker.config.TemplatesPath, wantTemplatesPath)
	}
}

func writeConfigFixture(t *testing.T, content string) string {
	t.Helper()

	configPath := t.TempDir() + "/service.conf"
	if err := os.WriteFile(configPath, []byte(strings.TrimSpace(content)), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	return configPath
}

func buildServiceConfigFixture() string {
	return strings.Join([]string{
		"[messaging]",
		"url=nats://127.0.0.1:4222",
		"timeout=5s",
		"",
		"[logging]",
		"output=stdout",
		"format=text",
		"level=info",
		"",
		"[email]",
		"to=ops@example.com",
		"from=system-alert-notifier@lite-nas.com",
		"",
		"[smtp]",
		"host=127.0.0.1",
		"port=25",
		"timeout=10s",
		"helo=localhost",
		"",
	}, "\n")
}

func TestRunServiceRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	configPath := writeConfigFixture(t, "[messaging\n")

	err := RunService(context.Background(), ServiceRuntimeConfig{
		ConfigPath:      configPath,
		ServiceName:     "email-notifier",
		TemplatesPath:   t.TempDir(),
		AlertSubject:    "subject.alert",
		StartupMessage:  "started",
		ShutdownMessage: "stopping",
		InputBufferSize: 1,
	})
	if err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestRunServiceAttemptsSharedInfraSetup(t *testing.T) {
	t.Parallel()

	configPath := writeConfigFixture(t, buildServiceConfigFixture())

	err := RunService(context.Background(), ServiceRuntimeConfig{
		ConfigPath:      configPath,
		ServiceName:     "email-notifier",
		TemplatesPath:   mustWriteTemplateFixture(t),
		AlertSubject:    "subject.alert",
		StartupMessage:  "started",
		ShutdownMessage: "stopping",
		InputBufferSize: 1,
	})
	if err == nil {
		t.Fatal("expected messaging setup error")
	}
}

var _ sharedlogger.Logger = (*recordingLogger)(nil)
