package logger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type errFormatter struct {
	err error
}

func (f errFormatter) Format(slog.Record, []slog.Attr, RecordMetadata) ([]byte, error) {
	return nil, f.err
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	return len(p) - 1, nil
}

type errWriter struct {
	err error
}

func (w errWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestNewHandlerRejectsMissingDependencies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cfg  HandlerConfig
	}{
		{name: "missing writer", cfg: HandlerConfig{}},
		{name: "missing formatter", cfg: HandlerConfig{Writer: &bytes.Buffer{}}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if _, err := NewHandler(testCase.cfg); err == nil {
				t.Fatal("expected constructor error")
			}
		})
	}
}

func TestHandlerHandleReturnsExpectedError(t *testing.T) {
	t.Parallel()

	formatterErr := errors.New("format failed")
	writerErr := errors.New("write failed")

	testCases := []struct {
		name    string
		build   func(*testing.T) *Handler
		wantErr error
	}{
		{
			name: "formatter error",
			build: func(t *testing.T) *Handler {
				t.Helper()
				return newFormatterErrorHandler(t, formatterErr)
			},
			wantErr: formatterErr,
		},
		{
			name: "short write",
			build: func(t *testing.T) *Handler {
				t.Helper()
				return newShortWriteHandler(t)
			},
			wantErr: io.ErrShortWrite,
		},
		{
			name: "writer error",
			build: func(t *testing.T) *Handler {
				t.Helper()
				return newWriterErrorHandler(t, writerErr)
			},
			wantErr: writerErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			handler := testCase.build(t)
			record := slog.NewRecord(time.Time{}, slog.LevelInfo, "sample", 0)
			if err := handler.Handle(context.Background(), record); !errors.Is(err, testCase.wantErr) {
				t.Fatalf("Handle() error = %v, want %v", err, testCase.wantErr)
			}
		})
	}
}

func TestQualifyAttrHandlesGroupAndEmptyGroupName(t *testing.T) {
	t.Parallel()

	groupAttr := slog.Group("metrics", slog.String("cpu", "ok"))
	qualified := qualifyAttr(groupAttr, []string{"system"})
	if len(qualified) != 1 || qualified[0].Key != "system.metrics.cpu" {
		t.Fatalf("qualified group attr = %#v", qualified)
	}

	rootGroupAttr := slog.Attr{
		Value: slog.GroupValue(slog.String("cpu", "ok")),
	}
	rootQualified := qualifyAttr(rootGroupAttr, []string{"system"})
	if len(rootQualified) != 1 || rootQualified[0].Key != "system.cpu" {
		t.Fatalf("qualified root group attr = %#v", rootQualified)
	}
}

func TestQualifyAttrReturnsNilForEmptyAttr(t *testing.T) {
	t.Parallel()

	if qualified := qualifyAttr(slog.Attr{}, nil); qualified != nil {
		t.Fatalf("qualifyAttr() = %#v, want nil", qualified)
	}
}

func TestAppendGroupReturnsExistingGroupsWhenNameEmpty(t *testing.T) {
	t.Parallel()

	groups := []string{"system"}
	if got := appendGroup(groups, ""); len(got) != 1 || got[0] != "system" {
		t.Fatalf("appendGroup() = %#v", got)
	}
}

func TestNopLoggerConcreteMethodsAreCallable(t *testing.T) {
	t.Parallel()

	var logger NopLogger
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
}

func TestNewSlogLoggerReturnsNopForNilBase(t *testing.T) {
	t.Parallel()

	if _, ok := newSlogLogger(nil).(NopLogger); !ok {
		t.Fatalf("newSlogLogger(nil) returned %T, want NopLogger", newSlogLogger(nil))
	}
}

type internalStubFormatter struct{}

func (internalStubFormatter) Format(record slog.Record, attrs []slog.Attr, _ RecordMetadata) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(record.Message)
	for _, attr := range attrs {
		buffer.WriteString(attr.Key)
	}
	return buffer.Bytes(), nil
}

func mustNewInternalHandler(t *testing.T, cfg HandlerConfig) *Handler {
	t.Helper()

	handler, err := NewHandler(cfg)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	return handler
}

func newFormatterErrorHandler(t *testing.T, err error) *Handler {
	t.Helper()

	return mustNewInternalHandler(t, HandlerConfig{
		Writer:    io.Discard,
		Formatter: errFormatter{err: err},
	})
}

func newShortWriteHandler(t *testing.T) *Handler {
	t.Helper()

	return mustNewInternalHandler(t, HandlerConfig{
		Writer:    shortWriter{},
		Formatter: internalStubFormatter{},
	})
}

func newWriterErrorHandler(t *testing.T, err error) *Handler {
	t.Helper()

	return mustNewInternalHandler(t, HandlerConfig{
		Writer:    errWriter{err: err},
		Formatter: internalStubFormatter{},
	})
}
