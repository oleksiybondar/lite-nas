package config

import (
	"lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the web gateway service.
//
// It groups configuration by domain to reflect service responsibilities:
//   - HTTP: browser-facing HTTP listener settings
//   - Messaging: internal messaging system connectivity
//   - Logging: application logging behavior
type Config struct {
	HTTP      config.HTTPConfig
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
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
	data, err := reader.Read()
	if err != nil {
		return Config{}, err
	}

	cfgFile, err := ini.Load(data)
	if err != nil {
		return Config{}, err
	}

	httpConfig, err := config.LoadHTTPConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	messagingConfig, err := config.LoadMessagingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	loggingConfig, err := config.LoadLoggingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTP:      httpConfig,
		Messaging: messagingConfig,
		Logging:   loggingConfig,
	}, nil
}
