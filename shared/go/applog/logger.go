package applog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	sharedconfig "lite-nas/shared/config"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/logger/formatters/rfc5424"
)

var allowedLogDir = "/var/lib/lite-nas"

// NewAppLogger creates a project-standard application logger from the shared
// logging configuration.
func NewAppLogger(serviceName string, cfg sharedconfig.LoggingConfig) (sharedlogger.Logger, func(), error) {
	writer, cleanup, err := buildLogWriter(cfg)
	if err != nil {
		return nil, nil, err
	}

	log, err := sharedlogger.New(sharedlogger.Config{
		ServiceName:  serviceName,
		Hostname:     hostnameOrFallback(),
		Writer:       writer,
		Formatter:    newFormatter(serviceName),
		MinimumLevel: mapLogLevel(cfg.Level),
	})
	if err != nil {
		cleanup()
		return nil, nil, err
	}

	return log, cleanup, nil
}

func newFormatter(serviceName string) rfc5424.Formatter {
	return rfc5424.New(rfc5424.Config{
		ProcessID: strconv.Itoa(os.Getpid()),
		MessageID: buildMessageID(serviceName),
	})
}

func buildMessageID(serviceName string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(serviceName), "-", "_"))
	if normalized == "" {
		return "APP"
	}

	return normalized
}

func buildLogWriter(cfg sharedconfig.LoggingConfig) (io.Writer, func(), error) {
	switch cfg.Output {
	case "stderr":
		return os.Stderr, func() {}, nil
	case "file":
		if err := validateLogFilePath(cfg.FilePath); err != nil {
			return nil, nil, err
		}

		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, nil, err
		}

		return file, func() {
			_ = file.Close()
		}, nil
	default:
		return os.Stdout, func() {}, nil
	}
}

func validateLogFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("log file path is required")
	}

	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("log file path must be absolute")
	}

	relPath, err := filepath.Rel(filepath.Clean(allowedLogDir), cleanPath)
	if err != nil {
		return err
	}

	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return fmt.Errorf("log file path must be within %s", allowedLogDir)
	}

	return nil
}

func mapLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func hostnameOrFallback() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "unknown-host"
	}

	return hostname
}
