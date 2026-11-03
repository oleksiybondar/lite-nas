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

func TestLoadMetricsConfigUsesDefaults(t *testing.T) {
	t.Parallel()

	cfg := configtest.MustLoadConfig(t, LoadConfig, defaultMetricsConfigINI())
	if cfg.Metrics.PollInterval != time.Second {
		t.Fatalf("Metrics.PollInterval = %v, want 1s", cfg.Metrics.PollInterval)
	}
	if cfg.Metrics.HistorySize != 120 {
		t.Fatalf("Metrics.HistorySize = %d, want 120", cfg.Metrics.HistorySize)
	}
}

func TestLoadConfigRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		LoadConfig,
		strings.Replace(validConfigINI(), "poll_interval = 2s", "poll_interval = nope", 1),
	)
}

func validConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("network-metrics") +
		"\n[metrics]\n" +
		"poll_interval = 2s\n" +
		"history_size = 64\n"
}

func defaultMetricsConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("network-metrics")
}
