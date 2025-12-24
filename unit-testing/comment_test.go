package unit_testing

import (
	"fmt"
	"testing"
	"time"

	"connecthub/database"
	"connecthub/repository"
	"connecthub/server/services"
)

func TestCommentCreation(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	userIDs, postIDs, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	postService := services.NewPostService(testDB.DB)

	t.Run("ValidCommentCreation", func(t *testing.T) {
		// Add comment to existing post
		err := postService.AddComment(postIDs[0], userIDs[1], "This is a test comment")
		AssertNoError(t, err, "Comment creation should succeed")

		// Verify comment was added
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[0])
		AssertNoError(t, err, "Should be able to get comments")

		// Find our test comment
		found := false
		for _, comment := range comments {
			if comment.Content == "This is a test comment" && comment.UserID == userIDs[1] {
				found = true
				break
			}
		}
		AssertTrue(t, found, "Test comment should be found")
	})

	t.Run("CommentWithEmptyContent", func(t *testing.T) {
		// Try to add comment with empty content
		err := postService.AddComment(postIDs[0], userIDs[0], "")
		AssertError(t, err, "Comment creation should fail with empty content")
	})

	t.Run("CommentWithWhitespaceOnly", func(t *testing.T) {
		// Try to add comment with whitespace-only content
		err := postService.AddComment(postIDs[0], userIDs[0], "   ")
		AssertError(t, err, "Comment creation should fail with whitespace-only content")
	})

	t.Run("CommentOnNonexistentPost", func(t *testing.T) {
		// Try to add comment to non-existent post
		err := postService.AddComment(99999, userIDs[0], "Comment on non-existent post")
		AssertError(t, err, "Comment creation should fail on non-existent post")
	})

	t.Run("CommentWithInvalidUserID", func(t *testing.T) {
		// Try to add comment with invalid user ID
		err := postService.AddComment(postIDs[0], 99999, "Comment from invalid user")
		AssertError(t, err, "Comment creation should fail with invalid user ID")
	})

	t.Run("MultipleCommentsOnSamePost", func(t *testing.T) {
		// Add multiple comments to the same post
		for i := 0; i < 3; i++ {
			err := postService.AddComment(postIDs[1], userIDs[i%len(userIDs)],
				fmt.Sprintf("Comment %d on post", i+1))
			AssertNoError(t, err, fmt.Sprintf("Comment %d creation should succeed", i+1))
		}

		// Verify all comments were added
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[1])
		AssertNoError(t, err, "Should be able to get comments")

		// Count our test comments
		testCommentCount := 0
		for _, comment := range comments {
			if comment.Content == "Comment 1 on post" ||
				comment.Content == "Comment 2 on post" ||
				comment.Content == "Comment 3 on post" {
				testCommentCount++
			}
		}
		AssertTrue(t, testCommentCount >= 3, "Should have at least 3 test comments")
	})
}

func TestCommentRetrieval(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	userIDs, postIDs, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	t.Run("GetCommentsForPost", func(t *testing.T) {
		// Get comments for a specific post
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[0])
		AssertNoError(t, err, "Should be able to get comments for post")

		// Verify comment data structure
		for _, comment := range comments {
			AssertTrue(t, comment.ID > 0, "Comment ID should be positive")
			AssertEqual(t, comment.PostID, postIDs[0], "Comment should belong to correct post")
			AssertTrue(t, comment.UserID > 0, "Comment user ID should be positive")
			AssertNotEqual(t, comment.Content, "", "Comment content should not be empty")
			AssertNotEqual(t, comment.Username, "", "Comment should include username")
			AssertNotEqual(t, comment.FirstName, "", "Comment should include first name")
			AssertNotEqual(t, comment.LastName, "", "Comment should include last name")
		}
	})

	t.Run("GetCommentsForNonexistentPost", func(t *testing.T) {
		// Get comments for non-existent post
		comments, err := database.GetCommentsForPost(testDB.DB, 99999)
		AssertNoError(t, err, "Should not error for non-existent post")
		AssertEqual(t, len(comments), 0, "Should return empty list for non-existent post")
	})

	t.Run("CommentsOrderedByDate", func(t *testing.T) {
		// Add comments with specific timing
		postService := services.NewPostService(testDB.DB)

		// Add multiple comments to ensure ordering
		for i := 0; i < 3; i++ {
			err := postService.AddComment(postIDs[2], userIDs[0],
				fmt.Sprintf("Ordered comment %d", i+1))
			AssertNoError(t, err, "Comment creation should succeed")
		}

		// Get comments and verify ordering
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[2])
		AssertNoError(t, err, "Should be able to get comments")

		if len(comments) > 1 {
			// Comments should be ordered by creation time
			for i := 0; i < len(comments)-1; i++ {
				AssertTrue(t, comments[i].CreatedAt.Before(comments[i+1].CreatedAt) ||
					comments[i].CreatedAt.Equal(comments[i+1].CreatedAt),
					"Comments should be ordered by creation time")
			}
		}
	})

	t.Run("CommentWithUserInformation", func(t *testing.T) {
		// Get comments and verify user information is included
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[0])
		AssertNoError(t, err, "Should be able to get comments")

		if len(comments) > 0 {
			comment := comments[0]
			AssertNotEqual(t, comment.Username, "", "Comment should include username")
			AssertNotEqual(t, comment.FirstName, "", "Comment should include first name")
			AssertNotEqual(t, comment.LastName, "", "Comment should include last name")

			// Verify avatar is included if available
			if comment.Avatar.Valid {
				AssertNotEqual(t, comment.Avatar.String, "", "Avatar should not be empty if valid")
			}
		}
	})
}

func TestCommentValidation(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	userIDs, postIDs, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	postService := services.NewPostService(testDB.DB)

	t.Run("CommentContentValidation", func(t *testing.T) {
		testCases := []struct {
			content     string
			shouldPass  bool
			description string
		}{
			{"Valid comment content", true, "Valid comment should pass"},
			{"", false, "Empty comment should fail"},
			{"   ", false, "Whitespace-only comment should fail"},
			{"A", true, "Single character comment should pass"},
			{"This is a very long comment that contains a lot of text to test if there are any length restrictions on comments in the system", true, "Long comment should pass"},
		}

		for _, tc := range testCases {
			err := postService.AddComment(postIDs[0], userIDs[0], tc.content)

			if tc.shouldPass {
				AssertNoError(t, err, tc.description)
			} else {
				AssertError(t, err, tc.description)
			}
		}
	})

	t.Run("CommentUserValidation", func(t *testing.T) {
		// Test with valid user
		err := postService.AddComment(postIDs[0], userIDs[0], "Valid user comment")
		AssertNoError(t, err, "Comment with valid user should succeed")

		// Test with invalid user ID
		err = postService.AddComment(postIDs[0], 0, "Invalid user comment")
		AssertError(t, err, "Comment with invalid user ID should fail")

		// Test with negative user ID
		err = postService.AddComment(postIDs[0], -1, "Negative user ID comment")
		AssertError(t, err, "Comment with negative user ID should fail")
	})

	t.Run("CommentPostValidation", func(t *testing.T) {
		// Test with valid post
		err := postService.AddComment(postIDs[0], userIDs[0], "Valid post comment")
		AssertNoError(t, err, "Comment on valid post should succeed")

		// Test with invalid post ID
		err = postService.AddComment(0, userIDs[0], "Invalid post comment")
		AssertError(t, err, "Comment on invalid post should fail")

		// Test with negative post ID
		err = postService.AddComment(-1, userIDs[0], "Negative post ID comment")
		AssertError(t, err, "Comment on negative post ID should fail")
	})
}

func TestCommentIntegration(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	userIDs, postIDs, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	postService := services.NewPostService(testDB.DB)
	postRepo := repository.NewPostRepository(testDB.DB)

	t.Run("CommentCountInPost", func(t *testing.T) {
		// Get initial comment count
		post, err := postRepo.GetPostByID(postIDs[0])
		AssertNoError(t, err, "Should be able to get post")
		initialCommentCount := post.Comments

		// Add a comment
		err = postService.AddComment(postIDs[0], userIDs[0], "New comment for count test")
		AssertNoError(t, err, "Comment creation should succeed")

		// Get updated post and verify comment count increased
		updatedPost, err := postRepo.GetPostByID(postIDs[0])
		AssertNoError(t, err, "Should be able to get updated post")
		AssertEqual(t, updatedPost.Comments, initialCommentCount+1, "Comment count should increase")
	})

	t.Run("CommentsByDifferentUsers", func(t *testing.T) {
		// Add comments from different users
		for i, userID := range userIDs {
			err := postService.AddComment(postIDs[1], userID,
				fmt.Sprintf("Comment from user %d", i+1))
			AssertNoError(t, err, "Comment creation should succeed")
		}

		// Get comments and verify different users
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[1])
		AssertNoError(t, err, "Should be able to get comments")

		// Check that we have comments from different users
		userSet := make(map[int]bool)
		for _, comment := range comments {
			userSet[comment.UserID] = true
		}
		AssertTrue(t, len(userSet) >= 2, "Should have comments from multiple users")
	})

	t.Run("CommentTimestamps", func(t *testing.T) {
		// Add comment and verify timestamp
		err := postService.AddComment(postIDs[2], userIDs[0], "Timestamp test comment")
		AssertNoError(t, err, "Comment creation should succeed")

		// Get comments and verify timestamp is recent
		comments, err := database.GetCommentsForPost(testDB.DB, postIDs[2])
		AssertNoError(t, err, "Should be able to get comments")

		// Find our test comment
		found := false
		for _, comment := range comments {
			if comment.Content == "Timestamp test comment" {
				found = true
				// Verify timestamp is recent (within last minute)
				AssertTrue(t, comment.CreatedAt.After(time.Now().Add(-time.Minute)),
					"Comment timestamp should be recent")
				break
			}
		}
		AssertTrue(t, found, "Test comment should be found")
	})
}

func TestCommentRepository(t *testing.T) {
	testDB := TestSetup(t)

	// Setup test data
	userIDs, postIDs, _, err := SetupCompleteTestData(testDB.DB)
	AssertNoError(t, err, "Failed to setup test data")

	postRepo := repository.NewPostRepository(testDB.DB)

	t.Run("AddCommentRepository", func(t *testing.T) {
		// Add comment directly through repository
		err := postRepo.AddComment(postIDs[0], userIDs[0], "Repository comment test")
		AssertNoError(t, err, "Repository comment creation should succeed")

		// Verify comment was added
		comments, err := postRepo.GetCommentsForPost(postIDs[0])
		AssertNoError(t, err, "Should be able to get comments")

		// Find our test comment
		found := false
		for _, comment := range comments {
			if comment.Content == "Repository comment test" {
				found = true
				AssertEqual(t, comment.UserID, userIDs[0], "Comment user ID should match")
				AssertEqual(t, comment.PostID, postIDs[0], "Comment post ID should match")
				break
			}
		}
		AssertTrue(t, found, "Repository comment should be found")
	})

	t.Run("GetCommentsForPostRepository", func(t *testing.T) {
		// Add multiple comments
		for i := 0; i < 3; i++ {
			err := postRepo.AddComment(postIDs[1], userIDs[i%len(userIDs)],
				fmt.Sprintf("Repo comment %d", i+1))
			AssertNoError(t, err, "Comment creation should succeed")
		}

		// Get comments through repository
		comments, err := postRepo.GetCommentsForPost(postIDs[1])
		AssertNoError(t, err, "Should be able to get comments through repository")

		// Verify we have the expected comments
		repoCommentCount := 0
		for _, comment := range comments {
			if comment.Content == "Repo comment 1" ||
				comment.Content == "Repo comment 2" ||
				comment.Content == "Repo comment 3" {
				repoCommentCount++
			}
		}
		AssertTrue(t, repoCommentCount >= 3, "Should have at least 3 repository comments")
	})
}
