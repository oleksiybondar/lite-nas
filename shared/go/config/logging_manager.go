package config

import (
	loggingmanagerconfig "lite-nas/shared/config/loggingmanager"
	"lite-nas/shared/fileio"
)

// LoggingManagerServiceConfig defines shared runtime configuration for
// logging-manager services.
//
// It composes shared transport connectivity and loggingmanager runtime settings so
// multiple logging-manager services can reuse the same config contract.
type LoggingManagerServiceConfig struct {
	Messaging      MessagingConfig
	LoggingManager loggingmanagerconfig.LoggingManagerConfig
}

// LoadLoggingManagerServiceConfig reads and parses logging-manager
// service-level configuration.
func LoadLoggingManagerServiceConfig(reader fileio.Reader) (LoggingManagerServiceConfig, error) {
	cfgFile, err := LoadINI(reader)
	if err != nil {
		return LoggingManagerServiceConfig{}, err
	}

	messagingConfig, err := LoadMessagingConfig(cfgFile)
	if err != nil {
		return LoggingManagerServiceConfig{}, err
	}

	loggingManagerConfig, err := loggingmanagerconfig.LoadLoggingManagerConfig(cfgFile)
	if err != nil {
		return LoggingManagerServiceConfig{}, err
	}

	return LoggingManagerServiceConfig{
		Messaging:      messagingConfig,
		LoggingManager: loggingManagerConfig,
	}, nil
}
