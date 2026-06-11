package config_test

import (
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
)

func TestLoadSharedEmailConfigParsedFields(t *testing.T) {
	t.Parallel()

	cfg := mustLoadSharedEmailConfig(t, sharedEmailConfigFixture())

	if cfg.Messaging.ClientName != "system-email-notifier" {
		t.Fatalf("cfg.Messaging.ClientName = %q, want %q", cfg.Messaging.ClientName, "system-email-notifier")
	}

	if cfg.Email.From != "system-alert-notifier@lite-nas.com" {
		t.Fatalf("cfg.Email.From = %q, want %q", cfg.Email.From, "system-alert-notifier@lite-nas.com")
	}

	if cfg.SMTP.Timeout != 10*time.Second {
		t.Fatalf("cfg.SMTP.Timeout = %s, want %s", cfg.SMTP.Timeout, 10*time.Second)
	}

	if cfg.Logging.FilePath != "/var/log/lite-nas/system-email-notifier.log" {
		t.Fatalf("cfg.Logging.FilePath = %q, want %q", cfg.Logging.FilePath, "/var/log/lite-nas/system-email-notifier.log")
	}
}

func mustLoadSharedEmailConfig(t *testing.T, raw string) config.SharedEmailConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(raw))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadSharedEmailConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadSharedEmailConfig() error = %v", err)
	}

	return cfg
}

func sharedEmailConfigFixture() string {
	return "[messaging]\n" +
		"url=tls://127.0.0.1:4222\n" +
		"client_name=system-email-notifier\n" +
		"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/transport/lite-nas-sys-email-notifier/client.crt\n" +
		"key=/etc/lite-nas/certificates/transport/lite-nas-sys-email-notifier/client.key\n" +
		"timeout=5s\n" +
		"[email]\n" +
		"to=a@example.com\n" +
		"cc=b@example.com\n" +
		"from=system-alert-notifier@lite-nas.com\n" +
		"subject_prefix=[LiteNAS]\n" +
		"[smtp]\n" +
		"host=127.0.0.1\n" +
		"port=25\n" +
		"timeout=10s\n" +
		"helo=localhost\n" +
		"[logging]\n" +
		"level=warn\n" +
		"format=rfc5424\n" +
		"output=file\n" +
		"file_path=/var/log/lite-nas/system-email-notifier.log\n"
}
