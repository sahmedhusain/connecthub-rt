package unit_testing

import (
	"fmt"
	"testing"

	"forum/repository"
	"forum/server/services"
)

func TestPostCreation(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create post repository and service
	postRepo := repository.NewPostRepository(testDB.DB)
	postService := services.NewPostService(testDB.DB)

	t.Run("ValidPostCreation", func(t *testing.T) {
		// Create a valid post
		postID, err := postService.CreatePost(
			userIDs[0], "Test Post Title", "This is a test post content",
			[]string{"General", "Technology"})
		AssertNoError(t, err, "Post creation should succeed")
		AssertTrue(t, postID > 0, "Post ID should be positive")

		// Verify post was created
		post, err := postRepo.GetPostByID(postID)
		AssertNoError(t, err, "Should be able to retrieve created post")
		AssertEqual(t, post.Title, "Test Post Title", "Post title should match")
		AssertEqual(t, post.Content, "This is a test post content", "Post content should match")
		AssertEqual(t, post.UserUserID, userIDs[0], "Post user ID should match")
	})

	t.Run("PostWithEmptyTitle", func(t *testing.T) {
		// Try to create post with empty title
		_, err := postService.CreatePost(
			userIDs[0], "", "Content without title",
			[]string{"General"})
		AssertError(t, err, "Post creation should fail with empty title")
	})

	t.Run("PostWithEmptyContent", func(t *testing.T) {
		// Try to create post with empty content
		_, err := postService.CreatePost(
			userIDs[0], "Title without content", "",
			[]string{"General"})
		AssertError(t, err, "Post creation should fail with empty content")
	})

	t.Run("PostWithWhitespaceOnly", func(t *testing.T) {
		// Try to create post with whitespace-only title
		_, err := postService.CreatePost(
			userIDs[0], "   ", "Valid content",
			[]string{"General"})
		AssertError(t, err, "Post creation should fail with whitespace-only title")

		// Try to create post with whitespace-only content
		_, err = postService.CreatePost(
			userIDs[0], "Valid title", "   ",
			[]string{"General"})
		AssertError(t, err, "Post creation should fail with whitespace-only content")
	})

	t.Run("PostWithMultipleCategories", func(t *testing.T) {
		// Create post with multiple categories
		postID, err := postService.CreatePost(
			userIDs[1], "Multi-Category Post", "This post belongs to multiple categories",
			[]string{"Technology", "Science", "General"})
		AssertNoError(t, err, "Post creation with multiple categories should succeed")

		// Verify post was created
		post, err := postRepo.GetPostByID(postID)
		AssertNoError(t, err, "Should be able to retrieve post")
		AssertTrue(t, len(post.Categories) >= 3, "Post should have multiple categories")
	})

	t.Run("PostWithInvalidUserID", func(t *testing.T) {
		// Try to create post with invalid user ID
		_, err := postService.CreatePost(
			99999, "Invalid User Post", "Content from invalid user",
			[]string{"General"})
		AssertError(t, err, "Post creation should fail with invalid user ID")
	})
}

func TestPostRetrieval(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	_, postIDs, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	postRepo := repository.NewPostRepository(testDB.DB)
	postService := services.NewPostService(testDB.DB)

	t.Run("GetAllPosts", func(t *testing.T) {
		// Get all posts
		posts, err := postService.GetAllPosts()
		AssertNoError(t, err, "Should be able to get all posts")
		AssertTrue(t, len(posts) >= 5, "Should have at least 5 test posts")

		// Verify post data structure
		for _, post := range posts {
			AssertTrue(t, post.PostID > 0, "Post ID should be positive")
			AssertNotEqual(t, post.Title, "", "Post title should not be empty")
			AssertNotEqual(t, post.Content, "", "Post content should not be empty")
			AssertTrue(t, post.UserUserID > 0, "User ID should be positive")
			AssertNotEqual(t, post.Username, "", "Username should not be empty")
		}
	})

	t.Run("GetPostByID", func(t *testing.T) {
		// Get specific post
		post, err := postRepo.GetPostByID(postIDs[0])
		AssertNoError(t, err, "Should be able to get post by ID")
		AssertEqual(t, post.PostID, postIDs[0], "Post ID should match")
		AssertNotEqual(t, post.Title, "", "Post title should not be empty")
		AssertNotEqual(t, post.Content, "", "Post content should not be empty")
	})

	t.Run("GetNonexistentPost", func(t *testing.T) {
		// Try to get non-existent post
		_, err := postRepo.GetPostByID(99999)
		AssertError(t, err, "Should fail to get non-existent post")
	})

	t.Run("PostsOrderedByDate", func(t *testing.T) {
		// Get all posts and verify they're ordered by date (newest first)
		posts, err := postService.GetAllPosts()
		AssertNoError(t, err, "Should be able to get all posts")

		if len(posts) > 1 {
			for i := 0; i < len(posts)-1; i++ {
				// Posts should be ordered by date descending (newest first)
				AssertTrue(t, posts[i].PostAt.After(posts[i+1].PostAt) || posts[i].PostAt.Equal(posts[i+1].PostAt),
					"Posts should be ordered by date (newest first)")
			}
		}
	})

	t.Run("PostWithUserInformation", func(t *testing.T) {
		// Get post and verify user information is included
		post, err := postRepo.GetPostByID(postIDs[0])
		AssertNoError(t, err, "Should be able to get post")
		AssertNotEqual(t, post.Username, "", "Post should include username")
		AssertNotEqual(t, post.FirstName, "", "Post should include first name")
		AssertNotEqual(t, post.LastName, "", "Post should include last name")
	})
}

func TestPostFiltering(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	userIDs, _, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	postService := services.NewPostService(testDB.DB)

	t.Run("FilterAllPosts", func(t *testing.T) {
		// Get all posts using filter
		posts, err := postService.GetFilteredPosts("all")
		AssertNoError(t, err, "Should be able to get all posts with filter")
		AssertTrue(t, len(posts) >= 5, "Should have at least 5 posts")
	})

	t.Run("FilterTopRatedPosts", func(t *testing.T) {
		// Get top-rated posts
		_, err = postService.GetFilteredPosts("top-rated")
		AssertNoError(t, err, "Should be able to get top-rated posts")
		// Note: This test assumes the filtering logic is implemented
		// The actual behavior depends on the implementation
	})

	t.Run("FilterOldestPosts", func(t *testing.T) {
		// Get oldest posts
		posts, err := postService.GetFilteredPosts("oldest")
		AssertNoError(t, err, "Should be able to get oldest posts")

		if len(posts) > 1 {
			// Verify posts are ordered by date ascending (oldest first)
			for i := 0; i < len(posts)-1; i++ {
				AssertTrue(t, posts[i].PostAt.Before(posts[i+1].PostAt) || posts[i].PostAt.Equal(posts[i+1].PostAt),
					"Oldest posts should be ordered by date (oldest first)")
			}
		}
	})

	t.Run("InvalidFilter", func(t *testing.T) {
		// Try invalid filter
		_, err := postService.GetFilteredPosts("invalid-filter")
		AssertError(t, err, "Should fail with invalid filter")
	})

	t.Run("GetPostsByUser", func(t *testing.T) {
		// Get posts by specific user
		posts, err := postService.GetPostsByUser(userIDs[0])
		AssertNoError(t, err, "Should be able to get posts by user")

		// Verify all posts belong to the specified user
		for _, post := range posts {
			AssertEqual(t, post.UserUserID, userIDs[0], "All posts should belong to specified user")
		}
	})

	t.Run("GetPostsByNonexistentUser", func(t *testing.T) {
		// Get posts by non-existent user
		posts, err := postService.GetPostsByUser(99999)
		AssertNoError(t, err, "Should not error for non-existent user")
		AssertEqual(t, len(posts), 0, "Should return empty list for non-existent user")
	})
}

func TestPostCategories(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	postService := services.NewPostService(testDB.DB)

	t.Run("CreatePostWithNewCategory", func(t *testing.T) {
		// Create post with a new category
		_, err = postService.CreatePost(
			userIDs[0], "New Category Post", "Post with new category",
			[]string{"NewCategory"})
		AssertNoError(t, err, "Should be able to create post with new category")

		// Verify category was created
		var categoryCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE name = ?", "NewCategory").Scan(&categoryCount)
		AssertNoError(t, err, "Should be able to query category")
		AssertEqual(t, categoryCount, 1, "New category should be created")
	})

	t.Run("CreatePostWithExistingCategory", func(t *testing.T) {
		// First, create a category
		_, err := testDB.DB.Exec("INSERT INTO categories (name) VALUES (?)", "ExistingCategory")
		AssertNoError(t, err, "Should be able to create category")

		// Create post with existing category
		_, err = postService.CreatePost(
			userIDs[0], "Existing Category Post", "Post with existing category",
			[]string{"ExistingCategory"})
		AssertNoError(t, err, "Should be able to create post with existing category")

		// Verify only one category exists
		var categoryCount int
		err = testDB.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE name = ?", "ExistingCategory").Scan(&categoryCount)
		AssertNoError(t, err, "Should be able to query category")
		AssertEqual(t, categoryCount, 1, "Should not duplicate existing category")
	})

	t.Run("GetPostsByCategory", func(t *testing.T) {
		// Create posts in specific category
		categoryName := "TestCategory"
		postID1, err := postService.CreatePost(
			userIDs[0], "Category Post 1", "First post in category",
			[]string{categoryName})
		AssertNoError(t, err, "Should be able to create first post")

		postID2, err := postService.CreatePost(
			userIDs[1], "Category Post 2", "Second post in category",
			[]string{categoryName})
		AssertNoError(t, err, "Should be able to create second post")

		// Get posts by category
		posts, err := postService.GetPostsByCategory(categoryName)
		AssertNoError(t, err, "Should be able to get posts by category")
		AssertTrue(t, len(posts) >= 2, "Should have at least 2 posts in category")

		// Verify posts belong to the category
		foundPost1, foundPost2 := false, false
		for _, post := range posts {
			if post.PostID == postID1 {
				foundPost1 = true
			}
			if post.PostID == postID2 {
				foundPost2 = true
			}
		}
		AssertTrue(t, foundPost1, "Should find first post in category")
		AssertTrue(t, foundPost2, "Should find second post in category")
	})

	t.Run("GetPostsByNonexistentCategory", func(t *testing.T) {
		// Get posts by non-existent category
		posts, err := postService.GetPostsByCategory("NonexistentCategory")
		AssertNoError(t, err, "Should not error for non-existent category")
		AssertEqual(t, len(posts), 0, "Should return empty list for non-existent category")
	})
}

func TestPostTestRepository(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	postRepo := repository.NewPostRepository(testDB.DB)

	t.Run("CreateAndRetrievePost", func(t *testing.T) {
		// Create post directly through repository
		postID, err := postRepo.CreatePost(
			userIDs[0], "Repository Test Post", "Content from repository test",
			[]string{"General"})
		AssertNoError(t, err, "Repository post creation should succeed")
		AssertTrue(t, postID > 0, "Post ID should be positive")

		// Retrieve post
		post, err := postRepo.GetPostByID(postID)
		AssertNoError(t, err, "Should be able to retrieve post")
		AssertEqual(t, post.PostID, postID, "Post ID should match")
		AssertEqual(t, post.Title, "Repository Test Post", "Title should match")
		AssertEqual(t, post.Content, "Content from repository test", "Content should match")
	})

	t.Run("GetAllPostsRepository", func(t *testing.T) {
		// Create multiple posts
		for i := 0; i < 3; i++ {
			_, err := postRepo.CreatePost(
				userIDs[i%len(userIDs)], fmt.Sprintf("Repo Post %d", i), fmt.Sprintf("Content %d", i),
				[]string{"General"})
			AssertNoError(t, err, "Post creation should succeed")
		}

		// Get all posts
		posts, err := postRepo.GetAllPosts()
		AssertNoError(t, err, "Should be able to get all posts")
		AssertTrue(t, len(posts) >= 3, "Should have at least 3 posts")
	})

	t.Run("GetFilteredPostsRepository", func(t *testing.T) {
		// Test different filters
		filters := []string{"all", "top-rated", "oldest"}

		for _, filter := range filters {
			_, err := postRepo.GetFilteredPosts(filter)
			if filter == "invalid-filter" {
				AssertError(t, err, "Invalid filter should fail")
			} else {
				AssertNoError(t, err, fmt.Sprintf("Filter %s should work", filter))
			}
		}
	})
}
