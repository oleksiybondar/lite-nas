package config

import "gopkg.in/ini.v1"

// EventStoreConfig defines the composed eventstore configuration contract.
//
// The structure mirrors INI section ownership:
//   - [eventstore] -> Storage
//   - [eventstore_writer] -> Writer
//   - [eventstore_cleanup] -> Cleanup
type EventStoreConfig struct {
	Storage EventStoreStorageConfig
	Writer  EventStoreWriterConfig
	Cleanup EventStoreCleanupConfig
}

// LoadEventStoreConfig extracts and validates eventstore sections from the INI file.
func LoadEventStoreConfig(cfgFile *ini.File) (EventStoreConfig, error) {
	storage, err := loadEventStoreStorageConfig(cfgFile.Section("eventstore"))
	if err != nil {
		return EventStoreConfig{}, err
	}

	writer, err := loadEventStoreWriterConfig(cfgFile.Section("eventstore_writer"))
	if err != nil {
		return EventStoreConfig{}, err
	}

	cleanup, err := loadEventStoreCleanupConfig(cfgFile.Section("eventstore_cleanup"))
	if err != nil {
		return EventStoreConfig{}, err
	}

	return EventStoreConfig{
		Storage: storage,
		Writer:  writer,
		Cleanup: cleanup,
	}, nil
}
