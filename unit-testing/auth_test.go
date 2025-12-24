package unit_testing

import (
	"fmt"
	"testing"

	"connecthub/repository"
	"connecthub/server/services"
)

func TestUserAuthentication(t *testing.T) {
	testDB := TestSetup(t)

	// Create test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create user repository and service
	userRepo := repository.NewUserRepository(testDB.DB)
	userService := services.NewUserService(userRepo)

	t.Run("ValidUserAuthentication", func(t *testing.T) {
		// Test authentication with valid credentials
		user, err := userService.AuthenticateUser("johndoe", "password123")
		AssertNoError(t, err, "Authentication should succeed with valid credentials")
		AssertNotEqual(t, user, nil, "User should not be nil")
		AssertEqual(t, user.Username, "johndoe", "Username should match")
		AssertEqual(t, user.Email, "john@example.com", "Email should match")
	})

	t.Run("AuthenticationWithEmail", func(t *testing.T) {
		// Test authentication with email instead of username
		user, err := userService.AuthenticateUser("jane@example.com", "password123")
		AssertNoError(t, err, "Authentication should succeed with email")
		AssertNotEqual(t, user, nil, "User should not be nil")
		AssertEqual(t, user.Username, "janesmith", "Username should match")
		AssertEqual(t, user.Email, "jane@example.com", "Email should match")
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		// Test authentication with invalid password
		_, err := userService.AuthenticateUser("johndoe", "wrongpassword")
		AssertError(t, err, "Authentication should fail with invalid password")
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		// Test authentication with nonexistent user
		_, err := userService.AuthenticateUser("nonexistent", "password123")
		AssertError(t, err, "Authentication should fail with nonexistent user")
	})

	t.Run("EmptyCredentials", func(t *testing.T) {
		// Test authentication with empty credentials
		_, err := userService.AuthenticateUser("", "password123")
		AssertError(t, err, "Authentication should fail with empty username")

		_, err = userService.AuthenticateUser("johndoe", "")
		AssertError(t, err, "Authentication should fail with empty password")
	})

	t.Run("SessionCreation", func(t *testing.T) {
		// Test session creation after successful authentication
		sessionToken, err := userService.CreateUserSession(userIDs[0])
		AssertNoError(t, err, "Session creation should succeed")
		AssertNotEqual(t, sessionToken, "", "Session token should not be empty")

		// Verify session token is stored in database
		user, err := userService.GetUserBySession(sessionToken)
		AssertNoError(t, err, "Should be able to retrieve user by session")
		AssertEqual(t, user.ID, userIDs[0], "User ID should match")
	})

	t.Run("InvalidSessionToken", func(t *testing.T) {
		// Test retrieval with invalid session token
		_, err := userService.GetUserBySession("invalid_token")
		AssertError(t, err, "Should fail with invalid session token")
	})
}

func TestUserRegistration(t *testing.T) {
	testDB := TestSetup(t)

	// Create user repository and service
	userRepo := repository.NewUserRepository(testDB.DB)
	userService := services.NewUserService(userRepo)

	t.Run("ValidUserRegistration", func(t *testing.T) {
		// Test user registration with valid data
		userID, err := userService.RegisterUser(
			"Test", "User", "testuser", "test@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User registration should succeed")
		AssertTrue(t, userID > 0, "User ID should be positive")

		// Verify user was created
		user, err := userRepo.GetUserByID(userID)
		AssertNoError(t, err, "Should be able to retrieve created user")
		AssertEqual(t, user.Username, "testuser", "Username should match")
		AssertEqual(t, user.Email, "test@example.com", "Email should match")
	})

	t.Run("DuplicateUsername", func(t *testing.T) {
		// First registration
		_, err := userService.RegisterUser(
			"User", "One", "duplicate", "user1@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "First registration should succeed")

		// Second registration with same username
		_, err = userService.RegisterUser(
			"User", "Two", "duplicate", "user2@example.com",
			"female", "1992-01-01", "password123")
		AssertError(t, err, "Registration should fail with duplicate username")
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		// First registration
		_, err := userService.RegisterUser(
			"User", "One", "user1", "duplicate@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "First registration should succeed")

		// Second registration with same email
		_, err = userService.RegisterUser(
			"User", "Two", "user2", "duplicate@example.com",
			"female", "1992-01-01", "password123")
		AssertError(t, err, "Registration should fail with duplicate email")
	})

	t.Run("InvalidEmailFormat", func(t *testing.T) {
		// Test registration with invalid email format
		_, err := userService.RegisterUser(
			"Test", "User", "testuser2", "invalid-email",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Registration should fail with invalid email format")
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		// Test registration with missing first name
		_, err := userService.RegisterUser(
			"", "User", "testuser3", "test3@example.com",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Registration should fail with missing first name")

		// Test registration with missing last name
		_, err = userService.RegisterUser(
			"Test", "", "testuser4", "test4@example.com",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Registration should fail with missing last name")

		// Test registration with missing username
		_, err = userService.RegisterUser(
			"Test", "User", "", "test5@example.com",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Registration should fail with missing username")

		// Test registration with missing email
		_, err = userService.RegisterUser(
			"Test", "User", "testuser6", "",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Registration should fail with missing email")

		// Test registration with missing password
		_, err = userService.RegisterUser(
			"Test", "User", "testuser7", "test7@example.com",
			"male", "1990-01-01", "")
		AssertError(t, err, "Registration should fail with missing password")
	})

	t.Run("AvatarAssignment", func(t *testing.T) {
		// Test male user avatar assignment
		userID, err := userService.RegisterUser(
			"Male", "User", "maleuser", "male@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "Male user registration should succeed")

		user, err := userRepo.GetUserByID(userID)
		AssertNoError(t, err, "Should be able to retrieve male user")
		AssertEqual(t, user.Avatar.String, "/static/assets/male-avatar-boy-face-man-user-7.svg", "Male avatar should be assigned")

		// Test female user avatar assignment
		userID, err = userService.RegisterUser(
			"Female", "User", "femaleuser", "female@example.com",
			"female", "1990-01-01", "password123")
		AssertNoError(t, err, "Female user registration should succeed")

		user, err = userRepo.GetUserByID(userID)
		AssertNoError(t, err, "Should be able to retrieve female user")
		AssertEqual(t, user.Avatar.String, "/static/assets/female-avatar-girl-face-woman-user-9.svg", "Female avatar should be assigned")
	})
}

func TestAuthLoginAPI(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	_, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create HTTP test helper
	// Note: This would need to be integrated with the actual server setup
	// For now, we'll test the service layer directly

	t.Run("SuccessfulLogin", func(t *testing.T) {
		userRepo := repository.NewUserRepository(testDB.DB)
		userService := services.NewUserService(userRepo)

		// Authenticate user
		user, err := userService.AuthenticateUser("johndoe", "password123")
		AssertNoError(t, err, "Authentication should succeed")

		// Create session
		sessionToken, err := userService.CreateUserSession(user.ID)
		AssertNoError(t, err, "Session creation should succeed")
		AssertNotEqual(t, sessionToken, "", "Session token should not be empty")
	})

	t.Run("FailedLogin", func(t *testing.T) {
		userRepo := repository.NewUserRepository(testDB.DB)
		userService := services.NewUserService(userRepo)

		// Try to authenticate with wrong password
		_, err := userService.AuthenticateUser("johndoe", "wrongpassword")
		AssertError(t, err, "Authentication should fail with wrong password")
	})
}

func TestAuthSignupAPI(t *testing.T) {
	testDB := TestSetup(t)

	t.Run("SuccessfulSignup", func(t *testing.T) {
		userRepo := repository.NewUserRepository(testDB.DB)
		userService := services.NewUserService(userRepo)

		// Register new user
		userID, err := userService.RegisterUser(
			"New", "User", "newuser", "new@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User registration should succeed")
		AssertTrue(t, userID > 0, "User ID should be positive")
	})

	t.Run("SignupWithExistingUser", func(t *testing.T) {
		userRepo := repository.NewUserRepository(testDB.DB)
		userService := services.NewUserService(userRepo)

		// First registration
		_, err := userService.RegisterUser(
			"First", "User", "existinguser", "existing@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "First registration should succeed")

		// Try to register with same username
		_, err = userService.RegisterUser(
			"Second", "User", "existinguser", "different@example.com",
			"female", "1992-01-01", "password123")
		AssertError(t, err, "Second registration should fail with duplicate username")
	})
}

func TestAuthUserRepository(t *testing.T) {
	testDB := TestSetup(t)
	userRepo := repository.NewUserRepository(testDB.DB)

	t.Run("CreateAndRetrieveUser", func(t *testing.T) {
		// Create user
		userID, err := userRepo.CreateUser(
			"Repository", "Test", "repotest", "repo@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")
		AssertTrue(t, userID > 0, "User ID should be positive")

		// Retrieve user by ID
		user, err := userRepo.GetUserByID(userID)
		AssertNoError(t, err, "User retrieval should succeed")
		AssertEqual(t, user.Username, "repotest", "Username should match")
		AssertEqual(t, user.Email, "repo@example.com", "Email should match")
	})

	t.Run("UserExists", func(t *testing.T) {
		// Create user
		_, err := userRepo.CreateUser(
			"Exists", "Test", "existstest", "exists@example.com",
			"female", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Check if user exists by username
		exists, err := userRepo.UserExists("existstest", "")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertTrue(t, exists, "User should exist")

		// Check if user exists by email
		exists, err = userRepo.UserExists("", "exists@example.com")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertTrue(t, exists, "User should exist")

		// Check non-existent user
		exists, err = userRepo.UserExists("nonexistent", "")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertFalse(t, exists, "User should not exist")
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Create multiple users
		for i := 0; i < 3; i++ {
			_, err := userRepo.CreateUser(
				"User", "Test", fmt.Sprintf("user%d", i), fmt.Sprintf("user%d@example.com", i),
				"male", "1990-01-01", "password123")
			AssertNoError(t, err, "User creation should succeed")
		}

		// Get all users
		users, err := userRepo.GetAllUsers()
		AssertNoError(t, err, "GetAllUsers should succeed")
		AssertTrue(t, len(users) >= 3, "Should have at least 3 users")
	})

	t.Run("SessionManagement", func(t *testing.T) {
		// Create user
		userID, err := userRepo.CreateUser(
			"Session", "Test", "sessiontest", "session@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "User creation should succeed")

		// Update session
		sessionToken := "test_session_token"
		err = userRepo.UpdateUserSession(userID, sessionToken)
		AssertNoError(t, err, "Session update should succeed")

		// Retrieve user by session
		user, err := userRepo.GetUserBySession(sessionToken)
		AssertNoError(t, err, "User retrieval by session should succeed")
		AssertEqual(t, user.ID, userID, "User ID should match")

		// Validate session
		validatedUserID, err := userRepo.ValidateSession(sessionToken)
		AssertNoError(t, err, "Session validation should succeed")
		AssertEqual(t, validatedUserID, userID, "Validated user ID should match")
	})
}

func TestUserManagementAPI(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	userRepo := repository.NewUserRepository(testDB.DB)
	userService := services.NewUserService(userRepo)

	t.Run("GetCurrentUser", func(t *testing.T) {
		// Create session for first user
		sessionToken, err := userService.CreateUserSession(userIDs[0])
		AssertNoError(t, err, "Session creation should succeed")

		// Get user by session (simulating GetCurrentUser API)
		user, err := userService.GetUserBySession(sessionToken)
		AssertNoError(t, err, "Should be able to get current user")
		AssertEqual(t, user.ID, userIDs[0], "User ID should match")
		AssertEqual(t, user.Username, "johndoe", "Username should match")
	})

	t.Run("GetAllUsers", func(t *testing.T) {
		// Get all users (simulating GetUsers API)
		users, err := userRepo.GetAllUsers()
		AssertNoError(t, err, "Should be able to get all users")
		AssertTrue(t, len(users) >= 5, "Should have at least 5 test users")

		// Verify user data structure
		for _, user := range users {
			AssertTrue(t, user.ID > 0, "User ID should be positive")
			AssertNotEqual(t, user.Username, "", "Username should not be empty")
			AssertNotEqual(t, user.Email, "", "Email should not be empty")
			AssertNotEqual(t, user.FirstName, "", "First name should not be empty")
			AssertNotEqual(t, user.LastName, "", "Last name should not be empty")
		}
	})

	t.Run("UserProfileData", func(t *testing.T) {
		// Get specific user
		user, err := userRepo.GetUserByID(userIDs[0])
		AssertNoError(t, err, "Should be able to get user by ID")

		// Verify profile data completeness
		AssertEqual(t, user.Username, "johndoe", "Username should match")
		AssertEqual(t, user.Email, "john@example.com", "Email should match")
		AssertEqual(t, user.FirstName, "John", "First name should match")
		AssertEqual(t, user.LastName, "Doe", "Last name should match")
		AssertEqual(t, user.Gender, "male", "Gender should match")
		AssertNotEqual(t, user.Avatar.String, "", "Avatar should be set")
	})

	t.Run("SessionValidation", func(t *testing.T) {
		// Create session
		sessionToken, err := userService.CreateUserSession(userIDs[1])
		AssertNoError(t, err, "Session creation should succeed")

		// Validate session
		validatedUserID, err := userRepo.ValidateSession(sessionToken)
		AssertNoError(t, err, "Session validation should succeed")
		AssertEqual(t, validatedUserID, userIDs[1], "Validated user ID should match")

		// Test invalid session
		_, err = userRepo.ValidateSession("invalid_session_token")
		AssertError(t, err, "Invalid session should fail validation")
	})

	t.Run("UserExistenceChecks", func(t *testing.T) {
		// Check existing user by username
		exists, err := userRepo.UserExists("johndoe", "")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertTrue(t, exists, "User should exist")

		// Check existing user by email
		exists, err = userRepo.UserExists("", "jane@example.com")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertTrue(t, exists, "User should exist")

		// Check non-existent user
		exists, err = userRepo.UserExists("nonexistentuser", "")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertFalse(t, exists, "User should not exist")

		// Check non-existent email
		exists, err = userRepo.UserExists("", "nonexistent@example.com")
		AssertNoError(t, err, "UserExists check should succeed")
		AssertFalse(t, exists, "User should not exist")
	})
}

func TestPasswordSecurity(t *testing.T) {
	testDB := TestSetup(t)
	userRepo := repository.NewUserRepository(testDB.DB)
	userService := services.NewUserService(userRepo)

	t.Run("PasswordHashing", func(t *testing.T) {
		// Create user with password
		userID, err := userService.RegisterUser(
			"Password", "Test", "passwordtest", "password@example.com",
			"male", "1990-01-01", "plainpassword")
		AssertNoError(t, err, "User registration should succeed")

		// Verify password is hashed in database
		var storedPassword string
		err = testDB.DB.QueryRow("SELECT password FROM user WHERE userid = ?", userID).Scan(&storedPassword)
		AssertNoError(t, err, "Should be able to query stored password")
		AssertNotEqual(t, storedPassword, "plainpassword", "Password should be hashed")
		AssertTrue(t, len(storedPassword) > 20, "Hashed password should be longer than plain password")

		// Verify authentication works with plain password
		user, err := userService.AuthenticateUser("passwordtest", "plainpassword")
		AssertNoError(t, err, "Authentication should work with plain password")
		AssertEqual(t, user.ID, userID, "User ID should match")
	})

	t.Run("PasswordValidation", func(t *testing.T) {
		// Test various password scenarios
		testCases := []struct {
			password    string
			shouldPass  bool
			description string
		}{
			{"password123", true, "Normal password should work"},
			{"", false, "Empty password should fail"},
			{"a", true, "Single character password should work (no minimum length enforced)"}, // Adjust based on actual requirements
		}

		for _, tc := range testCases {
			_, err := userService.RegisterUser(
				"Test", "User", fmt.Sprintf("user_%s", tc.password), fmt.Sprintf("test_%s@example.com", tc.password),
				"male", "1990-01-01", tc.password)

			if tc.shouldPass {
				AssertNoError(t, err, tc.description)
			} else {
				AssertError(t, err, tc.description)
			}
		}
	})
}

func TestUserDataValidation(t *testing.T) {
	testDB := TestSetup(t)
	userRepo := repository.NewUserRepository(testDB.DB)
	userService := services.NewUserService(userRepo)

	t.Run("EmailValidation", func(t *testing.T) {
		testCases := []struct {
			email       string
			shouldPass  bool
			description string
		}{
			{"valid@example.com", true, "Valid email should pass"},
			{"user.name@domain.co.uk", true, "Complex valid email should pass"},
			{"invalid-email", false, "Invalid email format should fail"},
			{"@example.com", false, "Email without username should fail"},
			{"user@", false, "Email without domain should fail"},
			{"", false, "Empty email should fail"},
		}

		for i, tc := range testCases {
			_, err := userService.RegisterUser(
				"Test", "User", fmt.Sprintf("emailtest%d", i), tc.email,
				"male", "1990-01-01", "password123")

			if tc.shouldPass {
				AssertNoError(t, err, tc.description)
			} else {
				AssertError(t, err, tc.description)
			}
		}
	})

	t.Run("UsernameValidation", func(t *testing.T) {
		testCases := []struct {
			username    string
			shouldPass  bool
			description string
		}{
			{"validuser", true, "Valid username should pass"},
			{"user123", true, "Username with numbers should pass"},
			{"", false, "Empty username should fail"},
		}

		for i, tc := range testCases {
			_, err := userService.RegisterUser(
				"Test", "User", tc.username, fmt.Sprintf("usernametest%d@example.com", i),
				"male", "1990-01-01", "password123")

			if tc.shouldPass {
				AssertNoError(t, err, tc.description)
			} else {
				AssertError(t, err, tc.description)
			}
		}
	})

	t.Run("NameValidation", func(t *testing.T) {
		// Test first name validation
		_, err := userService.RegisterUser(
			"", "User", "nametest1", "nametest1@example.com",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Empty first name should fail")

		// Test last name validation
		_, err = userService.RegisterUser(
			"Test", "", "nametest2", "nametest2@example.com",
			"male", "1990-01-01", "password123")
		AssertError(t, err, "Empty last name should fail")

		// Test valid names
		_, err = userService.RegisterUser(
			"Valid", "Name", "nametest3", "nametest3@example.com",
			"male", "1990-01-01", "password123")
		AssertNoError(t, err, "Valid names should pass")
	})
}
