package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewInfraModuleReturnsReaderErrorForMissingConfig(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/non-existent/network-metrics.conf", "network-metrics")
	if err == nil {
		t.Fatal("NewInfraModule() error = nil, want missing config error")
	}
}

func TestNewInfraModuleReturnsConfigErrorAfterReadingFile(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "network-metrics.conf")
	if err := os.WriteFile(configPath, []byte(invalidInfraConfigINI()), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	_, err := NewInfraModule(configPath, "network-metrics")
	if err == nil {
		t.Fatal("NewInfraModule() error = nil, want config parsing error")
	}
}

func TestNewInfraModuleLoadsConfigAndBuildsInfra(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "network-metrics.conf")
	if err := os.WriteFile(configPath, []byte(validInfraConfigINI()), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	infra, err := NewInfraModule(configPath, "network-metrics")
	if err != nil {
		t.Fatalf("NewInfraModule() error = %v", err)
	}
	defer infra.Close()

	if infra.Config.Metrics.HistorySize != 10 {
		t.Fatalf("Config.Metrics.HistorySize = %d, want 10", infra.Config.Metrics.HistorySize)
	}
}

func invalidInfraConfigINI() string {
	return `[messaging]
url = nats://127.0.0.1:1
client_name = network-metrics
timeout = 1s

[logging]
level = info
format = rfc5424
output = stdout

[auth]

[metrics]
poll_interval = nope
history_size = 10
`
}

func validInfraConfigINI() string {
	return `[messaging]
url = nats://127.0.0.1:1
client_name = network-metrics
timeout = 1s

[logging]
level = info
format = rfc5424
output = stdout

[auth]

[metrics]
poll_interval = 1s
history_size = 10
`
}
