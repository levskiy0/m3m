package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/dop251/goja"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/config"
	"m3m/internal/plugin"
	"m3m/internal/runtime/modules"
	"m3m/internal/service"
)

// ProjectRuntime represents a running project instance
type ProjectRuntime struct {
	ProjectID primitive.ObjectID
	VM        *goja.Runtime
	Cancel    context.CancelFunc
	Logger    *modules.LoggerModule
	Router    *modules.RouterModule
	Scheduler *modules.ScheduleModule
	Service   *modules.ServiceModule
	StartedAt time.Time
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

	// Create log file
	logFile, err := m.storageService.CreateLogFile(projectIDStr)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Initialize modules
	loggerModule := modules.NewLoggerModule(logFile)
	routerModule := modules.NewRouterModule()
	routerModule.SetVM(vm)
	schedulerModule := modules.NewScheduleModule(m.logger)
	serviceModule := modules.NewServiceModule(vm, m.config.Runtime.Timeout)

	// Get environment variables
	envMap, _ := m.envService.GetEnvMap(ctx, projectID)

	// Register built-in modules
	if err := m.registerModules(vm, projectID, loggerModule, routerModule, schedulerModule, serviceModule, envMap); err != nil {
		cancel()
		return fmt.Errorf("failed to register modules: %w", err)
	}

	// Register plugins
	if err := m.plugins.RegisterAll(vm); err != nil {
		cancel()
		return fmt.Errorf("failed to register plugins: %w", err)
	}

	// Create runtime instance
	runtime := &ProjectRuntime{
		ProjectID: projectID,
		VM:        vm,
		Cancel:    cancel,
		Logger:    loggerModule,
		Router:    routerModule,
		Scheduler: schedulerModule,
		Service:   serviceModule,
		StartedAt: time.Now(),
	}

	// Execute code and lifecycle
	go func() {
		defer func() {
			if r := recover(); r != nil {
				loggerModule.Error(fmt.Sprintf("Runtime panic: %v", r))
			}
		}()

		// 1. Execute code (registers lifecycle callbacks)
		_, err := vm.RunString(code)
		if err != nil {
			loggerModule.Error(fmt.Sprintf("Runtime error: %v", err))
			m.logger.Error("Runtime execution error", "project", projectIDStr, "error", err)
			return
		}

		// 2. Execute boot callbacks
		loggerModule.Info("Executing boot phase...")
		if err := serviceModule.ExecuteBoot(); err != nil {
			loggerModule.Error(fmt.Sprintf("Boot error: %v", err))
			m.logger.Error("Boot error", "project", projectIDStr, "error", err)
		}

		// 3. Execute start callbacks
		loggerModule.Info("Executing start phase...")
		if err := serviceModule.ExecuteStart(); err != nil {
			loggerModule.Error(fmt.Sprintf("Start error: %v", err))
			m.logger.Error("Start error", "project", projectIDStr, "error", err)
		}

		// 4. Start scheduler after start phase
		schedulerModule.Start()

		loggerModule.Info("Service is running")

		// 5. Wait for context cancellation
		<-runtimeCtx.Done()

		// 6. Execute shutdown callbacks
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

// stopRuntime stops a runtime without lock
func (m *Manager) stopRuntime(runtime *ProjectRuntime) {
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
	for _, runtime := range m.runtimes {
		ids = append(ids, runtime.ProjectID)
	}
	return ids
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

	// Service lifecycle
	vm.Set("service", serviceModule.GetJSObject())

	// Logger
	vm.Set("logger", map[string]interface{}{
		"debug": loggerModule.Debug,
		"info":  loggerModule.Info,
		"warn":  loggerModule.Warn,
		"error": loggerModule.Error,
	})

	// Router
	vm.Set("router", map[string]interface{}{
		"get":      routerModule.Get,
		"post":     routerModule.Post,
		"put":      routerModule.Put,
		"delete":   routerModule.Delete,
		"response": routerModule.Response,
	})

	// Schedule
	vm.Set("schedule", map[string]interface{}{
		"daily":  schedulerModule.Daily,
		"hourly": schedulerModule.Hourly,
		"cron":   schedulerModule.Cron,
	})

	// Environment
	envModule := modules.NewEnvModule(envMap)
	vm.Set("env", map[string]interface{}{
		"get": envModule.Get,
	})

	// Storage
	storageModule := modules.NewStorageModule(m.storageService, projectIDStr)
	vm.Set("storage", map[string]interface{}{
		"read":   storageModule.Read,
		"write":  storageModule.Write,
		"exists": storageModule.Exists,
		"delete": storageModule.Delete,
		"list":   storageModule.List,
		"mkdir":  storageModule.MkDir,
	})

	// Database
	databaseModule := modules.NewDatabaseModule(m.modelService, projectID)
	vm.Set("database", map[string]interface{}{
		"collection": databaseModule.Collection,
	})

	// Goals
	goalsModule := modules.NewGoalsModule(m.goalService, projectID)
	vm.Set("goals", map[string]interface{}{
		"increment": goalsModule.Increment,
	})

	// HTTP
	httpModule := modules.NewHTTPModule(m.config.Runtime.Timeout)
	vm.Set("http", map[string]interface{}{
		"get":    httpModule.Get,
		"post":   httpModule.Post,
		"put":    httpModule.Put,
		"delete": httpModule.Delete,
	})

	// Crypto
	cryptoModule := modules.NewCryptoModule()
	vm.Set("crypto", map[string]interface{}{
		"md5":         cryptoModule.MD5,
		"sha256":      cryptoModule.SHA256,
		"randomBytes": cryptoModule.RandomBytes,
	})

	// Encoding
	encodingModule := modules.NewEncodingModule()
	vm.Set("encoding", map[string]interface{}{
		"base64Encode":  encodingModule.Base64Encode,
		"base64Decode":  encodingModule.Base64Decode,
		"jsonParse":     encodingModule.JSONParse,
		"jsonStringify": encodingModule.JSONStringify,
		"urlEncode":     encodingModule.URLEncode,
		"urlDecode":     encodingModule.URLDecode,
	})

	// Utils
	utilsModule := modules.NewUtilsModule()
	vm.Set("utils", map[string]interface{}{
		"sleep":        utilsModule.Sleep,
		"random":       utilsModule.Random,
		"randomInt":    utilsModule.RandomInt,
		"uuid":         utilsModule.UUID,
		"slugify":      utilsModule.Slugify,
		"truncate":     utilsModule.Truncate,
		"capitalize":   utilsModule.Capitalize,
		"regexMatch":   utilsModule.RegexMatch,
		"regexReplace": utilsModule.RegexReplace,
		"formatDate":   utilsModule.FormatDate,
		"parseDate":    utilsModule.ParseDate,
		"timestamp":    utilsModule.Timestamp,
	})

	// Delayed
	delayedModule := modules.NewDelayedModule(m.config.Runtime.WorkerPoolSize)
	vm.Set("delayed", map[string]interface{}{
		"run": delayedModule.Run,
	})

	// Console (alias for logger)
	vm.Set("console", map[string]interface{}{
		"log":   loggerModule.Info,
		"info":  loggerModule.Info,
		"warn":  loggerModule.Warn,
		"error": loggerModule.Error,
		"debug": loggerModule.Debug,
	})

	return nil
}

// GetTypeDefinitions returns TypeScript definitions for Monaco
func (m *Manager) GetTypeDefinitions() string {
	baseTypes := modules.GetBaseTypeDefinitions()
	pluginTypes := m.plugins.GetTypeDefinitions()
	return baseTypes + "\n" + pluginTypes
}
