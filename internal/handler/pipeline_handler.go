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
		pipeline.GET("/branches", h.ListBranches)
		pipeline.POST("/branches", h.CreateBranch)
		pipeline.GET("/branches/:branchId", h.GetBranch)
		pipeline.PUT("/branches/:branchId", h.UpdateBranch)
		pipeline.POST("/branches/:branchId/reset", h.ResetBranch)
		pipeline.DELETE("/branches/:branchId", h.DeleteBranch)

		pipeline.GET("/releases", h.ListReleases)
		pipeline.POST("/releases", h.CreateRelease)
		pipeline.DELETE("/releases/:releaseId", h.DeleteRelease)
		pipeline.POST("/releases/:releaseId/activate", h.ActivateRelease)
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

	branchID, err := primitive.ObjectIDFromHex(c.Param("branchId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}

	branch, err := h.pipelineService.GetBranchByID(c.Request.Context(), projectID, branchID)
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

	branchID, err := primitive.ObjectIDFromHex(c.Param("branchId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}

	var req domain.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch, err := h.pipelineService.UpdateBranchByID(c.Request.Context(), projectID, branchID, &req)
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

	branchID, err := primitive.ObjectIDFromHex(c.Param("branchId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}

	var req domain.ResetBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	branch, err := h.pipelineService.ResetBranchByID(c.Request.Context(), projectID, branchID, &req)
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

	branchID, err := primitive.ObjectIDFromHex(c.Param("branchId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid branch id"})
		return
	}

	if err := h.pipelineService.DeleteBranchByID(c.Request.Context(), projectID, branchID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "branch deleted successfully"})
}

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

	releaseID, err := primitive.ObjectIDFromHex(c.Param("releaseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
		return
	}

	if err := h.pipelineService.DeleteReleaseByID(c.Request.Context(), projectID, releaseID); err != nil {
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

	releaseID, err := primitive.ObjectIDFromHex(c.Param("releaseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release id"})
		return
	}

	if err := h.pipelineService.ActivateReleaseByID(c.Request.Context(), projectID, releaseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "release activated successfully"})
}
