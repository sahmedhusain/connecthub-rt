package server

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

func ReverseMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		requestPath := r.URL.Path

		log.Printf("[DEBUG] ReverseMiddleware checking authenticated state for %s %s from %s",
			r.Method, requestPath, clientIP)

		db, err := sql.Open("sqlite3", "./database/main.db")
		if err != nil {
			log.Printf("[ERROR] ReverseMiddleware: Database connection failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		seshCok, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("[DEBUG] ReverseMiddleware: No session cookie found for request to %s from %s, proceeding to login page",
				requestPath, clientIP)
			next.ServeHTTP(w, r)
			return
		}

		seshVal := seshCok.Value
		maskedToken := maskSessionToken(seshVal)

		if seshVal == "" {
			log.Printf("[DEBUG] ReverseMiddleware: Empty session cookie value for request to %s from %s, proceeding to login page",
				requestPath, clientIP)
			next.ServeHTTP(w, r)
			return
		}

		var userID int
		var username string
		err = db.QueryRow("SELECT userid, username FROM user WHERE current_session = ?", seshVal).Scan(&userID, &username)
		userContext := ""

		if err == nil {
			userContext = " for user " + username + " (ID: " + strconv.Itoa(userID) + ")"
		}

		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE current_session = ?)", seshVal).Scan(&exists)
		if err != nil {
			log.Printf("[ERROR] ReverseMiddleware: Error checking session existence%s: %v", userContext, err)

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			next.ServeHTTP(w, r)
			return
		}

		if exists {
			log.Printf("[INFO] ReverseMiddleware: Valid session %s found%s, redirecting from %s to /home",
				maskedToken, userContext, requestPath)
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		} else {
			log.Printf("[WARN] ReverseMiddleware: Invalid session %s found in DB from %s, clearing cookie",
				maskedToken, clientIP)

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				Expires:  time.Now().Add(-time.Hour),
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})

			next.ServeHTTP(w, r)
			return
		}
	})
}
