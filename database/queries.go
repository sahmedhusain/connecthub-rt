package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID               int            `json:"id"`
	FirstName        string         `json:"first_name"`
	LastName         string         `json:"last_name"`
	Username         string         `json:"username"`
	Email            string         `json:"email"`
	Password         string         `json:"password"`
	SessionSessionID int            `json:"current_session"`
	Avatar           sql.NullString `json:"avatar"`
	Gender           string         `json:"gender"`
	DateOfBirth      string         `json:"date_of_birth"`
}

type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	FirstName string
	LastName  string
	Username  string
	Content   string
	CreatedAt time.Time
	Avatar    sql.NullString
}

type Post struct {
	PostID      int
	Image       sql.NullString
	Title       string
	Content     string
	PostAt      time.Time
	UserUserID  int
	Username    string
	FirstName   string
	LastName    string
	Avatar      sql.NullString
	Comments    int
	Categories  []Category
	ImageBase64 string
}

type UserSession struct {
	ID     int
	UserID int
	Start  time.Time
	End    time.Time
}

func Select(colToReturn, table, where, input string) (string, error) {
	log.Printf("[DEBUG] Attempting to connect to SQLite database for Select operation on table %s", table)
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed in Select for table %s: %v", table, err)
		return "", err
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for Select operation on table %s", table)

	statement := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", colToReturn, table, where)
	log.Printf("[DEBUG] Executing query: %s with input: %s", statement, input)

	var returnedValue string
	err = db.QueryRow(statement, input).Scan(&returnedValue)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[INFO] No rows found for %s in %s where %s = %s", colToReturn, table, where, input)
		} else {
			log.Printf("[ERROR] Query failed for %s in %s where %s = %s: %v", colToReturn, table, where, input, err)
		}
		return "", err
	}

	log.Printf("[INFO] Successfully retrieved %s from %s where %s = %s", colToReturn, table, where, input)
	return returnedValue, nil
}

func ExecuteQuery(query string, args ...interface{}) (*sql.Rows, error) {
	log.Printf("[DEBUG] Attempting to connect to SQLite database for ExecuteQuery operation")
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed in ExecuteQuery: %v", err)
		return nil, err
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for ExecuteQuery operation")

	truncatedQuery := truncateSQL(query)
	log.Printf("[DEBUG] Executing query: %s with %d arguments", truncatedQuery, len(args))

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("[ERROR] Query execution failed: %v", err)
		return nil, err
	}

	log.Printf("[INFO] Query executed successfully: %s", truncatedQuery)
	return rows, nil
}

func ExecuteNonQuery(query string, args ...interface{}) (sql.Result, error) {
	log.Printf("[DEBUG] Attempting to connect to SQLite database for ExecuteNonQuery operation")
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed in ExecuteNonQuery: %v", err)
		return nil, err
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for ExecuteNonQuery operation")

	truncatedQuery := truncateSQL(query)
	log.Printf("[DEBUG] Executing non-query: %s with %d arguments", truncatedQuery, len(args))

	result, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("[ERROR] Non-query execution failed: %v", err)
		return nil, err
	}

	affected, _ := result.RowsAffected()
	log.Printf("[INFO] Non-query executed successfully: %s (Affected rows: %d)", truncatedQuery, affected)
	return result, nil
}

func CheckExists(table, condition string, args ...interface{}) (bool, error) {
	log.Printf("[DEBUG] Attempting to connect to SQLite database for CheckExists operation on table %s", table)
	db, err := sql.Open("sqlite3", "./database/main.db")
	if err != nil {
		log.Printf("[ERROR] Database connection failed in CheckExists for table %s: %v", table, err)
		return false, err
	}
	defer db.Close()
	log.Printf("[INFO] Successfully connected to SQLite database for CheckExists operation on table %s", table)

	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", table, condition)
	log.Printf("[DEBUG] Checking existence with query: %s with %d arguments", truncateSQL(query), len(args))

	var exists bool
	err = db.QueryRow(query, args...).Scan(&exists)
	if err != nil {
		log.Printf("[ERROR] Error checking existence in %s where %s: %v", table, condition, err)
		return false, err
	}

	log.Printf("[INFO] Existence check in %s where %s: %v", table, condition, exists)
	return exists, nil
}

func GetCategories(db *sql.DB) ([]Category, error) {
	log.Printf("[DEBUG] Retrieving all categories")

	rows, err := db.Query("SELECT idcategories, name FROM categories")
	if err != nil {
		log.Printf("[ERROR] Failed to query categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Printf("[ERROR] Failed to scan category row: %v", err)
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating category rows: %v", err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d categories", len(categories))
	return categories, nil
}

func GetCommentsForPost(db *sql.DB, postID int) ([]Comment, error) {
	log.Printf("[DEBUG] Retrieving comments for post ID %d", postID)

	query := `
        SELECT comment.commentid, comment.post_postid, comment.user_userid, user.F_name, user.L_name, user.Username, comment.content, comment.comment_at, user.Avatar
        FROM comment
        JOIN user ON comment.user_userid = user.userid
        WHERE comment.post_postid = ?`
	rows, err := db.Query(query, postID)
	if err != nil {
		log.Printf("[ERROR] Failed to query comments for post ID %d: %v", postID, err)
		return nil, fmt.Errorf("GetCommentsForPost query failed: %v", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var commentAt time.Time
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.FirstName, &comment.LastName, &comment.Username, &comment.Content, &commentAt, &comment.Avatar); err != nil {
			log.Printf("[ERROR] Failed to scan comment row for post ID %d: %v", postID, err)
			return nil, fmt.Errorf("GetCommentsForPost scan failed: %v", err)
		}
		comment.CreatedAt = commentAt
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating comment rows for post ID %d: %v", postID, err)
		return nil, fmt.Errorf("GetCommentsForPost row iteration error: %v", err)
	}

	log.Printf("[INFO] Retrieved %d comments for post ID %d", len(comments), postID)
	return comments, nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving all posts")

	query := `
        SELECT post.postid, post.title, post.content, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        ORDER BY post.post_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("[ERROR] Failed to query all posts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Title, &post.Content, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan post row: %v", err)
			return nil, err
		}
		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' with multiple formats: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating post rows: %v", err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d posts", len(posts))
	return posts, nil
}

func GetCategoriesForPost(db *sql.DB, postID int) ([]Category, error) {
	log.Printf("[DEBUG] Retrieving categories for post ID %d", postID)

	query := `
        SELECT c.idcategories, c.name
        FROM categories c
        JOIN post_has_categories phc ON c.idcategories = phc.categories_idcategories
        WHERE phc.post_postid = ?`
	rows, err := db.Query(query, postID)
	if err != nil {
		log.Printf("[ERROR] Failed to query categories for post ID %d: %v", postID, err)
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	rowCount := 0
	for rows.Next() {
		rowCount++
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Printf("[ERROR] Failed to scan category row for post ID %d: %v", postID, err)
			return categories, err
		}
		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating category rows for post ID %d: %v", postID, err)
		return categories, err
	}

	log.Printf("[INFO] Retrieved %d categories for post ID %d", len(categories), postID)
	return categories, nil
}

func GetComments(db *sql.DB) ([]Comment, error) {
	log.Printf("[DEBUG] Retrieving all comments")

	rows, err := db.Query("SELECT commentid, content, comment_at, post_postid, user_userid FROM comment")
	if err != nil {
		log.Printf("[ERROR] Failed to query all comments: %v", err)
		return nil, err
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var commentAt time.Time
		if err := rows.Scan(&comment.ID, &comment.Content, &commentAt, &comment.PostID, &comment.UserID); err != nil {
			log.Printf("[ERROR] Failed to scan comment row: %v", err)
			return nil, err
		}

		comment.CreatedAt = commentAt
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating comment rows: %v", err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d comments", len(comments))
	return comments, nil
}

func GetUserCommentedPosts(db *sql.DB, userid int, filter string) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving posts commented by user ID %d with filter '%s'", userid, filter)

	order := "DESC"
	if filter == "oldest" {
		order = "ASC"
	}

	query := fmt.Sprintf(`
        SELECT DISTINCT post.postid, post.title, post.content, post.post_at, post.user_userid, u.Username, u.F_name, u.L_name, u.Avatar,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN comment c ON post.postid = c.post_postid
        JOIN user u ON post.user_userid = u.userid -- Join post user, not comment user for post details
        WHERE c.user_userid = ? -- Filter by the user who commented
        ORDER BY post.post_at %s
    `, order)

	rows, err := db.Query(query, userid)
	if err != nil {
		log.Printf("[ERROR] Failed to query posts commented by user ID %d: %v", userid, err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Title, &post.Content, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan post row for user ID %d's commented posts: %v", userid, err)
			return nil, err
		}

		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' in GetUserCommentedPosts: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating post rows for user ID %d's commented posts: %v", userid, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d posts commented by user ID %d", len(posts), userid)
	return posts, nil
}

func GetUserByID(db *sql.DB, userID int) (User, error) {
	log.Printf("[DEBUG] Retrieving user with ID %d", userID)

	var user User
	err := db.QueryRow("SELECT userid, F_name, L_name, Username, Email, Avatar, gender, date_of_birth FROM user WHERE userid = ?", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &user.Avatar, &user.Gender, &user.DateOfBirth)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[INFO] No user found with ID %d", userID)
		} else {
			log.Printf("[ERROR] Failed to query user with ID %d: %v", userID, err)
		}
		return user, err
	}

	log.Printf("[INFO] Retrieved user with ID %d: username '%s'", userID, user.Username)
	return user, nil
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	log.Printf("[DEBUG] Retrieving all users")

	rows, err := db.Query("SELECT userid, F_name, L_name, Username, Email, Avatar FROM user")
	if err != nil {
		log.Printf("[ERROR] Failed to query all users: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Username, &user.Email, &avatar); err != nil {
			log.Printf("[ERROR] Failed to scan user row: %v", err)
			return nil, err
		}
		user.Avatar = avatar
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating user rows: %v", err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d users", len(users))
	return users, nil
}

func GetFilteredPosts(db *sql.DB, filter string) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving posts with filter '%s'", filter)

	var rows *sql.Rows
	var err error
	var query string

	switch filter {
	case "oldest":
		query = `
            SELECT post.postid, post.content, post.title, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
            ORDER BY post.post_at ASC
        `
		rows, err = db.Query(query)
	case "all":
		fallthrough
	default:
		query = `
            SELECT post.postid, post.content, post.title, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
                   (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
            FROM post
            JOIN user ON post.user_userid = user.userid
            ORDER BY post.post_at DESC
        `
		rows, err = db.Query(query)
	}

	if err != nil {
		log.Printf("[ERROR] Failed to query filtered posts with filter '%s': %v", filter, err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Content, &post.Title, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan post row with filter '%s': %v", filter, err)
			return nil, err
		}

		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' in GetFilteredPosts: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories

		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating post rows with filter '%s': %v", filter, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d posts with filter '%s'", len(posts), filter)
	return posts, nil
}

func GetPostsByMultiCategory(db *sql.DB, categoryName string) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving posts by multi-category '%s'", categoryName)

	rows, err := db.Query(`
        SELECT post.postid, post.content, post.title, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        JOIN post_has_categories phc ON post.postid = phc.post_postid
        JOIN categories c ON phc.categories_idcategories = c.idcategories
           WHERE c.name = ?
        ORDER BY post.post_at DESC
    `, categoryName)
	if err != nil {
		log.Printf("[ERROR] Failed to query posts by category '%s': %v", categoryName, err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Content, &post.Title, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan post row for category '%s': %v", categoryName, err)
			return nil, err
		}

		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' in GetPostsByMultiCategory: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating post rows for category '%s': %v", categoryName, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d posts for category '%s'", len(posts), categoryName)
	return posts, nil
}

func GetPostsByCategory(db *sql.DB, categoryName string) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving posts by category '%s'", categoryName)

	rows, err := db.Query(`
        SELECT post.postid, post.content, post.title, post.post_at, post.user_userid, user.Username, user.F_name, user.L_name, user.Avatar,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
        FROM post
        JOIN user ON post.user_userid = user.userid
        JOIN post_has_categories phc ON post.postid = phc.post_postid
        JOIN categories c ON phc.categories_idcategories = c.idcategories
        WHERE c.name = ?
        ORDER BY post.post_at DESC
    `, categoryName)
	if err != nil {
		log.Printf("[ERROR] Failed to query posts by category '%s': %v", categoryName, err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Content, &post.Title, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan post row for category '%s': %v", categoryName, err)
			return nil, err
		}

		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' in GetPostsByCategory: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating post rows for category '%s': %v", categoryName, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d posts for category '%s'", len(posts), categoryName)
	return posts, nil
}

func InsertPost(db *sql.DB, content string, title string, userID string) (int, error) {
	log.Printf("[DEBUG] Inserting new post for user ID %s with title '%s'", userID, title)

	stmt, err := db.Prepare("INSERT INTO post (content, title, post_at, user_userid) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("[ERROR] Failed to prepare insert post statement: %v", err)
		return 0, err
	}
	defer stmt.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")

	res, err := stmt.Exec(content, title, currentTime, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to execute insert post statement: %v", err)
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		log.Printf("[ERROR] Failed to get last insert ID for post: %v", err)
		return 0, err
	}

	log.Printf("[INFO] Inserted new post with ID %d for user ID %s", lastID, userID)
	return int(lastID), nil
}

func InsertPostCategory(db *sql.DB, postID int, categoryID int) error {
	log.Printf("[DEBUG] Linking post ID %d with category ID %d", postID, categoryID)

	stmt, err := db.Prepare("INSERT INTO post_has_categories (post_postid, categories_idcategories) VALUES (?, ?)")
	if err != nil {
		log.Printf("[ERROR] Failed to prepare insert post category statement: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(postID, categoryID)
	if err != nil {
		log.Printf("[ERROR] Failed to link post ID %d with category ID %d: %v", postID, categoryID, err)
		return err
	}

	log.Printf("[INFO] Successfully linked post ID %d with category ID %d", postID, categoryID)
	return nil
}

func GetUserPosts(db *sql.DB, userID int, filter string) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving posts for user ID %d with filter '%s'", userID, filter)

	var x string
	if filter == "oldest" {
		x = "post.post_at ASC"
	} else {
		x = "post.post_at DESC"
	}

	query := `SELECT
		post.postid, post.content, post.title, post.post_at, post.user_userid,
		user.avatar, user.F_name, user.L_name, user.Username,
               (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
	FROM post
	JOIN user ON post.user_userid = user.userid
	WHERE post.user_userid = ? ORDER BY ` + x

	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to query posts for user ID %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Content, &post.Title, &postAt, &post.UserUserID, &post.Avatar, &post.FirstName, &post.LastName, &post.Username, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan post row for user ID %d: %v", userID, err)
			return nil, err
		}

		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' in GetUserPosts: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating post rows for user ID %d: %v", userID, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d posts for user ID %d", len(posts), userID)
	return posts, nil
}

// AuthenticateUser authenticates a user by identifier (username or email) and password
func AuthenticateUser(db *sql.DB, identifier, password string) (*User, error) {
	log.Printf("[DEBUG] Authenticating user with identifier: %s", identifier)

	var user User
	var hashedPassword string

	query := `
		SELECT userid, F_name, L_name, Username, Email, password, Avatar, gender, date_of_birth
		FROM user
		WHERE Username = ? OR Email = ?
	`

	err := db.QueryRow(query, identifier, identifier).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Username,
		&user.Email, &hashedPassword, &user.Avatar, &user.Gender, &user.DateOfBirth,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] No user found with identifier: %s", identifier)
			return nil, fmt.Errorf("invalid credentials")
		}
		log.Printf("[ERROR] Database error during authentication: %v", err)
		return nil, err
	}

	// Verify password using bcrypt
	if !verifyPassword(password, hashedPassword) {
		log.Printf("[WARN] Password verification failed for user: %s", user.Username)
		return nil, fmt.Errorf("invalid credentials")
	}

	log.Printf("[INFO] User authenticated successfully: %s (ID: %d)", user.Username, user.ID)
	return &user, nil
}

// UpdateUserSession updates the user's current session token
func UpdateUserSession(db *sql.DB, userID int, sessionToken string) error {
	log.Printf("[DEBUG] Updating session for user ID %d", userID)

	query := `UPDATE user SET current_session = ? WHERE userid = ?`
	_, err := db.Exec(query, sessionToken, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to update session for user ID %d: %v", userID, err)
		return err
	}

	log.Printf("[INFO] Session updated successfully for user ID %d", userID)
	return nil
}

// UserExists checks if a user with the given username or email already exists
func UserExists(db *sql.DB, username, email string) (bool, error) {
	log.Printf("[DEBUG] Checking if user exists with username: %s or email: %s", username, email)

	var count int
	query := `SELECT COUNT(*) FROM user WHERE Username = ? OR Email = ?`
	err := db.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		log.Printf("[ERROR] Failed to check user existence: %v", err)
		return false, err
	}

	exists := count > 0
	log.Printf("[INFO] User existence check: %v (username: %s, email: %s)", exists, username, email)
	return exists, nil
}

// CreateUser creates a new user in the database
func CreateUser(db *sql.DB, firstName, lastName, username, email, gender, dateOfBirth, password string) (int, error) {
	log.Printf("[DEBUG] Creating new user: %s (%s)", username, email)

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password for user %s: %v", username, err)
		return 0, err
	}

	// Determine avatar based on gender
	var avatarPath string
	switch gender {
	case "male":
		avatarPath = "/static/assets/male-avatar-boy-face-man-user-7.svg"
	case "female":
		avatarPath = "/static/assets/female-avatar-girl-face-woman-user-9.svg"
	default:
		avatarPath = "/static/assets/default-avatar.png"
	}

	query := `
		INSERT INTO user (F_name, L_name, Username, Email, gender, date_of_birth, password, Avatar)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := db.Exec(query, firstName, lastName, username, email, gender, dateOfBirth, hashedPassword, avatarPath)
	if err != nil {
		log.Printf("[ERROR] Failed to create user %s: %v", username, err)
		return 0, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		log.Printf("[ERROR] Failed to get last insert ID for user %s: %v", username, err)
		return 0, err
	}

	log.Printf("[INFO] User created successfully: %s (ID: %d)", username, int(userID))
	return int(userID), nil
}

// GetPostByID retrieves a single post by its ID
func GetPostByID(db *sql.DB, postID int) (Post, error) {
	log.Printf("[DEBUG] Retrieving post with ID %d", postID)

	var post Post
	query := `
		SELECT post.postid, post.title, post.content, post.post_at, post.user_userid,
		       user.Username, user.F_name, user.L_name, user.Avatar,
		       (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
		FROM post
		JOIN user ON post.user_userid = user.userid
		WHERE post.postid = ?
	`

	var postAt string
	err := db.QueryRow(query, postID).Scan(
		&post.PostID, &post.Title, &post.Content, &postAt, &post.UserUserID,
		&post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[INFO] No post found with ID %d", postID)
		} else {
			log.Printf("[ERROR] Failed to query post with ID %d: %v", postID, err)
		}
		return post, err
	}

	// Parse the timestamp
	post.PostAt, err = time.Parse(time.RFC3339, postAt)
	if err != nil {
		layout := "2006-01-02 15:04:05"
		post.PostAt, err = time.Parse(layout, postAt)
		if err != nil {
			log.Printf("[WARN] Failed to parse post_at '%s' for post ID %d: %v", postAt, postID, err)
			post.PostAt = time.Time{}
		}
	}

	// Get categories for the post
	categories, err := GetCategoriesForPost(db, post.PostID)
	if err != nil {
		log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
	}
	post.Categories = categories

	log.Printf("[INFO] Retrieved post with ID %d: title '%s'", postID, post.Title)
	return post, nil
}

// CreatePost creates a new post with categories
func CreatePost(db *sql.DB, userID int, title, content string, categories []string) (int, error) {
	log.Printf("[DEBUG] Creating new post for user ID %d with title '%s'", userID, title)

	// Insert the post
	postID, err := InsertPost(db, content, title, fmt.Sprintf("%d", userID))
	if err != nil {
		log.Printf("[ERROR] Failed to insert post: %v", err)
		return 0, err
	}

	// Link categories to the post
	for _, categoryIDStr := range categories {
		// Convert category ID string to integer
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			log.Printf("[WARN] Invalid category ID '%s', skipping: %v", categoryIDStr, err)
			continue
		}

		// Verify category exists
		var exists int
		err = db.QueryRow("SELECT COUNT(*) FROM categories WHERE idcategories = ?", categoryID).Scan(&exists)
		if err != nil || exists == 0 {
			log.Printf("[WARN] Category ID %d not found, skipping: %v", categoryID, err)
			continue
		}

		// Link post to category
		err = InsertPostCategory(db, postID, categoryID)
		if err != nil {
			log.Printf("[ERROR] Failed to link post %d to category %d: %v", postID, categoryID, err)
		}
	}

	log.Printf("[INFO] Created post with ID %d for user %d", postID, userID)
	return postID, nil
}

// AddComment adds a comment to a post
func AddComment(db *sql.DB, postID, userID int, content string) error {
	log.Printf("[DEBUG] Adding comment to post ID %d by user ID %d", postID, userID)

	query := `
		INSERT INTO comment (post_postid, user_userid, content, comment_at)
		VALUES (?, ?, ?, ?)
	`

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err := db.Exec(query, postID, userID, content, currentTime)
	if err != nil {
		log.Printf("[ERROR] Failed to add comment to post ID %d: %v", postID, err)
		return err
	}

	log.Printf("[INFO] Comment added successfully to post ID %d by user ID %d", postID, userID)
	return nil
}

// GetPostsByUser retrieves posts created by a specific user
func GetPostsByUser(db *sql.DB, userID int) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving posts by user ID %d", userID)
	return GetUserPosts(db, userID, "newest")
}

// GetLikedPostsByUser retrieves posts liked by a specific user
func GetLikedPostsByUser(db *sql.DB, userID int) ([]Post, error) {
	log.Printf("[DEBUG] Retrieving liked posts by user ID %d", userID)

	// This is a placeholder implementation since we don't have a likes table yet
	// In a real implementation, you'd have a likes/reactions table
	query := `
		SELECT DISTINCT post.postid, post.title, post.content, post.post_at, post.user_userid,
		       user.Username, user.F_name, user.L_name, user.Avatar,
		       (SELECT COUNT(*) FROM comment WHERE comment.post_postid = post.postid) AS Comments
		FROM post
		JOIN user ON post.user_userid = user.userid
		WHERE post.postid IN (
			SELECT post_postid FROM comment WHERE user_userid = ?
		)
		ORDER BY post.post_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		log.Printf("[ERROR] Failed to query liked posts for user ID %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var postAt string
		if err := rows.Scan(&post.PostID, &post.Title, &post.Content, &postAt, &post.UserUserID, &post.Username, &post.FirstName, &post.LastName, &post.Avatar, &post.Comments); err != nil {
			log.Printf("[ERROR] Failed to scan liked post row for user ID %d: %v", userID, err)
			return nil, err
		}

		post.PostAt, err = time.Parse(time.RFC3339, postAt)
		if err != nil {
			layout := "2006-01-02 15:04:05"
			post.PostAt, err = time.Parse(layout, postAt)
			if err != nil {
				log.Printf("[WARN] Failed to parse post_at '%s' in GetLikedPostsByUser: %v", postAt, err)
				post.PostAt = time.Time{}
			}
		}

		categories, err := GetCategoriesForPost(db, post.PostID)
		if err != nil {
			log.Printf("[WARN] Failed to fetch categories for post ID %d: %v", post.PostID, err)
		}
		post.Categories = categories
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Printf("[ERROR] Error iterating liked post rows for user ID %d: %v", userID, err)
		return nil, err
	}

	log.Printf("[INFO] Retrieved %d liked posts for user ID %d", len(posts), userID)
	return posts, nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// verifyPassword checks if a provided password matches the stored hashed password
func verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
