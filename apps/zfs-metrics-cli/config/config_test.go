package config

import (
	"testing"
	"time"

	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
)

func TestLoadConfigReturnsMessagingURL(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfig(t, validConfigFixture())
	if cfg.Messaging.URL != "nats://127.0.0.1:4222" {
		t.Fatalf("Messaging.URL = %q, want nats://127.0.0.1:4222", cfg.Messaging.URL)
	}
}

func TestLoadConfigReturnsMessagingClientName(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfig(t, validConfigFixture())
	if cfg.Messaging.ClientName != "zfs-metrics-cli" {
		t.Fatalf("Messaging.ClientName = %q, want zfs-metrics-cli", cfg.Messaging.ClientName)
	}
}

func TestLoadConfigReturnsMessagingTimeout(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfig(t, validConfigFixture())
	if cfg.Messaging.Timeout != 5*time.Second {
		t.Fatalf("Messaging.Timeout = %v, want 5s", cfg.Messaging.Timeout)
	}
}

func TestLoadConfigReturnsLoggingFilePath(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfig(t, validConfigFixture())
	if cfg.Logging.FilePath != "/var/log/lite-nas/zfs-metrics-cli.log" {
		t.Fatalf("Logging.FilePath = %q, want /var/log/lite-nas/zfs-metrics-cli.log", cfg.Logging.FilePath)
	}
}

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, LoadConfig)
}

func TestLoadConfigReturnsINIParseError(t *testing.T) {
	t.Parallel()

	configtest.RunINIParseErrorCase(t, LoadConfig)
}

func TestLoadConfigReturnsSharedConfigError(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		LoadConfig,
		"[messaging]\n"+
			"url=nats://127.0.0.1:4222\n"+
			"timeout=5s\n"+
			"[logging]\n"+
			"output=file\n",
	)
}

func mustLoadConfig(t *testing.T, configData string) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{Data: []byte(configData)})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}

func validConfigFixture() string {
	return "[messaging]\n" +
		"url=nats://127.0.0.1:4222\n" +
		"client_name=zfs-metrics-cli\n" +
		"timeout=5s\n" +
		"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics-cli/client.crt\n" +
		"key=/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics-cli/client.key\n" +
		"[logging]\n" +
		"level=info\n" +
		"format=rfc5424\n" +
		"output=file\n" +
		"file_path=/var/log/lite-nas/zfs-metrics-cli.log\n"
}
