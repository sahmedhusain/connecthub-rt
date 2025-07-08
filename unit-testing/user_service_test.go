package unit_testing

import (
	"testing"

	"forum/database"
	"forum/repository"
	"forum/server/services"
)

func TestUserService(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Create user repository and service
	userRepo := repository.NewUserRepository(testDB.DB)
	userService := services.NewUserService(userRepo)

	t.Run("AuthenticateUser", func(t *testing.T) {
		// Setup test users
		_, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidCredentials", func(t *testing.T) {
			user, err := userService.AuthenticateUser("johndoe", "password123")
			AssertNoError(t, err, "Authentication should succeed")
			AssertNotEqual(t, user, nil, "User should be returned")
			AssertEqual(t, user.Username, "johndoe", "Username should match")
			AssertEqual(t, user.Email, "john@example.com", "Email should match")
		})

		t.Run("ValidEmailCredentials", func(t *testing.T) {
			user, err := userService.AuthenticateUser("jane@example.com", "password123")
			AssertNoError(t, err, "Authentication with email should succeed")
			AssertNotEqual(t, user, nil, "User should be returned")
			AssertEqual(t, user.Username, "janesmith", "Username should match")
			AssertEqual(t, user.Email, "jane@example.com", "Email should match")
		})

		t.Run("InvalidPassword", func(t *testing.T) {
			user, err := userService.AuthenticateUser("johndoe", "wrongpassword")
			AssertNotEqual(t, err, nil, "Authentication should fail")
			AssertEqual(t, user, nil, "User should not be returned")
		})

		t.Run("NonexistentUser", func(t *testing.T) {
			user, err := userService.AuthenticateUser("nonexistent", "password123")
			AssertNotEqual(t, err, nil, "Authentication should fail")
			AssertEqual(t, user, nil, "User should not be returned")
		})

		t.Run("EmptyCredentials", func(t *testing.T) {
			user, err := userService.AuthenticateUser("", "")
			AssertNotEqual(t, err, nil, "Authentication should fail with empty credentials")
			AssertEqual(t, user, nil, "User should not be returned")
		})

		t.Run("EmptyUsername", func(t *testing.T) {
			user, err := userService.AuthenticateUser("", "password123")
			AssertNotEqual(t, err, nil, "Authentication should fail with empty username")
			AssertEqual(t, user, nil, "User should not be returned")
		})

		t.Run("EmptyPassword", func(t *testing.T) {
			user, err := userService.AuthenticateUser("johndoe", "")
			AssertNotEqual(t, err, nil, "Authentication should fail with empty password")
			AssertEqual(t, user, nil, "User should not be returned")
		})
	})

	t.Run("CreateUserSession", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidUserID", func(t *testing.T) {
			sessionToken, err := userService.CreateUserSession(userIDs[0])
			AssertNoError(t, err, "Session creation should succeed")
			AssertNotEqual(t, sessionToken, "", "Session token should not be empty")
			AssertGreaterThan(t, len(sessionToken), 10, "Session token should be sufficiently long")

			// Verify session is stored in database
			user, err := userService.GetUserBySession(sessionToken)
			AssertNoError(t, err, "Should be able to retrieve user by session")
			AssertEqual(t, user.ID, userIDs[0], "User ID should match")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			sessionToken, err := userService.CreateUserSession(99999)
			AssertNotEqual(t, err, nil, "Session creation should fail for invalid user ID")
			AssertEqual(t, sessionToken, "", "Session token should be empty")
		})

		t.Run("ZeroUserID", func(t *testing.T) {
			sessionToken, err := userService.CreateUserSession(0)
			AssertNotEqual(t, err, nil, "Session creation should fail for zero user ID")
			AssertEqual(t, sessionToken, "", "Session token should be empty")
		})

		t.Run("MultipleSessionsForSameUser", func(t *testing.T) {
			// Create first session
			sessionToken1, err := userService.CreateUserSession(userIDs[0])
			AssertNoError(t, err, "First session creation should succeed")

			// Create second session (should replace first)
			sessionToken2, err := userService.CreateUserSession(userIDs[0])
			AssertNoError(t, err, "Second session creation should succeed")

			AssertNotEqual(t, sessionToken1, sessionToken2, "Session tokens should be different")

			// First session should no longer be valid
			user, err := userService.GetUserBySession(sessionToken1)
			AssertNotEqual(t, err, nil, "First session should be invalid")
			AssertEqual(t, user, nil, "Should not return user for old session")

			// Second session should be valid
			user, err = userService.GetUserBySession(sessionToken2)
			AssertNoError(t, err, "Second session should be valid")
			AssertEqual(t, user.ID, userIDs[0], "User ID should match")
		})
	})

	t.Run("RegisterUser", func(t *testing.T) {
		t.Run("ValidRegistration", func(t *testing.T) {
			userID, err := userService.RegisterUser(
				"New", "User", "newuser", "newuser@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNoError(t, err, "User registration should succeed")
			AssertNotEqual(t, userID, 0, "User ID should be set")

			// Verify user can authenticate
			user, err := userService.AuthenticateUser("newuser", "password123")
			AssertNoError(t, err, "New user should be able to authenticate")
			AssertEqual(t, user.ID, userID, "User ID should match")
			AssertEqual(t, user.Username, "newuser", "Username should match")
		})

		t.Run("DuplicateUsername", func(t *testing.T) {
			// Setup existing users
			_, err := SetupTestUsers(testDB.DB)
			AssertNoError(t, err, "Failed to setup test users")

			userID, err := userService.RegisterUser(
				"Duplicate", "User", "johndoe", "duplicate@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for duplicate username")
			AssertEqual(t, userID, 0, "User ID should be zero")
		})

		t.Run("DuplicateEmail", func(t *testing.T) {
			userID, err := userService.RegisterUser(
				"Duplicate", "Email", "duplicateemail", "john@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for duplicate email")
			AssertEqual(t, userID, 0, "User ID should be zero")
		})

		t.Run("InvalidEmail", func(t *testing.T) {
			userID, err := userService.RegisterUser(
				"Invalid", "Email", "invalidemail", "invalid-email",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for invalid email")
			AssertEqual(t, userID, 0, "User ID should be zero")
		})

		t.Run("MissingRequiredFields", func(t *testing.T) {
			// Missing first name
			userID, err := userService.RegisterUser(
				"", "User", "missingfirst", "missing@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for missing first name")
			AssertEqual(t, userID, 0, "User ID should be zero")

			// Missing last name
			userID, err = userService.RegisterUser(
				"Missing", "", "missinglast", "missing2@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for missing last name")
			AssertEqual(t, userID, 0, "User ID should be zero")

			// Missing username
			userID, err = userService.RegisterUser(
				"Missing", "Username", "", "missing3@example.com",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for missing username")
			AssertEqual(t, userID, 0, "User ID should be zero")

			// Missing email
			userID, err = userService.RegisterUser(
				"Missing", "Email", "missingemail", "",
				"male", "1990-01-01", "password123",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for missing email")
			AssertEqual(t, userID, 0, "User ID should be zero")

			// Missing password
			userID, err = userService.RegisterUser(
				"Missing", "Password", "missingpass", "missing4@example.com",
				"male", "1990-01-01", "",
			)
			AssertNotEqual(t, err, nil, "Registration should fail for missing password")
			AssertEqual(t, userID, 0, "User ID should be zero")
		})

		t.Run("OptionalFields", func(t *testing.T) {
			// Gender and date of birth are optional
			userID, err := userService.RegisterUser(
				"Optional", "Fields", "optionalfields", "optional@example.com",
				"", "", "password123",
			)
			AssertNoError(t, err, "Registration should succeed with optional fields empty")
			AssertNotEqual(t, userID, 0, "User ID should be set")
		})
	})

	t.Run("GetUserBySession", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Create session
		sessionToken, err := userService.CreateUserSession(userIDs[0])
		AssertNoError(t, err, "Session creation should succeed")

		t.Run("ValidSession", func(t *testing.T) {
			user, err := userService.GetUserBySession(sessionToken)
			AssertNoError(t, err, "Should retrieve user by valid session")
			AssertNotEqual(t, user, nil, "User should be returned")
			AssertEqual(t, user.ID, userIDs[0], "User ID should match")
			AssertEqual(t, user.Username, "johndoe", "Username should match")
		})

		t.Run("InvalidSession", func(t *testing.T) {
			user, err := userService.GetUserBySession("invalid_session_token")
			AssertNotEqual(t, err, nil, "Should fail for invalid session")
			AssertEqual(t, user, nil, "User should not be returned")
		})

		t.Run("EmptySession", func(t *testing.T) {
			user, err := userService.GetUserBySession("")
			AssertNotEqual(t, err, nil, "Should fail for empty session")
			AssertEqual(t, user, nil, "User should not be returned")
		})
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		users, err := userService.GetAllUsers()
		AssertNoError(t, err, "Should retrieve all users")
		AssertGreaterThanOrEqual(t, len(users), len(userIDs), "Should return at least the test users")

		// Verify test users are included
		userIDMap := make(map[int]bool)
		for _, user := range users {
			userIDMap[user.ID] = true
		}

		for _, userID := range userIDs {
			AssertEqual(t, userIDMap[userID], true, "Test user should be included in results")
		}
	})

	t.Run("LogoutUser", func(t *testing.T) {
		// Setup test users
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		// Create session
		sessionToken, err := userService.CreateUserSession(userIDs[0])
		AssertNoError(t, err, "Session creation should succeed")

		// Verify session is valid
		user, err := userService.GetUserBySession(sessionToken)
		AssertNoError(t, err, "Session should be valid before logout")
		AssertNotEqual(t, user, nil, "User should be returned")

		// Logout user by clearing session in database directly
		// (UserService doesn't have LogoutUser method - it's handled in HTTP handler)
		_, err = testDB.DB.Exec("UPDATE user SET current_session = NULL WHERE userid = ?", userIDs[0])
		AssertNoError(t, err, "Logout should succeed")

		// Verify session is no longer valid
		user, err = userService.GetUserBySession(sessionToken)
		AssertNotEqual(t, err, nil, "Session should be invalid after logout")
		AssertEqual(t, user, (*database.User)(nil), "User should not be returned")
	})
}
