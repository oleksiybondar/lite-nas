package config

import (
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the zfs-metrics service.
type Config struct {
	Metrics   MetricsConfig
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

// MetricsConfig defines ZFS snapshot collection settings.
type MetricsConfig struct {
	PollInterval time.Duration
	ZpoolPath    string
}

// LoadConfig loads service configuration from the provided reader.
func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := sharedconfig.LoadSharedConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	metricsCfg, err := loadMetricsConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Metrics:   metricsCfg,
		Messaging: sharedCfg.Messaging,
		Logging:   sharedCfg.Logging,
	}, nil
}

// loadMetricsConfig loads the [metrics] section specific to zfs-metrics.
func loadMetricsConfig(cfgFile *ini.File) (MetricsConfig, error) {
	section := cfgFile.Section("metrics")
	pollInterval, err := time.ParseDuration(section.Key("poll_interval").MustString("5s"))
	if err != nil {
		return MetricsConfig{}, err
	}

	return MetricsConfig{
		PollInterval: pollInterval,
		ZpoolPath:    section.Key("zpool_path").MustString("/sbin/zpool"),
	}, nil
}
