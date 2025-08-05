package logi

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestDaemonLogger_NewDemonLogger(t *testing.T) {
	tests := []struct {
		name        string
		daemonName  string
		debugEnv    string
		silentEnv   string
		expectDebug bool
		expectLevel slog.Level
	}{
		{
			name:        "creates daemon logger with default settings",
			daemonName:  "test-daemon",
			debugEnv:    "",
			silentEnv:   "",
			expectDebug: false,
			expectLevel: slog.LevelInfo,
		},
		{
			name:        "creates daemon logger with debug enabled",
			daemonName:  "test-daemon",
			debugEnv:    "1",
			silentEnv:   "",
			expectDebug: true,
			expectLevel: slog.LevelDebug,
		},
		{
			name:        "creates daemon logger with silent mode",
			daemonName:  "test-daemon",
			debugEnv:    "",
			silentEnv:   "1",
			expectDebug: false,
			expectLevel: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("DEBUG")
			os.Unsetenv("SILENT_LOGS")

			// Set environment variables
			if tt.debugEnv != "" {
				os.Setenv("DEBUG", tt.debugEnv)
			}
			if tt.silentEnv != "" {
				os.Setenv("SILENT_LOGS", tt.silentEnv)
			}

			logger := NewDemonLogger(tt.daemonName)

			if logger.daemon != tt.daemonName {
				t.Errorf("Expected daemon name %q, got %q", tt.daemonName, logger.daemon)
			}

			if logger.debugOn != tt.expectDebug {
				t.Errorf("Expected debugOn %v, got %v", tt.expectDebug, logger.debugOn)
			}

			if logger.logger == nil {
				t.Error("Expected logger to be initialized")
			}

			// Clean up
			os.Unsetenv("DEBUG")
			os.Unsetenv("SILENT_LOGS")
		})
	}
}

func TestDaemonLogger_NewSilentLogger(t *testing.T) {
	// Clean environment first
	os.Unsetenv("SILENT_LOGS")

	logger := NewSilentLogger()

	if logger.daemon != "silent" {
		t.Errorf("Expected daemon name 'silent', got %q", logger.daemon)
	}

	// Check that SILENT_LOGS environment variable was set
	if os.Getenv("SILENT_LOGS") != "1" {
		t.Error("Expected SILENT_LOGS environment variable to be set to '1'")
	}

	// Clean up
	os.Unsetenv("SILENT_LOGS")
}

func TestDaemonLogger_LoggingMethods(t *testing.T) {
	// Redirect stdout to capture log output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a channel to read from the pipe
	done := make(chan string)
	go func() {
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	// Clean environment
	os.Unsetenv("DEBUG")
	os.Unsetenv("SILENT_LOGS")

	logger := NewDemonLogger("test")

	// Test Info logging
	logger.Info("info message", "key", "value")

	// Test Error logging
	logger.Error("error message", "error", "test error")

	// Test Warn logging
	logger.Warn("warn message", "level", "high")

	// Test Debug logging (should not appear without DEBUG=1)
	logger.Debug("debug message", "level", "verbose")

	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	output := <-done

	// Check that output contains expected log entries
	if !strings.Contains(output, "info message") {
		t.Error("Expected output to contain info message")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Expected output to contain error message")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Expected output to contain warn message")
	}
	// Debug message should not appear without DEBUG=1
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not appear without DEBUG=1")
	}
	// Check for daemon prefix
	if !strings.Contains(output, "daemon=[test]") {
		t.Error("Expected output to contain daemon prefix")
	}
}

func TestDaemonLogger_DebugLogging(t *testing.T) {
	// Redirect stdout to capture log output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a channel to read from the pipe
	done := make(chan string)
	go func() {
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	// Set DEBUG environment variable
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	logger := NewDemonLogger("test")

	// Test Debug logging with DEBUG=1
	logger.Debug("debug message", "level", "verbose")

	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	output := <-done

	// Check that debug message appears with DEBUG=1
	if !strings.Contains(output, "debug message") {
		t.Error("Expected debug message to appear with DEBUG=1")
	}
}

func TestDaemonLogger_SilentLogging(t *testing.T) {
	// Set SILENT_LOGS environment variable
	os.Setenv("SILENT_LOGS", "1")
	defer os.Unsetenv("SILENT_LOGS")

	// Capture stdout to ensure nothing is written
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	logger := NewDemonLogger("test")

	// Test logging methods with silent mode
	logger.Info("info message")
	logger.Error("error message")
	logger.Warn("warn message")

	// Close the writer and restore stdout
	w.Close()
	os.Stdout = oldStdout
	output := <-done

	// Output should be empty in silent mode
	if output != "" {
		t.Errorf("Expected no output in silent mode, got: %q", output)
	}
}

func TestDaemonLogger_SLogger(t *testing.T) {
	logger := NewDemonLogger("test")

	slogLogger := logger.SLogger()

	if slogLogger == nil {
		t.Error("Expected SLogger to return a non-nil *slog.Logger")
	}

	// Verify it's the same logger instance
	if slogLogger != logger.logger {
		t.Error("Expected SLogger to return the same logger instance")
	}
}

func TestDaemonLogger_ImplementsLoggerInterface(t *testing.T) {
	var logger Logger = NewDemonLogger("test")

	// Test that DaemonLogger implements the Logger interface
	// We'll use silent mode to avoid output during testing
	os.Setenv("SILENT_LOGS", "1")
	defer os.Unsetenv("SILENT_LOGS")

	// Create a new logger with silent mode
	logger = NewDemonLogger("test")

	// These should not panic if the interface is implemented correctly
	logger.Info("info test")
	logger.Error("error test")
	logger.Debug("debug test")
	logger.Warn("warn test")
}

func TestDaemonLogger_EnvironmentVariableHandling(t *testing.T) {
	tests := []struct {
		name      string
		debugEnv  string
		silentEnv string
		testFunc  func(t *testing.T, logger *DaemonLogger)
	}{
		{
			name:     "DEBUG case insensitive - uppercase",
			debugEnv: "1",
			testFunc: func(t *testing.T, logger *DaemonLogger) {
				if !logger.debugOn {
					t.Error("Expected debug to be enabled with DEBUG=1")
				}
			},
		},
		{
			name:      "SILENT_LOGS case insensitive - uppercase",
			silentEnv: "1",
			testFunc: func(t *testing.T, logger *DaemonLogger) {
				// We can't easily test if the output is discarded without complex setup,
				// but we can at least verify the logger was created without error
				if logger.logger == nil {
					t.Error("Expected logger to be created even in silent mode")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("DEBUG")
			os.Unsetenv("SILENT_LOGS")

			if tt.debugEnv != "" {
				os.Setenv("DEBUG", tt.debugEnv)
			}
			if tt.silentEnv != "" {
				os.Setenv("SILENT_LOGS", tt.silentEnv)
			}

			logger := NewDemonLogger("test")
			tt.testFunc(t, logger)

			// Clean up
			os.Unsetenv("DEBUG")
			os.Unsetenv("SILENT_LOGS")
		})
	}
}
