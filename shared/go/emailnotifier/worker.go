package emailnotifier

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/smtp"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sharedconfig "lite-nas/shared/config"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

const alertTemplateFileName = "alert.html"

var (
	errNilInputChannel      = errors.New("email notifier input channel is required")
	errEmptyHostname        = errors.New("email notifier hostname is required")
	errEmptyTemplatesPath   = errors.New("email notifier templates path is required")
	errMissingSenderAddress = errors.New("email notifier sender address is required")
)

// WorkerConfig defines runtime inputs for one email notifier worker instance.
//
// Contract:
//   - Hostname is rendered into the email body and subject.
//   - TemplatesPath points to the notifier-specific template directory.
//   - Email and SMTP settings are already parsed from service config and are
//     validated again here at the worker boundary.
type WorkerConfig struct {
	Hostname      string
	TemplatesPath string
	Email         sharedconfig.EmailConfig
	SMTP          sharedconfig.SMTPConfig
}

// Worker consumes alert payloads and delivers rendered emails through SMTP.
//
// Architectural role:
//   - The worker sits below transport adapters such as NATS subscriptions.
//   - It accepts already-decoded alert payloads and owns only rendering and
//     SMTP delivery.
type Worker struct {
	config WorkerConfig
	input  <-chan loggingmanagercontract.AlertPayload
	send   func(context.Context, smtpRequest) error
}

type smtpRequest struct {
	SMTP       sharedconfig.SMTPConfig
	From       string
	Recipients []string
	Message    []byte
}

type alertTemplateData struct {
	Hostname string
	Alert    alertTemplateAlert
}

type alertTemplateAlert struct {
	EventID       string
	Category      string
	Severity      string
	SeverityBG    string
	SeverityColor string
	Priority      string
	CreatedAt     string
	Source        string
	Message       string
	TriggerValue  string
}

// NewWorker validates configuration and constructs one notifier worker.
func NewWorker(
	config WorkerConfig,
	input <-chan loggingmanagercontract.AlertPayload,
) (Worker, error) {
	if input == nil {
		return Worker{}, errNilInputChannel
	}

	config.Hostname = strings.TrimSpace(config.Hostname)
	if config.Hostname == "" {
		return Worker{}, errEmptyHostname
	}

	config.TemplatesPath = strings.TrimSpace(config.TemplatesPath)
	if config.TemplatesPath == "" {
		return Worker{}, errEmptyTemplatesPath
	}

	config.Email.From = strings.TrimSpace(config.Email.From)
	if config.Email.From == "" {
		return Worker{}, errMissingSenderAddress
	}

	return Worker{
		config: config,
		input:  input,
		send:   sendSMTPMessage,
	}, nil
}

// Run processes alert payloads until context cancellation or channel close.
func (worker Worker) Run(ctx context.Context) error {
	for {
		shouldStop, err := worker.runOnce(ctx)
		if err != nil {
			return err
		}
		if shouldStop {
			return nil
		}
	}
}

// runOnce executes one worker loop iteration and reports whether processing should stop.
func (worker Worker) runOnce(ctx context.Context) (bool, error) {
	alert, ok, err := worker.nextAlert(ctx)
	if err != nil {
		return false, err
	}
	if !ok {
		return true, nil
	}

	return false, worker.processAlert(ctx, alert)
}

// nextAlert waits for either context cancellation or the next alert payload.
func (worker Worker) nextAlert(
	ctx context.Context,
) (loggingmanagercontract.AlertPayload, bool, error) {
	select {
	case <-ctx.Done():
		return loggingmanagercontract.AlertPayload{}, false, ctx.Err()
	case alert, ok := <-worker.input:
		return alert, ok, nil
	}
}

// processAlert renders and sends one email for the provided alert payload.
func (worker Worker) processAlert(ctx context.Context, alert loggingmanagercontract.AlertPayload) error {
	recipients := worker.recipientList()
	if len(recipients) == 0 {
		return nil
	}

	templateName := templateNameForAlert(alert)
	htmlBody, err := worker.renderTemplate(templateName, alert)
	if err != nil {
		return err
	}

	subject := worker.subjectForAlert(alert)
	message := buildEmailMessage(worker.config.Email, recipients, subject, htmlBody)

	return worker.send(ctx, smtpRequest{
		SMTP:       worker.config.SMTP,
		From:       worker.config.Email.From,
		Recipients: recipients,
		Message:    message,
	})
}

// recipientList returns all SMTP recipients in transport order.
func (worker Worker) recipientList() []string {
	recipients := make([]string, 0, len(worker.config.Email.To)+len(worker.config.Email.CC))
	recipients = append(recipients, worker.config.Email.To...)
	recipients = append(recipients, worker.config.Email.CC...)
	return recipients
}

// renderTemplate renders one notifier template against one alert payload.
func (worker Worker) renderTemplate(
	templateName string,
	alert loggingmanagercontract.AlertPayload,
) (string, error) {
	templatePath := filepath.Join(worker.config.TemplatesPath, templateName)
	parsedTemplate, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var rendered bytes.Buffer
	if err = parsedTemplate.Execute(&rendered, buildAlertTemplateData(worker.config.Hostname, alert)); err != nil {
		return "", err
	}

	return rendered.String(), nil
}

// subjectForAlert constructs one human-readable email subject line.
func (worker Worker) subjectForAlert(alert loggingmanagercontract.AlertPayload) string {
	baseSubject := fmt.Sprintf(
		"%s alert on %s: %s",
		strings.ToUpper(normalizeSeverity(alert.Severity)),
		worker.config.Hostname,
		strings.TrimSpace(alert.Category),
	)

	prefix := strings.TrimSpace(worker.config.Email.SubjectPrefix)
	if prefix == "" {
		return baseSubject
	}

	return fmt.Sprintf("%s %s", prefix, baseSubject)
}

// templateNameForAlert chooses the renderer template for one alert payload.
func templateNameForAlert(_ loggingmanagercontract.AlertPayload) string {
	return alertTemplateFileName
}

// buildAlertTemplateData maps one alert payload into the template view model.
func buildAlertTemplateData(
	hostname string,
	alert loggingmanagercontract.AlertPayload,
) alertTemplateData {
	return alertTemplateData{
		Hostname: strings.TrimSpace(hostname),
		Alert: alertTemplateAlert{
			EventID:       strings.TrimSpace(alert.EventID),
			Category:      strings.TrimSpace(alert.Category),
			Severity:      normalizeSeverity(alert.Severity),
			SeverityBG:    severityBackground(alert.Severity),
			SeverityColor: severityForeground(alert.Severity),
			Priority:      formatPriority(alert.Priority),
			CreatedAt:     strings.TrimSpace(alert.CreatedAt),
			Source:        strings.TrimSpace(alert.Source),
			Message:       strings.TrimSpace(alert.Message),
			TriggerValue:  strings.TrimSpace(alert.TriggerValue),
		},
	}
}

// normalizeSeverity collapses empty or unknown values into a stable template value.
func normalizeSeverity(severity any) string {
	normalized := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", severity)))
	switch normalized {
	case "info", "warning", "error", "critical":
		return normalized
	default:
		return "info"
	}
}

// severityBackground returns one email-safe header background color.
func severityBackground(severity any) string {
	switch normalizeSeverity(severity) {
	case "warning":
		return "#3d2807"
	case "error":
		return "#3f1517"
	case "critical":
		return "#321641"
	default:
		return "#0f2740"
	}
}

// severityForeground returns one email-safe severity accent color.
func severityForeground(severity any) string {
	switch normalizeSeverity(severity) {
	case "warning":
		return "#fbbf24"
	case "error":
		return "#f87171"
	case "critical":
		return "#c084fc"
	default:
		return "#90caf9"
	}
}

// formatPriority renders one optional priority into a stable display value.
func formatPriority(priority *int) string {
	if priority == nil {
		return "-"
	}

	return "P" + strconv.Itoa(*priority)
}

// buildEmailMessage formats one HTML email with transport headers.
func buildEmailMessage(
	config sharedconfig.EmailConfig,
	recipients []string,
	subject string,
	htmlBody string,
) []byte {
	buffer := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buffer)

	_, _ = fmt.Fprintf(writer, "From: %s\r\n", config.From)
	if len(config.To) > 0 {
		_, _ = fmt.Fprintf(writer, "To: %s\r\n", strings.Join(config.To, ", "))
	}
	if len(config.CC) > 0 {
		_, _ = fmt.Fprintf(writer, "Cc: %s\r\n", strings.Join(config.CC, ", "))
	}
	_, _ = fmt.Fprintf(writer, "Subject: %s\r\n", subject)
	_, _ = fmt.Fprintf(writer, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	_, _ = fmt.Fprintf(writer, "MIME-Version: 1.0\r\n")
	_, _ = fmt.Fprintf(writer, "Content-Type: text/html; charset=UTF-8\r\n")
	_, _ = fmt.Fprintf(writer, "Content-Transfer-Encoding: 8bit\r\n")
	_, _ = fmt.Fprintf(writer, "X-LiteNAS-Recipients: %s\r\n", strings.Join(recipients, ","))
	_, _ = fmt.Fprintf(writer, "\r\n%s", htmlBody)
	_ = writer.Flush()

	return buffer.Bytes()
}

// sendSMTPMessage delivers one email through the configured SMTP endpoint.
func sendSMTPMessage(ctx context.Context, request smtpRequest) error {
	connection, err := dialSMTPConnection(ctx, request)
	if err != nil {
		return err
	}
	defer func() { _ = connection.Close() }()

	client, err := newSMTPClient(connection, request)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	if err = greetSMTPClient(client, request); err != nil {
		return err
	}

	if err = addSMTPRecipients(client, request); err != nil {
		return err
	}

	if err = writeSMTPData(client, request.Message); err != nil {
		return err
	}

	return client.Quit()
}

// dialSMTPConnection establishes one outbound SMTP TCP connection.
func dialSMTPConnection(ctx context.Context, request smtpRequest) (net.Conn, error) {
	address := net.JoinHostPort(request.SMTP.Host, strconv.Itoa(request.SMTP.Port))
	dialer := net.Dialer{Timeout: request.SMTP.Timeout}
	connection, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}

	if request.SMTP.Timeout > 0 {
		_ = connection.SetDeadline(time.Now().Add(request.SMTP.Timeout))
	}

	return connection, nil
}

// newSMTPClient wraps one connection in an SMTP client.
func newSMTPClient(connection net.Conn, request smtpRequest) (*smtp.Client, error) {
	return smtp.NewClient(connection, request.SMTP.Host)
}

// greetSMTPClient sends the configured HELO when one is configured.
func greetSMTPClient(client *smtp.Client, request smtpRequest) error {
	if request.SMTP.HELO == "" {
		return nil
	}

	return client.Hello(request.SMTP.HELO)
}

// addSMTPRecipients sends envelope sender and recipients to the SMTP server.
func addSMTPRecipients(client *smtp.Client, request smtpRequest) error {
	if err := client.Mail(request.From); err != nil {
		return err
	}

	for _, recipient := range request.Recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	return nil
}

// writeSMTPData streams the RFC 5322 message body through the SMTP DATA command.
func writeSMTPData(client *smtp.Client, message []byte) error {
	dataWriter, err := client.Data()
	if err != nil {
		return err
	}

	if _, err = dataWriter.Write(message); err != nil {
		_ = dataWriter.Close()
		return err
	}

	return dataWriter.Close()
}
