package loggingmanager

import "errors"

var (
	errInvalidLoggingManagerWriterBatchSize        = errors.New("loggingmanager_writer batch_size must be greater than zero")
	errInvalidLoggingManagerWriterFlushInterval    = errors.New("loggingmanager_writer flush_interval must be greater than zero")
	errInvalidLoggingManagerWriterFlushIntervalFmt = errors.New("loggingmanager_writer flush_interval has invalid duration")
)
