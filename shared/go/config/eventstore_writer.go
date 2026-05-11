package config

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/ini.v1"
)

var (
	errInvalidEventStoreWriterBatchSize        = errors.New("eventstore_writer batch_size must be greater than zero")
	errInvalidEventStoreWriterFlushInterval    = errors.New("eventstore_writer flush_interval must be greater than zero")
	errInvalidEventStoreWriterFlushIntervalFmt = errors.New("eventstore_writer flush_interval has invalid duration")
)

// EventStoreWriterConfig defines eventstore write-batching runtime behavior.
type EventStoreWriterConfig struct {
	BatchSize     int
	FlushInterval time.Duration
}

// loadEventStoreWriterConfig validates and parses the [eventstore_writer] section.
func loadEventStoreWriterConfig(section *ini.Section) (EventStoreWriterConfig, error) {
	batchSize, err := section.Key("batch_size").Int()
	if err != nil {
		return EventStoreWriterConfig{}, fmt.Errorf("eventstore_writer batch_size: %w", err)
	}

	flushInterval, err := time.ParseDuration(section.Key("flush_interval").String())
	if err != nil {
		return EventStoreWriterConfig{}, fmt.Errorf("%w: %v", errInvalidEventStoreWriterFlushIntervalFmt, err)
	}

	if batchSize <= 0 {
		return EventStoreWriterConfig{}, errInvalidEventStoreWriterBatchSize
	}

	if flushInterval <= 0 {
		return EventStoreWriterConfig{}, errInvalidEventStoreWriterFlushInterval
	}

	return EventStoreWriterConfig{
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
	}, nil
}
