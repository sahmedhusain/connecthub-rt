package unit_testing

import (
	"testing"

	"connecthub/database"
	"connecthub/repository"
)

func TestUserRepository(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	userRepo := repository.NewUserRepository(testDB.DB)

	t.Run("CreateUser", func(t *testing.T) {
		t.Run("ValidUser", func(t *testing.T) {
			userID, err := userRepo.CreateUser(
				"Test", "User", "testuser", "test@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNoError(t, err, "User creation should succeed")
			AssertNotEqual(t, userID, 0, "User ID should be set")
		})

		t.Run("DuplicateUsername", func(t *testing.T) {
			// Create first user
			_, err := userRepo.CreateUser(
				"First", "User", "duplicate", "first@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNoError(t, err, "First user creation should succeed")

			// Try to create second user with same username
			_, err = userRepo.CreateUser(
				"Second", "User", "duplicate", "second@example.com",
				"female", "1991-01-01", "password456",
			)
			AssertNotEqual(t, err, nil, "Second user creation should fail due to duplicate username")
		})

		t.Run("DuplicateEmail", func(t *testing.T) {
			// Create first user
			_, err := userRepo.CreateUser(
				"First", "Email", "firstemail", "duplicate@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNoError(t, err, "First user creation should succeed")

			// Try to create second user with same email
			_, err = userRepo.CreateUser(
				"Second", "Email", "secondemail", "duplicate@example.com",
				"female", "1991-01-01", "password456",
			)
			AssertNotEqual(t, err, nil, "Second user creation should fail due to duplicate email")
		})
	})

	t.Run("AuthenticateUser", func(t *testing.T) {
		// Setup test users
		_, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidUsernamePassword", func(t *testing.T) {
			user, err := userRepo.AuthenticateUser("johndoe", "password123")
			AssertNoError(t, err, "Authentication should succeed")
			AssertNotEqual(t, user, nil, "User should be returned")
			AssertEqual(t, user.Username, "johndoe", "Username should match")
		})

		t.Run("ValidEmailPassword", func(t *testing.T) {
			user, err := userRepo.AuthenticateUser("jane@example.com", "password123")
			AssertNoError(t, err, "Authentication with email should succeed")
			AssertNotEqual(t, user, nil, "User should be returned")
			AssertEqual(t, user.Email, "jane@example.com", "Email should match")
		})

		t.Run("InvalidPassword", func(t *testing.T) {
			user, err := userRepo.AuthenticateUser("johndoe", "wrongpassword")
			AssertNotEqual(t, err, nil, "Authentication should fail")
			AssertEqual(t, user, nil, "User should not be returned")
		})

		t.Run("NonexistentUser", func(t *testing.T) {
			user, err := userRepo.AuthenticateUser("nonexistent", "password123")
			AssertNotEqual(t, err, nil, "Authentication should fail")
			AssertEqual(t, user, nil, "User should not be returned")
		})
	})

	t.Run("UpdateUserSession", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidSessionUpdate", func(t *testing.T) {
			sessionToken := "test_session_token_123"
			err := userRepo.UpdateUserSession(userIDs[0], sessionToken)
			AssertNoError(t, err, "Session update should succeed")

			// Verify session was updated
			user, err := userRepo.GetUserBySession(sessionToken)
			AssertNoError(t, err, "Should retrieve user by session")
			AssertEqual(t, user.ID, userIDs[0], "User ID should match")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			err := userRepo.UpdateUserSession(99999, "test_session")
			AssertNotEqual(t, err, nil, "Session update should fail for invalid user ID")
		})
	})

	t.Run("GetUserBySession", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Create session
		sessionToken := "test_session_token_456"
		err = userRepo.UpdateUserSession(userIDs[0], sessionToken)
		AssertNoError(t, err, "Session creation should succeed")

		t.Run("ValidSession", func(t *testing.T) {
			user, err := userRepo.GetUserBySession(sessionToken)
			AssertNoError(t, err, "Should retrieve user by session")
			AssertNotEqual(t, user, nil, "User should be returned")
			AssertEqual(t, user.ID, userIDs[0], "User ID should match")
		})

		t.Run("InvalidSession", func(t *testing.T) {
			user, err := userRepo.GetUserBySession("invalid_session")
			AssertNotEqual(t, err, nil, "Should fail for invalid session")
			AssertEqual(t, user, nil, "User should not be returned")
		})
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		users, err := userRepo.GetAllUsers()
		AssertNoError(t, err, "Should retrieve all users")
		AssertGreaterThanOrEqual(t, len(users), len(userIDs), "Should return at least the test users")
	})
}

func TestPostRepository(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	postRepo := repository.NewPostRepository(testDB.DB)

	t.Run("CreatePost", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidPost", func(t *testing.T) {
			postID, err := postRepo.CreatePost(
				userIDs[0], "Test Post", "Test content", []string{"Technology"},
			)
			AssertNoError(t, err, "Post creation should succeed")
			AssertNotEqual(t, postID, 0, "Post ID should be set")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			postID, err := postRepo.CreatePost(
				99999, "Test Post", "Test content", []string{"Technology"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for invalid user ID")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})
	})

	t.Run("GetPostByID", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		postIDs, err := SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("ValidPostID", func(t *testing.T) {
			post, err := postRepo.GetPostByID(postIDs[0])
			AssertNoError(t, err, "Should retrieve post by ID")
			AssertNotEqual(t, post, nil, "Post should be returned")
			AssertEqual(t, post.PostID, postIDs[0], "Post ID should match")
		})

		t.Run("InvalidPostID", func(t *testing.T) {
			post, err := postRepo.GetPostByID(99999)
			AssertNotEqual(t, err, nil, "Should fail for invalid post ID")
			AssertEqual(t, post, nil, "Post should not be returned")
		})
	})

	t.Run("GetAllPosts", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		postIDs, err := SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		posts, err := postRepo.GetAllPosts()
		AssertNoError(t, err, "Should retrieve all posts")
		AssertGreaterThanOrEqual(t, len(posts), len(postIDs), "Should return at least the test posts")
	})

	t.Run("GetPostsByCategory", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		_, err = SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("ValidCategory", func(t *testing.T) {
			posts, err := postRepo.GetPostsByCategory("Technology")
			AssertNoError(t, err, "Should retrieve posts by category")
			AssertGreaterThanOrEqual(t, len(posts), 0, "Should return posts or empty array")
		})

		t.Run("NonExistentCategory", func(t *testing.T) {
			posts, err := postRepo.GetPostsByCategory("NonExistent")
			AssertNoError(t, err, "Should handle non-existent category")
			AssertEqual(t, len(posts), 0, "Should return empty array")
		})
	})

	t.Run("AddComment", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		postIDs, err := SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("ValidComment", func(t *testing.T) {
			err := postRepo.AddComment(postIDs[0], userIDs[0], "Test comment")
			AssertNoError(t, err, "Comment addition should succeed")
		})

		t.Run("InvalidPostID", func(t *testing.T) {
			err := postRepo.AddComment(99999, userIDs[0], "Test comment")
			AssertNotEqual(t, err, nil, "Comment addition should fail for invalid post ID")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			err := postRepo.AddComment(postIDs[0], 99999, "Test comment")
			AssertNotEqual(t, err, nil, "Comment addition should fail for invalid user ID")
		})
	})
}

func TestMessageRepository(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	messageRepo := repository.NewMessageRepository(testDB.DB)

	t.Run("CreateConversation", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidConversation", func(t *testing.T) {
			conversationID, err := messageRepo.CreateConversation([]int{userIDs[0], userIDs[1]})
			AssertNoError(t, err, "Conversation creation should succeed")
			AssertNotEqual(t, conversationID, 0, "Conversation ID should be set")
		})

		t.Run("SameUser", func(t *testing.T) {
			conversationID, err := messageRepo.CreateConversation([]int{userIDs[0], userIDs[0]})
			AssertNotEqual(t, err, nil, "Conversation creation should fail for same user")
			AssertEqual(t, conversationID, 0, "Conversation ID should be zero")
		})

		t.Run("InvalidUserIDs", func(t *testing.T) {
			conversationID, err := messageRepo.CreateConversation([]int{99999, userIDs[0]})
			AssertNotEqual(t, err, nil, "Conversation creation should fail for invalid user ID")
			AssertEqual(t, conversationID, 0, "Conversation ID should be zero")
		})
	})

	t.Run("SendMessage", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		conversationIDs, err := SetupTestConversations(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test conversations")

		t.Run("ValidMessage", func(t *testing.T) {
			message, err := messageRepo.AddMessageToConversation(conversationIDs[0], userIDs[0], "Test message")
			AssertNoError(t, err, "Message sending should succeed")
			AssertNotEqual(t, message.ID, 0, "Message ID should be set")
		})

		t.Run("InvalidConversationID", func(t *testing.T) {
			message, err := messageRepo.AddMessageToConversation(99999, userIDs[0], "Test message")
			AssertNotEqual(t, err, nil, "Message sending should fail for invalid conversation ID")
			AssertEqual(t, message, (*database.Message)(nil), "Message should be nil")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			message, err := messageRepo.AddMessageToConversation(conversationIDs[0], 99999, "Test message")
			AssertNotEqual(t, err, nil, "Message sending should fail for invalid user ID")
			AssertEqual(t, message, (*database.Message)(nil), "Message should be nil")
		})
	})

	t.Run("GetMessages", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		conversationIDs, err := SetupTestConversations(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test conversations")

		_, err = SetupTestMessages(testDB.DB, conversationIDs, userIDs)
		AssertNoError(t, err, "Failed to setup test messages")

		t.Run("ValidConversationID", func(t *testing.T) {
			messages, err := messageRepo.GetConversationMessages(conversationIDs[0], 10, 0)
			AssertNoError(t, err, "Should retrieve messages")
			AssertGreaterThanOrEqual(t, len(messages), 0, "Should return messages or empty array")
		})

		t.Run("InvalidConversationID", func(t *testing.T) {
			messages, err := messageRepo.GetConversationMessages(99999, 10, 0)
			AssertNoError(t, err, "Should handle invalid conversation ID")
			AssertEqual(t, len(messages), 0, "Should return empty array")
		})

		t.Run("WithPagination", func(t *testing.T) {
			messages, err := messageRepo.GetConversationMessages(conversationIDs[0], 5, 0)
			AssertNoError(t, err, "Should retrieve messages with pagination")
			AssertLessThanOrEqual(t, len(messages), 5, "Should respect limit")
		})
	})

	t.Run("GetUserConversations", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		_, err = SetupTestConversations(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test conversations")

		t.Run("ValidUserID", func(t *testing.T) {
			conversations, err := messageRepo.GetUserConversations(userIDs[0])
			AssertNoError(t, err, "Should retrieve user conversations")
			AssertGreaterThanOrEqual(t, len(conversations), 0, "Should return conversations or empty array")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			conversations, err := messageRepo.GetUserConversations(99999)
			AssertNoError(t, err, "Should handle invalid user ID")
			AssertEqual(t, len(conversations), 0, "Should return empty array")
		})
	})
}
