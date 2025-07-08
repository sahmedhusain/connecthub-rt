package websocket

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var db *sql.DB

// updateUserStatusInDB updates a user's online status in the database
func updateUserStatusInDB(userID int, status string) error {
	if db == nil {
		return fmt.Errorf("database connection not initialized")
	}

	query := `
        INSERT INTO online_status (user_id, status, last_seen)
        VALUES (?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(user_id) DO UPDATE SET
            status = excluded.status,
            last_seen = CURRENT_TIMESTAMP
    `

	_, err := db.Exec(query, userID, status)
	if err != nil {
		log.Printf("[ERROR] Failed to update online status for user %d: %v", userID, err)
		return err
	}

	log.Printf("[INFO] Updated online status for user %d to %s", userID, status)
	return nil
}

// SetDB sets the database connection for the WebSocket hub
func SetDB(database *sql.DB) {
	db = database
}

// Use constants from types.go

type Logger struct {
	debug bool
}

func NewLogger(debug bool) *Logger {
	return &Logger{debug: debug}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.debug {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("[WEBSOCKET-DEBUG] %s:%d "+format, append([]interface{}{file, line}, v...)...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	log.Printf("[WEBSOCKET-INFO] "+format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	log.Printf("[WEBSOCKET-ERROR] %s:%d "+format, append([]interface{}{file, line}, v...)...)
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// User ID to client mapping
	userConnections map[int]*Client

	// Inbound messages from the clients
	broadcast chan Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger instance
	logger *Logger

	// Metrics and monitoring
	stats struct {
		messagesSent      uint64
		messagesReceived  uint64
		connectionsTotal  uint64
		connectionsActive uint64
		lastActivity      time.Time
		errors            uint64
	}

	// Configuration
	config HubConfig
}

func NewHub() *Hub {
	return NewHubWithLogging(false)
}

func NewHubWithLogging(debug bool) *Hub {
	hub := &Hub{
		broadcast:       make(chan Message, messageBufferSize),
		register:        make(chan *Client, 8),
		unregister:      make(chan *Client, 8),
		clients:         make(map[*Client]bool),
		userConnections: make(map[int]*Client),
		logger:          NewLogger(debug),
	}

	hub.config = HubConfig{
		MaxClients:      DefaultMaxClients,
		RateLimitPeriod: DefaultRateLimitPeriod,
		MessageRate:     DefaultMessageRate,
	}
	hub.stats.lastActivity = time.Now()

	return hub
}

func (h *Hub) Run() {
	h.logger.Info("WebSocket hub started")
	for {
		select {
		case client := <-h.register:
			// Check max clients limit
			if len(h.clients) >= h.config.MaxClients {
				h.logger.Error("Max clients limit reached, rejecting connection")
				close(client.send)
				continue
			}

			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			atomic.AddUint64(&h.stats.messagesReceived, 1)
			h.stats.lastActivity = time.Now()

			h.logger.Debug("Broadcasting message type: %s, from user: %d", message.Type, message.UserID)
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
	atomic.AddUint64(&h.stats.connectionsTotal, 1)
	atomic.AddUint64(&h.stats.connectionsActive, 1)
	h.stats.lastActivity = time.Now()

	h.logger.Debug("Client connected: %v", client.UserID)

	if client.UserID > 0 {
		h.mu.Lock()
		if existingClient, ok := h.userConnections[client.UserID]; ok {
			h.logger.Info("Replacing existing connection for user %d", client.UserID)
			existingClient.close()
		}
		h.userConnections[client.UserID] = client
		h.mu.Unlock()

		// Update online status in database
		if db != nil {
			err := updateUserStatusInDB(client.UserID, "online")
			if err != nil {
				h.logger.Error("Failed to update online status in database: %v", err)
			}
		}

		// Broadcast online status to other users
		h.broadcastUserStatus(client.UserID, true)

		// Send current online users list to new client
		onlineUsers := h.GetOnlineUsers()
		client.send <- Message{
			Type: MessageTypeOnlineUsers,
			Content: map[string]interface{}{
				"users": onlineUsers,
			},
			Timestamp: time.Now(),
			UserID:    client.UserID,
		}

		h.logger.Info("User %d connected and is now online", client.UserID)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		atomic.AddUint64(&h.stats.connectionsActive, ^uint64(0)) // Decrement
		h.stats.lastActivity = time.Now()

		h.logger.Debug("Client disconnected: %v", client.UserID)

		if client.UserID > 0 {
			h.mu.Lock()
			if currentClient, ok := h.userConnections[client.UserID]; ok && currentClient == client {
				delete(h.userConnections, client.UserID)
			}
			h.mu.Unlock()

			// Update online status in database
			if db != nil {
				err := updateUserStatusInDB(client.UserID, "offline")
				if err != nil {
					h.logger.Error("Failed to update offline status in database: %v", err)
				}
			}

			// Broadcast offline status to other users
			h.broadcastUserStatus(client.UserID, false)
			h.logger.Info("User %d disconnected and is now offline", client.UserID)
		}
	}
}

func (h *Hub) broadcastMessage(message Message) {
	start := time.Now()
	recipientCount := 0
	errorCount := 0

	if message.Type == MessageTypePrivate {
		// Handle private messages with database integration
		h.mu.RLock()
		recipientClient, ok := h.userConnections[message.RecipientID]
		senderClient := h.userConnections[message.UserID]
		h.mu.RUnlock()

		if !ok || !recipientClient.hub.IsUserOnline(message.RecipientID) {
			// Recipient is offline, send user-friendly error back to sender
			if senderClient != nil {
				senderClient.send <- Message{
					Type:    "error",
					Content: "The recipient is currently offline. Your message will be delivered when they come online.",
					Code:    "RECIPIENT_OFFLINE",
				}
			}
			return
		}

		// Process the message with database operations
		responseMessage, err := h.processPrivateMessage(message)
		if err != nil {
			h.logger.Error("Failed to process private message: %v", err)
			if senderClient != nil {
				// Provide user-friendly error message based on error type
				errorMessage := "Failed to send message. Please try again."
				errorCode := "MESSAGE_SEND_FAILED"

				if strings.Contains(err.Error(), "conversation") {
					errorMessage = "Conversation not found. It may have been deleted or you don't have access to it."
					errorCode = "CONVERSATION_NOT_FOUND"
				} else if strings.Contains(err.Error(), "database") {
					errorMessage = "We're experiencing technical difficulties. Please try again in a moment."
					errorCode = "DATABASE_ERROR"
				}

				senderClient.send <- Message{
					Type:    "error",
					Content: errorMessage,
					Code:    errorCode,
				}
			}
			return
		}

		// Send the processed message to recipient
		select {
		case recipientClient.send <- responseMessage:
			recipientCount++
			atomic.AddUint64(&h.stats.messagesSent, 1)
			h.logger.Debug("Message sent to recipient %d", message.RecipientID)
		default:
			errorCount++
			atomic.AddUint64(&h.stats.errors, 1)
			h.logger.Error("Failed to send message to recipient %d", message.RecipientID)
			if senderClient != nil {
				senderClient.send <- Message{
					Type:    "error",
					Content: "Failed to send message. Please check your connection and try again.",
					Code:    "MESSAGE_SEND_FAILED",
				}
			}
		}

		// This ensures the sender sees their message confirmed and gets proper status indicators
		if senderClient != nil {
			select {
			case senderClient.send <- responseMessage:
				recipientCount++
				atomic.AddUint64(&h.stats.messagesSent, 1)
				h.logger.Debug("Message confirmation sent to sender %d", message.UserID)
			default:
				errorCount++
				atomic.AddUint64(&h.stats.errors, 1)
				h.logger.Error("Failed to send message confirmation to sender %d", message.UserID)
			}
		}

		// Send confirmation back to sender with database-populated fields
		if senderClient != nil {
			select {
			case senderClient.send <- responseMessage:
				recipientCount++
				atomic.AddUint64(&h.stats.messagesSent, 1)
			default:
				errorCount++
				atomic.AddUint64(&h.stats.errors, 1)
				h.logger.Error("Failed to send confirmation to sender %d", message.UserID)
			}
		}
	} else if message.Type == MessageTypeTyping {
		// Handle typing indicators - send only to recipient
		h.mu.RLock()
		recipientClient, ok := h.userConnections[message.RecipientID]
		h.mu.RUnlock()

		if ok && recipientClient.hub.IsUserOnline(message.RecipientID) {
			// Get sender name for typing indicator
			var senderName string
			if db != nil {
				err := db.QueryRow("SELECT Username FROM user WHERE userid = ?", message.UserID).Scan(&senderName)
				if err != nil {
					h.logger.Error("Failed to get sender name for typing indicator: %v", err)
					senderName = "Someone"
				}
			} else {
				senderName = "Someone"
			}

			// Create typing indicator message
			typingMessage := Message{
				Type:           MessageTypeTyping,
				UserID:         message.UserID,
				RecipientID:    message.RecipientID,
				ConversationID: message.ConversationID,
				Action:         message.Action,
				SenderID:       message.UserID,
				SenderName:     senderName,
				Timestamp:      time.Now(),
			}

			select {
			case recipientClient.send <- typingMessage:
				recipientCount++
				atomic.AddUint64(&h.stats.messagesSent, 1)
				h.logger.Debug("Typing indicator sent to user %d: %s", message.RecipientID, message.Action)
			default:
				errorCount++
				atomic.AddUint64(&h.stats.errors, 1)
				h.logger.Error("Failed to send typing indicator to user %d", message.RecipientID)
			}
		}
	} else {
		// Handle other message types (broadcast, status, etc.)
		for client := range h.clients {
			if h.shouldSendToClient(message, client) {
				select {
				case client.send <- message:
					recipientCount++
					atomic.AddUint64(&h.stats.messagesSent, 1)
				default:
					errorCount++
					atomic.AddUint64(&h.stats.errors, 1)
					h.logger.Error("Failed to send message to client %d, removing client", client.UserID)
					client.close()
				}
			}
		}
	}

	duration := time.Since(start)
	h.logger.Debug("Message delivered to %d recipients (%d errors) in %v",
		recipientCount, errorCount, duration)
}

// processPrivateMessage handles database operations for private messages
func (h *Hub) processPrivateMessage(message Message) (Message, error) {
	if db == nil {
		return message, fmt.Errorf("database connection not available")
	}

	var conversationID int
	var err error

	// Get sender name from database
	var senderName string
	err = db.QueryRow("SELECT Username FROM user WHERE userid = ?", message.UserID).Scan(&senderName)
	if err != nil {
		h.logger.Error("Failed to get sender name for user %d: %v", message.UserID, err)
		senderName = "Unknown User"
	}

	if message.IsNewConversation {
		// Create new conversation
		h.logger.Info("Creating new conversation between users %d and %d", message.UserID, message.RecipientID)
		participants := []int{message.UserID, message.RecipientID}

		// Use the database package function to create conversation
		conversationID, err = h.createConversation(participants)
		if err != nil {
			return message, fmt.Errorf("failed to create conversation: %v", err)
		}
		h.logger.Info("Created conversation %d for new private message", conversationID)

		// Send new conversation notification to recipient
		h.sendNewConversationNotification(conversationID, message.UserID, message.RecipientID)
	} else {
		conversationID = message.ConversationID
		if conversationID <= 0 {
			return message, fmt.Errorf("invalid conversation ID for existing conversation")
		}
	}

	// Save message to database
	contentStr, ok := message.Content.(string)
	if !ok {
		return message, fmt.Errorf("message content must be a string")
	}

	// Use the database package function to add message
	dbMessage, err := h.addMessageToConversation(conversationID, message.UserID, contentStr)
	if err != nil {
		return message, fmt.Errorf("failed to save message to database: %v", err)
	}

	// Construct response message with database-populated fields
	responseMessage := Message{
		Type:              message.Type,
		UserID:            message.UserID,
		RecipientID:       message.RecipientID,
		Content:           message.Content,
		Timestamp:         time.Now(),
		ConversationID:    conversationID,
		IsNewConversation: message.IsNewConversation,

		// Database-populated fields for frontend compatibility
		ID:         dbMessage.ID,
		MessageID:  dbMessage.ID,
		SenderID:   dbMessage.SenderID,
		SenderName: dbMessage.SenderName,
		SentAt:     dbMessage.SentAt,
		IsRead:     dbMessage.IsRead,
	}

	h.logger.Info("Successfully processed private message %d in conversation %d", dbMessage.ID, conversationID)
	return responseMessage, nil
}

func (h *Hub) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"messagesSent":      h.stats.messagesSent,
		"messagesReceived":  h.stats.messagesReceived,
		"connectionsTotal":  h.stats.connectionsTotal,
		"connectionsActive": h.stats.connectionsActive,
		"lastActivity":      h.stats.lastActivity,
		"onlineUsers":       len(h.GetOnlineUsers()),
	}
}

func (h *Hub) SetDebugMode(debug bool) {
	h.logger.debug = debug
	h.logger.Info("Debug mode set to: %v", debug)
}

func (h *Hub) shouldSendToClient(message Message, client *Client) bool {
	switch message.Type {
	case MessageTypePrivate:
		return client.UserID == message.RecipientID
	case MessageTypeUserStatus:
		return client.UserID != message.UserID
	case MessageTypeTyping:
		// Typing indicators should only go to the recipient
		return client.UserID == message.RecipientID
	default:
		return true
	}
}

func (h *Hub) SendToUser(userID int, message Message) bool {
	h.mu.RLock()
	client, ok := h.userConnections[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.send <- message:
			h.stats.messagesSent++
			h.logger.Debug("Direct message sent to user %d", userID)
			return true
		default:
			h.logger.Error("Failed to send direct message to user %d", userID)
			return false
		}
	}
	h.logger.Debug("Attempted to send message to offline user %d", userID)
	return false
}

func (h *Hub) BroadcastToAll(message Message) {
	h.logger.Debug("Broadcasting message to all users: type=%s", message.Type)

	// Add basic rate limiting
	select {
	case h.broadcast <- message:
		h.logger.Debug("Message queued for broadcast")
	default:
		h.logger.Error("Broadcast channel full, message dropped")
		atomic.AddUint64(&h.stats.errors, 1)
	}
}

func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	_, online := h.userConnections[userID]
	h.mu.RUnlock()
	return online
}

func (h *Hub) GetOnlineUsers() []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]int, 0, len(h.userConnections))
	for userID := range h.userConnections {
		users = append(users, userID)
	}
	return users
}

func (h *Hub) broadcastUserStatus(userID int, online bool) {
	status := "offline"
	if online {
		status = "online"
	}

	h.logger.Info("Broadcasting user %d status change: %s", userID, status)
	h.broadcast <- Message{
		Type:   MessageTypeUserStatus,
		UserID: userID,
		Content: map[string]interface{}{
			"status": status,
			"userId": userID,
		},
	}
}

// sendNewConversationNotification sends a notification to the recipient about a new conversation
func (h *Hub) sendNewConversationNotification(conversationID int, senderID int, recipientID int) {
	h.mu.RLock()
	recipientClient, ok := h.userConnections[recipientID]
	h.mu.RUnlock()

	if !ok || !h.IsUserOnline(recipientID) {
		h.logger.Debug("Recipient %d is offline, skipping new conversation notification", recipientID)
		return
	}

	// Get sender information
	var senderName, senderFirstName, senderLastName, senderUsername string
	if db != nil {
		err := db.QueryRow("SELECT firstname, lastname, username FROM user WHERE userid = ?", senderID).Scan(&senderFirstName, &senderLastName, &senderUsername)
		if err != nil {
			h.logger.Error("Failed to get sender info for new conversation notification: %v", err)
			senderName = "Someone"
		} else {
			senderName = senderFirstName + " " + senderLastName
		}
	} else {
		senderName = "Someone"
		senderUsername = "unknown"
	}

	// Create new conversation notification
	notification := Message{
		Type:           MessageTypeNewConversation,
		UserID:         senderID,
		RecipientID:    recipientID,
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderName:     senderName,
		Timestamp:      time.Now(),
		Content: map[string]interface{}{
			"conversation_id": conversationID,
			"sender_id":       senderID,
			"sender_name":     senderName,
			"sender_username": senderUsername,
			"message":         "New conversation started",
		},
	}

	select {
	case recipientClient.send <- notification:
		h.logger.Info("New conversation notification sent to user %d for conversation %d", recipientID, conversationID)
	default:
		h.logger.Error("Failed to send new conversation notification to user %d", recipientID)
	}
}

// Helper methods for database operations

// createConversation creates a new conversation between participants
func (h *Hub) createConversation(participants []int) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database connection not available")
	}

	// Insert conversation
	result, err := db.Exec("INSERT INTO conversation (created_at) VALUES (?)", time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %v", err)
	}

	conversationID64, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get conversation ID: %v", err)
	}
	conversationID := int(conversationID64)

	// Add participants to conversation
	for _, participantID := range participants {
		_, err = db.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)",
			conversationID, participantID)
		if err != nil {
			h.logger.Error("Failed to add participant %d to conversation %d: %v", participantID, conversationID, err)
			// Continue with other participants even if one fails
		}
	}

	h.logger.Info("Created conversation %d with %d participants", conversationID, len(participants))
	return conversationID, nil
}

// DatabaseMessage represents a message from the database
type DatabaseMessage struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Content    string    `json:"content"`
	SentAt     time.Time `json:"sent_at"`
	IsRead     bool      `json:"is_read"`
}

// addMessageToConversation adds a message to a conversation
func (h *Hub) addMessageToConversation(conversationID, senderID int, content string) (*DatabaseMessage, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	// Insert message
	result, err := db.Exec("INSERT INTO message (conversation_id, sender_id, content, sent_at, is_read) VALUES (?, ?, ?, ?, ?)",
		conversationID, senderID, content, time.Now(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %v", err)
	}

	messageID64, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get message ID: %v", err)
	}
	messageID := int(messageID64)

	// Get sender name
	var senderName string
	err = db.QueryRow("SELECT Username FROM user WHERE userid = ?", senderID).Scan(&senderName)
	if err != nil {
		h.logger.Error("Failed to get sender name for user %d: %v", senderID, err)
		senderName = "Unknown User"
	}

	dbMessage := &DatabaseMessage{
		ID:         messageID,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    content,
		SentAt:     time.Now(),
		IsRead:     false,
	}

	h.logger.Info("Added message %d to conversation %d from user %d", messageID, conversationID, senderID)
	return dbMessage, nil
}
func (h *Hub) SendReadStatusUpdate(conversationID int, readerID int) {
	if db == nil {
		h.logger.Error("Database connection not available for read status update")
		return
	}

	// Get all participants in the conversation except the reader
	query := `
		SELECT DISTINCT cp.user_id, u.Username
		FROM conversation_participant cp
		JOIN user u ON cp.user_id = u.userid
		WHERE cp.conversation_id = ? AND cp.user_id != ?
	`

	rows, err := db.Query(query, conversationID, readerID)
	if err != nil {
		h.logger.Error("Failed to get conversation participants for read status: %v", err)
		return
	}
	defer rows.Close()

	// Get reader name
	var readerName string
	err = db.QueryRow("SELECT Username FROM user WHERE userid = ?", readerID).Scan(&readerName)
	if err != nil {
		h.logger.Error("Failed to get reader name: %v", err)
		readerName = "Someone"
	}

	// Send read status update to all other participants
	for rows.Next() {
		var participantID int
		var participantName string

		if err := rows.Scan(&participantID, &participantName); err != nil {
			h.logger.Error("Failed to scan participant: %v", err)
			continue
		}

		h.mu.RLock()
		participantClient, ok := h.userConnections[participantID]
		h.mu.RUnlock()

		if ok && h.IsUserOnline(participantID) {
			readStatusMessage := Message{
				Type:           MessageTypeReadStatus,
				ConversationID: conversationID,
				UserID:         readerID,
				SenderID:       readerID,
				SenderName:     readerName,
				Timestamp:      time.Now(),
				Content: map[string]interface{}{
					"conversation_id": conversationID,
					"reader_id":       readerID,
					"reader_name":     readerName,
				},
			}

			select {
			case participantClient.send <- readStatusMessage:
				h.logger.Debug("Read status update sent to user %d for conversation %d", participantID, conversationID)
			default:
				h.logger.Error("Failed to send read status update to user %d", participantID)
			}
		}
	}

	if err = rows.Err(); err != nil {
		h.logger.Error("Error iterating conversation participants: %v", err)
	}
}
