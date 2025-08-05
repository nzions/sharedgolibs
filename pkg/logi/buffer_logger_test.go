package logi

import (
	"os"
	"testing"
	"time"
)

func TestBufferLogger_NewBufferLogger(t *testing.T) {
	tests := []struct {
		name        string
		daemonName  string
		debugEnv    string
		expectDebug bool
	}{
		{
			name:        "creates buffer logger with debug disabled",
			daemonName:  "test-daemon",
			debugEnv:    "",
			expectDebug: false,
		},
		{
			name:        "creates buffer logger with debug enabled",
			daemonName:  "test-daemon",
			debugEnv:    "1",
			expectDebug: true,
		},
		{
			name:        "creates buffer logger with debug enabled case insensitive",
			daemonName:  "test-daemon",
			debugEnv:    "1",
			expectDebug: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.debugEnv != "" {
				os.Setenv("DEBUG", tt.debugEnv)
				defer os.Unsetenv("DEBUG")
			}

			logger := NewBufferLogger(tt.daemonName)

			if logger.daemon != tt.daemonName {
				t.Errorf("Expected daemon name %q, got %q", tt.daemonName, logger.daemon)
			}

			if logger.debugOn != tt.expectDebug {
				t.Errorf("Expected debugOn %v, got %v", tt.expectDebug, logger.debugOn)
			}

			if logger.Messages == nil {
				t.Error("Expected Messages slice to be initialized")
			}

			if len(logger.Messages) != 0 {
				t.Errorf("Expected empty Messages slice, got %d items", len(logger.Messages))
			}
		})
	}
}

func TestBufferLogger_Info(t *testing.T) {
	logger := NewBufferLogger("test")

	logger.Info("test message")

	if len(logger.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(logger.Messages))
	}

	expected := "INFO daemon=[test] test message"
	if logger.Messages[0] != expected {
		t.Errorf("Expected %q, got %q", expected, logger.Messages[0])
	}
}

func TestBufferLogger_InfoWithKeyValues(t *testing.T) {
	logger := NewBufferLogger("test")

	logger.Info("test message", "key1", "value1", "key2", 42)

	if len(logger.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(logger.Messages))
	}

	expected := "INFO daemon=[test] test message key1=value1 key2=42"
	if logger.Messages[0] != expected {
		t.Errorf("Expected %q, got %q", expected, logger.Messages[0])
	}
}

func TestBufferLogger_Error(t *testing.T) {
	logger := NewBufferLogger("test")

	logger.Error("error message", "error", "something went wrong")

	if len(logger.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(logger.Messages))
	}

	expected := "ERROR daemon=[test] error message error=something went wrong"
	if logger.Messages[0] != expected {
		t.Errorf("Expected %q, got %q", expected, logger.Messages[0])
	}
}

func TestBufferLogger_Warn(t *testing.T) {
	logger := NewBufferLogger("test")

	logger.Warn("warning message", "level", "high")

	if len(logger.Messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(logger.Messages))
	}

	expected := "WARN daemon=[test] warning message level=high"
	if logger.Messages[0] != expected {
		t.Errorf("Expected %q, got %q", expected, logger.Messages[0])
	}
}

func TestBufferLogger_Debug(t *testing.T) {
	tests := []struct {
		name        string
		debugEnv    string
		expectCount int
	}{
		{
			name:        "debug disabled - no message stored",
			debugEnv:    "",
			expectCount: 0,
		},
		{
			name:        "debug enabled - message stored",
			debugEnv:    "1",
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.debugEnv != "" {
				os.Setenv("DEBUG", tt.debugEnv)
				defer os.Unsetenv("DEBUG")
			} else {
				os.Unsetenv("DEBUG")
			}

			logger := NewBufferLogger("test")
			logger.Debug("debug message", "level", "verbose")

			if len(logger.Messages) != tt.expectCount {
				t.Errorf("Expected %d messages, got %d", tt.expectCount, len(logger.Messages))
			}

			if tt.expectCount > 0 {
				expected := "DEBUG daemon=[test] debug message level=verbose"
				if logger.Messages[0] != expected {
					t.Errorf("Expected %q, got %q", expected, logger.Messages[0])
				}
			}
		})
	}
}

func TestBufferLogger_MultipleMessages(t *testing.T) {
	logger := NewBufferLogger("test")

	logger.Info("first message")
	logger.Error("second message")
	logger.Warn("third message")

	if len(logger.Messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(logger.Messages))
	}

	expectedMessages := []string{
		"INFO daemon=[test] first message",
		"ERROR daemon=[test] second message",
		"WARN daemon=[test] third message",
	}

	for i, expected := range expectedMessages {
		if logger.Messages[i] != expected {
			t.Errorf("Message %d: expected %q, got %q", i, expected, logger.Messages[i])
		}
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "string value",
			input:    "test string",
			expected: "test string",
		},
		{
			name:     "int value",
			input:    42,
			expected: "42",
		},
		{
			name:     "int64 value",
			input:    int64(9223372036854775807),
			expected: "9223372036854775807",
		},
		{
			name:     "duration value",
			input:    5 * time.Second,
			expected: "5s",
		},
		{
			name:     "bool value",
			input:    true,
			expected: "true",
		},
		{
			name:     "float value",
			input:    3.14,
			expected: "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBufferLogger_formatMessage(t *testing.T) {
	logger := NewBufferLogger("test-daemon")

	tests := []struct {
		name          string
		level         string
		msg           string
		keysAndValues []any
		expected      string
	}{
		{
			name:          "message without key-value pairs",
			level:         "INFO",
			msg:           "simple message",
			keysAndValues: nil,
			expected:      "INFO daemon=[test-daemon] simple message",
		},
		{
			name:          "message with one key-value pair",
			level:         "ERROR",
			msg:           "error occurred",
			keysAndValues: []any{"error", "file not found"},
			expected:      "ERROR daemon=[test-daemon] error occurred error=file not found",
		},
		{
			name:          "message with multiple key-value pairs",
			level:         "DEBUG",
			msg:           "processing data",
			keysAndValues: []any{"count", 5, "duration", 250 * time.Millisecond},
			expected:      "DEBUG daemon=[test-daemon] processing data count=5 duration=250ms",
		},
		{
			name:          "message with odd number of arguments (last key ignored)",
			level:         "WARN",
			msg:           "incomplete data",
			keysAndValues: []any{"key1", "value1", "key2"},
			expected:      "WARN daemon=[test-daemon] incomplete data key1=value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.formatMessage(tt.level, tt.msg, tt.keysAndValues...)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBufferLogger_ImplementsLoggerInterface(t *testing.T) {
	var logger Logger = NewBufferLogger("test")

	// Test that BufferLogger implements the Logger interface
	logger.Info("info test")
	logger.Error("error test")
	logger.Debug("debug test")
	logger.Warn("warn test")

	bufferLogger := logger.(*BufferLogger)
	if len(bufferLogger.Messages) != 3 { // Debug won't be logged without DEBUG=1
		t.Errorf("Expected 3 messages, got %d", len(bufferLogger.Messages))
	}
}
