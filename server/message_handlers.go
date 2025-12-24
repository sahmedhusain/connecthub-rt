package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"connecthub/database"
	"connecthub/websocket"
)

// Global WebSocket manager for message handlers
var globalWSManager *websocket.Manager

// SetWebSocketManager sets the global WebSocket manager
func SetWebSocketManager(manager *websocket.Manager) {
	globalWSManager = manager
}

// Message-related request/response types
type SendMessageRequest struct {
	ConversationID int    `json:"conversation_id"`
	Content        string `json:"content"`
}

type SendMessageResponse struct {
	Success bool        `json:"success"`
	Message interface{} `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateConversationRequest struct {
	Participants []int `json:"participants"`
}

type CreateConversationResponse struct {
	Success        bool   `json:"success"`
	ConversationID int    `json:"conversation_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

// SendMessageAPI handles POST /api/messages
func SendMessageAPI(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	if r.Method != "POST" {
		log.Printf("[WARN] SendMessageAPI: Method not allowed: %s from %s", r.Method, clientIP)
		WriteAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	log.Printf("[INFO] SendMessageAPI: Processing POST request from %s", clientIP)

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] SendMessageAPI: Failed to decode request: %v", err)
		WriteAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request format")
		return
	}

	if req.ConversationID == 0 || strings.TrimSpace(req.Content) == "" {
		log.Printf("[WARN] SendMessageAPI: Missing conversation_id or content: conversation_id=%v, content='%v'", req.ConversationID, req.Content)
		WriteAPIError(w, http.StatusBadRequest, "MISSING_PARAMETER", "Missing conversation_id or content")
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] SendMessageAPI: Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendMessageResponse{Success: false, Error: "Database connection failed"})
		return
	}
	defer db.Close()

	// Get sender user ID from session
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] SendMessageAPI: No session cookie found from %s: %v", clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(SendMessageResponse{Success: false, Error: "Unauthorized"})
		return
	}

	var senderID int
	maskedToken := maskSessionToken(seshCok.Value)
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&senderID)
	if err != nil {
		log.Printf("[WARN] SendMessageAPI: Invalid session token %s from %s: %v", maskedToken, clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(SendMessageResponse{Success: false, Error: "Invalid session"})
		return
	}

	// Insert the message
	msg, err := database.AddMessageToConversation(db, req.ConversationID, senderID, req.Content)
	if err != nil {
		log.Printf("[ERROR] SendMessageAPI: Failed to insert message for conversation ID %d: %v", req.ConversationID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SendMessageResponse{Success: false, Error: "Failed to send message"})
		return
	}

	log.Printf("[INFO] SendMessageAPI: Message sent successfully for conversation ID %d from sender ID %d", req.ConversationID, senderID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SendMessageResponse{
		Success: true,
		Message: msg,
	})
}

// GetMessages handles GET /api/messages
func GetMessages(w http.ResponseWriter, r *http.Request) {
	conversationIDStr := r.URL.Query().Get("conversation_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if conversationIDStr == "" {
		log.Printf("[WARN] GetMessages: Missing conversation_id parameter")
		http.Error(w, "Missing conversation_id parameter", http.StatusBadRequest)
		return
	}

	conversationID, err := strconv.Atoi(conversationIDStr)
	if err != nil {
		log.Printf("[WARN] GetMessages: Invalid conversation_id: %s", conversationIDStr)
		http.Error(w, "Invalid conversation_id", http.StatusBadRequest)
		return
	}

	limit := 50 // Default limit - increased for better user experience
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] GetMessages: Database connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Verify user has access to this conversation
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] GetMessages: No session cookie found")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID)
	if err != nil {
		log.Printf("[WARN] GetMessages: Invalid session: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is participant in conversation
	var participantCount int
	err = db.QueryRow("SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?", conversationID, userID).Scan(&participantCount)
	if err != nil || participantCount == 0 {
		log.Printf("[WARN] GetMessages: User %d not authorized for conversation %d", userID, conversationID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	messages, err := database.GetConversationMessages(db, conversationID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetMessages: Failed to fetch messages: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] GetMessages: Retrieved %d messages for conversation %d", len(messages), conversationID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetConversations handles GET /api/conversations
func GetConversations(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] GetConversations: Database connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] GetConversations: No session cookie found")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID)
	if err != nil {
		log.Printf("[WARN] GetConversations: Invalid session: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conversations, err := database.GetUserConversations(db, userID)
	if err != nil {
		log.Printf("[ERROR] GetConversations: Failed to fetch conversations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] GetConversations: Retrieved %d conversations for user %d", len(conversations), userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// MarkMessagesAsReadAPI handles POST /api/messages/read
func MarkMessagesAsReadAPI(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	if r.Method != "POST" {
		log.Printf("[WARN] MarkMessagesAsReadAPI: Method not allowed: %s from %s", r.Method, clientIP)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	log.Printf("[INFO] MarkMessagesAsReadAPI: Processing POST request from %s", clientIP)

	var req struct {
		ConversationID int `json:"conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] MarkMessagesAsReadAPI: Failed to decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid request format"})
		return
	}

	if req.ConversationID <= 0 {
		log.Printf("[WARN] MarkMessagesAsReadAPI: Invalid conversation_id: %v", req.ConversationID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid conversation_id"})
		return
	}

	// Get database connection
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] MarkMessagesAsReadAPI: Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Internal server error"})
		return
	}
	defer db.Close()

	// Get user ID from session
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] MarkMessagesAsReadAPI: No session cookie found")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Unauthorized"})
		return
	}

	var userID int
	maskedToken := maskSessionToken(seshCok.Value)
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID)
	if err != nil {
		log.Printf("[WARN] MarkMessagesAsReadAPI: Invalid session token %s from %s: %v", maskedToken, clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid session"})
		return
	}

	// Check if user is participant in conversation
	var participantCount int
	err = db.QueryRow("SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?", req.ConversationID, userID).Scan(&participantCount)
	if err != nil || participantCount == 0 {
		log.Printf("[WARN] MarkMessagesAsReadAPI: User %d not authorized for conversation %d", userID, req.ConversationID)
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Forbidden"})
		return
	}

	// Mark messages as read
	err = database.MarkMessagesAsRead(db, req.ConversationID, userID)
	if err != nil {
		log.Printf("[ERROR] MarkMessagesAsReadAPI: Failed to mark messages as read: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to mark messages as read"})
		return
	}

	log.Printf("[INFO] MarkMessagesAsReadAPI: Messages marked as read for conversation %d by user %d", req.ConversationID, userID)

	// CRITICAL FIX: Send read status update via WebSocket to notify message senders
	if globalWSManager != nil {
		globalWSManager.SendReadStatusUpdate(req.ConversationID, userID)
		log.Printf("[INFO] MarkMessagesAsReadAPI: Read status update sent via WebSocket for conversation %d", req.ConversationID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// CreateConversationAPI handles POST /api/conversations
func CreateConversationAPI(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	if r.Method != "POST" {
		log.Printf("[WARN] CreateConversationAPI: Method not allowed: %s from %s", r.Method, clientIP)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "Method not allowed"})
		return
	}

	log.Printf("[INFO] CreateConversationAPI: Processing POST request from %s", clientIP)

	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] CreateConversationAPI: Failed to decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "Invalid request format"})
		return
	}

	if len(req.Participants) < 2 {
		log.Printf("[WARN] CreateConversationAPI: At least two participants required, received %d", len(req.Participants))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "At least two participants required"})
		return
	}

	// Get current user from session
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] CreateConversationAPI: No session cookie found from %s: %v", clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "Unauthorized"})
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] CreateConversationAPI: Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "Database connection failed"})
		return
	}
	defer db.Close()

	var currentUserID int
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&currentUserID)
	if err != nil {
		log.Printf("[WARN] CreateConversationAPI: Invalid session: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "Invalid session"})
		return
	}

	// Ensure current user is included in participants
	userIncluded := false
	for _, participantID := range req.Participants {
		if participantID == currentUserID {
			userIncluded = true
			break
		}
	}
	if !userIncluded {
		req.Participants = append(req.Participants, currentUserID)
	}

	convID, err := database.CreateConversation(req.Participants)
	if err != nil {
		log.Printf("[ERROR] CreateConversationAPI: Failed to create conversation: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreateConversationResponse{Success: false, Error: "Failed to create conversation"})
		return
	}

	log.Printf("[INFO] CreateConversationAPI: Successfully created conversation ID %d with %d participants", convID, len(req.Participants))

	json.NewEncoder(w).Encode(CreateConversationResponse{
		Success:        true,
		ConversationID: convID,
	})
}
