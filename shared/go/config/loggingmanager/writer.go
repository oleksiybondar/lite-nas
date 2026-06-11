package loggingmanager

import (
	"time"

	"gopkg.in/ini.v1"
)

// LoggingManagerWriterConfig defines loggingmanager write-batching runtime behavior.
type LoggingManagerWriterConfig struct {
	BatchSize     int
	FlushInterval time.Duration
}

// loadLoggingManagerWriterConfig validates and parses the [loggingmanager_writer] section.
func loadLoggingManagerWriterConfig(section *ini.Section) (LoggingManagerWriterConfig, error) {
	batchSize, flushInterval, err := loadLoggingManagerBatchRuntime(
		section,
		"loggingmanager_writer",
		"flush_interval",
		errInvalidLoggingManagerWriterFlushIntervalFmt,
	)
	if err != nil {
		return LoggingManagerWriterConfig{}, err
	}

	if batchSize <= 0 {
		return LoggingManagerWriterConfig{}, errInvalidLoggingManagerWriterBatchSize
	}

	if flushInterval <= 0 {
		return LoggingManagerWriterConfig{}, errInvalidLoggingManagerWriterFlushInterval
	}

	return LoggingManagerWriterConfig{
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
	}, nil
}
