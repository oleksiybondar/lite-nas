package config

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/ini.v1"
)

const (
	defaultLoggingLevel  = "info"
	defaultLoggingFormat = "rfc5424"
	defaultLoggingOutput = "stdout"
)

var (
	errInvalidLoggingLevel  = errors.New("unsupported logging level")
	errInvalidLoggingFormat = errors.New("unsupported logging format")
	errInvalidLoggingOutput = errors.New("unsupported logging output")
	errMissingLogFilePath   = errors.New("logging file_path is required when output=file")
)

// LoggingConfig defines the plain [logging] INI settings used at bootstrap time.
type LoggingConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}

// LoadLoggingConfig extracts and validates the [logging] section from the INI file.
func LoadLoggingConfig(cfgFile *ini.File) (LoggingConfig, error) {
	section := cfgFile.Section("logging")

	config := LoggingConfig{
		Level:    strings.ToLower(section.Key("level").MustString(defaultLoggingLevel)),
		Format:   strings.ToLower(section.Key("format").MustString(defaultLoggingFormat)),
		Output:   strings.ToLower(section.Key("output").MustString(defaultLoggingOutput)),
		FilePath: section.Key("file_path").String(),
	}

	if !isSupportedLoggingLevel(config.Level) {
		return LoggingConfig{}, fmt.Errorf("%w: %s", errInvalidLoggingLevel, config.Level)
	}

	if !isSupportedLoggingFormat(config.Format) {
		return LoggingConfig{}, fmt.Errorf("%w: %s", errInvalidLoggingFormat, config.Format)
	}

	if !isSupportedLoggingOutput(config.Output) {
		return LoggingConfig{}, fmt.Errorf("%w: %s", errInvalidLoggingOutput, config.Output)
	}

	if config.Output == "file" && strings.TrimSpace(config.FilePath) == "" {
		return LoggingConfig{}, errMissingLogFilePath
	}

	return config, nil
}

func isSupportedLoggingLevel(level string) bool {
	switch level {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func isSupportedLoggingFormat(format string) bool {
	return format == "rfc5424"
}

func isSupportedLoggingOutput(output string) bool {
	switch output {
	case "stdout", "stderr", "file":
		return true
	default:
		return false
	}
}
