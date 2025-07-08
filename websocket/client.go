package websocket

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a websocket client connection
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan Message

	// User ID associated with this connection
	UserID int

	// Time of the last ping/pong
	lastPing time.Time

	// Closed flag to prevent duplicate closes
	closed   bool
	closeMux sync.Mutex
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan Message, messageBufferSize),
		UserID:   userID,
		lastPing: time.Now(),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.lastPing = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("Unexpected close error: %v", err)
			} else {
				c.hub.logger.Debug("Connection closed: %v", err)
			}
			break
		}

		// Clean the message
		message = bytes.TrimSpace(bytes.Replace(message, []byte{'\n'}, []byte{' '}, -1))

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.hub.logger.Error("Error unmarshalling message: %v", err)
			// Send error message back to client
			c.send <- Message{
				Type:    "error",
				Content: "Invalid message format",
			}
			continue
		}

		// Validate message
		if err := c.validateMessage(&msg); err != nil {
			c.hub.logger.Error("Invalid message: %v", err)
			c.send <- Message{
				Type:    "error",
				Content: err.Error(),
			}
			continue
		}

		msg.UserID = c.UserID
		msg.Timestamp = time.Now()

		c.hub.logger.Debug("Received message from user %d of type %s", c.UserID, msg.Type)
		c.hub.broadcast <- msg
	}
}

// close safely closes the client connection
func (c *Client) close() {
	c.closeMux.Lock()
	defer c.closeMux.Unlock()

	if !c.closed {
		c.hub.unregister <- c
		c.conn.Close()
		c.closed = true
		c.hub.logger.Debug("Closed connection for user %d", c.UserID)
	}
}

// validateMessage checks if a message is valid based on its type
func (c *Client) validateMessage(msg *Message) error {
	if msg == nil {
		return errors.New("message cannot be nil")
	}

	switch msg.Type {
	case MessageTypePrivate:
		if msg.RecipientID <= 0 {
			return fmt.Errorf("private message requires valid recipient ID, got %d", msg.RecipientID)
		}
		if msg.Content == nil || msg.Content == "" {
			return errors.New("message content cannot be empty")
		}
		// Check if recipient is online
		if !c.hub.IsUserOnline(msg.RecipientID) {
			return errors.New("cannot send message: recipient is offline")
		}

		// For new conversations, check if both users are online
		if msg.IsNewConversation {
			if !c.hub.IsUserOnline(c.UserID) {
				return errors.New("cannot start conversation: you are offline")
			}
		}
		// For existing conversations, require conversation ID
		if !msg.IsNewConversation && msg.ConversationID <= 0 {
			return errors.New("conversation ID required for existing conversations")
		}
	case MessageTypeBroadcast:
		if msg.Content == nil || msg.Content == "" {
			return errors.New("message content cannot be empty")
		}
	case MessageTypeUserStatus:
		// Status updates are handled internally by the hub
		return errors.New("status updates are handled automatically")
	case MessageTypeNotification:
		if msg.Content == nil || msg.Content == "" {
			return errors.New("notification content cannot be empty")
		}
	case "get_online_users":
		// Handle get_online_users request without broadcasting
		onlineUsers := c.hub.GetOnlineUsers()
		c.send <- Message{
			Type: MessageTypeOnlineUsers,
			Content: map[string]interface{}{
				"users": onlineUsers,
			},
			Timestamp: time.Now(),
			UserID:    c.UserID,
		}
		// Return nil to silently handle request without error
		return nil
	case "ping":
		// Handle ping messages from client - respond with pong
		c.send <- Message{
			Type:      "pong",
			Content:   "pong",
			Timestamp: time.Now(),
			UserID:    c.UserID,
		}
		// Return nil to silently handle ping without error
		return nil
	case MessageTypeOnlineUsers:
		return errors.New("online_users updates are handled automatically")
	case MessageTypeTyping:
		// Validate typing indicator message
		if msg.RecipientID <= 0 {
			return fmt.Errorf("typing indicator requires valid recipient ID, got %d", msg.RecipientID)
		}
		if msg.Action != TypingActionStart && msg.Action != TypingActionStop {
			return fmt.Errorf("typing indicator requires valid action (start/stop), got %s", msg.Action)
		}
		// Check if recipient is online
		if !c.hub.IsUserOnline(msg.RecipientID) {
			return errors.New("cannot send typing indicator: recipient is offline")
		}
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return nil
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.hub.logger.Debug("Hub closed channel for user %d", c.UserID)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send each message individually instead of batching
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			messageBytes, err := json.Marshal(message)
			if err != nil {
				c.hub.logger.Error("Error marshalling message: %v", err)
				continue
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				c.hub.logger.Error("Error writing message: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.hub.logger.Debug("Ping failed for user %d: %v", c.UserID, err)
				return
			}
			c.lastPing = time.Now()
		}
	}
}
