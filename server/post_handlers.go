package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"connecthub/database"
)

// Post-related request/response types
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreatePostRequest struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
}

type CreatePostResponse struct {
	Success bool   `json:"success"`
	PostID  int    `json:"post_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// GetPosts handles GET /api/posts
func GetPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] GetPosts: Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	filter := r.URL.Query().Get("filter")
	selectedTab := r.URL.Query().Get("tab")

	selectedTab = strings.ReplaceAll(selectedTab, " ", "+")

	log.Printf("[INFO] GetPosts: Raw tab parameter: '%s', filter: '%s'", selectedTab, filter)

	var userID int
	seshCok, err := r.Cookie("session_token")
	if err == nil && seshCok.Value != "" {
		maskedToken := maskSessionToken(seshCok.Value)
		log.Printf("[DEBUG] GetPosts: Retrieving user ID for session %s", maskedToken)
		err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID)
		log.Printf("[INFO] GetPosts: Session token found: %s, userID: %d, err: %v", maskedToken, userID, err)
	} else {
		log.Printf("[INFO] GetPosts: No session token found, userID will be 0")
	}

	log.Printf("[INFO] GetPosts: Selected tab: %s, filter: %s, userID: %d", selectedTab, filter, userID)

	if selectedTab == "" {
		selectedTab = "posts"
	}

	if filter == "" {
		filter = "all"
	}

	var posts []database.Post
	var fetchErr error

	switch selectedTab {
	case "posts":
		if filter == "" || filter == "all" {
			filter = "all"
		}
		switch filter {
		case "all":
			log.Printf("[DEBUG] GetPosts: Fetching all posts")
			posts, fetchErr = database.GetAllPosts(db)
		case "top-rated", "oldest":
			log.Printf("[DEBUG] GetPosts: Fetching posts with filter %s", filter)
			posts, fetchErr = database.GetFilteredPosts(db, filter)
		default:
			log.Printf("[ERROR] Invalid filter '%s' for tab 'posts'", filter)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid filter for posts"})
			return
		}

	case "tags":
		categories, err := database.GetCategories(db)
		if err != nil {
			log.Printf("[ERROR] GetPosts: Fetching categories failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch categories"})
			return
		}

		categoryNames := make([]string, len(categories))
		for i, category := range categories {
			categoryNames[i] = category.Name
		}

		if filter == "" || filter == "all" {
			log.Printf("[DEBUG] GetPosts: Fetching all posts for tags tab with no specific filter")
			posts, fetchErr = database.GetAllPosts(db)
		} else if CheckFilter(filter, categoryNames) {
			log.Printf("[DEBUG] GetPosts: Fetching posts by category %s", filter)
			posts, fetchErr = database.GetPostsByCategory(db, filter)
		} else {
			log.Printf("[ERROR] Invalid category filter '%s' for tab 'tags'", filter)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid category filter"})
			return
		}

	case "my+posts", "your+posts":
		if userID == 0 {
			log.Printf("[WARN] GetPosts: User not authenticated for 'your posts' tab")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}
		log.Printf("[DEBUG] GetPosts: Fetching posts by user ID %d", userID)
		posts, fetchErr = database.GetPostsByUser(db, userID)

	case "liked+posts", "your+replies":
		if userID == 0 {
			log.Printf("[WARN] GetPosts: User not authenticated for 'your replies' tab")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
			return
		}
		log.Printf("[DEBUG] GetPosts: Fetching liked posts by user ID %d", userID)
		posts, fetchErr = database.GetLikedPostsByUser(db, userID)

	default:
		log.Printf("[ERROR] Invalid tab '%s'", selectedTab)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid tab"})
		return
	}

	if fetchErr != nil {
		log.Printf("[ERROR] GetPosts: Fetching posts failed: %v", fetchErr)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch posts"})
		return
	}

	log.Printf("[INFO] GetPosts: Retrieved %d posts for tab '%s' with filter '%s'", len(posts), selectedTab, filter)
	json.NewEncoder(w).Encode(posts)
}

// GetPostByID handles GET /api/post
func GetPostByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		log.Printf("[WARN] GetPostByID: Missing post ID parameter")
		WriteAPIError(w, http.StatusBadRequest, "MISSING_PARAMETER", "Missing post ID")
		return
	}

	postIDInt, err := strconv.Atoi(postIDStr)
	if err != nil {
		log.Printf("[WARN] GetPostByID: Invalid post ID: %s", postIDStr)
		WriteAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER", "Invalid post ID")
		return
	}

	log.Printf("[INFO] GetPostByID: Fetching post with ID %d", postIDInt)

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] GetPostByID: Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database connection failed"})
		return
	}
	defer db.Close()

	post, err := database.GetPostByID(db, postIDInt)
	if err != nil {
		log.Printf("[ERROR] GetPostByID: Fetching post failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch post"})
		return
	}

	comments, err := database.GetCommentsForPost(db, postIDInt)
	if err != nil {
		log.Printf("[ERROR] GetPostByID: Fetching comments failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch comments"})
		return
	}

	categories, err := database.GetCategoriesForPost(db, post.PostID)
	if err != nil {
		log.Printf("[ERROR] GetPostByID: Fetching categories failed: %v", err)
	}

	response := map[string]interface{}{
		"post":       post,
		"comments":   comments,
		"categories": categories,
	}

	json.NewEncoder(w).Encode(response)
}

// CreatePostAPI handles POST /api/post/create
func CreatePostAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clientIP := getClientIP(r)

	if r.Method != "POST" {
		log.Printf("[WARN] CreatePostAPI: Method not allowed: %s from %s", r.Method, clientIP)
		WriteAPIError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	log.Printf("[INFO] CreatePostAPI: Processing create post request from %s", clientIP)

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] CreatePostAPI: Failed to decode request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreatePostResponse{Success: false, Error: "Invalid request format"})
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		log.Printf("[WARN] CreatePostAPI: Missing title or content from %s", clientIP)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CreatePostResponse{Success: false, Error: "Title and content are required"})
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] CreatePostAPI: Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreatePostResponse{Success: false, Error: "Database connection failed"})
		return
	}
	defer db.Close()

	// Get user ID from session
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		log.Printf("[WARN] CreatePostAPI: No session cookie found from %s: %v", clientIP, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(CreatePostResponse{Success: false, Error: "Unauthorized"})
		return
	}

	var userID int
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(CreatePostResponse{Success: false, Error: "Invalid session"})
		return
	}

	// Create post
	postID, err := database.CreatePost(db, userID, req.Title, req.Content, req.Categories)
	if err != nil {
		log.Printf("[ERROR] CreatePostAPI: Failed to create post: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CreatePostResponse{Success: false, Error: "Failed to create post"})
		return
	}

	log.Printf("[INFO] CreatePostAPI: Post created successfully with ID %d by user %d", postID, userID)

	json.NewEncoder(w).Encode(CreatePostResponse{
		Success: true,
		PostID:  postID,
	})
}

// CategoriesAPI handles GET /api/categories
func CategoriesAPI(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] CategoriesAPI: Database connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	categories, err := database.GetCategories(db)
	if err != nil {
		log.Printf("[ERROR] CategoriesAPI: Fetching categories failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// AddComment handles POST /addcomment
func AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postIDStr := r.FormValue("post_id")
	content := r.FormValue("content")

	if postIDStr == "" || strings.TrimSpace(content) == "" {
		http.Error(w, "Missing post ID or content", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] AddComment: Database connection failed: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Get user ID from session
	seshCok, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var userID int
	err = db.QueryRow("SELECT userid FROM user WHERE current_session = ?", seshCok.Value).Scan(&userID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Add comment
	err = database.AddComment(db, postID, userID, content)
	if err != nil {
		log.Printf("[ERROR] AddComment: Failed to add comment: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}

// CheckFilter checks if a filter is valid against a list of valid filters
func CheckFilter(filter string, validFilters []string) bool {
	for _, validFilter := range validFilters {
		if filter == validFilter {
			return true
		}
	}
	return false
}
