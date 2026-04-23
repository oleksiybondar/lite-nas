package config

import (
	"lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

type Config struct {
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
}

func LoadConfig(reader fileio.Reader) (Config, error) {
	data, err := reader.Read()
	if err != nil {
		return Config{}, err
	}

	cfgFile, err := ini.Load(data)
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
		Messaging: messagingConfig,
		Logging:   loggingConfig,
	}, nil
}
