package modules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewInfraModuleReturnsReaderErrorForMissingConfig(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/non-existent/process-metrics.conf", "process-metrics")
	if err == nil {
		t.Fatal("NewInfraModule() error = nil, want missing config error")
	}
}

func TestNewInfraModuleReturnsConfigErrorAfterReadingFile(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "process-metrics.conf")
	if err := os.WriteFile(configPath, []byte(invalidInfraConfigINI()), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	_, err := NewInfraModule(configPath, "process-metrics")
	if err == nil {
		t.Fatal("NewInfraModule() error = nil, want config parsing error")
	}
}

func TestNewInfraModuleLoadsConfigAndBuildsInfra(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "process-metrics.conf")
	if err := os.WriteFile(configPath, []byte(validInfraConfigINI()), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	infra, err := NewInfraModule(configPath, "process-metrics")
	if err != nil {
		t.Fatalf("NewInfraModule() error = %v", err)
	}
	defer infra.Close()

	if !infra.Config.ProcessMetrics.Enabled {
		t.Fatal("Config.ProcessMetrics.Enabled = false, want true")
	}
}

func invalidInfraConfigINI() string {
	return `[messaging]
url = nats://127.0.0.1:1
client_name = process-metrics
timeout = 1s

[logging]
level = info
format = rfc5424
output = stdout

[auth]

[process_metrics]
enabled = true
poll_interval = nope
`
}

func validInfraConfigINI() string {
	return `[messaging]
url = nats://127.0.0.1:1
client_name = process-metrics
timeout = 1s

[logging]
level = info
format = rfc5424
output = stdout

[auth]

[process_metrics]
enabled = true
poll_interval = 5s
`
}
