package loggingmanager

import (
	"fmt"
	"time"

	"gopkg.in/ini.v1"
)

// loadLoggingManagerBatchRuntime parses a section with batch_size and duration key.
func loadLoggingManagerBatchRuntime(
	section *ini.Section,
	sectionName string,
	durationKey string,
	durationFormatErr error,
) (int, time.Duration, error) {
	batchSize, err := loadPositiveBatchSize(section, sectionName)
	if err != nil {
		return 0, 0, err
	}

	duration, err := loadDurationKey(section, durationKey)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %w", durationFormatErr, err)
	}

	return batchSize, duration, nil
}
