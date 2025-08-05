package logi

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// BufferLogger stores log messages in memory for testing
type BufferLogger struct {
	daemon   string
	debugOn  bool
	Messages []string // Exported slice to store log messages
}

// NewBufferLogger creates a logger that stores messages in memory for testing purposes
func NewBufferLogger(daemonName string) *BufferLogger {
	debugOn := strings.ToLower(os.Getenv("DEBUG")) == "1"

	return &BufferLogger{
		daemon:   daemonName,
		debugOn:  debugOn,
		Messages: make([]string, 0),
	}
}

// Info logs an info message and stores it in Messages
func (l *BufferLogger) Info(msg string, keysAndValues ...any) {
	logEntry := l.formatMessage("INFO", msg, keysAndValues...)
	l.Messages = append(l.Messages, logEntry)
}

// Error logs an error message and stores it in Messages
func (l *BufferLogger) Error(msg string, keysAndValues ...any) {
	logEntry := l.formatMessage("ERROR", msg, keysAndValues...)
	l.Messages = append(l.Messages, logEntry)
}

// Debug logs a debug message and stores it in Messages (only if DEBUG=1)
func (l *BufferLogger) Debug(msg string, keysAndValues ...any) {
	if l.debugOn {
		logEntry := l.formatMessage("DEBUG", msg, keysAndValues...)
		l.Messages = append(l.Messages, logEntry)
	}
}

// Warn logs a warning message and stores it in Messages
func (l *BufferLogger) Warn(msg string, keysAndValues ...any) {
	logEntry := l.formatMessage("WARN", msg, keysAndValues...)
	l.Messages = append(l.Messages, logEntry)
}

// formatMessage formats a log message with key-value pairs
func (l *BufferLogger) formatMessage(level, msg string, keysAndValues ...any) string {
	entry := level + " daemon=[" + l.daemon + "] " + msg
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i]
			value := keysAndValues[i+1]
			entry += " " + key.(string) + "=" + formatValue(value)
		}
	}
	return entry
}

// formatValue formats different types of values for logging
func formatValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case time.Duration:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
