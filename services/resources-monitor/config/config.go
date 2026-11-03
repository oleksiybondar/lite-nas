package config

import (
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// Config defines runtime configuration for the resources monitor service.
//
// It groups configuration by domain to reflect service responsibilities:
//   - Messaging: external messaging system connectivity
//   - Auth: service auth identity settings
//   - Rules: rule file locations used by the monitor
//   - Logging: application logging behavior
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Auth      sharedconfig.AuthConfig
	Rules     sharedconfig.RulesConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads monitor configuration from a file abstraction and returns a
// parsed Config value.
func LoadConfig(reader fileio.Reader) (Config, error) {
	return sharedconfig.LoadConfigWithSharedSections(
		reader,
		func(cfgFile *ini.File, sharedCfg sharedconfig.SharedConfig) (Config, error) {
			rulesConfig, err := sharedconfig.LoadRulesConfig(cfgFile)
			if err != nil {
				return Config{}, err
			}

			return Config{
				Messaging: sharedCfg.Messaging,
				Auth:      sharedCfg.Auth,
				Rules:     rulesConfig,
				Logging:   sharedCfg.Logging,
			}, nil
		},
	)
}
