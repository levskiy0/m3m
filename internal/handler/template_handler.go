package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/levskiy0/m3m/internal/constants"
	"github.com/levskiy0/m3m/internal/middleware"
)

type TemplateHandler struct{}

func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{}
}

// Register registers template routes
func (h *TemplateHandler) Register(api *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	templates := api.Group("/templates")
	templates.Use(authMiddleware.Authenticate())
	{
		templates.GET("/service", h.GetServiceTemplate)
	}
}

// GetServiceTemplate returns the default service code template
func (h *TemplateHandler) GetServiceTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": constants.DefaultServiceCode,
	})
}
