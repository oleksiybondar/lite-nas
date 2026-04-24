package logger

import "log/slog"

type slogLogger struct {
	logger *slog.Logger
}

func newSlogLogger(base *slog.Logger) Logger {
	if base == nil {
		return NewNop()
	}

	return slogLogger{logger: base}
}

func (l slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l slogLogger) With(args ...any) Logger {
	return slogLogger{logger: l.logger.With(args...)}
}
