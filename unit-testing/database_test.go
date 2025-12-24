package unit_testing

import (
	"database/sql"
	"fmt"
	"testing"

	"connecthub/database"
	"connecthub/repository"
)

func TestDatabaseConnection(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("DatabaseConnectionValid", func(t *testing.T) {
		// Test that database connection is valid
		err := testDB.DB.Ping()
		AssertNoError(t, err, "Database connection should be valid")
	})

	t.Run("DatabaseTablesExist", func(t *testing.T) {
		// Check that all required tables exist
		tables := []string{
			"user", "post", "comment", "categories", "post_categories",
			"conversation", "conversation_participants", "message", "online_status",
		}

		for _, table := range tables {
			var count int
			query := "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?"
			err := testDB.DB.QueryRow(query, table).Scan(&count)
			AssertNoError(t, err, "Should be able to query table existence")
			AssertEqual(t, count, 1, fmt.Sprintf("Table %s should exist", table))
		}
	})

	t.Run("DatabaseIndexesExist", func(t *testing.T) {
		// Check that important indexes exist
		indexes := []string{
			"idx_message_conversation", "idx_message_sender",
			"idx_conversation_participants_user", "idx_conversation_participants_conv",
			"idx_online_status_user", "idx_post_user",
			"idx_comment_post", "idx_comment_user",
		}

		for _, index := range indexes {
			var count int
			query := "SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?"
			err := testDB.DB.QueryRow(query, index).Scan(&count)
			AssertNoError(t, err, "Should be able to query index existence")
			AssertEqual(t, count, 1, fmt.Sprintf("Index %s should exist", index))
		}
	})
}

func TestUserDatabaseOperations(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("CreateUser", func(t *testing.T) {
		// Test user creation
		userID, err := database.CreateUser(testDB.DB, "Test", "User", "testuser", "test@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")
		AssertTrue(t, userID > 0, "User ID should be positive")

		// Verify user was created
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM user WHERE userid = ?", userID).Scan(&count)
		AssertNoError(t, err, "Should be able to query user")
		AssertEqual(t, count, 1, "User should exist in database")
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Create user first
		userID, err := database.CreateUser(testDB.DB, "Get", "User", "getuser", "get@example.com",
			"female", "1992-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Get user by ID
		user, err := database.GetUserByID(testDB.DB, userID)
		AssertNoError(t, err, "Should be able to get user by ID")
		AssertEqual(t, user.ID, userID, "User ID should match")
		AssertEqual(t, user.Username, "getuser", "Username should match")
		AssertEqual(t, user.Email, "get@example.com", "Email should match")
		AssertEqual(t, user.FirstName, "Get", "First name should match")
		AssertEqual(t, user.LastName, "User", "Last name should match")
	})

	t.Run("AuthenticateUser", func(t *testing.T) {
		// Create user first
		_, err := database.CreateUser(testDB.DB, "Auth", "User", "authuser", "auth@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Authenticate with username
		user, err := database.AuthenticateUser(testDB.DB, "authuser", "password123")
		AssertNoError(t, err, "Authentication with username should succeed")
		AssertEqual(t, user.Username, "authuser", "Username should match")

		// Authenticate with email
		user, err = database.AuthenticateUser(testDB.DB, "auth@example.com", "password123")
		AssertNoError(t, err, "Authentication with email should succeed")
		AssertEqual(t, user.Email, "auth@example.com", "Email should match")

		// Test wrong password
		_, err = database.AuthenticateUser(testDB.DB, "authuser", "wrongpassword")
		AssertError(t, err, "Authentication with wrong password should fail")
	})

	t.Run("UpdateUserSession", func(t *testing.T) {
		// Create user first
		userID, err := database.CreateUser(testDB.DB, "Session", "User", "sessionuser", "session@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Update session
		sessionToken := "test_session_token_123"
		err = database.UpdateUserSession(testDB.DB, userID, sessionToken)
		AssertNoError(t, err, "Session update should succeed")

		// Verify session was updated
		var storedSession sql.NullString
		err = testDB.DB.QueryRow("SELECT current_session FROM user WHERE userid = ?", userID).Scan(&storedSession)
		AssertNoError(t, err, "Should be able to query session")
		AssertTrue(t, storedSession.Valid, "Session should be valid")
		AssertEqual(t, storedSession.String, sessionToken, "Session token should match")
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 3; i++ {
			_, err := database.CreateUser(testDB.DB, "User", fmt.Sprintf("Test%d", i),
				fmt.Sprintf("user%d", i), fmt.Sprintf("user%d@example.com", i),
				"male", "1990-01-01", "password123")
			AssertNoError(t, err, "User creation should succeed")
		}

		// Get all users
		users, err := database.GetAllUsers(testDB.DB)
		AssertNoError(t, err, "Should be able to get all users")
		AssertTrue(t, len(users) >= 3, "Should have at least 3 users")

		// Verify user data structure
		for _, user := range users {
			AssertTrue(t, user.ID > 0, "User ID should be positive")
			AssertNotEqual(t, user.Username, "", "Username should not be empty")
			AssertNotEqual(t, user.Email, "", "Email should not be empty")
		}
	})
}

func TestPostDatabaseOperations(t *testing.T) {
	testDB := TestSetup(t)

	// Create a test user first
	userID, err := database.CreateUser(testDB.DB, "Post", "User", "postuser", "post@example.com",
		"male", "1990-01-01", "password123")
	AssertNoError(t, err, "User creation should succeed")

	t.Run("CreatePost", func(t *testing.T) {
		// Create post
		postID, err := database.CreatePost(testDB.DB, userID, "Test Post", "This is a test post content",
			[]string{"General", "Technology"})
		AssertNoError(t, err, "Post creation should succeed")
		AssertTrue(t, postID > 0, "Post ID should be positive")

		// Verify post was created
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM post WHERE postid = ?", postID).Scan(&count)
		AssertNoError(t, err, "Should be able to query post")
		AssertEqual(t, count, 1, "Post should exist in database")
	})

	t.Run("GetPostByID", func(t *testing.T) {
		// Create post first
		postID, err := database.CreatePost(testDB.DB, userID, "Get Post", "Content for get test",
			[]string{"General"})
		AssertNoError(t, err, "Post creation should succeed")

		// Get post by ID
		post, err := database.GetPostByID(testDB.DB, postID)
		AssertNoError(t, err, "Should be able to get post by ID")
		AssertEqual(t, post.PostID, postID, "Post ID should match")
		AssertEqual(t, post.Title, "Get Post", "Post title should match")
		AssertEqual(t, post.Content, "Content for get test", "Post content should match")
		AssertEqual(t, post.UserUserID, userID, "Post user ID should match")
	})

	t.Run("GetAllPosts", func(t *testing.T) {
		// Create multiple posts
		for i := 0; i < 3; i++ {
			_, err := database.CreatePost(testDB.DB, userID, fmt.Sprintf("Post %d", i),
				fmt.Sprintf("Content %d", i), []string{"General"})
			AssertNoError(t, err, "Post creation should succeed")
		}

		// Get all posts
		posts, err := database.GetAllPosts(testDB.DB)
		AssertNoError(t, err, "Should be able to get all posts")
		AssertTrue(t, len(posts) >= 3, "Should have at least 3 posts")

		// Verify posts are ordered by date (newest first)
		if len(posts) > 1 {
			for i := 0; i < len(posts)-1; i++ {
				AssertTrue(t, posts[i].PostAt.After(posts[i+1].PostAt) || posts[i].PostAt.Equal(posts[i+1].PostAt),
					"Posts should be ordered by date (newest first)")
			}
		}
	})

	t.Run("GetFilteredPosts", func(t *testing.T) {
		// Test different filters
		filters := []string{"all", "top-rated", "oldest"}

		for _, filter := range filters {
			posts, err := database.GetFilteredPosts(testDB.DB, filter)
			AssertNoError(t, err, fmt.Sprintf("Filter %s should work", filter))

			if filter == "oldest" && len(posts) > 1 {
				// Verify oldest filter ordering
				for i := 0; i < len(posts)-1; i++ {
					AssertTrue(t, posts[i].PostAt.Before(posts[i+1].PostAt) || posts[i].PostAt.Equal(posts[i+1].PostAt),
						"Oldest filter should order posts by date (oldest first)")
				}
			}
		}
	})

	t.Run("AddComment", func(t *testing.T) {
		// Create post first
		postID, err := database.CreatePost(testDB.DB, userID, "Comment Post", "Post for comment test",
			[]string{"General"})
		AssertNoError(t, err, "Post creation should succeed")

		// Add comment
		err = database.AddComment(testDB.DB, postID, userID, "This is a test comment")
		AssertNoError(t, err, "Comment creation should succeed")

		// Verify comment was added
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM comment WHERE post_postid = ?", postID).Scan(&count)
		AssertNoError(t, err, "Should be able to query comment")
		AssertEqual(t, count, 1, "Comment should exist in database")
	})

	t.Run("GetCommentsForPost", func(t *testing.T) {
		// Create post first
		postID, err := database.CreatePost(testDB.DB, userID, "Comments Post", "Post for comments test",
			[]string{"General"})
		AssertNoError(t, err, "Post creation should succeed")

		// Add multiple comments
		commentContents := []string{"First comment", "Second comment", "Third comment"}
		for _, content := range commentContents {
			err = database.AddComment(testDB.DB, postID, userID, content)
			AssertNoError(t, err, "Comment creation should succeed")
		}

		// Get comments
		comments, err := database.GetCommentsForPost(testDB.DB, postID)
		AssertNoError(t, err, "Should be able to get comments")
		AssertTrue(t, len(comments) >= 3, "Should have at least 3 comments")

		// Verify comment data structure
		for _, comment := range comments {
			AssertTrue(t, comment.ID > 0, "Comment ID should be positive")
			AssertEqual(t, comment.PostID, postID, "Comment post ID should match")
			AssertEqual(t, comment.UserID, userID, "Comment user ID should match")
			AssertNotEqual(t, comment.Content, "", "Comment content should not be empty")
			AssertNotEqual(t, comment.Username, "", "Comment should include username")
		}
	})
}

func TestConversationDatabaseOperations(t *testing.T) {
	testDB := TestSetup(t)

	// Create test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	t.Run("CreateConversation", func(t *testing.T) {
		// Create conversation
		participants := []int{userIDs[0], userIDs[1]}
		conversationID, err := database.CreateConversation(participants)
		AssertNoError(t, err, "Conversation creation should succeed")
		AssertTrue(t, conversationID > 0, "Conversation ID should be positive")

		// Verify conversation was created
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM conversation WHERE conversation_id = ?", conversationID).Scan(&count)
		AssertNoError(t, err, "Should be able to query conversation")
		AssertEqual(t, count, 1, "Conversation should exist in database")

		// Verify participants were added
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ?", conversationID).Scan(&count)
		AssertNoError(t, err, "Should be able to query participants")
		AssertEqual(t, count, 2, "Should have 2 participants")
	})

	t.Run("AddMessageToConversation", func(t *testing.T) {
		// Create conversation first
		participants := []int{userIDs[0], userIDs[1]}
		conversationID, err := database.CreateConversation(participants)
		AssertNoError(t, err, "Conversation creation should succeed")

		// Add message
		message, err := database.AddMessageToConversation(testDB.DB, conversationID, userIDs[0], "Test message")
		AssertNoError(t, err, "Message creation should succeed")
		AssertNotEqual(t, message, nil, "Message should not be nil")
		AssertTrue(t, message.ID > 0, "Message ID should be positive")
		AssertEqual(t, message.ConversationID, conversationID, "Message conversation ID should match")
		AssertEqual(t, message.SenderID, userIDs[0], "Message sender ID should match")
		AssertEqual(t, message.Content, "Test message", "Message content should match")
	})

	t.Run("GetConversationMessages", func(t *testing.T) {
		// Create conversation and messages
		participants := []int{userIDs[0], userIDs[1]}
		conversationID, err := database.CreateConversation(participants)
		AssertNoError(t, err, "Conversation creation should succeed")

		// Add multiple messages
		messageContents := []string{"Message 1", "Message 2", "Message 3"}
		for i, content := range messageContents {
			_, err = database.AddMessageToConversation(testDB.DB, conversationID, userIDs[i%2], content)
			AssertNoError(t, err, "Message creation should succeed")
		}

		// Get messages
		messages, err := database.GetConversationMessages(testDB.DB, conversationID, 10, 0)
		AssertNoError(t, err, "Should be able to get messages")
		AssertTrue(t, len(messages) >= 3, "Should have at least 3 messages")

		// Verify message data structure
		for _, message := range messages {
			AssertTrue(t, message.ID > 0, "Message ID should be positive")
			AssertEqual(t, message.ConversationID, conversationID, "Message conversation ID should match")
			AssertTrue(t, message.SenderID > 0, "Message sender ID should be positive")
			AssertNotEqual(t, message.Content, "", "Message content should not be empty")
		}
	})

	t.Run("GetUserConversations", func(t *testing.T) {
		// Create multiple conversations for a user
		conv1Participants := []int{userIDs[0], userIDs[1]}
		conv1ID, err := database.CreateConversation(conv1Participants)
		AssertNoError(t, err, "First conversation creation should succeed")

		conv2Participants := []int{userIDs[0], userIDs[2]}
		conv2ID, err := database.CreateConversation(conv2Participants)
		AssertNoError(t, err, "Second conversation creation should succeed")

		// Get user conversations
		conversations, err := database.GetUserConversations(testDB.DB, userIDs[0])
		AssertNoError(t, err, "Should be able to get user conversations")
		AssertTrue(t, len(conversations) >= 2, "User should have at least 2 conversations")

		// Verify conversation IDs
		conversationIDs := make(map[int]bool)
		for _, conv := range conversations {
			conversationIDs[conv.ID] = true
		}
		AssertTrue(t, conversationIDs[conv1ID], "Should include first conversation")
		AssertTrue(t, conversationIDs[conv2ID], "Should include second conversation")
	})
}

func TestRepositoryPattern(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("UserRepositoryInterface", func(t *testing.T) {
		// Test that UserRepository implements the interface correctly
		userRepo := repository.NewUserRepository(testDB.DB)

		// Test interface methods exist and work
		userID, err := userRepo.CreateUser("Repo", "Test", "repotest", "repo@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "Repository CreateUser should work")
		AssertTrue(t, userID > 0, "User ID should be positive")

		// Test GetUserByID
		user, err := userRepo.GetUserByID(userID)
		AssertNoError(t, err, "Repository GetUserByID should work")
		AssertEqual(t, user.ID, userID, "User ID should match")

		// Test AuthenticateUser
		authUser, err := userRepo.AuthenticateUser("repotest", "password123")
		AssertNoError(t, err, "Repository AuthenticateUser should work")
		AssertEqual(t, authUser.ID, userID, "Authenticated user ID should match")

		// Test session management
		sessionToken := "test_session_repo"
		err = userRepo.UpdateUserSession(userID, sessionToken)
		AssertNoError(t, err, "Repository UpdateUserSession should work")

		sessionUser, err := userRepo.GetUserBySession(sessionToken)
		AssertNoError(t, err, "Repository GetUserBySession should work")
		AssertEqual(t, sessionUser.ID, userID, "Session user ID should match")

		validatedUserID, err := userRepo.ValidateSession(sessionToken)
		AssertNoError(t, err, "Repository ValidateSession should work")
		AssertEqual(t, validatedUserID, userID, "Validated user ID should match")

		// Test UserExists
		exists, err := userRepo.UserExists("repotest", "")
		AssertNoError(t, err, "Repository UserExists should work")
		AssertTrue(t, exists, "User should exist")

		// Test GetAllUsers
		users, err := userRepo.GetAllUsers()
		AssertNoError(t, err, "Repository GetAllUsers should work")
		AssertTrue(t, len(users) >= 1, "Should have at least 1 user")
	})

	t.Run("PostRepositoryInterface", func(t *testing.T) {
		// Create test user first
		userRepo := repository.NewUserRepository(testDB.DB)
		userID, err := userRepo.CreateUser("Post", "Repo", "postrepo", "postrepo@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Test PostRepository interface
		postRepo := repository.NewPostRepository(testDB.DB)

		// Test CreatePost
		postID, err := postRepo.CreatePost(userID, "Repo Post", "Repository test post",
			[]string{"General"})
		AssertNoError(t, err, "Repository CreatePost should work")
		AssertTrue(t, postID > 0, "Post ID should be positive")

		// Test GetPostByID
		post, err := postRepo.GetPostByID(postID)
		AssertNoError(t, err, "Repository GetPostByID should work")
		AssertEqual(t, post.PostID, postID, "Post ID should match")

		// Test GetAllPosts
		posts, err := postRepo.GetAllPosts()
		AssertNoError(t, err, "Repository GetAllPosts should work")
		AssertTrue(t, len(posts) >= 1, "Should have at least 1 post")

		// Test GetFilteredPosts
		filteredPosts, err := postRepo.GetFilteredPosts("all")
		AssertNoError(t, err, "Repository GetFilteredPosts should work")
		AssertTrue(t, len(filteredPosts) >= 1, "Should have at least 1 filtered post")

		// Test AddComment
		err = postRepo.AddComment(postID, userID, "Repository comment test")
		AssertNoError(t, err, "Repository AddComment should work")

		// Test GetCommentsForPost
		comments, err := postRepo.GetCommentsForPost(postID)
		AssertNoError(t, err, "Repository GetCommentsForPost should work")
		AssertTrue(t, len(comments) >= 1, "Should have at least 1 comment")
	})

	t.Run("MessageRepositoryInterface", func(t *testing.T) {
		// Create test users first
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Test MessageRepository interface
		messageRepo := repository.NewMessageRepository(testDB.DB)

		// Test CreateConversation
		participants := []int{userIDs[0], userIDs[1]}
		conversationID, err := messageRepo.CreateConversation(participants)
		AssertNoError(t, err, "Repository CreateConversation should work")
		AssertTrue(t, conversationID > 0, "Conversation ID should be positive")

		// Test GetConversationParticipants
		conversationParticipants, err := messageRepo.GetConversationParticipants(conversationID)
		AssertNoError(t, err, "Repository GetConversationParticipants should work")
		AssertEqual(t, len(conversationParticipants), 2, "Should have 2 participants")

		// Test IsUserParticipant
		isParticipant, err := messageRepo.IsUserParticipant(conversationID, userIDs[0])
		AssertNoError(t, err, "Repository IsUserParticipant should work")
		AssertTrue(t, isParticipant, "User should be participant")

		// Test AddMessageToConversation
		message, err := messageRepo.AddMessageToConversation(conversationID, userIDs[0], "Repo message test")
		AssertNoError(t, err, "Repository AddMessageToConversation should work")
		AssertNotEqual(t, message, nil, "Message should not be nil")

		// Test GetConversationMessages
		messages, err := messageRepo.GetConversationMessages(conversationID, 10, 0)
		AssertNoError(t, err, "Repository GetConversationMessages should work")
		AssertTrue(t, len(messages) >= 1, "Should have at least 1 message")

		// Test GetUserConversations
		conversations, err := messageRepo.GetUserConversations(userIDs[0])
		AssertNoError(t, err, "Repository GetUserConversations should work")
		AssertTrue(t, len(conversations) >= 1, "Should have at least 1 conversation")

		// Test MarkMessagesAsRead
		err = messageRepo.MarkMessagesAsRead(conversationID, userIDs[1])
		AssertNoError(t, err, "Repository MarkMessagesAsRead should work")

		// Test GetUnreadMessageCount
		unreadCount, err := messageRepo.GetUnreadMessageCount(conversationID, userIDs[1])
		AssertNoError(t, err, "Repository GetUnreadMessageCount should work")
		AssertEqual(t, unreadCount, 0, "Should have 0 unread messages after marking as read")
	})
}

func TestDataIntegrity(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("ForeignKeyConstraints", func(t *testing.T) {
		// Create user
		userID, err := database.CreateUser(testDB.DB, "FK", "Test", "fktest", "fk@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Create post
		postID, err := database.CreatePost(testDB.DB, userID, "FK Post", "Foreign key test post",
			[]string{"General"})
		AssertNoError(t, err, "Post creation should succeed")

		// Try to create comment with invalid post ID
		err = database.AddComment(testDB.DB, 99999, userID, "Comment on invalid post")
		AssertError(t, err, "Comment creation should fail with invalid post ID")

		// Try to create comment with invalid user ID
		err = database.AddComment(testDB.DB, postID, 99999, "Comment from invalid user")
		AssertError(t, err, "Comment creation should fail with invalid user ID")

		// Valid comment should work
		err = database.AddComment(testDB.DB, postID, userID, "Valid comment")
		AssertNoError(t, err, "Valid comment creation should succeed")
	})

	t.Run("UniqueConstraints", func(t *testing.T) {
		// Create first user
		_, err := database.CreateUser(testDB.DB, "Unique", "Test1", "uniquetest", "unique@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "First user creation should succeed")

		// Try to create user with same username
		_, err = database.CreateUser(testDB.DB, "Unique", "Test2", "uniquetest", "unique2@example.com",
			"female", "1992-01-01", "password123")
		AssertError(t, err, "User creation should fail with duplicate username")

		// Try to create user with same email
		_, err = database.CreateUser(testDB.DB, "Unique", "Test3", "uniquetest2", "unique@example.com",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "User creation should fail with duplicate email")
	})

	t.Run("DataConsistency", func(t *testing.T) {
		// Create user and post
		userID, err := database.CreateUser(testDB.DB, "Consistency", "Test", "consistencytest", "consistency@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		postID, err := database.CreatePost(testDB.DB, userID, "Consistency Post", "Data consistency test",
			[]string{"General"})
		AssertNoError(t, err, "Post creation should succeed")

		// Add comments
		for i := 0; i < 3; i++ {
			err = database.AddComment(testDB.DB, postID, userID, fmt.Sprintf("Comment %d", i+1))
			AssertNoError(t, err, "Comment creation should succeed")
		}

		// Verify comment count in post
		post, err := database.GetPostByID(testDB.DB, postID)
		AssertNoError(t, err, "Should be able to get post")
		AssertEqual(t, post.Comments, 3, "Post should show correct comment count")

		// Verify actual comments
		comments, err := database.GetCommentsForPost(testDB.DB, postID)
		AssertNoError(t, err, "Should be able to get comments")
		AssertEqual(t, len(comments), 3, "Should have 3 actual comments")
	})

	t.Run("TransactionIntegrity", func(t *testing.T) {
		// Test that operations are atomic
		userID, err := database.CreateUser(testDB.DB, "Transaction", "Test", "transactiontest", "transaction@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Create post with categories (this involves multiple table inserts)
		postID, err := database.CreatePost(testDB.DB, userID, "Transaction Post", "Transaction test post",
			[]string{"Category1", "Category2", "Category3"})
		AssertNoError(t, err, "Post creation with categories should succeed")

		// Verify all categories were created and linked
		var categoryCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM post_categories WHERE post_id = ?", postID).Scan(&categoryCount)
		AssertNoError(t, err, "Should be able to query post categories")
		AssertEqual(t, categoryCount, 3, "Should have 3 category links")

		// Verify categories exist
		for i := 1; i <= 3; i++ {
			var count int
			err = testDB.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE name = ?", fmt.Sprintf("Category%d", i)).Scan(&count)
			AssertNoError(t, err, "Should be able to query category")
			AssertEqual(t, count, 1, fmt.Sprintf("Category%d should exist", i))
		}
	})
}
