package config

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/ini.v1"
)

var (
	errMissingEventStoreSQLitePath     = errors.New("eventstore sqlite_path is required")
	errInvalidEventStoreMaxEvents      = errors.New("eventstore max_events must be greater than zero")
	errInvalidEventStoreMaxOccurrences = errors.New("eventstore max_occurrences must be greater than zero")
)

// EventStoreStorageConfig defines bounded SQLite storage settings for eventstore.
type EventStoreStorageConfig struct {
	SQLitePath     string
	MaxEvents      int
	MaxOccurrences int
}

// loadEventStoreStorageConfig validates and parses the [eventstore] section.
func loadEventStoreStorageConfig(section *ini.Section) (EventStoreStorageConfig, error) {
	sqlitePath := strings.TrimSpace(section.Key("sqlite_path").String())
	if sqlitePath == "" {
		return EventStoreStorageConfig{}, errMissingEventStoreSQLitePath
	}

	maxEvents, err := section.Key("max_events").Int()
	if err != nil {
		return EventStoreStorageConfig{}, fmt.Errorf("eventstore max_events: %w", err)
	}

	maxOccurrences, err := section.Key("max_occurrences").Int()
	if err != nil {
		return EventStoreStorageConfig{}, fmt.Errorf("eventstore max_occurrences: %w", err)
	}

	if maxEvents <= 0 {
		return EventStoreStorageConfig{}, errInvalidEventStoreMaxEvents
	}

	if maxOccurrences <= 0 {
		return EventStoreStorageConfig{}, errInvalidEventStoreMaxOccurrences
	}

	return EventStoreStorageConfig{
		SQLitePath:     sqlitePath,
		MaxEvents:      maxEvents,
		MaxOccurrences: maxOccurrences,
	}, nil
}
