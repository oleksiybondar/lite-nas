package config

import (
	"strings"
	"testing"
	"time"

	"lite-nas/shared/testutil/configtest"
)

func TestLoadConfigParsesProcessMetricsAndSharedSections(t *testing.T) {
	t.Parallel()

	cfg := configtest.MustLoadConfig(t, LoadConfig, validConfigINI())
	if !cfg.ProcessMetrics.Enabled {
		t.Fatal("ProcessMetrics.Enabled = false, want true")
	}
	if cfg.ProcessMetrics.PollInterval != 8*time.Second {
		t.Fatalf("ProcessMetrics.PollInterval = %v, want 8s", cfg.ProcessMetrics.PollInterval)
	}
}

func TestLoadConfigUsesDefaultProcessMetricsValues(t *testing.T) {
	t.Parallel()

	cfg := configtest.MustLoadConfig(t, LoadConfig, defaultMetricsConfigINI())
	if !cfg.ProcessMetrics.Enabled {
		t.Fatal("ProcessMetrics.Enabled = false, want true")
	}
	if cfg.ProcessMetrics.PollInterval != 5*time.Second {
		t.Fatalf("ProcessMetrics.PollInterval = %v, want 5s", cfg.ProcessMetrics.PollInterval)
	}
}

func TestLoadConfigRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	invalidINI := strings.Replace(validConfigINI(), "poll_interval = 8s", "poll_interval = nope", 1)
	configtest.RunRejectsInvalidConfigCase(t, LoadConfig, invalidINI)
}

func validConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("process-metrics") +
		"\n[process_metrics]\n" +
		"enabled = true\n" +
		"poll_interval = 8s\n"
}

func defaultMetricsConfigINI() string {
	return configtest.MetricsServiceSharedConfigFixture("process-metrics")
}
