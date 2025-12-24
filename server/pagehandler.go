package server

import (
	"database/sql"
	"encoding/json"
	"connecthub/database"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("[INFO] Processing request for /home route from %s", clientIP)
	log.Printf("[INFO] Serving index.html for /home route from %s", clientIP)
	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
	log.Printf("[DEBUG] Served index.html for /home route to %s", clientIP)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("[INFO] Processing request for / (login) route from %s", clientIP)
	log.Printf("[INFO] Serving index.html for / (login) route from %s", clientIP)
	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
	log.Printf("[DEBUG] Served index.html for / (login) route to %s", clientIP)
}

func NewPostPage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("[INFO] Processing %s request to /newpost from %s", r.Method, clientIP)
	log.Printf("[DEBUG] Handling %s request to /newpost from %s", r.Method, clientIP)

	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[ERROR] No session cookie found for /newpost request from %s: %v", clientIP, err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("[DEBUG] Session cookie retrieved for /newpost request from %s", clientIP)

	maskedToken := maskSessionToken(seshCok.Value)

	log.Printf("[DEBUG] Attempting to connect to SQLite database for /newpost with session %s", maskedToken)
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed for /newpost with session %s: %v", maskedToken, err)
		errData := NewErrorData("500", "Internal Server Error")
		ErrHandler(w, r, errData)
		return
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for /newpost with session %s", maskedToken)

	var userID int
	var userName string
	log.Printf("[DEBUG] Fetching user info for session %s from %s", maskedToken, clientIP)
	err = db.QueryRow("SELECT userid, Username FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID, &userName)
	if err != nil {
		log.Printf("[ERROR] Error fetching user info for session %s from %s: %v", maskedToken, clientIP, err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	log.Printf("[INFO] Retrieved user info for session %s: user %s (ID: %d)", maskedToken, userName, userID)

	switch r.Method {
	case "GET":
		log.Printf("[INFO] Serving /newpost page for user %s (ID: %d)", userName, userID)
		http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
		return

	case "POST":
		log.Printf("[INFO] Processing new post submission from user %s (ID: %d)", userName, userID)

		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			log.Printf("[ERROR] Failed to parse form data: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse form data"})
			return
		}

		content := strings.TrimSpace(r.FormValue("content"))
		title := strings.TrimSpace(r.FormValue("title"))

		if content == "" || title == "" {
			log.Printf("[WARN] Missing content or title in post submission")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Content and title are required"})
			return
		}

		userIDStr := strconv.Itoa(userID)

		postID, err := database.InsertPost(db, content, title, userIDStr)
		if err != nil {
			log.Printf("[ERROR] Failed to create post: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create post"})
			return
		}
		log.Printf("[INFO] Created post ID %d for user %s", postID, userName)

		categories := r.Form["categories"]
		categorySuccess := 0
		for _, categoryIDStr := range categories {
			categoryIDInt, err := strconv.Atoi(categoryIDStr)
			if err != nil {
				log.Printf("[WARN] Invalid category ID: %s", categoryIDStr)
				continue
			}
			err = database.InsertPostCategory(db, postID, categoryIDInt)
			if err != nil {
				log.Printf("[ERROR] Failed to assign category %d to post %d: %v", categoryIDInt, postID, err)
			} else {
				categorySuccess++
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := map[string]interface{}{
			"success":  true,
			"post_id":  postID,
			"message":  "Post created successfully",
			"redirect": "/post?id=" + strconv.Itoa(postID),
		}
		json.NewEncoder(w).Encode(response)
		return

	default:
		log.Printf("[WARN] Method %s not allowed for /newpost from %s", r.Method, clientIP)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func PostPage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	postID := r.URL.Query().Get("id")

	if postID == "" {
		log.Printf("[WARN] Post ID not provided for /post route from %s", clientIP)
	} else {
		log.Printf("[INFO] Serving index.html for /post route, ID: %s from %s", postID, clientIP)
	}

	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
}

func SignupPage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	log.Printf("[INFO] Serving index.html for /signup route from %s", clientIP)
	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
}

func ProfilePage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	userID := r.URL.Query().Get("id")

	if userID == "" {
		log.Printf("[WARN] User ID not provided for /profile route from %s", clientIP)
	} else {
		log.Printf("[INFO] Serving index.html for /profile route, User ID: %s from %s", userID, clientIP)
	}

	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
}

func ChatPage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	conversationID := r.URL.Query().Get("id")

	if conversationID == "" {
		log.Printf("[INFO] Serving chat interface (all conversations) for %s", clientIP)
	} else {
		log.Printf("[INFO] Serving chat interface for conversation ID: %s from %s", conversationID, clientIP)
	}

	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	clientIP := getClientIP(r)
	query := r.URL.Query().Get("q")

	if query == "" {
		log.Printf("[WARN] Empty search query from %s", clientIP)
	} else {
		sanitizedQuery := sanitizeSearchQuery(query)
		log.Printf("[INFO] Serving search results for query: '%s' from %s", sanitizedQuery, clientIP)
	}

	http.ServeFile(w, r, filepath.Join("src", "template", "index.html"))
}
