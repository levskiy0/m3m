package modules

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/pkg/schema"
	"github.com/robfig/cron/v3"
)

// JobStatus represents the status of a scheduled job
type JobStatus string

const (
	JobStatusActive  JobStatus = "active"
	JobStatusPaused  JobStatus = "paused"
	JobStatusOneTime JobStatus = "one-time"
)

// JobInfo contains metadata about a scheduled job
type JobInfo struct {
	ID             string        `json:"id"`
	Type           string        `json:"type"` // "cron", "interval", "once", "delay"
	Expression     string        `json:"expression,omitempty"`
	Interval       string        `json:"interval,omitempty"`
	Status         JobStatus     `json:"status"`
	LastRun        *time.Time    `json:"lastRun,omitempty"`
	NextRun        *time.Time    `json:"nextRun,omitempty"`
	ExecutionCount int64         `json:"executionCount"`
	SkipIfRunning  bool          `json:"skipIfRunning"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	CreatedAt      time.Time     `json:"createdAt"`
}

// JobOptions contains options for job execution
type JobOptions struct {
	SkipIfRunning bool          `json:"skipIfRunning"`
	Timeout       time.Duration `json:"timeout"`
	LockName      string        `json:"lockName"`
	LockTimeout   time.Duration `json:"lockTimeout"`
}

// internalJob wraps job information with internal state
type internalJob struct {
	info       JobInfo
	handler    goja.Callable
	cronID     cron.EntryID
	timer      *time.Timer
	isRunning  atomic.Bool
	cancelFunc func()
}

type ScheduleModule struct {
	cron           *cron.Cron
	jobs           map[string]*internalJob
	jobCounter     int64
	mu             sync.Mutex
	started        bool
	logger         *slog.Logger
	executionCount int64
}

func NewScheduleModule(logger *slog.Logger) *ScheduleModule {
	return &ScheduleModule{
		cron:   cron.New(),
		jobs:   make(map[string]*internalJob),
		logger: logger,
	}
}

// generateJobID generates a unique job ID
func (s *ScheduleModule) generateJobID() string {
	id := atomic.AddInt64(&s.jobCounter, 1)
	return fmt.Sprintf("job_%d", id)
}

// addCronJob adds a cron-based job and returns the job ID
func (s *ScheduleModule) addCronJob(spec string, handler goja.Callable, opts *JobOptions) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobID := s.generateJobID()

	job := &internalJob{
		info: JobInfo{
			ID:            jobID,
			Type:          "cron",
			Expression:    spec,
			Status:        JobStatusActive,
			SkipIfRunning: opts != nil && opts.SkipIfRunning,
			Timeout:       0,
			CreatedAt:     time.Now(),
		},
		handler: handler,
	}

	if opts != nil {
		job.info.Timeout = opts.Timeout
	}

	cronID, err := s.cron.AddFunc(spec, func() {
		s.executeJob(job)
	})

	if err != nil {
		s.logger.Error("Failed to add scheduled job", "spec", spec, "error", err)
		return ""
	}

	job.cronID = cronID
	s.jobs[jobID] = job

	// Update next run time
	entry := s.cron.Entry(cronID)
	if !entry.Next.IsZero() {
		next := entry.Next
		job.info.NextRun = &next
	}

	return jobID
}

// executeJob executes a job with proper handling
func (s *ScheduleModule) executeJob(job *internalJob) {
	// Check if should skip when already running
	if job.info.SkipIfRunning && !job.isRunning.CompareAndSwap(false, true) {
		s.logger.Debug("Skipping job execution - already running", "jobID", job.info.ID)
		return
	}

	defer func() {
		if job.info.SkipIfRunning {
			job.isRunning.Store(false)
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Scheduled job panic", "jobID", job.info.ID, "error", r)
		}
	}()

	// Increment execution counts
	atomic.AddInt64(&s.executionCount, 1)
	job.info.ExecutionCount++
	now := time.Now()
	job.info.LastRun = &now

	// Execute with timeout if specified
	if job.info.Timeout > 0 {
		done := make(chan struct{})
		go func() {
			defer close(done)
			_, err := job.handler(nil, nil)
			if err != nil {
				s.logger.Error("Scheduled job error", "jobID", job.info.ID, "error", err)
			}
		}()

		select {
		case <-done:
			// Job completed
		case <-time.After(job.info.Timeout):
			s.logger.Error("Scheduled job timeout", "jobID", job.info.ID, "timeout", job.info.Timeout)
		}
	} else {
		_, err := job.handler(nil, nil)
		if err != nil {
			s.logger.Error("Scheduled job error", "jobID", job.info.ID, "error", err)
		}
	}

	// Update next run time for cron jobs
	if job.cronID != 0 {
		entry := s.cron.Entry(job.cronID)
		if !entry.Next.IsZero() {
			next := entry.Next
			job.info.NextRun = &next
		}
	}
}

// parseDuration parses duration strings like "5m", "2h", "1d"
func parseDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`^(\d+)([smhd])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "s":
		return time.Duration(value) * time.Second, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %s", unit)
	}
}

// parseOptions parses job options from goja.Value
func (s *ScheduleModule) parseOptions(vm *goja.Runtime, val goja.Value) *JobOptions {
	if val == nil || goja.IsUndefined(val) || goja.IsNull(val) {
		return nil
	}

	obj := val.ToObject(vm)
	if obj == nil {
		return nil
	}

	opts := &JobOptions{}

	if skipVal := obj.Get("skipIfRunning"); skipVal != nil && !goja.IsUndefined(skipVal) {
		opts.SkipIfRunning = skipVal.ToBoolean()
	}

	if timeoutVal := obj.Get("timeout"); timeoutVal != nil && !goja.IsUndefined(timeoutVal) {
		opts.Timeout = time.Duration(timeoutVal.ToInteger()) * time.Millisecond
	}

	if lockVal := obj.Get("lockName"); lockVal != nil && !goja.IsUndefined(lockVal) {
		opts.LockName = lockVal.String()
	}

	if lockTimeoutVal := obj.Get("lockTimeout"); lockTimeoutVal != nil && !goja.IsUndefined(lockTimeoutVal) {
		opts.LockTimeout = time.Duration(lockTimeoutVal.ToInteger()) * time.Millisecond
	}

	return opts
}

// ================== Preset Methods ==================

// Daily schedules a job to run at midnight every day
func (s *ScheduleModule) Daily(handler goja.Callable) string {
	return s.addCronJob("0 0 * * *", handler, nil)
}

// Hourly schedules a job to run at the start of every hour
func (s *ScheduleModule) Hourly(handler goja.Callable) string {
	return s.addCronJob("0 * * * *", handler, nil)
}

// Minutely schedules a job to run every minute
func (s *ScheduleModule) Minutely(handler goja.Callable) string {
	return s.addCronJob("* * * * *", handler, nil)
}

// Weekly schedules a job to run weekly on a specific day
// dayOfWeek: 0 = Sunday, 1 = Monday, ..., 6 = Saturday
func (s *ScheduleModule) Weekly(dayOfWeek int, handler goja.Callable) string {
	if dayOfWeek < 0 || dayOfWeek > 6 {
		s.logger.Error("Invalid day of week", "dayOfWeek", dayOfWeek)
		return ""
	}
	spec := fmt.Sprintf("0 0 * * %d", dayOfWeek)
	return s.addCronJob(spec, handler, nil)
}

// Monthly schedules a job to run monthly on a specific day
// dayOfMonth: 1-31
func (s *ScheduleModule) Monthly(dayOfMonth int, handler goja.Callable) string {
	if dayOfMonth < 1 || dayOfMonth > 31 {
		s.logger.Error("Invalid day of month", "dayOfMonth", dayOfMonth)
		return ""
	}
	spec := fmt.Sprintf("0 0 %d * *", dayOfMonth)
	return s.addCronJob(spec, handler, nil)
}

// At schedules a job to run daily at a specific time
// timeStr: "HH:MM" format (24-hour)
func (s *ScheduleModule) At(timeStr string, handler goja.Callable) string {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		s.logger.Error("Invalid time format", "time", timeStr)
		return ""
	}

	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		s.logger.Error("Invalid time format", "time", timeStr)
		return ""
	}

	spec := fmt.Sprintf("%d %d * * *", minute, hour)
	return s.addCronJob(spec, handler, nil)
}

// Cron schedules a job using a cron expression with optional options
func (s *ScheduleModule) Cron(call goja.FunctionCall, vm *goja.Runtime) goja.Value {
	if len(call.Arguments) < 2 {
		s.logger.Error("Cron requires at least 2 arguments")
		return vm.ToValue("")
	}

	expression := call.Arguments[0].String()
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		s.logger.Error("Second argument must be a function")
		return vm.ToValue("")
	}

	var opts *JobOptions
	if len(call.Arguments) >= 3 {
		opts = s.parseOptions(vm, call.Arguments[2])
	}

	jobID := s.addCronJob(expression, handler, opts)
	return vm.ToValue(jobID)
}

// ================== Interval Methods ==================

// Every schedules a job to run at regular intervals
// interval: "5m", "2h", "1d", etc.
func (s *ScheduleModule) Every(call goja.FunctionCall, vm *goja.Runtime) goja.Value {
	if len(call.Arguments) < 2 {
		s.logger.Error("Every requires at least 2 arguments")
		return vm.ToValue("")
	}

	intervalStr := call.Arguments[0].String()
	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		s.logger.Error("Second argument must be a function")
		return vm.ToValue("")
	}

	duration, err := parseDuration(intervalStr)
	if err != nil {
		s.logger.Error("Invalid interval format", "interval", intervalStr, "error", err)
		return vm.ToValue("")
	}

	s.mu.Lock()
	jobID := s.generateJobID()

	job := &internalJob{
		info: JobInfo{
			ID:        jobID,
			Type:      "interval",
			Interval:  intervalStr,
			Status:    JobStatusActive,
			CreatedAt: time.Now(),
		},
		handler: handler,
	}

	// Start the interval timer
	var runJob func()
	runJob = func() {
		if job.info.Status != JobStatusActive {
			return
		}
		s.executeJob(job)
		job.timer = time.AfterFunc(duration, runJob)
		next := time.Now().Add(duration)
		job.info.NextRun = &next
	}

	next := time.Now().Add(duration)
	job.info.NextRun = &next
	job.timer = time.AfterFunc(duration, runJob)
	s.jobs[jobID] = job
	s.mu.Unlock()

	return vm.ToValue(jobID)
}

// ================== One-Time Tasks ==================

// Once schedules a job to run once at a specific time
// timestamp: Unix timestamp in milliseconds or Date object
func (s *ScheduleModule) Once(call goja.FunctionCall, vm *goja.Runtime) goja.Value {
	if len(call.Arguments) < 2 {
		s.logger.Error("Once requires 2 arguments")
		return vm.ToValue("")
	}

	var targetTime time.Time

	// Handle both Unix timestamp and Date object
	arg := call.Arguments[0]
	if obj, ok := arg.Export().(map[string]interface{}); ok {
		// Date object - try to get timestamp
		if getTime, ok := obj["getTime"]; ok {
			if fn, ok := getTime.(func(goja.FunctionCall) goja.Value); ok {
				ts := fn(goja.FunctionCall{})
				targetTime = time.UnixMilli(ts.ToInteger())
			}
		}
	} else {
		// Assume it's a Unix timestamp in milliseconds
		ts := arg.ToInteger()
		targetTime = time.UnixMilli(ts)
	}

	if targetTime.IsZero() {
		s.logger.Error("Invalid timestamp")
		return vm.ToValue("")
	}

	handler, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		s.logger.Error("Second argument must be a function")
		return vm.ToValue("")
	}

	s.mu.Lock()
	jobID := s.generateJobID()

	job := &internalJob{
		info: JobInfo{
			ID:        jobID,
			Type:      "once",
			Status:    JobStatusOneTime,
			NextRun:   &targetTime,
			CreatedAt: time.Now(),
		},
		handler: handler,
	}

	delay := time.Until(targetTime)
	if delay < 0 {
		s.logger.Warn("Target time is in the past, executing immediately", "jobID", jobID)
		delay = 0
	}

	job.timer = time.AfterFunc(delay, func() {
		s.executeJob(job)
		s.mu.Lock()
		delete(s.jobs, jobID)
		s.mu.Unlock()
	})

	s.jobs[jobID] = job
	s.mu.Unlock()

	return vm.ToValue(jobID)
}

// Delay schedules a job to run after a delay
// ms: delay in milliseconds
func (s *ScheduleModule) Delay(ms int64, handler goja.Callable) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	jobID := s.generateJobID()
	targetTime := time.Now().Add(time.Duration(ms) * time.Millisecond)

	job := &internalJob{
		info: JobInfo{
			ID:        jobID,
			Type:      "delay",
			Status:    JobStatusOneTime,
			NextRun:   &targetTime,
			CreatedAt: time.Now(),
		},
		handler: handler,
	}

	job.timer = time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
		s.executeJob(job)
		s.mu.Lock()
		delete(s.jobs, jobID)
		s.mu.Unlock()
	})

	s.jobs[jobID] = job
	return jobID
}

// ================== Job Management ==================

// Cancel cancels a scheduled job
func (s *ScheduleModule) Cancel(jobID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return false
	}

	// Stop cron job if applicable
	if job.cronID != 0 {
		s.cron.Remove(job.cronID)
	}

	// Stop timer if applicable
	if job.timer != nil {
		job.timer.Stop()
	}

	delete(s.jobs, jobID)
	return true
}

// Pause pauses a scheduled job
func (s *ScheduleModule) Pause(jobID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok || job.info.Status == JobStatusOneTime {
		return false
	}

	if job.info.Status == JobStatusPaused {
		return true // Already paused
	}

	job.info.Status = JobStatusPaused

	// Stop cron job
	if job.cronID != 0 {
		s.cron.Remove(job.cronID)
		job.cronID = 0
	}

	// Stop timer
	if job.timer != nil {
		job.timer.Stop()
		job.timer = nil
	}

	return true
}

// Resume resumes a paused job
func (s *ScheduleModule) Resume(jobID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok || job.info.Status != JobStatusPaused {
		return false
	}

	job.info.Status = JobStatusActive

	// Re-add cron job
	if job.info.Type == "cron" && job.info.Expression != "" {
		cronID, err := s.cron.AddFunc(job.info.Expression, func() {
			s.executeJob(job)
		})
		if err != nil {
			s.logger.Error("Failed to resume cron job", "jobID", jobID, "error", err)
			return false
		}
		job.cronID = cronID
	}

	// Re-start interval timer
	if job.info.Type == "interval" && job.info.Interval != "" {
		duration, err := parseDuration(job.info.Interval)
		if err != nil {
			return false
		}

		var runJob func()
		runJob = func() {
			if job.info.Status != JobStatusActive {
				return
			}
			s.executeJob(job)
			job.timer = time.AfterFunc(duration, runJob)
			next := time.Now().Add(duration)
			job.info.NextRun = &next
		}

		next := time.Now().Add(duration)
		job.info.NextRun = &next
		job.timer = time.AfterFunc(duration, runJob)
	}

	return true
}

// List returns information about all scheduled jobs
func (s *ScheduleModule) List() []map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]map[string]interface{}, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, s.jobToMap(job))
	}
	return result
}

// Get returns information about a specific job
func (s *ScheduleModule) Get(jobID string) map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return nil
	}

	return s.jobToMap(job)
}

// jobToMap converts a job to a map for JavaScript
func (s *ScheduleModule) jobToMap(job *internalJob) map[string]interface{} {
	result := map[string]interface{}{
		"id":             job.info.ID,
		"type":           job.info.Type,
		"status":         string(job.info.Status),
		"executionCount": job.info.ExecutionCount,
		"createdAt":      job.info.CreatedAt.UnixMilli(),
	}

	if job.info.Expression != "" {
		result["expression"] = job.info.Expression
	}
	if job.info.Interval != "" {
		result["interval"] = job.info.Interval
	}
	if job.info.LastRun != nil {
		result["lastRun"] = job.info.LastRun.UnixMilli()
	}
	if job.info.NextRun != nil {
		result["nextRun"] = job.info.NextRun.UnixMilli()
	}
	if job.info.SkipIfRunning {
		result["skipIfRunning"] = true
	}
	if job.info.Timeout > 0 {
		result["timeout"] = job.info.Timeout.Milliseconds()
	}

	return result
}

// LastRun returns the last run timestamp for a job
func (s *ScheduleModule) LastRun(jobID string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok || job.info.LastRun == nil {
		return nil
	}

	return job.info.LastRun.UnixMilli()
}

// NextRun returns the next run timestamp for a job
func (s *ScheduleModule) NextRun(jobID string) interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok || job.info.NextRun == nil {
		return nil
	}

	return job.info.NextRun.UnixMilli()
}

// ================== Lifecycle Methods ==================

func (s *ScheduleModule) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started && len(s.jobs) > 0 {
		s.cron.Start()
		s.started = true
		s.logger.Info("Scheduler started", "jobs", len(s.jobs))
	}
}

func (s *ScheduleModule) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		s.cron.Stop()
		s.started = false

		// Stop all timers
		for _, job := range s.jobs {
			if job.timer != nil {
				job.timer.Stop()
			}
		}

		s.logger.Info("Scheduler stopped")
	}
}

// JobsCount returns the number of scheduled jobs
func (s *ScheduleModule) JobsCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.jobs)
}

// IsStarted returns whether the scheduler is running
func (s *ScheduleModule) IsStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

// ExecutionCount returns the total number of job executions
func (s *ScheduleModule) ExecutionCount() int64 {
	return atomic.LoadInt64(&s.executionCount)
}

// Name returns the module name for JavaScript
func (s *ScheduleModule) Name() string {
	return "$schedule"
}

// Register registers the module into the JavaScript VM
func (s *ScheduleModule) Register(vm interface{}) {
	runtime := vm.(*goja.Runtime)

	runtime.Set(s.Name(), map[string]interface{}{
		// Presets
		"daily":    s.Daily,
		"hourly":   s.Hourly,
		"minutely": s.Minutely,
		"weekly":   s.Weekly,
		"monthly":  s.Monthly,
		"at":       s.At,

		// Cron
		"cron": func(call goja.FunctionCall) goja.Value {
			return s.Cron(call, runtime)
		},

		// Intervals
		"every": func(call goja.FunctionCall) goja.Value {
			return s.Every(call, runtime)
		},

		// One-time
		"once": func(call goja.FunctionCall) goja.Value {
			return s.Once(call, runtime)
		},
		"delay": s.Delay,

		// Job management
		"cancel":  s.Cancel,
		"pause":   s.Pause,
		"resume":  s.Resume,
		"list":    s.List,
		"get":     s.Get,
		"lastRun": s.LastRun,
		"nextRun": s.NextRun,
	})
}

// GetSchema implements JSSchemaProvider
func (s *ScheduleModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$schedule",
		Description: "Job scheduling for periodic tasks using cron expressions, intervals, and one-time execution",
		Types: []schema.TypeSchema{
			{
				Name:        "ScheduleJobInfo",
				Description: "Information about a scheduled job",
				Fields: []schema.ParamSchema{
					{Name: "id", Type: "string", Description: "Unique job identifier"},
					{Name: "type", Type: "string", Description: "Job type: 'cron', 'interval', 'once', 'delay'"},
					{Name: "status", Type: "string", Description: "Job status: 'active', 'paused', 'one-time'"},
					{Name: "expression", Type: "string", Description: "Cron expression (for cron jobs)", Optional: true},
					{Name: "interval", Type: "string", Description: "Interval string (for interval jobs)", Optional: true},
					{Name: "executionCount", Type: "number", Description: "Number of times the job has executed"},
					{Name: "lastRun", Type: "number", Description: "Unix timestamp of last execution", Optional: true},
					{Name: "nextRun", Type: "number", Description: "Unix timestamp of next scheduled execution", Optional: true},
					{Name: "skipIfRunning", Type: "boolean", Description: "Whether to skip if already running", Optional: true},
					{Name: "timeout", Type: "number", Description: "Execution timeout in milliseconds", Optional: true},
					{Name: "createdAt", Type: "number", Description: "Unix timestamp when job was created"},
				},
			},
			{
				Name:        "ScheduleJobOptions",
				Description: "Options for job configuration",
				Fields: []schema.ParamSchema{
					{Name: "skipIfRunning", Type: "boolean", Description: "Skip execution if previous run is still in progress", Optional: true},
					{Name: "timeout", Type: "number", Description: "Maximum execution time in milliseconds", Optional: true},
				},
			},
		},
		Methods: []schema.MethodSchema{
			// Presets
			{
				Name:        "daily",
				Description: "Schedule a job to run daily at midnight",
				Params:      []schema.ParamSchema{{Name: "handler", Type: "() => void", Description: "Function to execute"}},
				Returns:     &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			{
				Name:        "hourly",
				Description: "Schedule a job to run at the start of every hour",
				Params:      []schema.ParamSchema{{Name: "handler", Type: "() => void", Description: "Function to execute"}},
				Returns:     &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			{
				Name:        "minutely",
				Description: "Schedule a job to run every minute",
				Params:      []schema.ParamSchema{{Name: "handler", Type: "() => void", Description: "Function to execute"}},
				Returns:     &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			{
				Name:        "weekly",
				Description: "Schedule a job to run weekly on a specific day",
				Params: []schema.ParamSchema{
					{Name: "dayOfWeek", Type: "number", Description: "Day of week (0=Sunday, 1=Monday, ..., 6=Saturday)"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			{
				Name:        "monthly",
				Description: "Schedule a job to run monthly on a specific day",
				Params: []schema.ParamSchema{
					{Name: "dayOfMonth", Type: "number", Description: "Day of month (1-31)"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			{
				Name:        "at",
				Description: "Schedule a job to run daily at a specific time",
				Params: []schema.ParamSchema{
					{Name: "time", Type: "string", Description: "Time in HH:MM format (24-hour)"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			// Cron
			{
				Name:        "cron",
				Description: "Schedule a job using a cron expression with optional options",
				Params: []schema.ParamSchema{
					{Name: "expression", Type: "string", Description: "Cron expression (e.g., '0 0 * * *')"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
					{Name: "options", Type: "ScheduleJobOptions", Description: "Optional job options", Optional: true},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			// Intervals
			{
				Name:        "every",
				Description: "Schedule a job to run at regular intervals",
				Params: []schema.ParamSchema{
					{Name: "interval", Type: "string", Description: "Interval (e.g., '5s', '5m', '2h', '1d')"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			// One-time
			{
				Name:        "once",
				Description: "Schedule a job to run once at a specific timestamp",
				Params: []schema.ParamSchema{
					{Name: "timestamp", Type: "number | Date", Description: "Unix timestamp in milliseconds or Date object"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			{
				Name:        "delay",
				Description: "Schedule a job to run after a delay",
				Params: []schema.ParamSchema{
					{Name: "ms", Type: "number", Description: "Delay in milliseconds"},
					{Name: "handler", Type: "() => void", Description: "Function to execute"},
				},
				Returns: &schema.ParamSchema{Name: "jobId", Type: "string", Description: "Job ID"},
			},
			// Job management
			{
				Name:        "cancel",
				Description: "Cancel a scheduled job",
				Params:      []schema.ParamSchema{{Name: "jobId", Type: "string", Description: "Job ID to cancel"}},
				Returns:     &schema.ParamSchema{Name: "success", Type: "boolean", Description: "True if job was cancelled"},
			},
			{
				Name:        "pause",
				Description: "Pause a scheduled job",
				Params:      []schema.ParamSchema{{Name: "jobId", Type: "string", Description: "Job ID to pause"}},
				Returns:     &schema.ParamSchema{Name: "success", Type: "boolean", Description: "True if job was paused"},
			},
			{
				Name:        "resume",
				Description: "Resume a paused job",
				Params:      []schema.ParamSchema{{Name: "jobId", Type: "string", Description: "Job ID to resume"}},
				Returns:     &schema.ParamSchema{Name: "success", Type: "boolean", Description: "True if job was resumed"},
			},
			{
				Name:        "list",
				Description: "List all scheduled jobs",
				Params:      []schema.ParamSchema{},
				Returns:     &schema.ParamSchema{Name: "jobs", Type: "ScheduleJobInfo[]", Description: "Array of job information objects"},
			},
			{
				Name:        "get",
				Description: "Get information about a specific job",
				Params:      []schema.ParamSchema{{Name: "jobId", Type: "string", Description: "Job ID"}},
				Returns:     &schema.ParamSchema{Name: "job", Type: "ScheduleJobInfo | null", Description: "Job information or null if not found"},
			},
			{
				Name:        "lastRun",
				Description: "Get the last run timestamp for a job",
				Params:      []schema.ParamSchema{{Name: "jobId", Type: "string", Description: "Job ID"}},
				Returns:     &schema.ParamSchema{Name: "timestamp", Type: "number | null", Description: "Unix timestamp in milliseconds or null"},
			},
			{
				Name:        "nextRun",
				Description: "Get the next run timestamp for a job",
				Params:      []schema.ParamSchema{{Name: "jobId", Type: "string", Description: "Job ID"}},
				Returns:     &schema.ParamSchema{Name: "timestamp", Type: "number | null", Description: "Unix timestamp in milliseconds or null"},
			},
		},
	}
}

// GetScheduleSchema returns the schedule schema (static version)
func GetScheduleSchema() schema.ModuleSchema {
	return (&ScheduleModule{}).GetSchema()
}
