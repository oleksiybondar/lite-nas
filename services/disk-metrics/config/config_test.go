package config

import (
	"strings"
	"testing"
	"time"

	"lite-nas/shared/testutil/configtest"
)

func TestLoadConfigParsesMetricsAndSharedSections(t *testing.T) {
	t.Parallel()

	cfg := configtest.MustLoadConfig(t, LoadConfig, validConfigINI())
	assertConfiguredMetrics(t, cfg)
}

func TestLoadConfigUsesDefaultMetricsValues(t *testing.T) {
	t.Parallel()

	cfg := configtest.MustLoadConfig(t, LoadConfig, defaultMetricsConfigINI())
	assertDefaultMetrics(t, cfg)
}

func TestLoadConfigRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	invalidINI := strings.Replace(validConfigINI(), "poll_interval = 2s", "poll_interval = nope", 1)
	configtest.RunRejectsInvalidConfigCase(t, LoadConfig, invalidINI)
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
	return configtest.MetricsServiceSharedConfigFixture("disk-metrics") +
		"\n[metrics]\n" +
		"poll_interval = 2s\n" +
		"history_size = 64\n"
}

func defaultMetricsConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("disk-metrics")
}
