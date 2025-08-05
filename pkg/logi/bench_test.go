package logi

import (
	"os"
	"testing"
	"time"
)

func BenchmarkBufferLogger_Info(b *testing.B) {
	logger := NewBufferLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}

func BenchmarkBufferLogger_InfoWithMultipleKeyValues(b *testing.B) {
	logger := NewBufferLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			"iteration", i,
			"timestamp", time.Now(),
			"duration", 150*time.Millisecond,
			"status", "ok")
	}
}

func BenchmarkBufferLogger_Error(b *testing.B) {
	logger := NewBufferLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("benchmark error", "error", "test error", "code", 500)
	}
}

func BenchmarkBufferLogger_Debug_Disabled(b *testing.B) {
	// Ensure debug is disabled
	os.Unsetenv("DEBUG")
	logger := NewBufferLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message", "iteration", i)
	}
}

func BenchmarkBufferLogger_Debug_Enabled(b *testing.B) {
	// Enable debug
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	logger := NewBufferLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("debug message", "iteration", i)
	}
}

func BenchmarkDaemonLogger_Info(b *testing.B) {
	// Use silent mode to avoid I/O overhead in benchmarks
	os.Setenv("SILENT_LOGS", "1")
	defer os.Unsetenv("SILENT_LOGS")

	logger := NewDemonLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", "iteration", i)
	}
}

func BenchmarkDaemonLogger_Error(b *testing.B) {
	// Use silent mode to avoid I/O overhead in benchmarks
	os.Setenv("SILENT_LOGS", "1")
	defer os.Unsetenv("SILENT_LOGS")

	logger := NewDemonLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("benchmark error", "error", "test error", "code", 500)
	}
}

func BenchmarkFormatValue(b *testing.B) {
	testValues := []any{
		"string value",
		42,
		int64(9223372036854775807),
		5 * time.Second,
		true,
		3.14159,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testValues {
			_ = formatValue(val)
		}
	}
}

func BenchmarkBufferLogger_FormatMessage(b *testing.B) {
	logger := NewBufferLogger("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.formatMessage("INFO", "test message",
			"key1", "value1",
			"key2", i,
			"key3", 250*time.Millisecond)
	}
}
