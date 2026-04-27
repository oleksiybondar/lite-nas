package config

import (
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the auth service.
//
// It groups configuration by domain to reflect service responsibilities:
//   - Messaging: internal messaging system connectivity
//   - Logging: application logging behavior
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads auth-service configuration from a file abstraction and
// returns a parsed Config value.
func LoadConfig(reader fileio.Reader) (Config, error) {
	data, err := reader.Read()
	if err != nil {
		return Config{}, err
	}

	cfgFile, err := ini.Load(data)
	if err != nil {
		return Config{}, err
	}

	messagingConfig, err := sharedconfig.LoadMessagingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	loggingConfig, err := sharedconfig.LoadLoggingConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging: messagingConfig,
		Logging:   loggingConfig,
	}, nil
}
