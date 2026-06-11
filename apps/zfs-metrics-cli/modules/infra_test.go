package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewInfraModuleReturnsConfigReaderError(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/non-existent/zfs-metrics-cli.conf", "zfs-metrics-cli")
	if err == nil {
		t.Fatal("NewInfraModule() error = nil, want config reader error")
	}
}

func TestNewInfraModuleBuildsInfraFromConfig(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "zfs-metrics-cli.conf")
	if err := os.WriteFile(configPath, []byte("[messaging]\nurl=nats://127.0.0.1:4222\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	infra, err := NewInfraModule(configPath, "zfs-metrics-cli")
	if err != nil {
		t.Fatalf("NewInfraModule() error = %v", err)
	}
	infra.Close()
}
