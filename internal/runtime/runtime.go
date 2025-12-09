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
	"github.com/levskiy0/m3m/internal/plugin"
	"github.com/levskiy0/m3m/internal/runtime/modules"
	"github.com/levskiy0/m3m/internal/service"
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
	StartedAt     time.Time
	Metrics       *MetricsHistory
	metricsCancel context.CancelFunc
	lastHitCount  int64
	lastJobCount  int64
}

// LogBroadcaster interface for broadcasting log updates
type LogBroadcaster interface {
	BroadcastLogUpdate(projectID string)
}

// Manager manages all project runtimes
type Manager struct {
	runtimes       map[string]*ProjectRuntime
	mu             sync.RWMutex
	config         *config.Config
	logger         *slog.Logger
	plugins        *plugin.Loader
	envService     *service.EnvironmentService
	goalService    *service.GoalService
	modelService   *service.ModelService
	storageService *service.StorageService
	logBroadcaster LogBroadcaster
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

// Start starts a project with the given code
func (m *Manager) Start(ctx context.Context, projectID primitive.ObjectID, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	projectIDStr := projectID.Hex()

	// Stop existing runtime if running
	if existing, ok := m.runtimes[projectIDStr]; ok {
		m.stopRuntime(existing)
	}

	// Create new runtime context
	runtimeCtx, cancel := context.WithCancel(ctx)

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

	envMap, _ := m.envService.GetEnvMap(ctx, projectID)

	if err := m.registerModules(vm, projectID, loggerModule, routerModule, schedulerModule, serviceModule, envMap); err != nil {
		cancel()
		return fmt.Errorf("failed to register modules: %w", err)
	}

	if err := m.plugins.RegisterAll(vm); err != nil {
		cancel()
		return fmt.Errorf("failed to register plugins: %w", err)
	}

	metricsCtx, metricsCancel := context.WithCancel(context.Background())

	runtime := &ProjectRuntime{
		ProjectID:     projectID,
		VM:            vm,
		Cancel:        cancel,
		Logger:        loggerModule,
		Router:        routerModule,
		Scheduler:     schedulerModule,
		Service:       serviceModule,
		StartedAt:     time.Now(),
		Metrics:       NewMetricsHistory(),
		metricsCancel: metricsCancel,
	}

	go runtime.collectMetrics(metricsCtx)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				loggerModule.Error(fmt.Sprintf("Runtime panic: %v", r))
			}
		}()

		_, err := vm.RunString(code)
		if err != nil {
			loggerModule.Error(fmt.Sprintf("Runtime error: %v", err))
			m.logger.Error("Runtime execution error", "project", projectIDStr, "error", err)
			return
		}

		loggerModule.Info("Executing boot phase...")
		if err := serviceModule.ExecuteBoot(); err != nil {
			loggerModule.Error(fmt.Sprintf("Boot error: %v", err))
			m.logger.Error("Boot error", "project", projectIDStr, "error", err)
		}

		loggerModule.Info("Executing start phase...")
		if err := serviceModule.ExecuteStart(); err != nil {
			loggerModule.Error(fmt.Sprintf("Start error: %v", err))
			m.logger.Error("Start error", "project", projectIDStr, "error", err)
		}

		schedulerModule.Start()

		loggerModule.Info("Service is running")

		<-runtimeCtx.Done()

		loggerModule.Info("Executing shutdown phase...")
		if err := serviceModule.ExecuteShutdown(); err != nil {
			loggerModule.Error(fmt.Sprintf("Shutdown error: %v", err))
		}

		schedulerModule.Stop()
		loggerModule.Info("Service stopped")
		loggerModule.Close()
	}()

	m.runtimes[projectIDStr] = runtime
	m.logger.Info("Project started", "project", projectIDStr)

	return nil
}

func (m *Manager) stopRuntime(runtime *ProjectRuntime) {
	if runtime.metricsCancel != nil {
		runtime.metricsCancel()
	}
	if runtime.Service != nil {
		runtime.Service.ExecuteShutdown()
	}
	runtime.Cancel()
	if runtime.Scheduler != nil {
		runtime.Scheduler.Stop()
	}
	if runtime.Logger != nil {
		runtime.Logger.Close()
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
	runtime, ok := m.runtimes[projectIDStr]
	if !ok {
		return fmt.Errorf("project not running")
	}

	m.stopRuntime(runtime)
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

// StopAll stops all running runtimes (for graceful shutdown)
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for projectIDStr, runtime := range m.runtimes {
		m.logger.Info("Stopping project", "project", projectIDStr)
		m.stopRuntime(runtime)
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
	envMap map[string]interface{},
) error {
	projectIDStr := projectID.Hex()

	// Pre-initialized modules (passed as arguments)
	serviceModule.Register(vm)
	loggerModule.Register(vm) // Also registers console
	routerModule.Register(vm)
	schedulerModule.Register(vm)

	// Environment-dependent modules
	envModule := modules.NewEnvModule(envMap)
	envModule.Register(vm)

	smtpModule := modules.NewSMTPModule(envModule)
	smtpModule.Register(vm)

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

// GetTypeDefinitions returns TypeScript definitions for Monaco
func (m *Manager) GetTypeDefinitions() string {
	baseTypes := modules.GetBaseTypeDefinitions()
	pluginTypes := m.plugins.GetTypeDefinitions()
	return baseTypes + "\n" + pluginTypes
}
