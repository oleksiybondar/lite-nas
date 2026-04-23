// Package logger provides a small project-local logging abstraction backed by log/slog.
package logger

import (
	"errors"
	"io"
	"log/slog"
)

// Logger defines the project-local logging interface used by application code.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
}

// RecordMetadata contains contextual values that formatters may place in every record.
type RecordMetadata struct {
	ServiceName string
	Hostname    string
}

// Formatter serializes one slog record and its effective attributes.
type Formatter interface {
	Format(record slog.Record, attrs []slog.Attr, metadata RecordMetadata) ([]byte, error)
}

// Config defines the runtime settings used to construct a logger.
type Config struct {
	ServiceName  string
	Hostname     string
	Writer       io.Writer
	Formatter    Formatter
	MinimumLevel slog.Level
}

// New creates a slog-backed logger that writes formatted records to the configured writer.
func New(config Config) (Logger, error) {
	if config.Writer == nil {
		return nil, errors.New("logger writer is required")
	}

	if config.Formatter == nil {
		return nil, errors.New("logger formatter is required")
	}

	handler, err := NewHandler(HandlerConfig{
		Writer:       NewLockedWriter(config.Writer),
		Formatter:    config.Formatter,
		MinimumLevel: config.MinimumLevel,
		Metadata: RecordMetadata{
			ServiceName: config.ServiceName,
			Hostname:    config.Hostname,
		},
	})
	if err != nil {
		return nil, err
	}

	return newSlogLogger(slog.New(handler)), nil
}
