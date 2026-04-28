package config

import (
	"time"

	"lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the system metrics service.
//
// It groups configuration by domain to reflect service responsibilities:
//   - Metrics: polling and history behavior
//   - Messaging: external messaging system connectivity
type Config struct {
	Metrics   MetricsConfig
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
}

// MetricsConfig defines settings related to system metrics collection.
//
// PollInterval controls how often metrics are collected.
// HistorySize defines the number of snapshots retained in memory.
type MetricsConfig struct {
	PollInterval time.Duration
	HistorySize  int
}

// LoadConfig reads configuration data from the provided Reader and parses it
// into a Config struct.
//
// The function performs the following steps:
//   - reads raw configuration data using the Reader abstraction
//   - parses INI content using the ini library
//   - delegates section-specific parsing to dedicated helpers
//
// An error is returned if reading, parsing, or validation fails.
func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := config.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := config.LoadSharedConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	metricsConfig, err := loadMetricsConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Metrics:   metricsConfig,
		Messaging: sharedCfg.Messaging,
		Logging:   sharedCfg.Logging,
	}, nil
}

// loadMetricsConfig extracts and parses the [metrics] section from the INI file.
//
// Expected keys:
//   - poll_interval: duration string (e.g. "1s", "500ms")
//   - history_size: integer number of snapshots to retain
//
// poll_interval defaults to "1s" if not provided.
// An error is returned if parsing fails.
func loadMetricsConfig(cfgFile *ini.File) (MetricsConfig, error) {
	section := cfgFile.Section("metrics")

	pollInterval, err := time.ParseDuration(section.Key("poll_interval").MustString("1s"))
	if err != nil {
		return MetricsConfig{}, err
	}

	historySize, err := section.Key("history_size").Int()
	if err != nil {
		return MetricsConfig{}, err
	}

	return MetricsConfig{
		PollInterval: pollInterval,
		HistorySize:  historySize,
	}, nil
}
