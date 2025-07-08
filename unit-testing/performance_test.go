package unit_testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"forum/server"
)

// benchmarkTestSetup creates a test database for benchmarks
func benchmarkTestSetup(b *testing.B) *TestDatabase {
	// Create a dummy test to get TestSetup working
	t := &testing.T{}
	return TestSetup(t)
}

// benchmarkCreateTestSession creates a test session for benchmarks
func benchmarkCreateTestSession(b *testing.B, testDB *TestDatabase, userID int) string {
	// Create a dummy test to get CreateTestSession working
	t := &testing.T{}
	return CreateTestSession(t, testDB, userID)
}

// PerformanceMetrics holds performance test results
type PerformanceMetrics struct {
	TotalRequests     int           `json:"total_requests"`
	SuccessfulReqs    int           `json:"successful_requests"`
	FailedReqs        int           `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	MinLatency        time.Duration `json:"min_latency"`
	MaxLatency        time.Duration `json:"max_latency"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	TotalDuration     time.Duration `json:"total_duration"`
	P95Latency        time.Duration `json:"p95_latency"`
	P99Latency        time.Duration `json:"p99_latency"`
	ErrorRate         float64       `json:"error_rate"`
}

// LoadTestConfig defines load test parameters
type LoadTestConfig struct {
	ConcurrentUsers int           `json:"concurrent_users"`
	Duration        time.Duration `json:"duration"`
	RequestsPerUser int           `json:"requests_per_user"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	ThinkTime       time.Duration `json:"think_time"`
}

// BenchmarkUserRegistration tests user registration performance
func BenchmarkUserRegistration(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			signupReq := server.SignupRequest{
				FirstName:   fmt.Sprintf("Bench%d", counter),
				LastName:    "User",
				Username:    fmt.Sprintf("benchuser%d_%d", counter, time.Now().UnixNano()),
				Email:       fmt.Sprintf("bench%d@example.com", counter),
				Gender:      "other",
				DateOfBirth: "1990-01-01",
				Password:    "password123",
			}

			body, _ := json.Marshal(signupReq)
			req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.SignupAPI(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkUserLogin tests login performance
func BenchmarkUserLogin(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		b.Fatalf("Failed to setup test users: %v", err)
	}

	handler := createTestServer(testDB)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		userIndex := 0
		for pb.Next() {
			// Cycle through test users
			userIndex = (userIndex + 1) % len(userIDs)

			loginReq := server.LoginRequest{
				Identifier: fmt.Sprintf("testuser%d", userIndex+1),
				Password:   "Aa123456",
			}

			body, _ := json.Marshal(loginReq)
			req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkPostCreation tests post creation performance
func BenchmarkPostCreation(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		b.Fatalf("Failed to setup test users: %v", err)
	}

	handler := createTestServer(testDB)

	// Create sessions for users
	sessions := make([]string, len(userIDs))
	for i, userID := range userIDs {
		sessions[i] = benchmarkCreateTestSession(b, testDB, userID)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			sessionIndex := counter % len(sessions)

			createReq := server.CreatePostRequest{
				Title:      fmt.Sprintf("Benchmark Post %d", counter),
				Content:    fmt.Sprintf("This is benchmark post content %d with sufficient length to simulate real posts", counter),
				Categories: []string{"Technology"},
			}

			body, _ := json.Marshal(createReq)
			req := httptest.NewRequest("POST", "/api/post/create", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  "session_token",
				Value: sessions[sessionIndex],
			})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkPostRetrieval tests post retrieval performance
func BenchmarkPostRetrieval(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		b.Fatalf("Failed to setup test users: %v", err)
	}

	postIDs, err := SetupTestPosts(testDB.DB, userIDs)
	if err != nil {
		b.Fatalf("Failed to setup test posts: %v", err)
	}

	handler := createTestServer(testDB)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			_ = counter % len(postIDs) // postIndex not used

			req := httptest.NewRequest("GET", "/api/posts", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkCommentCreation tests comment creation performance
func BenchmarkCommentCreation(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		b.Fatalf("Failed to setup test users: %v", err)
	}

	postIDs, err := SetupTestPosts(testDB.DB, userIDs)
	if err != nil {
		b.Fatalf("Failed to setup test posts: %v", err)
	}

	handler := createTestServer(testDB)

	// Create sessions
	sessions := make([]string, len(userIDs))
	for i, userID := range userIDs {
		sessions[i] = benchmarkCreateTestSession(b, testDB, userID)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			counter++
			sessionIndex := counter % len(sessions)
			postIndex := counter % len(postIDs)

			form := url.Values{}
			form.Add("post_id", strconv.Itoa(postIDs[postIndex]))
			form.Add("content", fmt.Sprintf("Benchmark comment %d", counter))

			req := httptest.NewRequest("POST", "/addcomment", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.AddCookie(&http.Cookie{
				Name:  "session_token",
				Value: sessions[sessionIndex],
			})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusSeeOther {
				b.Errorf("Expected status 303, got %d", w.Code)
			}
		}
	})
}

// BenchmarkMessageSending tests messaging performance
func BenchmarkMessageSending(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		b.Fatalf("Failed to setup test users: %v", err)
	}

	handler := createTestServer(testDB)

	// Create sessions
	sessions := make([]string, len(userIDs))
	for i, userID := range userIDs {
		sessions[i] = benchmarkCreateTestSession(b, testDB, userID)
	}

	// Create conversations using simple approach
	conversationIDs := make([]int, 0)
	for i := 0; i < len(userIDs)-1; i++ {
		convReq := map[string]interface{}{
			"participants": []int{userIDs[i+1]},
		}

		body, _ := json.Marshal(convReq)
		req := httptest.NewRequest("POST", "/api/conversations", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessions[i],
		})

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			if success, ok := response["success"].(bool); ok && success {
				if convID, ok := response["conversation_id"].(float64); ok {
					conversationIDs = append(conversationIDs, int(convID))
				}
			}
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			if len(conversationIDs) == 0 {
				continue
			}

			counter++
			sessionIndex := counter % len(sessions)
			convIndex := counter % len(conversationIDs)

			msgReq := map[string]interface{}{
				"conversation_id": conversationIDs[convIndex],
				"content":         fmt.Sprintf("Benchmark message %d", counter),
			}

			body, _ := json.Marshal(msgReq)
			req := httptest.NewRequest("POST", "/api/messages", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  "session_token",
				Value: sessions[sessionIndex],
			})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkDatabaseOperations tests database performance
func BenchmarkDatabaseOperations(b *testing.B) {
	testDB := benchmarkTestSetup(b)
	defer testDB.Cleanup()

	b.Run("UserInsert", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				counter++
				_, err := testDB.DB.Exec(`
					INSERT INTO user (firstname, lastname, username, email, gender, dateofbirth, password, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				`, fmt.Sprintf("Bench%d", counter), "User", fmt.Sprintf("benchuser%d_%d", counter, time.Now().UnixNano()),
					fmt.Sprintf("bench%d@example.com", counter), "other", "1990-01-01", "hashedpassword", time.Now())

				if err != nil {
					b.Errorf("Database insert failed: %v", err)
				}
			}
		})
	})

	b.Run("UserSelect", func(b *testing.B) {
		// Setup some users first
		for i := 0; i < 100; i++ {
			testDB.DB.Exec(`
				INSERT INTO user (firstname, lastname, username, email, gender, dateofbirth, password, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, fmt.Sprintf("Select%d", i), "User", fmt.Sprintf("selectuser%d", i),
				fmt.Sprintf("select%d@example.com", i), "other", "1990-01-01", "hashedpassword", time.Now())
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			counter := 0
			for pb.Next() {
				counter++
				username := fmt.Sprintf("selectuser%d", counter%100)

				var userID int
				err := testDB.DB.QueryRow("SELECT userid FROM user WHERE username = ?", username).Scan(&userID)
				if err != nil {
					b.Errorf("Database select failed: %v", err)
				}
			}
		})
	})
}

// TestLoadTestUserRegistration performs load testing on user registration
func TestLoadTestUserRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Cleanup()

	config := LoadTestConfig{
		ConcurrentUsers: 50,
		Duration:        30 * time.Second,
		RequestsPerUser: 10,
		RampUpTime:      5 * time.Second,
		ThinkTime:       100 * time.Millisecond,
	}

	metrics := runLoadTest(t, testDB, "user_registration", config, func(userID int, sessionToken string) error {
		signupReq := server.SignupRequest{
			FirstName:   fmt.Sprintf("Load%d", userID),
			LastName:    "User",
			Username:    fmt.Sprintf("loaduser%d_%d", userID, time.Now().UnixNano()),
			Email:       fmt.Sprintf("load%d@example.com", userID),
			Gender:      "other",
			DateOfBirth: "1990-01-01",
			Password:    "password123",
		}

		body, _ := json.Marshal(signupReq)
		req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		handler := createTestServer(testDB)

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			return fmt.Errorf("signup failed with status %d", w.Code)
		}
		return nil
	})

	// Validate performance requirements
	validatePerformanceMetrics(t, "UserRegistration", metrics, PerformanceThresholds{
		MaxAverageLatency:    500 * time.Millisecond,
		MinRequestsPerSecond: 50,
		MaxErrorRate:         0.01, // 1%
		MaxP95Latency:        1 * time.Second,
	})
}

// PerformanceThresholds defines acceptable performance limits
type PerformanceThresholds struct {
	MaxAverageLatency    time.Duration
	MinRequestsPerSecond float64
	MaxErrorRate         float64
	MaxP95Latency        time.Duration
}

// validatePerformanceMetrics checks if metrics meet performance requirements
func validatePerformanceMetrics(t *testing.T, testName string, metrics PerformanceMetrics, thresholds PerformanceThresholds) {
	t.Logf("Performance Results for %s:", testName)
	t.Logf("  Total Requests: %d", metrics.TotalRequests)
	t.Logf("  Successful: %d", metrics.SuccessfulReqs)
	t.Logf("  Failed: %d", metrics.FailedReqs)
	t.Logf("  Average Latency: %v", metrics.AverageLatency)
	t.Logf("  P95 Latency: %v", metrics.P95Latency)
	t.Logf("  P99 Latency: %v", metrics.P99Latency)
	t.Logf("  Requests/Second: %.2f", metrics.RequestsPerSecond)
	t.Logf("  Error Rate: %.2f%%", metrics.ErrorRate*100)

	// Check thresholds
	if metrics.AverageLatency > thresholds.MaxAverageLatency {
		t.Errorf("Average latency %v exceeds threshold %v", metrics.AverageLatency, thresholds.MaxAverageLatency)
	}

	if metrics.RequestsPerSecond < thresholds.MinRequestsPerSecond {
		t.Errorf("Requests per second %.2f below threshold %.2f", metrics.RequestsPerSecond, thresholds.MinRequestsPerSecond)
	}

	if metrics.ErrorRate > thresholds.MaxErrorRate {
		t.Errorf("Error rate %.2f%% exceeds threshold %.2f%%", metrics.ErrorRate*100, thresholds.MaxErrorRate*100)
	}

	if metrics.P95Latency > thresholds.MaxP95Latency {
		t.Errorf("P95 latency %v exceeds threshold %v", metrics.P95Latency, thresholds.MaxP95Latency)
	}
}

// runLoadTest executes a load test with the given configuration
func runLoadTest(t *testing.T, testDB *TestDatabase, testName string, config LoadTestConfig, testFunc func(int, string) error) PerformanceMetrics {
	var wg sync.WaitGroup
	var mu sync.Mutex

	latencies := make([]time.Duration, 0)
	successCount := 0
	errorCount := 0

	startTime := time.Now()

	// Ramp up users gradually
	userDelay := config.RampUpTime / time.Duration(config.ConcurrentUsers)

	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)

		go func(userID int) {
			defer wg.Done()

			// Ramp up delay
			time.Sleep(time.Duration(userID) * userDelay)

			// Execute requests for this user
			for j := 0; j < config.RequestsPerUser; j++ {
				requestStart := time.Now()

				err := testFunc(userID, "")

				requestDuration := time.Since(requestStart)

				mu.Lock()
				latencies = append(latencies, requestDuration)
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()

				// Think time between requests
				if j < config.RequestsPerUser-1 {
					time.Sleep(config.ThinkTime)
				}
			}
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate metrics
	return calculateMetrics(latencies, successCount, errorCount, totalDuration)
}

// calculateMetrics computes performance metrics from raw data
func calculateMetrics(latencies []time.Duration, successCount, errorCount int, totalDuration time.Duration) PerformanceMetrics {
	totalRequests := successCount + errorCount

	if len(latencies) == 0 {
		return PerformanceMetrics{
			TotalRequests:  totalRequests,
			SuccessfulReqs: successCount,
			FailedReqs:     errorCount,
			TotalDuration:  totalDuration,
		}
	}

	// Sort latencies for percentile calculations
	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)

	// Simple bubble sort for small datasets
	for i := 0; i < len(sortedLatencies); i++ {
		for j := 0; j < len(sortedLatencies)-1-i; j++ {
			if sortedLatencies[j] > sortedLatencies[j+1] {
				sortedLatencies[j], sortedLatencies[j+1] = sortedLatencies[j+1], sortedLatencies[j]
			}
		}
	}

	// Calculate average
	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}
	avgLatency := totalLatency / time.Duration(len(latencies))

	// Calculate percentiles
	p95Index := int(float64(len(sortedLatencies)) * 0.95)
	p99Index := int(float64(len(sortedLatencies)) * 0.99)

	if p95Index >= len(sortedLatencies) {
		p95Index = len(sortedLatencies) - 1
	}
	if p99Index >= len(sortedLatencies) {
		p99Index = len(sortedLatencies) - 1
	}

	return PerformanceMetrics{
		TotalRequests:     totalRequests,
		SuccessfulReqs:    successCount,
		FailedReqs:        errorCount,
		AverageLatency:    avgLatency,
		MinLatency:        sortedLatencies[0],
		MaxLatency:        sortedLatencies[len(sortedLatencies)-1],
		RequestsPerSecond: float64(totalRequests) / totalDuration.Seconds(),
		TotalDuration:     totalDuration,
		P95Latency:        sortedLatencies[p95Index],
		P99Latency:        sortedLatencies[p99Index],
		ErrorRate:         float64(errorCount) / float64(totalRequests),
	}
}
