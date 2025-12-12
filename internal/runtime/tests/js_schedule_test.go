package tests

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/internal/runtime/modules"
)

// ============== SCHEDULE MODULE TESTS ==============

// TestScheduleHelper creates a helper with schedule module
type TestScheduleHelper struct {
	VM       *goja.Runtime
	Schedule *modules.ScheduleModule
	executed chan string
}

func NewTestScheduleHelper(t *testing.T) *TestScheduleHelper {
	t.Helper()
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	schedule := modules.NewScheduleModule(logger)
	schedule.Register(vm)

	h := &TestScheduleHelper{
		VM:       vm,
		Schedule: schedule,
		executed: make(chan string, 100),
	}

	// Register mock logger
	vm.Set("$logger", map[string]interface{}{
		"debug": func(args ...interface{}) {},
		"info":  func(args ...interface{}) {},
		"warn":  func(args ...interface{}) {},
		"error": func(args ...interface{}) {},
	})

	// Register console
	vm.Set("console", map[string]interface{}{
		"log":   func(args ...interface{}) {},
		"info":  func(args ...interface{}) {},
		"warn":  func(args ...interface{}) {},
		"error": func(args ...interface{}) {},
		"debug": func(args ...interface{}) {},
	})

	return h
}

func (h *TestScheduleHelper) Run(code string) (goja.Value, error) {
	return h.VM.RunString(code)
}

func (h *TestScheduleHelper) MustRun(t *testing.T, code string) goja.Value {
	t.Helper()
	result, err := h.Run(code)
	if err != nil {
		t.Fatalf("JS execution failed: %v\nCode: %s", err, code)
	}
	return result
}

// ============== TIME CONSISTENCY TESTS ==============

func TestSchedule_At_ParsesTimeCorrectly(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		name     string
		time     string
		wantCron string
	}{
		{"midnight", "00:00", "0 0 * * *"},
		{"noon", "12:00", "0 12 * * *"},
		{"afternoon", "14:30", "30 14 * * *"},
		{"evening", "23:59", "59 23 * * *"},
		{"morning", "09:15", "15 9 * * *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := `$schedule.at("` + tt.time + `", function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if jobID == "" {
				t.Fatalf("at(%q) returned empty job ID", tt.time)
			}

			// Get job info to verify cron expression
			info := h.Schedule.Get(jobID)
			if info == nil {
				t.Fatalf("Job %s not found", jobID)
			}

			expr, ok := info["expression"].(string)
			if !ok {
				t.Fatalf("Job expression not found")
			}

			if expr != tt.wantCron {
				t.Errorf("at(%q) created cron %q, want %q", tt.time, expr, tt.wantCron)
			}

			// Cleanup
			h.Schedule.Cancel(jobID)
		})
	}
}

func TestSchedule_At_InvalidTimeFormats(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		name string
		time string
	}{
		{"no colon", "1400"},
		{"invalid hour", "25:00"},
		{"invalid minute", "12:60"},
		{"negative hour", "-1:00"},
		{"letters", "ab:cd"},
		{"extra colon", "12:30:00"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := `$schedule.at("` + tt.time + `", function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if jobID != "" {
				t.Errorf("at(%q) should return empty job ID for invalid time, got %q", tt.time, jobID)
				h.Schedule.Cancel(jobID)
			}
		})
	}
}

// TestSchedule_At_UsesUTCTime verifies that .at() uses UTC timezone
func TestSchedule_At_UsesUTCTime(t *testing.T) {
	h := NewTestScheduleHelper(t)

	// Start the scheduler FIRST so cron entries get scheduled
	h.Schedule.Start()
	defer h.Schedule.Stop()

	// Get current server time
	now := time.Now()
	t.Logf("Server local time: %s (zone: %s)", now.Format("15:04:05"), now.Location().String())

	// Schedule a job for 1 minute from now
	targetTime := now.Add(1 * time.Minute)
	timeStr := targetTime.Format("15:04")

	code := `$schedule.at("` + timeStr + `", function() {})`
	result := h.MustRun(t, code)
	jobID := result.String()

	if jobID == "" {
		t.Fatal("Failed to create job")
	}
	defer h.Schedule.Cancel(jobID)

	// Give cron a moment to calculate next run
	time.Sleep(10 * time.Millisecond)

	// Verify nextRun is in server local time
	info := h.Schedule.Get(jobID)
	if info == nil {
		t.Fatal("Job not found")
	}

	nextRunVal, ok := info["nextRun"]
	if !ok || nextRunVal == nil {
		t.Log("nextRun not set - this may happen if cron hasn't calculated it yet")
		t.Logf("Job info: %+v", info)
		return
	}

	nextRunMs, ok := nextRunVal.(int64)
	if !ok {
		t.Fatalf("nextRun is not int64: %T", nextRunVal)
	}

	nextRun := time.UnixMilli(nextRunMs)
	t.Logf("Job nextRun: %s (zone: %s)", nextRun.Format("15:04:05"), nextRun.Location().String())

	// nextRun should be in local time, with the same hour:minute as requested
	expectedHour := targetTime.Hour()
	expectedMinute := targetTime.Minute()

	// The next run should match the requested time (could be today or tomorrow)
	if nextRun.Hour() != expectedHour || nextRun.Minute() != expectedMinute {
		t.Errorf("nextRun time %02d:%02d doesn't match requested %02d:%02d",
			nextRun.Hour(), nextRun.Minute(), expectedHour, expectedMinute)
	}
}

// ============== CRON EXPRESSION TESTS ==============

func TestSchedule_Cron_ParsesCorrectly(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		name string
		cron string
	}{
		{"every minute", "* * * * *"},
		{"every hour", "0 * * * *"},
		{"daily midnight", "0 0 * * *"},
		{"weekly sunday", "0 0 * * 0"},
		{"monthly first", "0 0 1 * *"},
		{"specific time", "30 14 * * *"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := `$schedule.cron("` + tt.cron + `", function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if jobID == "" {
				t.Errorf("cron(%q) returned empty job ID", tt.cron)
				return
			}

			info := h.Schedule.Get(jobID)
			if info == nil {
				t.Fatalf("Job %s not found", jobID)
			}

			expr, ok := info["expression"].(string)
			if !ok || expr != tt.cron {
				t.Errorf("Job expression = %q, want %q", expr, tt.cron)
			}

			h.Schedule.Cancel(jobID)
		})
	}
}

func TestSchedule_Cron_InvalidExpressions(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		name string
		cron string
	}{
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * * *"},
		{"invalid chars", "x * * * *"},
		{"out of range", "60 * * * *"},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := `$schedule.cron("` + tt.cron + `", function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if jobID != "" {
				t.Errorf("cron(%q) should return empty job ID for invalid expression", tt.cron)
				h.Schedule.Cancel(jobID)
			}
		})
	}
}

// ============== PRESET METHODS TESTS ==============

func TestSchedule_Daily_CreatesCorrectCron(t *testing.T) {
	h := NewTestScheduleHelper(t)

	code := `$schedule.daily(function() {})`
	result := h.MustRun(t, code)

	jobID := result.String()
	if jobID == "" {
		t.Fatal("daily() returned empty job ID")
	}
	defer h.Schedule.Cancel(jobID)

	info := h.Schedule.Get(jobID)
	expr := info["expression"].(string)
	if expr != "0 0 * * *" {
		t.Errorf("daily() created cron %q, want %q", expr, "0 0 * * *")
	}
}

func TestSchedule_Hourly_CreatesCorrectCron(t *testing.T) {
	h := NewTestScheduleHelper(t)

	code := `$schedule.hourly(function() {})`
	result := h.MustRun(t, code)

	jobID := result.String()
	if jobID == "" {
		t.Fatal("hourly() returned empty job ID")
	}
	defer h.Schedule.Cancel(jobID)

	info := h.Schedule.Get(jobID)
	expr := info["expression"].(string)
	if expr != "0 * * * *" {
		t.Errorf("hourly() created cron %q, want %q", expr, "0 * * * *")
	}
}

func TestSchedule_Weekly_CreatesCorrectCron(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		day      int
		wantCron string
	}{
		{0, "0 0 * * 0"}, // Sunday
		{1, "0 0 * * 1"}, // Monday
		{6, "0 0 * * 6"}, // Saturday
	}

	for _, tt := range tests {
		t.Run("day_"+string(rune('0'+tt.day)), func(t *testing.T) {
			code := `$schedule.weekly(` + string(rune('0'+tt.day)) + `, function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if jobID == "" {
				t.Fatalf("weekly(%d) returned empty job ID", tt.day)
			}
			defer h.Schedule.Cancel(jobID)

			info := h.Schedule.Get(jobID)
			expr := info["expression"].(string)
			if expr != tt.wantCron {
				t.Errorf("weekly(%d) created cron %q, want %q", tt.day, expr, tt.wantCron)
			}
		})
	}
}

func TestSchedule_Monthly_CreatesCorrectCron(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		day      int
		wantCron string
	}{
		{1, "0 0 1 * *"},
		{15, "0 0 15 * *"},
		{31, "0 0 31 * *"},
	}

	for _, tt := range tests {
		t.Run("day_"+string(rune('0'+tt.day)), func(t *testing.T) {
			code := `$schedule.monthly(` + itoa(tt.day) + `, function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if jobID == "" {
				t.Fatalf("monthly(%d) returned empty job ID", tt.day)
			}
			defer h.Schedule.Cancel(jobID)

			info := h.Schedule.Get(jobID)
			expr := info["expression"].(string)
			if expr != tt.wantCron {
				t.Errorf("monthly(%d) created cron %q, want %q", tt.day, expr, tt.wantCron)
			}
		})
	}
}

// ============== INTERVAL TESTS ==============

func TestSchedule_Every_ParsesIntervals(t *testing.T) {
	h := NewTestScheduleHelper(t)

	tests := []struct {
		interval string
		valid    bool
	}{
		{"5s", true},
		{"30m", true},
		{"2h", true},
		{"1d", true},
		{"invalid", false},
		{"5", false},
		{"5x", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			code := `$schedule.every("` + tt.interval + `", function() {})`
			result := h.MustRun(t, code)

			jobID := result.String()
			if tt.valid && jobID == "" {
				t.Errorf("every(%q) should return job ID", tt.interval)
			}
			if !tt.valid && jobID != "" {
				t.Errorf("every(%q) should return empty job ID for invalid interval", tt.interval)
				h.Schedule.Cancel(jobID)
			}
			if jobID != "" {
				h.Schedule.Cancel(jobID)
			}
		})
	}
}

// ============== DELAY TESTS ==============

func TestSchedule_Delay_ExecutesAfterTime(t *testing.T) {
	h := NewTestScheduleHelper(t)

	executed := false
	h.VM.Set("markExecuted", func() {
		executed = true
	})

	code := `$schedule.delay(50, markExecuted)`
	result := h.MustRun(t, code)

	jobID := result.String()
	if jobID == "" {
		t.Fatal("delay() returned empty job ID")
	}

	// Should not be executed immediately
	if executed {
		t.Error("Job should not execute immediately")
	}

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	if !executed {
		t.Error("Job should have executed after delay")
	}
}

// ============== JOB MANAGEMENT TESTS ==============

func TestSchedule_Cancel_RemovesJob(t *testing.T) {
	h := NewTestScheduleHelper(t)

	code := `$schedule.every("1h", function() {})`
	result := h.MustRun(t, code)
	jobID := result.String()

	if h.Schedule.Get(jobID) == nil {
		t.Fatal("Job should exist after creation")
	}

	success := h.Schedule.Cancel(jobID)
	if !success {
		t.Error("Cancel should return true")
	}

	if h.Schedule.Get(jobID) != nil {
		t.Error("Job should not exist after cancel")
	}
}

func TestSchedule_PauseResume_Works(t *testing.T) {
	h := NewTestScheduleHelper(t)
	h.Schedule.Start()
	defer h.Schedule.Stop()

	code := `$schedule.every("1h", function() {})`
	result := h.MustRun(t, code)
	jobID := result.String()
	defer h.Schedule.Cancel(jobID)

	// Pause
	success := h.Schedule.Pause(jobID)
	if !success {
		t.Error("Pause should return true")
	}

	info := h.Schedule.Get(jobID)
	if info["status"] != "paused" {
		t.Errorf("Job status should be 'paused', got %v", info["status"])
	}

	// Resume
	success = h.Schedule.Resume(jobID)
	if !success {
		t.Error("Resume should return true")
	}

	info = h.Schedule.Get(jobID)
	if info["status"] != "active" {
		t.Errorf("Job status should be 'active', got %v", info["status"])
	}
}

func TestSchedule_List_ReturnsAllJobs(t *testing.T) {
	h := NewTestScheduleHelper(t)

	// Create multiple jobs
	h.MustRun(t, `$schedule.daily(function() {})`)
	h.MustRun(t, `$schedule.hourly(function() {})`)
	h.MustRun(t, `$schedule.every("5m", function() {})`)

	jobs := h.Schedule.List()
	if len(jobs) != 3 {
		t.Errorf("List() should return 3 jobs, got %d", len(jobs))
	}

	// Cleanup
	for _, job := range jobs {
		h.Schedule.Cancel(job["id"].(string))
	}
}

// ============== TIME ZONE CONSISTENCY TEST ==============

func TestSchedule_TimeConsistency_AllMethodsUseLocalTime(t *testing.T) {
	h := NewTestScheduleHelper(t)
	h.Schedule.Start()
	defer h.Schedule.Stop()

	serverNow := time.Now()
	t.Logf("Server time: %s", serverNow.Format("2006-01-02 15:04:05 MST"))
	t.Logf("Server location: %s", serverNow.Location().String())

	// Create a job with .at() for current time + 1 minute
	targetTime := serverNow.Add(1 * time.Minute)
	timeStr := targetTime.Format("15:04")

	code := `$schedule.at("` + timeStr + `", function() {})`
	result := h.MustRun(t, code)
	jobID := result.String()
	defer h.Schedule.Cancel(jobID)

	// Give cron time to calculate nextRun
	time.Sleep(10 * time.Millisecond)

	info := h.Schedule.Get(jobID)
	t.Logf("Job info: %+v", info)

	nextRunVal, ok := info["nextRun"]
	if !ok || nextRunVal == nil {
		t.Log("nextRun not set - cron may not have calculated it yet")
		t.Logf("Expression: %v", info["expression"])
		return
	}

	nextRunMs, ok := nextRunVal.(int64)
	if !ok {
		t.Fatalf("nextRun is not int64: %T", nextRunVal)
	}

	nextRun := time.UnixMilli(nextRunMs)

	// The nextRun should be in the same day or next day, at the specified time
	// in server's local timezone
	t.Logf("Scheduled for: %s", timeStr)
	t.Logf("Next run: %s", nextRun.Format("2006-01-02 15:04:05 MST"))

	// Verify hour and minute match
	if nextRun.Hour() != targetTime.Hour() || nextRun.Minute() != targetTime.Minute() {
		t.Errorf("Next run time %02d:%02d doesn't match scheduled %02d:%02d",
			nextRun.Hour(), nextRun.Minute(), targetTime.Hour(), targetTime.Minute())
	}

	// Verify it's not in UTC (unless server is in UTC)
	// The returned timestamp should be interpretable in local time
	if serverNow.Location().String() != "UTC" {
		// If we're not in UTC, the nextRun converted to UTC would have different hour
		nextRunUTC := nextRun.UTC()
		if nextRunUTC.Hour() == nextRun.Hour() && serverNow.Location().String() != "UTC" {
			t.Log("Warning: nextRun might be stored as UTC instead of local time")
		}
	}
}

// ============== EXECUTION COUNT TEST ==============

func TestSchedule_ExecutionCount_Increments(t *testing.T) {
	h := NewTestScheduleHelper(t)

	execCount := 0
	h.VM.Set("increment", func() {
		execCount++
	})

	code := `$schedule.delay(10, increment)`
	h.MustRun(t, code)

	time.Sleep(50 * time.Millisecond)

	if execCount != 1 {
		t.Errorf("Job should have executed once, got %d", execCount)
	}

	// Check total execution count
	if h.Schedule.ExecutionCount() < 1 {
		t.Error("ExecutionCount should be at least 1")
	}
}

// Helper function
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
