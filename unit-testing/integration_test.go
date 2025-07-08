package unit_testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"forum/repository"
	"forum/server/services"
)

func TestUserRegistrationAndLoginFlow(t *testing.T) {
	testDB := TestSetup(t)

	// Create HTTP test helper
	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	t.Run("CompleteUserRegistrationFlow", func(t *testing.T) {
		// Test user registration
		registrationData := url.Values{
			"firstName":   {"Integration"},
			"lastName":    {"Test"},
			"username":    {"integrationtest"},
			"email":       {"integration@example.com"},
			"gender":      {"male"},
			"dateOfBirth": {"1990-01-01"},
			"password":    {"password123"},
		}

		resp, err := httpHelper.POSTForm("/api/signup", registrationData, nil)
		AssertNoError(t, err, "Registration request should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		// Verify user was created in database
		var userID int
		err = testDB.DB.QueryRow("SELECT userid FROM user WHERE username = ?", "integrationtest").Scan(&userID)
		AssertNoError(t, err, "User should be created in database")
		AssertTrue(t, userID > 0, "User ID should be positive")

		// Test login with created user
		loginData := map[string]string{
			"identifier": "integrationtest",
			"password":   "password123",
		}

		loginResp, err := httpHelper.POST("/api/login", loginData, nil)
		AssertNoError(t, err, "Login request should succeed")
		AssertStatusCode(t, loginResp, http.StatusOK)

		// Verify session cookie is set
		AssertCookieExists(t, loginResp, "session_token")

		// Test accessing protected endpoint with session
		sessionCookie := getSessionCookie(loginResp)
		protectedResp, err := httpHelper.AuthenticatedRequest("GET", "/api/user/current", nil, sessionCookie)
		AssertNoError(t, err, "Protected request should succeed")
		AssertStatusCode(t, protectedResp, http.StatusOK)

		// Verify user data in response
		var userData map[string]interface{}
		AssertJSONResponse(t, protectedResp, &userData)
		AssertEqual(t, userData["username"], "integrationtest", "Username should match")
		AssertEqual(t, userData["email"], "integration@example.com", "Email should match")
	})

	t.Run("LoginWithEmail", func(t *testing.T) {
		// Create user first
		userRepo := repository.NewUserRepository(testDB.DB)
		userService := services.NewUserService(userRepo)

		_, err := userService.RegisterUser("Email", "Test", "emailtest", "emailtest@example.com",
			"female", "1992-01-01", "password123")
		AssertNoError(t, err, "User registration should succeed")

		// Test login with email
		loginData := map[string]string{
			"identifier": "emailtest@example.com",
			"password":   "password123",
		}

		loginResp, err := httpHelper.POST("/api/login", loginData, nil)
		AssertNoError(t, err, "Login with email should succeed")
		AssertStatusCode(t, loginResp, http.StatusOK)
		AssertCookieExists(t, loginResp, "session_token")
	})

	t.Run("InvalidLoginAttempts", func(t *testing.T) {
		// Test login with wrong password
		loginData := map[string]string{
			"identifier": "integrationtest",
			"password":   "wrongpassword",
		}

		loginResp, err := httpHelper.POST("/api/login", loginData, nil)
		AssertNoError(t, err, "Login request should complete")
		AssertStatusCode(t, loginResp, http.StatusUnauthorized)

		// Test login with non-existent user
		loginData["identifier"] = "nonexistent"
		loginData["password"] = "password123"

		loginResp, err = httpHelper.POST("/api/login", loginData, nil)
		AssertNoError(t, err, "Login request should complete")
		AssertStatusCode(t, loginResp, http.StatusUnauthorized)
	})
}

func TestPostCreationAndInteractionFlow(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create HTTP test helper
	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	// Login user
	sessionCookie := loginTestUser(t, httpHelper, testDB, userIDs[0])

	t.Run("CompletePostCreationFlow", func(t *testing.T) {
		// Create post
		postData := map[string]interface{}{
			"title":      "Integration Test Post",
			"content":    "This is a test post created during integration testing",
			"categories": []string{"General", "Testing"},
		}

		postResp, err := httpHelper.AuthenticatedRequest("POST", "/api/post/create", postData, sessionCookie)
		AssertNoError(t, err, "Post creation should succeed")
		AssertStatusCode(t, postResp, http.StatusOK)

		// Get post ID from response
		var postResponse map[string]interface{}
		AssertJSONResponse(t, postResp, &postResponse)
		postID := int(postResponse["post_id"].(float64))

		// Verify post was created in database
		var dbPostID int
		var title, content string
		err = testDB.DB.QueryRow("SELECT postid, title, content FROM post WHERE postid = ?", postID).
			Scan(&dbPostID, &title, &content)
		AssertNoError(t, err, "Post should exist in database")
		AssertEqual(t, title, "Integration Test Post", "Post title should match")
		AssertEqual(t, content, "This is a test post created during integration testing", "Post content should match")

		// Verify categories were created and linked
		var categoryCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM post_categories WHERE post_id = ?", postID).Scan(&categoryCount)
		AssertNoError(t, err, "Should be able to query post categories")
		AssertEqual(t, categoryCount, 2, "Post should have 2 categories")

		// Get all posts and verify our post is included
		postsResp, err := httpHelper.GET("/api/posts", nil)
		AssertNoError(t, err, "Should be able to get all posts")
		AssertStatusCode(t, postsResp, http.StatusOK)

		var postsData map[string]interface{}
		AssertJSONResponse(t, postsResp, &postsData)
		posts := postsData["posts"].([]interface{})

		// Find our post
		found := false
		for _, post := range posts {
			postMap := post.(map[string]interface{})
			if int(postMap["postid"].(float64)) == postID {
				found = true
				AssertEqual(t, postMap["title"], "Integration Test Post", "Post title should match in list")
				break
			}
		}
		AssertTrue(t, found, "Created post should appear in posts list")
	})

	t.Run("PostCommentingFlow", func(t *testing.T) {
		// Create a post first
		postService := services.NewPostService(testDB.DB)
		postID, err := postService.CreatePost(userIDs[0], "Comment Test Post", "Post for testing comments", []string{"General"})
		AssertNoError(t, err, "Post creation should succeed")

		// Add comment to post
		commentData := map[string]interface{}{
			"postID":  postID,
			"content": "This is a test comment",
		}

		commentResp, err := httpHelper.AuthenticatedRequest("POST", "/addcomment", commentData, sessionCookie)
		AssertNoError(t, err, "Comment creation should succeed")
		AssertStatusCode(t, commentResp, http.StatusOK)

		// Verify comment was added to database
		var commentCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM comment WHERE post_postid = ?", postID).Scan(&commentCount)
		AssertNoError(t, err, "Should be able to query comments")
		AssertEqual(t, commentCount, 1, "Should have 1 comment")

		// Get post details and verify comment is included
		postResp, err := httpHelper.GET(fmt.Sprintf("/api/post?id=%d", postID), nil)
		AssertNoError(t, err, "Should be able to get post details")
		AssertStatusCode(t, postResp, http.StatusOK)

		var postData map[string]interface{}
		AssertJSONResponse(t, postResp, &postData)

		// Verify comment count increased
		comments := postData["comments"].([]interface{})
		AssertTrue(t, len(comments) >= 1, "Post should have at least 1 comment")

		// Find our comment
		found := false
		for _, comment := range comments {
			commentMap := comment.(map[string]interface{})
			if commentMap["content"] == "This is a test comment" {
				found = true
				break
			}
		}
		AssertTrue(t, found, "Our comment should be in the comments list")
	})
}

func TestMessagingWorkflow(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create HTTP test helper
	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	// Login both users
	sessionCookie1 := loginTestUser(t, httpHelper, testDB, userIDs[0])
	sessionCookie2 := loginTestUser(t, httpHelper, testDB, userIDs[1])

	t.Run("CompleteMessagingFlow", func(t *testing.T) {
		// Create conversation
		conversationData := map[string]interface{}{
			"participants": []int{userIDs[1]}, // User 0 creates conversation with User 1
		}

		convResp, err := httpHelper.AuthenticatedRequest("POST", "/api/conversations", conversationData, sessionCookie1)
		AssertNoError(t, err, "Conversation creation should succeed")
		AssertStatusCode(t, convResp, http.StatusOK)

		var convResponse map[string]interface{}
		AssertJSONResponse(t, convResp, &convResponse)
		conversationID := int(convResponse["conversation_id"].(float64))

		// Send message in conversation
		messageData := map[string]interface{}{
			"conversation_id": conversationID,
			"content":         "Hello from integration test!",
		}

		msgResp, err := httpHelper.AuthenticatedRequest("POST", "/api/messages", messageData, sessionCookie1)
		AssertNoError(t, err, "Message sending should succeed")
		AssertStatusCode(t, msgResp, http.StatusOK)

		// Verify message was stored in database
		var messageCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM message WHERE conversation_id = ?", conversationID).Scan(&messageCount)
		AssertNoError(t, err, "Should be able to query messages")
		AssertEqual(t, messageCount, 1, "Should have 1 message")

		// Get messages from conversation
		messagesResp, err := httpHelper.AuthenticatedRequest("GET",
			fmt.Sprintf("/api/messages?conversation_id=%d", conversationID), nil, sessionCookie1)
		AssertNoError(t, err, "Should be able to get messages")
		AssertStatusCode(t, messagesResp, http.StatusOK)

		var messagesData map[string]interface{}
		AssertJSONResponse(t, messagesResp, &messagesData)
		messages := messagesData["messages"].([]interface{})
		AssertEqual(t, len(messages), 1, "Should have 1 message")

		message := messages[0].(map[string]interface{})
		AssertEqual(t, message["content"], "Hello from integration test!", "Message content should match")
		AssertEqual(t, int(message["sender_id"].(float64)), userIDs[0], "Sender ID should match")

		// Get conversations for both users
		conv1Resp, err := httpHelper.AuthenticatedRequest("GET", "/api/conversations", nil, sessionCookie1)
		AssertNoError(t, err, "User 1 should be able to get conversations")
		AssertStatusCode(t, conv1Resp, http.StatusOK)

		conv2Resp, err := httpHelper.AuthenticatedRequest("GET", "/api/conversations", nil, sessionCookie2)
		AssertNoError(t, err, "User 2 should be able to get conversations")
		AssertStatusCode(t, conv2Resp, http.StatusOK)

		// Both users should see the conversation
		var conv1Data, conv2Data map[string]interface{}
		AssertJSONResponse(t, conv1Resp, &conv1Data)
		AssertJSONResponse(t, conv2Resp, &conv2Data)

		conversations1 := conv1Data["conversations"].([]interface{})
		conversations2 := conv2Data["conversations"].([]interface{})

		AssertTrue(t, len(conversations1) >= 1, "User 1 should have at least 1 conversation")
		AssertTrue(t, len(conversations2) >= 1, "User 2 should have at least 1 conversation")
	})

	t.Run("MessageReadStatusFlow", func(t *testing.T) {
		// Create conversation and send message
		messageRepo := repository.NewMessageRepository(testDB.DB)
		messageService := services.NewMessageService(testDB.DB)

		conversationID, err := messageRepo.CreateConversation([]int{userIDs[0], userIDs[1]})
		AssertNoError(t, err, "Conversation creation should succeed")

		_, err = messageService.SendMessage(conversationID, userIDs[0], "Test message for read status")
		AssertNoError(t, err, "Message sending should succeed")

		// Mark messages as read
		readData := map[string]interface{}{
			"conversation_id": conversationID,
		}

		readResp, err := httpHelper.AuthenticatedRequest("POST", "/api/messages/read", readData, sessionCookie2)
		AssertNoError(t, err, "Marking messages as read should succeed")
		AssertStatusCode(t, readResp, http.StatusOK)

		// Verify messages are marked as read in database
		var unreadCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM message WHERE conversation_id = ? AND is_read = 0", conversationID).Scan(&unreadCount)
		AssertNoError(t, err, "Should be able to query unread messages")
		AssertEqual(t, unreadCount, 0, "Should have 0 unread messages after marking as read")
	})
}

func TestEndToEndUserJourney(t *testing.T) {
	testDB := TestSetup(t)

	// Create HTTP test helper
	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	t.Run("CompleteUserJourney", func(t *testing.T) {
		// 1. User registration
		registrationData := url.Values{
			"firstName":   {"Journey"},
			"lastName":    {"User"},
			"username":    {"journeyuser"},
			"email":       {"journey@example.com"},
			"gender":      {"female"},
			"dateOfBirth": {"1995-01-01"},
			"password":    {"journey123"},
		}

		regResp, err := httpHelper.POSTForm("/api/signup", registrationData, nil)
		AssertNoError(t, err, "Registration should succeed")
		AssertStatusCode(t, regResp, http.StatusOK)

		// 2. User login
		loginData := map[string]string{
			"identifier": "journeyuser",
			"password":   "journey123",
		}

		loginResp, err := httpHelper.POST("/api/login", loginData, nil)
		AssertNoError(t, err, "Login should succeed")
		AssertStatusCode(t, loginResp, http.StatusOK)
		sessionCookie := getSessionCookie(loginResp)

		// 3. Create multiple posts
		for i := 1; i <= 3; i++ {
			postData := map[string]interface{}{
				"title":      fmt.Sprintf("Journey Post %d", i),
				"content":    fmt.Sprintf("This is journey post number %d", i),
				"categories": []string{"General", "Journey"},
			}

			postResp, err := httpHelper.AuthenticatedRequest("POST", "/api/post/create", postData, sessionCookie)
			AssertNoError(t, err, "Post creation should succeed")
			AssertStatusCode(t, postResp, http.StatusOK)
		}

		// 4. Get all posts and verify user's posts are included
		postsResp, err := httpHelper.GET("/api/posts", nil)
		AssertNoError(t, err, "Should be able to get posts")
		AssertStatusCode(t, postsResp, http.StatusOK)

		var postsData map[string]interface{}
		AssertJSONResponse(t, postsResp, &postsData)
		posts := postsData["posts"].([]interface{})

		journeyPostCount := 0
		for _, post := range posts {
			postMap := post.(map[string]interface{})
			if strings.Contains(postMap["title"].(string), "Journey Post") {
				journeyPostCount++
			}
		}
		AssertEqual(t, journeyPostCount, 3, "Should find all 3 journey posts")

		// 5. Comment on posts
		if len(posts) > 0 {
			firstPost := posts[0].(map[string]interface{})
			postID := int(firstPost["postid"].(float64))

			commentData := map[string]interface{}{
				"postID":  postID,
				"content": "Great post! Thanks for sharing.",
			}

			commentResp, err := httpHelper.AuthenticatedRequest("POST", "/addcomment", commentData, sessionCookie)
			AssertNoError(t, err, "Comment creation should succeed")
			AssertStatusCode(t, commentResp, http.StatusOK)
		}

		// 6. Get user profile
		profileResp, err := httpHelper.AuthenticatedRequest("GET", "/api/user/current", nil, sessionCookie)
		AssertNoError(t, err, "Should be able to get user profile")
		AssertStatusCode(t, profileResp, http.StatusOK)

		var profileData map[string]interface{}
		AssertJSONResponse(t, profileResp, &profileData)
		AssertEqual(t, profileData["username"], "journeyuser", "Username should match")
		AssertEqual(t, profileData["email"], "journey@example.com", "Email should match")

		// 7. Logout
		logoutResp, err := httpHelper.AuthenticatedRequest("POST", "/api/logout", nil, sessionCookie)
		AssertNoError(t, err, "Logout should succeed")
		AssertStatusCode(t, logoutResp, http.StatusOK)

		// 8. Verify session is invalidated
		protectedResp, err := httpHelper.AuthenticatedRequest("GET", "/api/user/current", nil, sessionCookie)
		AssertNoError(t, err, "Request should complete")
		AssertStatusCode(t, protectedResp, http.StatusUnauthorized)
	})
}

// Helper functions

func createTestServer(testDB *TestDatabase) http.Handler {
	// This would create a test server with the actual application routes
	// For now, return a simple mux for demonstration
	mux := http.NewServeMux()

	// Add basic routes for testing
	mux.HandleFunc("/api/signup", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: "test_session_token",
		})
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	return mux
}

func loginTestUser(t *testing.T, httpHelper *HTTPTestHelper, testDB *TestDatabase, userID int) *http.Cookie {
	// Create session for user
	sessionToken := CreateTestSession(t, testDB, userID)

	return &http.Cookie{
		Name:  "session_token",
		Value: sessionToken,
	}
}

func getSessionCookie(resp *http.Response) *http.Cookie {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_token" {
			return cookie
		}
	}
	return nil
}
