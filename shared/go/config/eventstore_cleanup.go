package config

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/ini.v1"
)

var (
	errInvalidEventStoreCleanupBatchSize   = errors.New("eventstore_cleanup batch_size must be greater than zero")
	errInvalidEventStoreCleanupInterval    = errors.New("eventstore_cleanup interval must be greater than zero")
	errInvalidEventStoreCleanupIntervalFmt = errors.New("eventstore_cleanup interval has invalid duration")
)

// EventStoreCleanupConfig defines eventstore occurrence cleanup runtime behavior.
type EventStoreCleanupConfig struct {
	BatchSize int
	Interval  time.Duration
}

// loadEventStoreCleanupConfig validates and parses the [eventstore_cleanup] section.
func loadEventStoreCleanupConfig(section *ini.Section) (EventStoreCleanupConfig, error) {
	batchSize, err := section.Key("batch_size").Int()
	if err != nil {
		return EventStoreCleanupConfig{}, fmt.Errorf("eventstore_cleanup batch_size: %w", err)
	}

	interval, err := time.ParseDuration(section.Key("interval").String())
	if err != nil {
		return EventStoreCleanupConfig{}, fmt.Errorf("%w: %v", errInvalidEventStoreCleanupIntervalFmt, err)
	}

	if batchSize <= 0 {
		return EventStoreCleanupConfig{}, errInvalidEventStoreCleanupBatchSize
	}

	if interval <= 0 {
		return EventStoreCleanupConfig{}, errInvalidEventStoreCleanupInterval
	}

	return EventStoreCleanupConfig{
		BatchSize: batchSize,
		Interval:  interval,
	}, nil
}
