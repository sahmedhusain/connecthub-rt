package unit_testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// HTTPTestHelper provides utilities for HTTP testing
type HTTPTestHelper struct {
	Server *httptest.Server
	Client *http.Client
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(handler http.Handler) *HTTPTestHelper {
	server := httptest.NewServer(handler)
	return &HTTPTestHelper{
		Server: server,
		Client: server.Client(),
	}
}

// Close closes the test server
func (h *HTTPTestHelper) Close() {
	h.Server.Close()
}

// GET performs a GET request
func (h *HTTPTestHelper) GET(path string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("GET", h.Server.URL+path, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.Client.Do(req)
}

// POST performs a POST request with JSON body
func (h *HTTPTestHelper) POST(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("POST", h.Server.URL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.Client.Do(req)
}

// POSTForm performs a POST request with form data
func (h *HTTPTestHelper) POSTForm(path string, formData url.Values, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("POST", h.Server.URL+path, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.Client.Do(req)
}

// PUT performs a PUT request with JSON body
func (h *HTTPTestHelper) PUT(path string, body interface{}, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest("PUT", h.Server.URL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.Client.Do(req)
}

// DELETE performs a DELETE request
func (h *HTTPTestHelper) DELETE(path string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", h.Server.URL+path, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return h.Client.Do(req)
}

// AssertStatusCode checks if the response has the expected status code
func AssertStatusCode(t *testing.T, resp *http.Response, expectedCode int) {
	if resp.StatusCode != expectedCode {
		t.Fatalf("Expected status code %d, got %d", expectedCode, resp.StatusCode)
	}
}

// AssertJSONResponse checks if the response is valid JSON and unmarshals it
func AssertJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v\nBody: %s", err, string(body))
	}
}

// AssertResponseContains checks if the response body contains a specific string
func AssertResponseContains(t *testing.T, resp *http.Response, expected string) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	bodyStr := string(body)
	if !strings.Contains(bodyStr, expected) {
		t.Fatalf("Response body does not contain expected string '%s'\nBody: %s", expected, bodyStr)
	}
}

// AssertResponseNotContains checks if the response body does not contain a specific string
func AssertResponseNotContains(t *testing.T, resp *http.Response, unexpected string) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()

	bodyStr := string(body)
	if strings.Contains(bodyStr, unexpected) {
		t.Fatalf("Response body contains unexpected string '%s'\nBody: %s", unexpected, bodyStr)
	}
}

// GetResponseBody reads and returns the response body as a string
func GetResponseBody(t *testing.T, resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	defer resp.Body.Close()
	return string(body)
}

// AssertHeader checks if a response header has the expected value
func AssertHeader(t *testing.T, resp *http.Response, headerName, expectedValue string) {
	actualValue := resp.Header.Get(headerName)
	if actualValue != expectedValue {
		t.Fatalf("Expected header %s to be '%s', got '%s'", headerName, expectedValue, actualValue)
	}
}

// AssertHeaderExists checks if a response header exists
func AssertHeaderExists(t *testing.T, resp *http.Response, headerName string) {
	if resp.Header.Get(headerName) == "" {
		t.Fatalf("Expected header %s to exist", headerName)
	}
}

// AssertCookie checks if a response cookie has the expected value
func AssertCookie(t *testing.T, resp *http.Response, cookieName, expectedValue string) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == cookieName {
			if cookie.Value != expectedValue {
				t.Fatalf("Expected cookie %s to be '%s', got '%s'", cookieName, expectedValue, cookie.Value)
			}
			return
		}
	}
	t.Fatalf("Cookie %s not found in response", cookieName)
}

// AssertCookieExists checks if a response cookie exists
func AssertCookieExists(t *testing.T, resp *http.Response, cookieName string) {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == cookieName {
			return
		}
	}
	t.Fatalf("Cookie %s not found in response", cookieName)
}

// LoginTestUser performs a login request and returns the session cookie
func (h *HTTPTestHelper) LoginTestUser(username, password string) (*http.Cookie, error) {
	loginData := map[string]string{
		"identifier": username,
		"password":   password,
	}

	resp, err := h.POST("/api/login", loginData, nil)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Find session cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session_token" {
			return cookie, nil
		}
	}

	return nil, fmt.Errorf("session cookie not found")
}

// AuthenticatedRequest performs a request with authentication cookie
func (h *HTTPTestHelper) AuthenticatedRequest(method, path string, body interface{}, sessionCookie *http.Cookie) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON body: %v", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, h.Server.URL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}

	return h.Client.Do(req)
}

// MockWebSocketConnection represents a mock WebSocket connection for testing
type MockWebSocketConnection struct {
	Messages []interface{}
	Closed   bool
}

// SendMessage simulates sending a message through WebSocket
func (m *MockWebSocketConnection) SendMessage(message interface{}) {
	if !m.Closed {
		m.Messages = append(m.Messages, message)
	}
}

// Close simulates closing the WebSocket connection
func (m *MockWebSocketConnection) Close() {
	m.Closed = true
}

// GetLastMessage returns the last message sent through the connection
func (m *MockWebSocketConnection) GetLastMessage() interface{} {
	if len(m.Messages) == 0 {
		return nil
	}
	return m.Messages[len(m.Messages)-1]
}

// GetMessageCount returns the number of messages sent
func (m *MockWebSocketConnection) GetMessageCount() int {
	return len(m.Messages)
}

// ClearMessages clears all messages from the connection
func (m *MockWebSocketConnection) ClearMessages() {
	m.Messages = []interface{}{}
}

// NewMockWebSocketConnection creates a new mock WebSocket connection
func NewMockWebSocketConnection() *MockWebSocketConnection {
	return &MockWebSocketConnection{
		Messages: make([]interface{}, 0),
		Closed:   false,
	}
}
