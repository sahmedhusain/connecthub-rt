package server

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		requestPath := r.URL.Path

		log.Printf("[INFO] Starting authentication check for request: %s %s from %s", r.Method, requestPath, clientIP)
		log.Printf("[DEBUG] Auth check for request: %s %s from %s", r.Method, requestPath, clientIP)

		log.Printf("[DEBUG] Attempting to connect to SQLite database for auth check")
		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Printf("[ERROR] Database connection failed during auth check: %v", err)
			errData := NewErrorData("500", "Internal Server Error")
			ErrHandler(w, r, errData)
			return
		}
		defer db.Close()
		log.Printf("[INFO] Successfully connected to SQLite database for auth check")

		sessionCookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("[WARN] No session cookie found for request to %s from %s: %v", requestPath, clientIP, err)
			errData := NewErrorData("401", "Authentication Required")
			log.Printf("[INFO] Redirecting to authentication due to missing session cookie for %s from %s", requestPath, clientIP)
			AutherrHandler(w, r, errData)
			return
		}
		log.Printf("[DEBUG] Session cookie retrieved for request to %s from %s", requestPath, clientIP)

		sessionToken := sessionCookie.Value
		maskedToken := maskSessionToken(sessionToken)

		if sessionToken == "" {
			log.Printf("[WARN] Empty session token for request to %s from %s", requestPath, clientIP)
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			errData := NewErrorData("401", "Invalid Session")
			log.Printf("[INFO] Redirecting to authentication due to empty session token for %s from %s", requestPath, clientIP)
			AutherrHandler(w, r, errData)
			return
		}
		log.Printf("[DEBUG] Session token %s retrieved for request to %s from %s", maskedToken, requestPath, clientIP)

		var userID int
		var username string
		log.Printf("[DEBUG] Validating session token %s in database", maskedToken)
		err = db.QueryRow("SELECT userid, username FROM user WHERE current_session = ?", sessionToken).Scan(&userID, &username)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("[WARN] Invalid session token %s for request to %s from %s", maskedToken, requestPath, clientIP)
				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    "",
					Path:     "/",
					Expires:  time.Now().Add(-time.Hour),
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				})
				errData := NewErrorData("401", "Invalid Session")
				log.Printf("[INFO] Redirecting to authentication due to invalid session token %s for %s from %s", maskedToken, requestPath, clientIP)
				AutherrHandler(w, r, errData)
				return
			}

			log.Printf("[ERROR] Database error while validating session %s: %v", maskedToken, err)
			errData := NewErrorData("500", "Internal Server Error")
			log.Printf("[INFO] Returning internal server error due to database issue for session validation from %s", clientIP)
			ErrHandler(w, r, errData)
			return
		}
		log.Printf("[INFO] Session token %s validated for user %s (ID: %d)", maskedToken, username, userID)

		// Session is valid (already validated above by checking user table)
		log.Printf("[INFO] Session token %s is valid for user %s (ID: %d)", maskedToken, username, userID)

		log.Printf("[DEBUG] Authentication successful for user %s (ID: %d) accessing %s",
			username, userID, requestPath)

		log.Printf("[INFO] Proceeding to next handler for authenticated request %s %s from %s", r.Method, requestPath, clientIP)
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		clientIP := getClientIP(r)
		requestPath := r.URL.Path

		next.ServeHTTP(w, r)

		duration := time.Since(startTime)

		// Only log non-static requests
		if !strings.HasPrefix(r.URL.Path, "/static/") && !strings.HasPrefix(r.URL.Path, "/favicon.ico") {
			log.Printf("[INFO] %s - %s %s - %v", clientIP, r.Method, requestPath, duration)
		}
	})
}
