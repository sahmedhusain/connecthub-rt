package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"connecthub/repository"
	"connecthub/server/services"
)

// User-related request/response types
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Success     bool   `json:"success"`
	UserID      int    `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Gender      string `json:"gender,omitempty"`
	DateOfBirth string `json:"dateOfBirth,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Error       string `json:"error,omitempty"`
}

type SignupRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
	DateOfBirth string `json:"dateOfBirth"`
	Password    string `json:"password"`
}

type SignupResponse struct {
	Success     bool   `json:"success"`
	UserID      int    `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"`
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Gender      string `json:"gender,omitempty"`
	DateOfBirth string `json:"dateOfBirth,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	Error       string `json:"error,omitempty"`
}

// LoginAPI handles POST /api/login
func LoginAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientIP := getClientIP(r)

	if r.Method != "POST" {
		log.Printf("[WARN] LoginAPI: Login attempt with invalid method: %s from %s", r.Method, clientIP)
		WriteAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	log.Printf("[INFO] LoginAPI: Processing login request from %s", clientIP)

	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		log.Printf("[ERROR] LoginAPI: Failed to decode login request from %s: %v", clientIP, err)
		WriteAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request format")
		return
	}

	// Open database connection
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] LoginAPI: Database connection failed during login: %v", err)
		WriteAPIError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Internal server error")
		return
	}
	defer db.Close()

	// Create user repository and service
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	// Authenticate user using service
	user, err := userService.AuthenticateUser(loginReq.Identifier, loginReq.Password)
	if err != nil {
		log.Printf("[WARN] LoginAPI: Authentication failed for %s from %s: %v", loginReq.Identifier, clientIP, err)

		// Use the enhanced error message from the service
		errorCode := "INVALID_CREDENTIALS"
		if strings.Contains(err.Error(), "couldn't find an account") {
			errorCode = "USER_NOT_FOUND"
		} else if strings.Contains(err.Error(), "password you entered is incorrect") {
			errorCode = "INCORRECT_PASSWORD"
		}

		WriteAPIError(w, http.StatusUnauthorized, errorCode, err.Error())
		return
	}

	// Create session using service
	sessionToken, err := userService.CreateUserSession(user.ID)
	if err != nil {
		log.Printf("[ERROR] LoginAPI: Failed to create session for user %d: %v", user.ID, err)
		WriteAPIError(w, http.StatusInternalServerError, "SESSION_ERROR", "Session creation failed")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	// Handle avatar field properly
	avatarStr := ""
	if user.Avatar.Valid {
		avatarStr = user.Avatar.String
	}

	log.Printf("[INFO] LoginAPI: User logged in successfully: %s (ID: %d)", user.Username, user.ID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Success:     true,
		UserID:      user.ID,
		Username:    user.Username,
		Email:       loginReq.Identifier,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Gender:      user.Gender,
		DateOfBirth: user.DateOfBirth,
		Avatar:      avatarStr,
	})
}

// SignupAPI handles POST /api/signup
func SignupAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientIP := getClientIP(r)

	if r.Method != "POST" {
		log.Printf("[WARN] SignupAPI: Signup attempt with invalid method: %s from %s", r.Method, clientIP)
		WriteAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	log.Printf("[INFO] SignupAPI: Processing signup request from %s", clientIP)

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[WARN] SignupAPI: Invalid JSON from %s: %v", clientIP, err)
		WriteAPIError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request format")
		return
	}

	// Open database connection
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] SignupAPI: Database connection failed: %v", err)
		WriteAPIError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database connection failed")
		return
	}
	defer db.Close()

	// Create user repository and service
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	// Register user using service (includes validation)
	userID, err := userService.RegisterUser(req.FirstName, req.LastName, req.Username, req.Email, req.Gender, req.DateOfBirth, req.Password)
	if err != nil {
		log.Printf("[WARN] SignupAPI: Registration failed from %s: %v", clientIP, err)

		// Determine appropriate status code and error code based on enhanced error messages
		statusCode := http.StatusBadRequest
		errorCode := "VALIDATION_ERROR"
		errorMessage := err.Error()

		if strings.Contains(errorMessage, "email already exists") {
			statusCode = http.StatusConflict
			errorCode = "EMAIL_EXISTS"
		} else if strings.Contains(errorMessage, "username is already taken") {
			statusCode = http.StatusConflict
			errorCode = "USERNAME_EXISTS"
		} else if strings.Contains(errorMessage, "technical difficulties") {
			statusCode = http.StatusInternalServerError
			errorCode = "DATABASE_ERROR"
		} else if strings.Contains(errorMessage, "required") {
			errorCode = "MISSING_FIELD"
		} else if strings.Contains(errorMessage, "valid email") {
			errorCode = "INVALID_EMAIL"
		} else if strings.Contains(errorMessage, "username must be") {
			errorCode = "INVALID_USERNAME"
		}

		WriteAPIError(w, statusCode, errorCode, errorMessage)
		return
	}

	// Create session for the new user automatically
	sessionToken, err := userService.CreateUserSession(userID)
	if err != nil {
		log.Printf("[ERROR] SignupAPI: Failed to create session for new user %d: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SignupResponse{Success: false, Error: "Session creation failed"})
		return
	}

	// Get the created user to retrieve avatar information
	user, err := userService.GetUserByID(userID)
	if err != nil {
		log.Printf("[ERROR] SignupAPI: Failed to retrieve user data for new user %d: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SignupResponse{Success: false, Error: "Failed to retrieve user data"})
		return
	}

	// Handle avatar field properly
	avatarStr := ""
	if user.Avatar.Valid {
		avatarStr = user.Avatar.String
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	log.Printf("[INFO] SignupAPI: User %s (ID: %d) created successfully with session from %s", req.Username, userID, clientIP)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignupResponse{
		Success:     true,
		UserID:      userID,
		Username:    req.Username,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Gender:      req.Gender,
		DateOfBirth: req.DateOfBirth,
		Avatar:      avatarStr,
	})
}

// LogoutAPI handles POST /api/logout
func LogoutAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientIP := getClientIP(r)

	if r.Method != "POST" {
		log.Printf("[WARN] LogoutAPI: Logout attempt with invalid method: %s from %s", r.Method, clientIP)
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	log.Printf("[INFO] LogoutAPI: Processing logout request from %s", clientIP)

	// Get session cookie
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] LogoutAPI: No session cookie found from %s: %v", clientIP, err)
		// Still clear cookie and return success even if no session
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Logged out successfully",
		})
		return
	}

	// Connect to database to clear session
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] LogoutAPI: Database connection failed: %v", err)
		// Still clear cookie even if database fails
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Logged out successfully",
		})
		return
	}
	defer db.Close()

	// Clear session from database
	maskedToken := maskSessionToken(sessionCookie.Value)
	log.Printf("[DEBUG] LogoutAPI: Clearing session %s from database", maskedToken)

	_, err = db.Exec("UPDATE user SET current_session = NULL WHERE current_session = ?", sessionCookie.Value)
	if err != nil {
		log.Printf("[ERROR] LogoutAPI: Failed to clear session %s from database: %v", maskedToken, err)
	} else {
		log.Printf("[DEBUG] LogoutAPI: Successfully cleared session %s from database", maskedToken)
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	log.Printf("[INFO] LogoutAPI: User logged out successfully from %s", clientIP)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// GetUsers handles GET /api/users
func GetUsers(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] GetUsers: Database connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Create user repository and service
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	// Get all users using service
	users, err := userService.GetAllUsers()
	if err != nil {
		log.Printf("[ERROR] GetUsers: Fetching users failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetCurrentUser handles GET /api/user/current
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] GetCurrentUser: No session cookie from %s: %v", clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No session"})
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] GetCurrentUser: Database connection error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Internal server error"})
		return
	}
	defer db.Close()

	// Create user repository and service
	userRepo := repository.NewUserRepository(db)
	userService := services.NewUserService(userRepo)

	// Get user by session using service
	user, err := userService.GetUserBySession(sessionCookie.Value)
	if err != nil {
		log.Printf("[WARN] GetCurrentUser: Invalid session from %s: %v", clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid session"})
		return
	}

	avatarStr := ""
	if user.Avatar.Valid {
		avatarStr = user.Avatar.String
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"userId":      user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"avatar":      avatarStr,
		"firstName":   user.FirstName,
		"lastName":    user.LastName,
		"gender":      user.Gender,
		"dateOfBirth": user.DateOfBirth,
	})
}
