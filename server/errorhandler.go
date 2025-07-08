package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
)

// Enhanced Error Messages - User-friendly, actionable messages
const (
	// Authentication Error Messages
	ErrInvalidCredentials = "The username/email or password you entered is incorrect. Please check your credentials and try again."
	ErrUserNotFound       = "We couldn't find an account with that username or email. Would you like to sign up instead?"
	ErrIncorrectPassword  = "The password you entered is incorrect. Please try again or reset your password if you've forgotten it."
	ErrAccountExists      = "An account with this email already exists. Try logging in instead, or use a different email address."
	ErrUsernameExists     = "This username is already taken. Please choose a different username."
	ErrSessionExpired     = "Your session has expired for security reasons. Please log in again to continue."
	ErrSessionInvalid     = "Your session is no longer valid. Please log in again to access this content."
	ErrAuthRequired       = "You need to be logged in to access this content. Please log in to continue."
	ErrAccessDenied       = "You don't have permission to access this content. Contact support if you believe this is an error."

	// Form Validation Error Messages
	ErrEmailRequired     = "Email address is required. Please enter your email address."
	ErrEmailInvalid      = "Please enter a valid email address (e.g., user@example.com)."
	ErrPasswordRequired  = "Password is required. Please enter your password."
	ErrPasswordTooShort  = "Password must be at least 8 characters long. Please choose a stronger password."
	ErrPasswordWeak      = "Password must contain at least one uppercase letter, one lowercase letter, and one number."
	ErrUsernameRequired  = "Username is required. Please choose a username."
	ErrUsernameInvalid   = "Username must be 3-20 characters long and contain only letters, numbers, and underscores."
	ErrFirstNameRequired = "First name is required. Please enter your first name."
	ErrFirstNameInvalid  = "First name should contain only letters and be 2-30 characters long."
	ErrLastNameRequired  = "Last name is required. Please enter your last name."
	ErrLastNameInvalid   = "Last name should contain only letters and be 2-30 characters long."
	ErrPasswordMismatch  = "Passwords don't match. Please make sure both password fields are identical."

	// Content Validation Error Messages
	ErrTitleRequired      = "Post title is required. Please enter a title for your post."
	ErrTitleTooLong       = "Post title is too long. Please keep it under 200 characters."
	ErrContentRequired    = "Post content is required. Please write some content for your post."
	ErrContentTooLong     = "Post content is too long. Please keep it under 10,000 characters."
	ErrCategoriesRequired = "Please select at least one category for your post."
	ErrCommentRequired    = "Comment content is required. Please write your comment."
	ErrCommentTooLong     = "Comment is too long. Please keep it under 1,000 characters."

	// System Error Messages
	ErrDatabaseConnection = "We're experiencing technical difficulties. Please try again in a moment."
	ErrDatabaseOperation  = "Something went wrong while processing your request. Please try again."
	ErrServerError        = "An unexpected error occurred. Our team has been notified. Please try again later."
	ErrNetworkError       = "Network connection failed. Please check your internet connection and try again."
	ErrFileUploadFailed   = "File upload failed. Please check the file size and format, then try again."
	ErrFileTooBig         = "File is too large. Please choose a file smaller than 5MB."
	ErrFileInvalidFormat  = "Invalid file format. Please upload a JPEG, PNG, or GIF image."

	// API Error Messages
	ErrAPINotFound         = "The requested resource was not found. Please check the URL and try again."
	ErrAPIMethodNotAllowed = "This action is not allowed. Please use the correct method for this request."
	ErrAPIRateLimit        = "Too many requests. Please wait a moment before trying again."
	ErrAPIInvalidRequest   = "Invalid request format. Please check your input and try again."
	ErrAPIUnauthorized     = "Authentication required. Please log in to access this resource."
	ErrAPIForbidden        = "Access denied. You don't have permission to perform this action."

	// Chat and Messaging Error Messages
	ErrMessageSendFailed    = "Failed to send message. Please check your connection and try again."
	ErrRecipientOffline     = "The recipient is currently offline. Your message will be delivered when they come online."
	ErrConversationNotFound = "Conversation not found. It may have been deleted or you don't have access to it."
	ErrMessageTooLong       = "Message is too long. Please keep it under 1,000 characters."
	ErrInvalidRecipient     = "Invalid recipient. Please select a valid user to send the message to."

	// General User-Friendly Messages
	ErrGenericBadRequest  = "There was a problem with your request. Please check your input and try again."
	ErrGenericNotFound    = "The page or resource you're looking for doesn't exist. Please check the URL."
	ErrGenericServerError = "Something went wrong on our end. Please try again in a few minutes."
	ErrGenericTimeout     = "The request took too long to process. Please try again."
)

type ErrorPageData struct {
	Code     string `json:"code"`
	ErrorMsg string `json:"message"`
	Path     string `json:"path,omitempty"`
	Source   string `json:"source,omitempty"`
	Type     string `json:"type,omitempty"` // "navigation", "authentication", "api", "critical"
}

func ErrHandler(w http.ResponseWriter, r *http.Request, errData *ErrorPageData) {
	if errData.Path == "" {
		errData.Path = r.URL.Path
	}

	// Set default error type if not specified
	if errData.Type == "" {
		errData.Type = "navigation"
	}

	pc, file, line, ok := runtime.Caller(1)
	source := "unknown"
	if ok {
		parts := strings.Split(file, "/")
		funcName := runtime.FuncForPC(pc).Name()
		funcParts := strings.Split(funcName, ".")
		source = fmt.Sprintf("%s:%d in %s", parts[len(parts)-1], line, funcParts[len(funcParts)-1])
		errData.Source = source
	}

	log.Printf("[ERROR] %s error (%s): %s from %s (Request: %s %s, User-Agent: %s, Type: %s)",
		errData.Code,
		source,
		errData.ErrorMsg,
		getClientIP(r),
		r.Method,
		r.URL.Path,
		r.UserAgent(),
		errData.Type,
	)

	// Check if this is an API request
	if strings.HasPrefix(r.URL.Path, "/api/") {
		// For API requests, return JSON error response
		WriteAPIError(w, getStatusCodeFromErrorCode(errData.Code), errData.Code, errData.ErrorMsg)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	redirectURL := fmt.Sprintf("%s://%s/#/error?code=%s&message=%s&path=%s&type=%s",
		scheme,
		host,
		url.QueryEscape(errData.Code),
		url.QueryEscape(errData.ErrorMsg),
		url.QueryEscape(errData.Path),
		url.QueryEscape(errData.Type),
	)
	log.Printf("[ERROR] Redirecting to: %s", redirectURL)
	w.Header().Set("Location", redirectURL)
	w.WriteHeader(http.StatusSeeOther)
}

func AutherrHandler(w http.ResponseWriter, r *http.Request, errData *ErrorPageData) {
	pc, file, line, ok := runtime.Caller(1)
	source := "unknown"
	if ok {
		parts := strings.Split(file, "/")
		funcName := runtime.FuncForPC(pc).Name()
		funcParts := strings.Split(funcName, ".")
		source = fmt.Sprintf("%s:%d in %s", parts[len(parts)-1], line, funcParts[len(funcParts)-1])
	}

	log.Printf("[WARN] Authentication error (%s): %s from %s (Request: %s %s, User-Agent: %s)",
		source,
		errData.ErrorMsg,
		getClientIP(r),
		r.Method,
		r.URL.Path,
		r.UserAgent(),
	)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewErrorData(code, message string) *ErrorPageData {
	return &ErrorPageData{
		Code:     code,
		ErrorMsg: message,
	}
}

// NewErrorDataWithType creates a new ErrorPageData with specified type
func NewErrorDataWithType(code, message, errorType string) *ErrorPageData {
	return &ErrorPageData{
		Code:     code,
		ErrorMsg: message,
		Type:     errorType,
	}
}

// getStatusCodeFromErrorCode converts error code to HTTP status code
func getStatusCodeFromErrorCode(code string) int {
	switch code {
	case "400", "INVALID_JSON", "INVALID_PARAMETER", "MISSING_PARAMETER", "VALIDATION_ERROR", "MISSING_FIELD", "INVALID_EMAIL", "INVALID_USERNAME":
		return http.StatusBadRequest
	case "401", "INVALID_CREDENTIALS", "INVALID_SESSION", "USER_NOT_FOUND", "INCORRECT_PASSWORD":
		return http.StatusUnauthorized
	case "403", "FORBIDDEN":
		return http.StatusForbidden
	case "404", "NOT_FOUND":
		return http.StatusNotFound
	case "405", "METHOD_NOT_ALLOWED":
		return http.StatusMethodNotAllowed
	case "409", "DUPLICATE_ENTRY", "EMAIL_EXISTS", "USERNAME_EXISTS":
		return http.StatusConflict
	case "429", "RATE_LIMITED":
		return http.StatusTooManyRequests
	case "500", "DATABASE_ERROR", "SESSION_ERROR", "ENCODING_ERROR":
		return http.StatusInternalServerError
	default:
		// Try to parse as integer
		if statusCode, err := strconv.Atoi(code); err == nil && statusCode >= 100 && statusCode < 600 {
			return statusCode
		}
		return http.StatusInternalServerError
	}
}
