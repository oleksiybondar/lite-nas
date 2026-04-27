package modules

import (
	"os"
	"path/filepath"
	"testing"

	serviceconfig "lite-nas/services/system-metrics/config"
)

func TestNewInfraModuleReturnsConfigReaderError(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/missing/system-metrics.conf", "system-metrics")
	if err == nil {
		t.Fatal("expected config reader error")
	}
}

func TestNewInfraModuleReturnsLoggerError(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "system-metrics.conf")
	configData := []byte(
		"[metrics]\n" +
			"poll_interval=1s\n" +
			"history_size=2\n" +
			"[messaging]\n" +
			"url=nats://localhost:4222\n" +
			"client_name=system-metrics\n" +
			"timeout=1s\n" +
			"[logging]\n" +
			"level=info\n" +
			"format=rfc5424\n" +
			"output=file\n" +
			"file_path=/tmp/system-metrics.log\n",
	)

	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := NewInfraModule(configPath, "system-metrics")
	if err == nil {
		t.Fatal("expected logger initialization error")
	}
}

func TestInfraConfigFieldExposesConfig(t *testing.T) {
	t.Parallel()

	module, _, _, _ := loadInfraFixture()

	if module.Config != (serviceconfig.Config{}) {
		t.Fatal("expected Config field to expose module config")
	}
}

func TestInfraLoggerFieldExposesLogger(t *testing.T) {
	t.Parallel()

	module, _, _, _ := loadInfraFixture()

	if module.Logger == nil {
		t.Fatal("expected Logger field to expose module logger")
	}
}

func TestInfraClientFieldExposesClient(t *testing.T) {
	t.Parallel()

	module, client, _, _ := loadInfraFixture()

	if module.Client != client {
		t.Fatal("expected Client field to expose module client")
	}
}

func TestInfraServerFieldExposesServer(t *testing.T) {
	t.Parallel()

	module, _, server, _ := loadInfraFixture()

	if module.Server != server {
		t.Fatal("expected Server field to expose module server")
	}
}

func TestInfraCloseDrainsAndClosesClient(t *testing.T) {
	t.Parallel()

	module, client, _, _ := loadInfraFixture()

	module.Close()

	if !client.drained {
		t.Fatal("expected client to be drained")
	}

	if !client.closed {
		t.Fatal("expected client to be closed")
	}
}

func TestInfraCloseDrainsAndClosesServer(t *testing.T) {
	t.Parallel()

	module, _, server, _ := loadInfraFixture()

	module.Close()

	if !server.drained {
		t.Fatal("expected server to be drained")
	}

	if !server.closed {
		t.Fatal("expected server to be closed")
	}
}

func TestInfraCloseRunsLogCleanup(t *testing.T) {
	t.Parallel()

	module, _, _, cleanupCalls := loadInfraFixture()

	module.Close()

	if *cleanupCalls != 1 {
		t.Fatalf("cleanupCalls = %d, want 1", *cleanupCalls)
	}
}
