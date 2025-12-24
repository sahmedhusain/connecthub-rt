package repository

import (
	"database/sql"
	"log"

	"connecthub/database"
)

// PostRepositoryImpl implements the PostRepository interface
type PostRepositoryImpl struct {
	db *sql.DB
}

// NewPostRepository creates a new PostRepository instance
func NewPostRepository(db *sql.DB) PostRepository {
	return &PostRepositoryImpl{db: db}
}

// GetAllPosts retrieves all posts from the database
func (r *PostRepositoryImpl) GetAllPosts() ([]database.Post, error) {
	log.Printf("[DEBUG] PostRepository: Getting all posts")
	return database.GetAllPosts(r.db)
}

// GetPostByID retrieves a single post by its ID
func (r *PostRepositoryImpl) GetPostByID(postID int) (database.Post, error) {
	log.Printf("[DEBUG] PostRepository: Getting post by ID: %d", postID)
	return database.GetPostByID(r.db, postID)
}

// GetFilteredPosts retrieves posts with specific filters
func (r *PostRepositoryImpl) GetFilteredPosts(filter string) ([]database.Post, error) {
	log.Printf("[DEBUG] PostRepository: Getting filtered posts with filter: %s", filter)
	return database.GetFilteredPosts(r.db, filter)
}

// GetPostsByCategory retrieves posts by category name
func (r *PostRepositoryImpl) GetPostsByCategory(categoryName string) ([]database.Post, error) {
	log.Printf("[DEBUG] PostRepository: Getting posts by category: %s", categoryName)
	return database.GetPostsByCategory(r.db, categoryName)
}

// GetPostsByUser retrieves posts created by a specific user
func (r *PostRepositoryImpl) GetPostsByUser(userID int) ([]database.Post, error) {
	log.Printf("[DEBUG] PostRepository: Getting posts by user ID: %d", userID)
	return database.GetPostsByUser(r.db, userID)
}

// GetLikedPostsByUser retrieves posts liked by a specific user
func (r *PostRepositoryImpl) GetLikedPostsByUser(userID int) ([]database.Post, error) {
	log.Printf("[DEBUG] PostRepository: Getting liked posts by user ID: %d", userID)
	return database.GetLikedPostsByUser(r.db, userID)
}

// CreatePost creates a new post with categories
func (r *PostRepositoryImpl) CreatePost(userID int, title, content string, categories []string) (int, error) {
	log.Printf("[DEBUG] PostRepository: Creating post for user ID: %d, title: %s", userID, title)
	return database.CreatePost(r.db, userID, title, content, categories)
}

// GetCommentsForPost retrieves comments for a specific post
func (r *PostRepositoryImpl) GetCommentsForPost(postID int) ([]database.Comment, error) {
	log.Printf("[DEBUG] PostRepository: Getting comments for post ID: %d", postID)
	return database.GetCommentsForPost(r.db, postID)
}

// AddComment adds a comment to a post
func (r *PostRepositoryImpl) AddComment(postID, userID int, content string) error {
	log.Printf("[DEBUG] PostRepository: Adding comment to post ID: %d by user ID: %d", postID, userID)
	return database.AddComment(r.db, postID, userID, content)
}

// GetCategories retrieves all available categories
func (r *PostRepositoryImpl) GetCategories() ([]database.Category, error) {
	log.Printf("[DEBUG] PostRepository: Getting all categories")
	return database.GetCategories(r.db)
}

// GetCategoriesForPost retrieves categories for a specific post
func (r *PostRepositoryImpl) GetCategoriesForPost(postID int) ([]database.Category, error) {
	log.Printf("[DEBUG] PostRepository: Getting categories for post ID: %d", postID)
	return database.GetCategoriesForPost(r.db, postID)
}
