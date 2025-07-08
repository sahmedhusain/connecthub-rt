package repository

import (
	"database/sql"
	"log"

	"forum/database"
)

// MessageRepositoryImpl implements the MessageRepository interface
type MessageRepositoryImpl struct {
	db *sql.DB
}

// NewMessageRepository creates a new MessageRepository instance
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

// CreateConversation creates a new conversation with participants
func (r *MessageRepositoryImpl) CreateConversation(participants []int) (int, error) {
	log.Printf("[DEBUG] MessageRepository: Creating conversation with %d participants", len(participants))
	return database.CreateConversation(participants)
}

// GetUserConversations retrieves all conversations for a user
func (r *MessageRepositoryImpl) GetUserConversations(userID int) ([]database.Conversation, error) {
	log.Printf("[DEBUG] MessageRepository: Getting conversations for user %d", userID)
	return database.GetUserConversations(r.db, userID)
}

// GetConversationParticipants retrieves participants for a conversation
func (r *MessageRepositoryImpl) GetConversationParticipants(conversationID int) ([]database.User, error) {
	log.Printf("[DEBUG] MessageRepository: Getting participants for conversation %d", conversationID)

	// This function needs to be implemented in the database package
	// For now, return empty slice
	participants := []database.User{}

	log.Printf("[INFO] MessageRepository: Retrieved %d participants for conversation %d", len(participants), conversationID)
	return participants, nil
}

// IsUserParticipant checks if a user is a participant in a conversation
func (r *MessageRepositoryImpl) IsUserParticipant(conversationID, userID int) (bool, error) {
	log.Printf("[DEBUG] MessageRepository: Checking if user %d is participant in conversation %d", userID, conversationID)

	var count int
	query := `SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ? AND user_id = ?`
	err := r.db.QueryRow(query, conversationID, userID).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] MessageRepository: Failed to check participant status: %v", err)
		return false, err
	}

	isParticipant := count > 0
	log.Printf("[DEBUG] MessageRepository: User %d participant status for conversation %d: %v", userID, conversationID, isParticipant)
	return isParticipant, nil
}

// AddMessageToConversation adds a message to a conversation
func (r *MessageRepositoryImpl) AddMessageToConversation(conversationID, senderID int, content string) (*database.Message, error) {
	log.Printf("[DEBUG] MessageRepository: Adding message to conversation %d from user %d", conversationID, senderID)
	return database.AddMessageToConversation(r.db, conversationID, senderID, content)
}

// GetConversationMessages retrieves messages for a conversation with pagination
func (r *MessageRepositoryImpl) GetConversationMessages(conversationID, limit, offset int) ([]database.Message, error) {
	log.Printf("[DEBUG] MessageRepository: Getting messages for conversation %d (limit: %d, offset: %d)", conversationID, limit, offset)
	return database.GetConversationMessages(r.db, conversationID, limit, offset)
}

// MarkMessagesAsRead marks messages in a conversation as read by a user
func (r *MessageRepositoryImpl) MarkMessagesAsRead(conversationID, userID int) error {
	log.Printf("[DEBUG] MessageRepository: Marking messages as read for conversation %d by user %d", conversationID, userID)
	return database.MarkMessagesAsRead(r.db, conversationID, userID)
}

// GetUnreadMessageCount gets the count of unread messages for a user in a conversation
func (r *MessageRepositoryImpl) GetUnreadMessageCount(conversationID, userID int) (int, error) {
	log.Printf("[DEBUG] MessageRepository: Getting unread message count for user %d in conversation %d", userID, conversationID)
	return database.GetUnreadMessageCount(r.db, conversationID, userID)
}
