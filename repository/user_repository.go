package repository

import (
	"database/sql"
	"fmt"
	"log"

	"connecthub/database"
)

// UserRepositoryImpl implements the UserRepository interface
type UserRepositoryImpl struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// AuthenticateUser validates user credentials and returns user data
func (r *UserRepositoryImpl) AuthenticateUser(identifier, password string) (*database.User, error) {
	log.Printf("[DEBUG] UserRepository: Authenticating user with identifier: %s", identifier)
	return database.AuthenticateUser(r.db, identifier, password)
}

// UpdateUserSession updates the user's current session token
func (r *UserRepositoryImpl) UpdateUserSession(userID int, sessionToken string) error {
	log.Printf("[DEBUG] UserRepository: Updating session for user ID %d", userID)
	return database.UpdateUserSession(r.db, userID, sessionToken)
}

// GetUserBySession retrieves user information by session token
func (r *UserRepositoryImpl) GetUserBySession(sessionToken string) (*database.User, error) {
	log.Printf("[DEBUG] UserRepository: Getting user by session token")

	var user database.User
	query := `
		SELECT userid, username, Email, F_name, L_name, date_of_birth, Avatar 
		FROM user 
		WHERE current_session = ?
	`

	err := r.db.QueryRow(query, sessionToken).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &user.Password, &user.Avatar,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] UserRepository: Invalid session token")
			return nil, fmt.Errorf("invalid session")
		}
		log.Printf("[ERROR] UserRepository: Database error during session lookup: %v", err)
		return nil, err
	}

	log.Printf("[INFO] UserRepository: User found for session: %s (ID: %d)", user.Username, user.ID)
	return &user, nil
}

// ValidateSession checks if a session token is valid and returns user ID
func (r *UserRepositoryImpl) ValidateSession(sessionToken string) (int, error) {
	log.Printf("[DEBUG] UserRepository: Validating session token")

	var userID int
	err := r.db.QueryRow("SELECT userid FROM user WHERE current_session = ?", sessionToken).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] UserRepository: Invalid session token")
			return 0, fmt.Errorf("invalid session")
		}
		log.Printf("[ERROR] UserRepository: Database error during session validation: %v", err)
		return 0, err
	}

	log.Printf("[INFO] UserRepository: Session validated for user ID %d", userID)
	return userID, nil
}

// CreateUser creates a new user in the database
func (r *UserRepositoryImpl) CreateUser(firstName, lastName, username, email, gender, dateOfBirth, password string) (int, error) {
	log.Printf("[DEBUG] UserRepository: Creating new user: %s (%s)", username, email)
	return database.CreateUser(r.db, firstName, lastName, username, email, gender, dateOfBirth, password)
}

// GetUserByID retrieves a user by their ID
func (r *UserRepositoryImpl) GetUserByID(userID int) (*database.User, error) {
	log.Printf("[DEBUG] UserRepository: Getting user by ID: %d", userID)

	user, err := database.GetUserByID(r.db, userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAllUsers retrieves all users from the database
func (r *UserRepositoryImpl) GetAllUsers() ([]database.User, error) {
	log.Printf("[DEBUG] UserRepository: Getting all users")
	return database.GetAllUsers(r.db)
}

// UserExists checks if a user with the given username or email already exists
func (r *UserRepositoryImpl) UserExists(username, email string) (bool, error) {
	log.Printf("[DEBUG] UserRepository: Checking if user exists with username: %s or email: %s", username, email)
	return database.UserExists(r.db, username, email)
}

// EmailExists checks if a user with the given email already exists
func (r *UserRepositoryImpl) EmailExists(email string) (bool, error) {
	log.Printf("[DEBUG] UserRepository: Checking if email exists: %s", email)

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM user WHERE Email = ?", email).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] UserRepository: Database error during email existence check: %v", err)
		return false, err
	}

	exists := count > 0
	log.Printf("[DEBUG] UserRepository: Email %s exists: %t", email, exists)
	return exists, nil
}

// UsernameExists checks if a user with the given username already exists
func (r *UserRepositoryImpl) UsernameExists(username string) (bool, error) {
	log.Printf("[DEBUG] UserRepository: Checking if username exists: %s", username)

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM user WHERE username = ?", username).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] UserRepository: Database error during username existence check: %v", err)
		return false, err
	}

	exists := count > 0
	log.Printf("[DEBUG] UserRepository: Username %s exists: %t", username, exists)
	return exists, nil
}
