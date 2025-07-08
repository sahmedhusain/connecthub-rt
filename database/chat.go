package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

type ChatMessage struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	SenderID       int       `json:"sender_id"`
	Content        string    `json:"content"`
	SentAt         time.Time `json:"sent_at"`
	IsRead         bool      `json:"is_read"`
	SenderName     string    `json:"sender_name,omitempty"`
}

type Message struct {
	ID              int       `json:"id"`
	ConversationID  int       `json:"conversation_id"`
	SenderID        int       `json:"sender_id"`
	SenderName      string    `json:"sender_name,omitempty"`
	Content         string    `json:"content"`
	SentAt          time.Time `json:"sent_at"`
	IsRead          bool      `json:"is_read"`
	RecipientOnline bool      `json:"recipient_online"`
}

type Conversation struct {
	ID           int          `json:"id"`
	CreatedAt    time.Time    `json:"created_at"`
	Participants []*User      `json:"participants"`
	LastMessage  *ChatMessage `json:"last_message,omitempty"`
}

var DB *sql.DB

func CreateConversation(participants []int) (int, error) {
	if DB == nil {
		var err error
		log.Printf("[DEBUG] Attempting to connect to SQLite database for creating conversation")
		DB, err = sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Printf("[ERROR] Database connection failed in CreateConversation: %v", err)
			return 0, err
		}
		log.Printf("[INFO] Successfully connected to SQLite database for creating conversation")
	} else {
		log.Printf("[DEBUG] Using existing database connection for creating conversation")
	}

	tx, err := DB.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction in CreateConversation: %v", err)
		return 0, err
	}
	log.Printf("[DEBUG] Started transaction for creating conversation with %d participants", len(participants))

	if len(participants) == 2 {
		var existingConvID int
		log.Printf("[DEBUG] Checking for existing conversation between users %d and %d", participants[0], participants[1])
		err := tx.QueryRow(`
			SELECT cp1.conversation_id
			FROM conversation_participants cp1
			JOIN conversation_participants cp2 ON cp1.conversation_id = cp2.conversation_id
			WHERE cp1.user_id = ? AND cp2.user_id = ?
			GROUP BY cp1.conversation_id
			HAVING COUNT(DISTINCT cp1.user_id) = 1 AND COUNT(DISTINCT cp2.user_id) = 1
		`, participants[0], participants[1]).Scan(&existingConvID)

		if err == nil {
			tx.Rollback()
			log.Printf("[INFO] Found existing conversation (ID: %d) between users %d and %d", existingConvID, participants[0], participants[1])
			return existingConvID, nil
		} else if err != sql.ErrNoRows {
			tx.Rollback()
			log.Printf("[ERROR] Failed checking for existing conversation between users %d and %d: %v", participants[0], participants[1], err)
			return 0, err
		}
		log.Printf("[DEBUG] No existing conversation found between users %d and %d, creating new one", participants[0], participants[1])
	}

	res, err := tx.Exec("INSERT INTO conversation (created_at) VALUES (CURRENT_TIMESTAMP)")
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to insert into conversation table: %v", err)
		return 0, err
	}
	log.Printf("[DEBUG] Successfully inserted new conversation record")

	convID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to get last insert ID for conversation: %v", err)
		return 0, err
	}
	log.Printf("[DEBUG] Retrieved new conversation ID: %d", convID)

	stmt, err := tx.Prepare("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to prepare statement for conversation_participants: %v", err)
		return 0, err
	}
	defer stmt.Close()
	log.Printf("[DEBUG] Prepared statement for adding participants to conversation %d", convID)

	for _, userID := range participants {
		_, err := stmt.Exec(convID, userID)
		if err != nil {
			tx.Rollback()
			log.Printf("[ERROR] Failed to add user %d to conversation %d: %v", userID, convID, err)
			return 0, err
		}
		log.Printf("[INFO] Added user %d to conversation %d", userID, convID)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction for creating conversation: %v", err)
		return 0, err
	}
	log.Printf("[DEBUG] Committed transaction for creating conversation %d", convID)

	log.Printf("[INFO] Created conversation %d with participants: %v", int(convID), participants)
	return int(convID), nil
}

func GetConversationBetweenUsers(db *sql.DB, userID1, userID2 int) (int, error) {
	query := `
		SELECT cp1.conversation_id
		FROM conversation_participants cp1
		JOIN conversation_participants cp2 ON cp1.conversation_id = cp2.conversation_id
		WHERE cp1.user_id = ? AND cp2.user_id = ?
		GROUP BY cp1.conversation_id
		HAVING COUNT(DISTINCT cp1.user_id) + COUNT(DISTINCT cp2.user_id) = 2
		LIMIT 1
	`

	log.Printf("[DEBUG] Checking for conversation between users %d and %d", userID1, userID2)
	var conversationID int
	err := db.QueryRow(query, userID1, userID2).Scan(&conversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[INFO] No conversation found between users %d and %d", userID1, userID2)
			return 0, nil
		}
		log.Printf("[ERROR] Failed to get conversation between users %d and %d: %v", userID1, userID2, err)
		return 0, err
	}
	log.Printf("[DEBUG] Successfully retrieved conversation ID %d between users %d and %d", conversationID, userID1, userID2)

	log.Printf("[INFO] Found conversation %d between users %d and %d", conversationID, userID1, userID2)
	return conversationID, nil
}

func SaveChatMessage(db *sql.DB, senderID, conversationID int, content string) (int, error) {
	query := `
		INSERT INTO message (conversation_id, sender_id, content, sent_at, is_read)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, 0)
	`

	contentPreview := truncateContent(content)
	log.Printf("[DEBUG] Saving message from user %d in conversation %d: '%s'", senderID, conversationID, contentPreview)
	result, err := db.Exec(query, conversationID, senderID, content)
	if err != nil {
		log.Printf("[ERROR] Failed to save message from user %d in conversation %d: %v", senderID, conversationID, err)
		return 0, err
	}
	log.Printf("[INFO] Message saved successfully from user %d in conversation %d", senderID, conversationID)

	messageID, err := result.LastInsertId()
	if err != nil {
		log.Printf("[ERROR] Failed to get last insert ID for message: %v", err)
		return 0, err
	}
	log.Printf("[DEBUG] Retrieved new message ID: %d", messageID)

	log.Printf("[INFO] Saved message %d from user %d in conversation %d", int(messageID), senderID, conversationID)
	return int(messageID), nil
}

func IsUserInConversation(db *sql.DB, userID, conversationID int) (bool, error) {
	var count int

	query := `
		SELECT COUNT(*) FROM conversation_participants
		WHERE user_id = ? AND conversation_id = ?
	`

	log.Printf("[DEBUG] Checking if user %d is in conversation %d", userID, conversationID)
	err := db.QueryRow(query, userID, conversationID).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] Failed to check if user %d is in conversation %d: %v", userID, conversationID, err)
		return false, err
	}

	isInConversation := count > 0
	log.Printf("[INFO] User %d is%s in conversation %d", userID, map[bool]string{true: "", false: " not"}[isInConversation], conversationID)
	return isInConversation, nil
}

func GetConversationParticipants(db *sql.DB, conversationID int) ([]int, error) {
	participants := []int{}

	query := `
		SELECT user_id FROM conversation_participants
		WHERE conversation_id = ?
	`

	log.Printf("[DEBUG] Retrieving participants for conversation %d", conversationID)
	rows, err := db.Query(query, conversationID)
	if err != nil {
		log.Printf("[ERROR] Failed to get participants for conversation %d: %v", conversationID, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("[DEBUG] Successfully queried participants for conversation %d", conversationID)

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			log.Printf("[ERROR] Failed to scan participant ID for conversation %d: %v", conversationID, err)
			return nil, err
		}
		log.Printf("[DEBUG] Found participant ID %d in conversation %d", userID, conversationID)
		participants = append(participants, userID)
	}

	log.Printf("[INFO] Found %d participants in conversation %d: %v", len(participants), conversationID, participants)
	return participants, nil
}

func GetConversationMessages(db *sql.DB, conversationID, limit, offset int) ([]Message, error) {
	messages := []Message{}

	// PAGINATION FIX: Order by sent_at DESC to get newest messages first for proper pagination
	// This allows offset to work correctly - offset 0 gets the newest messages
	// Frontend will reverse the order for display if needed
	query := `
		SELECT m.message_id, m.conversation_id, m.sender_id, u.Username, m.content, m.sent_at, m.is_read
		FROM message m
		JOIN user u ON m.sender_id = u.userid
		WHERE m.conversation_id = ?
		ORDER BY m.sent_at DESC
		LIMIT ? OFFSET ?
	`

	log.Printf("[DEBUG] Retrieving messages for conversation %d with limit %d and offset %d", conversationID, limit, offset)
	rows, err := db.Query(query, conversationID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve messages for conversation %d (limit: %d, offset: %d): %v", conversationID, limit, offset, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("[DEBUG] Successfully queried messages for conversation %d", conversationID)

	for rows.Next() {
		var msg Message
		var sentAtStr string
		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.SenderName,
			&msg.Content, &sentAtStr, &msg.IsRead,
		)
		if err != nil {
			log.Printf("[ERROR] Failed to scan message from conversation %d: %v", conversationID, err)
			return nil, err
		}
		log.Printf("[DEBUG] Scanned message ID %d from conversation %d", msg.ID, conversationID)

		msg.SentAt, err = time.Parse(time.RFC3339, sentAtStr)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			msg.SentAt, err = time.Parse(layout, sentAtStr)
			if err != nil {
				log.Printf("[WARN] Failed to parse timestamp '%s' for message %d: %v", sentAtStr, msg.ID, err)
				msg.SentAt = time.Time{}
			}
			log.Printf("[DEBUG] Parsed timestamp for message %d: %v", msg.ID, msg.SentAt)
		}

		messages = append(messages, msg)
	}

	// PAGINATION FIX: Remove array reversal since we now order by DESC in the query
	// Messages are returned in newest-first order, frontend will handle display order

	log.Printf("[INFO] Retrieved %d messages from conversation %d (limit: %d, offset: %d)", len(messages), conversationID, limit, offset)
	return messages, nil
}

func MarkMessagesAsRead(db *sql.DB, conversationID, userID int) error {
	query := `
		UPDATE message
		SET is_read = 1
		WHERE conversation_id = ? AND sender_id != ? AND is_read = 0
	`

	log.Printf("[DEBUG] Marking messages as read in conversation %d for user %d", conversationID, userID)
	result, err := db.Exec(query, conversationID, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to mark messages as read in conversation %d for user %d: %v", conversationID, userID, err)
		return err
	}

	affected, _ := result.RowsAffected()
	log.Printf("[INFO] Marked %d messages as read in conversation %d for user %d", affected, conversationID, userID)
	return nil
}

func GetUserConversations(db *sql.DB, userID int) ([]Conversation, error) {
	conversations := []Conversation{}

	log.Printf("[DEBUG] Retrieving conversations for user %d", userID)
	rows, err := db.Query(`
		SELECT c.conversation_id, c.created_at
		FROM conversation c
		JOIN conversation_participants cp ON c.conversation_id = cp.conversation_id
		WHERE cp.user_id = ?
		ORDER BY (
			SELECT MAX(sent_at)
			FROM message
			WHERE conversation_id = c.conversation_id
		) DESC
	`, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to get conversations for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("[DEBUG] Successfully queried conversations for user %d", userID)

	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.CreatedAt)
		if err != nil {
			log.Printf("[ERROR] Failed to scan conversation for user %d: %v", userID, err)
			return nil, err
		}
		log.Printf("[DEBUG] Scanned conversation ID %d for user %d", conv.ID, userID)

		users, err := getConversationParticipants(conv.ID, db)
		if err != nil {
			log.Printf("[ERROR] Failed to get participants for conversation %d: %v", conv.ID, err)
			return nil, err
		}
		participants := make([]*User, len(users))
		for i := range users {
			participants[i] = &users[i]
		}
		conv.Participants = participants
		log.Printf("[DEBUG] Retrieved %d participants for conversation %d", len(participants), conv.ID)

		lastMsg, err := getLastMessage(conv.ID, db)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("[ERROR] Failed to get last message for conversation %d: %v", conv.ID, err)
		} else if err == sql.ErrNoRows {
			conv.LastMessage = nil
			log.Printf("[DEBUG] No last message found for conversation %d", conv.ID)
		} else {
			conv.LastMessage = &ChatMessage{
				ID:             lastMsg.ID,
				ConversationID: lastMsg.ConversationID,
				SenderID:       lastMsg.SenderID,
				SenderName:     lastMsg.SenderName,
				Content:        lastMsg.Content,
				SentAt:         lastMsg.SentAt,
				IsRead:         lastMsg.IsRead,
			}
			log.Printf("[DEBUG] Retrieved last message ID %d for conversation %d", lastMsg.ID, conv.ID)
		}

		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating conversations for user %d: %v", userID, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d conversations for user %d", len(conversations), userID)
	return conversations, nil
}

func getConversationParticipants(conversationID int, db *sql.DB) ([]User, error) {
	participants := []User{}

	log.Printf("[DEBUG] Querying participants for conversation %d", conversationID)
	rows, err := db.Query(`
		SELECT u.userid, u.Username, u.Email, u.F_name, u.L_name, u.Avatar
		FROM user u
		JOIN conversation_participants cp ON u.userid = cp.user_id
		WHERE cp.conversation_id = ?
	`, conversationID)
	if err != nil {
		log.Printf("[ERROR] Failed to query participants for conversation %d: %v", conversationID, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("[DEBUG] Successfully queried participants for conversation %d", conversationID)

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Avatar)
		if err != nil {
			log.Printf("[ERROR] Failed to scan participant for conversation %d: %v", conversationID, err)
			return nil, err
		}
		log.Printf("[DEBUG] Scanned participant ID %d for conversation %d", user.ID, conversationID)
		participants = append(participants, user)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating participants for conversation %d: %v", conversationID, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d participants for conversation %d", len(participants), conversationID)
	return participants, nil
}

func getLastMessage(conversationID int, db *sql.DB) (*Message, error) {
	var msg Message
	var sentAtStr string

	log.Printf("[DEBUG] Retrieving last message for conversation %d", conversationID)
	err := db.QueryRow(`
		SELECT m.message_id, m.conversation_id, m.sender_id, u.Username, m.content, m.sent_at, m.is_read
		FROM message m
		JOIN user u ON m.sender_id = u.userid
		WHERE m.conversation_id = ?
		ORDER BY m.sent_at DESC
		LIMIT 1
	`, conversationID).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.SenderName,
		&msg.Content, &sentAtStr, &msg.IsRead,
	)
	log.Printf("[DEBUG] Successfully queried last message for conversation %d", conversationID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[INFO] No messages found in conversation %d", conversationID)
			return nil, err
		}
		log.Printf("[ERROR] Failed to get last message for conversation %d: %v", conversationID, err)
		return nil, err
	}

	msg.SentAt, err = time.Parse(time.RFC3339, sentAtStr)
	if err != nil {
		layout := "2006-01-02 15:04:05"
		msg.SentAt, err = time.Parse(layout, sentAtStr)
		if err != nil {
			log.Printf("[WARN] Failed to parse timestamp '%s' for last message in conversation %d: %v", sentAtStr, conversationID, err)
			msg.SentAt = time.Time{}
		}
	}
	log.Printf("[DEBUG] Retrieved last message ID %d for conversation %d", msg.ID, conversationID)

	return &msg, nil
}

func AddMessageToConversation(db *sql.DB, conversationID, senderID int, content string) (*Message, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction for adding message: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Started transaction for adding message from user %d to conversation %d", senderID, conversationID)
	contentPreview := truncateContent(content)
	log.Printf("[DEBUG] Content of message to be added: '%s'", contentPreview)

	// Get recipient ID and check if they're online
	var recipientID int
	err = tx.QueryRow(`
        SELECT user_id 
        FROM conversation_participants 
        WHERE conversation_id = ? AND user_id != ?
        LIMIT 1
    `, conversationID, senderID).Scan(&recipientID)
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to get recipient ID: %v", err)
		return nil, err
	}

	// Check if recipient is online (for notification purposes, but allow offline messaging)
	var isOnline bool
	err = tx.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM online_status
            WHERE user_id = ?
            AND status = 'online'
            AND last_seen > datetime('now', '-5 minutes')
        )
    `, recipientID).Scan(&isOnline)
	if err != nil {
		log.Printf("[WARN] Failed to check recipient online status (continuing anyway): %v", err)
		isOnline = false // Default to offline if check fails
	}

	log.Printf("[DEBUG] Recipient %d online status: %v (allowing message regardless)", recipientID, isOnline)

	// Insert message regardless of recipient online status (modern chat behavior)
	res, err := tx.Exec(`
        INSERT INTO message (conversation_id, sender_id, content, sent_at, is_read)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP, 0)
    `, conversationID, senderID, content)

	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to insert message from user %d in conversation %d: %v", senderID, conversationID, err)
		return nil, err
	}
	log.Printf("[DEBUG] Successfully inserted message from user %d in conversation %d", senderID, conversationID)

	messageID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to get last insert ID for message: %v", err)
		return nil, err
	}
	log.Printf("[DEBUG] Retrieved new message ID: %d", messageID)

	var msg Message
	var sentAtStr string
	err = tx.QueryRow(`
		SELECT m.message_id, m.conversation_id, m.sender_id, u.Username, m.content, m.sent_at, m.is_read
		FROM message m
		JOIN user u ON m.sender_id = u.userid
		WHERE m.message_id = ?
	`, messageID).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.SenderName,
		&msg.Content, &sentAtStr, &msg.IsRead,
	)

	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to fetch message %d after insertion: %v", messageID, err)
		return nil, err
	}
	log.Printf("[DEBUG] Fetched details for message ID %d", messageID)

	msg.SentAt, err = time.Parse(time.RFC3339, sentAtStr)
	if err != nil {
		layout := "2006-01-02 15:04:05"
		msg.SentAt, err = time.Parse(layout, sentAtStr)
		if err != nil {
			log.Printf("[WARN] Failed to parse timestamp '%s' for new message %d: %v", sentAtStr, messageID, err)
			msg.SentAt = time.Time{}
		}
	}
	log.Printf("[DEBUG] Parsed timestamp for message ID %d: %v", messageID, msg.SentAt)

	if err := tx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction for message %d: %v", messageID, err)
		return nil, err
	}
	log.Printf("[DEBUG] Committed transaction for message ID %d", messageID)

	log.Printf("[INFO] Added message %d from user %d to conversation %d: '%s'", messageID, senderID, conversationID, truncateContent(content))
	return &msg, nil
}

func GetConversationsWithIDs(db *sql.DB, conversationIDs []int) ([]Conversation, error) {
	if len(conversationIDs) == 0 {
		log.Printf("[INFO] GetConversationsWithIDs called with empty ID list")
		return []Conversation{}, nil
	}

	query := "SELECT conversation_id, created_at FROM conversation WHERE conversation_id IN ("
	placeholders := make([]string, len(conversationIDs))
	args := make([]interface{}, len(conversationIDs))
	for i, id := range conversationIDs {
		placeholders[i] = "?"
		args[i] = id
	}
	query += fmt.Sprintf("%s)", strings.Join(placeholders, ","))

	log.Printf("[DEBUG] Querying conversations with IDs: %v", conversationIDs)
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("[ERROR] Failed to query conversations with IDs %v: %v", conversationIDs, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("[DEBUG] Successfully queried conversations with IDs: %v", conversationIDs)

	conversations := []Conversation{}
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.CreatedAt)
		if err != nil {
			log.Printf("[ERROR] Failed to scan conversation: %v", err)
			return nil, err
		}
		log.Printf("[DEBUG] Scanned conversation ID %d", conv.ID)

		users, err := getConversationParticipants(conv.ID, db)
		if err != nil {
			log.Printf("[ERROR] Failed to get participants for conversation %d: %v", conv.ID, err)
			return nil, err
		}
		participants := make([]*User, len(users))
		for i := range users {
			participants[i] = &users[i]
		}
		conv.Participants = participants
		log.Printf("[DEBUG] Retrieved %d participants for conversation %d", len(participants), conv.ID)

		lastMsg, err := getLastMessage(conv.ID, db)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("[ERROR] Failed to get last message for conversation %d: %v", conv.ID, err)
		} else if err == sql.ErrNoRows {
			conv.LastMessage = nil
			log.Printf("[DEBUG] No last message found for conversation %d", conv.ID)
		} else {
			conv.LastMessage = &ChatMessage{
				ID:             lastMsg.ID,
				ConversationID: lastMsg.ConversationID,
				SenderID:       lastMsg.SenderID,
				SenderName:     lastMsg.SenderName,
				Content:        lastMsg.Content,
				SentAt:         lastMsg.SentAt,
				IsRead:         lastMsg.IsRead,
			}
			log.Printf("[DEBUG] Retrieved last message ID %d for conversation %d", lastMsg.ID, conv.ID)
		}

		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating conversations with IDs %v: %v", conversationIDs, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d/%d conversations with requested IDs", len(conversations), len(conversationIDs))
	return conversations, nil
}

func CheckUserExists(db *sql.DB, userID int) (bool, error) {
	var count int

	log.Printf("[DEBUG] Checking if user %d exists", userID)
	err := db.QueryRow("SELECT COUNT(*) FROM user WHERE userid = ?", userID).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] Failed to check if user %d exists: %v", userID, err)
		return false, err
	}

	exists := count > 0
	log.Printf("[INFO] User %d exists: %v", userID, exists)
	return exists, nil
}

func GetUnreadMessageCount(db *sql.DB, conversationID, userID int) (int, error) {
	var count int

	query := `
		SELECT COUNT(*) FROM message m
		WHERE m.conversation_id = ? 
		AND m.sender_id != ? 
		AND m.is_read = 0
	`

	log.Printf("[DEBUG] Retrieving unread message count for user %d in conversation %d", userID, conversationID)
	err := db.QueryRow(query, conversationID, userID).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] Failed to get unread message count for user %d in conversation %d: %v", userID, conversationID, err)
		return 0, err
	}

	log.Printf("[INFO] User %d has %d unread messages in conversation %d", userID, count, conversationID)
	return count, nil
}

func GetConversationParticipantsDetails(db *sql.DB, conversationID int) ([]*User, error) {
	var participants []*User

	query := `
		SELECT u.userid, u.Username, u.Email, u.F_name, u.L_name, u.Avatar
		FROM user u
		JOIN conversation_participants cp ON u.userid = cp.user_id
		WHERE cp.conversation_id = ?
	`

	log.Printf("[DEBUG] Retrieving participant details for conversation %d", conversationID)
	rows, err := db.Query(query, conversationID)
	if err != nil {
		log.Printf("[ERROR] Failed to get participant details for conversation %d: %v", conversationID, err)
		return nil, err
	}
	defer rows.Close()
	log.Printf("[DEBUG] Successfully queried participant details for conversation %d", conversationID)

	for rows.Next() {
		user := &User{}
		var avatarNullable sql.NullString

		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &avatarNullable,
		); err != nil {
			log.Printf("[ERROR] Failed to scan participant details in conversation %d: %v", conversationID, err)
			return nil, err
		}

		user.Avatar = avatarNullable
		log.Printf("[DEBUG] Scanned participant ID %d details for conversation %d", user.ID, conversationID)
		participants = append(participants, user)
	}

	log.Printf("[INFO] Retrieved %d participant details for conversation %d", len(participants), conversationID)
	return participants, nil
}

func truncateContent(content string) string {
	if len(content) > 50 {
		return content[:47] + "..."
	}
	return content
}
