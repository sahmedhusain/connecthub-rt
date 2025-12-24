package unit_testing

import (
	"testing"

	"connecthub/server/services"
)

func TestPostService(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Create post service
	postService := services.NewPostService(testDB.DB)

	t.Run("GetAllPosts", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		postIDs, err := SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		posts, err := postService.GetAllPosts()
		AssertNoError(t, err, "Should retrieve all posts")
		AssertGreaterThanOrEqual(t, len(posts), len(postIDs), "Should return at least the test posts")

		// Verify test posts are included
		postIDMap := make(map[int]bool)
		for _, post := range posts {
			postIDMap[post.PostID] = true
		}

		for _, postID := range postIDs {
			AssertEqual(t, postIDMap[postID], true, "Test post should be included in results")
		}

		// Verify post structure
		for _, post := range posts {
			AssertNotEqual(t, post.PostID, 0, "Post ID should be set")
			AssertNotEqual(t, post.Title, "", "Post title should not be empty")
			AssertNotEqual(t, post.Content, "", "Post content should not be empty")
			AssertNotEqual(t, post.UserUserID, 0, "Post user ID should be set")
			AssertNotEqual(t, post.Username, "", "Post username should not be empty")
		}
	})

	t.Run("GetFilteredPosts", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		_, err = SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("AllFilter", func(t *testing.T) {
			posts, err := postService.GetFilteredPosts("all")
			AssertNoError(t, err, "Should retrieve all posts")
			AssertGreaterThan(t, len(posts), 0, "Should return posts")
		})

		t.Run("TopRatedFilter", func(t *testing.T) {
			posts, err := postService.GetFilteredPosts("top-rated")
			AssertNoError(t, err, "Should retrieve top-rated posts")
			AssertGreaterThanOrEqual(t, len(posts), 0, "Should return posts or empty array")
		})

		t.Run("OldestFilter", func(t *testing.T) {
			posts, err := postService.GetFilteredPosts("oldest")
			AssertNoError(t, err, "Should retrieve oldest posts")
			AssertGreaterThanOrEqual(t, len(posts), 0, "Should return posts or empty array")
		})

		t.Run("InvalidFilter", func(t *testing.T) {
			posts, err := postService.GetFilteredPosts("invalid_filter")
			AssertNotEqual(t, err, nil, "Should fail for invalid filter")
			AssertEqual(t, len(posts), 0, "Should return empty array")
		})
	})

	t.Run("GetPostsByCategory", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		_, err = SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("ValidCategory", func(t *testing.T) {
			posts, err := postService.GetPostsByCategory("Technology")
			AssertNoError(t, err, "Should retrieve posts by category")
			AssertGreaterThanOrEqual(t, len(posts), 0, "Should return posts or empty array")

			// Verify all posts belong to the category
			for _, post := range posts {
				found := false
				for _, category := range post.Categories {
					if category.Name == "Technology" {
						found = true
						break
					}
				}
				AssertEqual(t, found, true, "Post should belong to Technology category")
			}
		})

		t.Run("NonExistentCategory", func(t *testing.T) {
			posts, err := postService.GetPostsByCategory("NonExistentCategory")
			AssertNoError(t, err, "Should handle non-existent category")
			AssertEqual(t, len(posts), 0, "Should return empty array")
		})

		t.Run("EmptyCategory", func(t *testing.T) {
			posts, err := postService.GetPostsByCategory("")
			AssertNoError(t, err, "Should handle empty category")
			AssertEqual(t, len(posts), 0, "Should return empty array")
		})
	})

	t.Run("GetPostByID", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		postIDs, err := SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("ValidPostID", func(t *testing.T) {
			post, err := postService.GetPostByID(postIDs[0])
			AssertNoError(t, err, "Should retrieve post by ID")
			AssertNotEqual(t, post, nil, "Post should be returned")
			AssertEqual(t, post.PostID, postIDs[0], "Post ID should match")
			AssertNotEqual(t, post.Title, "", "Post title should not be empty")
			AssertNotEqual(t, post.Content, "", "Post content should not be empty")
		})

		t.Run("InvalidPostID", func(t *testing.T) {
			post, err := postService.GetPostByID(99999)
			AssertNotEqual(t, err, nil, "Should fail for invalid post ID")
			AssertEqual(t, post, nil, "Post should not be returned")
		})

		t.Run("ZeroPostID", func(t *testing.T) {
			post, err := postService.GetPostByID(0)
			AssertNotEqual(t, err, nil, "Should fail for zero post ID")
			AssertEqual(t, post, nil, "Post should not be returned")
		})
	})

	t.Run("CreatePost", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		t.Run("ValidPost", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"Test Post Title",
				"This is a test post content",
				[]string{"Technology", "Programming"},
			)
			AssertNoError(t, err, "Post creation should succeed")
			AssertNotEqual(t, postID, 0, "Post ID should be set")

			// Verify post was created
			post, err := postService.GetPostByID(postID)
			AssertNoError(t, err, "Should retrieve created post")
			AssertEqual(t, post.PostID, postID, "Post ID should match")
			AssertEqual(t, post.Title, "Test Post Title", "Post title should match")
			AssertEqual(t, post.Content, "This is a test post content", "Post content should match")
			AssertEqual(t, post.UserUserID, userIDs[0], "Post user ID should match")
		})

		t.Run("EmptyTitle", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"", // Empty title
				"This is a test post content",
				[]string{"Technology"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for empty title")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("EmptyContent", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"Test Post Title",
				"", // Empty content
				[]string{"Technology"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for empty content")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("WhitespaceOnlyTitle", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"   ", // Whitespace only title
				"This is a test post content",
				[]string{"Technology"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for whitespace-only title")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("WhitespaceOnlyContent", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"Test Post Title",
				"   ", // Whitespace only content
				[]string{"Technology"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for whitespace-only content")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			postID, err := postService.CreatePost(
				99999, // Invalid user ID
				"Test Post Title",
				"This is a test post content",
				[]string{"Technology"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for invalid user ID")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("InvalidCategories", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"Test Post Title",
				"This is a test post content",
				[]string{"InvalidCategory"},
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for invalid categories")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("EmptyCategories", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"Test Post Title",
				"This is a test post content",
				[]string{}, // Empty categories
			)
			AssertNotEqual(t, err, nil, "Post creation should fail for empty categories")
			AssertEqual(t, postID, 0, "Post ID should be zero")
		})

		t.Run("MultipleValidCategories", func(t *testing.T) {
			postID, err := postService.CreatePost(
				userIDs[0],
				"Multi-Category Post",
				"This post belongs to multiple categories",
				[]string{"Technology", "Programming", "Science"},
			)
			AssertNoError(t, err, "Post creation should succeed with multiple categories")
			AssertNotEqual(t, postID, 0, "Post ID should be set")

			// Verify categories were assigned
			post, err := postService.GetPostByID(postID)
			AssertNoError(t, err, "Should retrieve created post")
			AssertEqual(t, len(post.Categories), 3, "Should have 3 categories")
		})
	})

	t.Run("AddComment", func(t *testing.T) {
		// Setup test data
		userIDs, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		postIDs, err := SetupTestPosts(testDB.DB, userIDs)
		AssertNoError(t, err, "Failed to setup test posts")

		t.Run("ValidComment", func(t *testing.T) {
			err := postService.AddComment(postIDs[0], userIDs[0], "This is a test comment")
			AssertNoError(t, err, "Comment addition should succeed")

			// Verify comment was added by checking post details
			post, err := postService.GetPostByID(postIDs[0])
			AssertNoError(t, err, "Should retrieve post with comment")
			AssertGreaterThan(t, post.Comments, 0, "Post should have comments")
		})

		t.Run("EmptyComment", func(t *testing.T) {
			err := postService.AddComment(postIDs[0], userIDs[0], "")
			AssertNotEqual(t, err, nil, "Comment addition should fail for empty content")
		})

		t.Run("WhitespaceOnlyComment", func(t *testing.T) {
			err := postService.AddComment(postIDs[0], userIDs[0], "   ")
			AssertNotEqual(t, err, nil, "Comment addition should fail for whitespace-only content")
		})

		t.Run("InvalidPostID", func(t *testing.T) {
			err := postService.AddComment(99999, userIDs[0], "This is a test comment")
			AssertNotEqual(t, err, nil, "Comment addition should fail for invalid post ID")
		})

		t.Run("InvalidUserID", func(t *testing.T) {
			err := postService.AddComment(postIDs[0], 99999, "This is a test comment")
			AssertNotEqual(t, err, nil, "Comment addition should fail for invalid user ID")
		})

		t.Run("ZeroPostID", func(t *testing.T) {
			err := postService.AddComment(0, userIDs[0], "This is a test comment")
			AssertNotEqual(t, err, nil, "Comment addition should fail for zero post ID")
		})

		t.Run("ZeroUserID", func(t *testing.T) {
			err := postService.AddComment(postIDs[0], 0, "This is a test comment")
			AssertNotEqual(t, err, nil, "Comment addition should fail for zero user ID")
		})
	})

	t.Run("GetCategories", func(t *testing.T) {
		categories, err := postService.GetCategories()
		AssertNoError(t, err, "Should retrieve categories")
		AssertGreaterThan(t, len(categories), 0, "Should return categories")

		// Check for expected categories
		expectedCategories := []string{"Technology", "Programming", "Science", "Art", "Music", "Sports"}
		categoryMap := make(map[string]bool)
		for _, category := range categories {
			categoryMap[category.Name] = true
		}

		for _, expected := range expectedCategories {
			AssertEqual(t, categoryMap[expected], true, "Should contain category: "+expected)
		}
	})
}
