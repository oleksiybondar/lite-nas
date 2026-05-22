package config

import (
	"lite-nas/shared/fileio"

	"gopkg.in/ini.v1"
)

// SharedConfig groups the common runtime bootstrap sections reused across
// LiteNAS services and apps.
type SharedConfig struct {
	Messaging MessagingConfig
	Logging   LoggingConfig
	Auth      AuthConfig
}

// SharedAuthTokenConfig groups shared bootstrap sections reused by services
// that need local JWT verification policy.
type SharedAuthTokenConfig struct {
	Messaging  MessagingConfig
	Logging    LoggingConfig
	AuthTokens AuthTokenConfig
}

// LoadINI reads configuration bytes from the supplied reader and parses them as
// an INI document.
func LoadINI(reader fileio.Reader) (*ini.File, error) {
	data, err := reader.Read()
	if err != nil {
		return nil, err
	}

	return ini.Load(data)
}

// LoadSharedConfig extracts shared bootstrap sections from
// a parsed INI document.
func LoadSharedConfig(cfgFile *ini.File) (SharedConfig, error) {
	messagingConfig, err := LoadMessagingConfig(cfgFile)
	if err != nil {
		return SharedConfig{}, err
	}

	loggingConfig, err := LoadLoggingConfig(cfgFile)
	if err != nil {
		return SharedConfig{}, err
	}
	authConfig, err := LoadAuthConfig(cfgFile)
	if err != nil {
		return SharedConfig{}, err
	}

	return SharedConfig{
		Messaging: messagingConfig,
		Logging:   loggingConfig,
		Auth:      authConfig,
	}, nil
}

// LoadSharedAuthTokenConfig extracts shared bootstrap sections plus
// auth-token policy from a parsed INI document.
func LoadSharedAuthTokenConfig(cfgFile *ini.File) (SharedAuthTokenConfig, error) {
	messagingConfig, err := LoadMessagingConfig(cfgFile)
	if err != nil {
		return SharedAuthTokenConfig{}, err
	}

	loggingConfig, err := LoadLoggingConfig(cfgFile)
	if err != nil {
		return SharedAuthTokenConfig{}, err
	}

	authTokenConfig, err := LoadAuthTokenConfig(cfgFile)
	if err != nil {
		return SharedAuthTokenConfig{}, err
	}

	return SharedAuthTokenConfig{
		Messaging:  messagingConfig,
		Logging:    loggingConfig,
		AuthTokens: authTokenConfig,
	}, nil
}
