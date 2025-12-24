package services

import (
	"fmt"
	"log"
	"regexp"

	"connecthub/database"
	"connecthub/repository"
	"connecthub/security"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// AuthenticateUser validates user credentials and returns user data
func (s *UserService) AuthenticateUser(identifier, password string) (*database.User, error) {
	log.Printf("[DEBUG] UserService: Authenticating user with identifier: %s", maskIdentifier(identifier))

	if identifier == "" {
		return nil, fmt.Errorf("username or email is required. Please enter your username or email address")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required. Please enter your password")
	}

	user, err := s.userRepo.AuthenticateUser(identifier, password)
	if err != nil {
		log.Printf("[WARN] UserService: Authentication failed for %s: %v", maskIdentifier(identifier), err)

		// Check if user exists to provide more specific error message
		if s.isValidEmail(identifier) {
			emailExists, _ := s.userRepo.EmailExists(identifier)
			if !emailExists {
				return nil, fmt.Errorf("we couldn't find an account with that email address. Would you like to sign up instead?")
			}
		} else {
			usernameExists, _ := s.userRepo.UsernameExists(identifier)
			if !usernameExists {
				return nil, fmt.Errorf("we couldn't find an account with that username. Would you like to sign up instead?")
			}
		}

		// If user exists but authentication failed, it's likely a password issue
		return nil, fmt.Errorf("the password you entered is incorrect. Please try again or reset your password if you've forgotten it")
	}

	log.Printf("[INFO] UserService: User authenticated successfully: %s (ID: %d)", user.Username, user.ID)
	return user, nil
}

// CreateUserSession generates a session token and updates user session
func (s *UserService) CreateUserSession(userID int) (string, error) {
	log.Printf("[DEBUG] UserService: Creating session for user ID %d", userID)

	// Generate session token
	sessionTokenUUID, err := security.GenerateToken()
	if err != nil {
		log.Printf("[ERROR] UserService: Failed to generate session token: %v", err)
		return "", fmt.Errorf("session creation failed")
	}

	sessionToken := sessionTokenUUID.String()
	err = s.userRepo.UpdateUserSession(userID, sessionToken)
	if err != nil {
		log.Printf("[ERROR] UserService: Failed to update session for user %d: %v", userID, err)
		return "", fmt.Errorf("session creation failed")
	}

	log.Printf("[INFO] UserService: Session created successfully for user ID %d", userID)
	return sessionToken, nil
}

// RegisterUser creates a new user account with validation
func (s *UserService) RegisterUser(firstName, lastName, username, email, gender, dateOfBirth, password string) (int, error) {
	log.Printf("[DEBUG] UserService: Registering new user: %s (%s)", username, email)

	// Validate required fields with specific error messages
	if firstName == "" {
		return 0, fmt.Errorf("first name is required. Please enter your first name")
	}
	if lastName == "" {
		return 0, fmt.Errorf("last name is required. Please enter your last name")
	}
	if username == "" {
		return 0, fmt.Errorf("username is required. Please choose a username")
	}
	if email == "" {
		return 0, fmt.Errorf("email address is required. Please enter your email address")
	}
	if password == "" {
		return 0, fmt.Errorf("password is required. Please enter your password")
	}

	// Validate email format
	if !s.isValidEmail(email) {
		log.Printf("[WARN] UserService: Invalid email format: %s", email)
		return 0, fmt.Errorf("please enter a valid email address (e.g., user@example.com)")
	}

	// Validate username format
	if !s.isValidUsername(username) {
		log.Printf("[WARN] UserService: Invalid username format: %s", username)
		return 0, fmt.Errorf("username must be 3-20 characters long and contain only letters, numbers, and underscores")
	}

	// Check if user already exists with specific error messages
	exists, err := s.userRepo.UserExists(username, email)
	if err != nil {
		log.Printf("[ERROR] UserService: Failed to check user existence: %v", err)
		return 0, fmt.Errorf("we're experiencing technical difficulties. Please try again in a moment")
	}
	if exists {
		// Check which field already exists for more specific error
		emailExists, _ := s.userRepo.EmailExists(email)
		usernameExists, _ := s.userRepo.UsernameExists(username)

		if emailExists {
			log.Printf("[WARN] UserService: Email already exists: %s", email)
			return 0, fmt.Errorf("an account with this email already exists. Try logging in instead, or use a different email address")
		}
		if usernameExists {
			log.Printf("[WARN] UserService: Username already exists: %s", username)
			return 0, fmt.Errorf("this username is already taken. Please choose a different username")
		}

		log.Printf("[WARN] UserService: User already exists: %s", username)
		return 0, fmt.Errorf("an account with this information already exists. Please check your details")
	}

	// Create user
	userID, err := s.userRepo.CreateUser(firstName, lastName, username, email, gender, dateOfBirth, password)
	if err != nil {
		log.Printf("[ERROR] UserService: Failed to create user: %v", err)
		return 0, fmt.Errorf("failed to create user")
	}

	log.Printf("[INFO] UserService: User %s (ID: %d) created successfully", username, userID)
	return userID, nil
}

// GetUserBySession retrieves user information by session token
func (s *UserService) GetUserBySession(sessionToken string) (*database.User, error) {
	log.Printf("[DEBUG] UserService: Getting user by session token")
	return s.userRepo.GetUserBySession(sessionToken)
}

// GetUserByID retrieves user information by user ID
func (s *UserService) GetUserByID(userID int) (*database.User, error) {
	log.Printf("[DEBUG] UserService: Getting user by ID: %d", userID)
	return s.userRepo.GetUserByID(userID)
}

// GetAllUsers retrieves all users from the database
func (s *UserService) GetAllUsers() ([]database.User, error) {
	log.Printf("[DEBUG] UserService: Getting all users")
	return s.userRepo.GetAllUsers()
}

// ValidateSession checks if a session token is valid and returns user ID
func (s *UserService) ValidateSession(sessionToken string) (int, error) {
	log.Printf("[DEBUG] UserService: Validating session token")
	return s.userRepo.ValidateSession(sessionToken)
}

// Helper methods

// isValidEmail validates email format using regex
func (s *UserService) isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// isValidUsername validates username format using regex
func (s *UserService) isValidUsername(username string) bool {
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	return usernameRegex.MatchString(username)
}

// maskIdentifier masks an identifier for logging purposes
func maskIdentifier(identifier string) string {
	if len(identifier) <= 3 {
		return identifier
	}

	// Check if it's an email
	if regexp.MustCompile(`@`).MatchString(identifier) {
		parts := regexp.MustCompile(`@`).Split(identifier, 2)
		if len(parts) == 2 {
			username := parts[0]
			domain := parts[1]

			if len(username) <= 2 {
				return username + "@" + domain
			}

			maskedUsername := username[:2] + regexp.MustCompile(`.`).ReplaceAllString(username[2:], "*")
			return maskedUsername + "@" + domain
		}
	}

	// For regular usernames
	return identifier[:2] + regexp.MustCompile(`.`).ReplaceAllString(identifier[2:len(identifier)-1], "*") + identifier[len(identifier)-1:]
}
