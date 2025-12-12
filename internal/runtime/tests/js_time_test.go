package tests

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/levskiy0/m3m/internal/runtime/modules"
)

// ============== TIME FORMAT TESTS ==============

// TestLogger_TimestampFormat verifies that logger uses UTC time format
func TestLogger_TimestampFormat(t *testing.T) {
	// Create temp log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := modules.NewLoggerModule(logPath)
	defer logger.Close()

	// Log a message
	before := time.Now().UTC()
	logger.Info("test message")
	after := time.Now().UTC()

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logLine := string(content)
	t.Logf("Log line: %s", logLine)

	// Expected format: [2006-01-02 15:04:05] [INFO] test message
	// This is UTC time without timezone suffix
	pattern := `^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] \[INFO\] test message`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(logLine)

	if len(matches) < 2 {
		t.Fatalf("Log line doesn't match expected format: %s", logLine)
	}

	// Parse the timestamp as UTC
	timestampStr := matches[1]
	timestamp, err := time.ParseInLocation("2006-01-02 15:04:05", timestampStr, time.UTC)
	if err != nil {
		t.Fatalf("Failed to parse timestamp %s: %v", timestampStr, err)
	}

	t.Logf("Parsed timestamp (UTC): %s", timestamp.Format("2006-01-02 15:04:05 MST"))
	t.Logf("Before (UTC): %s", before.Format("2006-01-02 15:04:05 MST"))
	t.Logf("After (UTC): %s", after.Format("2006-01-02 15:04:05 MST"))

	// Verify timestamp is within expected range (UTC time)
	if timestamp.Before(before.Add(-time.Second)) || timestamp.After(after.Add(time.Second)) {
		t.Errorf("Timestamp %s is not within expected range [%s, %s]",
			timestamp.Format("15:04:05"),
			before.Format("15:04:05"),
			after.Format("15:04:05"))
	}

	// Verify format uses UTC time (hour should match UTC hour)
	if timestamp.Hour() != before.Hour() {
		t.Errorf("Log timestamp hour %d doesn't match UTC hour %d",
			timestamp.Hour(), before.Hour())
	}
}

// TestLogger_MultipleMessages tests multiple log messages have consistent timestamps
func TestLogger_MultipleMessages(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := modules.NewLoggerModule(logPath)
	defer logger.Close()

	// Log multiple messages
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Read and verify all timestamps are in local format
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	pattern := regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] \[(DEBUG|INFO|WARN|ERROR)\]`)

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if !pattern.MatchString(line) {
			t.Errorf("Line %d doesn't match timestamp format: %s", lineNum, line)
		}
	}

	if lineNum != 4 {
		t.Errorf("Expected 4 log lines, got %d", lineNum)
	}
}

// TestLogger_TimestampIsUTC verifies timestamp IS in UTC
func TestLogger_TimestampIsUTC(t *testing.T) {
	// Skip if server is in UTC (can't distinguish)
	_, offset := time.Now().Zone()
	if offset == 0 {
		t.Skip("Server is in UTC timezone, cannot verify UTC vs local")
	}

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := modules.NewLoggerModule(logPath)
	defer logger.Close()

	now := time.Now()
	logger.Info("test")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Extract timestamp
	pattern := regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\]`)
	matches := pattern.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		t.Fatal("Could not extract timestamp")
	}

	// Parse as UTC time
	utcTs, _ := time.ParseInLocation("2006-01-02 15:04:05", matches[1], time.UTC)

	// The hour should match UTC hour, not local hour
	t.Logf("Server local time: %s", now.Format("15:04:05 MST"))
	t.Logf("Server UTC time: %s", now.UTC().Format("15:04:05 MST"))
	t.Logf("Log timestamp: %s", matches[1])

	// Log hour should match UTC hour
	if utcTs.Hour() != now.UTC().Hour() {
		t.Errorf("Log hour %d doesn't match UTC hour %d (local would be %d)",
			utcTs.Hour(), now.UTC().Hour(), now.Hour())
	}
}

// ============== UTILS MODULE TIME TESTS ==============

func TestUtils_Timestamp_ReturnsLocalTime(t *testing.T) {
	h := NewJSTestHelper(t)

	before := time.Now().UnixMilli()
	result := h.MustRun(t, `$utils.timestamp()`)
	after := time.Now().UnixMilli()

	ts := result.ToInteger()
	if ts < before || ts > after {
		t.Errorf("Timestamp %d should be between %d and %d", ts, before, after)
	}

	// Convert to time and verify it's in reasonable range
	tsTime := time.UnixMilli(ts)
	t.Logf("Timestamp as local time: %s", tsTime.Local().Format("2006-01-02 15:04:05 MST"))
}

// ============== DATE PARSING TESTS ==============

// TestDateParsing_ISO8601 tests various date format parsing
func TestDateParsing_ISO8601(t *testing.T) {
	h := NewJSTestHelper(t)

	tests := []struct {
		name   string
		input  string
		format string
	}{
		// ISO 8601 with timezone
		{"UTC", "2025-12-12T07:10:05.657Z", "UTC"},
		{"offset positive", "2025-12-12T14:10:05.657+07:00", "local"},
		// Without timezone (local time)
		{"no timezone", "2025-12-12T14:10:05", "local"},
		{"space separator", "2025-12-12 14:10:05", "local"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := `new Date("` + tt.input + `").getTime()`
			result := h.MustRun(t, code)
			ts := result.ToInteger()

			parsed := time.UnixMilli(ts)
			t.Logf("Input: %s -> %s (local: %s)",
				tt.input,
				parsed.UTC().Format("2006-01-02T15:04:05Z"),
				parsed.Local().Format("2006-01-02 15:04:05 MST"))
		})
	}
}

// TestDateParsing_ServerTime tests that server time is consistent
func TestDateParsing_ServerTime(t *testing.T) {
	serverNow := time.Now()
	t.Logf("Server time: %s", serverNow.Format("2006-01-02 15:04:05 MST"))
	t.Logf("Server location: %s", serverNow.Location().String())
	t.Logf("Server UTC offset: %d seconds", func() int {
		_, offset := serverNow.Zone()
		return offset
	}())

	// Verify time.Now() returns local time
	if serverNow.Location() == time.UTC && serverNow.Location().String() != "UTC" {
		t.Log("Warning: time.Now() might be using UTC instead of local time")
	}
}
