package repository

import (
	"connecthub/database"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Authentication and session management
	AuthenticateUser(identifier, password string) (*database.User, error)
	UpdateUserSession(userID int, sessionToken string) error
	GetUserBySession(sessionToken string) (*database.User, error)
	ValidateSession(sessionToken string) (int, error)

	// User management
	CreateUser(firstName, lastName, username, email, gender, dateOfBirth, password string) (int, error)
	GetUserByID(userID int) (*database.User, error)
	GetAllUsers() ([]database.User, error)
	UserExists(username, email string) (bool, error)
	EmailExists(email string) (bool, error)
	UsernameExists(username string) (bool, error)
}

// PostRepository defines the interface for post data operations
type PostRepository interface {
	// Post retrieval
	GetAllPosts() ([]database.Post, error)
	GetPostByID(postID int) (database.Post, error)
	GetFilteredPosts(filter string) ([]database.Post, error)
	GetPostsByCategory(categoryName string) ([]database.Post, error)
	GetPostsByUser(userID int) ([]database.Post, error)
	GetLikedPostsByUser(userID int) ([]database.Post, error)

	// Post management
	CreatePost(userID int, title, content string, categories []string) (int, error)

	// Comments
	GetCommentsForPost(postID int) ([]database.Comment, error)
	AddComment(postID, userID int, content string) error

	// Categories
	GetCategories() ([]database.Category, error)
	GetCategoriesForPost(postID int) ([]database.Category, error)
}

// MessageRepository defines the interface for message and conversation data operations
type MessageRepository interface {
	// Conversation management
	CreateConversation(participants []int) (int, error)
	GetUserConversations(userID int) ([]database.Conversation, error)
	GetConversationParticipants(conversationID int) ([]database.User, error)
	IsUserParticipant(conversationID, userID int) (bool, error)

	// Message management
	AddMessageToConversation(conversationID, senderID int, content string) (*database.Message, error)
	GetConversationMessages(conversationID, limit, offset int) ([]database.Message, error)
	MarkMessagesAsRead(conversationID, userID int) error
	GetUnreadMessageCount(conversationID, userID int) (int, error)
}
