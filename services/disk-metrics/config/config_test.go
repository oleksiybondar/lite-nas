package config

import (
	"strings"
	"testing"
	"time"

	"lite-nas/shared/testutil/fileiotest"
)

func TestLoadConfigParsesMetricsAndSharedSections(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfig(t, validConfigINI())
	assertConfiguredMetrics(t, cfg)
}

func TestLoadConfigUsesDefaultMetricsValues(t *testing.T) {
	t.Parallel()

	cfg := mustLoadConfig(t, defaultMetricsConfigINI())
	assertDefaultMetrics(t, cfg)
}

func TestLoadConfigRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	invalidINI := strings.Replace(validConfigINI(), "poll_interval = 2s", "poll_interval = nope", 1)
	if _, err := LoadConfig(fileiotest.Reader{Data: []byte(invalidINI)}); err == nil {
		t.Fatal("LoadConfig() error = nil, want invalid duration error")
	}
}

func mustLoadConfig(t *testing.T, iniData string) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{Data: []byte(iniData)})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}

func assertConfiguredMetrics(t *testing.T, cfg Config) {
	t.Helper()

	if cfg.Metrics.PollInterval != 2*time.Second {
		t.Fatalf("Metrics.PollInterval = %v, want 2s", cfg.Metrics.PollInterval)
	}
	if cfg.Metrics.HistorySize != 64 {
		t.Fatalf("Metrics.HistorySize = %d, want 64", cfg.Metrics.HistorySize)
	}
	if cfg.Messaging.URL != "nats://127.0.0.1:4222" {
		t.Fatalf("Messaging.URL = %q, want nats://127.0.0.1:4222", cfg.Messaging.URL)
	}
	if cfg.Logging.Level != "info" {
		t.Fatalf("Logging.Level = %q, want info", cfg.Logging.Level)
	}
}

func assertDefaultMetrics(t *testing.T, cfg Config) {
	t.Helper()

	if cfg.Metrics.PollInterval != time.Second {
		t.Fatalf("Metrics.PollInterval = %v, want 1s", cfg.Metrics.PollInterval)
	}
	if cfg.Metrics.HistorySize != 120 {
		t.Fatalf("Metrics.HistorySize = %d, want 120", cfg.Metrics.HistorySize)
	}
}

func validConfigINI() string {
	return `[messaging]
url = nats://127.0.0.1:4222
client_name = disk-metrics
timeout = 3s

[logging]
level = info
format = rfc5424
output = stdout

[auth]

[metrics]
poll_interval = 2s
history_size = 64
`
}

func defaultMetricsConfigINI() string {
	return `[messaging]
url = nats://127.0.0.1:4222
client_name = disk-metrics
timeout = 3s

[logging]
level = info
format = rfc5424
output = stdout

[auth]
`
}
