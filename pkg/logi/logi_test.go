package logi

import (
	"testing"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}

	// Test that version follows expected format (basic check)
	if len(Version) < 5 { // At minimum "v0.0.0" would be 6 chars, but we use "0.1.0" format
		t.Errorf("Version %q seems too short", Version)
	}

	expectedVersion := "0.2.0"
	if Version != expectedVersion {
		t.Errorf("Expected version %q, got %q", expectedVersion, Version)
	}
}

func TestLoggerInterface(t *testing.T) {
	// Test that both implementations satisfy the Logger interface
	var bufferLogger Logger = NewBufferLogger("test")
	var daemonLogger Logger = NewDemonLogger("test")

	// Test that we can call interface methods without compilation errors
	testLoggerInterface(t, bufferLogger, "BufferLogger")
	testLoggerInterface(t, daemonLogger, "DaemonLogger")
}

// Helper function to test Logger interface methods
func testLoggerInterface(t *testing.T, logger Logger, loggerType string) {
	// These should not panic if the interface is properly implemented
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s panicked during interface method call: %v", loggerType, r)
		}
	}()

	logger.Info("test info message")
	logger.Error("test error message")
	logger.Debug("test debug message")
	logger.Warn("test warn message")
}

func TestLoggerInterfaceWithKeyValues(t *testing.T) {
	// Test with BufferLogger to verify key-value pairs work through interface
	var logger Logger = NewBufferLogger("test")

	logger.Info("test message", "key1", "value1", "key2", 42)
	logger.Error("error message", "error", "test error")
	logger.Debug("debug message", "level", "verbose")
	logger.Warn("warn message", "severity", "high")

	// Cast back to BufferLogger to check messages
	bufferLogger := logger.(*BufferLogger)

	if len(bufferLogger.Messages) != 3 { // Debug won't be logged without DEBUG=1
		t.Errorf("Expected 3 messages, got %d", len(bufferLogger.Messages))
	}

	// Check that key-value pairs are properly formatted
	expectedMessages := []string{
		"INFO daemon=[test] test message key1=value1 key2=42",
		"ERROR daemon=[test] error message error=test error",
		"WARN daemon=[test] warn message severity=high",
	}

	for i, expected := range expectedMessages {
		if i < len(bufferLogger.Messages) && bufferLogger.Messages[i] != expected {
			t.Errorf("Message %d: expected %q, got %q", i, expected, bufferLogger.Messages[i])
		}
	}
}
