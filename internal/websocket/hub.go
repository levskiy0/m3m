package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	EventMonitor EventType = "monitor"
	EventLog     EventType = "log"
	EventRunning EventType = "running"
	EventGoals   EventType = "goals"
)

// Event represents a WebSocket event message
type Event struct {
	ProjectID string    `json:"projectId"`
	Event     EventData `json:"event"`
}

// EventData represents the event payload
type EventData struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	projectIDs map[string]bool
	mu         sync.RWMutex
	userID     string
}

// Hub manages WebSocket connections and broadcasts
type Hub struct {
	// Registered clients by project ID
	projectClients map[string]map[*Client]bool

	// All clients
	clients map[*Client]bool

	// Register requests
	register chan *Client

	// Unregister requests
	unregister chan *Client

	// Subscribe to project
	subscribe chan *Subscription

	// Unsubscribe from project
	unsubscribe chan *Subscription

	// Broadcast to project subscribers
	broadcast chan *Broadcast

	mu     sync.RWMutex
	logger *slog.Logger
}

// Subscription represents a subscribe/unsubscribe request
type Subscription struct {
	client    *Client
	projectID string
}

// Broadcast represents a message to broadcast to project subscribers
type Broadcast struct {
	ProjectID string
	Event     EventData
}

// NewHub creates a new Hub
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		projectClients: make(map[string]map[*Client]bool),
		clients:        make(map[*Client]bool),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		subscribe:      make(chan *Subscription),
		unsubscribe:    make(chan *Subscription),
		broadcast:      make(chan *Broadcast, 256),
		logger:         logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			h.logger.Debug("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				client.mu.RLock()
				for projectID := range client.projectIDs {
					if clients, ok := h.projectClients[projectID]; ok {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.projectClients, projectID)
						}
					}
				}
				client.mu.RUnlock()

				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Debug("WebSocket client disconnected")

		case sub := <-h.subscribe:
			h.mu.Lock()
			if _, ok := h.projectClients[sub.projectID]; !ok {
				h.projectClients[sub.projectID] = make(map[*Client]bool)
			}
			h.projectClients[sub.projectID][sub.client] = true
			clientCount := len(h.projectClients[sub.projectID])
			h.mu.Unlock()

			sub.client.mu.Lock()
			sub.client.projectIDs[sub.projectID] = true
			sub.client.mu.Unlock()

			h.logger.Info("Client subscribed to project", "projectId", sub.projectID, "subscriberCount", clientCount)

		case sub := <-h.unsubscribe:
			h.mu.Lock()
			if clients, ok := h.projectClients[sub.projectID]; ok {
				delete(clients, sub.client)
				if len(clients) == 0 {
					delete(h.projectClients, sub.projectID)
				}
			}
			h.mu.Unlock()

			sub.client.mu.Lock()
			delete(sub.client.projectIDs, sub.projectID)
			sub.client.mu.Unlock()

			h.logger.Debug("Client unsubscribed from project", "projectId", sub.projectID)

		case broadcast := <-h.broadcast:
			h.mu.RLock()
			clients := h.projectClients[broadcast.ProjectID]
			h.mu.RUnlock()

			if len(clients) == 0 {
				continue
			}

			event := Event{
				ProjectID: broadcast.ProjectID,
				Event:     broadcast.Event,
			}

			message, err := json.Marshal(event)
			if err != nil {
				h.logger.Error("Failed to marshal event", "error", err)
				continue
			}

			for client := range clients {
				select {
				case client.send <- message:
				default:
					h.logger.Warn("Client buffer full, skipping message")
				}
			}
		}
	}
}

// Register adds a client to the hub
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// BroadcastToProject sends an event to all clients subscribed to a project
func (h *Hub) BroadcastToProject(projectID string, eventType EventType, data interface{}) {
	h.broadcast <- &Broadcast{
		ProjectID: projectID,
		Event: EventData{
			Type: eventType,
			Data: data,
		},
	}
}

// GetSubscriberCount returns the number of subscribers for a project
func (h *Hub) GetSubscriberCount(projectID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.projectClients[projectID])
}

// HasSubscribers returns true if project has any subscribers
func (h *Hub) HasSubscribers(projectID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.projectClients[projectID]) > 0
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		projectIDs: make(map[string]bool),
		userID:     userID,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		// Parse client message (subscribe/unsubscribe)
		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.hub.logger.Warn("Invalid client message", "error", err)
			continue
		}

		switch msg.Action {
		case "subscribe":
			if msg.ProjectID != "" {
				c.hub.subscribe <- &Subscription{client: c, projectID: msg.ProjectID}
			}
		case "unsubscribe":
			if msg.ProjectID != "" {
				c.hub.unsubscribe <- &Subscription{client: c, projectID: msg.ProjectID}
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ClientMessage represents a message from the client
type ClientMessage struct {
	Action    string `json:"action"` // "subscribe" or "unsubscribe"
	ProjectID string `json:"projectId"`
}
