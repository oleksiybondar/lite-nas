// Package hostmanagertest supplies shared service-manager and session-manager fixtures.
package hostmanagertest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lite-nas/shared/config"
	"lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/testutil/authtokentest"
)

// MustConfig writes a valid manager config and token certificate into a test directory.
func MustConfig(t *testing.T) string {
	t.Helper()
	public, private := authtokentest.MustGenerateEd25519Key(t)
	directory := t.TempDir()
	certificatePath := filepath.Join(directory, "token.crt")
	certificate := authtokentest.MustCreateEd25519CertificatePEM(t, public, private)
	if err := os.WriteFile(certificatePath, certificate, 0o600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(directory, "manager.conf")
	config := fmt.Sprintf("[messaging]\nurl=nats://127.0.0.1:4222\nclient_name=test\n[auth_tokens]\nissuer=test\naudience=test\nverification_cert=%s\n[logging]\nlevel=warn\nformat=rfc5424\noutput=stdout\n", certificatePath)
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatal(err)
	}
	return configPath
}

// MalformedPayloads exercises JSON framing, token validation, and size limits.
func MalformedPayloads(token string) map[string][]byte {
	return map[string][]byte{
		"unknown field": []byte(`{"access_token":"` + token + `","extra":true}`),
		"trailing JSON": []byte(`{"access_token":"` + token + `"}{}`),
		"invalid token": []byte(`{"access_token":"invalid"}`),
		"oversized":     []byte(strings.Repeat("x", 16*1024+1)),
	}
}

// RecordingLogger stores only error messages relevant to lifecycle tests.
type RecordingLogger struct{ Errors []string }

// Debug ignores debug entries in manager lifecycle tests.
func (*RecordingLogger) Debug(string, ...any) {}

// Info ignores startup entries in manager lifecycle tests.
func (*RecordingLogger) Info(string, ...any) {}

// Warn ignores warning entries in manager lifecycle tests.
func (*RecordingLogger) Warn(string, ...any) {}

// Error records a failed drain without losing its observable path.
func (l *RecordingLogger) Error(message string, _ ...any) { l.Errors = append(l.Errors, message) }

// With retains the same logger for structured lifecycle messages.
func (l *RecordingLogger) With(...any) logger.Logger { return l }

// InfraFactory injects an in-memory server and logger without opening NATS.
func InfraFactory(server messaging.Server, log logger.Logger) modules.CoreInfraConstructor {
	return func(_ string, _ config.LoggingConfig, _ config.MessagingConfig) (modules.CoreInfra, error) {
		return modules.CoreInfra{Server: server, Logger: log}, nil
	}
}
