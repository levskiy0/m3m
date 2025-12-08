package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	"m3m/internal/config"
	"m3m/internal/middleware"
	ws "m3m/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type WebSocketHandler struct {
	hub    *ws.Hub
	config *config.Config
}

func NewWebSocketHandler(hub *ws.Hub, cfg *config.Config) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		config: cfg,
	}
}

func (h *WebSocketHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// WebSocket endpoint - custom auth via query param (can't use headers with WS)
	r.GET("/ws", h.HandleWebSocket)
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter (WebSocket can't send custom headers)
	token := c.Query("token")
	if token == "" {
		// Also try Authorization header for compatibility
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	// Validate token
	userID, err := h.validateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(h.hub, conn, userID)

	// Register client
	h.hub.Register(client)

	// Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}

func (h *WebSocketHandler) validateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWT.Secret), nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}

// GetHub returns the WebSocket hub
func (h *WebSocketHandler) GetHub() *ws.Hub {
	return h.hub
}
