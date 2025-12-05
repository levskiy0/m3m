package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/middleware"
	"m3m/internal/service"
)

type StorageHandler struct {
	storageService *service.StorageService
	projectService *service.ProjectService
}

func NewStorageHandler(storageService *service.StorageService, projectService *service.ProjectService) *StorageHandler {
	return &StorageHandler{
		storageService: storageService,
		projectService: projectService,
	}
}

func (h *StorageHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	storage := r.Group("/projects/:id/storage")
	storage.Use(authMiddleware.Authenticate())
	{
		storage.GET("", h.List)
		storage.POST("/mkdir", h.MkDir)
		storage.POST("/upload", h.Upload)
		storage.GET("/download/*path", h.Download)
		storage.PUT("/rename", h.Rename)
		storage.DELETE("/*path", h.Delete)
		storage.POST("/file", h.CreateFile)
		storage.PUT("/file/*path", h.UpdateFile)
		storage.GET("/thumbnail/*path", h.Thumbnail)
	}
}

// RegisterPublicRoutes registers CDN routes at root level (not under /api)
func (h *StorageHandler) RegisterPublicRoutes(r *gin.Engine) {
	// /cdn/download/{project-id}/path/to/file - force download
	r.GET("/cdn/download/:id/*path", h.PublicDownload)
	// /cdn/resize/{WxH}/{project-id}/path/to/file - resized image
	r.GET("/cdn/resize/:size/:id/*path", h.CDNResize)
	// /cdn/{project-id}/path/to/file - direct view (must be last - catches all)
	r.GET("/cdn/:id/*path", h.CDN)
}

func (h *StorageHandler) checkAccess(c *gin.Context) (string, bool) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return "", false
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return "", false
	}

	return projectID.Hex(), true
}

func (h *StorageHandler) List(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	files, err := h.storageService.List(projectID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, files)
}

func (h *StorageHandler) MkDir(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storageService.MkDir(projectID, req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "directory created successfully"})
}

func (h *StorageHandler) Upload(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	path := c.PostForm("path")
	if path == "" {
		path = "/"
	}

	fullPath := filepath.Join(path, file.Filename)

	if err := h.storageService.Upload(projectID, fullPath, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "file uploaded successfully",
		"path":    fullPath,
	})
}

func (h *StorageHandler) Download(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	filePath, err := h.storageService.Download(projectID, path)
	if err != nil {
		if err == service.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(filePath)
}

func (h *StorageHandler) Rename(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req struct {
		OldPath string `json:"old_path" binding:"required"`
		NewPath string `json:"new_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storageService.Rename(projectID, req.OldPath, req.NewPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "renamed successfully"})
}

func (h *StorageHandler) Delete(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	if err := h.storageService.Delete(projectID, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
}

func (h *StorageHandler) CreateFile(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	var req struct {
		Path    string `json:"path" binding:"required"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storageService.Write(projectID, req.Path, []byte(req.Content)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "file created successfully",
		"path":    req.Path,
	})
}

func (h *StorageHandler) UpdateFile(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	if err := h.storageService.Write(projectID, path, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file updated successfully"})
}

func (h *StorageHandler) Thumbnail(c *gin.Context) {
	projectID, ok := h.checkAccess(c)
	if !ok {
		return
	}

	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	data, err := h.storageService.GenerateThumbnail(projectID, path, 50, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

// CDN serves files publicly without authentication (inline)
func (h *StorageHandler) CDN(c *gin.Context) {
	projectID := c.Param("id")
	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	filePath, err := h.storageService.Download(projectID, path)
	if err != nil {
		if err == service.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(filePath)
}

// PublicDownload serves files with Content-Disposition: attachment (force download)
func (h *StorageHandler) PublicDownload(c *gin.Context) {
	projectID := c.Param("id")
	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	filePath, err := h.storageService.Download(projectID, path)
	if err != nil {
		if err == service.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.FileAttachment(filePath, filepath.Base(path))
}

// CDNResize serves resized images (e.g., /cdn/50x50/{project-id}/path/to/image.png)
func (h *StorageHandler) CDNResize(c *gin.Context) {
	sizeParam := c.Param("size")
	projectID := c.Param("id")
	path := c.Param("path")
	path = strings.TrimPrefix(path, "/")

	// Check if size param is "download" - redirect to PublicDownload
	if sizeParam == "download" {
		h.PublicDownload(c)
		return
	}

	// Parse size (e.g., "50x50", "100x100")
	var width, height int
	if _, err := fmt.Sscanf(sizeParam, "%dx%d", &width, &height); err != nil {
		// Not a valid size format, treat as regular CDN request
		// This shouldn't happen with proper routing, but handle gracefully
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size format, use WxH (e.g., 50x50)"})
		return
	}

	if width <= 0 || height <= 0 || width > 2000 || height > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dimensions"})
		return
	}

	filePath, err := h.storageService.GetResizedImage(projectID, path, width, height)
	if err != nil {
		if err == service.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.File(filePath)
}
