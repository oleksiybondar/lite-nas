package logger

// NopLogger is a Logger implementation that discards all log calls.
type NopLogger struct{}

// NewNop returns a logger that performs no work.
func NewNop() Logger {
	return NopLogger{}
}

func (NopLogger) Debug(string, ...any) {}

func (NopLogger) Info(string, ...any) {}

func (NopLogger) Warn(string, ...any) {}

func (NopLogger) Error(string, ...any) {}

func (NopLogger) With(...any) Logger {
	return NopLogger{}
}
