package logger

import (
	"io"
	"sync"
)

// LockedWriter serializes writes to the wrapped writer.
type LockedWriter struct {
	writer io.Writer
	mu     sync.Mutex
}

// NewLockedWriter wraps the provided writer with a mutex.
func NewLockedWriter(writer io.Writer) *LockedWriter {
	return &LockedWriter{writer: writer}
}

// Write writes one record while holding the mutex for the full call.
func (w *LockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.writer.Write(p)
}
