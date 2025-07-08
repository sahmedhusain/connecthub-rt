package server

import (
	"database/sql"
	UUID "forum/security"
	"log"
	"net/http"
	"strconv"
	"time"
)

func CreateSession(w http.ResponseWriter, r *http.Request, userID int) {
	clientIP := getClientIP(r)
	log.Printf("[DEBUG] Creating new session for user ID %d from %s", userID, clientIP)

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed during session creation for user %d: %v", userID, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}
	defer db.Close()

	var username string
	err = db.QueryRow("SELECT username FROM user WHERE userid = ?", userID).Scan(&username)
	if err != nil {
		log.Printf("[WARN] Unable to retrieve username for user ID %d: %v", userID, err)
		username = "unknown"
	}

	sessionToken, err := UUID.GenerateToken()
	if err != nil {
		log.Printf("[ERROR] Error generating UUID for user %s (ID: %d): %v", username, userID, err)
		errData := NewErrorData("500", "Failed to generate secure session")
		ErrHandler(w, r, errData)
		return
	}

	stringToken := sessionToken.String()
	maskedToken := maskSessionToken(stringToken)
	log.Printf("[DEBUG] Generated session token %s for user %s (ID: %d)", maskedToken, username, userID)

	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction for session creation for user %s (ID: %d): %v",
			username, userID, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}

	sessionExpiry := time.Now().Add(24 * time.Hour)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    stringToken,
		Path:     "/",
		Expires:  sessionExpiry,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
	})

	result, err := tx.Exec(
		"UPDATE session SET sessionid = ?, endtime = ? WHERE userid = ?",
		stringToken, sessionExpiry, userID,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Error updating session for user %s (ID: %d): %v", username, userID, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Error checking rows affected for session update for user %s (ID: %d): %v",
			username, userID, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}

	if rowsAffected == 0 {
		log.Printf("[DEBUG] No existing session found for user %s (ID: %d), creating new session",
			username, userID)

		_, err := tx.Exec(
			"INSERT INTO session (sessionid, userid, endtime) VALUES (?, ?, ?)",
			stringToken, userID, sessionExpiry,
		)
		if err != nil {
			tx.Rollback()
			log.Printf("[ERROR] Error creating new session for user %s (ID: %d): %v", username, userID, err)
			errData := NewErrorData("500", "Internal Server Error")
			ErrHandler(w, r, errData)
			return
		}
	} else {
		log.Printf("[DEBUG] Updated existing session for user %s (ID: %d)", username, userID)
	}

	_, err = tx.Exec(
		"UPDATE user SET current_session = ? WHERE userid = ?",
		stringToken, userID,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Error updating user session reference for user %s (ID: %d): %v",
			username, userID, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction for session creation for user %s (ID: %d): %v",
			username, userID, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}

	log.Printf("[INFO] Session created successfully for user %s (ID: %d), expires at %v",
		username, userID, sessionExpiry.Format(time.RFC3339))
}

func DeleteSession(w http.ResponseWriter, r *http.Request, sessionToken string) {
	clientIP := getClientIP(r)
	maskedToken := maskSessionToken(sessionToken)
	log.Printf("[DEBUG] Deleting session %s from %s", maskedToken, clientIP)

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed during session deletion: %v", err)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction for session deletion: %v", err)
		return
	}

	var userID int
	var username string
	err = tx.QueryRow("SELECT u.userid, u.username FROM user u WHERE u.current_session = ?",
		sessionToken).Scan(&userID, &username)

	userInfo := ""
	if err == nil {
		userInfo = " for user " + username + " (ID: " + strconv.Itoa(userID) + ")"
	}

	result, err := tx.Exec("UPDATE user SET current_session = NULL WHERE current_session = ?", sessionToken)
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to clear session from user table%s: %v", userInfo, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("[DEBUG] Cleared session from %d user records%s", rowsAffected, userInfo)

	result, err = tx.Exec("DELETE FROM session WHERE sessionid = ?", sessionToken)
	if err != nil {
		tx.Rollback()
		log.Printf("[ERROR] Failed to delete session from session table%s: %v", userInfo, err)
		return
	}

	sessionRowsAffected, _ := result.RowsAffected()
	log.Printf("[DEBUG] Deleted %d session records%s", sessionRowsAffected, userInfo)

	if err = tx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction for session deletion%s: %v", userInfo, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	log.Printf("[INFO] Session deleted successfully%s", userInfo)
}

func ValidateSession(r *http.Request) (bool, int, string) {
	sessionCookie, err := r.Cookie("session_token")
	if err != nil {
		return false, 0, ""
	}

	sessionToken := sessionCookie.Value
	maskedToken := maskSessionToken(sessionToken)

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed during session validation: %v", err)
		return false, 0, ""
	}
	defer db.Close()

	var userID int
	var username string
	var expiryTime time.Time

	err = db.QueryRow(`
        SELECT u.userid, u.username, s.endtime 
        FROM user u 
        JOIN session s ON u.current_session = s.sessionid 
        WHERE u.current_session = ?`,
		sessionToken).Scan(&userID, &username, &expiryTime)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("[ERROR] Error validating session %s: %v", maskedToken, err)
		}
		return false, 0, ""
	}

	if time.Now().After(expiryTime) {
		log.Printf("[INFO] Session %s for user %s (ID: %d) is expired", maskedToken, username, userID)
		return false, 0, ""
	}

	return true, userID, username
}
