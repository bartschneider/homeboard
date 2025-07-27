package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/bartosz/homeboard/internal/config"
)

func TestWebSocketManager(t *testing.T) {
	manager := NewWebSocketManager()

	t.Run("ManagerInitialization", func(t *testing.T) {
		if manager.clients == nil {
			t.Error("Expected clients map to be initialized")
		}

		if manager.broadcast == nil {
			t.Error("Expected broadcast channel to be initialized")
		}

		if manager.register == nil {
			t.Error("Expected register channel to be initialized")
		}

		if manager.unregister == nil {
			t.Error("Expected unregister channel to be initialized")
		}

		if manager.upgrader.ReadBufferSize != 1024 {
			t.Errorf("Expected read buffer size 1024, got %d", manager.upgrader.ReadBufferSize)
		}

		if manager.upgrader.WriteBufferSize != 1024 {
			t.Errorf("Expected write buffer size 1024, got %d", manager.upgrader.WriteBufferSize)
		}
	})

	t.Run("ClientConnections", func(t *testing.T) {
		// Start manager in goroutine
		go manager.Run()

		// Test client count
		initialCount := manager.GetConnectedClients()
		if initialCount != 0 {
			t.Errorf("Expected 0 connected clients initially, got %d", initialCount)
		}

		// Create mock client
		client := &Client{
			conn:     nil, // Would be a real connection in practice
			send:     make(chan AdminMessage, 256),
			manager:  manager,
			id:       "test_client_1",
			connTime: time.Now(),
		}

		// Register client
		manager.register <- client

		// Give some time for registration
		time.Sleep(10 * time.Millisecond)

		// Check client count
		count := manager.GetConnectedClients()
		if count != 1 {
			t.Errorf("Expected 1 connected client, got %d", count)
		}

		// Unregister client
		manager.unregister <- client

		// Give some time for unregistration
		time.Sleep(10 * time.Millisecond)

		// Check client count
		count = manager.GetConnectedClients()
		if count != 0 {
			t.Errorf("Expected 0 connected clients after unregistration, got %d", count)
		}
	})

	t.Run("MessageBroadcasting", func(t *testing.T) {
		manager := NewWebSocketManager()
		go manager.Run()

		// Create multiple mock clients
		clients := make([]*Client, 3)
		for i := 0; i < 3; i++ {
			client := &Client{
				conn:     nil,
				send:     make(chan AdminMessage, 256),
				manager:  manager,
				id:       generateClientID(),
				connTime: time.Now(),
			}
			clients[i] = client
			manager.register <- client
		}

		// Give time for registration
		time.Sleep(10 * time.Millisecond)

		// Broadcast message
		message := AdminMessage{
			Type: "test_message",
			Payload: map[string]interface{}{
				"content": "test broadcast",
			},
			Timestamp: time.Now(),
		}

		manager.broadcast <- message

		// Give time for broadcast
		time.Sleep(10 * time.Millisecond)

		// Check that all clients received the message
		for i, client := range clients {
			select {
			case receivedMessage := <-client.send:
				if receivedMessage.Type != "test_message" {
					t.Errorf("Client %d expected message type 'test_message', got '%s'", i, receivedMessage.Type)
				}
			default:
				t.Errorf("Client %d did not receive broadcast message", i)
			}
		}
	})

	t.Run("BroadcastMethods", func(t *testing.T) {
		manager := NewWebSocketManager()
		go manager.Run()

		// Create mock client to receive messages
		client := &Client{
			conn:     nil,
			send:     make(chan AdminMessage, 256),
			manager:  manager,
			id:       generateClientID(),
			connTime: time.Now(),
		}
		manager.register <- client
		time.Sleep(10 * time.Millisecond)

		// Test config update broadcast
		testConfig := &config.Config{
			Title: "Test Config",
		}
		manager.BroadcastConfigUpdate(testConfig)

		// Verify message received
		select {
		case message := <-client.send:
			if message.Type != WSTypeConfigUpdate {
				t.Errorf("Expected config update message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive config update message")
		}

		// Test widget status broadcast
		widgetUpdate := WidgetStatusUpdate{
			WidgetName: "test_widget",
			Status:     "active",
		}
		manager.BroadcastWidgetStatus(widgetUpdate)

		select {
		case message := <-client.send:
			if message.Type != WSTypeWidgetStatus {
				t.Errorf("Expected widget status message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive widget status message")
		}

		// Test system metrics broadcast
		metricsUpdate := SystemMetricsUpdate{
			Timestamp: time.Now(),
			Metrics: SystemMetrics{
				CPUUsage:    50.0,
				MemoryUsage: 60.0,
			},
		}
		manager.BroadcastSystemMetrics(metricsUpdate)

		select {
		case message := <-client.send:
			if message.Type != WSTypeSystemMetrics {
				t.Errorf("Expected system metrics message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive system metrics message")
		}

		// Test log entry broadcast
		logEntry := LogEntry{
			Level:   "info",
			Message: "Test log",
		}
		manager.BroadcastLogEntry(logEntry)

		select {
		case message := <-client.send:
			if message.Type != WSTypeLogEntry {
				t.Errorf("Expected log entry message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive log entry message")
		}

		// Test notification broadcast
		notification := NotificationMessage{
			Type:    "success",
			Title:   "Test",
			Message: "Test notification",
		}
		manager.BroadcastNotification(notification)

		select {
		case message := <-client.send:
			if message.Type != WSTypeNotification {
				t.Errorf("Expected notification message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive notification message")
		}

		// Test error broadcast
		manager.BroadcastError("Test error", map[string]interface{}{"code": 500})

		select {
		case message := <-client.send:
			if message.Type != WSTypeError {
				t.Errorf("Expected error message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive error message")
		}
	})

	t.Run("ClientInfo", func(t *testing.T) {
		manager := NewWebSocketManager()
		go manager.Run()

		// Create client with known ID
		clientID := "test_client_info"
		client := &Client{
			conn:     nil,
			send:     make(chan AdminMessage, 256),
			manager:  manager,
			id:       clientID,
			connTime: time.Now(),
		}
		manager.register <- client
		time.Sleep(10 * time.Millisecond)

		// Get client info
		clientsInfo := manager.GetClientInfo()
		if len(clientsInfo) != 1 {
			t.Errorf("Expected 1 client info, got %d", len(clientsInfo))
		}

		if clientsInfo[0]["id"] != clientID {
			t.Errorf("Expected client ID '%s', got '%s'", clientID, clientsInfo[0]["id"])
		}

		if clientsInfo[0]["connected_at"] == nil {
			t.Error("Expected connected_at to be set")
		}
	})
}

func TestWebSocketConnection(t *testing.T) {
	// Create test server
	manager := NewWebSocketManager()
	go manager.Run()

	server := httptest.NewServer(http.HandlerFunc(manager.HandleConnection))
	defer server.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	t.Run("ConnectionUpgrade", func(t *testing.T) {
		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Give time for connection to be registered
		time.Sleep(50 * time.Millisecond)

		// Check that client was registered
		clientCount := manager.GetConnectedClients()
		if clientCount != 1 {
			t.Errorf("Expected 1 connected client, got %d", clientCount)
		}

		// Read welcome message
		var message AdminMessage
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatalf("Failed to read welcome message: %v", err)
		}

		if message.Type != "connection_established" {
			t.Errorf("Expected connection_established message, got '%s'", message.Type)
		}

		// Close connection
		conn.Close()

		// Give time for disconnection
		time.Sleep(50 * time.Millisecond)

		// Check that client was unregistered
		clientCount = manager.GetConnectedClients()
		if clientCount != 0 {
			t.Errorf("Expected 0 connected clients after disconnect, got %d", clientCount)
		}
	})

	t.Run("PingPongHandling", func(t *testing.T) {
		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Skip welcome message
		var welcomeMessage AdminMessage
		conn.ReadJSON(&welcomeMessage)

		// Send ping message
		pingMessage := AdminMessage{
			Type: "ping",
			Payload: map[string]interface{}{
				"timestamp": time.Now(),
			},
			Timestamp: time.Now(),
		}

		if err := conn.WriteJSON(pingMessage); err != nil {
			t.Fatalf("Failed to send ping message: %v", err)
		}

		// Read pong response
		var pongMessage AdminMessage
		if err := conn.ReadJSON(&pongMessage); err != nil {
			t.Fatalf("Failed to read pong message: %v", err)
		}

		if pongMessage.Type != "pong" {
			t.Errorf("Expected pong message, got '%s'", pongMessage.Type)
		}
	})

	t.Run("SubscriptionHandling", func(t *testing.T) {
		// Connect to WebSocket
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Skip welcome message
		var welcomeMessage AdminMessage
		conn.ReadJSON(&welcomeMessage)

		// Send subscription message
		subscribeMessage := AdminMessage{
			Type: "subscribe",
			Payload: map[string]interface{}{
				"topics": []interface{}{"config_updates", "widget_status"},
			},
			Timestamp: time.Now(),
		}

		if err := conn.WriteJSON(subscribeMessage); err != nil {
			t.Fatalf("Failed to send subscribe message: %v", err)
		}

		// Read subscription confirmation
		var confirmMessage AdminMessage
		if err := conn.ReadJSON(&confirmMessage); err != nil {
			t.Fatalf("Failed to read subscription confirmation: %v", err)
		}

		if confirmMessage.Type != "subscription_confirmed" {
			t.Errorf("Expected subscription_confirmed message, got '%s'", confirmMessage.Type)
		}

		// Send unsubscription message
		unsubscribeMessage := AdminMessage{
			Type: "unsubscribe",
			Payload: map[string]interface{}{
				"topics": []interface{}{"config_updates"},
			},
			Timestamp: time.Now(),
		}

		if err := conn.WriteJSON(unsubscribeMessage); err != nil {
			t.Fatalf("Failed to send unsubscribe message: %v", err)
		}

		// Read unsubscription confirmation
		var unconfirmMessage AdminMessage
		if err := conn.ReadJSON(&unconfirmMessage); err != nil {
			t.Fatalf("Failed to read unsubscription confirmation: %v", err)
		}

		if unconfirmMessage.Type != "unsubscription_confirmed" {
			t.Errorf("Expected unsubscription_confirmed message, got '%s'", unconfirmMessage.Type)
		}
	})
}

func TestMessageBroadcaster(t *testing.T) {
	manager := NewWebSocketManager()
	go manager.Run()

	broadcaster := NewMessageBroadcaster(manager)

	// Create mock client to receive messages
	client := &Client{
		conn:     nil,
		send:     make(chan AdminMessage, 256),
		manager:  manager,
		id:       generateClientID(),
		connTime: time.Now(),
	}
	manager.register <- client
	time.Sleep(10 * time.Millisecond)

	t.Run("NotifyConfigChange", func(t *testing.T) {
		testConfig := &config.Config{
			Title: "Test Config",
		}

		broadcaster.NotifyConfigChange(testConfig, "updated")

		// Should receive notification message
		select {
		case message := <-client.send:
			if message.Type != WSTypeNotification {
				t.Errorf("Expected notification message type, got '%s'", message.Type)
			}

			payload, ok := message.Payload.(NotificationMessage)
			if !ok {
				t.Error("Expected NotificationMessage payload")
			} else {
				if payload.Type != "info" {
					t.Errorf("Expected notification type 'info', got '%s'", payload.Type)
				}
				if payload.Title != "Configuration Updated" {
					t.Errorf("Expected title 'Configuration Updated', got '%s'", payload.Title)
				}
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive notification message")
		}

		// Should also receive config update message
		select {
		case message := <-client.send:
			if message.Type != WSTypeConfigUpdate {
				t.Errorf("Expected config update message type, got '%s'", message.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive config update message")
		}
	})

	t.Run("NotifyWidgetChange", func(t *testing.T) {
		broadcaster.NotifyWidgetChange("test_widget", "updated", "Successfully updated")

		select {
		case message := <-client.send:
			if message.Type != WSTypeNotification {
				t.Errorf("Expected notification message type, got '%s'", message.Type)
			}

			payload, ok := message.Payload.(NotificationMessage)
			if !ok {
				t.Error("Expected NotificationMessage payload")
			} else {
				if payload.Type != "info" {
					t.Errorf("Expected notification type 'info', got '%s'", payload.Type)
				}
				if payload.Title != "Widget updated" {
					t.Errorf("Expected title 'Widget updated', got '%s'", payload.Title)
				}
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive widget change notification")
		}
	})

	t.Run("NotifyError", func(t *testing.T) {
		broadcaster.NotifyError("Test Error", "Something went wrong", false)

		select {
		case message := <-client.send:
			if message.Type != WSTypeNotification {
				t.Errorf("Expected notification message type, got '%s'", message.Type)
			}

			payload, ok := message.Payload.(NotificationMessage)
			if !ok {
				t.Error("Expected NotificationMessage payload")
			} else {
				if payload.Type != "error" {
					t.Errorf("Expected notification type 'error', got '%s'", payload.Type)
				}
				if payload.Title != "Test Error" {
					t.Errorf("Expected title 'Test Error', got '%s'", payload.Title)
				}
				if payload.Duration != 10 {
					t.Errorf("Expected duration 10, got %d", payload.Duration)
				}
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive error notification")
		}
	})

	t.Run("NotifySuccess", func(t *testing.T) {
		broadcaster.NotifySuccess("Success", "Operation completed")

		select {
		case message := <-client.send:
			if message.Type != WSTypeNotification {
				t.Errorf("Expected notification message type, got '%s'", message.Type)
			}

			payload, ok := message.Payload.(NotificationMessage)
			if !ok {
				t.Error("Expected NotificationMessage payload")
			} else {
				if payload.Type != "success" {
					t.Errorf("Expected notification type 'success', got '%s'", payload.Type)
				}
				if payload.Title != "Success" {
					t.Errorf("Expected title 'Success', got '%s'", payload.Title)
				}
				if payload.Duration != 3 {
					t.Errorf("Expected duration 3, got %d", payload.Duration)
				}
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive success notification")
		}
	})

	t.Run("PersistentErrorNotification", func(t *testing.T) {
		broadcaster.NotifyError("Persistent Error", "This error persists", true)

		select {
		case message := <-client.send:
			payload, ok := message.Payload.(NotificationMessage)
			if !ok {
				t.Error("Expected NotificationMessage payload")
			} else {
				if payload.Duration != 0 {
					t.Errorf("Expected persistent notification (duration 0), got %d", payload.Duration)
				}
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive persistent error notification")
		}
	})
}

func TestClientIDGeneration(t *testing.T) {
	t.Run("UniqueIDs", func(t *testing.T) {
		ids := make(map[string]bool)

		// Generate multiple IDs
		for i := 0; i < 100; i++ {
			id := generateClientID()
			if ids[id] {
				t.Errorf("Generated duplicate client ID: %s", id)
			}
			ids[id] = true

			if !strings.HasPrefix(id, "client_") {
				t.Errorf("Client ID should start with 'client_', got: %s", id)
			}
		}
	})
}

func TestNotificationIDGeneration(t *testing.T) {
	t.Run("UniqueNotificationIDs", func(t *testing.T) {
		ids := make(map[string]bool)

		// Generate multiple IDs
		for i := 0; i < 100; i++ {
			id := generateNotificationID()
			if ids[id] {
				t.Errorf("Generated duplicate notification ID: %s", id)
			}
			ids[id] = true

			if !strings.HasPrefix(id, "notif_") {
				t.Errorf("Notification ID should start with 'notif_', got: %s", id)
			}
		}
	})
}

func TestConcurrentWebSocketOperations(t *testing.T) {
	manager := NewWebSocketManager()
	go manager.Run()

	t.Run("ConcurrentClientRegistration", func(t *testing.T) {
		numClients := 50
		clients := make([]*Client, numClients)

		// Register clients concurrently
		for i := 0; i < numClients; i++ {
			go func(index int) {
				client := &Client{
					conn:     nil,
					send:     make(chan AdminMessage, 256),
					manager:  manager,
					id:       generateClientID(),
					connTime: time.Now(),
				}
				clients[index] = client
				manager.register <- client
			}(i)
		}

		// Wait for all registrations
		time.Sleep(100 * time.Millisecond)

		// Check client count
		count := manager.GetConnectedClients()
		if count != numClients {
			t.Errorf("Expected %d connected clients, got %d", numClients, count)
		}

		// Unregister all clients
		for i := 0; i < numClients; i++ {
			if clients[i] != nil {
				manager.unregister <- clients[i]
			}
		}

		// Wait for all unregistrations
		time.Sleep(100 * time.Millisecond)

		// Check client count
		count = manager.GetConnectedClients()
		if count != 0 {
			t.Errorf("Expected 0 connected clients after unregistration, got %d", count)
		}
	})

	t.Run("ConcurrentBroadcasting", func(t *testing.T) {
		// Register some clients
		numClients := 10
		clients := make([]*Client, numClients)

		for i := 0; i < numClients; i++ {
			client := &Client{
				conn:     nil,
				send:     make(chan AdminMessage, 256),
				manager:  manager,
				id:       generateClientID(),
				connTime: time.Now(),
			}
			clients[i] = client
			manager.register <- client
		}

		time.Sleep(50 * time.Millisecond)

		// Broadcast messages concurrently
		numMessages := 20
		for i := 0; i < numMessages; i++ {
			go func(index int) {
				message := AdminMessage{
					Type: "test_message",
					Payload: map[string]interface{}{
						"index": index,
					},
					Timestamp: time.Now(),
				}
				manager.broadcast <- message
			}(i)
		}

		// Wait for broadcasts
		time.Sleep(100 * time.Millisecond)

		// Check that clients received messages
		for i, client := range clients {
			if client == nil {
				continue
			}

			receivedCount := 0
			for {
				select {
				case <-client.send:
					receivedCount++
				default:
					goto checkCount
				}
			}

		checkCount:
			if receivedCount != numMessages {
				t.Errorf("Client %d expected %d messages, got %d", i, numMessages, receivedCount)
			}
		}
	})
}
