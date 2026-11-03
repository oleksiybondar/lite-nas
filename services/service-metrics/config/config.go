package config

import (
	"time"

	"lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the service metrics service.
type Config struct {
	Metrics   MetricsConfig
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
}

// MetricsConfig defines settings related to service snapshot collection.
type MetricsConfig struct {
	PollInterval time.Duration
	HistorySize  int
}

// LoadConfig reads configuration data from the provided Reader and parses it
// into a Config struct.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return config.LoadConfigWithSharedSections(
		reader,
		func(cfgFile *ini.File, sharedCfg config.SharedConfig) (Config, error) {
			metricsConfig, err := loadMetricsConfig(cfgFile)
			if err != nil {
				return Config{}, err
			}

			return Config{
				Metrics:   metricsConfig,
				Messaging: sharedCfg.Messaging,
				Logging:   sharedCfg.Logging,
			}, nil
		},
	)
}

// loadMetricsConfig extracts and parses the [metrics] section from the INI
// file.
func loadMetricsConfig(cfgFile *ini.File) (MetricsConfig, error) {
	section := cfgFile.Section("metrics")

	pollInterval, err := time.ParseDuration(section.Key("poll_interval").MustString("5s"))
	if err != nil {
		return MetricsConfig{}, err
	}

	return MetricsConfig{
		PollInterval: pollInterval,
		HistorySize:  section.Key("history_size").MustInt(10),
	}, nil
}
