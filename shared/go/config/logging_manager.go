package config

import "lite-nas/shared/fileio"

// LoggingManagerConfig defines shared runtime configuration for logging-manager services.
//
// It composes shared transport connectivity and eventstore runtime settings so
// multiple logging-manager services can reuse the same config contract.
type LoggingManagerConfig struct {
	Messaging  MessagingConfig
	EventStore EventStoreConfig
}

// LoadLoggingManagerConfig reads and parses logging-manager configuration.
func LoadLoggingManagerConfig(reader fileio.Reader) (LoggingManagerConfig, error) {
	cfgFile, err := LoadINI(reader)
	if err != nil {
		return LoggingManagerConfig{}, err
	}

	messagingConfig, err := LoadMessagingConfig(cfgFile)
	if err != nil {
		return LoggingManagerConfig{}, err
	}

	eventStoreConfig, err := LoadEventStoreConfig(cfgFile)
	if err != nil {
		return LoggingManagerConfig{}, err
	}

	return LoggingManagerConfig{
		Messaging:  messagingConfig,
		EventStore: eventStoreConfig,
	}, nil
}
