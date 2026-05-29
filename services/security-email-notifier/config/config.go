package config

import (
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// Config defines runtime configuration for the security email notifier
// service.
//
// It groups shared infrastructure settings needed by the service bootstrap:
//   - Messaging: external messaging system connectivity
//   - Logging: application logging behavior
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads service configuration from a file abstraction and returns a
// parsed Config value.
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
