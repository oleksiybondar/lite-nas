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

// SharedEmailConfig groups the common runtime bootstrap sections reused by
// email notifier services.
type SharedEmailConfig struct {
	Messaging MessagingConfig
	Logging   LoggingConfig
	Email     EmailConfig
	SMTP      SMTPConfig
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

// LoadConfigWithSharedSections parses an INI document, loads the shared
// runtime sections, and delegates service-specific parsing to the supplied
// builder.
func LoadConfigWithSharedSections[T any](
	reader fileio.Reader,
	build func(cfgFile *ini.File, sharedCfg SharedConfig) (T, error),
) (T, error) {
	cfgFile, err := LoadINI(reader)
	if err != nil {
		var zero T
		return zero, err
	}

	sharedCfg, err := LoadSharedConfig(cfgFile)
	if err != nil {
		var zero T
		return zero, err
	}

	return build(cfgFile, sharedCfg)
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

// LoadSharedEmailConfig extracts shared bootstrap sections plus email delivery
// configuration from a parsed INI document.
func LoadSharedEmailConfig(cfgFile *ini.File) (SharedEmailConfig, error) {
	messagingConfig, err := LoadMessagingConfig(cfgFile)
	if err != nil {
		return SharedEmailConfig{}, err
	}

	loggingConfig, err := LoadLoggingConfig(cfgFile)
	if err != nil {
		return SharedEmailConfig{}, err
	}

	emailConfig, err := LoadEmailConfig(cfgFile)
	if err != nil {
		return SharedEmailConfig{}, err
	}

	smtpConfig, err := LoadSMTPConfig(cfgFile)
	if err != nil {
		return SharedEmailConfig{}, err
	}

	return SharedEmailConfig{
		Messaging: messagingConfig,
		Logging:   loggingConfig,
		Email:     emailConfig,
		SMTP:      smtpConfig,
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
