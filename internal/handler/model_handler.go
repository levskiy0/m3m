package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/middleware"
	"m3m/internal/service"
)

type ModelHandler struct {
	modelService   *service.ModelService
	projectService *service.ProjectService
}

func NewModelHandler(modelService *service.ModelService, projectService *service.ProjectService) *ModelHandler {
	return &ModelHandler{
		modelService:   modelService,
		projectService: projectService,
	}
}

func (h *ModelHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	models := r.Group("/projects/:id/models")
	models.Use(authMiddleware.Authenticate())
	{
		models.GET("", h.List)
		models.POST("", h.Create)
		models.GET("/:modelId", h.Get)
		models.PUT("/:modelId", h.Update)
		models.DELETE("/:modelId", h.Delete)

		// Data routes
		models.GET("/:modelId/data", h.ListData)
		models.POST("/:modelId/data", h.CreateData)
		models.POST("/:modelId/data/query", h.QueryData) // Advanced filtering
		models.GET("/:modelId/data/:dataId", h.GetData)
		models.PUT("/:modelId/data/:dataId", h.UpdateData)
		models.DELETE("/:modelId/data/:dataId", h.DeleteData)
		models.POST("/:modelId/data/bulk-delete", h.BulkDeleteData) // Bulk delete
	}
}

func (h *ModelHandler) checkAccess(c *gin.Context) (primitive.ObjectID, bool) {
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

func (h *ModelHandler) List(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	models, err := h.modelService.GetByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models)
}

func (h *ModelHandler) Create(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req domain.CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.modelService.Create(c.Request.Context(), projectID, &req)
	if err != nil {
		var validationErr service.ValidationErrors
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": validationErr.Errors,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, model)
}

func (h *ModelHandler) Get(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	model, err := h.modelService.GetByID(c.Request.Context(), modelID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
		return
	}

	c.JSON(http.StatusOK, model)
}

func (h *ModelHandler) Update(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var req domain.UpdateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model, err := h.modelService.Update(c.Request.Context(), modelID, &req)
	if err != nil {
		var validationErr service.ValidationErrors
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": validationErr.Errors,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model)
}

func (h *ModelHandler) Delete(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	if err := h.modelService.Delete(c.Request.Context(), modelID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "model deleted successfully"})
}

// Data handlers

func (h *ModelHandler) ListData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var query domain.DataQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse filter params from query string (format: filter[field]=value)
	if query.Filters == nil {
		query.Filters = make(map[string]string)
	}
	for key, values := range c.Request.URL.Query() {
		if len(key) > 7 && key[:7] == "filter[" && key[len(key)-1] == ']' {
			field := key[7 : len(key)-1]
			if len(values) > 0 {
				query.Filters[field] = values[0]
			}
		}
	}

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Page <= 0 {
		query.Page = 1
	}

	data, total, err := h.modelService.GetData(c.Request.Context(), modelID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pagination metadata
	totalPages := int64(0)
	if query.Limit > 0 {
		totalPages = (total + int64(query.Limit) - 1) / int64(query.Limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"total":      total,
		"page":       query.Page,
		"limit":      query.Limit,
		"totalPages": totalPages,
	})
}

func (h *ModelHandler) CreateData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.modelService.CreateData(c.Request.Context(), modelID, data)
	if err != nil {
		var validationErr service.ValidationErrors
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": validationErr.Errors,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// QueryData handles advanced data queries with filtering
func (h *ModelHandler) QueryData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var query domain.AdvancedDataQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Page <= 0 {
		query.Page = 1
	}

	data, total, err := h.modelService.GetDataAdvanced(c.Request.Context(), modelID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pagination metadata
	totalPages := int64(0)
	if query.Limit > 0 {
		totalPages = (total + int64(query.Limit) - 1) / int64(query.Limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"total":      total,
		"page":       query.Page,
		"limit":      query.Limit,
		"totalPages": totalPages,
	})
}

func (h *ModelHandler) GetData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	dataID, err := primitive.ObjectIDFromHex(c.Param("dataId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data id"})
		return
	}

	data, err := h.modelService.GetDataByID(c.Request.Context(), modelID, dataID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "data not found"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *ModelHandler) UpdateData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	dataID, err := primitive.ObjectIDFromHex(c.Param("dataId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data id"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.modelService.UpdateData(c.Request.Context(), modelID, dataID, data); err != nil {
		var validationErr service.ValidationErrors
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": validationErr.Errors,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "data updated successfully"})
}

func (h *ModelHandler) DeleteData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	dataID, err := primitive.ObjectIDFromHex(c.Param("dataId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data id"})
		return
	}

	if err := h.modelService.DeleteData(c.Request.Context(), modelID, dataID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "data deleted successfully"})
}

// BulkDeleteData deletes multiple data records by their IDs
func (h *ModelHandler) BulkDeleteData(c *gin.Context) {
	_, ok := h.checkAccess(c)
	if !ok {
		return
	}

	modelID, err := primitive.ObjectIDFromHex(c.Param("modelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model id"})
		return
	}

	var req struct {
		IDs []string `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert string IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data id: " + idStr})
			return
		}
		objectIDs = append(objectIDs, id)
	}

	deletedCount, err := h.modelService.DeleteManyData(c.Request.Context(), modelID, objectIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "data deleted successfully",
		"deleted_count": deletedCount,
	})
}
