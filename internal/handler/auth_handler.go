package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"m3m/internal/domain"
	"m3m/internal/middleware"
	"m3m/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", h.Login)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RegisterProtected(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	auth := r.Group("/auth")
	auth.Use(authMiddleware.Authenticate())
	{
		auth.POST("/logout", h.Logout)
	}
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// With JWT, logout is typically handled client-side by removing the token
	// Server-side could implement a token blacklist if needed
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
