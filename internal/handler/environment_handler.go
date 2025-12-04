package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/middleware"
	"m3m/internal/service"
)

type EnvironmentHandler struct {
	envService     *service.EnvironmentService
	projectService *service.ProjectService
}

func NewEnvironmentHandler(envService *service.EnvironmentService, projectService *service.ProjectService) *EnvironmentHandler {
	return &EnvironmentHandler{
		envService:     envService,
		projectService: projectService,
	}
}

func (h *EnvironmentHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	env := r.Group("/projects/:projectId/env")
	env.Use(authMiddleware.Authenticate())
	{
		env.GET("", h.List)
		env.POST("", h.Create)
		env.PUT("/:key", h.Update)
		env.DELETE("/:key", h.Delete)
	}
}

func (h *EnvironmentHandler) checkAccess(c *gin.Context) (primitive.ObjectID, bool) {
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

func (h *EnvironmentHandler) List(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	envVars, err := h.envService.GetByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, envVars)
}

func (h *EnvironmentHandler) Create(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req domain.CreateEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	envVar, err := h.envService.Create(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, envVar)
}

func (h *EnvironmentHandler) Update(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	key := c.Param("key")
	var req domain.UpdateEnvVarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	envVar, err := h.envService.Update(c.Request.Context(), projectID, key, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, envVar)
}

func (h *EnvironmentHandler) Delete(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	key := c.Param("key")
	if err := h.envService.Delete(c.Request.Context(), projectID, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "environment variable deleted successfully"})
}
