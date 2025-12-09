package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/service"
)

const (
	AuthorizationHeader = "Authorization"
	UserContextKey      = "user"
	UserIDContextKey    = "user_id"
)

type AuthMiddleware struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthMiddleware(authService *service.AuthService, userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		userService: userService,
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader(AuthorizationHeader)
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := m.authService.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		user, err := m.userService.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		c.Set(UserContextKey, user)
		c.Set(UserIDContextKey, userID)
		c.Next()
	}
}

func (m *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(UserContextKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		u := user.(*domain.User)

		// Root has all permissions
		if u.IsRoot {
			c.Next()
			return
		}

		switch permission {
		case "create_projects":
			if !u.Permissions.CreateProjects {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
		case "manage_users":
			if !u.Permissions.ManageUsers {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
		}

		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *domain.User {
	user, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	return user.(*domain.User)
}

func GetCurrentUserID(c *gin.Context) primitive.ObjectID {
	userID, exists := c.Get(UserIDContextKey)
	if !exists {
		return primitive.NilObjectID
	}
	return userID.(primitive.ObjectID)
}
