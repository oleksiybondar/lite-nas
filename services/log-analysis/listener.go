package main

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/workers"
)

// tailer owns one physical rsyslog output file, including rotation and
// truncation detection. Startup seeks EOF so historical lines do not trigger
// new security actions after service restart.
type tailer struct {
	path       string
	file       *os.File
	offset     int64
	pending    []byte
	discarding bool
}

// newTailer opens a configured file once. Tests can start from the beginning
// while production listeners establish a current-end baseline.
func newTailer(path string, startAtEnd bool) (*tailer, error) {
	file, err := os.Open(path) // #nosec G304 -- validated administrator source path
	if errors.Is(err, os.ErrNotExist) {
		// rsyslog may create auth.log only after the first matching record.
		return &tailer{path: path}, nil
	}
	if err != nil {
		return nil, err
	}
	tailer := &tailer{path: path, file: file}
	if startAtEnd {
		tailer.offset, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	return tailer, nil
}

// close releases the currently open inode after its listener stops.
func (t *tailer) close() error {
	if t.file == nil {
		return nil
	}
	return t.file.Close()
}

// readLines calls emit for every complete bounded line added since the last
// poll. It leaves a partial final line buffered for the next poll.
func (t *tailer) readLines(emit func(string) error) error {
	current, err := os.Stat(t.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if t.file == nil {
		t.file, err = os.Open(t.path) // #nosec G304 -- same configured source path
		if err != nil {
			return err
		}
	}
	opened, err := t.file.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(current, opened) || current.Size() < t.offset {
		file, err := os.Open(t.path) // #nosec G304 -- same configured source path
		if err != nil {
			return err
		}
		_ = t.file.Close()
		t.file = file
		t.offset = 0
		t.pending = nil
		t.discarding = false
	}
	if _, err := t.file.Seek(t.offset, io.SeekStart); err != nil {
		return err
	}
	var buffer [8192]byte
	for {
		count, err := t.file.Read(buffer[:])
		for _, char := range buffer[:count] {
			if char == '\n' {
				if t.discarding {
					t.discarding = false
					t.pending = nil
					continue
				}
				line := string(t.pending)
				t.pending = nil
				if line != "" {
					if sendErr := emit(line); sendErr != nil {
						return sendErr
					}
				}
				continue
			}
			if t.discarding {
				continue
			}
			if len(t.pending) >= maxLogLineBytes {
				t.pending = nil
				t.discarding = true
				continue
			}
			t.pending = append(t.pending, char)
		}
		t.offset += int64(count)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// logListener is a shared-worker-convention listener for one source. It owns
// a bounded send into the dispatch channel and closes no shared channels.
type logListener struct {
	source sourceSpec
	tailer *tailer
	output chan<- logRecord
	timer  workers.TimerWorker
	ticks  <-chan struct{}
	worker workers.BaseWorker
	done   chan struct{}
	log    sharedlogger.Logger
}

// newLogListener constructs a listener with a pre-opened tailer. No plugin
// constructor, registry or Go code loading is reachable from this type.
func newLogListener(source sourceSpec, interval time.Duration, output chan<- logRecord, log sharedlogger.Logger) (*logListener, error) {
	tailer, err := newTailer(source.Path, true)
	if err != nil {
		return nil, err
	}
	timer, ticks, err := workers.NewPollingTimerWorker(interval, 1)
	if err != nil {
		tailer.close()
		return nil, err
	}
	listener := &logListener{source: source, tailer: tailer, output: output, timer: timer, ticks: ticks, done: make(chan struct{}), log: log}
	listener.worker = workers.NewBaseWorker(listener.run)
	return listener, nil
}

// Start launches the timer and line reader under the service context.
func (l *logListener) Start(ctx context.Context) { l.timer.Start(ctx); l.worker.Start(ctx) }

// Done closes after the last read; the runtime then closes dispatch input.
func (l *logListener) Done() <-chan struct{} { return l.done }

func (l *logListener) run(ctx context.Context) {
	defer close(l.done)
	defer l.tailer.close()
	for {
		select {
		case <-ctx.Done():
			return
		case <-l.ticks:
			err := l.tailer.readLines(func(line string) error {
				record, err := parseAuthRecord(l.source.ID, line)
				if err != nil {
					l.log.Warn("invalid auth log record", "source", l.source.ID, "error", err)
					return nil
				}
				select {
				case l.output <- record:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
			if err != nil && ctx.Err() == nil {
				l.log.Error("log source read failed", "source", l.source.ID, "error", err)
			}
		}
	}
}
