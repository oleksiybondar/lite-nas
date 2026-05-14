package config

import (
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// Config defines runtime configuration for the resources monitor service.
//
// It groups configuration by domain to reflect service responsibilities:
//   - Messaging: external messaging system connectivity
//   - Rules: rule file locations used by the monitor
//   - Logging: application logging behavior
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Rules     sharedconfig.RulesConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads monitor configuration from a file abstraction and returns a
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

	rulesConfig, err := sharedconfig.LoadRulesConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging: sharedCfg.Messaging,
		Rules:     rulesConfig,
		Logging:   sharedCfg.Logging,
	}, nil
}
