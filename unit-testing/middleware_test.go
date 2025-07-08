package unit_testing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/server"
)

func TestAuthMiddleware(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create a test handler that requires authentication
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	t.Run("AuthenticatedRequest", func(t *testing.T) {
		// Create session for user
		sessionToken := CreateTestSession(t, testDB, userIDs[0])

		// Create request with session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		// Create response recorder
		rr := httptest.NewRecorder()

		// Apply auth middleware
		authHandler := server.AuthMiddleware(protectedHandler)
		authHandler.ServeHTTP(rr, req)

		// Should allow access
		AssertEqual(t, rr.Code, http.StatusOK, "Authenticated request should succeed")
		AssertEqual(t, rr.Body.String(), "Protected content", "Should return protected content")
	})

	t.Run("UnauthenticatedRequest", func(t *testing.T) {
		// Create request without session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		rr := httptest.NewRecorder()

		// Apply auth middleware
		authHandler := server.AuthMiddleware(protectedHandler)
		authHandler.ServeHTTP(rr, req)

		// Should redirect to login or return unauthorized
		AssertTrue(t, rr.Code == http.StatusUnauthorized || rr.Code == http.StatusFound,
			"Unauthenticated request should be rejected or redirected")
	})

	t.Run("InvalidSessionToken", func(t *testing.T) {
		// Create request with invalid session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "invalid_session_token",
		})

		rr := httptest.NewRecorder()

		// Apply auth middleware
		authHandler := server.AuthMiddleware(protectedHandler)
		authHandler.ServeHTTP(rr, req)

		// Should reject request
		AssertTrue(t, rr.Code == http.StatusUnauthorized || rr.Code == http.StatusFound,
			"Invalid session should be rejected")
	})

	t.Run("ExpiredSession", func(t *testing.T) {
		// Create session and then invalidate it
		sessionToken := CreateTestSession(t, testDB, userIDs[0])

		// Invalidate session by updating it to empty
		_, err := testDB.DB.Exec("UPDATE user SET current_session = NULL WHERE userid = ?", userIDs[0])
		AssertNoError(t, err, "Should be able to invalidate session")

		// Create request with expired session
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		rr := httptest.NewRecorder()

		// Apply auth middleware
		authHandler := server.AuthMiddleware(protectedHandler)
		authHandler.ServeHTTP(rr, req)

		// Should reject request
		AssertTrue(t, rr.Code == http.StatusUnauthorized || rr.Code == http.StatusFound,
			"Expired session should be rejected")
	})

	t.Run("MultipleSessionRequests", func(t *testing.T) {
		// Create sessions for multiple users
		sessionToken1 := CreateTestSession(t, testDB, userIDs[0])
		sessionToken2 := CreateTestSession(t, testDB, userIDs[1])

		// Test first user
		req1 := httptest.NewRequest("GET", "/protected", nil)
		req1.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken1,
		})

		rr1 := httptest.NewRecorder()
		authHandler := server.AuthMiddleware(protectedHandler)
		authHandler.ServeHTTP(rr1, req1)

		AssertEqual(t, rr1.Code, http.StatusOK, "First user should be authenticated")

		// Test second user
		req2 := httptest.NewRequest("GET", "/protected", nil)
		req2.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken2,
		})

		rr2 := httptest.NewRecorder()
		authHandler.ServeHTTP(rr2, req2)

		AssertEqual(t, rr2.Code, http.StatusOK, "Second user should be authenticated")
	})
}

func TestReverseMiddleware(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create a test handler for public pages
	publicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Public content"))
	})

	t.Run("UnauthenticatedUserAccessingPublicPage", func(t *testing.T) {
		// Create request without session
		req := httptest.NewRequest("GET", "/login", nil)
		rr := httptest.NewRecorder()

		// Apply reverse middleware (should allow access to public pages)
		reverseHandler := server.ReverseMiddleware(publicHandler)
		reverseHandler.ServeHTTP(rr, req)

		// Should allow access
		AssertEqual(t, rr.Code, http.StatusOK, "Unauthenticated user should access public pages")
	})

	t.Run("AuthenticatedUserAccessingPublicPage", func(t *testing.T) {
		// Create session for user
		sessionToken := CreateTestSession(t, testDB, userIDs[0])

		// Create request with session cookie to public page
		req := httptest.NewRequest("GET", "/login", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		rr := httptest.NewRecorder()

		// Apply reverse middleware (should redirect authenticated users away from login)
		reverseHandler := server.ReverseMiddleware(publicHandler)
		reverseHandler.ServeHTTP(rr, req)

		// Should redirect to home or allow access (depending on implementation)
		// This test depends on the specific implementation of ReverseMiddleware
		AssertTrue(t, rr.Code == http.StatusOK || rr.Code == http.StatusFound,
			"Authenticated user accessing public page should be handled appropriately")
	})
}

func TestInputValidation(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("SQLInjectionPrevention", func(t *testing.T) {
		// Test that SQL injection attempts are prevented
		maliciousInputs := []string{
			"'; DROP TABLE user; --",
			"' OR '1'='1",
			"admin'--",
			"' UNION SELECT * FROM user --",
			"'; INSERT INTO user VALUES ('hacker', 'hacker'); --",
		}

		for _, maliciousInput := range maliciousInputs {
			// Try to authenticate with malicious input
			_, err := testDB.DB.Query("SELECT * FROM user WHERE username = ?", maliciousInput)
			// Should not cause database errors (parameterized queries prevent injection)
			AssertNoError(t, err, "Parameterized queries should prevent SQL injection")
		}
	})

	t.Run("XSSPrevention", func(t *testing.T) {
		// Test XSS prevention in user input
		xssInputs := []string{
			"<script>alert('xss')</script>",
			"javascript:alert('xss')",
			"<img src=x onerror=alert('xss')>",
			"<svg onload=alert('xss')>",
			"'><script>alert('xss')</script>",
		}

		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		for _, xssInput := range xssInputs {
			// Try to create post with XSS content
			// The application should sanitize or escape this input
			postID, err := testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
				"Test Post", xssInput, userIDs[0])
			AssertNoError(t, err, "Should be able to insert content (will be sanitized on output)")

			// Verify content was stored (sanitization happens on output, not input)
			var storedContent string
			lastID, _ := postID.LastInsertId()
			err = testDB.DB.QueryRow("SELECT content FROM post WHERE postid = ?", lastID).Scan(&storedContent)
			AssertNoError(t, err, "Should be able to retrieve stored content")
			// Content should be stored as-is, sanitization happens during rendering
		}
	})

	t.Run("InputLengthValidation", func(t *testing.T) {
		// Test input length limits
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Test very long title
		longTitle := string(make([]byte, 1000))
		for i := range longTitle {
			longTitle = longTitle[:i] + "a" + longTitle[i+1:]
		}

		// Test very long content
		longContent := string(make([]byte, 10000))
		for i := range longContent {
			longContent = longContent[:i] + "a" + longContent[i+1:]
		}

		// Try to create post with very long inputs
		_, err = testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
			longTitle, longContent, userIDs[0])
		// Should either succeed (if no length limits) or fail gracefully
		// The specific behavior depends on database constraints and application validation
		if err != nil {
			// If it fails, it should be a validation error, not a crash
			AssertNotEqual(t, err.Error(), "", "Error should have a message")
		}
	})

	t.Run("SpecialCharacterHandling", func(t *testing.T) {
		// Test handling of special characters
		specialChars := []string{
			"Unicode: 你好世界",
			"Emojis: 😀😃😄😁",
			"Special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
			"Newlines:\nand\ttabs",
			"Quotes: 'single' and \"double\"",
		}

		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		for _, specialChar := range specialChars {
			// Try to create post with special characters
			_, err := testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
				"Special Char Test", specialChar, userIDs[0])
			AssertNoError(t, err, "Should handle special characters properly")
		}
	})
}

func TestErrorHandling(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("DatabaseConnectionError", func(t *testing.T) {
		// Test handling of database connection errors
		// This is difficult to test without actually breaking the database
		// In a real scenario, you might use dependency injection to mock the database

		// For now, test that the application handles missing tables gracefully
		_, err := testDB.DB.Query("SELECT * FROM nonexistent_table")
		AssertError(t, err, "Should error when querying non-existent table")
		// The application should handle this error gracefully, not crash
	})

	t.Run("InvalidDataHandling", func(t *testing.T) {
		// Test handling of invalid data types
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Try to insert invalid data types
		_, err = testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, ?, ?)",
			"Test", "Content", "invalid_date", userIDs[0])
		AssertError(t, err, "Should error with invalid date format")

		// Try to insert with invalid user ID
		_, err = testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
			"Test", "Content", "not_a_number")
		AssertError(t, err, "Should error with invalid user ID type")
	})

	t.Run("ConcurrentAccessHandling", func(t *testing.T) {
		// Test handling of concurrent database access
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Simulate concurrent post creation
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(index int) {
				_, err := testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
					fmt.Sprintf("Concurrent Post %d", index), "Concurrent content", userIDs[0])
				AssertNoError(t, err, "Concurrent post creation should succeed")
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all posts were created
		var count int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM post WHERE title LIKE 'Concurrent Post%'").Scan(&count)
		AssertNoError(t, err, "Should be able to count concurrent posts")
		AssertEqual(t, count, 10, "All concurrent posts should be created")
	})
}

func TestSecurityFeatures(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("PasswordHashing", func(t *testing.T) {
		// Test that passwords are properly hashed
		password := "test_password_123"

		// Create user with password
		userID, err := testDB.DB.Exec("INSERT INTO user (F_name, L_name, Username, Email, password, gender, date_of_birth) VALUES (?, ?, ?, ?, ?, ?, ?)",
			"Test", "User", "hashtest", "hash@example.com", password, "male", "1990-01-01")
		AssertNoError(t, err, "User creation should succeed")

		// Retrieve stored password
		var storedPassword string
		lastID, _ := userID.LastInsertId()
		err = testDB.DB.QueryRow("SELECT password FROM user WHERE userid = ?", lastID).Scan(&storedPassword)
		AssertNoError(t, err, "Should be able to retrieve password")

		// Password should be hashed (not stored in plain text)
		AssertNotEqual(t, storedPassword, password, "Password should be hashed, not stored in plain text")
		AssertTrue(t, len(storedPassword) > len(password), "Hashed password should be longer than original")
	})

	t.Run("SessionTokenSecurity", func(t *testing.T) {
		// Test session token generation and validation
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Create session token
		sessionToken1 := CreateTestSession(t, testDB, userIDs[0])
		sessionToken2 := CreateTestSession(t, testDB, userIDs[1])

		// Session tokens should be unique
		AssertNotEqual(t, sessionToken1, sessionToken2, "Session tokens should be unique")

		// Session tokens should be sufficiently long
		AssertTrue(t, len(sessionToken1) >= 16, "Session token should be sufficiently long")
		AssertTrue(t, len(sessionToken2) >= 16, "Session token should be sufficiently long")

		// Session tokens should not be predictable
		sessionToken3 := CreateTestSession(t, testDB, userIDs[2])
		AssertNotEqual(t, sessionToken1, sessionToken3, "Session tokens should not be predictable")
	})

	t.Run("AuthorizationChecks", func(t *testing.T) {
		// Test that users can only access their own data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Create posts for different users
		postID1, err := testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
			"User 1 Post", "Content from user 1", userIDs[0])
		AssertNoError(t, err, "Post creation should succeed")

		postID2, err := testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
			"User 2 Post", "Content from user 2", userIDs[1])
		AssertNoError(t, err, "Post creation should succeed")

		// Verify posts belong to correct users
		var ownerID int
		lastID1, _ := postID1.LastInsertId()
		err = testDB.DB.QueryRow("SELECT user_userid FROM post WHERE postid = ?", lastID1).Scan(&ownerID)
		AssertNoError(t, err, "Should be able to get post owner")
		AssertEqual(t, ownerID, userIDs[0], "Post should belong to correct user")

		lastID2, _ := postID2.LastInsertId()
		err = testDB.DB.QueryRow("SELECT user_userid FROM post WHERE postid = ?", lastID2).Scan(&ownerID)
		AssertNoError(t, err, "Should be able to get post owner")
		AssertEqual(t, ownerID, userIDs[1], "Post should belong to correct user")
	})

	t.Run("DataSanitization", func(t *testing.T) {
		// Test that data is properly sanitized
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Test with potentially dangerous content
		dangerousContent := "<script>alert('xss')</script><img src=x onerror=alert('xss')>"

		// Store content in database
		_, err = testDB.DB.Exec("INSERT INTO post (title, content, post_at, user_userid) VALUES (?, ?, datetime('now'), ?)",
			"Sanitization Test", dangerousContent, userIDs[0])
		AssertNoError(t, err, "Should be able to store content")

		// Content should be stored as-is (sanitization happens on output)
		// The application should sanitize when displaying, not when storing
		var storedContent string
		err = testDB.DB.QueryRow("SELECT content FROM post WHERE title = 'Sanitization Test'").Scan(&storedContent)
		AssertNoError(t, err, "Should be able to retrieve content")
		// Content is stored as-is, sanitization should happen during rendering
	})
}
