package config

import (
	"testing"
	"time"

	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
)

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, LoadConfig)
}

func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	cfg := loadConfigFixture(t)
	assertMessagingConfig(t, cfg)
}

func TestLoadConfigRulesFields(t *testing.T) {
	t.Parallel()

	cfg := loadConfigFixture(t)
	if len(cfg.Rules.Files) != 2 {
		t.Fatalf("len(cfg.Rules.Files) = %d, want 2", len(cfg.Rules.Files))
	}
	if cfg.Rules.Files[0] != "/etc/lite-nas/resources-monitor/rules/system-metrics.json" {
		t.Fatalf("cfg.Rules.Files[0] = %q", cfg.Rules.Files[0])
	}
	if cfg.Rules.Files[1] != "/etc/lite-nas/resources-monitor/rules/network-metrics.json" {
		t.Fatalf("cfg.Rules.Files[1] = %q", cfg.Rules.Files[1])
	}
}

func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	cfg := loadConfigFixture(t)
	assertLoggingConfig(t, cfg)
}

func TestLoadConfigAuthFields(t *testing.T) {
	t.Parallel()

	cfg := loadConfigFixture(t)
	if cfg.Auth.CA != "/etc/lite-nas/certificates/identities/root-ca.crt" {
		t.Fatalf("cfg.Auth.CA = %q", cfg.Auth.CA)
	}
	if cfg.Auth.Cert != "/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt" {
		t.Fatalf("cfg.Auth.Cert = %q", cfg.Auth.Cert)
	}
	if cfg.Auth.Key != "/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.key" {
		t.Fatalf("cfg.Auth.Key = %q", cfg.Auth.Key)
	}
	if cfg.Auth.ServiceName != "resources-monitor" {
		t.Fatalf("cfg.Auth.ServiceName = %q", cfg.Auth.ServiceName)
	}
	if cfg.Auth.ServiceLogin != "lite-nas-resources-monitor" {
		t.Fatalf("cfg.Auth.ServiceLogin = %q", cfg.Auth.ServiceLogin)
	}
}

func TestLoadConfigRejectsMissingRulesFiles(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		LoadConfig,
		"[messaging]\n"+
			"url=nats://localhost:4222\n"+
			"[logging]\n"+
			"level=info\n"+
			"format=rfc5424\n"+
			"output=stdout\n"+
			"[rules]\n",
	)
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{
		Data: []byte(
			"[rules]\n" +
				"files=/etc/lite-nas/resources-monitor/rules/system-metrics.json,/etc/lite-nas/resources-monitor/rules/network-metrics.json\n" +
				"[messaging]\n" +
				"url=nats://localhost:4222\n" +
				"client_name=resources-monitor\n" +
				"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.crt\n" +
				"key=/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.key\n" +
				"timeout=9s\n" +
				"[auth]\n" +
				"ca=/etc/lite-nas/certificates/identities/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt\n" +
				"key=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.key\n" +
				"service_name=resources-monitor\n" +
				"service_login=lite-nas-resources-monitor\n" +
				"[logging]\n" +
				"level=debug\n" +
				"format=rfc5424\n" +
				"output=file\n" +
				"file_path=/var/log/lite-nas/resources-monitor.log\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}

func assertMessagingConfig(t *testing.T, cfg Config) {
	t.Helper()

	checks := []struct {
		name string
		got  any
		want any
	}{
		{name: "url", got: cfg.Messaging.URL, want: "nats://localhost:4222"},
		{name: "client_name", got: cfg.Messaging.ClientName, want: "resources-monitor"},
		{name: "ca", got: cfg.Messaging.CA, want: "/etc/lite-nas/certificates/transport/root-ca.crt"},
		{name: "cert", got: cfg.Messaging.Cert, want: "/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.crt"},
		{name: "key", got: cfg.Messaging.Key, want: "/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.key"},
		{name: "timeout", got: cfg.Messaging.Timeout, want: 9 * time.Second},
	}

	for _, check := range checks {
		if check.got != check.want {
			t.Fatalf("cfg.Messaging.%s = %v, want %v", check.name, check.got, check.want)
		}
	}
}

func assertLoggingConfig(t *testing.T, cfg Config) {
	t.Helper()

	if cfg.Logging.Level != "debug" {
		t.Fatalf("cfg.Logging.Level = %q", cfg.Logging.Level)
	}
	if cfg.Logging.Output != "file" {
		t.Fatalf("cfg.Logging.Output = %q", cfg.Logging.Output)
	}
	if cfg.Logging.FilePath != "/var/log/lite-nas/resources-monitor.log" {
		t.Fatalf("cfg.Logging.FilePath = %q", cfg.Logging.FilePath)
	}
}
