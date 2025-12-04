package handler

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	runtime := r.Group("/projects/:projectId")
	runtime.Use(authMiddleware.Authenticate())
	{
		runtime.POST("/start", h.Start)
		runtime.POST("/stop", h.Stop)
		runtime.POST("/restart", h.Restart)
		runtime.GET("/status", h.Status)
		runtime.GET("/logs", h.Logs)
		runtime.GET("/logs/download", h.DownloadLogs)
	}

	// Runtime types for Monaco
	r.GET("/runtime/types", h.Types)

	// Plugins info
	r.GET("/plugins", h.ListPlugins)

	// Project routes (public, handled by runtime router)
	r.Any("/r/:projectSlug/*route", h.HandleRoute)
}

func (h *RuntimeHandler) checkAccess(c *gin.Context) (primitive.ObjectID, bool) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("projectId"))
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
		Version string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get release code
	var code string
	if req.Version != "" {
		release, err := h.pipelineService.GetRelease(c.Request.Context(), projectID, req.Version)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
			return
		}
		code = release.Code

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
		} else {
			code = release.Code
		}
	}

	// Clear old logs
	h.storageService.ClearLogs(projectID.Hex())

	// Start runtime
	if err := h.runtimeManager.Start(c.Request.Context(), projectID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update project status
	h.projectService.UpdateStatus(c.Request.Context(), projectID, domain.ProjectStatusRunning)

	c.JSON(http.StatusOK, gin.H{"message": "project started"})
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

	// Update project status
	h.projectService.UpdateStatus(c.Request.Context(), projectID, domain.ProjectStatusStopped)

	c.JSON(http.StatusOK, gin.H{"message": "project stopped"})
}

func (h *RuntimeHandler) Restart(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	// Stop first
	h.runtimeManager.Stop(projectID)

	// Get current active release
	project, err := h.projectService.GetByID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	var code string
	if project.ActiveRelease != "" {
		release, err := h.pipelineService.GetRelease(c.Request.Context(), projectID, project.ActiveRelease)
		if err == nil {
			code = release.Code
		}
	}

	if code == "" {
		branch, err := h.pipelineService.GetBranch(c.Request.Context(), projectID, "develop")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no code to run"})
			return
		}
		code = branch.Code
	}

	// Clear old logs
	h.storageService.ClearLogs(projectID.Hex())

	// Start runtime
	if err := h.runtimeManager.Start(c.Request.Context(), projectID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project restarted"})
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
	if runtime, ok := h.runtimeManager.GetRuntime(projectID); ok {
		t := runtime.StartedAt.Format("2006-01-02 15:04:05")
		startedAt = &t
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"started_at": startedAt,
	})
}

func (h *RuntimeHandler) Logs(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	logsPath := h.storageService.GetLogsPath(projectID.Hex())
	entries, err := os.ReadDir(logsPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"logs": []string{}})
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
		c.JSON(http.StatusOK, gin.H{"logs": []string{}})
		return
	}

	// Read log file
	file, err := os.Open(latestLog)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	var logs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}

	// Return last 1000 lines max
	if len(logs) > 1000 {
		logs = logs[len(logs)-1000:]
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
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
