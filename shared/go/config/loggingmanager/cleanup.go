package loggingmanager

import (
	"time"

	"gopkg.in/ini.v1"
)

// LoggingManagerCleanupConfig defines loggingmanager occurrence cleanup runtime behavior.
type LoggingManagerCleanupConfig struct {
	BatchSize int
	Interval  time.Duration
}

// loadLoggingManagerCleanupConfig validates and parses the [loggingmanager_cleanup] section.
func loadLoggingManagerCleanupConfig(section *ini.Section) (LoggingManagerCleanupConfig, error) {
	batchSize, interval, err := loadLoggingManagerBatchRuntime(
		section,
		"loggingmanager_cleanup",
		"interval",
		errInvalidLoggingManagerCleanupIntervalFmt,
	)
	if err != nil {
		return LoggingManagerCleanupConfig{}, err
	}

	if batchSize <= 0 {
		return LoggingManagerCleanupConfig{}, errInvalidLoggingManagerCleanupBatchSize
	}

	if interval <= 0 {
		return LoggingManagerCleanupConfig{}, errInvalidLoggingManagerCleanupInterval
	}

	return LoggingManagerCleanupConfig{
		BatchSize: batchSize,
		Interval:  interval,
	}, nil
}
