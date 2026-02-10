package logger

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger for structured logging
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New(service string) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	logger := slog.New(handler).With(
		slog.String("service", service),
	)

	return &Logger{Logger: logger}
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{Logger: l.Logger.With(args...)}
}
