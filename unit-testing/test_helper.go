package unit_testing

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// TestConfig holds configuration for test execution
type TestConfig struct {
	DBPath          string
	TestDataEnabled bool
	VerboseLogging  bool
	CleanupAfter    bool
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DBPath:          "./test_database.db",
		TestDataEnabled: false, // Disabled by default to avoid conflicts with test fixtures
		VerboseLogging:  false,
		CleanupAfter:    true,
	}
}

// TestDatabase manages test database lifecycle
type TestDatabase struct {
	DB     *sql.DB
	Path   string
	Config *TestConfig
}

// NewTestDatabase creates a new test database instance
func NewTestDatabase(config *TestConfig) (*TestDatabase, error) {
	if config == nil {
		config = DefaultTestConfig()
	}

	// Ensure test database directory exists
	dir := filepath.Dir(config.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create test database directory: %v", err)
	}

	// Remove existing test database
	if err := os.RemoveAll(config.DBPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing test database: %v", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open test database: %v", err)
	}

	testDB := &TestDatabase{
		DB:     db,
		Path:   config.DBPath,
		Config: config,
	}

	// Initialize database schema
	if err := testDB.InitializeSchema(); err != nil {
		testDB.Close()
		return nil, fmt.Errorf("failed to initialize test database schema: %v", err)
	}

	// Load test data if enabled (disabled by default to avoid conflicts)
	if config.TestDataEnabled {
		if err := testDB.LoadTestData(); err != nil {
			testDB.Close()
			return nil, fmt.Errorf("failed to load test data: %v", err)
		}
	}

	return testDB, nil
}

// InitializeSchema creates all necessary database tables
func (tdb *TestDatabase) InitializeSchema() error {
	log.Printf("[TEST] Initializing test database schema")

	createTables := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			idcategories INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS user (
			userid INTEGER PRIMARY KEY AUTOINCREMENT,
			F_name TEXT NOT NULL,
			L_name TEXT NOT NULL,
			Username TEXT NOT NULL UNIQUE,
			Email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			current_session TEXT,
			Avatar TEXT,
			gender TEXT,
			date_of_birth DATE
		);`,

		`CREATE TABLE IF NOT EXISTS post (
			postid INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NULL,
			title TEXT NULL,
			post_at DATETIME NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`,

		`CREATE TABLE IF NOT EXISTS comment (
			commentid INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NULL,
			comment_at DATETIME NULL,
			post_postid INTEGER NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`,

		`CREATE TABLE IF NOT EXISTS post_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_id INTEGER NOT NULL,
			category_id INTEGER NOT NULL,
			FOREIGN KEY (post_id) REFERENCES post(postid),
			FOREIGN KEY (category_id) REFERENCES categories(idcategories)
		);`,

		`CREATE TABLE IF NOT EXISTS conversation (
			conversation_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS conversation_participants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (conversation_id) REFERENCES conversation(conversation_id),
			FOREIGN KEY (user_id) REFERENCES user(userid),
			UNIQUE(conversation_id, user_id)
		);`,

		`CREATE TABLE IF NOT EXISTS message (
			message_id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_read BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (conversation_id) REFERENCES conversation(conversation_id),
			FOREIGN KEY (sender_id) REFERENCES user(userid)
		);`,

		`CREATE TABLE IF NOT EXISTS online_status (
			user_id INTEGER PRIMARY KEY,
			status TEXT NOT NULL DEFAULT 'offline',
			last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(userid)
		);`,

		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_message_conversation ON message(conversation_id);`,
		`CREATE INDEX IF NOT EXISTS idx_message_sender ON message(sender_id);`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_participants_user ON conversation_participants(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_participants_conv ON conversation_participants(conversation_id);`,
		`CREATE INDEX IF NOT EXISTS idx_online_status_user ON online_status(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_post_user ON post(user_userid);`,
		`CREATE INDEX IF NOT EXISTS idx_comment_post ON comment(post_postid);`,
		`CREATE INDEX IF NOT EXISTS idx_comment_user ON comment(user_userid);`,
	}

	for _, query := range createTables {
		if _, err := tdb.DB.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %v", err)
		}
	}

	log.Printf("[TEST] Test database schema initialized successfully")
	return nil
}

// LoadTestData loads test data into the database
func (tdb *TestDatabase) LoadTestData() error {
	log.Printf("[TEST] Loading test data into test database")

	// Insert default categories
	categories := []string{"General", "Technology", "Sports", "Entertainment", "Science", "Politics"}
	for _, category := range categories {
		_, err := tdb.DB.Exec("INSERT INTO categories (name) VALUES (?)", category)
		if err != nil {
			return fmt.Errorf("failed to insert category %s: %v", category, err)
		}
	}

	// Insert test users
	testUsers := []struct {
		FirstName, LastName, Username, Email, Gender, DateOfBirth, Password string
	}{
		{"John", "Doe", "johndoe", "john@example.com", "male", "1990-01-01", "$2a$10$hashedpassword1"},
		{"Jane", "Smith", "janesmith", "jane@example.com", "female", "1992-05-15", "$2a$10$hashedpassword2"},
		{"Bob", "Johnson", "bobjohnson", "bob@example.com", "male", "1988-12-10", "$2a$10$hashedpassword3"},
		{"Alice", "Brown", "alicebrown", "alice@example.com", "female", "1995-03-20", "$2a$10$hashedpassword4"},
		{"Charlie", "Wilson", "charliewilson", "charlie@example.com", "male", "1987-08-05", "$2a$10$hashedpassword5"},
	}

	for _, user := range testUsers {
		avatar := "/static/assets/default-avatar.png"
		if user.Gender == "male" {
			avatar = "/static/assets/male-avatar-boy-face-man-user-7.svg"
		} else if user.Gender == "female" {
			avatar = "/static/assets/female-avatar-girl-face-woman-user-9.svg"
		}

		_, err := tdb.DB.Exec(`
			INSERT INTO user (F_name, L_name, Username, Email, gender, date_of_birth, password, Avatar)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			user.FirstName, user.LastName, user.Username, user.Email,
			user.Gender, user.DateOfBirth, user.Password, avatar)
		if err != nil {
			return fmt.Errorf("failed to insert user %s: %v", user.Username, err)
		}
	}

	// Insert test posts
	testPosts := []struct {
		Title, Content string
		UserID         int
	}{
		{"Welcome to the Forum", "This is the first post on our forum!", 1},
		{"Technology Discussion", "Let's talk about the latest tech trends", 2},
		{"Sports Update", "Latest sports news and updates", 3},
		{"Entertainment News", "What's happening in entertainment", 4},
		{"Science Discoveries", "Amazing scientific breakthroughs", 5},
	}

	for _, post := range testPosts {
		result, err := tdb.DB.Exec(`
			INSERT INTO post (title, content, post_at, user_userid)
			VALUES (?, ?, ?, ?)`,
			post.Title, post.Content, time.Now(), post.UserID)
		if err != nil {
			return fmt.Errorf("failed to insert post %s: %v", post.Title, err)
		}

		// Get the post ID and assign it to a category
		postID, _ := result.LastInsertId()
		_, err = tdb.DB.Exec(`
			INSERT INTO post_categories (post_id, category_id)
			VALUES (?, ?)`, postID, 1) // Assign to General category
		if err != nil {
			return fmt.Errorf("failed to assign category to post: %v", err)
		}
	}

	// Insert test comments
	testComments := []struct {
		Content string
		PostID  int
		UserID  int
	}{
		{"Great post! Thanks for sharing.", 1, 2},
		{"I agree with your points.", 1, 3},
		{"Very interesting topic.", 2, 1},
		{"Looking forward to more updates.", 3, 4},
		{"Excellent analysis!", 4, 5},
	}

	for _, comment := range testComments {
		_, err := tdb.DB.Exec(`
			INSERT INTO comment (content, comment_at, post_postid, user_userid)
			VALUES (?, ?, ?, ?)`,
			comment.Content, time.Now(), comment.PostID, comment.UserID)
		if err != nil {
			return fmt.Errorf("failed to insert comment: %v", err)
		}
	}

	log.Printf("[TEST] Test data loaded successfully")
	return nil
}

// Close closes the test database connection and optionally cleans up files
func (tdb *TestDatabase) Close() error {
	if tdb.DB != nil {
		if err := tdb.DB.Close(); err != nil {
			log.Printf("[TEST] Error closing test database: %v", err)
		}
	}

	if tdb.Config.CleanupAfter {
		if err := os.RemoveAll(tdb.Path); err != nil && !os.IsNotExist(err) {
			log.Printf("[TEST] Error removing test database file: %v", err)
		}
	}

	return nil
}

// TestSetup provides common test setup functionality
func TestSetup(t *testing.T) *TestDatabase {
	config := DefaultTestConfig()
	config.DBPath = fmt.Sprintf("./test_db_%d.db", time.Now().UnixNano())

	testDB, err := NewTestDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Cleanup function to be called at the end of tests
	t.Cleanup(func() {
		testDB.Close()
	})

	return testDB
}

// AssertNoError is a helper function to check for errors in tests
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError is a helper function to check that an error occurred
func AssertError(t *testing.T, err error, message string) {
	if err == nil {
		t.Fatalf("%s: expected error but got none", message)
	}
}

// AssertEqual is a helper function to check equality
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual is a helper function to check inequality
func AssertNotEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected == actual {
		t.Fatalf("%s: expected %v to not equal %v", message, expected, actual)
	}
}

// AssertTrue is a helper function to check boolean true
func AssertTrue(t *testing.T, condition bool, message string) {
	if !condition {
		t.Fatalf("%s: expected true but got false", message)
	}
}

// AssertFalse is a helper function to check boolean false
func AssertFalse(t *testing.T, condition bool, message string) {
	if condition {
		t.Fatalf("%s: expected false but got true", message)
	}
}

// AssertGreaterThan is a helper function to check if actual > expected
func AssertGreaterThan(t *testing.T, actual, expected int, message string) {
	if actual <= expected {
		t.Fatalf("%s: expected %d > %d", message, actual, expected)
	}
}

// AssertGreaterThanOrEqual is a helper function to check if actual >= expected
func AssertGreaterThanOrEqual(t *testing.T, actual, expected int, message string) {
	if actual < expected {
		t.Fatalf("%s: expected %d >= %d", message, actual, expected)
	}
}

// AssertLessThanOrEqual is a helper function to check if actual <= expected
func AssertLessThanOrEqual(t *testing.T, actual, expected int, message string) {
	if actual > expected {
		t.Fatalf("%s: expected %d <= %d", message, actual, expected)
	}
}

// Cleanup is an alias for Close to maintain compatibility with existing tests
func (tdb *TestDatabase) Cleanup() error {
	return tdb.Close()
}

// CreateTestSession creates a test session for a user
func CreateTestSession(t *testing.T, db *TestDatabase, userID int) string {
	sessionToken := fmt.Sprintf("test_session_%d_%d", userID, time.Now().UnixNano())

	_, err := db.DB.Exec("UPDATE user SET current_session = ? WHERE userid = ?", sessionToken, userID)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return sessionToken
}

// SetupTestConversations creates test conversations between users
func SetupTestConversations(db *sql.DB, userIDs []int) ([]int, error) {
	var conversationIDs []int

	// Create conversations between pairs of users
	for i := 0; i < len(userIDs)-1; i++ {
		for j := i + 1; j < len(userIDs) && len(conversationIDs) < 5; j++ {
			// Create conversation
			result, err := db.Exec("INSERT INTO conversation (created_at) VALUES (?)", time.Now())
			if err != nil {
				return nil, fmt.Errorf("failed to create conversation: %v", err)
			}

			conversationID, err := result.LastInsertId()
			if err != nil {
				return nil, fmt.Errorf("failed to get conversation ID: %v", err)
			}

			// Add participants
			_, err = db.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, userIDs[i])
			if err != nil {
				return nil, fmt.Errorf("failed to add participant: %v", err)
			}

			_, err = db.Exec("INSERT INTO conversation_participants (conversation_id, user_id) VALUES (?, ?)", conversationID, userIDs[j])
			if err != nil {
				return nil, fmt.Errorf("failed to add participant: %v", err)
			}

			conversationIDs = append(conversationIDs, int(conversationID))
		}
	}

	return conversationIDs, nil
}

// SetupTestMessages creates test messages in conversations
func SetupTestMessages(db *sql.DB, conversationIDs, userIDs []int) ([]int, error) {
	var messageIDs []int

	for i, conversationID := range conversationIDs {
		// Create 2-3 messages per conversation
		for j := 0; j < 3; j++ {
			userIndex := (i + j) % len(userIDs)
			content := fmt.Sprintf("Test message %d in conversation %d", j+1, conversationID)

			result, err := db.Exec(`
				INSERT INTO message (conversation_id, sender_id, content, sent_at, is_read)
				VALUES (?, ?, ?, ?, ?)
			`, conversationID, userIDs[userIndex], content, time.Now(), false)

			if err != nil {
				return nil, fmt.Errorf("failed to create message: %v", err)
			}

			messageID, err := result.LastInsertId()
			if err != nil {
				return nil, fmt.Errorf("failed to get message ID: %v", err)
			}

			messageIDs = append(messageIDs, int(messageID))
		}
	}

	return messageIDs, nil
}
