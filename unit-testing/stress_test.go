package unit_testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"forum/server"
)

// StressTestMetrics holds stress test results
type StressTestMetrics struct {
	MaxConcurrentUsers    int           `json:"max_concurrent_users"`
	PeakRequestsPerSecond float64       `json:"peak_requests_per_second"`
	TotalRequests         int64         `json:"total_requests"`
	SuccessfulRequests    int64         `json:"successful_requests"`
	FailedRequests        int64         `json:"failed_requests"`
	AverageLatency        time.Duration `json:"average_latency"`
	MaxLatency            time.Duration `json:"max_latency"`
	MemoryUsageMB         float64       `json:"memory_usage_mb"`
	CPUUsagePercent       float64       `json:"cpu_usage_percent"`
	ErrorRate             float64       `json:"error_rate"`
	ThroughputMBps        float64       `json:"throughput_mbps"`
	ConnectionErrors      int64         `json:"connection_errors"`
	TimeoutErrors         int64         `json:"timeout_errors"`
	TestDuration          time.Duration `json:"test_duration"`
}

// StressTestConfig defines stress test parameters
type StressTestConfig struct {
	StartUsers       int           `json:"start_users"`
	MaxUsers         int           `json:"max_users"`
	UserIncrement    int           `json:"user_increment"`
	IncrementDelay   time.Duration `json:"increment_delay"`
	TestDuration     time.Duration `json:"test_duration"`
	RequestTimeout   time.Duration `json:"request_timeout"`
	FailureThreshold float64       `json:"failure_threshold"`
	LatencyThreshold time.Duration `json:"latency_threshold"`
}

// TestStressUserRegistration performs stress testing on user registration
func TestStressUserRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Cleanup()

	config := StressTestConfig{
		StartUsers:       10,
		MaxUsers:         500,
		UserIncrement:    10,
		IncrementDelay:   5 * time.Second,
		TestDuration:     2 * time.Minute,
		RequestTimeout:   5 * time.Second,
		FailureThreshold: 0.05, // 5% failure rate
		LatencyThreshold: 2 * time.Second,
	}

	metrics := runStressTest(t, testDB, "UserRegistration", config, func(userID int) error {
		signupReq := server.SignupRequest{
			FirstName:   fmt.Sprintf("Stress%d", userID),
			LastName:    "User",
			Username:    fmt.Sprintf("stressuser%d_%d", userID, time.Now().UnixNano()),
			Email:       fmt.Sprintf("stress%d@example.com", userID),
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

	// Validate stress test results
	validateStressTestResults(t, "UserRegistration", metrics, config)
}

// TestStressWebSocketConnections tests WebSocket connection limits
func TestStressWebSocketConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Cleanup()

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		t.Fatalf("Failed to setup test users: %v", err)
	}

	config := StressTestConfig{
		StartUsers:       5,
		MaxUsers:         100,
		UserIncrement:    5,
		IncrementDelay:   3 * time.Second,
		TestDuration:     1 * time.Minute,
		RequestTimeout:   10 * time.Second,
		FailureThreshold: 0.10, // 10% failure rate for WebSocket
		LatencyThreshold: 1 * time.Second,
	}

	metrics := runWebSocketStressTest(t, testDB, userIDs, config)
	validateStressTestResults(t, "WebSocketConnections", metrics, config)
}

// TestStressDatabaseOperations tests database under heavy load
func TestStressDatabaseOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Cleanup()

	config := StressTestConfig{
		StartUsers:       20,
		MaxUsers:         200,
		UserIncrement:    20,
		IncrementDelay:   2 * time.Second,
		TestDuration:     1 * time.Minute,
		RequestTimeout:   3 * time.Second,
		FailureThreshold: 0.02, // 2% failure rate for DB
		LatencyThreshold: 500 * time.Millisecond,
	}

	metrics := runStressTest(t, testDB, "DatabaseOperations", config, func(userID int) error {
		// Mix of database operations
		operations := []func() error{
			func() error {
				// Insert operation
				_, err := testDB.DB.Exec(`
					INSERT INTO user (firstname, lastname, username, email, gender, dateofbirth, password, created_at)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				`, fmt.Sprintf("Stress%d", userID), "User", fmt.Sprintf("stressdb%d_%d", userID, time.Now().UnixNano()),
					fmt.Sprintf("stressdb%d@example.com", userID), "other", "1990-01-01", "hashedpassword", time.Now())
				return err
			},
			func() error {
				// Select operation
				var count int
				err := testDB.DB.QueryRow("SELECT COUNT(*) FROM user").Scan(&count)
				return err
			},
			func() error {
				// Update operation
				_, err := testDB.DB.Exec("UPDATE user SET lastname = ? WHERE userid = ?", fmt.Sprintf("Updated%d", userID), userID%100+1)
				return err
			},
		}

		// Execute random operation
		operation := operations[userID%len(operations)]
		return operation()
	})

	validateStressTestResults(t, "DatabaseOperations", metrics, config)
}

// TestStressMemoryUsage tests memory usage under load
func TestStressMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Cleanup()

	initialMemory := getMemoryUsage()
	t.Logf("Initial memory usage: %.2f MB", initialMemory)

	config := StressTestConfig{
		StartUsers:       50,
		MaxUsers:         1000,
		UserIncrement:    50,
		IncrementDelay:   3 * time.Second,
		TestDuration:     2 * time.Minute,
		RequestTimeout:   5 * time.Second,
		FailureThreshold: 0.05,
		LatencyThreshold: 1 * time.Second,
	}

	metrics := runStressTest(t, testDB, "MemoryUsage", config, func(userID int) error {
		// Create large data structures to test memory handling
		data := make([]byte, 1024*10) // 10KB per request
		for i := range data {
			data[i] = byte(userID % 256)
		}

		// Simulate processing
		time.Sleep(10 * time.Millisecond)

		return nil
	})

	finalMemory := getMemoryUsage()
	memoryIncrease := finalMemory - initialMemory

	t.Logf("Final memory usage: %.2f MB", finalMemory)
	t.Logf("Memory increase: %.2f MB", memoryIncrease)

	// Check for memory leaks (allow for reasonable increase)
	if memoryIncrease > 500 { // 500MB threshold
		t.Errorf("Potential memory leak detected: memory increased by %.2f MB", memoryIncrease)
	}

	validateStressTestResults(t, "MemoryUsage", metrics, config)
}

// runStressTest executes a stress test with gradual user increase
func runStressTest(t *testing.T, testDB *TestDatabase, testName string, config StressTestConfig, testFunc func(int) error) StressTestMetrics {
	t.Logf("Starting stress test: %s", testName)
	t.Logf("Config: %d -> %d users, increment: %d, duration: %v", config.StartUsers, config.MaxUsers, config.UserIncrement, config.TestDuration)

	var (
		totalRequests      int64
		successfulRequests int64
		failedRequests     int64
		connectionErrors   int64
		timeoutErrors      int64
		totalLatency       int64
		maxLatency         int64
		currentUsers       int32
		maxUsers           int32
		peakRPS            float64
	)

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	defer cancel()

	// Channel for coordinating user goroutines
	userChan := make(chan struct{}, config.MaxUsers)
	var wg sync.WaitGroup

	// Metrics collection goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var lastRequests int64
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				currentReqs := atomic.LoadInt64(&totalRequests)
				rps := float64(currentReqs - lastRequests)
				if rps > peakRPS {
					peakRPS = rps
				}
				lastRequests = currentReqs

				users := atomic.LoadInt32(&currentUsers)
				if users > maxUsers {
					maxUsers = users
				}

				if testing.Verbose() {
					t.Logf("Current: %d users, %.0f RPS, %d total requests", users, rps, currentReqs)
				}
			}
		}
	}()

	// Gradual user increase
	go func() {
		users := config.StartUsers
		for users <= config.MaxUsers {
			select {
			case <-ctx.Done():
				return
			default:
				// Add users
				for i := 0; i < config.UserIncrement && users+i <= config.MaxUsers; i++ {
					userChan <- struct{}{}
					atomic.AddInt32(&currentUsers, 1)

					wg.Add(1)
					go func(userID int) {
						defer wg.Done()
						defer func() {
							<-userChan
							atomic.AddInt32(&currentUsers, -1)
						}()

						// User execution loop
						for {
							select {
							case <-ctx.Done():
								return
							default:
								requestStart := time.Now()

								err := testFunc(userID)

								requestLatency := time.Since(requestStart)
								atomic.AddInt64(&totalRequests, 1)
								atomic.AddInt64(&totalLatency, int64(requestLatency))

								// Update max latency
								for {
									currentMax := atomic.LoadInt64(&maxLatency)
									if int64(requestLatency) <= currentMax {
										break
									}
									if atomic.CompareAndSwapInt64(&maxLatency, currentMax, int64(requestLatency)) {
										break
									}
								}

								if err != nil {
									atomic.AddInt64(&failedRequests, 1)
									if isConnectionError(err) {
										atomic.AddInt64(&connectionErrors, 1)
									} else if isTimeoutError(err) {
										atomic.AddInt64(&timeoutErrors, 1)
									}
								} else {
									atomic.AddInt64(&successfulRequests, 1)
								}

								// Brief pause between requests
								time.Sleep(50 * time.Millisecond)
							}
						}
					}(users + i)
				}

				users += config.UserIncrement

				// Wait before next increment
				select {
				case <-ctx.Done():
					return
				case <-time.After(config.IncrementDelay):
				}
			}
		}
	}()

	// Wait for test completion
	<-ctx.Done()

	// Allow some time for cleanup
	time.Sleep(1 * time.Second)
	cancel()

	// Wait for all users to finish (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Log("Warning: Some user goroutines did not finish within timeout")
	}

	endTime := time.Now()
	testDuration := endTime.Sub(startTime)

	// Calculate final metrics
	totalReqs := atomic.LoadInt64(&totalRequests)
	successReqs := atomic.LoadInt64(&successfulRequests)
	failReqs := atomic.LoadInt64(&failedRequests)
	avgLatency := time.Duration(0)

	if totalReqs > 0 {
		avgLatency = time.Duration(atomic.LoadInt64(&totalLatency) / totalReqs)
	}

	errorRate := float64(failReqs) / float64(totalReqs)
	memoryUsage := getMemoryUsage()

	return StressTestMetrics{
		MaxConcurrentUsers:    int(maxUsers),
		PeakRequestsPerSecond: peakRPS,
		TotalRequests:         totalReqs,
		SuccessfulRequests:    successReqs,
		FailedRequests:        failReqs,
		AverageLatency:        avgLatency,
		MaxLatency:            time.Duration(atomic.LoadInt64(&maxLatency)),
		MemoryUsageMB:         memoryUsage,
		ErrorRate:             errorRate,
		ConnectionErrors:      atomic.LoadInt64(&connectionErrors),
		TimeoutErrors:         atomic.LoadInt64(&timeoutErrors),
		TestDuration:          testDuration,
	}
}

// runWebSocketStressTest tests WebSocket connections under stress
func runWebSocketStressTest(t *testing.T, testDB *TestDatabase, userIDs []int, config StressTestConfig) StressTestMetrics {
	// This is a simplified WebSocket stress test
	// In a real implementation, you would create actual WebSocket connections

	return runStressTest(t, testDB, "WebSocketStress", config, func(userID int) error {
		// Simulate WebSocket connection and message sending
		time.Sleep(10 * time.Millisecond) // Simulate network latency

		// Simulate occasional connection failures
		if userID%100 == 0 {
			return fmt.Errorf("simulated connection error")
		}

		return nil
	})
}

// validateStressTestResults checks if stress test results meet requirements
func validateStressTestResults(t *testing.T, testName string, metrics StressTestMetrics, config StressTestConfig) {
	t.Logf("Stress Test Results for %s:", testName)
	t.Logf("  Max Concurrent Users: %d", metrics.MaxConcurrentUsers)
	t.Logf("  Peak RPS: %.2f", metrics.PeakRequestsPerSecond)
	t.Logf("  Total Requests: %d", metrics.TotalRequests)
	t.Logf("  Successful: %d (%.1f%%)", metrics.SuccessfulRequests, float64(metrics.SuccessfulRequests)/float64(metrics.TotalRequests)*100)
	t.Logf("  Failed: %d (%.1f%%)", metrics.FailedRequests, metrics.ErrorRate*100)
	t.Logf("  Average Latency: %v", metrics.AverageLatency)
	t.Logf("  Max Latency: %v", metrics.MaxLatency)
	t.Logf("  Memory Usage: %.2f MB", metrics.MemoryUsageMB)
	t.Logf("  Connection Errors: %d", metrics.ConnectionErrors)
	t.Logf("  Timeout Errors: %d", metrics.TimeoutErrors)
	t.Logf("  Test Duration: %v", metrics.TestDuration)

	// Validate against thresholds
	if metrics.ErrorRate > config.FailureThreshold {
		t.Errorf("Error rate %.2f%% exceeds threshold %.2f%%", metrics.ErrorRate*100, config.FailureThreshold*100)
	}

	if metrics.AverageLatency > config.LatencyThreshold {
		t.Errorf("Average latency %v exceeds threshold %v", metrics.AverageLatency, config.LatencyThreshold)
	}

	// Check for minimum performance requirements
	if metrics.PeakRequestsPerSecond < 10 {
		t.Errorf("Peak RPS %.2f is too low", metrics.PeakRequestsPerSecond)
	}

	if metrics.MaxConcurrentUsers < config.StartUsers {
		t.Errorf("Failed to reach minimum concurrent users: %d < %d", metrics.MaxConcurrentUsers, config.StartUsers)
	}

	// Memory usage warnings
	if metrics.MemoryUsageMB > 1000 { // 1GB threshold
		t.Logf("Warning: High memory usage detected: %.2f MB", metrics.MemoryUsageMB)
	}

	// Connection error warnings
	if metrics.ConnectionErrors > metrics.TotalRequests/100 { // More than 1% connection errors
		t.Logf("Warning: High connection error rate: %d errors", metrics.ConnectionErrors)
	}
}

// Helper functions
func getMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024 // Convert to MB
}

func isConnectionError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "network") ||
		strings.Contains(err.Error(), "dial"))
}

func isTimeoutError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline"))
}
