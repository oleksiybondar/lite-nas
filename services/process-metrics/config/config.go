package config

import (
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the process metrics service.
type Config struct {
	ProcessMetrics ProcessMetricsConfig
	Messaging      sharedconfig.MessagingConfig
	Logging        sharedconfig.LoggingConfig
}

// ProcessMetricsConfig defines settings related to live process collection.
type ProcessMetricsConfig struct {
	Enabled      bool
	PollInterval time.Duration
}

// LoadConfig reads and parses service configuration from the provided reader.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return sharedconfig.LoadConfigWithSharedSections(
		reader,
		func(cfgFile *ini.File, sharedCfg sharedconfig.SharedConfig) (Config, error) {
			processCfg, err := loadProcessMetricsConfig(cfgFile)
			if err != nil {
				return Config{}, err
			}

			return Config{
				ProcessMetrics: processCfg,
				Messaging:      sharedCfg.Messaging,
				Logging:        sharedCfg.Logging,
			}, nil
		},
	)
}

// loadProcessMetricsConfig loads the process-metrics-specific
// [process_metrics] section.
func loadProcessMetricsConfig(cfgFile *ini.File) (ProcessMetricsConfig, error) {
	section := cfgFile.Section("process_metrics")

	pollInterval, err := time.ParseDuration(section.Key("poll_interval").MustString("5s"))
	if err != nil {
		return ProcessMetricsConfig{}, err
	}

	return ProcessMetricsConfig{
		Enabled:      section.Key("enabled").MustBool(true),
		PollInterval: pollInterval,
	}, nil
}
