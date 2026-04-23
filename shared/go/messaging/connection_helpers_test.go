package messaging

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	sharedlogger "lite-nas/shared/logger"

	"github.com/nats-io/nats.go"
)

type logEntry struct {
	level string
	msg   string
	args  []any
}

type recordingLogger struct {
	mu      sync.Mutex
	entries []logEntry
}

func (l *recordingLogger) Debug(msg string, args ...any) {
	l.append("debug", msg, args...)
}

func (l *recordingLogger) Info(msg string, args ...any) {
	l.append("info", msg, args...)
}

func (l *recordingLogger) Warn(msg string, args ...any) {
	l.append("warn", msg, args...)
}

func (l *recordingLogger) Error(msg string, args ...any) {
	l.append("error", msg, args...)
}

func (l *recordingLogger) With(args ...any) sharedlogger.Logger {
	return l
}

func (l *recordingLogger) append(level string, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, logEntry{
		level: level,
		msg:   msg,
		args:  append([]any(nil), args...),
	})
}

func (l *recordingLogger) Entries() []logEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	copied := make([]logEntry, len(l.entries))
	copy(copied, l.entries)
	return copied
}

var _ sharedlogger.Logger = (*recordingLogger)(nil)

type connHandlerCapture struct {
	called bool
	log    sharedlogger.Logger
	conn   *nats.Conn
}

func (c *connHandlerCapture) record(log sharedlogger.Logger, nc *nats.Conn) {
	c.called = true
	c.log = log
	c.conn = nc
}

type connErrHandlerCapture struct {
	called bool
	log    sharedlogger.Logger
	conn   *nats.Conn
	err    error
}

func (c *connErrHandlerCapture) record(log sharedlogger.Logger, nc *nats.Conn, err error) {
	c.called = true
	c.log = log
	c.conn = nc
	c.err = err
}

type asyncErrHandlerCapture struct {
	called bool
	log    sharedlogger.Logger
	conn   *nats.Conn
	sub    *nats.Subscription
	err    error
}

func (c *asyncErrHandlerCapture) record(log sharedlogger.Logger, nc *nats.Conn, sub *nats.Subscription, err error) {
	c.called = true
	c.log = log
	c.conn = nc
	c.sub = sub
	c.err = err
}

func assertErrorIs(t *testing.T, err error, want error) {
	t.Helper()

	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func assertSingleLogEntry(t *testing.T, log *recordingLogger, want logEntry) {
	t.Helper()

	entries := log.Entries()
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	assertLogEntry(t, entries[0], want)
}

func assertLogEntry(t *testing.T, got logEntry, want logEntry) {
	t.Helper()

	if got.level != want.level {
		t.Fatalf("level = %q, want %q", got.level, want.level)
	}

	if got.msg != want.msg {
		t.Fatalf("msg = %q, want %q", got.msg, want.msg)
	}

	if !reflect.DeepEqual(got.args, want.args) {
		t.Fatalf("args = %#v, want %#v", got.args, want.args)
	}
}
