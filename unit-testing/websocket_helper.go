package unit_testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a WebSocket message for testing
type WebSocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// WebSocketTestHelper provides utilities for WebSocket testing
type WebSocketTestHelper struct {
	Server      *httptest.Server
	Upgrader    websocket.Upgrader
	Connections map[string]*TestWebSocketConnection
	mutex       sync.RWMutex
}

// TestWebSocketConnection represents a test WebSocket connection
type TestWebSocketConnection struct {
	Conn           *websocket.Conn
	Messages       []WebSocketMessage
	SentMessages   []interface{}
	UserID         int
	Connected      bool
	LastActivity   time.Time
	MessageHandler func(WebSocketMessage)
	mutex          sync.RWMutex
}

// NewWebSocketTestHelper creates a new WebSocket test helper
func NewWebSocketTestHelper() *WebSocketTestHelper {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}

	return &WebSocketTestHelper{
		Upgrader:    upgrader,
		Connections: make(map[string]*TestWebSocketConnection),
	}
}

// StartServer starts the test WebSocket server
func (h *WebSocketTestHelper) StartServer(handler http.HandlerFunc) {
	h.Server = httptest.NewServer(handler)
}

// Close closes the test server and all connections
func (h *WebSocketTestHelper) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, conn := range h.Connections {
		conn.Close()
	}

	if h.Server != nil {
		h.Server.Close()
	}
}

// ConnectUser creates a WebSocket connection for a user
func (h *WebSocketTestHelper) ConnectUser(userID int, sessionToken string) (*TestWebSocketConnection, error) {
	if h.Server == nil {
		return nil, fmt.Errorf("server not started")
	}

	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(h.Server.URL, "http://", "ws://", 1) + "/ws"

	// Add query parameters for authentication
	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse WebSocket URL: %v", err)
	}

	q := u.Query()
	q.Set("user_id", fmt.Sprintf("%d", userID))
	q.Set("session_token", sessionToken)
	u.RawQuery = q.Encode()

	// Create WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %v", err)
	}

	testConn := &TestWebSocketConnection{
		Conn:         conn,
		Messages:     make([]WebSocketMessage, 0),
		SentMessages: make([]interface{}, 0),
		UserID:       userID,
		Connected:    true,
		LastActivity: time.Now(),
	}

	// Start message reading goroutine
	go testConn.readMessages()

	h.mutex.Lock()
	h.Connections[fmt.Sprintf("user_%d", userID)] = testConn
	h.mutex.Unlock()

	return testConn, nil
}

// GetConnection returns a connection for a user
func (h *WebSocketTestHelper) GetConnection(userID int) *TestWebSocketConnection {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.Connections[fmt.Sprintf("user_%d", userID)]
}

// DisconnectUser disconnects a user's WebSocket connection
func (h *WebSocketTestHelper) DisconnectUser(userID int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	key := fmt.Sprintf("user_%d", userID)
	if conn, exists := h.Connections[key]; exists {
		conn.Close()
		delete(h.Connections, key)
	}
}

// BroadcastMessage broadcasts a message to all connected users
func (h *WebSocketTestHelper) BroadcastMessage(message interface{}) error {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, conn := range h.Connections {
		if err := conn.SendMessage(message); err != nil {
			return fmt.Errorf("failed to send message to user %d: %v", conn.UserID, err)
		}
	}

	return nil
}

// SendMessage sends a message from one user to another
func (h *WebSocketTestHelper) SendMessage(fromUserID, toUserID int, message interface{}) error {
	fromConn := h.GetConnection(fromUserID)
	if fromConn == nil {
		return fmt.Errorf("user %d not connected", fromUserID)
	}

	return fromConn.SendMessage(message)
}

// WaitForMessage waits for a message to be received by a user
func (h *WebSocketTestHelper) WaitForMessage(userID int, timeout time.Duration) (WebSocketMessage, error) {
	conn := h.GetConnection(userID)
	if conn == nil {
		return WebSocketMessage{}, fmt.Errorf("user %d not connected", userID)
	}

	return conn.WaitForMessage(timeout)
}

// AssertMessageReceived checks if a user received a specific message
func (h *WebSocketTestHelper) AssertMessageReceived(t *testing.T, userID int, expectedType string, timeout time.Duration) WebSocketMessage {
	message, err := h.WaitForMessage(userID, timeout)
	if err != nil {
		t.Fatalf("Failed to receive message for user %d: %v", userID, err)
	}

	if message.Type != expectedType {
		t.Fatalf("Expected message type %s, got %s", expectedType, message.Type)
	}

	return message
}

// TestWebSocketConnection methods

// SendMessage sends a message through the WebSocket connection
func (c *TestWebSocketConnection) SendMessage(message interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.Connected {
		return fmt.Errorf("connection is closed")
	}

	if err := c.Conn.WriteJSON(message); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	c.SentMessages = append(c.SentMessages, message)
	c.LastActivity = time.Now()

	return nil
}

// WaitForMessage waits for a message to be received
func (c *TestWebSocketConnection) WaitForMessage(timeout time.Duration) (WebSocketMessage, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		c.mutex.RLock()
		if len(c.Messages) > 0 {
			message := c.Messages[0]
			c.Messages = c.Messages[1:]
			c.mutex.RUnlock()
			return message, nil
		}
		c.mutex.RUnlock()

		time.Sleep(10 * time.Millisecond)
	}

	return WebSocketMessage{}, fmt.Errorf("timeout waiting for message")
}

// GetReceivedMessages returns all received messages
func (c *TestWebSocketConnection) GetReceivedMessages() []WebSocketMessage {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	messages := make([]WebSocketMessage, len(c.Messages))
	copy(messages, c.Messages)
	return messages
}

// GetSentMessages returns all sent messages
func (c *TestWebSocketConnection) GetSentMessages() []interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	messages := make([]interface{}, len(c.SentMessages))
	copy(messages, c.SentMessages)
	return messages
}

// ClearMessages clears all received and sent messages
func (c *TestWebSocketConnection) ClearMessages() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Messages = make([]WebSocketMessage, 0)
	c.SentMessages = make([]interface{}, 0)
}

// Close closes the WebSocket connection
func (c *TestWebSocketConnection) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Connected {
		c.Connected = false
		if c.Conn != nil {
			c.Conn.Close()
		}
	}
}

// IsConnected returns whether the connection is active
func (c *TestWebSocketConnection) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Connected
}

// SetMessageHandler sets a custom message handler
func (c *TestWebSocketConnection) SetMessageHandler(handler func(WebSocketMessage)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.MessageHandler = handler
}

// readMessages reads messages from the WebSocket connection
func (c *TestWebSocketConnection) readMessages() {
	defer c.Close()

	for c.Connected {
		var rawMessage json.RawMessage
		err := c.Conn.ReadJSON(&rawMessage)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log unexpected close error
			}
			break
		}

		// Parse the message to determine its type
		var baseMessage struct {
			Type string `json:"type"`
		}

		if err := json.Unmarshal(rawMessage, &baseMessage); err != nil {
			continue // Skip malformed messages
		}

		message := WebSocketMessage{
			Type: baseMessage.Type,
			Data: rawMessage,
		}

		c.mutex.Lock()
		c.Messages = append(c.Messages, message)
		c.LastActivity = time.Now()

		// Call custom message handler if set
		if c.MessageHandler != nil {
			go c.MessageHandler(message)
		}
		c.mutex.Unlock()
	}
}

// CreateTestWebSocketMessage creates a test WebSocket message
func CreateTestWebSocketMessage(messageType string, data interface{}) WebSocketMessage {
	jsonData, _ := json.Marshal(data)
	return WebSocketMessage{
		Type: messageType,
		Data: jsonData,
	}
}

// AssertWebSocketMessage checks if a WebSocket message matches expected values
func AssertWebSocketMessage(t *testing.T, message WebSocketMessage, expectedType string, expectedData interface{}) {
	if message.Type != expectedType {
		t.Fatalf("Expected message type %s, got %s", expectedType, message.Type)
	}

	if expectedData != nil {
		var actualData interface{}
		if err := json.Unmarshal(message.Data, &actualData); err != nil {
			t.Fatalf("Failed to unmarshal message data: %v", err)
		}

		expectedJSON, _ := json.Marshal(expectedData)
		actualJSON, _ := json.Marshal(actualData)

		if string(expectedJSON) != string(actualJSON) {
			t.Fatalf("Expected message data %s, got %s", string(expectedJSON), string(actualJSON))
		}
	}
}

// SimulateTypingIndicator simulates a typing indicator message
func (h *WebSocketTestHelper) SimulateTypingIndicator(userID, conversationID int, isTyping bool) error {
	message := map[string]interface{}{
		"type":            "typing_indicator",
		"user_id":         userID,
		"conversation_id": conversationID,
		"is_typing":       isTyping,
		"timestamp":       time.Now().Unix(),
	}

	return h.BroadcastMessage(message)
}

// SimulateOnlineStatusChange simulates an online status change
func (h *WebSocketTestHelper) SimulateOnlineStatusChange(userID int, status string) error {
	message := map[string]interface{}{
		"type":      "online_status",
		"user_id":   userID,
		"status":    status,
		"timestamp": time.Now().Unix(),
	}

	return h.BroadcastMessage(message)
}
