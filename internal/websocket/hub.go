package websocket

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// EventType represents the type of WebSocket event
type EventType string

const (
	EventMonitor   EventType = "monitor"
	EventLog       EventType = "log"
	EventRunning   EventType = "running"
	EventGoals     EventType = "goals"
	EventActions   EventType = "actions"
	EventUIRequest EventType = "ui_request"
	EventTime      EventType = "time"
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

// GlobalEvent represents a global event (not tied to a project)
type GlobalEvent struct {
	Event EventData `json:"event"`
}

// Client represents a connected WebSocket client
type Client struct {
	hub        *Hub
	conn       *websocket.Conn
	send       chan []byte
	projectIDs map[string]bool
	mu         sync.RWMutex
	userID     string
	sessionID  string // unique session ID for this connection
}

// Hub manages WebSocket connections and broadcasts
type Hub struct {
	// Registered clients by project ID
	projectClients map[string]map[*Client]bool

	// All clients
	clients map[*Client]bool

	// Clients by session ID (for targeted UI requests)
	sessionClients map[string]*Client

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

	// Broadcast to all clients
	globalBroadcast chan *EventData

	// UI response handler
	uiResponseHandler UIResponseHandler

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
		projectClients:  make(map[string]map[*Client]bool),
		clients:         make(map[*Client]bool),
		sessionClients:  make(map[string]*Client),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		subscribe:       make(chan *Subscription),
		unsubscribe:     make(chan *Subscription),
		broadcast:       make(chan *Broadcast, 256),
		globalBroadcast: make(chan *EventData, 64),
		logger:          logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.sessionClients[client.sessionID] = client
			h.mu.Unlock()
			h.logger.Debug("WebSocket client connected", "sessionId", client.sessionID)

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
				delete(h.sessionClients, client.sessionID)
				close(client.send)
			}
			h.mu.Unlock()
			h.logger.Debug("WebSocket client disconnected", "sessionId", client.sessionID)

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

		case eventData := <-h.globalBroadcast:
			h.mu.RLock()
			allClients := make([]*Client, 0, len(h.clients))
			for client := range h.clients {
				allClients = append(allClients, client)
			}
			h.mu.RUnlock()

			if len(allClients) == 0 {
				continue
			}

			globalEvent := GlobalEvent{
				Event: *eventData,
			}

			message, err := json.Marshal(globalEvent)
			if err != nil {
				h.logger.Error("Failed to marshal global event", "error", err)
				continue
			}

			for _, client := range allClients {
				select {
				case client.send <- message:
				default:
					// skip
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

// BroadcastToAll sends an event to all connected clients
func (h *Hub) BroadcastToAll(eventType EventType, data interface{}) {
	h.globalBroadcast <- &EventData{
		Type: eventType,
		Data: data,
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

// SetUIResponseHandler sets the handler for UI responses
func (h *Hub) SetUIResponseHandler(handler UIResponseHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.uiResponseHandler = handler
}

// SendToUser sends an event to a specific user subscribed to a project
func (h *Hub) SendToUser(projectID, userID string, eventType EventType, data interface{}) {
	h.mu.RLock()
	clients := h.projectClients[projectID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		h.logger.Warn("No clients subscribed to project", "projectId", projectID)
		return
	}

	event := Event{
		ProjectID: projectID,
		Event: EventData{
			Type: eventType,
			Data: data,
		},
	}

	message, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal event", "error", err)
		return
	}

	sent := false
	for client := range clients {
		if client.userID == userID {
			select {
			case client.send <- message:
				sent = true
			default:
				h.logger.Warn("Client buffer full, skipping message")
			}
		}
	}

	if !sent {
		h.logger.Warn("User not subscribed to project", "projectId", projectID, "userId", userID)
	}
}

// SendToSession sends an event to a specific session (for targeted UI requests)
func (h *Hub) SendToSession(projectID, sessionID string, eventType EventType, data interface{}) {
	h.mu.RLock()
	client, ok := h.sessionClients[sessionID]
	h.mu.RUnlock()

	if !ok {
		h.logger.Warn("Session not found", "sessionId", sessionID)
		return
	}

	// Verify client is subscribed to this project
	client.mu.RLock()
	subscribed := client.projectIDs[projectID]
	client.mu.RUnlock()

	if !subscribed {
		h.logger.Warn("Session not subscribed to project", "sessionId", sessionID, "projectId", projectID)
		return
	}

	event := Event{
		ProjectID: projectID,
		Event: EventData{
			Type: eventType,
			Data: data,
		},
	}

	message, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal event", "error", err)
		return
	}

	select {
	case client.send <- message:
		// sent successfully
	default:
		h.logger.Warn("Client buffer full, skipping message")
	}
}

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (increased for ui_response form data)
	maxMessageSize = 8192
)

// NewClient creates a new WebSocket client with a unique session ID
func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		hub:        hub,
		conn:       conn,
		send:       make(chan []byte, 256),
		projectIDs: make(map[string]bool),
		userID:     userID,
		sessionID:  uuid.New().String(),
	}
}

// SessionID returns the client's unique session ID
func (c *Client) SessionID() string {
	return c.sessionID
}

// SendSessionInfo sends the session ID to the client on connect
func (c *Client) SendSessionInfo() {
	msg := struct {
		Type      string `json:"type"`
		SessionID string `json:"sessionId"`
	}{
		Type:      "session",
		SessionID: c.sessionID,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	// Write directly to connection (before pumps start)
	c.conn.WriteMessage(websocket.TextMessage, data)
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
		case "ui_response":
			if msg.ProjectID != "" && msg.RequestID != "" {
				c.hub.mu.RLock()
				handler := c.hub.uiResponseHandler
				c.hub.mu.RUnlock()
				if handler != nil {
					go handler(msg.ProjectID, msg.RequestID, msg.Data)
				}
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
	Action    string      `json:"action"` // "subscribe", "unsubscribe", or "ui_response"
	ProjectID string      `json:"projectId"`
	RequestID string      `json:"requestId,omitempty"` // for ui_response
	Data      interface{} `json:"data,omitempty"`      // for ui_response
}

// UIResponseHandler is called when a UI response is received
type UIResponseHandler func(projectID, requestID string, data interface{})
