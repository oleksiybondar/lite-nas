package config

import (
	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

// Config defines runtime configuration for the system email notifier service.
//
// It groups shared infrastructure settings needed by the service bootstrap:
//   - Messaging: external messaging system connectivity
//   - Email: outbound recipient and sender policy
//   - SMTP: local SMTP delivery settings
//   - Logging: application logging behavior
type Config struct {
	Messaging sharedconfig.MessagingConfig
	Email     sharedconfig.EmailConfig
	SMTP      sharedconfig.SMTPConfig
	Logging   sharedconfig.LoggingConfig
}

// LoadConfig reads service configuration from a file abstraction and returns a
// parsed Config value.
func LoadConfig(reader fileio.Reader) (Config, error) {
	cfgFile, err := sharedconfig.LoadINI(reader)
	if err != nil {
		return Config{}, err
	}

	sharedCfg, err := sharedconfig.LoadSharedEmailConfig(cfgFile)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Messaging: sharedCfg.Messaging,
		Email:     sharedCfg.Email,
		SMTP:      sharedCfg.SMTP,
		Logging:   sharedCfg.Logging,
	}, nil
}
