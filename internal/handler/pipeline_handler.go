package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/middleware"
	"m3m/internal/service"
)

type PipelineHandler struct {
	pipelineService *service.PipelineService
	projectService  *service.ProjectService
}

func NewPipelineHandler(pipelineService *service.PipelineService, projectService *service.ProjectService) *PipelineHandler {
	return &PipelineHandler{
		pipelineService: pipelineService,
		projectService:  projectService,
	}
}

func (h *PipelineHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	pipeline := r.Group("/projects/:id/pipeline")
	pipeline.Use(authMiddleware.Authenticate())
	{
		// Branches
		pipeline.GET("/branches", h.ListBranches)
		pipeline.POST("/branches", h.CreateBranch)
		pipeline.GET("/branches/:name", h.GetBranch)
		pipeline.PUT("/branches/:name", h.UpdateBranch)
		pipeline.POST("/branches/:name/reset", h.ResetBranch)
		pipeline.DELETE("/branches/:name", h.DeleteBranch)

		// Releases
		pipeline.GET("/releases", h.ListReleases)
		pipeline.POST("/releases", h.CreateRelease)
		pipeline.DELETE("/releases/:version", h.DeleteRelease)
		pipeline.POST("/releases/:version/activate", h.ActivateRelease)
	}
}

func (h *PipelineHandler) checkAccess(c *gin.Context) (primitive.ObjectID, bool) {
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

// Branches

func (h *PipelineHandler) ListBranches(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	branches, err := h.pipelineService.GetBranchSummaries(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branches)
}

func (h *PipelineHandler) CreateBranch(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req domain.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch, err := h.pipelineService.CreateBranch(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, branch)
}

func (h *PipelineHandler) GetBranch(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	name := c.Param("name")
	branch, err := h.pipelineService.GetBranch(c.Request.Context(), projectID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "branch not found"})
		return
	}

	c.JSON(http.StatusOK, branch)
}

func (h *PipelineHandler) UpdateBranch(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	name := c.Param("name")
	var req domain.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch, err := h.pipelineService.UpdateBranch(c.Request.Context(), projectID, name, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branch)
}

func (h *PipelineHandler) ResetBranch(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	name := c.Param("name")
	var req domain.ResetBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch, err := h.pipelineService.ResetBranch(c.Request.Context(), projectID, name, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, branch)
}

func (h *PipelineHandler) DeleteBranch(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	name := c.Param("name")
	if name == "develop" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete develop branch"})
		return
	}

	if err := h.pipelineService.DeleteBranch(c.Request.Context(), projectID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "branch deleted successfully"})
}

// Releases

func (h *PipelineHandler) ListReleases(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	releases, err := h.pipelineService.GetReleaseSummaries(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, releases)
}

func (h *PipelineHandler) CreateRelease(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req domain.CreateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	release, err := h.pipelineService.CreateRelease(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, release)
}

func (h *PipelineHandler) DeleteRelease(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	version := c.Param("version")
	if err := h.pipelineService.DeleteRelease(c.Request.Context(), projectID, version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "release deleted successfully"})
}

func (h *PipelineHandler) ActivateRelease(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	version := c.Param("version")
	if err := h.pipelineService.ActivateRelease(c.Request.Context(), projectID, version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "release activated successfully"})
}
