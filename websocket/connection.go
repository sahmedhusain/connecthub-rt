package websocket

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, implement proper origin checking
		return true
	},
}

type Manager struct {
	hub    *Hub
	logger *Logger
}

func NewManager() *Manager {
	hub := NewHub()
	go hub.Run()

	return &Manager{
		hub:    hub,
		logger: NewLogger(false),
	}
}

func NewManagerWithDebug(debug bool) *Manager {
	hub := NewHubWithLogging(debug)
	go hub.Run()

	return &Manager{
		hub:    hub,
		logger: NewLogger(debug),
	}
}

func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Extract and validate user ID
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		m.logger.Error("Missing user_id parameter from %s", r.RemoteAddr)
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		m.logger.Error("Invalid user ID '%s' from %s", userIDStr, r.RemoteAddr)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Authenticate the WebSocket connection
	if !m.authenticateWebSocketConnection(r, userID) {
		m.logger.Error("Authentication failed for WebSocket connection from user %d at %s", userID, r.RemoteAddr)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.logger.Error("Error upgrading connection for user %d: %v", userID, err)
		return
	}

	// Set connection properties
	conn.SetReadLimit(maxMessageSize)

	m.logger.Debug("New WebSocket connection established for user %d from %s",
		userID, r.RemoteAddr)

	// Create and register new client
	client := NewClient(m.hub, conn, userID)

	// Use select to prevent blocking if hub is busy
	select {
	case client.hub.register <- client:
		m.logger.Info("Client registered, starting pumps for user %d", userID)
		go client.WritePump()
		go client.ReadPump()
	default:
		m.logger.Error("Failed to register client, hub busy. Closing connection for user %d", userID)
		conn.Close()
	}
}

func (m *Manager) SendToUser(userID int, message Message) bool {
	if userID <= 0 {
		m.logger.Error("Invalid user ID for message sending: %d", userID)
		return false
	}

	success := m.hub.SendToUser(userID, message)
	if !success {
		m.logger.Debug("Failed to send message to user %d - user might be offline", userID)
	}
	return success
}

func (m *Manager) BroadcastMessage(message Message) {
	if message.Type == "" {
		m.logger.Error("Attempted to broadcast message with empty type")
		return
	}

	m.logger.Debug("Broadcasting message type: %s from manager", message.Type)

	// Use select to prevent blocking if hub is busy
	select {
	case m.hub.broadcast <- message:
		m.logger.Debug("Message queued for broadcast")
	default:
		m.logger.Error("Failed to broadcast message, hub busy")
	}
}

func (m *Manager) IsUserOnline(userID int) bool {
	return m.hub.IsUserOnline(userID)
}

func (m *Manager) GetOnlineUsers() []int {
	return m.hub.GetOnlineUsers()
}

func (m *Manager) SetDebugMode(debug bool) {
	m.logger.debug = debug
	m.hub.SetDebugMode(debug)
	m.logger.Info("Debug mode set to: %v", debug)
}

func (m *Manager) GetStats() map[string]interface{} {
	return m.hub.GetStats()
}

func (m *Manager) SendReadStatusUpdate(conversationID int, readerID int) {
	m.hub.SendReadStatusUpdate(conversationID, readerID)
}

// authenticateWebSocketConnection validates the user's session for WebSocket connections
func (m *Manager) authenticateWebSocketConnection(r *http.Request, userID int) bool {
	// Get session cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		m.logger.Error("No session cookie found for WebSocket connection: %v", err)
		return false
	}

	sessionToken := sessionCookie.Value
	if sessionToken == "" {
		m.logger.Error("Empty session token for WebSocket connection")
		return false
	}

	// Validate session in database
	if db == nil {
		m.logger.Error("Database connection not available for WebSocket authentication")
		return false
	}

	var dbUserID int
	var username string
	query := `SELECT userid, Username FROM user WHERE current_session = ?`

	err = db.QueryRow(query, sessionToken).Scan(&dbUserID, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			m.logger.Error("Invalid or expired session token for WebSocket connection")
		} else {
			m.logger.Error("Database error during WebSocket authentication: %v", err)
		}
		return false
	}

	// Verify that the user ID from the URL matches the session
	if dbUserID != userID {
		m.logger.Error("User ID mismatch: URL has %d, session has %d", userID, dbUserID)
		return false
	}

	m.logger.Info("WebSocket authentication successful for user %s (ID: %d)", username, userID)
	return true
}
