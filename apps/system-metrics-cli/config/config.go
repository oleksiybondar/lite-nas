package config

import (
	"lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

type Config struct {
	Messaging config.MessagingConfig
	Logging   config.LoggingConfig
}

func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := config.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := config.LoadSharedConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging: sharedCfg.Messaging,
		Logging:   sharedCfg.Logging,
	}, nil
}
