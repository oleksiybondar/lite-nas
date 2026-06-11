package emailnotifier

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	sharedconfig "lite-nas/shared/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
)

func TestNewWorkerRejectsNilInputChannel(t *testing.T) {
	t.Parallel()

	_, err := NewWorker(buildWorkerConfig(t), nil)
	if !errors.Is(err, errNilInputChannel) {
		t.Fatalf("NewWorker() error = %v, want %v", err, errNilInputChannel)
	}
}

func TestRunReturnsNilWhenInputChannelCloses(t *testing.T) {
	t.Parallel()

	input := make(chan loggingmanagercontract.AlertPayload)
	worker := mustNewWorker(t, buildWorkerConfig(t), input)
	close(input)

	if err := worker.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestProcessAlertSkipsDeliveryWhenRecipientsEmpty(t *testing.T) {
	t.Parallel()

	cfg := buildWorkerConfig(t)
	cfg.Email.To = nil
	cfg.Email.CC = nil

	worker := mustNewWorker(t, cfg, make(chan loggingmanagercontract.AlertPayload))
	deliveries := 0
	worker.send = func(_ context.Context, _ smtpRequest) error {
		deliveries++
		return nil
	}

	if err := worker.processAlert(context.Background(), buildAlertPayload()); err != nil {
		t.Fatalf("processAlert() error = %v", err)
	}

	if deliveries != 0 {
		t.Fatalf("deliveries = %d, want 0", deliveries)
	}
}

func TestProcessAlertDeliversRenderedHTML(t *testing.T) {
	t.Parallel()

	worker := mustNewWorker(t, buildWorkerConfig(t), make(chan loggingmanagercontract.AlertPayload))
	var sentRequest smtpRequest
	worker.send = func(_ context.Context, request smtpRequest) error {
		sentRequest = request
		return nil
	}

	if err := worker.processAlert(context.Background(), buildAlertPayload()); err != nil {
		t.Fatalf("processAlert() error = %v", err)
	}

	message := string(sentRequest.Message)
	assertContains(t, message, "Subject: [LiteNAS] WARNING alert on lite-nas-host: system.metrics.mem.used")
	assertContains(t, message, "Envelope-Source")
	assertContains(t, message, "RAM usage is above threshold")
	assertContains(t, message, "93")
	assertContains(t, message, "bgcolor=\"#3d2807\"")
	assertContains(t, message, "color:#fbbf24")
	if sentRequest.From != "system-alert-notifier@lite-nas.com" {
		t.Fatalf("sentRequest.From = %q, want %q", sentRequest.From, "system-alert-notifier@lite-nas.com")
	}
	if len(sentRequest.Recipients) != 2 {
		t.Fatalf("recipient count = %d, want 2", len(sentRequest.Recipients))
	}
}

func TestBuildAlertTemplateDataNormalizesPriorityAndSeverity(t *testing.T) {
	t.Parallel()

	data := buildAlertTemplateData("host-1", loggingmanagercontract.AlertPayload{
		Category: "system.metrics.mem.used",
		Severity: "",
	})

	if data.Alert.Severity != "info" {
		t.Fatalf("data.Alert.Severity = %q, want info", data.Alert.Severity)
	}
	if data.Alert.Priority != "-" {
		t.Fatalf("data.Alert.Priority = %q, want -", data.Alert.Priority)
	}
	if data.Alert.SeverityBG != "#0f2740" {
		t.Fatalf("data.Alert.SeverityBG = %q, want %q", data.Alert.SeverityBG, "#0f2740")
	}
	if data.Alert.SeverityColor != "#90caf9" {
		t.Fatalf("data.Alert.SeverityColor = %q, want %q", data.Alert.SeverityColor, "#90caf9")
	}
}

func TestSendSMTPMessageDeliversMessage(t *testing.T) {
	t.Parallel()

	server := mustStartSMTPTestServer(t)

	err := sendSMTPMessage(context.Background(), smtpRequest{
		SMTP: sharedconfig.SMTPConfig{
			Host:    server.host,
			Port:    server.port,
			Timeout: 10 * time.Second,
			HELO:    "localhost",
		},
		From:       "system-alert-notifier@lite-nas.com",
		Recipients: []string{"ops@example.com", "audit@example.com"},
		Message:    []byte("Subject: [LiteNAS] test\r\n\r\nhello\r\n"),
	})
	if err != nil {
		t.Fatalf("sendSMTPMessage() error = %v", err)
	}

	commands, message := server.waitForDelivery(t)
	assertContains(t, strings.Join(commands, "\n"), "EHLO localhost")
	assertContains(t, strings.Join(commands, "\n"), "MAIL FROM:<system-alert-notifier@lite-nas.com>")
	assertContains(t, strings.Join(commands, "\n"), "RCPT TO:<ops@example.com>")
	assertContains(t, strings.Join(commands, "\n"), "RCPT TO:<audit@example.com>")
	assertContains(t, message, "Subject: [LiteNAS] test")
	assertContains(t, message, "hello")
}

func buildWorkerConfig(t *testing.T) WorkerConfig {
	t.Helper()

	return WorkerConfig{
		Hostname:      "lite-nas-host",
		TemplatesPath: mustWriteTemplateFixture(t),
		Email: sharedconfig.EmailConfig{
			To:            []string{"ops@example.com"},
			CC:            []string{"audit@example.com"},
			From:          "system-alert-notifier@lite-nas.com",
			SubjectPrefix: "[LiteNAS]",
		},
		SMTP: sharedconfig.SMTPConfig{
			Host:    "127.0.0.1",
			Port:    25,
			Timeout: 10 * time.Second,
			HELO:    "localhost",
		},
	}
}

func mustWriteTemplateFixture(t *testing.T) string {
	t.Helper()

	templateDir := t.TempDir()
	templatePath := filepath.Join(templateDir, alertTemplateFileName)
	templateBody := `<html><body bgcolor="{{.Alert.SeverityBG}}" style="color:{{.Alert.SeverityColor}}"><span>{{.Hostname}}</span><span>Envelope-Source</span><span>{{.Alert.Message}}</span><span>{{.Alert.TriggerValue}}</span></body></html>`
	if err := os.WriteFile(templatePath, []byte(templateBody), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return templateDir
}

func buildAlertPayload() loggingmanagercontract.AlertPayload {
	priority := 2
	return loggingmanagercontract.AlertPayload{
		EventID:      "sysram_00000001",
		Category:     "system.metrics.mem.used",
		Severity:     sharedloggingenum.SeverityWarning,
		Priority:     &priority,
		CreatedAt:    "2026-05-31T10:00:00Z",
		Source:       "system-metrics",
		Message:      "RAM usage is above threshold",
		TriggerValue: "93",
	}
}

func mustNewWorker(
	t *testing.T,
	config WorkerConfig,
	input <-chan loggingmanagercontract.AlertPayload,
) Worker {
	t.Helper()

	worker, err := NewWorker(config, input)
	if err != nil {
		t.Fatalf("NewWorker() error = %v", err)
	}

	return worker
}

func assertContains(t *testing.T, actual string, expected string) {
	t.Helper()

	if !strings.Contains(actual, expected) {
		t.Fatalf("expected %q to contain %q", actual, expected)
	}
}

type smtpTestServer struct {
	host string
	port int

	listener net.Listener
	done     chan struct{}

	mu       sync.Mutex
	commands []string
	message  string
}

func mustStartSMTPTestServer(t *testing.T) *smtpTestServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := &smtpTestServer{
		host:     "127.0.0.1",
		port:     listener.Addr().(*net.TCPAddr).Port,
		listener: listener,
		done:     make(chan struct{}),
	}

	go server.serve()
	t.Cleanup(func() {
		_ = server.listener.Close()
	})

	return server
}

func (server *smtpTestServer) serve() {
	defer close(server.done)

	connection, err := server.listener.Accept()
	if err != nil {
		return
	}
	defer func() { _ = connection.Close() }()

	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)
	writeSMTPReply(writer, 220, "test-smtp")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		command := strings.TrimRight(line, "\r\n")
		if server.handleCommand(reader, writer, command) {
			return
		}
	}
}

func (server *smtpTestServer) handleCommand(
	reader *bufio.Reader,
	writer *bufio.Writer,
	command string,
) bool {
	server.recordCommand(command)

	switch {
	case strings.HasPrefix(command, "HELO "), strings.HasPrefix(command, "EHLO "):
		writeSMTPReply(writer, 250, "hello")
	case strings.HasPrefix(command, "MAIL FROM:"):
		writeSMTPReply(writer, 250, "sender ok")
	case strings.HasPrefix(command, "RCPT TO:"):
		writeSMTPReply(writer, 250, "recipient ok")
	case command == "DATA":
		writeSMTPReply(writer, 354, "end with <CR><LF>.<CR><LF>")
		server.recordMessage(readSMTPData(reader))
		writeSMTPReply(writer, 250, "queued")
	case command == "QUIT":
		writeSMTPReply(writer, 221, "bye")
		return true
	default:
		writeSMTPReply(writer, 250, "ok")
	}

	return false
}

func (server *smtpTestServer) waitForDelivery(t *testing.T) ([]string, string) {
	t.Helper()

	select {
	case <-server.done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SMTP delivery")
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	return append([]string(nil), server.commands...), server.message
}

func (server *smtpTestServer) recordCommand(command string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.commands = append(server.commands, command)
}

func (server *smtpTestServer) recordMessage(message string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.message = message
}

func writeSMTPReply(writer *bufio.Writer, code int, message string) {
	_, _ = fmt.Fprintf(writer, "%d %s\r\n", code, message)
	_ = writer.Flush()
}

func readSMTPData(reader *bufio.Reader) string {
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return strings.Join(lines, "")
		}
		if line == ".\r\n" {
			return strings.Join(lines, "")
		}
		lines = append(lines, line)
	}
}
