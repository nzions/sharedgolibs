package testicle

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger handles structured logging for testicle
type Logger struct {
	debug  bool
	logger *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(debug bool) *Logger {
	return &Logger{
		debug:  debug,
		logger: log.New(os.Stdout, "", 0),
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s", timestamp, message)
}

// Debug logs a debug message (only if debug mode is enabled)
func (l *Logger) Debug(format string, args ...interface{}) {
	if !l.debug {
		return
	}
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] DEBUG: %s", timestamp, message)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] ERROR: %s", timestamp, message)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] WARN: %s", timestamp, message)
}
