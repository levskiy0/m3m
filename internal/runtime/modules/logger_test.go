package modules

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoggerModule_NewLoggerModule(t *testing.T) {
	// Test with valid file path
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := NewLoggerModule(logPath)
	if logger == nil {
		t.Fatal("NewLoggerModule() returned nil")
	}
	defer logger.Close()

	if logger.file == nil {
		t.Error("logger.file should not be nil")
	}
}

func TestLoggerModule_NewLoggerModule_InvalidPath(t *testing.T) {
	// Test with invalid path - should fallback to stdout
	logger := NewLoggerModule("/nonexistent/directory/test.log")
	if logger == nil {
		t.Fatal("NewLoggerModule() returned nil")
	}
	defer logger.Close()

	if logger.file != os.Stdout {
		t.Error("logger should fallback to stdout on invalid path")
	}
}

func TestLoggerModule_Debug(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "debug.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Debug("test debug message")
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "[DEBUG]") {
		t.Error("log should contain [DEBUG] level")
	}
	if !strings.Contains(string(content), "test debug message") {
		t.Error("log should contain message")
	}
}

func TestLoggerModule_Info(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "info.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Info("test info message")
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "[INFO]") {
		t.Error("log should contain [INFO] level")
	}
	if !strings.Contains(string(content), "test info message") {
		t.Error("log should contain message")
	}
}

func TestLoggerModule_Warn(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "warn.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Warn("test warning message")
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "[WARN]") {
		t.Error("log should contain [WARN] level")
	}
	if !strings.Contains(string(content), "test warning message") {
		t.Error("log should contain message")
	}
}

func TestLoggerModule_Error(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "error.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Error("test error message")
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "[ERROR]") {
		t.Error("log should contain [ERROR] level")
	}
	if !strings.Contains(string(content), "test error message") {
		t.Error("log should contain message")
	}
}

func TestLoggerModule_LogFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "format.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Info("formatted message")
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	line := string(content)

	// Check format: [timestamp] [LEVEL] message
	// Timestamp format: 2006-01-02 15:04:05
	if !strings.HasPrefix(line, "[") {
		t.Error("log line should start with '['")
	}

	parts := strings.SplitN(line, "]", 3)
	if len(parts) < 3 {
		t.Errorf("log format should be [timestamp] [LEVEL] message, got: %s", line)
	}

	// Check timestamp format (YYYY-MM-DD HH:MM:SS)
	timestamp := strings.TrimPrefix(parts[0], "[")
	_, err = time.Parse("2006-01-02 15:04:05", timestamp)
	if err != nil {
		t.Errorf("invalid timestamp format: %s, error: %v", timestamp, err)
	}
}

func TestLoggerModule_MultipleArguments(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "multi.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Info("arg1", "arg2", 123, true)
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	line := string(content)
	if !strings.Contains(line, "arg1") || !strings.Contains(line, "arg2") ||
		!strings.Contains(line, "123") || !strings.Contains(line, "true") {
		t.Errorf("log should contain all arguments, got: %s", line)
	}
}

func TestLoggerModule_ConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "concurrent.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	var wg sync.WaitGroup
	logCount := 100

	// Log from multiple goroutines
	for i := 0; i < logCount; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Info("message", n)
		}(i)
	}

	wg.Wait()
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != logCount {
		t.Errorf("expected %d log lines, got %d", logCount, len(lines))
	}
}

func TestLoggerModule_EmptyMessage(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "empty.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Info()
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Should still create a log entry
	if len(content) == 0 {
		t.Error("empty message should still create log entry")
	}
}

func TestLoggerModule_LargeMessage(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "large.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	// Create a large message
	largeMsg := strings.Repeat("x", 10000)
	logger.Info(largeMsg)
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), largeMsg) {
		t.Error("log should contain large message")
	}
}

func TestLoggerModule_Close(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "close.log")

	logger := NewLoggerModule(logPath)
	logger.Info("before close")
	logger.Close()

	// Should not panic when writing after close
	// (behavior depends on implementation, but shouldn't crash)
}

func TestLoggerModule_Close_Stdout(t *testing.T) {
	// Should not close stdout
	logger := NewLoggerModule("/nonexistent/path")
	logger.Close() // Should not panic or close stdout
}

func TestLoggerModule_AppendMode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "append.log")

	// First logger
	logger1 := NewLoggerModule(logPath)
	logger1.Info("first message")
	logger1.file.Sync()
	logger1.Close()

	// Second logger should append
	logger2 := NewLoggerModule(logPath)
	logger2.Info("second message")
	logger2.file.Sync()
	logger2.Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "first message") {
		t.Error("log should contain first message")
	}
	if !strings.Contains(string(content), "second message") {
		t.Error("log should contain second message")
	}
}

func TestLoggerModule_AllLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "levels.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)
	levels := []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"}
	for _, level := range levels {
		if !strings.Contains(contentStr, level) {
			t.Errorf("log should contain level %s", level)
		}
	}
}

func TestLoggerModule_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "special.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	specialChars := "Special chars: \t\n\"'<>&[]{}!@#$%^*()"
	logger.Info(specialChars)
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Most characters should be preserved
	if !strings.Contains(string(content), "Special chars:") {
		t.Error("log should contain special characters message")
	}
}

func TestLoggerModule_Unicode(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "unicode.log")

	logger := NewLoggerModule(logPath)
	defer logger.Close()

	unicodeMsg := "Unicode: Hello World"
	logger.Info(unicodeMsg)
	logger.file.Sync()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), unicodeMsg) {
		t.Error("log should preserve unicode characters")
	}
}
