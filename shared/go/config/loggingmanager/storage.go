package loggingmanager

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	errMissingLoggingManagerSQLitePath     = errors.New("loggingmanager sqlite_path is required")
	errInvalidLoggingManagerMaxEvents      = errors.New("loggingmanager max_events must be greater than zero")
	errInvalidLoggingManagerMaxOccurrences = errors.New("loggingmanager max_occurrences must be greater than zero")
)

// LoggingManagerStorageConfig defines bounded SQLite storage settings for loggingmanager.
type LoggingManagerStorageConfig struct {
	SQLitePath     string
	MaxEvents      int
	MaxOccurrences int
}

// loadLoggingManagerStorageConfig validates and parses the [loggingmanager] section.
func loadLoggingManagerStorageConfig(section *ini.Section) (LoggingManagerStorageConfig, error) {
	sqlitePath := strings.TrimSpace(section.Key("sqlite_path").String())
	if sqlitePath == "" {
		return LoggingManagerStorageConfig{}, errMissingLoggingManagerSQLitePath
	}

	maxEvents, err := section.Key("max_events").Int()
	if err != nil {
		return LoggingManagerStorageConfig{}, fmt.Errorf("loggingmanager max_events: %w", err)
	}

	maxOccurrences, err := section.Key("max_occurrences").Int()
	if err != nil {
		return LoggingManagerStorageConfig{}, fmt.Errorf("loggingmanager max_occurrences: %w", err)
	}

	if maxEvents <= 0 {
		return LoggingManagerStorageConfig{}, errInvalidLoggingManagerMaxEvents
	}

	if maxOccurrences <= 0 {
		return LoggingManagerStorageConfig{}, errInvalidLoggingManagerMaxOccurrences
	}

	return LoggingManagerStorageConfig{
		SQLitePath:     sqlitePath,
		MaxEvents:      maxEvents,
		MaxOccurrences: maxOccurrences,
	}, nil
}
