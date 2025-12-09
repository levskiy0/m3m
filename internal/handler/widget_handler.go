package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/middleware"
	"github.com/levskiy0/m3m/internal/service"
)

type WidgetHandler struct {
	widgetService  *service.WidgetService
	projectService *service.ProjectService
}

func NewWidgetHandler(widgetService *service.WidgetService, projectService *service.ProjectService) *WidgetHandler {
	return &WidgetHandler{
		widgetService:  widgetService,
		projectService: projectService,
	}
}

func (h *WidgetHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	widgets := r.Group("/projects/:id/widgets")
	widgets.Use(authMiddleware.Authenticate())
	{
		widgets.GET("", h.List)
		widgets.POST("", h.Create)
		widgets.PUT("/:widgetId", h.Update)
		widgets.DELETE("/:widgetId", h.Delete)
		widgets.POST("/reorder", h.Reorder)
	}
}

func (h *WidgetHandler) List(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	widgets, err := h.widgetService.GetByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, widgets)
}

func (h *WidgetHandler) Create(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req domain.CreateWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	widget, err := h.widgetService.Create(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, widget)
}

func (h *WidgetHandler) Update(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	widgetID, err := primitive.ObjectIDFromHex(c.Param("widgetId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid widget id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req domain.UpdateWidgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	widget, err := h.widgetService.Update(c.Request.Context(), projectID, widgetID, &req)
	if err != nil {
		if err == service.ErrWidgetNotInProject {
			c.JSON(http.StatusForbidden, gin.H{"error": "widget does not belong to this project"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, widget)
}

func (h *WidgetHandler) Delete(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	widgetID, err := primitive.ObjectIDFromHex(c.Param("widgetId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid widget id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.widgetService.Delete(c.Request.Context(), projectID, widgetID); err != nil {
		if err == service.ErrWidgetNotInProject {
			c.JSON(http.StatusForbidden, gin.H{"error": "widget does not belong to this project"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "widget deleted successfully"})
}

func (h *WidgetHandler) Reorder(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req domain.ReorderWidgetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.widgetService.Reorder(c.Request.Context(), projectID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "widgets reordered successfully"})
}
