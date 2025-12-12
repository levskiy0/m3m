package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/middleware"
	"github.com/levskiy0/m3m/internal/repository"
	"github.com/levskiy0/m3m/internal/runtime"
	"github.com/levskiy0/m3m/internal/service"
)

type ActionHandler struct {
	actionService  *service.ActionService
	projectService *service.ProjectService
	runtimeManager *runtime.Manager
}

func NewActionHandler(
	actionService *service.ActionService,
	projectService *service.ProjectService,
	runtimeManager *runtime.Manager,
) *ActionHandler {
	return &ActionHandler{
		actionService:  actionService,
		projectService: projectService,
		runtimeManager: runtimeManager,
	}
}

func (h *ActionHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	actions := r.Group("/projects/:id/actions")
	actions.Use(authMiddleware.Authenticate())
	{
		actions.GET("", h.List)
		actions.POST("", h.Create)
		actions.PUT("/:actionId", h.Update)
		actions.DELETE("/:actionId", h.Delete)
		actions.GET("/states", h.GetStates)
		actions.POST("/:actionSlug/trigger", h.Trigger)
	}
}

func (h *ActionHandler) checkAccess(c *gin.Context) (primitive.ObjectID, bool) {
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

func (h *ActionHandler) List(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	actions, err := h.actionService.GetByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, actions)
}

func (h *ActionHandler) Create(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req domain.CreateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	action, err := h.actionService.Create(c.Request.Context(), projectID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrActionExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "action with this slug already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, action)
}

func (h *ActionHandler) Update(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	actionID, err := primitive.ObjectIDFromHex(c.Param("actionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}

	// Verify action belongs to this project
	action, err := h.actionService.GetByID(c.Request.Context(), actionID)
	if err != nil {
		if errors.Is(err, repository.ErrActionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "action not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if action.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "action does not belong to this project"})
		return
	}

	var req domain.UpdateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedAction, err := h.actionService.Update(c.Request.Context(), actionID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedAction)
}

func (h *ActionHandler) Delete(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	actionID, err := primitive.ObjectIDFromHex(c.Param("actionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action id"})
		return
	}

	// Verify action belongs to this project
	action, err := h.actionService.GetByID(c.Request.Context(), actionID)
	if err != nil {
		if errors.Is(err, repository.ErrActionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "action not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if action.ProjectID != projectID {
		c.JSON(http.StatusForbidden, gin.H{"error": "action does not belong to this project"})
		return
	}

	if err := h.actionService.Delete(c.Request.Context(), actionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "action deleted successfully"})
}

func (h *ActionHandler) GetStates(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	states := h.runtimeManager.GetActionStates(projectID)
	if states == nil {
		states = []domain.ActionRuntimeState{}
	}

	c.JSON(http.StatusOK, states)
}

// TriggerActionRequest is the request body for action trigger
type TriggerActionRequest struct {
	SessionID string `json:"sessionId"`
}

// Trigger triggers an action with session context for $ui dialogs
func (h *ActionHandler) Trigger(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	actionSlug := c.Param("actionSlug")

	// Parse request body for sessionId
	var req TriggerActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// sessionId is optional for backward compatibility
		req.SessionID = ""
	}

	// Verify action exists
	_, err := h.actionService.GetBySlug(c.Request.Context(), projectID, actionSlug)
	if err != nil {
		if errors.Is(err, repository.ErrActionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "action not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if project is running
	if !h.runtimeManager.IsRunning(projectID) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "project not running"})
		return
	}

	// Check action state - only allow triggering if enabled
	state, ok := h.runtimeManager.GetActionState(projectID, actionSlug)
	if ok && state != domain.ActionStateEnabled {
		c.JSON(http.StatusConflict, gin.H{"error": "action is " + string(state)})
		return
	}

	// Get current user ID for action context
	user := middleware.GetCurrentUser(c)
	userID := user.ID.Hex()

	// Trigger action with session context
	if err := h.runtimeManager.TriggerActionWithSession(projectID, actionSlug, userID, req.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "action triggered successfully"})
}
