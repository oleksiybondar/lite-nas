package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
)

// HandlerConfig defines the components used by Handler.
type HandlerConfig struct {
	Writer       io.Writer
	Formatter    Formatter
	MinimumLevel slog.Level
	Metadata     RecordMetadata
}

// Handler bridges slog records to project-local formatters and writers.
type Handler struct {
	writer       io.Writer
	formatter    Formatter
	minimumLevel slog.Level
	metadata     RecordMetadata
	attrs        []slog.Attr
	groups       []string
}

// NewHandler creates a Handler with the provided writer, formatter, and level filter.
func NewHandler(config HandlerConfig) (*Handler, error) {
	if config.Writer == nil {
		return nil, errors.New("handler writer is required")
	}

	if config.Formatter == nil {
		return nil, errors.New("handler formatter is required")
	}

	return &Handler{
		writer:       config.Writer,
		formatter:    config.Formatter,
		minimumLevel: config.MinimumLevel,
		metadata:     config.Metadata,
	}, nil
}

// Enabled reports whether the record level passes the minimum level filter.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minimumLevel
}

// Handle formats the record and writes it as one output record.
func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	attrs := make([]slog.Attr, 0, len(h.attrs)+record.NumAttrs())
	attrs = append(attrs, h.attrs...)
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, qualifyAttr(attr, h.groups)...)
		return true
	})

	payload, err := h.formatter.Format(record, attrs, h.metadata)
	if err != nil {
		return err
	}

	written, err := h.writer.Write(payload)
	if err != nil {
		return err
	}

	if written != len(payload) {
		return io.ErrShortWrite
	}

	return nil
}

// WithAttrs returns a new handler with additional attributes bound to it.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	qualified := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	qualified = append(qualified, h.attrs...)

	for _, attr := range attrs {
		qualified = append(qualified, qualifyAttr(attr, h.groups)...)
	}

	return &Handler{
		writer:       h.writer,
		formatter:    h.formatter,
		minimumLevel: h.minimumLevel,
		metadata:     h.metadata,
		attrs:        qualified,
		groups:       append([]string(nil), h.groups...),
	}
}

// WithGroup returns a new handler that prefixes later attributes with the group name.
func (h *Handler) WithGroup(name string) slog.Handler {
	groups := append([]string(nil), h.groups...)
	groups = append(groups, name)

	return &Handler{
		writer:       h.writer,
		formatter:    h.formatter,
		minimumLevel: h.minimumLevel,
		metadata:     h.metadata,
		attrs:        append([]slog.Attr(nil), h.attrs...),
		groups:       groups,
	}
}

func qualifyAttr(attr slog.Attr, groups []string) []slog.Attr {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return nil
	}

	if attr.Value.Kind() == slog.KindGroup {
		return qualifyGroupAttr(attr, groups)
	}

	return []slog.Attr{prefixAttrKey(attr, groups)}
}

func qualifyGroupAttr(attr slog.Attr, groups []string) []slog.Attr {
	qualified := make([]slog.Attr, 0, len(attr.Value.Group()))
	nextGroups := appendGroup(groups, attr.Key)

	for _, child := range attr.Value.Group() {
		qualified = append(qualified, qualifyAttr(child, nextGroups)...)
	}

	return qualified
}

func prefixAttrKey(attr slog.Attr, groups []string) slog.Attr {
	if len(groups) == 0 {
		return attr
	}

	attr.Key = strings.Join(append(append([]string(nil), groups...), attr.Key), ".")
	return attr
}

func appendGroup(groups []string, name string) []string {
	if name == "" {
		return groups
	}

	return append(append([]string(nil), groups...), name)
}
