package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/dop251/goja"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/config"
	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/plugin"
	"github.com/levskiy0/m3m/internal/runtime/modules"
	"github.com/levskiy0/m3m/internal/service"
)

// CrashReason indicates why a runtime stopped
type CrashReason string

const (
	CrashReasonNone     CrashReason = ""
	CrashReasonPanic    CrashReason = "panic"
	CrashReasonError    CrashReason = "error"
	CrashReasonShutdown CrashReason = "shutdown"
)

// Auto-restart settings
const (
	MaxRestartsPerHour    = 5                // max restarts within 1 hour before giving up
	RestartCooldownPeriod = time.Hour        // reset restart counter after this period
	InitialRestartDelay   = 1 * time.Second  // initial delay before restart
	MaxRestartDelay       = 30 * time.Second // max delay between restarts
)

// ProjectRuntime represents a running project instance
type ProjectRuntime struct {
	ProjectID     primitive.ObjectID
	VM            *goja.Runtime
	Cancel        context.CancelFunc
	Logger        *modules.LoggerModule
	Router        *modules.RouterModule
	Scheduler     *modules.ScheduleModule
	Service       *modules.ServiceModule
	Actions       *modules.ActionsModule
	StartedAt     time.Time
	Metrics       *MetricsHistory
	metricsCancel context.CancelFunc
	lastHitCount  int64
	lastJobCount  int64
	code          string      // stored for auto-restart
	crashReason   CrashReason // reason for last crash
	crashMessage  string      // crash error message
	restartCount  int         // number of restarts
	lastRestartAt time.Time   // last restart time
	stopRequested bool        // true if stop was explicitly requested (no auto-restart)
}

// LogBroadcaster interface for broadcasting log updates
type LogBroadcaster interface {
	BroadcastLogUpdate(projectID string)
}

// ActionBroadcaster interface for broadcasting action state changes
type ActionBroadcaster interface {
	BroadcastActionStates(projectID string, states []domain.ActionRuntimeState)
}

// RuntimeStopHandler is called when a runtime stops (either normally or due to crash)
type RuntimeStopHandler interface {
	OnRuntimeStopped(projectID primitive.ObjectID, reason CrashReason, message string)
}

// Manager manages all project runtimes
type Manager struct {
	runtimes          map[string]*ProjectRuntime
	mu                sync.RWMutex
	config            *config.Config
	logger            *slog.Logger
	plugins           *plugin.Loader
	envService        *service.EnvironmentService
	goalService       *service.GoalService
	modelService      *service.ModelService
	storageService    *service.StorageService
	logBroadcaster    LogBroadcaster
	actionBroadcaster ActionBroadcaster
	stopHandler       RuntimeStopHandler
}

func NewManager(
	cfg *config.Config,
	logger *slog.Logger,
	plugins *plugin.Loader,
	envService *service.EnvironmentService,
	goalService *service.GoalService,
	modelService *service.ModelService,
	storageService *service.StorageService,
) *Manager {
	return &Manager{
		runtimes:       make(map[string]*ProjectRuntime),
		config:         cfg,
		logger:         logger,
		plugins:        plugins,
		envService:     envService,
		goalService:    goalService,
		modelService:   modelService,
		storageService: storageService,
	}
}

// SetLogBroadcaster sets the log broadcaster for notifying about new logs
func (m *Manager) SetLogBroadcaster(broadcaster LogBroadcaster) {
	m.logBroadcaster = broadcaster
}

// SetActionBroadcaster sets the action broadcaster for notifying about action state changes
func (m *Manager) SetActionBroadcaster(broadcaster ActionBroadcaster) {
	m.actionBroadcaster = broadcaster
}

// SetStopHandler sets the handler for runtime stop events
func (m *Manager) SetStopHandler(handler RuntimeStopHandler) {
	m.stopHandler = handler
}

// Start starts a project with the given code
func (m *Manager) Start(ctx context.Context, projectID primitive.ObjectID, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectIDStr := projectID.Hex()

	// Stop existing runtime if running
	if existing, ok := m.runtimes[projectIDStr]; ok {
		m.stopRuntime(existing)
	}

	// Create new runtime context - use Background() so runtime survives parent context cancellation
	runtimeCtx, cancel := context.WithCancel(context.Background())

	// Create GOJA VM
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	logFile, err := m.storageService.CreateLogFile(projectIDStr)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create log file: %w", err)
	}

	loggerModule := modules.NewLoggerModule(logFile)

	if m.logBroadcaster != nil {
		loggerModule.SetOnLog(func() {
			m.logBroadcaster.BroadcastLogUpdate(projectIDStr)
		})
	}

	routerModule := modules.NewRouterModule()
	routerModule.SetVM(vm)
	schedulerModule := modules.NewScheduleModule(m.logger)
	serviceModule := modules.NewServiceModule(vm, m.config.Runtime.Timeout)
	actionsModule := modules.NewActionsModule(vm, projectID, m.actionBroadcaster)

	// Create env getter for lazy loading (enables hot reload of env vars)
	envGetter := func() map[string]interface{} {
		envMap, _ := m.envService.GetEnvMap(context.Background(), projectID)
		return envMap
	}

	if err := m.registerModules(vm, projectID, loggerModule, routerModule, schedulerModule, serviceModule, actionsModule, envGetter); err != nil {
		cancel()
		return fmt.Errorf("failed to register modules: %w", err)
	}

	if m.plugins != nil {
		if err := m.plugins.RegisterAll(vm); err != nil {
			cancel()
			return fmt.Errorf("failed to register plugins: %w", err)
		}
	}

	metricsCtx, metricsCancel := context.WithCancel(context.Background())

	rt := &ProjectRuntime{
		ProjectID:     projectID,
		VM:            vm,
		Cancel:        cancel,
		Logger:        loggerModule,
		Router:        routerModule,
		Scheduler:     schedulerModule,
		Service:       serviceModule,
		Actions:       actionsModule,
		StartedAt:     time.Now(),
		Metrics:       NewMetricsHistory(),
		metricsCancel: metricsCancel,
		code:          code,
	}

	go rt.collectMetrics(metricsCtx)

	go func() {
		var crashReason CrashReason
		var crashMessage string

		defer func() {
			if r := recover(); r != nil {
				crashReason = CrashReasonPanic
				crashMessage = fmt.Sprintf("%v", r)
				// Get stack trace
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				loggerModule.Error(fmt.Sprintf("FATAL PANIC: %v", r))
				loggerModule.Error(fmt.Sprintf("Stack trace:\n%s", stackTrace))
				m.logger.Error("Runtime panic",
					"project", projectIDStr,
					"panic", r,
					"stack", stackTrace,
				)
			}

			// Log shutdown reason
			rt.crashReason = crashReason
			rt.crashMessage = crashMessage

			uptime := time.Since(rt.StartedAt)
			shouldRestart := false

			if crashReason != "" && crashReason != CrashReasonShutdown {
				loggerModule.Error(fmt.Sprintf("Service crashed after %s: [%s] %s",
					formatDuration(uptime), crashReason, crashMessage))
				m.logger.Error("Service crashed",
					"project", projectIDStr,
					"reason", crashReason,
					"message", crashMessage,
					"uptime", uptime.String(),
				)

				// Check if we should auto-restart
				if !rt.stopRequested {
					// Reset counter if cooldown period passed
					if time.Since(rt.lastRestartAt) > RestartCooldownPeriod {
						rt.restartCount = 0
					}

					if rt.restartCount < MaxRestartsPerHour {
						shouldRestart = true
					} else {
						loggerModule.Error(fmt.Sprintf("Auto-restart disabled: exceeded max restarts (%d) per hour", MaxRestartsPerHour))
						m.logger.Error("Auto-restart limit exceeded",
							"project", projectIDStr,
							"restarts", rt.restartCount,
							"max", MaxRestartsPerHour,
						)
					}
				}
			} else {
				loggerModule.Info(fmt.Sprintf("Service stopped after %s (normal shutdown)",
					formatDuration(uptime)))
				m.logger.Info("Service stopped normally",
					"project", projectIDStr,
					"uptime", uptime.String(),
				)
			}

			// Notify stop handler (for updating status in DB)
			if m.stopHandler != nil {
				m.stopHandler.OnRuntimeStopped(projectID, crashReason, crashMessage)
			}

			// Auto-restart if needed
			if shouldRestart {
				restartCount := rt.restartCount + 1
				lastRestartAt := time.Now()

				// Calculate backoff delay
				delay := InitialRestartDelay * time.Duration(1<<uint(restartCount-1))
				if delay > MaxRestartDelay {
					delay = MaxRestartDelay
				}

				loggerModule.Info(fmt.Sprintf("Auto-restarting in %v (attempt %d/%d)...",
					delay, restartCount, MaxRestartsPerHour))
				m.logger.Info("Auto-restarting project",
					"project", projectIDStr,
					"attempt", restartCount,
					"delay", delay.String(),
				)

				code := rt.code

				time.Sleep(delay)

				// Restart with preserved restart info
				go func() {
					if err := m.startWithRestartInfo(projectID, code, restartCount, lastRestartAt); err != nil {
						m.logger.Error("Auto-restart failed",
							"project", projectIDStr,
							"error", err,
						)
					}
				}()
			} else {
				// Remove from runtimes map only if not restarting
				m.mu.Lock()
				delete(m.runtimes, projectIDStr)
				m.mu.Unlock()
			}
		}()

		_, err := vm.RunString(code)
		if err != nil {
			crashReason = CrashReasonError
			crashMessage = err.Error()
			loggerModule.Error(fmt.Sprintf("Runtime error: %v", err))
			m.logger.Error("Runtime execution error", "project", projectIDStr, "error", err)
			return
		}

		loggerModule.Info("Executing boot phase...")
		if err := serviceModule.ExecuteBoot(); err != nil {
			crashReason = CrashReasonError
			crashMessage = fmt.Sprintf("boot: %v", err)
			loggerModule.Error(fmt.Sprintf("Boot error: %v", err))
			m.logger.Error("Boot error", "project", projectIDStr, "error", err)
			return
		}

		loggerModule.Info("Executing start phase...")
		if err := serviceModule.ExecuteStart(); err != nil {
			loggerModule.Error(fmt.Sprintf("Start error: %v", err))
			m.logger.Error("Start error", "project", projectIDStr, "error", err)
			// Don't return - continue running even if start callback has error
		}

		schedulerModule.Start()

		loggerModule.Info("Service is running")

		// Wait for shutdown signal
		<-runtimeCtx.Done()

		// Normal shutdown - not a crash
		loggerModule.Info("Shutdown signal received")
		loggerModule.Info("Executing shutdown phase...")
		if err := serviceModule.ExecuteShutdown(); err != nil {
			loggerModule.Error(fmt.Sprintf("Shutdown error: %v", err))
		}

		schedulerModule.Stop()
		loggerModule.Info("Service stopped")
		loggerModule.Close()
	}()

	m.runtimes[projectIDStr] = rt
	m.logger.Info("Project started", "project", projectIDStr)

	return nil
}

func (m *Manager) stopRuntime(rt *ProjectRuntime) {
	// Log who is stopping the runtime with stack trace
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	m.logger.Debug("stopRuntime called",
		"project", rt.ProjectID.Hex(),
		"stack", string(buf[:n]),
	)

	if rt.metricsCancel != nil {
		rt.metricsCancel()
	}
	if rt.Service != nil {
		rt.Service.ExecuteShutdown()
	}
	rt.Cancel()
	if rt.Scheduler != nil {
		rt.Scheduler.Stop()
	}
	if rt.Logger != nil {
		rt.Logger.Close()
	}
}

func (rt *ProjectRuntime) collectMetrics(ctx context.Context) {
	rt.collectSnapshot()

	ticker := time.NewTicker(MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rt.collectSnapshot()
		}
	}
}

// collectSnapshot collects a single metrics snapshot
func (rt *ProjectRuntime) collectSnapshot() {
	var requestsDelta int64
	if rt.Router != nil {
		currentHits := rt.Router.HitCount()
		requestsDelta = currentHits - rt.lastHitCount
		rt.lastHitCount = currentHits
	}

	var jobsDelta int64
	if rt.Scheduler != nil {
		currentJobs := rt.Scheduler.ExecutionCount()
		jobsDelta = currentJobs - rt.lastJobCount
		rt.lastJobCount = currentJobs
	}

	// CPU tracking is simplified - for proper tracking use gopsutil
	var cpuPercent float64 = 0

	snapshot := rt.Metrics.CollectSnapshot(requestsDelta, jobsDelta, cpuPercent)
	rt.Metrics.AddSnapshot(snapshot)
}

// Stop stops a running project
func (m *Manager) Stop(projectID primitive.ObjectID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectIDStr := projectID.Hex()
	rt, ok := m.runtimes[projectIDStr]
	if !ok {
		return fmt.Errorf("project not running")
	}

	// Mark as explicitly stopped (no auto-restart)
	rt.stopRequested = true

	m.stopRuntime(rt)
	delete(m.runtimes, projectIDStr)
	m.logger.Info("Project stopped", "project", projectIDStr)

	return nil
}

// IsRunning checks if a project is running
func (m *Manager) IsRunning(projectID primitive.ObjectID) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.runtimes[projectID.Hex()]
	return ok
}

// GetRuntime returns the runtime for a project
func (m *Manager) GetRuntime(projectID primitive.ObjectID) (*ProjectRuntime, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	runtime, ok := m.runtimes[projectID.Hex()]
	return runtime, ok
}

// HandleRoute handles an HTTP request for a project route
func (m *Manager) HandleRoute(projectID primitive.ObjectID, method, path string, ctx *modules.RequestContext) (*modules.ResponseData, error) {
	m.mu.RLock()
	runtime, ok := m.runtimes[projectID.Hex()]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("project not running")
	}

	return runtime.Router.Handle(method, path, ctx)
}

// GetCORSConfig returns CORS configuration for a project
func (m *Manager) GetCORSConfig(projectID primitive.ObjectID) *modules.CORSConfig {
	m.mu.RLock()
	runtime, ok := m.runtimes[projectID.Hex()]
	m.mu.RUnlock()

	if !ok || runtime.Router == nil {
		return nil
	}

	return runtime.Router.GetCORSConfig()
}

// TriggerAction triggers an action handler in a running project
func (m *Manager) TriggerAction(projectID primitive.ObjectID, slug string) error {
	m.mu.RLock()
	runtime, ok := m.runtimes[projectID.Hex()]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("project not running")
	}

	if runtime.Actions == nil {
		return fmt.Errorf("actions module not initialized")
	}

	return runtime.Actions.Trigger(slug)
}

// GetActionStates returns current action states for a project
func (m *Manager) GetActionStates(projectID primitive.ObjectID) []domain.ActionRuntimeState {
	m.mu.RLock()
	runtime, ok := m.runtimes[projectID.Hex()]
	m.mu.RUnlock()

	if !ok || runtime.Actions == nil {
		return nil
	}

	return runtime.Actions.GetStates()
}

// StopAll stops all running runtimes (for graceful shutdown)
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for projectIDStr, rt := range m.runtimes {
		m.logger.Info("Stopping project", "project", projectIDStr)
		rt.stopRequested = true // No auto-restart on graceful shutdown
		m.stopRuntime(rt)
	}
	m.runtimes = make(map[string]*ProjectRuntime)
}

// GetRunningProjects returns list of running project IDs
func (m *Manager) GetRunningProjects() []primitive.ObjectID {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]primitive.ObjectID, 0, len(m.runtimes))
	for _, rt := range m.runtimes {
		ids = append(ids, rt.ProjectID)
	}
	return ids
}

// RuntimeStats contains statistics about a running runtime
type RuntimeStats struct {
	ProjectID       string           `json:"project_id"`
	Status          string           `json:"status"`
	StartedAt       time.Time        `json:"started_at"`
	UptimeSeconds   int64            `json:"uptime_seconds"`
	UptimeFormatted string           `json:"uptime_formatted"`
	RoutesCount     int              `json:"routes_count"`
	RoutesByMethod  map[string]int   `json:"routes_by_method"`
	ScheduledJobs   int              `json:"scheduled_jobs"`
	SchedulerActive bool             `json:"scheduler_active"`
	Memory          MemoryStats      `json:"memory"`
	TotalRequests   int64            `json:"total_requests"`
	HitsByPath      map[string]int64 `json:"hits_by_path"`
	History         *SparklineData   `json:"history,omitempty"`
	// Extended metrics
	StorageBytes  int64   `json:"storage_bytes"`
	DatabaseBytes int64   `json:"database_bytes"`
	CPUPercent    float64 `json:"cpu_percent"`
}

// MemoryStats contains memory statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`       // bytes allocated and still in use
	TotalAlloc uint64 `json:"total_alloc"` // bytes allocated (even if freed)
	Sys        uint64 `json:"sys"`         // bytes obtained from system
	NumGC      uint32 `json:"num_gc"`      // number of completed GC cycles
}

// GetStats returns statistics for a running project
func (m *Manager) GetStats(projectID primitive.ObjectID) (*RuntimeStats, error) {
	m.mu.RLock()
	rt, ok := m.runtimes[projectID.Hex()]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("project not running")
	}

	uptime := time.Since(rt.StartedAt)

	// Get memory stats (process-wide, not per-VM)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := &RuntimeStats{
		ProjectID:       projectID.Hex(),
		Status:          "running",
		StartedAt:       rt.StartedAt,
		UptimeSeconds:   int64(uptime.Seconds()),
		UptimeFormatted: formatDuration(uptime),
		RoutesCount:     0,
		RoutesByMethod:  make(map[string]int),
		ScheduledJobs:   0,
		SchedulerActive: false,
		Memory: MemoryStats{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
	}

	// Get router stats
	if rt.Router != nil {
		stats.RoutesCount = rt.Router.RoutesCount()
		stats.RoutesByMethod = rt.Router.RoutesByMethod()
		stats.TotalRequests = rt.Router.HitCount()
		stats.HitsByPath = rt.Router.HitsByPath()
	}

	// Get scheduler stats
	if rt.Scheduler != nil {
		stats.ScheduledJobs = rt.Scheduler.JobsCount()
		stats.SchedulerActive = rt.Scheduler.IsStarted()
	}

	// Get metrics history for sparklines
	if rt.Metrics != nil {
		sparklineData := rt.Metrics.GetSparklineData()
		stats.History = &sparklineData
	}

	// Get storage size
	if m.storageService != nil {
		if size, err := m.storageService.GetStorageSize(projectID.Hex()); err == nil {
			stats.StorageBytes = size
		}
	}

	// Get database size
	if m.modelService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if size, err := m.modelService.GetProjectDataSize(ctx, projectID); err == nil {
			stats.DatabaseBytes = size
		}
	}

	// Get CPU percent from latest metrics snapshot
	if rt.Metrics != nil {
		latestSnapshots := rt.Metrics.GetLatest(1)
		if len(latestSnapshots) > 0 {
			stats.CPUPercent = latestSnapshots[0].CPUPercent
		}
	}

	return stats, nil
}

// GetBasicStats returns storage and database sizes for a project (even when not running)
func (m *Manager) GetBasicStats(projectID primitive.ObjectID) *RuntimeStats {
	stats := &RuntimeStats{
		ProjectID:      projectID.Hex(),
		Status:         "stopped",
		RoutesByMethod: make(map[string]int),
	}

	// Get storage size
	if m.storageService != nil {
		if size, err := m.storageService.GetStorageSize(projectID.Hex()); err == nil {
			stats.StorageBytes = size
		}
	}

	// Get database size
	if m.modelService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if size, err := m.modelService.GetProjectDataSize(ctx, projectID); err == nil {
			stats.DatabaseBytes = size
		}
	}

	return stats
}

// formatDuration formats duration as human-readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func (m *Manager) registerModules(
	vm *goja.Runtime,
	projectID primitive.ObjectID,
	loggerModule *modules.LoggerModule,
	routerModule *modules.RouterModule,
	schedulerModule *modules.ScheduleModule,
	serviceModule *modules.ServiceModule,
	actionsModule *modules.ActionsModule,
	envGetter func() map[string]interface{},
) error {
	projectIDStr := projectID.Hex()

	// Pre-initialized modules (passed as arguments)
	serviceModule.Register(vm)
	loggerModule.Register(vm) // Also registers console
	routerModule.Register(vm)
	schedulerModule.Register(vm)
	actionsModule.Register(vm)

	// Environment-dependent modules (lazy loading for hot reload)
	envModule := modules.NewEnvModule(envGetter)
	envModule.Register(vm)

	mailModule := modules.NewMailModule(envModule)
	mailModule.Register(vm)

	// Storage-dependent modules
	storageModule := modules.NewStorageModule(m.storageService, projectIDStr)
	storageModule.Register(vm)

	imageModule := modules.NewImageModule(m.storageService, projectIDStr)
	imageModule.Register(vm)

	drawModule := modules.NewDrawModule(m.storageService, projectIDStr)
	drawModule.Register(vm)

	// Service-dependent modules
	databaseModule := modules.NewDatabaseModule(m.modelService, projectID)
	databaseModule.Register(vm)

	goalsModule := modules.NewGoalsModule(m.goalService, projectID)
	goalsModule.Register(vm)

	// HTTP module (needs storage for download functionality)
	httpModule := modules.NewHTTPModule(m.config.Runtime.Timeout, m.storageService, projectIDStr)
	httpModule.Register(vm)

	modules.NewCryptoModule().Register(vm)
	modules.NewEncodingModule().Register(vm)
	modules.NewUtilsModule().Register(vm)
	modules.NewValidatorModule().Register(vm)

	delayedModule := modules.NewDelayedModule(m.config.Runtime.WorkerPoolSize)
	delayedModule.Register(vm)

	return nil
}

// startWithRestartInfo starts a project with preserved restart information
func (m *Manager) startWithRestartInfo(projectID primitive.ObjectID, code string, restartCount int, lastRestartAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectIDStr := projectID.Hex()

	// Stop existing runtime if running
	if existing, ok := m.runtimes[projectIDStr]; ok {
		m.stopRuntime(existing)
	}

	// Create new runtime context - use Background() so runtime survives parent context cancellation
	runtimeCtx, cancel := context.WithCancel(context.Background())

	// Create GOJA VM
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	logFile, err := m.storageService.CreateLogFile(projectIDStr)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create log file: %w", err)
	}

	loggerModule := modules.NewLoggerModule(logFile)

	if m.logBroadcaster != nil {
		loggerModule.SetOnLog(func() {
			m.logBroadcaster.BroadcastLogUpdate(projectIDStr)
		})
	}

	routerModule := modules.NewRouterModule()
	routerModule.SetVM(vm)
	schedulerModule := modules.NewScheduleModule(m.logger)
	serviceModule := modules.NewServiceModule(vm, m.config.Runtime.Timeout)
	actionsModule := modules.NewActionsModule(vm, projectID, m.actionBroadcaster)

	// Create env getter for lazy loading (enables hot reload of env vars)
	envGetter := func() map[string]interface{} {
		envMap, _ := m.envService.GetEnvMap(context.Background(), projectID)
		return envMap
	}

	if err := m.registerModules(vm, projectID, loggerModule, routerModule, schedulerModule, serviceModule, actionsModule, envGetter); err != nil {
		cancel()
		return fmt.Errorf("failed to register modules: %w", err)
	}

	if m.plugins != nil {
		if err := m.plugins.RegisterAll(vm); err != nil {
			cancel()
			return fmt.Errorf("failed to register plugins: %w", err)
		}
	}

	metricsCtx, metricsCancel := context.WithCancel(context.Background())

	rt := &ProjectRuntime{
		ProjectID:     projectID,
		VM:            vm,
		Cancel:        cancel,
		Logger:        loggerModule,
		Router:        routerModule,
		Scheduler:     schedulerModule,
		Service:       serviceModule,
		Actions:       actionsModule,
		StartedAt:     time.Now(),
		Metrics:       NewMetricsHistory(),
		metricsCancel: metricsCancel,
		code:          code,
		restartCount:  restartCount,
		lastRestartAt: lastRestartAt,
	}

	go rt.collectMetrics(metricsCtx)

	go func() {
		var crashReason CrashReason
		var crashMessage string

		defer func() {
			if r := recover(); r != nil {
				crashReason = CrashReasonPanic
				crashMessage = fmt.Sprintf("%v", r)
				// Get stack trace
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				stackTrace := string(buf[:n])

				loggerModule.Error(fmt.Sprintf("FATAL PANIC: %v", r))
				loggerModule.Error(fmt.Sprintf("Stack trace:\n%s", stackTrace))
				m.logger.Error("Runtime panic",
					"project", projectIDStr,
					"panic", r,
					"stack", stackTrace,
				)
			}

			// Log shutdown reason
			rt.crashReason = crashReason
			rt.crashMessage = crashMessage

			uptime := time.Since(rt.StartedAt)
			shouldRestart := false

			if crashReason != "" && crashReason != CrashReasonShutdown {
				loggerModule.Error(fmt.Sprintf("Service crashed after %s: [%s] %s",
					formatDuration(uptime), crashReason, crashMessage))
				m.logger.Error("Service crashed",
					"project", projectIDStr,
					"reason", crashReason,
					"message", crashMessage,
					"uptime", uptime.String(),
				)

				// Check if we should auto-restart
				if !rt.stopRequested {
					// Reset counter if cooldown period passed
					if time.Since(rt.lastRestartAt) > RestartCooldownPeriod {
						rt.restartCount = 0
					}

					if rt.restartCount < MaxRestartsPerHour {
						shouldRestart = true
					} else {
						loggerModule.Error(fmt.Sprintf("Auto-restart disabled: exceeded max restarts (%d) per hour", MaxRestartsPerHour))
						m.logger.Error("Auto-restart limit exceeded",
							"project", projectIDStr,
							"restarts", rt.restartCount,
							"max", MaxRestartsPerHour,
						)
					}
				}
			} else {
				loggerModule.Info(fmt.Sprintf("Service stopped after %s (normal shutdown)",
					formatDuration(uptime)))
				m.logger.Info("Service stopped normally",
					"project", projectIDStr,
					"uptime", uptime.String(),
				)
			}

			// Notify stop handler (for updating status in DB)
			if m.stopHandler != nil {
				m.stopHandler.OnRuntimeStopped(projectID, crashReason, crashMessage)
			}

			// Auto-restart if needed
			if shouldRestart {
				newRestartCount := rt.restartCount + 1
				newLastRestartAt := time.Now()

				// Calculate backoff delay
				delay := InitialRestartDelay * time.Duration(1<<uint(newRestartCount-1))
				if delay > MaxRestartDelay {
					delay = MaxRestartDelay
				}

				loggerModule.Info(fmt.Sprintf("Auto-restarting in %v (attempt %d/%d)...",
					delay, newRestartCount, MaxRestartsPerHour))
				m.logger.Info("Auto-restarting project",
					"project", projectIDStr,
					"attempt", newRestartCount,
					"delay", delay.String(),
				)

				savedCode := rt.code

				time.Sleep(delay)

				// Restart with preserved restart info
				go func() {
					if err := m.startWithRestartInfo(projectID, savedCode, newRestartCount, newLastRestartAt); err != nil {
						m.logger.Error("Auto-restart failed",
							"project", projectIDStr,
							"error", err,
						)
					}
				}()
			} else {
				// Remove from runtimes map only if not restarting
				m.mu.Lock()
				delete(m.runtimes, projectIDStr)
				m.mu.Unlock()
			}
		}()

		loggerModule.Info(fmt.Sprintf("Service restarted (attempt %d/%d)", restartCount, MaxRestartsPerHour))

		_, err := vm.RunString(code)
		if err != nil {
			crashReason = CrashReasonError
			crashMessage = err.Error()
			loggerModule.Error(fmt.Sprintf("Runtime error: %v", err))
			m.logger.Error("Runtime execution error", "project", projectIDStr, "error", err)
			return
		}

		loggerModule.Info("Executing boot phase...")
		if err := serviceModule.ExecuteBoot(); err != nil {
			crashReason = CrashReasonError
			crashMessage = fmt.Sprintf("boot: %v", err)
			loggerModule.Error(fmt.Sprintf("Boot error: %v", err))
			m.logger.Error("Boot error", "project", projectIDStr, "error", err)
			return
		}

		loggerModule.Info("Executing start phase...")
		if err := serviceModule.ExecuteStart(); err != nil {
			loggerModule.Error(fmt.Sprintf("Start error: %v", err))
			m.logger.Error("Start error", "project", projectIDStr, "error", err)
			// Don't return - continue running even if start callback has error
		}

		schedulerModule.Start()

		loggerModule.Info("Service is running")

		// Wait for shutdown signal
		<-runtimeCtx.Done()

		// Normal shutdown - not a crash
		loggerModule.Info("Shutdown signal received")
		loggerModule.Info("Executing shutdown phase...")
		if err := serviceModule.ExecuteShutdown(); err != nil {
			loggerModule.Error(fmt.Sprintf("Shutdown error: %v", err))
		}

		schedulerModule.Stop()
		loggerModule.Info("Service stopped")
		loggerModule.Close()
	}()

	m.runtimes[projectIDStr] = rt
	m.logger.Info("Project restarted", "project", projectIDStr, "attempt", restartCount)

	return nil
}

// GetTypeDefinitions returns TypeScript definitions for Monaco
func (m *Manager) GetTypeDefinitions() string {
	baseTypes := modules.GetBaseTypeDefinitions()
	pluginTypes := m.plugins.GetTypeDefinitions()
	return baseTypes + "\n" + pluginTypes
}
