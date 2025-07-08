package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"forum/database"
)

// MessageService handles message and conversation-related business logic
type MessageService struct {
	db *sql.DB
}

// NewMessageService creates a new MessageService instance
func NewMessageService(db *sql.DB) *MessageService {
	return &MessageService{db: db}
}

// SendMessage sends a message to a conversation with validation
func (s *MessageService) SendMessage(conversationID, senderID int, content string) (*database.Message, error) {
	log.Printf("[DEBUG] MessageService: Sending message to conversation %d from user %d", conversationID, senderID)

	// Validate input
	if conversationID <= 0 {
		return nil, fmt.Errorf("invalid conversation ID")
	}

	if senderID <= 0 {
		return nil, fmt.Errorf("invalid sender ID")
	}

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("message content cannot be empty")
	}

	// Verify user is participant in conversation
	isParticipant, err := s.isUserParticipant(conversationID, senderID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to verify participant: %v", err)
		return nil, fmt.Errorf("failed to verify conversation access")
	}

	if !isParticipant {
		log.Printf("[WARN] MessageService: User %d not authorized for conversation %d", senderID, conversationID)
		return nil, fmt.Errorf("user not authorized for this conversation")
	}

	// Send the message
	message, err := database.AddMessageToConversation(s.db, conversationID, senderID, content)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to send message: %v", err)
		return nil, err
	}

	log.Printf("[INFO] MessageService: Message sent successfully to conversation %d", conversationID)
	return message, nil
}

// GetConversationMessages retrieves messages for a conversation with pagination
func (s *MessageService) GetConversationMessages(conversationID, userID, limit, offset int) ([]database.Message, error) {
	log.Printf("[DEBUG] MessageService: Getting messages for conversation %d (limit: %d, offset: %d)", conversationID, limit, offset)

	// Validate input
	if conversationID <= 0 {
		return nil, fmt.Errorf("invalid conversation ID")
	}

	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Verify user is participant in conversation
	isParticipant, err := s.isUserParticipant(conversationID, userID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to verify participant: %v", err)
		return nil, fmt.Errorf("failed to verify conversation access")
	}

	if !isParticipant {
		log.Printf("[WARN] MessageService: User %d not authorized for conversation %d", userID, conversationID)
		return nil, fmt.Errorf("user not authorized for this conversation")
	}

	// Set default pagination values
	if limit <= 0 {
		limit = 50 // Increased default for better user experience
	}
	if offset < 0 {
		offset = 0
	}

	// Get messages
	messages, err := database.GetConversationMessages(s.db, conversationID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to get messages: %v", err)
		return nil, err
	}

	log.Printf("[INFO] MessageService: Retrieved %d messages for conversation %d", len(messages), conversationID)
	return messages, nil
}

// GetUserConversations retrieves all conversations for a user
func (s *MessageService) GetUserConversations(userID int) ([]database.Conversation, error) {
	log.Printf("[DEBUG] MessageService: Getting conversations for user %d", userID)

	// Validate input
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Get conversations
	conversations, err := database.GetUserConversations(s.db, userID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to get conversations: %v", err)
		return nil, err
	}

	log.Printf("[INFO] MessageService: Retrieved %d conversations for user %d", len(conversations), userID)
	return conversations, nil
}

// CreateConversation creates a new conversation with participants
func (s *MessageService) CreateConversation(participants []int, currentUserID int) (int, error) {
	log.Printf("[DEBUG] MessageService: Creating conversation with %d participants", len(participants))

	// Validate input
	if len(participants) < 2 {
		return 0, fmt.Errorf("at least two participants required")
	}

	if currentUserID <= 0 {
		return 0, fmt.Errorf("invalid current user ID")
	}

	// Validate all participant IDs
	for _, participantID := range participants {
		if participantID <= 0 {
			return 0, fmt.Errorf("invalid participant ID: %d", participantID)
		}

		// Verify participant exists
		_, err := database.GetUserByID(s.db, participantID)
		if err != nil {
			log.Printf("[ERROR] MessageService: Participant %d not found: %v", participantID, err)
			return 0, fmt.Errorf("participant %d not found", participantID)
		}
	}

	// Ensure current user is included in participants
	userIncluded := false
	for _, participantID := range participants {
		if participantID == currentUserID {
			userIncluded = true
			break
		}
	}
	if !userIncluded {
		participants = append(participants, currentUserID)
	}

	// Create conversation
	conversationID, err := database.CreateConversation(participants)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to create conversation: %v", err)
		return 0, err
	}

	log.Printf("[INFO] MessageService: Created conversation %d with %d participants", conversationID, len(participants))
	return conversationID, nil
}

// GetConversationParticipants retrieves participants for a conversation
func (s *MessageService) GetConversationParticipants(conversationID, userID int) ([]database.User, error) {
	log.Printf("[DEBUG] MessageService: Getting participants for conversation %d", conversationID)

	// Validate input
	if conversationID <= 0 {
		return nil, fmt.Errorf("invalid conversation ID")
	}

	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Verify user is participant in conversation
	isParticipant, err := s.isUserParticipant(conversationID, userID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to verify participant: %v", err)
		return nil, fmt.Errorf("failed to verify conversation access")
	}

	if !isParticipant {
		log.Printf("[WARN] MessageService: User %d not authorized for conversation %d", userID, conversationID)
		return nil, fmt.Errorf("user not authorized for this conversation")
	}

	// Get participants (this function needs to be implemented in database package)
	// For now, we'll return an empty slice
	participants := []database.User{}

	log.Printf("[INFO] MessageService: Retrieved %d participants for conversation %d", len(participants), conversationID)
	return participants, nil
}

// MarkMessagesAsRead marks messages in a conversation as read by a user
func (s *MessageService) MarkMessagesAsRead(conversationID, userID int) error {
	log.Printf("[DEBUG] MessageService: Marking messages as read for conversation %d by user %d", conversationID, userID)

	// Validate input
	if conversationID <= 0 {
		return fmt.Errorf("invalid conversation ID")
	}

	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	// Verify user is participant in conversation
	isParticipant, err := s.isUserParticipant(conversationID, userID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to verify participant: %v", err)
		return fmt.Errorf("failed to verify conversation access")
	}

	if !isParticipant {
		log.Printf("[WARN] MessageService: User %d not authorized for conversation %d", userID, conversationID)
		return fmt.Errorf("user not authorized for this conversation")
	}

	// Mark messages as read
	err = database.MarkMessagesAsRead(s.db, conversationID, userID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to mark messages as read: %v", err)
		return err
	}

	log.Printf("[INFO] MessageService: Messages marked as read for conversation %d by user %d", conversationID, userID)
	return nil
}

// Helper methods

// isUserParticipant checks if a user is a participant in a conversation
func (s *MessageService) isUserParticipant(conversationID, userID int) (bool, error) {
	log.Printf("[DEBUG] MessageService: Checking if user %d is participant in conversation %d", userID, conversationID)

	var count int
	query := `SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?`
	err := s.db.QueryRow(query, conversationID, userID).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to check participant status: %v", err)
		return false, err
	}

	isParticipant := count > 0
	log.Printf("[DEBUG] MessageService: User %d participant status for conversation %d: %v", userID, conversationID, isParticipant)
	return isParticipant, nil
}

// GetUnreadMessageCount gets the count of unread messages for a user in a conversation
func (s *MessageService) GetUnreadMessageCount(conversationID, userID int) (int, error) {
	log.Printf("[DEBUG] MessageService: Getting unread message count for user %d in conversation %d", userID, conversationID)

	// Validate input
	if conversationID <= 0 {
		return 0, fmt.Errorf("invalid conversation ID")
	}

	if userID <= 0 {
		return 0, fmt.Errorf("invalid user ID")
	}

	// Get unread count
	count, err := database.GetUnreadMessageCount(s.db, conversationID, userID)
	if err != nil {
		log.Printf("[ERROR] MessageService: Failed to get unread message count: %v", err)
		return 0, err
	}

	log.Printf("[INFO] MessageService: User %d has %d unread messages in conversation %d", userID, count, conversationID)
	return count, nil
}
