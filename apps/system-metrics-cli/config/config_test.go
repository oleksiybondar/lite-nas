package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sharedfileio "lite-nas/shared/fileio"
	"lite-nas/shared/testutil/fileiotest"
)

func TestLoadConfigReturnsMessagingURL(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfigFromFixture(t)
	if cfg.Messaging.URL != "nats://127.0.0.1:4222" {
		t.Fatalf("Messaging.URL = %q, want nats://127.0.0.1:4222", cfg.Messaging.URL)
	}
}

func TestLoadConfigReturnsMessagingClientName(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfigFromFixture(t)
	if cfg.Messaging.ClientName != "system-metrics-cli" {
		t.Fatalf("Messaging.ClientName = %q, want system-metrics-cli", cfg.Messaging.ClientName)
	}
}

func TestLoadConfigReturnsMessagingTimeout(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfigFromFixture(t)
	if cfg.Messaging.Timeout != 5*time.Second {
		t.Fatalf("Messaging.Timeout = %v, want 5s", cfg.Messaging.Timeout)
	}
}

func TestLoadConfigReturnsLoggingOutput(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfigFromFixture(t)
	if cfg.Logging.Output != "stderr" {
		t.Fatalf("Logging.Output = %q, want stderr", cfg.Logging.Output)
	}
}

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")

	_, err := LoadConfig(fileiotest.Reader{Err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("LoadConfig() error = %v, want %v", err, wantErr)
	}
}

func TestLoadConfigReturnsINIParseError(t *testing.T) {
	t.Parallel()

	_, err := LoadConfig(fileiotest.Reader{Data: []byte("[messaging")})
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want parse error")
	}
}

func TestLoadConfigReturnsLoggingConfigError(t *testing.T) {
	t.Parallel()

	_, err := LoadConfig(fileiotest.Reader{
		Data: []byte(
			"[messaging]\n" +
				"url=nats://127.0.0.1:4222\n" +
				"timeout=5s\n" +
				"[logging]\n" +
				"output=file\n",
		),
	})
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want logging config error")
	}
}

func loadConfigFromFixture(t *testing.T) (Config, error) {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "system-metrics-cli.conf")
	configData := []byte(
		"[messaging]\n" +
			"url=nats://127.0.0.1:4222\n" +
			"client_name=system-metrics-cli\n" +
			"timeout=5s\n" +
			"ca=/etc/lite-nas/certificates/root-ca.crt\n" +
			"cert=/etc/lite-nas/certificates/lite-nas-system-metrics-cli/client.crt\n" +
			"key=/etc/lite-nas/certificates/lite-nas-system-metrics-cli/client.key\n" +
			"[logging]\n" +
			"level=info\n" +
			"format=rfc5424\n" +
			"output=stderr\n",
	)

	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		return Config{}, err
	}

	reader, err := sharedfileio.NewFileReader(configPath)
	if err != nil {
		return Config{}, err
	}

	return LoadConfig(reader)
}

func mustLoadConfigFromFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := loadConfigFromFixture(t)
	if err != nil {
		t.Fatalf("loadConfigFromFixture() error = %v", err)
	}

	return cfg
}
