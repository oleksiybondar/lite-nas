package loggingmanager

import (
	"fmt"
	"time"

	"gopkg.in/ini.v1"
)

// loadPositiveBatchSize parses and validates a positive batch_size key.
func loadPositiveBatchSize(section *ini.Section, sectionName string) (int, error) {
	batchSize, err := section.Key("batch_size").Int()
	if err != nil {
		return 0, fmt.Errorf("%s batch_size: %w", sectionName, err)
	}

	return batchSize, nil
}

// loadDurationKey parses a duration key from an INI section.
func loadDurationKey(section *ini.Section, key string) (time.Duration, error) {
	return time.ParseDuration(section.Key(key).String())
}
