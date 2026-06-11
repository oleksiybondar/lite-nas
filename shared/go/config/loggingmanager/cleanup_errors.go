package loggingmanager

import "errors"

var (
	errInvalidLoggingManagerCleanupBatchSize   = errors.New("loggingmanager_cleanup batch_size must be greater than zero")
	errInvalidLoggingManagerCleanupInterval    = errors.New("loggingmanager_cleanup interval must be greater than zero")
	errInvalidLoggingManagerCleanupIntervalFmt = errors.New("loggingmanager_cleanup interval has invalid duration")
)
