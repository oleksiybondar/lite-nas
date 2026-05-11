package loggingmanager

import "gopkg.in/ini.v1"

// LoggingManagerConfig defines the composed loggingmanager configuration contract.
//
// The structure mirrors INI section ownership:
//   - [loggingmanager] -> Storage
//   - [loggingmanager_writer] -> Writer
//   - [loggingmanager_cleanup] -> Cleanup
type LoggingManagerConfig struct {
	Storage LoggingManagerStorageConfig
	Writer  LoggingManagerWriterConfig
	Cleanup LoggingManagerCleanupConfig
}

// LoadLoggingManagerConfig extracts and validates loggingmanager sections from the INI file.
func LoadLoggingManagerConfig(cfgFile *ini.File) (LoggingManagerConfig, error) {
	storage, err := loadLoggingManagerStorageConfig(cfgFile.Section("loggingmanager"))
	if err != nil {
		return LoggingManagerConfig{}, err
	}

	writer, err := loadLoggingManagerWriterConfig(cfgFile.Section("loggingmanager_writer"))
	if err != nil {
		return LoggingManagerConfig{}, err
	}

	cleanup, err := loadLoggingManagerCleanupConfig(cfgFile.Section("loggingmanager_cleanup"))
	if err != nil {
		return LoggingManagerConfig{}, err
	}

	return LoggingManagerConfig{
		Storage: storage,
		Writer:  writer,
		Cleanup: cleanup,
	}, nil
}
