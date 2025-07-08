package unit_testing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"forum/database"
	"forum/server"
)

func TestCompleteUserJourneyE2E(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	t.Run("NewUserCompleteJourney", func(t *testing.T) {
		var bodyBytes []byte
		// Step 1: User visits the site and registers
		registrationData := url.Values{
			"firstName":   {"Journey"},
			"lastName":    {"User"},
			"username":    {"journeyuser"},
			"email":       {"journey@example.com"},
			"gender":      {"female"},
			"dateOfBirth": {"1995-05-15"},
			"password":    {"securepass123"},
		}

		resp, err := httpHelper.POSTForm("/api/signup", registrationData, nil)
		AssertNoError(t, err, "Registration should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		// Extract session from registration
		sessionCookie := extractSessionCookie(resp)
		AssertNotEqual(t, sessionCookie, "", "Session should be created after registration")

		// Step 2: User browses posts
		resp, err = httpHelper.GET("/api/posts", map[string]string{"session_token": sessionCookie})
		AssertNoError(t, err, "Should be able to browse posts")
		AssertStatusCode(t, resp, http.StatusOK)

		var posts []database.Post
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &posts)
		AssertNoError(t, err, "Should parse posts response")

		// Step 3: User creates their first post
		postData := map[string]interface{}{
			"title":      "My First Post on the Forum",
			"content":    "Hello everyone! I'm new here and excited to be part of this community. Looking forward to great discussions!",
			"categories": []string{"General", "Introduction"},
		}

		resp, err = httpHelper.POST("/api/post/create", postData, map[string]string{"session_token": sessionCookie})
		AssertNoError(t, err, "Post creation should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		var createPostResp server.CreatePostResponse
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &createPostResp)
		AssertNoError(t, err, "Should parse create post response")
		AssertEqual(t, createPostResp.Success, true, "Post creation should be successful")

		firstPostID := createPostResp.PostID

		// Step 4: User views their created post
		resp, err = httpHelper.GET("/api/post?id="+strconv.Itoa(firstPostID), map[string]string{"session_token": sessionCookie})
		AssertNoError(t, err, "Should be able to view created post")
		AssertStatusCode(t, resp, http.StatusOK)

		var createdPost database.Post
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &createdPost)
		AssertNoError(t, err, "Should parse post response")
		AssertEqual(t, createdPost.Title, "My First Post on the Forum", "Post title should match")

		// Step 5: User comments on another post (if available)
		if len(posts) > 0 {
			commentData := url.Values{
				"post_id": {strconv.Itoa(posts[0].PostID)},
				"content": {"Great post! Thanks for sharing this information."},
			}

			resp, err = httpHelper.POSTForm("/addcomment", commentData, map[string]string{"session_token": sessionCookie})
			AssertNoError(t, err, "Comment addition should succeed")
			AssertStatusCode(t, resp, http.StatusSeeOther) // Redirect after comment
		}

		// Step 6: User starts a conversation with another user
		// First, get list of users
		resp, err = httpHelper.GET("/api/users", map[string]string{"session_token": sessionCookie})
		if err == nil && resp.StatusCode == http.StatusOK {
			var users []database.User
			bodyBytes, err := io.ReadAll(resp.Body)
			if err == nil {
				err = json.Unmarshal(bodyBytes, &users)
			}
			if err == nil && len(users) > 1 {
				// Find another user
				var otherUserID int
				for _, user := range users {
					if user.Username != "journeyuser" {
						otherUserID = user.ID
						break
					}
				}

				if otherUserID > 0 {
					// Create conversation
					convData := map[string]interface{}{
						"user_id": otherUserID,
					}

					resp, err = httpHelper.POST("/api/create-conversation", convData, map[string]string{"session_token": sessionCookie})
					if err == nil && resp.StatusCode == http.StatusOK {
						var convResp server.CreateConversationResponse
						bodyBytes, err := io.ReadAll(resp.Body)
						if err == nil {
							err = json.Unmarshal(bodyBytes, &convResp)
						}
						if err == nil && convResp.Success {
							// Send a message
							msgData := map[string]interface{}{
								"conversation_id": convResp.ConversationID,
								"content":         "Hi! I'm new to the forum. Nice to meet you!",
							}

							resp, err = httpHelper.POST("/api/send-message", msgData, map[string]string{"session_token": sessionCookie})
							AssertNoError(t, err, "Message sending should succeed")
							AssertStatusCode(t, resp, http.StatusOK)
						}
					}
				}
			}
		}

		// Step 7: User logs out
		resp, err = httpHelper.POST("/api/logout", nil, map[string]string{"session_token": sessionCookie})
		AssertNoError(t, err, "Logout should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		// Step 8: User logs back in
		loginData := map[string]string{
			"identifier": "journeyuser",
			"password":   "securepass123",
		}

		resp, err = httpHelper.POST("/api/login", loginData, nil)
		AssertNoError(t, err, "Login should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		var loginResp server.LoginResponse
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &loginResp)
		AssertNoError(t, err, "Should parse login response")
		AssertEqual(t, loginResp.Success, true, "Login should be successful")
		AssertEqual(t, loginResp.Username, "journeyuser", "Username should match")
	})
}

func TestMultiUserInteractionE2E(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	t.Run("MultiUserPostInteraction", func(t *testing.T) {
		var bodyBytes []byte
		// Create multiple users
		users := []struct {
			username string
			email    string
			session  string
		}{
			{"alice", "alice@example.com", ""},
			{"bob", "bob@example.com", ""},
			{"charlie", "charlie@example.com", ""},
		}

		// Register all users
		for i := range users {
			registrationData := url.Values{
				"firstName":   {users[i].username},
				"lastName":    {"User"},
				"username":    {users[i].username},
				"email":       {users[i].email},
				"gender":      {"other"},
				"dateOfBirth": {"1990-01-01"},
				"password":    {"password123"},
			}

			resp, err := httpHelper.POSTForm("/api/signup", registrationData, nil)
			AssertNoError(t, err, fmt.Sprintf("Registration for %s should succeed", users[i].username))
			AssertStatusCode(t, resp, http.StatusOK)

			users[i].session = extractSessionCookie(resp)
			AssertNotEqual(t, users[i].session, "", fmt.Sprintf("Session for %s should be created", users[i].username))
		}

		// Alice creates a post
		postData := map[string]interface{}{
			"title":      "Discussion: Best Programming Practices",
			"content":    "What are your thoughts on the best programming practices for beginners? I'd love to hear different perspectives!",
			"categories": []string{"Programming", "Discussion"},
		}

		resp, err := httpHelper.POST("/api/post/create", postData, map[string]string{"session_token": users[0].session})
		AssertNoError(t, err, "Alice's post creation should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		var createPostResp server.CreatePostResponse
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &createPostResp)
		AssertNoError(t, err, "Should parse create post response")
		postID := createPostResp.PostID

		// Bob comments on Alice's post
		commentData := url.Values{
			"post_id": {strconv.Itoa(postID)},
			"content": {"Great question! I think code readability is the most important aspect for beginners."},
		}

		resp, err = httpHelper.POSTForm("/addcomment", commentData, map[string]string{"session_token": users[1].session})
		AssertNoError(t, err, "Bob's comment should succeed")
		AssertStatusCode(t, resp, http.StatusSeeOther)

		// Charlie also comments
		commentData = url.Values{
			"post_id": {strconv.Itoa(postID)},
			"content": {"I agree with Bob! Also, writing tests early helps a lot."},
		}

		resp, err = httpHelper.POSTForm("/addcomment", commentData, map[string]string{"session_token": users[2].session})
		AssertNoError(t, err, "Charlie's comment should succeed")
		AssertStatusCode(t, resp, http.StatusSeeOther)

		// Alice responds to the comments
		commentData = url.Values{
			"post_id": {strconv.Itoa(postID)},
			"content": {"Thanks for the insights! Testing is definitely something I need to work on."},
		}

		resp, err = httpHelper.POSTForm("/addcomment", commentData, map[string]string{"session_token": users[0].session})
		AssertNoError(t, err, "Alice's response should succeed")
		AssertStatusCode(t, resp, http.StatusSeeOther)

		// Verify all comments are present
		resp, err = httpHelper.GET("/api/post?id="+strconv.Itoa(postID), map[string]string{"session_token": users[0].session})
		AssertNoError(t, err, "Should retrieve post with comments")
		AssertStatusCode(t, resp, http.StatusOK)

		var post database.Post
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &post)
		AssertNoError(t, err, "Should parse post response")
		AssertGreaterThanOrEqual(t, post.Comments, 3, "Post should have at least 3 comments")

		// Note: post.Comments is a count (int), not a slice of comments
		// To get actual comments, we would need to make a separate API call
		// For now, we just verify that the comment count is correct
	})
}

func TestMessagingSystemE2E(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	httpHelper := NewHTTPTestHelper(createTestServer(testDB))
	defer httpHelper.Close()

	t.Run("CompleteMessagingFlow", func(t *testing.T) {
		var bodyBytes []byte
		// Create two users
		user1Data := url.Values{
			"firstName":   {"Sender"},
			"lastName":    {"User"},
			"username":    {"sender"},
			"email":       {"sender@example.com"},
			"gender":      {"male"},
			"dateOfBirth": {"1990-01-01"},
			"password":    {"password123"},
		}

		user2Data := url.Values{
			"firstName":   {"Receiver"},
			"lastName":    {"User"},
			"username":    {"receiver"},
			"email":       {"receiver@example.com"},
			"gender":      {"female"},
			"dateOfBirth": {"1992-02-02"},
			"password":    {"password123"},
		}

		// Register users
		resp1, err := httpHelper.POSTForm("/api/signup", user1Data, nil)
		AssertNoError(t, err, "User 1 registration should succeed")
		AssertStatusCode(t, resp1, http.StatusOK)
		session1 := extractSessionCookie(resp1)

		resp2, err := httpHelper.POSTForm("/api/signup", user2Data, nil)
		AssertNoError(t, err, "User 2 registration should succeed")
		AssertStatusCode(t, resp2, http.StatusOK)
		session2 := extractSessionCookie(resp2)

		// Get user IDs
		var user1ID, user2ID int
		err = testDB.DB.QueryRow("SELECT userid FROM user WHERE username = ?", "sender").Scan(&user1ID)
		AssertNoError(t, err, "Should find user 1")
		err = testDB.DB.QueryRow("SELECT userid FROM user WHERE username = ?", "receiver").Scan(&user2ID)
		AssertNoError(t, err, "Should find user 2")

		// User 1 creates a conversation with User 2
		convData := map[string]interface{}{
			"user_id": user2ID,
		}

		resp, err := httpHelper.POST("/api/create-conversation", convData, map[string]string{"session_token": session1})
		AssertNoError(t, err, "Conversation creation should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		var convResp server.CreateConversationResponse
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &convResp)
		AssertNoError(t, err, "Should parse conversation response")
		AssertEqual(t, convResp.Success, true, "Conversation creation should be successful")
		conversationID := convResp.ConversationID

		// User 1 sends multiple messages
		messages := []string{
			"Hello! How are you doing?",
			"I saw your post about programming practices.",
			"Would love to discuss it further!",
		}

		for _, msgContent := range messages {
			msgData := map[string]interface{}{
				"conversation_id": conversationID,
				"content":         msgContent,
			}

			resp, err = httpHelper.POST("/api/send-message", msgData, map[string]string{"session_token": session1})
			AssertNoError(t, err, "Message sending should succeed")
			AssertStatusCode(t, resp, http.StatusOK)

			// Small delay between messages
			time.Sleep(100 * time.Millisecond)
		}

		// User 2 checks conversations
		resp, err = httpHelper.GET("/api/conversations", map[string]string{"session_token": session2})
		AssertNoError(t, err, "Should retrieve conversations")
		AssertStatusCode(t, resp, http.StatusOK)

		var conversations []database.Conversation
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &conversations)
		AssertNoError(t, err, "Should parse conversations")
		AssertGreaterThan(t, len(conversations), 0, "User 2 should have conversations")

		// User 2 reads messages
		resp, err = httpHelper.GET("/api/messages?conversation_id="+strconv.Itoa(conversationID), map[string]string{"session_token": session2})
		AssertNoError(t, err, "Should retrieve messages")
		AssertStatusCode(t, resp, http.StatusOK)

		var retrievedMessages []database.Message
		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &retrievedMessages)
		AssertNoError(t, err, "Should parse messages")
		AssertEqual(t, len(retrievedMessages), len(messages), "Should retrieve all sent messages")

		// Verify message content and order
		for i, msg := range retrievedMessages {
			AssertEqual(t, msg.Content, messages[i], fmt.Sprintf("Message %d content should match", i+1))
			AssertEqual(t, msg.SenderID, user1ID, "Sender ID should match")
		}

		// User 2 replies
		replyData := map[string]interface{}{
			"conversation_id": conversationID,
			"content":         "Hi! I'm doing well, thanks for asking. Yes, let's definitely discuss programming!",
		}

		resp, err = httpHelper.POST("/api/send-message", replyData, map[string]string{"session_token": session2})
		AssertNoError(t, err, "Reply should succeed")
		AssertStatusCode(t, resp, http.StatusOK)

		// User 1 checks for new messages
		resp, err = httpHelper.GET("/api/messages?conversation_id="+strconv.Itoa(conversationID), map[string]string{"session_token": session1})
		AssertNoError(t, err, "Should retrieve updated messages")
		AssertStatusCode(t, resp, http.StatusOK)

		bodyBytes, err = io.ReadAll(resp.Body)
		AssertNoError(t, err, "Should read response body")
		err = json.Unmarshal(bodyBytes, &retrievedMessages)
		AssertNoError(t, err, "Should parse updated messages")
		AssertEqual(t, len(retrievedMessages), len(messages)+1, "Should have original messages plus reply")

		// Verify the reply is present
		lastMessage := retrievedMessages[len(retrievedMessages)-1]
		AssertEqual(t, lastMessage.SenderID, user2ID, "Last message should be from user 2")
		AssertEqual(t, lastMessage.Content, "Hi! I'm doing well, thanks for asking. Yes, let's definitely discuss programming!", "Reply content should match")
	})
}

// Helper function to extract session cookie from response
func extractSessionCookie(resp *http.Response) string {
	if resp.Header == nil {
		return ""
	}

	setCookieHeaders, exists := resp.Header["Set-Cookie"]
	if !exists {
		return ""
	}

	for _, cookieHeader := range setCookieHeaders {
		if strings.Contains(cookieHeader, "session_token=") {
			// Extract the session token value
			parts := strings.Split(cookieHeader, ";")
			for _, part := range parts {
				if strings.HasPrefix(strings.TrimSpace(part), "session_token=") {
					return strings.TrimPrefix(strings.TrimSpace(part), "session_token=")
				}
			}
		}
	}

	return ""
}
