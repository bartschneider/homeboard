package admin

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/bartosz/homeboard/internal/config"
)

// WebSocketManager manages WebSocket connections for real-time updates
type WebSocketManager struct {
	clients    map[*Client]bool
	broadcast  chan AdminMessage
	register   chan *Client
	unregister chan *Client
	upgrader   websocket.Upgrader
	mutex      sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	conn     *websocket.Conn
	send     chan AdminMessage
	manager  *WebSocketManager
	id       string
	connTime time.Time
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan AdminMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
	}
}

// Run starts the WebSocket manager hub
func (wsm *WebSocketManager) Run() {
	for {
		select {
		case client := <-wsm.register:
			wsm.registerClient(client)

		case client := <-wsm.unregister:
			wsm.unregisterClient(client)

		case message := <-wsm.broadcast:
			wsm.broadcastMessage(message)
		}
	}
}

// HandleConnection handles new WebSocket connections
func (wsm *WebSocketManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := wsm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client
	client := &Client{
		conn:     conn,
		send:     make(chan AdminMessage, 256),
		manager:  wsm,
		id:       generateClientID(),
		connTime: time.Now(),
	}

	// Register client
	wsm.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// registerClient adds a new client to the manager
func (wsm *WebSocketManager) registerClient(client *Client) {
	wsm.mutex.Lock()
	wsm.clients[client] = true
	wsm.mutex.Unlock()

	log.Printf("WebSocket client connected: %s", client.id)

	// Send welcome message with current status
	welcomeMessage := AdminMessage{
		Type: "connection_established",
		Payload: map[string]interface{}{
			"client_id":   client.id,
			"server_time": time.Now(),
			"message":     "Connected to admin panel",
		},
		Timestamp: time.Now(),
	}

	select {
	case client.send <- welcomeMessage:
	default:
		close(client.send)
		delete(wsm.clients, client)
	}
}

// unregisterClient removes a client from the manager
func (wsm *WebSocketManager) unregisterClient(client *Client) {
	wsm.mutex.Lock()
	if _, ok := wsm.clients[client]; ok {
		delete(wsm.clients, client)
		close(client.send)
		wsm.mutex.Unlock()
		log.Printf("WebSocket client disconnected: %s", client.id)
	} else {
		wsm.mutex.Unlock()
	}
}

// broadcastMessage sends a message to all connected clients
func (wsm *WebSocketManager) broadcastMessage(message AdminMessage) {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()

	for client := range wsm.clients {
		select {
		case client.send <- message:
		default:
			// Client's send channel is full, disconnect client
			close(client.send)
			delete(wsm.clients, client)
		}
	}
}

// Broadcast methods for different message types

// BroadcastConfigUpdate sends configuration update to all clients
func (wsm *WebSocketManager) BroadcastConfigUpdate(config *config.Config) {
	message := AdminMessage{
		Type:      WSTypeConfigUpdate,
		Payload:   config,
		Timestamp: time.Now(),
	}
	wsm.broadcast <- message
}

// BroadcastWidgetStatus sends widget status update to all clients
func (wsm *WebSocketManager) BroadcastWidgetStatus(update WidgetStatusUpdate) {
	message := AdminMessage{
		Type:      WSTypeWidgetStatus,
		Payload:   update,
		Timestamp: time.Now(),
	}
	wsm.broadcast <- message
}

// BroadcastSystemMetrics sends system metrics to all clients
func (wsm *WebSocketManager) BroadcastSystemMetrics(metrics SystemMetricsUpdate) {
	message := AdminMessage{
		Type:      WSTypeSystemMetrics,
		Payload:   metrics,
		Timestamp: time.Now(),
	}
	wsm.broadcast <- message
}

// BroadcastLogEntry sends log entry to all clients
func (wsm *WebSocketManager) BroadcastLogEntry(logEntry LogEntry) {
	message := AdminMessage{
		Type:      WSTypeLogEntry,
		Payload:   logEntry,
		Timestamp: time.Now(),
	}
	wsm.broadcast <- message
}

// BroadcastNotification sends notification to all clients
func (wsm *WebSocketManager) BroadcastNotification(notification NotificationMessage) {
	message := AdminMessage{
		Type:      WSTypeNotification,
		Payload:   notification,
		Timestamp: time.Now(),
	}
	wsm.broadcast <- message
}

// BroadcastError sends error message to all clients
func (wsm *WebSocketManager) BroadcastError(errorMsg string, details map[string]interface{}) {
	payload := map[string]interface{}{
		"message": errorMsg,
		"details": details,
	}

	message := AdminMessage{
		Type:      WSTypeError,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	wsm.broadcast <- message
}

// GetConnectedClients returns the number of connected clients
func (wsm *WebSocketManager) GetConnectedClients() int {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()
	return len(wsm.clients)
}

// GetClientInfo returns information about connected clients
func (wsm *WebSocketManager) GetClientInfo() []map[string]interface{} {
	wsm.mutex.RLock()
	defer wsm.mutex.RUnlock()

	clients := make([]map[string]interface{}, 0, len(wsm.clients))
	for client := range wsm.clients {
		clients = append(clients, map[string]interface{}{
			"id":           client.id,
			"connected_at": client.connTime,
			"remote_addr":  client.conn.RemoteAddr().String(),
		})
	}
	return clients
}

// Client methods

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	// Set read deadline and pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message AdminMessage
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages from client
		c.handleIncomingMessage(message)
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncomingMessage processes messages received from clients
func (c *Client) handleIncomingMessage(message AdminMessage) {
	switch message.Type {
	case "ping":
		// Respond to ping with pong
		pongMessage := AdminMessage{
			Type: "pong",
			Payload: map[string]interface{}{
				"timestamp": time.Now(),
			},
			Timestamp: time.Now(),
		}
		select {
		case c.send <- pongMessage:
		default:
			// Channel full, ignore
		}

	case "subscribe":
		// Handle subscription requests
		c.handleSubscription(message)

	case "unsubscribe":
		// Handle unsubscription requests
		c.handleUnsubscription(message)

	default:
		log.Printf("Unknown message type from client %s: %s", c.id, message.Type)
	}
}

// handleSubscription handles client subscription requests
func (c *Client) handleSubscription(message AdminMessage) {
	// Parse subscription payload
	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		return
	}

	topics, ok := payload["topics"].([]interface{})
	if !ok {
		return
	}

	// Send confirmation
	response := AdminMessage{
		Type: "subscription_confirmed",
		Payload: map[string]interface{}{
			"topics":    topics,
			"client_id": c.id,
		},
		Timestamp: time.Now(),
	}

	select {
	case c.send <- response:
	default:
		// Channel full, ignore
	}

	log.Printf("Client %s subscribed to topics: %v", c.id, topics)
}

// handleUnsubscription handles client unsubscription requests
func (c *Client) handleUnsubscription(message AdminMessage) {
	// Parse unsubscription payload
	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		return
	}

	topics, ok := payload["topics"].([]interface{})
	if !ok {
		return
	}

	// Send confirmation
	response := AdminMessage{
		Type: "unsubscription_confirmed",
		Payload: map[string]interface{}{
			"topics":    topics,
			"client_id": c.id,
		},
		Timestamp: time.Now(),
	}

	select {
	case c.send <- response:
	default:
		// Channel full, ignore
	}

	log.Printf("Client %s unsubscribed from topics: %v", c.id, topics)
}

// Helper functions

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// MessageBroadcaster provides a high-level interface for broadcasting messages
type MessageBroadcaster struct {
	wsManager *WebSocketManager
}

// NewMessageBroadcaster creates a new message broadcaster
func NewMessageBroadcaster(wsManager *WebSocketManager) *MessageBroadcaster {
	return &MessageBroadcaster{
		wsManager: wsManager,
	}
}

// NotifyConfigChange broadcasts configuration changes
func (mb *MessageBroadcaster) NotifyConfigChange(config *config.Config, changeType string) {
	notification := NotificationMessage{
		ID:       generateNotificationID(),
		Type:     "info",
		Title:    "Configuration Updated",
		Message:  fmt.Sprintf("Configuration %s successfully", changeType),
		Duration: 5, // 5 seconds
	}

	mb.wsManager.BroadcastNotification(notification)
	mb.wsManager.BroadcastConfigUpdate(config)
}

// NotifyWidgetChange broadcasts widget status changes
func (mb *MessageBroadcaster) NotifyWidgetChange(widgetName, status, message string) {
	notification := NotificationMessage{
		ID:       generateNotificationID(),
		Type:     "info",
		Title:    fmt.Sprintf("Widget %s", status),
		Message:  fmt.Sprintf("%s: %s", widgetName, message),
		Duration: 3,
	}

	mb.wsManager.BroadcastNotification(notification)
}

// NotifyError broadcasts error notifications
func (mb *MessageBroadcaster) NotifyError(title, message string, persistent bool) {
	duration := 10
	if persistent {
		duration = 0 // Persistent notification
	}

	notification := NotificationMessage{
		ID:       generateNotificationID(),
		Type:     "error",
		Title:    title,
		Message:  message,
		Duration: duration,
	}

	mb.wsManager.BroadcastNotification(notification)
}

// NotifySuccess broadcasts success notifications
func (mb *MessageBroadcaster) NotifySuccess(title, message string) {
	notification := NotificationMessage{
		ID:       generateNotificationID(),
		Type:     "success",
		Title:    title,
		Message:  message,
		Duration: 3,
	}

	mb.wsManager.BroadcastNotification(notification)
}

// generateNotificationID generates a unique notification ID
func generateNotificationID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}
