package unit_testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connecthub/server"
)

func TestLoginAPI(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test users
	_, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	t.Run("ValidLogin", func(t *testing.T) {
		loginReq := server.LoginRequest{
			Identifier: "johndoe",
			Password:   "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.LoginAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var response server.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, true, "Login should succeed")
		AssertEqual(t, response.Username, "johndoe", "Username should match")
		AssertEqual(t, response.Email, "john@example.com", "Email should match")

		// Check if session cookie is set
		cookies := w.Result().Cookies()
		sessionCookieFound := false
		for _, cookie := range cookies {
			if cookie.Name == "session_token" && cookie.Value != "" {
				sessionCookieFound = true
				break
			}
		}
		AssertEqual(t, sessionCookieFound, true, "Session cookie should be set")
	})

	t.Run("LoginWithEmail", func(t *testing.T) {
		loginReq := server.LoginRequest{
			Identifier: "jane@example.com",
			Password:   "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.LoginAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var response server.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, true, "Login should succeed")
		AssertEqual(t, response.Username, "janesmith", "Username should match")
		AssertEqual(t, response.Email, "jane@example.com", "Email should match")
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		loginReq := server.LoginRequest{
			Identifier: "johndoe",
			Password:   "wrongpassword",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.LoginAPI(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")

		var response server.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Login should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		loginReq := server.LoginRequest{
			Identifier: "nonexistent",
			Password:   "password123",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.LoginAPI(w, req)

		AssertEqual(t, w.Code, http.StatusUnauthorized, "Expected status Unauthorized")

		var response server.LoginResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Login should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("EmptyCredentials", func(t *testing.T) {
		loginReq := server.LoginRequest{
			Identifier: "",
			Password:   "",
		}

		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.LoginAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/login", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.LoginAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")
	})
}

func TestSignupAPI(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	t.Run("ValidSignup", func(t *testing.T) {
		signupReq := server.SignupRequest{
			FirstName:   "New",
			LastName:    "User",
			Username:    "newuser",
			Email:       "newuser@example.com",
			Gender:      "male",
			DateOfBirth: "1990-01-01",
			Password:    "password123",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.SignupAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		var response server.SignupResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, true, "Signup should succeed")
		AssertEqual(t, response.Username, "newuser", "Username should match")
		AssertEqual(t, response.Email, "newuser@example.com", "Email should match")
		AssertNotEqual(t, response.UserID, 0, "User ID should be set")

		// Check if session cookie is set
		cookies := w.Result().Cookies()
		sessionCookieFound := false
		for _, cookie := range cookies {
			if cookie.Name == "session_token" && cookie.Value != "" {
				sessionCookieFound = true
				break
			}
		}
		AssertEqual(t, sessionCookieFound, true, "Session cookie should be set")
	})

	t.Run("DuplicateUsername", func(t *testing.T) {
		// Setup existing user
		_, err := SetupTestUsers(testDB.DB)
		AssertNoError(t, err, "Failed to setup test users")

		signupReq := server.SignupRequest{
			FirstName:   "Duplicate",
			LastName:    "User",
			Username:    "johndoe", // This username already exists
			Email:       "duplicate@example.com",
			Gender:      "male",
			DateOfBirth: "1990-01-01",
			Password:    "password123",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.SignupAPI(w, req)

		AssertEqual(t, w.Code, http.StatusConflict, "Expected status Conflict")

		var response server.SignupResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Signup should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		signupReq := server.SignupRequest{
			FirstName:   "Duplicate",
			LastName:    "User",
			Username:    "duplicateuser",
			Email:       "john@example.com", // This email already exists
			Gender:      "male",
			DateOfBirth: "1990-01-01",
			Password:    "password123",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.SignupAPI(w, req)

		AssertEqual(t, w.Code, http.StatusConflict, "Expected status Conflict")

		var response server.SignupResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Signup should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		signupReq := server.SignupRequest{
			FirstName:   "Invalid",
			LastName:    "Email",
			Username:    "invalidemail",
			Email:       "invalid-email", // Invalid email format
			Gender:      "male",
			DateOfBirth: "1990-01-01",
			Password:    "password123",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.SignupAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.SignupResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Signup should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		signupReq := server.SignupRequest{
			FirstName: "Missing",
			// LastName is missing
			Username:    "missingfields",
			Email:       "missing@example.com",
			Gender:      "male",
			DateOfBirth: "1990-01-01",
			Password:    "password123",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.SignupAPI(w, req)

		AssertEqual(t, w.Code, http.StatusBadRequest, "Expected status Bad Request")

		var response server.SignupResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		AssertNoError(t, err, "Failed to unmarshal response")

		AssertEqual(t, response.Success, false, "Signup should fail")
		AssertNotEqual(t, response.Error, "", "Error message should be present")
	})
}

func TestLogoutAPI(t *testing.T) {
	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test users and create a session
	_, err := SetupTestUsers(testDB.DB)
	AssertNoError(t, err, "Failed to setup test users")

	// Create a session for testing
	sessionToken := CreateTestSession(t, testDB, 1)

	t.Run("ValidLogout", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/logout", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})

		w := httptest.NewRecorder()
		server.LogoutAPI(w, req)

		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")

		// Check if session cookie is cleared
		cookies := w.Result().Cookies()
		sessionCookieCleared := false
		for _, cookie := range cookies {
			if cookie.Name == "session_token" && cookie.Value == "" {
				sessionCookieCleared = true
				break
			}
		}
		AssertEqual(t, sessionCookieCleared, true, "Session cookie should be cleared")
	})

	t.Run("LogoutWithoutSession", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/logout", nil)

		w := httptest.NewRecorder()
		server.LogoutAPI(w, req)

		// Should still succeed even without session
		AssertEqual(t, w.Code, http.StatusOK, "Expected status OK")
	})
}
