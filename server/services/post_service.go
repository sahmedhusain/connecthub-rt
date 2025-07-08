package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"forum/database"
)

// PostService handles post-related business logic
type PostService struct {
	db *sql.DB
}

// NewPostService creates a new PostService instance
func NewPostService(db *sql.DB) *PostService {
	return &PostService{db: db}
}

// GetAllPosts retrieves all posts from the database
func (s *PostService) GetAllPosts() ([]database.Post, error) {
	log.Printf("[DEBUG] PostService: Getting all posts")

	posts, err := database.GetAllPosts(s.db)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get all posts: %v", err)
		return nil, err
	}

	log.Printf("[INFO] PostService: Retrieved %d posts", len(posts))
	return posts, nil
}

// GetFilteredPosts retrieves posts with specific filters
func (s *PostService) GetFilteredPosts(filter string) ([]database.Post, error) {
	log.Printf("[DEBUG] PostService: Getting filtered posts with filter: %s", filter)

	var posts []database.Post
	var err error

	switch filter {
	case "all":
		posts, err = database.GetAllPosts(s.db)
	case "top-rated", "oldest":
		posts, err = database.GetFilteredPosts(s.db, filter)
	default:
		log.Printf("[ERROR] PostService: Invalid filter: %s", filter)
		return nil, fmt.Errorf("invalid filter")
	}

	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get filtered posts: %v", err)
		return nil, err
	}

	log.Printf("[INFO] PostService: Retrieved %d posts with filter: %s", len(posts), filter)
	return posts, nil
}

// GetPostsByCategory retrieves posts by category name
func (s *PostService) GetPostsByCategory(categoryName string) ([]database.Post, error) {
	log.Printf("[DEBUG] PostService: Getting posts by category: %s", categoryName)

	// Validate category exists
	categories, err := database.GetCategories(s.db)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get categories: %v", err)
		return nil, fmt.Errorf("failed to validate category")
	}

	categoryExists := false
	for _, category := range categories {
		if category.Name == categoryName {
			categoryExists = true
			break
		}
	}

	if !categoryExists {
		log.Printf("[WARN] PostService: Invalid category: %s", categoryName)
		return nil, fmt.Errorf("invalid category")
	}

	posts, err := database.GetPostsByCategory(s.db, categoryName)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get posts by category: %v", err)
		return nil, err
	}

	log.Printf("[INFO] PostService: Retrieved %d posts for category: %s", len(posts), categoryName)
	return posts, nil
}

// GetPostsByUser retrieves posts created by a specific user
func (s *PostService) GetPostsByUser(userID int) ([]database.Post, error) {
	log.Printf("[DEBUG] PostService: Getting posts by user ID: %d", userID)

	posts, err := database.GetPostsByUser(s.db, userID)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get posts by user: %v", err)
		return nil, err
	}

	log.Printf("[INFO] PostService: Retrieved %d posts for user ID: %d", len(posts), userID)
	return posts, nil
}

// GetLikedPostsByUser retrieves posts liked by a specific user
func (s *PostService) GetLikedPostsByUser(userID int) ([]database.Post, error) {
	log.Printf("[DEBUG] PostService: Getting liked posts by user ID: %d", userID)

	posts, err := database.GetLikedPostsByUser(s.db, userID)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get liked posts by user: %v", err)
		return nil, err
	}

	log.Printf("[INFO] PostService: Retrieved %d liked posts for user ID: %d", len(posts), userID)
	return posts, nil
}

// GetPostByID retrieves a single post by its ID
func (s *PostService) GetPostByID(postID int) (database.Post, error) {
	log.Printf("[DEBUG] PostService: Getting post by ID: %d", postID)

	post, err := database.GetPostByID(s.db, postID)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get post by ID: %v", err)
		return post, err
	}

	log.Printf("[INFO] PostService: Retrieved post ID: %d, title: %s", postID, post.Title)
	return post, nil
}

// GetPostWithComments retrieves a post with its comments
func (s *PostService) GetPostWithComments(postID int) (map[string]interface{}, error) {
	log.Printf("[DEBUG] PostService: Getting post with comments for ID: %d", postID)

	// Get the post
	post, err := database.GetPostByID(s.db, postID)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get post: %v", err)
		return nil, err
	}

	// Get comments for the post
	comments, err := database.GetCommentsForPost(s.db, postID)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get comments: %v", err)
		return nil, err
	}

	// Get categories for the post
	categories, err := database.GetCategoriesForPost(s.db, postID)
	if err != nil {
		log.Printf("[WARN] PostService: Failed to get categories: %v", err)
		categories = []database.Category{} // Set empty slice if error
	}

	response := map[string]interface{}{
		"post":       post,
		"comments":   comments,
		"categories": categories,
	}

	log.Printf("[INFO] PostService: Retrieved post %d with %d comments and %d categories", postID, len(comments), len(categories))
	return response, nil
}

// CreatePost creates a new post with validation
func (s *PostService) CreatePost(userID int, title, content string, categories []string) (int, error) {
	log.Printf("[DEBUG] PostService: Creating post for user ID: %d, title: %s", userID, title)

	// Validate input
	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" {
		return 0, fmt.Errorf("title and content are required")
	}

	// Validate categories if provided
	if len(categories) > 0 {
		validCategories, err := database.GetCategories(s.db)
		if err != nil {
			log.Printf("[ERROR] PostService: Failed to get valid categories: %v", err)
			return 0, fmt.Errorf("failed to validate categories")
		}

		validCategoryNames := make(map[string]bool)
		for _, cat := range validCategories {
			validCategoryNames[cat.Name] = true
		}

		for _, category := range categories {
			if !validCategoryNames[category] {
				log.Printf("[WARN] PostService: Invalid category: %s", category)
				return 0, fmt.Errorf("invalid category: %s", category)
			}
		}
	}

	// Create the post
	postID, err := database.CreatePost(s.db, userID, title, content, categories)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to create post: %v", err)
		return 0, err
	}

	log.Printf("[INFO] PostService: Post created successfully with ID: %d", postID)
	return postID, nil
}

// AddComment adds a comment to a post with validation
func (s *PostService) AddComment(postID, userID int, content string) error {
	log.Printf("[DEBUG] PostService: Adding comment to post ID: %d by user ID: %d", postID, userID)

	// Validate input with user-friendly messages
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("comment content is required. Please write your comment")
	}

	if len(content) > 1000 {
		return fmt.Errorf("comment is too long. Please keep it under 1,000 characters")
	}

	// Verify post exists
	_, err := database.GetPostByID(s.db, postID)
	if err != nil {
		log.Printf("[ERROR] PostService: Post not found: %v", err)
		return fmt.Errorf("the post you're trying to comment on was not found. It may have been deleted")
	}

	// Add the comment
	err = database.AddComment(s.db, postID, userID, content)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to add comment: %v", err)
		return err
	}

	log.Printf("[INFO] PostService: Comment added successfully to post ID: %d", postID)
	return nil
}

// GetCategories retrieves all available categories
func (s *PostService) GetCategories() ([]database.Category, error) {
	log.Printf("[DEBUG] PostService: Getting all categories")

	categories, err := database.GetCategories(s.db)
	if err != nil {
		log.Printf("[ERROR] PostService: Failed to get categories: %v", err)
		return nil, err
	}

	log.Printf("[INFO] PostService: Retrieved %d categories", len(categories))
	return categories, nil
}

// ValidateFilter checks if a filter is valid for the given context
func (s *PostService) ValidateFilter(filter string, context string) bool {
	switch context {
	case "posts":
		validFilters := []string{"all", "top-rated", "oldest"}
		for _, validFilter := range validFilters {
			if filter == validFilter {
				return true
			}
		}
		return false
	case "tags":
		// For tags context, we need to check against actual categories
		categories, err := database.GetCategories(s.db)
		if err != nil {
			log.Printf("[ERROR] PostService: Failed to get categories for validation: %v", err)
			return false
		}

		if filter == "all" {
			return true
		}

		for _, category := range categories {
			if filter == category.Name {
				return true
			}
		}
		return false
	default:
		return false
	}
}
