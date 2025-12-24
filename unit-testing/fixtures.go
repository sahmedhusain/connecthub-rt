package unit_testing

import (
	"database/sql"
	"fmt"
	"time"

	"connecthub/database"
)

// TestUser represents a test user fixture
type TestUser struct {
	ID           int
	FirstName    string
	LastName     string
	Username     string
	Email        string
	Password     string
	Gender       string
	DateOfBirth  string
	Avatar       string
	SessionToken string
}

// TestPost represents a test post fixture
type TestPost struct {
	ID         int
	Title      string
	Content    string
	UserID     int
	PostAt     time.Time
	Categories []string
}

// TestComment represents a test comment fixture
type TestComment struct {
	ID        int
	Content   string
	PostID    int
	UserID    int
	CommentAt time.Time
}

// TestConversation represents a test conversation fixture
type TestConversation struct {
	ID           int
	Participants []int
	CreatedAt    time.Time
}

// TestMessage represents a test message fixture
type TestMessage struct {
	ID             int
	ConversationID int
	SenderID       int
	Content        string
	SentAt         time.Time
	IsRead         bool
}

// UserFixtures provides predefined test users
var UserFixtures = []TestUser{
	{
		FirstName:   "John",
		LastName:    "Doe",
		Username:    "johndoe",
		Email:       "john@example.com",
		Password:    "password123",
		Gender:      "male",
		DateOfBirth: "1990-01-01",
		Avatar:      "/static/assets/male-avatar-boy-face-man-user-7.svg",
	},
	{
		FirstName:   "Jane",
		LastName:    "Smith",
		Username:    "janesmith",
		Email:       "jane@example.com",
		Password:    "password123",
		Gender:      "female",
		DateOfBirth: "1992-05-15",
		Avatar:      "/static/assets/female-avatar-girl-face-woman-user-9.svg",
	},
	{
		FirstName:   "Bob",
		LastName:    "Johnson",
		Username:    "bobjohnson",
		Email:       "bob@example.com",
		Password:    "password123",
		Gender:      "male",
		DateOfBirth: "1988-12-10",
		Avatar:      "/static/assets/male-avatar-boy-face-man-user-7.svg",
	},
	{
		FirstName:   "Alice",
		LastName:    "Brown",
		Username:    "alicebrown",
		Email:       "alice@example.com",
		Password:    "password123",
		Gender:      "female",
		DateOfBirth: "1995-03-20",
		Avatar:      "/static/assets/female-avatar-girl-face-woman-user-9.svg",
	},
	{
		FirstName:   "Charlie",
		LastName:    "Wilson",
		Username:    "charliewilson",
		Email:       "charlie@example.com",
		Password:    "password123",
		Gender:      "male",
		DateOfBirth: "1987-08-05",
		Avatar:      "/static/assets/male-avatar-boy-face-man-user-7.svg",
	},
}

// PostFixtures provides predefined test posts
var PostFixtures = []TestPost{
	{
		Title:      "Welcome to the Forum",
		Content:    "This is the first post on our forum! Welcome everyone!",
		UserID:     1,
		Categories: []string{"General"},
	},
	{
		Title:      "Technology Discussion",
		Content:    "Let's talk about the latest tech trends and innovations.",
		UserID:     2,
		Categories: []string{"Technology"},
	},
	{
		Title:      "Sports Update",
		Content:    "Latest sports news and updates from around the world.",
		UserID:     3,
		Categories: []string{"Sports"},
	},
	{
		Title:      "Entertainment News",
		Content:    "What's happening in the world of entertainment?",
		UserID:     4,
		Categories: []string{"Entertainment"},
	},
	{
		Title:      "Science Discoveries",
		Content:    "Amazing scientific breakthroughs and discoveries.",
		UserID:     5,
		Categories: []string{"Science"},
	},
}

// CommentFixtures provides predefined test comments
var CommentFixtures = []TestComment{
	{
		Content: "Great post! Thanks for sharing this information.",
		PostID:  1,
		UserID:  2,
	},
	{
		Content: "I completely agree with your points here.",
		PostID:  1,
		UserID:  3,
	},
	{
		Content: "Very interesting topic, looking forward to more discussions.",
		PostID:  2,
		UserID:  1,
	},
	{
		Content: "Thanks for the update! This is very helpful.",
		PostID:  3,
		UserID:  4,
	},
	{
		Content: "Excellent analysis and well-written post.",
		PostID:  4,
		UserID:  5,
	},
}

// CreateTestUser creates a test user in the database and returns the user ID
func CreateTestUser(db *sql.DB, user TestUser) (int, error) {
	// Use the database.CreateUser function which handles password hashing internally
	userID, err := database.CreateUser(db, user.FirstName, user.LastName, user.Username,
		user.Email, user.Gender, user.DateOfBirth, user.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to create test user: %v", err)
	}

	return userID, nil
}

// CreateTestPost creates a test post in the database and returns the post ID
func CreateTestPost(db *sql.DB, post TestPost) (int, error) {
	query := `
		INSERT INTO post (title, content, post_at, user_userid)
		VALUES (?, ?, ?, ?)
	`

	postAt := post.PostAt
	if postAt.IsZero() {
		postAt = time.Now()
	}

	result, err := db.Exec(query, post.Title, post.Content, postAt, post.UserID)
	if err != nil {
		return 0, fmt.Errorf("failed to create test post: %v", err)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get post ID: %v", err)
	}

	// Add categories if specified
	for _, categoryName := range post.Categories {
		// Get or create category
		var categoryID int
		err := db.QueryRow("SELECT idcategories FROM categories WHERE name = ?", categoryName).Scan(&categoryID)
		if err == sql.ErrNoRows {
			// Create category
			result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
			if err != nil {
				return 0, fmt.Errorf("failed to create category: %v", err)
			}
			categoryID64, _ := result.LastInsertId()
			categoryID = int(categoryID64)
		} else if err != nil {
			return 0, fmt.Errorf("failed to query category: %v", err)
		}

		// Link post to category
		_, err = db.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, categoryID)
		if err != nil {
			return 0, fmt.Errorf("failed to link post to category: %v", err)
		}
	}

	return int(postID), nil
}

// CreateTestComment creates a test comment in the database and returns the comment ID
func CreateTestComment(db *sql.DB, comment TestComment) (int, error) {
	query := `
		INSERT INTO comment (content, comment_at, post_postid, user_userid)
		VALUES (?, ?, ?, ?)
	`

	commentAt := comment.CommentAt
	if commentAt.IsZero() {
		commentAt = time.Now()
	}

	result, err := db.Exec(query, comment.Content, commentAt, comment.PostID, comment.UserID)
	if err != nil {
		return 0, fmt.Errorf("failed to create test comment: %v", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get comment ID: %v", err)
	}

	return int(commentID), nil
}

// CreateTestConversation creates a test conversation with participants
func CreateTestConversation(db *sql.DB, participants []int) (int, error) {
	// Create conversation
	result, err := db.Exec("INSERT INTO conversation (created_at) VALUES (?)", time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %v", err)
	}

	conversationID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get conversation ID: %v", err)
	}

	// Add participants
	for _, userID := range participants {
		_, err := db.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)",
			conversationID, userID)
		if err != nil {
			return 0, fmt.Errorf("failed to add participant %d: %v", userID, err)
		}
	}

	return int(conversationID), nil
}

// CreateTestMessage creates a test message in the database
func CreateTestMessage(db *sql.DB, message TestMessage) (int, error) {
	query := `
		INSERT INTO message (conversation_id, sender_id, content, sent_at, is_read)
		VALUES (?, ?, ?, ?, ?)
	`

	sentAt := message.SentAt
	if sentAt.IsZero() {
		sentAt = time.Now()
	}

	result, err := db.Exec(query, message.ConversationID, message.SenderID,
		message.Content, sentAt, message.IsRead)
	if err != nil {
		return 0, fmt.Errorf("failed to create test message: %v", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get message ID: %v", err)
	}

	return int(messageID), nil
}

// SetupTestUsers creates all test users and returns their IDs
func SetupTestUsers(db *sql.DB) ([]int, error) {
	var userIDs []int

	for _, user := range UserFixtures {
		userID, err := CreateTestUser(db, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create test user %s: %v", user.Username, err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// SetupTestPosts creates all test posts and returns their IDs
func SetupTestPosts(db *sql.DB, userIDs []int) ([]int, error) {
	var postIDs []int

	for i, post := range PostFixtures {
		if i < len(userIDs) {
			post.UserID = userIDs[i]
		}

		postID, err := CreateTestPost(db, post)
		if err != nil {
			return nil, fmt.Errorf("failed to create test post %s: %v", post.Title, err)
		}
		postIDs = append(postIDs, postID)
	}

	return postIDs, nil
}

// SetupTestComments creates all test comments and returns their IDs
func SetupTestComments(db *sql.DB, postIDs, userIDs []int) ([]int, error) {
	var commentIDs []int

	for i, comment := range CommentFixtures {
		if i < len(postIDs) {
			comment.PostID = postIDs[i%len(postIDs)]
		}
		if i < len(userIDs) {
			comment.UserID = userIDs[(i+1)%len(userIDs)] // Different user than post author
		}

		commentID, err := CreateTestComment(db, comment)
		if err != nil {
			return nil, fmt.Errorf("failed to create test comment: %v", err)
		}
		commentIDs = append(commentIDs, commentID)
	}

	return commentIDs, nil
}

// SetupCompleteTestData creates a complete set of test data
func SetupCompleteTestData(db *sql.DB) (userIDs, postIDs, commentIDs []int, err error) {
	// Create users
	userIDs, err = SetupTestUsers(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup test users: %v", err)
	}

	// Create posts
	postIDs, err = SetupTestPosts(db, userIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup test posts: %v", err)
	}

	// Create comments
	commentIDs, err = SetupTestComments(db, postIDs, userIDs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup test comments: %v", err)
	}

	return userIDs, postIDs, commentIDs, nil
}
