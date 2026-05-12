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
	errInvalidLoggingManagerEventIDPrefix  = errors.New("loggingmanager event_id_prefix must be 1..10 ascii letters, digits, or underscore")
)

// LoggingManagerStorageConfig defines bounded SQLite storage settings for loggingmanager.
type LoggingManagerStorageConfig struct {
	SQLitePath     string
	MaxEvents      int
	MaxOccurrences int
	EventIDPrefix  string
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

	eventIDPrefix := strings.TrimSpace(section.Key("event_id_prefix").MustString("event"))
	if err = validateLoggingManagerStorageProperties(maxEvents, maxOccurrences, eventIDPrefix); err != nil {
		return LoggingManagerStorageConfig{}, err
	}

	return LoggingManagerStorageConfig{
		SQLitePath:     sqlitePath,
		MaxEvents:      maxEvents,
		MaxOccurrences: maxOccurrences,
		EventIDPrefix:  eventIDPrefix,
	}, nil
}

func validateLoggingManagerStorageProperties(
	maxEvents int,
	maxOccurrences int,
	eventIDPrefix string,
) error {
	if maxEvents <= 0 {
		return errInvalidLoggingManagerMaxEvents
	}

	if maxOccurrences <= 0 {
		return errInvalidLoggingManagerMaxOccurrences
	}

	if !isValidEventIDPrefix(eventIDPrefix) {
		return errInvalidLoggingManagerEventIDPrefix
	}

	return nil
}

func isValidEventIDPrefix(prefix string) bool {
	if !isEventPrefixValidLength(prefix) {
		return false
	}
	return isEventPrefixValidFormat(prefix)
}

func isEventPrefixValidLength(prefix string) bool {
	return len(prefix) > 0 && len(prefix) <= 10
}

func isEventPrefixValidFormat(prefix string) bool {
	for _, char := range prefix {
		if !isAlphanumeric(char) {
			return false
		}
	}
	return true
}

func isChar(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || char == '_'
}

func isDigit(char rune) bool {
	return char >= '0' && char <= '9'
}

func isAlphanumeric(char rune) bool {
	return isChar(char) || isDigit(char)
}
