package logi

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// DaemonLogger wraps slog with daemon name prefix and debug level control
type DaemonLogger struct {
	logger  *slog.Logger
	daemon  string
	debugOn bool
}

// NewSilentLogger creates a logger that discards all logs
func NewSilentLogger() *DaemonLogger {
	os.Setenv("SILENT_LOGS", "1")
	return NewDemonLogger("silent")
}

// NewDemonLogger creates a new logger for the specified daemon name
// Debug logging is enabled if env var DEBUG=1
func NewDemonLogger(daemonName string) *DaemonLogger {
	silent := strings.ToLower(os.Getenv("SILENT_LOGS")) == "1"
	debugOn := strings.ToLower(os.Getenv("DEBUG")) == "1"

	level := slog.LevelInfo
	if debugOn {
		level = slog.LevelDebug
	}

	var handlerOut io.Writer = os.Stdout
	if silent {
		handlerOut = io.Discard
	}

	handler := slog.NewTextHandler(handlerOut, &slog.HandlerOptions{
		Level:     level,
		AddSource: false,
	})

	logger := slog.New(handler).With("daemon", "["+daemonName+"]")

	return &DaemonLogger{
		logger:  logger,
		daemon:  daemonName,
		debugOn: debugOn,
	}
}

// Info logs an info message with daemon prefix
func (l *DaemonLogger) Info(msg string, keysAndValues ...any) {
	l.logger.Info(msg, keysAndValues...)
}

// Debug logs a debug message with daemon prefix (only if DEBUG=1)
func (l *DaemonLogger) Debug(msg string, keysAndValues ...any) {
	if l.debugOn {
		l.logger.Debug(msg, keysAndValues...)
	}
}

// Error logs an error message with daemon prefix
func (l *DaemonLogger) Error(msg string, keysAndValues ...any) {
	l.logger.Error(msg, keysAndValues...)
}

// Warn logs a warning message with daemon prefix
func (l *DaemonLogger) Warn(msg string, keysAndValues ...any) {
	l.logger.Warn(msg, keysAndValues...)
}

// SLogger returns the underlying *slog.Logger with daemon prefix
func (l *DaemonLogger) SLogger() *slog.Logger {
	return l.logger
}
