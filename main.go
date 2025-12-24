package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	db "connecthub/database"
	"connecthub/server"
)

// Command line flags
var (
	loadTestData = flag.Bool("test-data", false, "Load seed/test data into database")
	serverPort   = flag.String("port", "8080", "Override default port 8080 with custom port")
	resetDB      = flag.Bool("reset", false, "Clear existing database and create fresh empty database")
)

func init() {
	setupLogging()
}

// initializeDatabase handles database initialization based on command line flags
func initializeDatabase() {
	log.Printf("[INFO] Database initialization started")

	// Always initialize database schema
	db.DataBase()
	log.Printf("[INFO] Database schema initialized successfully")

	// Handle reset flag
	if *resetDB {
		log.Printf("[INFO] Reset flag detected - dropping existing database tables")
		db.DropDataBase()
		log.Printf("[INFO] Database tables dropped, reinitializing database")
		db.DataBase()
		log.Printf("[INFO] Database reinitialized successfully")
	}

	// Handle test-data flag or load test data by default if no users exist
	if *loadTestData || shouldLoadTestDataByDefault() {
		log.Printf("[INFO] Loading test data with properly hashed passwords")
		err := db.LoadTestData()
		if err != nil {
			log.Printf("[ERROR] Failed to load test data: %v", err)
		} else {
			log.Printf("[INFO] Test data loaded successfully with hashed passwords")
		}
	}
}

// shouldLoadTestDataByDefault checks if test data should be loaded when no explicit flag is provided
func shouldLoadTestDataByDefault() bool {
	// Only load test data by default if no explicit flags are provided and user table is empty
	if *loadTestData || *resetDB {
		return false // Explicit flags take precedence
	}

	log.Printf("[DEBUG] Checking if test data should be loaded by default")
	dbConn, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed during default test data check: %v", err)
		return false
	}
	defer dbConn.Close()

	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
	if err != nil {
		log.Printf("[WARN] Failed to query user count, not loading test data by default: %v", err)
		return false
	}

	shouldLoad := count == 0
	log.Printf("[INFO] User table has %d records, loading test data by default: %v", count, shouldLoad)
	return shouldLoad
}

func setupLogging() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		err := os.Mkdir("logs", 0755)
		if err != nil {
			log.Printf("Failed to create logs directory: %v", err)
		}
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile, err := os.OpenFile(fmt.Sprintf("logs/forum_%s.log", timestamp),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	// Parse command line flags
	flag.Parse()

	log.Printf("[INFO] Initializing application...")

	// Initialize database
	initializeDatabase()

	// Create and initialize server
	srv := server.NewHTTPServer(*serverPort)
	if err := srv.Initialize(); err != nil {
		log.Fatalf("[FATAL] Failed to initialize server: %v", err)
	}

	// Start server
	log.Fatal(srv.Start())
}
