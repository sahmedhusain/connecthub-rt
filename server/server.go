package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"forum/websocket"
)

// HTTPServer represents the HTTP server with its configuration
type HTTPServer struct {
	router    *mux.Router
	wsManager *websocket.Manager
	port      string
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(port string) *HTTPServer {
	return &HTTPServer{
		router: mux.NewRouter(),
		port:   port,
	}
}

// Initialize sets up the server with all routes and middleware
func (s *HTTPServer) Initialize() error {
	log.Printf("[INFO] Initializing server...")

	// Initialize WebSocket manager
	s.wsManager = websocket.NewManager()
	log.Printf("[INFO] WebSocket manager initialized")

	// Set global WebSocket manager for message handlers
	SetWebSocketManager(s.wsManager)
	log.Printf("[INFO] Global WebSocket manager set for message handlers")

	// Set up database connection for WebSocket operations
	dbConn, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Failed to open database connection for WebSocket: %v", err)
		return fmt.Errorf("failed to open database connection: %v", err)
	}
	websocket.SetDB(dbConn)
	log.Printf("[INFO] Database connection set for WebSocket operations")

	// Configure static file servers
	s.setupStaticRoutes()
	log.Printf("[INFO] Static file servers configured")

	// Register API and page routes
	s.registerAPIRoutes()
	log.Printf("[INFO] API routes registered")

	s.registerPageRoutes()
	log.Printf("[INFO] Page routes registered")

	// Block access to source directory
	s.router.HandleFunc("/src/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WARN] Blocked access attempt to source directory: %s from %s",
			r.URL.Path, getClientIP(r))
		http.NotFound(w, r)
	})

	// Add 404 handler for unmatched routes
	s.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[WARN] 404 Not Found: %s %s from %s", r.Method, r.URL.Path, getClientIP(r))

		// Check if this is an API request
		if strings.HasPrefix(r.URL.Path, "/api/") {
			WriteAPIError(w, http.StatusNotFound, "NOT_FOUND", "API endpoint not found")
			return
		}

		// For page requests, redirect to SPA with 404 error
		errData := NewErrorDataWithType("404", "Page Not Found", "navigation")
		ErrHandler(w, r, errData)
	})

	// Apply logging middleware
	s.router.Use(LoggingMiddleware)
	log.Printf("[INFO] Logging middleware applied to all routes")

	log.Printf("[INFO] Server initialization completed")
	return nil
}

func (s *HTTPServer) setupStaticRoutes() {
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		secureFileServer("./src/static/")))

	s.router.PathPrefix("/js/").Handler(http.StripPrefix("/js/",
		secureFileServer("./src/js/")))

	s.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/",
		secureFileServer("./src/static/assets")))

	s.router.PathPrefix("/src/").Handler(http.StripPrefix("/src/",
		secureFileServer("./src")))
}

func secureFileServer(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)

		fullPath := filepath.Join(root, path)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			clientIP := getClientIP(r)
			log.Printf("[WARN] Blocked directory browsing attempt: %s from %s",
				r.URL.Path, clientIP)

			http.NotFound(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	})
}

// registerAPIRoutes sets up all API endpoints
func (s *HTTPServer) registerAPIRoutes() {
	// Post-related routes
	s.router.HandleFunc("/api/posts", GetPosts)
	s.router.HandleFunc("/api/post", GetPostByID)
	s.router.HandleFunc("/api/categories", CategoriesAPI)
	s.router.HandleFunc("/api/post/create", CreatePostAPI)
	s.router.HandleFunc("/addcomment", AddComment)

	// User-related routes
	s.router.HandleFunc("/api/login", LoginAPI)
	s.router.HandleFunc("/api/signup", SignupAPI)
	s.router.HandleFunc("/api/logout", LogoutAPI)
	s.router.HandleFunc("/api/users", AuthMiddleware(GetUsers))
	s.router.HandleFunc("/api/user/current", AuthMiddleware(GetCurrentUser))

	// Message-related routes
	s.router.HandleFunc("/api/conversations", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			CreateConversationAPI(w, r)
		} else {
			GetConversations(w, r)
		}
	}))
	s.router.HandleFunc("/api/messages", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			SendMessageAPI(w, r)
		} else {
			GetMessages(w, r)
		}
	}))
	s.router.HandleFunc("/api/messages/read", AuthMiddleware(MarkMessagesAsReadAPI))
}

// registerPageRoutes sets up all page endpoints
func (s *HTTPServer) registerPageRoutes() {
	// Public pages
	s.router.HandleFunc("/", LoginPage)
	s.router.HandleFunc("/login", LoginPage)
	s.router.HandleFunc("/signup", SignupPage)
	s.router.HandleFunc("/post", PostPage)

	// Protected pages (require authentication)
	s.router.HandleFunc("/home", AuthMiddleware(HomePage))
	s.router.HandleFunc("/create-post", AuthMiddleware(NewPostPage))
	s.router.HandleFunc("/chat", AuthMiddleware(ChatPage))

	// WebSocket endpoint
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.wsManager.HandleConnection(w, r)
	})
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	serverAddr := ":" + s.port
	log.Printf("[INFO] Server starting on http://localhost%s", serverAddr)
	fmt.Printf("Server running on http://localhost%s\nTo stop the server press Ctrl+C\n", serverAddr)

	return http.ListenAndServe(serverAddr, s.router)
}

// GetRouter returns the server's router (useful for testing)
func (s *HTTPServer) GetRouter() *mux.Router {
	return s.router
}

// GetWebSocketManager returns the WebSocket manager
func (s *HTTPServer) GetWebSocketManager() *websocket.Manager {
	return s.wsManager
}
