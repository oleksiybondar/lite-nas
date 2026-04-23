package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sharedfileio "lite-nas/shared/fileio"
)

type stubReader struct {
	data []byte
	err  error
}

func (reader stubReader) Read() ([]byte, error) {
	if reader.err != nil {
		return nil, reader.err
	}

	return reader.data, nil
}

func TestLoadConfigReturnsMessagingURL(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFromFixture(t)
	if err != nil {
		t.Fatalf("loadConfigFromFixture() error = %v", err)
	}

	if cfg.Messaging.URL != "nats://127.0.0.1:4222" {
		t.Fatalf("Messaging.URL = %q, want nats://127.0.0.1:4222", cfg.Messaging.URL)
	}
}

func TestLoadConfigReturnsMessagingClientName(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFromFixture(t)
	if err != nil {
		t.Fatalf("loadConfigFromFixture() error = %v", err)
	}

	if cfg.Messaging.ClientName != "system-metrics-cli" {
		t.Fatalf("Messaging.ClientName = %q, want system-metrics-cli", cfg.Messaging.ClientName)
	}
}

func TestLoadConfigReturnsMessagingTimeout(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFromFixture(t)
	if err != nil {
		t.Fatalf("loadConfigFromFixture() error = %v", err)
	}

	if cfg.Messaging.Timeout != 5*time.Second {
		t.Fatalf("Messaging.Timeout = %v, want 5s", cfg.Messaging.Timeout)
	}
}

func TestLoadConfigReturnsLoggingOutput(t *testing.T) {
	t.Parallel()

	cfg, err := loadConfigFromFixture(t)
	if err != nil {
		t.Fatalf("loadConfigFromFixture() error = %v", err)
	}

	if cfg.Logging.Output != "stderr" {
		t.Fatalf("Logging.Output = %q, want stderr", cfg.Logging.Output)
	}
}

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")

	_, err := LoadConfig(stubReader{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("LoadConfig() error = %v, want %v", err, wantErr)
	}
}

func TestLoadConfigReturnsINIParseError(t *testing.T) {
	t.Parallel()

	_, err := LoadConfig(stubReader{data: []byte("[messaging")})
	if err == nil {
		t.Fatal("LoadConfig() error = nil, want parse error")
	}
}

func TestLoadConfigReturnsLoggingConfigError(t *testing.T) {
	t.Parallel()

	_, err := LoadConfig(stubReader{
		data: []byte(
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
			"ca=/etc/liteNAS/certificates/root-ca.crt\n" +
			"cert=/etc/liteNAS/certificates/lite-nas-system-metrics-cli/client.crt\n" +
			"key=/etc/liteNAS/certificates/lite-nas-system-metrics-cli/client.key\n" +
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
