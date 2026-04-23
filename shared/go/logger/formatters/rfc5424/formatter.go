// Package rfc5424 formats slog records as single-line RFC 5424 style syslog messages.
package rfc5424

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	sharedlogger "lite-nas/shared/logger"
)

const (
	defaultFacility = 1
	debugSeverity   = 7
	infoSeverity    = 6
	warnSeverity    = 4
	errorSeverity   = 3
)

// Config defines RFC 5424-specific formatter options.
type Config struct {
	Facility  int
	ProcessID string
	MessageID string
}

// Formatter emits RFC 5424-style syslog lines.
type Formatter struct {
	facility  int
	processID string
	messageID string
}

var primitiveValueFormatters = map[slog.Kind]func(slog.Value) string{
	slog.KindString: func(value slog.Value) string {
		return value.String()
	},
	slog.KindBool: func(value slog.Value) string {
		return strconv.FormatBool(value.Bool())
	},
	slog.KindInt64: func(value slog.Value) string {
		return strconv.FormatInt(value.Int64(), 10)
	},
	slog.KindUint64: func(value slog.Value) string {
		return strconv.FormatUint(value.Uint64(), 10)
	},
	slog.KindFloat64: func(value slog.Value) string {
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	},
	slog.KindDuration: func(value slog.Value) string {
		return value.Duration().String()
	},
	slog.KindTime: func(value slog.Value) string {
		return value.Time().UTC().Format(time.RFC3339Nano)
	},
}

// New creates an RFC 5424 formatter with deterministic placeholder defaults.
func New(config Config) Formatter {
	facility := config.Facility
	if facility == 0 {
		facility = defaultFacility
	}

	processID := config.ProcessID
	if processID == "" {
		processID = "-"
	}

	messageID := config.MessageID
	if messageID == "" {
		messageID = "-"
	}

	return Formatter{
		facility:  facility,
		processID: processID,
		messageID: messageID,
	}
}

// Format serializes one slog record to a single RFC 5424-style line.
func (f Formatter) Format(
	record slog.Record,
	attrs []slog.Attr,
	metadata sharedlogger.RecordMetadata,
) ([]byte, error) {
	pri := f.facility*8 + Severity(record.Level)
	timestamp := formatTimestamp(record.Time)
	hostname := sanitizeHeaderValue(metadata.Hostname)
	appName := sanitizeHeaderValue(metadata.ServiceName)
	procID := sanitizeHeaderValue(f.processID)
	msgID := sanitizeHeaderValue(f.messageID)
	message := formatMessage(record.Message, attrs)

	line := fmt.Sprintf(
		"<%d>1 %s %s %s %s %s - %s\n",
		pri,
		timestamp,
		hostname,
		appName,
		procID,
		msgID,
		message,
	)

	return []byte(line), nil
}

// Severity maps slog levels to RFC 5424 syslog severities.
func Severity(level slog.Level) int {
	switch {
	case level >= slog.LevelError:
		return errorSeverity
	case level >= slog.LevelWarn:
		return warnSeverity
	case level >= slog.LevelInfo:
		return infoSeverity
	default:
		return debugSeverity
	}
}

func formatTimestamp(value time.Time) string {
	if value.IsZero() {
		return "-"
	}

	return value.UTC().Format(time.RFC3339Nano)
}

func formatMessage(message string, attrs []slog.Attr) string {
	parts := []string{sanitizeMessageValue(message)}

	for _, attr := range attrs {
		attr.Value = attr.Value.Resolve()
		key := sanitizeKey(attr.Key)
		value := sanitizeMessageValue(formatValue(attr.Value))
		parts = append(parts, key+"="+value)
	}

	return strings.TrimSpace(strings.Join(parts, " "))
}

func formatValue(value slog.Value) string {
	if rendered, ok := formatPrimitiveValue(value); ok {
		return rendered
	}

	return formatFallbackValue(value)
}

func formatPrimitiveValue(value slog.Value) (string, bool) {
	formatter, ok := primitiveValueFormatters[value.Kind()]
	if !ok {
		return "", false
	}

	return formatter(value), true
}

func formatFallbackValue(value slog.Value) string {
	if value.Kind() == slog.KindAny {
		if err, ok := value.Any().(error); ok {
			return err.Error()
		}
	}

	return fmt.Sprint(value.Any())
}

func sanitizeHeaderValue(value string) string {
	value = sanitizeMessageValue(value)
	if value == "" {
		return "-"
	}

	return strings.Join(strings.Fields(value), "_")
}

func sanitizeKey(value string) string {
	value = sanitizeHeaderValue(value)
	if value == "-" {
		return "attr"
	}

	return value
}

func sanitizeMessageValue(value string) string {
	replacer := strings.NewReplacer("\r", " ", "\n", " ")
	return strings.TrimSpace(replacer.Replace(value))
}
