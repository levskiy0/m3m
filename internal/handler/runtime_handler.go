package handler

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/middleware"
	"m3m/internal/plugin"
	"m3m/internal/runtime"
	"m3m/internal/runtime/modules"
	"m3m/internal/service"
)

type RuntimeHandler struct {
	runtimeManager  *runtime.Manager
	projectService  *service.ProjectService
	pipelineService *service.PipelineService
	storageService  *service.StorageService
	pluginLoader    *plugin.Loader
}

func NewRuntimeHandler(
	runtimeManager *runtime.Manager,
	projectService *service.ProjectService,
	pipelineService *service.PipelineService,
	storageService *service.StorageService,
	pluginLoader *plugin.Loader,
) *RuntimeHandler {
	return &RuntimeHandler{
		runtimeManager:  runtimeManager,
		projectService:  projectService,
		pipelineService: pipelineService,
		storageService:  storageService,
		pluginLoader:    pluginLoader,
	}
}

func (h *RuntimeHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// Runtime management
	runtime := r.Group("/projects/:id")
	runtime.Use(authMiddleware.Authenticate())
	{
		runtime.POST("/start", h.Start)
		runtime.POST("/stop", h.Stop)
		runtime.POST("/restart", h.Restart)
		runtime.GET("/status", h.Status)
		runtime.GET("/monitor", h.Monitor)
		runtime.GET("/logs", h.Logs)
		runtime.GET("/logs/download", h.DownloadLogs)
	}

	// Runtime types for Monaco
	r.GET("/runtime/types", h.Types)

	// Plugins info
	r.GET("/plugins", h.ListPlugins)

	// System info (requires auth)
	system := r.Group("/system")
	system.Use(authMiddleware.Authenticate())
	{
		system.GET("/info", h.SystemInfo)
	}
}

// RegisterPublicRoutes registers public routes on the root router (not under /api)
func (h *RuntimeHandler) RegisterPublicRoutes(r *gin.Engine) {
	// Project routes (public, handled by runtime router)
	r.Any("/r/:projectSlug/*route", h.HandleRoute)
}

func (h *RuntimeHandler) checkAccess(c *gin.Context) (primitive.ObjectID, bool) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return primitive.NilObjectID, false
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return primitive.NilObjectID, false
	}

	return projectID, true
}

func (h *RuntimeHandler) Start(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req struct {
		Version string `json:"version"` // Release version to run
		Branch  string `json:"branch"`  // Branch name to run (debug mode)
	}
	// Body is optional - ignore EOF errors
	c.ShouldBindJSON(&req)

	var code string
	var runningSource string

	if req.Branch != "" {
		// Debug mode: run code from branch
		branch, err := h.pipelineService.GetBranch(c.Request.Context(), projectID, req.Branch)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "branch not found"})
			return
		}
		code = branch.Code
		runningSource = "debug:" + req.Branch
	} else if req.Version != "" {
		// Run specific release version
		release, err := h.pipelineService.GetRelease(c.Request.Context(), projectID, req.Version)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
			return
		}
		code = release.Code
		runningSource = "release:" + req.Version

		// Activate this release
		h.pipelineService.ActivateRelease(c.Request.Context(), projectID, req.Version)
		h.projectService.SetActiveRelease(c.Request.Context(), projectID, req.Version)
	} else {
		// Use active release or develop branch
		release, err := h.pipelineService.GetActiveRelease(c.Request.Context(), projectID)
		if err != nil {
			// Fallback to develop branch
			branch, err := h.pipelineService.GetBranch(c.Request.Context(), projectID, "develop")
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "no code to run"})
				return
			}
			code = branch.Code
			runningSource = "debug:develop"
		} else {
			code = release.Code
			runningSource = "release:" + release.Version
		}
	}

	// Clear old logs
	h.storageService.ClearLogs(projectID.Hex())

	// Start runtime with background context (runtime should outlive HTTP request)
	if err := h.runtimeManager.Start(context.Background(), projectID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update project status and running source
	h.projectService.UpdateStatus(c.Request.Context(), projectID, domain.ProjectStatusRunning)
	h.projectService.SetRunningSource(c.Request.Context(), projectID, runningSource)

	c.JSON(http.StatusOK, gin.H{"message": "project started", "runningSource": runningSource})
}

func (h *RuntimeHandler) Stop(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	if err := h.runtimeManager.Stop(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update project status and clear running source
	h.projectService.UpdateStatus(c.Request.Context(), projectID, domain.ProjectStatusStopped)
	h.projectService.SetRunningSource(c.Request.Context(), projectID, "")

	c.JSON(http.StatusOK, gin.H{"message": "project stopped"})
}

func (h *RuntimeHandler) Restart(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	// Stop first
	h.runtimeManager.Stop(projectID)

	// Get current project to check running source
	project, err := h.projectService.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var code string
	runningSource := project.RunningSource

	// Use the same source that was running before
	if strings.HasPrefix(runningSource, "debug:") {
		branchName := strings.TrimPrefix(runningSource, "debug:")
		branch, err := h.pipelineService.GetBranch(c.Request.Context(), projectID, branchName)
		if err == nil {
			code = branch.Code
		}
	} else if strings.HasPrefix(runningSource, "release:") {
		version := strings.TrimPrefix(runningSource, "release:")
		release, err := h.pipelineService.GetRelease(c.Request.Context(), projectID, version)
		if err == nil {
			code = release.Code
		}
	}

	// Fallback to active release or develop branch
	if code == "" {
		if project.ActiveRelease != "" {
			release, err := h.pipelineService.GetRelease(c.Request.Context(), projectID, project.ActiveRelease)
			if err == nil {
				code = release.Code
				runningSource = "release:" + project.ActiveRelease
			}
		}
	}

	if code == "" {
		branch, err := h.pipelineService.GetBranch(c.Request.Context(), projectID, "develop")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no code to run"})
			return
		}
		code = branch.Code
		runningSource = "debug:develop"
	}

	// Clear old logs
	h.storageService.ClearLogs(projectID.Hex())

	// Start runtime with background context (runtime should outlive HTTP request)
	if err := h.runtimeManager.Start(context.Background(), projectID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update running source if it changed
	if runningSource != project.RunningSource {
		h.projectService.SetRunningSource(c.Request.Context(), projectID, runningSource)
	}

	c.JSON(http.StatusOK, gin.H{"message": "project restarted", "runningSource": runningSource})
}

func (h *RuntimeHandler) Status(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	isRunning := h.runtimeManager.IsRunning(projectID)
	status := "stopped"
	if isRunning {
		status = "running"
	}

	var startedAt *string
	if rt, ok := h.runtimeManager.GetRuntime(projectID); ok {
		t := rt.StartedAt.Format("2006-01-02 15:04:05")
		startedAt = &t
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"started_at": startedAt,
	})
}

func (h *RuntimeHandler) Monitor(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	stats, err := h.runtimeManager.GetStats(projectID)
	if err != nil {
		// Project not running - still return storage and database sizes
		basicStats := h.runtimeManager.GetBasicStats(projectID)
		c.JSON(http.StatusOK, basicStats)
		return
	}

	c.JSON(http.StatusOK, stats)
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

func (h *RuntimeHandler) Logs(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	logsPath := h.storageService.GetLogsPath(projectID.Hex())
	entries, err := os.ReadDir(logsPath)
	if err != nil {
		c.JSON(http.StatusOK, []LogEntry{})
		return
	}

	// Find latest log file
	var latestLog string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			latestLog = filepath.Join(logsPath, entry.Name())
		}
	}

	if latestLog == "" {
		c.JSON(http.StatusOK, []LogEntry{})
		return
	}

	// Read log file
	file, err := os.Open(latestLog)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	logs := make([]LogEntry, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLogLine(line)
		logs = append(logs, entry)
	}

	// Return last 1000 entries max
	if len(logs) > 1000 {
		logs = logs[len(logs)-1000:]
	}

	c.JSON(http.StatusOK, logs)
}

// parseLogLine parses log line in format: [2006-01-02 15:04:05] [LEVEL] message
func parseLogLine(line string) LogEntry {
	entry := LogEntry{
		Timestamp: "",
		Level:     "info",
		Message:   line,
	}

	// Try to parse timestamp: [2006-01-02 15:04:05]
	if len(line) > 21 && line[0] == '[' {
		closeBracket := strings.Index(line, "]")
		if closeBracket > 0 {
			// Convert "2006-01-02 15:04:05" to ISO format "2006-01-02T15:04:05"
			ts := line[1:closeBracket]
			entry.Timestamp = strings.Replace(ts, " ", "T", 1)
			line = strings.TrimPrefix(line[closeBracket+1:], " ")

			// Try to parse level: [LEVEL]
			if len(line) > 2 && line[0] == '[' {
				levelEnd := strings.Index(line, "]")
				if levelEnd > 0 {
					level := strings.ToLower(line[1:levelEnd])
					if level == "debug" || level == "info" || level == "warn" || level == "error" {
						entry.Level = level
					}
					entry.Message = strings.TrimPrefix(line[levelEnd+1:], " ")
				}
			}
		}
	}

	return entry
}

func (h *RuntimeHandler) DownloadLogs(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	logsPath := h.storageService.GetLogsPath(projectID.Hex())
	entries, err := os.ReadDir(logsPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no logs found"})
		return
	}

	// Find latest log file
	var latestLog string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			latestLog = filepath.Join(logsPath, entry.Name())
		}
	}

	if latestLog == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "no logs found"})
		return
	}

	c.FileAttachment(latestLog, "project.log")
}

func (h *RuntimeHandler) Types(c *gin.Context) {
	types := h.runtimeManager.GetTypeDefinitions()
	c.String(http.StatusOK, types)
}

func (h *RuntimeHandler) ListPlugins(c *gin.Context) {
	plugins := h.pluginLoader.GetPluginInfos()
	c.JSON(http.StatusOK, plugins)
}

// SystemInfo returns system information including runtime and loaded plugins
func (h *RuntimeHandler) SystemInfo(c *gin.Context) {
	// Get Go runtime info
	var memStats goruntime.MemStats
	goruntime.ReadMemStats(&memStats)

	// Get running projects
	runningProjects := h.runtimeManager.GetRunningProjects()

	// Get plugins
	plugins := h.pluginLoader.GetPluginInfos()

	info := gin.H{
		"version":       "1.0.0",
		"go_version":    goruntime.Version(),
		"go_os":         goruntime.GOOS,
		"go_arch":       goruntime.GOARCH,
		"num_cpu":       goruntime.NumCPU(),
		"num_goroutine": goruntime.NumGoroutine(),
		"memory": gin.H{
			"alloc":       memStats.Alloc,
			"total_alloc": memStats.TotalAlloc,
			"sys":         memStats.Sys,
			"num_gc":      memStats.NumGC,
		},
		"running_projects_count": len(runningProjects),
		"plugins":                plugins,
	}

	c.JSON(http.StatusOK, info)
}

func (h *RuntimeHandler) HandleRoute(c *gin.Context) {
	projectSlug := c.Param("projectSlug")
	route := c.Param("route")

	// Get project by slug
	project, err := h.projectService.GetBySlug(c.Request.Context(), projectSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	// Check if project is running
	if !h.runtimeManager.IsRunning(project.ID) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "project not running"})
		return
	}

	// Build request context
	body, _ := io.ReadAll(c.Request.Body)
	headers := make(map[string]string)
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	query := make(map[string]string)
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	ctx := &modules.RequestContext{
		Method:  c.Request.Method,
		Path:    route,
		Params:  make(map[string]string),
		Query:   query,
		Headers: headers,
		Body:    string(body),
	}

	// Handle route
	resp, err := h.runtimeManager.HandleRoute(project.ID, c.Request.Method, route, ctx)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Set response headers
	if resp.Headers != nil {
		for k, v := range resp.Headers {
			c.Header(k, v)
		}
	}

	c.JSON(resp.Status, resp.Body)
}
