package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// maskSessionToken masks a session token for logging purposes to avoid exposing sensitive information.
func maskSessionToken(token string) string {
	if len(token) < 12 {
		log.Printf("[WARN] Invalid token format for masking: length %d", len(token))
		return "invalid-token-format"
	}

	return token[:4] + "..." + token[len(token)-4:]
}

// getClientIP extracts the client IP address from the request.
func getClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// truncateContent shortens content for logging to avoid overly long log entries.
func truncateContent(content string) string {
	if len(content) > 50 {
		return content[:47] + "..."
	}
	return content
}

// sanitizeSearchQuery sanitizes a search query for safe logging and display.
func sanitizeSearchQuery(query string) string {
	if len(query) > 100 {
		query = query[:97] + "..."
	}

	query = strings.Replace(query, "\n", " ", -1)
	query = strings.Replace(query, "\r", " ", -1)

	return query
}

// APIError represents a standardized API error response
type APIError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// APISuccess represents a standardized API success response
type APISuccess struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// WriteAPIError writes a standardized error response to the client
func WriteAPIError(w http.ResponseWriter, statusCode int, errorCode, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIError{
		Success: false,
		Error:   errorMessage,
		Code:    errorCode,
	}

	// Log the error for debugging
	log.Printf("[API_ERROR] %s (%s): %s", errorCode, http.StatusText(statusCode), errorMessage)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode API error response: %v", err)
	}
}

// WriteAPISuccess writes a standardized success response to the client
func WriteAPISuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := APISuccess{
		Success: true,
		Data:    data,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[ERROR] Failed to encode API success response: %v", err)
		WriteAPIError(w, http.StatusInternalServerError, "ENCODING_ERROR", "Failed to encode response")
	}
}
