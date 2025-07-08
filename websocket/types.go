package websocket

import "time"

// Buffer sizes
const (
	readBufferSize    = 1024
	writeBufferSize   = 1024
	maxMessageSize    = 512 * 1024 // 512KB max message size
	messageBufferSize = 256        // Size of message buffer per client
)

// Timeouts
const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Message types
const (
	MessageTypePrivate         = "private"
	MessageTypeBroadcast       = "broadcast"
	MessageTypeUserStatus      = "user_status"
	MessageTypeNotification    = "notification"
	MessageTypeOnlineUsers     = "online_users"
	MessageTypeTyping          = "typing"
	MessageTypeNewConversation = "new_conversation"
	MessageTypeReadStatus      = "read_status" // CRITICAL FIX: Add read status message type
)

// Typing action types
const (
	TypingActionStart = "start"
	TypingActionStop  = "stop"
)

// Hub configuration defaults
const (
	DefaultMaxClients      = 10000
	DefaultRateLimitPeriod = time.Minute
	DefaultMessageRate     = 100 // messages per rate limit period
)

// Message represents a message in the chat system
type Message struct {
	Type              string      `json:"type"`
	From              int         `json:"from"`
	To                int         `json:"to,omitempty"`
	Content           interface{} `json:"content"`
	Timestamp         time.Time   `json:"timestamp"`
	UserID            int         `json:"user_id"`                       // ID of the user sending the message
	RecipientID       int         `json:"recipient_id"`                  // ID of the recipient user
	Data              interface{} `json:"data,omitempty"`                // Optional data payload
	IsNewConversation bool        `json:"is_new_conversation,omitempty"` // Whether this starts a new conversation
	ConversationID    int         `json:"conversation_id,omitempty"`     // ID of the conversation this message belongs to
	Code              string      `json:"code,omitempty"`                // Error code for error messages

	// Additional fields for database integration and frontend compatibility
	ID         int       `json:"id,omitempty"`          // Message ID from database
	MessageID  int       `json:"message_id,omitempty"`  // Alternative message ID field
	SenderID   int       `json:"sender_id,omitempty"`   // Sender ID (alias for UserID for frontend compatibility)
	SenderName string    `json:"sender_name,omitempty"` // Sender username
	SentAt     time.Time `json:"sent_at,omitempty"`     // When the message was sent
	IsRead     bool      `json:"is_read,omitempty"`     // Whether the message has been read

	// Typing indicator fields
	Action string `json:"action,omitempty"` // For typing messages: "start" or "stop"
}

// HubConfig contains configuration options for the Hub
type HubConfig struct {
	MaxClients      int
	RateLimitPeriod time.Duration
	MessageRate     int
	Debug           bool
}
