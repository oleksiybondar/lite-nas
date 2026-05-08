package config

import (
	"lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// Config defines runtime configuration for the web gateway service.
//
// It groups configuration by domain to reflect service responsibilities:
//   - HTTP: browser-facing HTTP listener settings
//   - Messaging: internal messaging system connectivity
//   - AuthTokens: access-token verification policy
//   - Logging: application logging behavior
type Config struct {
	HTTP       config.HTTPConfig
	Messaging  config.MessagingConfig
	AuthTokens config.AuthTokenConfig
	Logging    config.LoggingConfig
}

// LoadConfig reads gateway configuration from a file abstraction and returns a
// parsed Config value.
//
// Parameters:
//   - reader: source of the INI configuration document
//
// The function performs the following steps:
//   - reads raw configuration data using the Reader abstraction
//   - parses INI content using the ini library
//   - delegates section-specific parsing to dedicated helpers
//
// Returns an error if reading, parsing, or validation fails.
func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := config.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := config.LoadSharedConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	httpConfig, err := config.LoadHTTPConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	authTokenConfig, err := config.LoadAuthTokenConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		HTTP:       httpConfig,
		Messaging:  sharedCfg.Messaging,
		AuthTokens: authTokenConfig,
		Logging:    sharedCfg.Logging,
	}, nil
}
