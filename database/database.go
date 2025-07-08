package database

import (
	"database/sql"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func DataBase() {
	log.Printf("[DEBUG] Attempting to connect to SQLite database at ./database/main.db")
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal("[FATAL] Failed to connect to the database: ", err)
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database")

	createTables := []string{
		`
		CREATE TABLE IF NOT EXISTS categories (
			idcategories INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);`,

		`
		CREATE TABLE IF NOT EXISTS comment (
			commentid INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NULL,
			comment_at DATETIME NULL,
			post_postid INTEGER NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`,

		`
		CREATE TABLE IF NOT EXISTS post (
			postid INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NULL,
			title  TEXT NULL,
			post_at DATETIME NOT NULL,
			user_userid INTEGER NOT NULL,
			FOREIGN KEY (user_userid) REFERENCES user(userid)
		);`,

		`
		CREATE TABLE IF NOT EXISTS post_has_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			post_postid INTEGER NOT NULL,
			categories_idcategories INTEGER NOT NULL,
			FOREIGN KEY (post_postid) REFERENCES post(postid),
			FOREIGN KEY (categories_idcategories) REFERENCES categories(idcategories)
		);`,

		`
		CREATE TABLE IF NOT EXISTS session (
			sessionid TEXT PRIMARY KEY,
			userid INTEGER NOT NULL UNIQUE,
			endtime DATETIME NOT NULL,
			FOREIGN KEY (userid) REFERENCES user(userid)
		);`,

		`
		CREATE TABLE IF NOT EXISTS user (
			userid INTEGER PRIMARY KEY AUTOINCREMENT,
			F_name TEXT NOT NULL,
			L_name TEXT NOT NULL,
			Username TEXT NOT NULL UNIQUE,
			Email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			current_session TEXT,
			Avatar TEXT,
			gender TEXT,
			date_of_birth DATE,
			FOREIGN KEY (current_session) REFERENCES session(sessionid)
		);`,

		`
		CREATE TABLE IF NOT EXISTS conversation (
			conversation_id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		`
		CREATE TABLE IF NOT EXISTS conversation_participants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			FOREIGN KEY (conversation_id) REFERENCES conversation(conversation_id),
			FOREIGN KEY (user_id) REFERENCES user(userid),
			UNIQUE(conversation_id, user_id)
		);`,

		`
		CREATE TABLE IF NOT EXISTS message (
			message_id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			is_read BOOLEAN NOT NULL DEFAULT 0,
			FOREIGN KEY (conversation_id) REFERENCES conversation(conversation_id),
			FOREIGN KEY (sender_id) REFERENCES user(userid)
		);`,

		`
		CREATE TABLE IF NOT EXISTS online_status (
			user_id INTEGER PRIMARY KEY,
			status TEXT NOT NULL DEFAULT 'offline',
			last_seen DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES user(userid)
		);`,

		`CREATE INDEX IF NOT EXISTS idx_message_conversation ON message(conversation_id);`,
		`CREATE INDEX IF NOT EXISTS idx_message_sender ON message(sender_id);`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_participants_user ON conversation_participants(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_conversation_participants_conv ON conversation_participants(conversation_id);`,
		`CREATE INDEX IF NOT EXISTS idx_online_status_user ON online_status(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_online_status_last_seen ON online_status(last_seen);`,
	}

	for i, query := range createTables {
		log.Printf("[DEBUG] Executing table creation query #%d", i+1)
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("[FATAL] Failed to create table (query #%d): %v", i+1, err)
		}
		log.Printf("[INFO] Table creation query #%d executed successfully", i+1)
	}

	log.Println("[INFO] Database tables initialized successfully")

	var count int
	log.Printf("[DEBUG] Checking if categories table is populated")
	err = db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		log.Fatalf("[FATAL] Failed to query category count: %v", err)
	}

	if count == 0 {
		log.Println("[INFO] Inserting initial categories...")

		insertCategories := []string{
			`INSERT INTO categories (name) VALUES ('Git');`,
			`INSERT INTO categories (name) VALUES ('Go');`,
			`INSERT INTO categories (name) VALUES ('JS');`,
			`INSERT INTO categories (name) VALUES ('SQL');`,
			`INSERT INTO categories (name) VALUES ('CSS');`,
			`INSERT INTO categories (name) VALUES ('HTML');`,
			`INSERT INTO categories (name) VALUES ('Unix');`,
			`INSERT INTO categories (name) VALUES ('Docker');`,
			`INSERT INTO categories (name) VALUES ('Rust');`,
			`INSERT INTO categories (name) VALUES ('C');`,
			`INSERT INTO categories (name) VALUES ('Shell');`,
			`INSERT INTO categories (name) VALUES ('PHP');`,
			`INSERT INTO categories (name) VALUES ('Python');`,
			`INSERT INTO categories (name) VALUES ('Ruby');`,
			`INSERT INTO categories (name) VALUES ('C++');`,
			`INSERT INTO categories (name) VALUES ('GraphQL');`,
			`INSERT INTO categories (name) VALUES ('Ruby on Rails');`,
			`INSERT INTO categories (name) VALUES ('Laravel');`,
			`INSERT INTO categories (name) VALUES ('Django');`,
			`INSERT INTO categories (name) VALUES ('Electron');`,
			`INSERT INTO categories (name) VALUES ('TCP/IP');`,
			`INSERT INTO categories (name) VALUES ('HTTP');`,
			`INSERT INTO categories (name) VALUES ('WebSocket');`,
			`INSERT INTO categories (name) VALUES ('AI');`,
			`INSERT INTO categories (name) VALUES ('Machine Learning');`,
			`INSERT INTO categories (name) VALUES ('Data Science');`,
			`INSERT INTO categories (name) VALUES ('DevOps');`,
			`INSERT INTO categories (name) VALUES ('Blockchain');`,
			`INSERT INTO categories (name) VALUES ('Cybersecurity');`,
			`INSERT INTO categories (name) VALUES ('Java');`,
			`INSERT INTO categories (name) VALUES ('Mobile Development');`,
			`INSERT INTO categories (name) VALUES ('Web Assembly');`,
			`INSERT INTO categories (name) VALUES ('Serverless');`,
			`INSERT INTO categories (name) VALUES ('Microservices');`,
			`INSERT INTO categories (name) VALUES ('Testing');`,
			`INSERT INTO categories (name) VALUES ('UI/UX');`,
			`INSERT INTO categories (name) VALUES ('Game Development');`,
			`INSERT INTO categories (name) VALUES ('Embedded Systems');`,
			`INSERT INTO categories (name) VALUES ('Cloud Computing');`,
			`INSERT INTO categories (name) VALUES ('Quantum Computing');`,
		}

		for i, stmt := range insertCategories {
			log.Printf("[DEBUG] Inserting category #%d", i+1)
			_, err := db.Exec(stmt)
			if err != nil {
				log.Printf("[ERROR] Failed to insert category #%d (%s): %v", i+1, strings.TrimPrefix(stmt, "INSERT INTO categories (name) VALUES ('"), err)
			} else {
				log.Printf("[INFO] Successfully inserted category #%d", i+1)
			}
		}
		log.Println("[INFO] Initial categories inserted successfully")
	} else {
		log.Printf("[INFO] Categories table already populated with %d entries, skipping insertion", count)
	}
}

func DropDataBase() {
	log.Printf("[DEBUG] Attempting to connect to SQLite database for dropping tables")
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal("[FATAL] Failed to connect to the database for dropping tables: ", err)
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for dropping tables")

	const DropCategoriesTable = `DROP TABLE IF EXISTS categories;`
	const DropCommentTable = `DROP TABLE IF EXISTS comment;`
	const DropPostTable = `DROP TABLE IF EXISTS post;`
	const DropPostHasCategoriesTable = `DROP TABLE IF EXISTS post_has_categories;`
	const DropSessionsTable = `DROP TABLE IF EXISTS session;`
	const DropUserTable = `DROP TABLE IF EXISTS user;`
	const DropConversationTable = `DROP TABLE IF EXISTS conversation;`
	const DropConversationParticipantsTable = `DROP TABLE IF EXISTS conversation_participants;`
	const DropMessageTable = `DROP TABLE IF EXISTS message;`
	const DropOnlineStatusTable = `DROP TABLE IF EXISTS online_status;`

	dropTableStatements := []string{
		DropCategoriesTable,
		DropCommentTable,
		DropPostTable,
		DropPostHasCategoriesTable,
		DropSessionsTable,
		DropUserTable,
		DropConversationTable,
		DropConversationParticipantsTable,
		DropMessageTable,
		DropOnlineStatusTable,
	}

	for i, stmt := range dropTableStatements {
		log.Printf("[DEBUG] Executing drop table statement #%d", i+1)
		_, err = db.Exec(stmt)
		if err != nil {
			log.Fatalf("[FATAL] Failed to drop table (statement #%d): %v", i+1, err)
		}
		log.Printf("[INFO] Drop table statement #%d executed successfully", i+1)
	}

	log.Println("[INFO] Database tables dropped successfully")
}

func LoadTestData() error {
	log.Printf("[DEBUG] Attempting to connect to SQLite database for loading test data")
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Fatal("[FATAL] Failed to connect to the database for loading test data: ", err)
		return err
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for loading test data")

	// Check if user table is already populated
	var count int
	log.Printf("[DEBUG] Checking if user table is populated before loading test data")
	err = db.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
	if err != nil {
		log.Printf("[ERROR] Failed to query user table count before loading test data: %v", err)
		return err
	}
	log.Printf("[INFO] User table contains %d records", count)
	if count > 0 {
		log.Printf("[INFO] Skipping test data loading as user table is already populated")
		return nil
	}

	log.Printf("[DEBUG] Reading seed data file for test data loading")
	fileContent, err := os.ReadFile("./database/seed_data.sql")
	if err != nil {
		log.Printf("[ERROR] Failed to read seed data file: %v", err)
		return err
	}
	log.Printf("[INFO] Successfully read seed data file for test data loading")

	log.Printf("[INFO] Loaded %d bytes from seed_data.sql file", len(fileContent))

	statements := strings.Split(string(fileContent), ";")
	log.Printf("[INFO] Found %d SQL statements to execute", len(statements))

	tx, err := db.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to start transaction for test data: %v", err)
		return err
	}
	log.Printf("[DEBUG] Started transaction for loading test data")

	executedCount := 0
	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		_, err := tx.Exec(statement)
		if err != nil {
			tx.Rollback()
			log.Printf("[ERROR] Failed to execute statement #%d: %v", i+1, err)
			log.Printf("[ERROR] Statement: %s", truncateSQL(statement))
			return err
		}
		executedCount++
		if (executedCount % 10) == 0 {
			log.Printf("[INFO] Executed %d statements for test data", executedCount)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] Failed to commit transaction for test data: %v", err)
		return err
	}
	log.Printf("[DEBUG] Committed transaction for loading test data")

	log.Printf("[INFO] Test data loaded successfully! Executed %d statements", executedCount)
	return nil
}

func truncateSQL(sql string) string {
	if len(sql) > 100 {
		return sql[:97] + "..."
	}
	return sql
}
