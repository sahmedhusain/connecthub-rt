package unit_testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"connecthub/database"
	"connecthub/server"
)

func TestGetPosts(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	postIDs, err := SetupTestPosts(testDB.DB, userIDs)
	AssertNoError(t, err, "Failed to setup test posts")

	t.Run("GetAllPosts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/posts", nil)
		w := httptest.NewRecorder()

		server.GetPosts(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var posts []database.Post
		err := json.Unmarshal(w.Body.Bytes(), &posts)
		AssertNoError(t, err, "Failed to unmarshal posts")

		AssertGreaterThan(t, len(posts), 0, "Should return posts")
		AssertEqual(t, len(posts), len(postIDs), "Should return all posts")
	})

	t.Run("GetPostsWithFilter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/posts?filter=top-rated", nil)
		w := httptest.NewRecorder()

		server.GetPosts(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var posts []database.Post
		err := json.Unmarshal(w.Body.Bytes(), &posts)
		AssertNoError(t, err, "Failed to unmarshal posts")

		AssertGreaterThanOrEqual(t, len(posts), 0, "Should return posts or empty array")
	})

	t.Run("GetPostsWithCategory", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/posts?tab=Technology", nil)
		w := httptest.NewRecorder()

		server.GetPosts(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var posts []database.Post
		err := json.Unmarshal(w.Body.Bytes(), &posts)
		AssertNoError(t, err, "Failed to unmarshal posts")

		// Verify all posts belong to Technology category
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

	t.Run("GetPostsWithSession", func(t *testing.T) {
		sessionToken := CreateTestSession(t, testDB, userIDs[0])

		req := httptest.NewRequest("GET", "/api/posts", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})
		w := httptest.NewRecorder()

		server.GetPosts(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var posts []database.Post
		err := json.Unmarshal(w.Body.Bytes(), &posts)
		AssertNoError(t, err, "Failed to unmarshal posts")

		AssertGreaterThan(t, len(posts), 0, "Should return posts")
	})
}

func TestGetPostByID(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	postIDs, err := SetupTestPosts(testDB.DB, userIDs)
	AssertNoError(t, err, "Failed to setup test posts")

	t.Run("ValidPostID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/post?id="+strconv.Itoa(postIDs[0]), nil)
		w := httptest.NewRecorder()

		server.GetPostByID(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var post database.Post
		err := json.Unmarshal(w.Body.Bytes(), &post)
		AssertNoError(t, err, "Failed to unmarshal post")

		AssertEqual(t, post.PostID, postIDs[0], "Post ID should match")
		AssertNotEqual(t, post.Title, "", "Post title should not be empty")
		AssertNotEqual(t, post.Content, "", "Post content should not be empty")
	})

	t.Run("InvalidPostID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/post?id=99999", nil)
		w := httptest.NewRecorder()

		server.GetPostByID(w, req)

		AssertEqual(t, w.Code, http.StatusNotFound, "Expected status Not Found")
	})

	t.Run("MissingPostID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/post", nil)
		w := httptest.NewRecorder()

		server.GetPostByID(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})

	t.Run("InvalidPostIDFormat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/post?id=invalid", nil)
		w := httptest.NewRecorder()

		server.GetPostByID(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})
}

func TestCreatePostAPI(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	sessionToken := CreateTestSession(t, testDB, userIDs[0])

	t.Run("ValidPostCreation", func(t *testing.T) {
		createReq := server.CreatePostRequest{
			Title:      "Test Post",
			Content:    "This is a test post content",
			Categories: []string{"Technology", "Programming"},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/post/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreatePostAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var response server.CreatePostResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, true, "Post creation should succeed")
		AssertNotEqual(t, response.PostID, 0, "Post ID should be set")
	})

	t.Run("PostCreationWithoutSession", func(t *testing.T) {
		createReq := server.CreatePostRequest{
			Title:      "Test Post",
			Content:    "This is a test post content",
			Categories: []string{"Technology"},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/post/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.CreatePostAPI(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})

	t.Run("PostCreationWithEmptyTitle", func(t *testing.T) {
		createReq := server.CreatePostRequest{
			Title:      "", // Empty title
			Content:    "This is a test post content",
			Categories: []string{"Technology"},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/post/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreatePostAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.CreatePostResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Post creation should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("PostCreationWithEmptyContent", func(t *testing.T) {
		createReq := server.CreatePostRequest{
			Title:      "Test Post",
			Content:    "", // Empty content
			Categories: []string{"Technology"},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/post/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreatePostAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.CreatePostResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Post creation should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("PostCreationWithInvalidCategories", func(t *testing.T) {
		createReq := server.CreatePostRequest{
			Title:      "Test Post",
			Content:    "This is a test post content",
			Categories: []string{"InvalidCategory"},
		}

		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/post/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.CreatePostAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.CreatePostResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Post creation should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})
}

func TestAddComment(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test data
	userIDs, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	postIDs, err := SetupTestPosts(testDB.DB, userIDs)
	AssertNoError(t, err, "Failed to setup test posts")

	sessionToken := CreateTestSession(t, testDB, userIDs[0])

	t.Run("ValidCommentAddition", func(t *testing.T) {
		form := url.Values{}
		form.Add("post_id", strconv.Itoa(postIDs[0]))
		form.Add("content", "This is a test comment")

		req := httptest.NewRequest("POST", "/addcomment", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.AddComment(w, req)

		// Should redirect back to the post
		AssertEqual(t, w.Code, http.StatusSeeOther, "Expected status See Other (redirect)")

		location := w.Header().Get("Location")
		expectedLocation := "/post?id=" + strconv.Itoa(postIDs[0])
		AssertEqual(t, location, expectedLocation, "Should redirect to post page")
	})

	t.Run("CommentAdditionWithoutSession", func(t *testing.T) {
		form := url.Values{}
		form.Add("post_id", strconv.Itoa(postIDs[0]))
		form.Add("content", "This is a test comment")

		req := httptest.NewRequest("POST", "/addcomment", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		server.AddComment(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")
	})

	t.Run("CommentAdditionWithEmptyContent", func(t *testing.T) {
		form := url.Values{}
		form.Add("post_id", strconv.Itoa(postIDs[0]))
		form.Add("content", "") // Empty content

		req := httptest.NewRequest("POST", "/addcomment", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.AddComment(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})

	t.Run("CommentAdditionWithInvalidPostID", func(t *testing.T) {
		form := url.Values{}
		form.Add("post_id", "99999") // Non-existent post ID
		form.Add("content", "This is a test comment")

		req := httptest.NewRequest("POST", "/addcomment", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.AddComment(w, req)

		AssertEqual(t, w.Code, http.StatusInternalServerError, "Expected status Internal Server Error")
	})

	t.Run("CommentAdditionWithMissingPostID", func(t *testing.T) {
		form := url.Values{}
		// Missing post_id
		form.Add("content", "This is a test comment")

		req := httptest.NewRequest("POST", "/addcomment", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.AddComment(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})
}

func TestCategoriesAPI(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	t.Run("GetCategories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/categories", nil)
		w := httptest.NewRecorder()

		server.CategoriesAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var categories []string
		err := json.Unmarshal(w.Body.Bytes(), &categories)
		AssertNoError(t, err, "Failed to unmarshal categories")

		AssertGreaterThan(t, len(categories), 0, "Should return categories")

		// Check for expected categories
		expectedCategories := []string{"Technology", "Programming", "Science", "Art", "Music", "Sports"}
		for _, expected := range expectedCategories {
			found := false
			for _, category := range categories {
				if category == expected {
					found = true
					break
				}
			}
			AssertEqual(t, found, true, "Should contain category: "+expected)
		}
	})
}
