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

// LoadSharedConfig extracts the shared [messaging] and [logging] sections from
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

	return SharedConfig{
		Messaging: messagingConfig,
		Logging:   loggingConfig,
	}, nil
}
