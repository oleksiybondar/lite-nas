package config

import (
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// Config defines runtime configuration required by this CLI.
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads and validates shared CLI configuration from INI.
func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := sharedconfig.LoadSharedConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging: sharedCfg.Messaging,
		Logging:   sharedCfg.Logging,
	}, nil
}
