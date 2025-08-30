package logger

import (
	"log/slog"
	"os"
)

type Logger interface {
	Info(msg string, keyvals ...interface{})

	Warn(msg string, keyvals ...interface{})

	Error(msg string, keyvals ...interface{})

	Debug(msg string, keyvals ...interface{})
}

func New() Logger {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug, // minimum log level - set to debug to enable debug logs
		AddSource: true,            // include file + line number
	}
	handler := slog.NewJSONHandler(os.Stderr, opts)
	return slog.New(handler)
}
