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
	if cfg.Metrics.PollInterval != 8*time.Second {
		t.Fatalf("Metrics.PollInterval = %v, want 8s", cfg.Metrics.PollInterval)
	}
	if cfg.Metrics.HistorySize != 25 {
		t.Fatalf("Metrics.HistorySize = %d, want 25", cfg.Metrics.HistorySize)
	}
}

func TestLoadConfigUsesDefaultMetricsValues(t *testing.T) {
	t.Parallel()

	cfg := configtest.MustLoadConfig(t, LoadConfig, defaultMetricsConfigINI())
	if cfg.Metrics.PollInterval != 5*time.Second {
		t.Fatalf("Metrics.PollInterval = %v, want 5s", cfg.Metrics.PollInterval)
	}
	if cfg.Metrics.HistorySize != 10 {
		t.Fatalf("Metrics.HistorySize = %d, want 10", cfg.Metrics.HistorySize)
	}
}

func TestLoadConfigRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	invalidINI := strings.Replace(validConfigINI(), "poll_interval = 8s", "poll_interval = nope", 1)
	configtest.RunRejectsInvalidConfigCase(t, LoadConfig, invalidINI)
}

// validConfigINI returns one complete service configuration fixture.
func validConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("service-metrics") +
		"\n[metrics]\n" +
		"poll_interval = 8s\n" +
		"history_size = 25\n"
}

// defaultMetricsConfigINI returns a fixture that relies on metric defaults.
func defaultMetricsConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("service-metrics")
}
