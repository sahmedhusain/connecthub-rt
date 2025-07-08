package unit_testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"
)

// SimplePerformanceMetrics holds basic performance metrics
type SimplePerformanceMetrics struct {
	TotalRequests     int           `json:"total_requests"`
	SuccessfulReqs    int           `json:"successful_requests"`
	FailedReqs        int           `json:"failed_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	MaxLatency        time.Duration `json:"max_latency"`
	MinLatency        time.Duration `json:"min_latency"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	MemoryUsageMB     float64       `json:"memory_usage_mb"`
	TestDuration      time.Duration `json:"test_duration"`
}

// BenchmarkSimpleHTTPRequests demonstrates basic HTTP request benchmarking
func BenchmarkSimpleHTTPRequests(b *testing.B) {
	// Create a simple test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(1 * time.Millisecond)

		switch r.URL.Path {
		case "/api/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		case "/api/data":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data":      "test data",
				"timestamp": time.Now().Unix(),
				"id":        123,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{Timeout: 5 * time.Second}
		for pb.Next() {
			resp, err := client.Get(server.URL + "/api/health")
			if err != nil {
				b.Errorf("Request failed: %v", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}

// BenchmarkJSONProcessing demonstrates JSON processing performance
func BenchmarkJSONProcessing(b *testing.B) {
	testData := map[string]interface{}{
		"id":       12345,
		"username": "testuser",
		"email":    "test@example.com",
		"posts":    []string{"post1", "post2", "post3"},
		"metadata": map[string]string{
			"created": "2024-01-01",
			"updated": "2024-01-02",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Marshal to JSON
		jsonData, err := json.Marshal(testData)
		if err != nil {
			b.Errorf("JSON marshal failed: %v", err)
			continue
		}

		// Unmarshal from JSON
		var result map[string]interface{}
		err = json.Unmarshal(jsonData, &result)
		if err != nil {
			b.Errorf("JSON unmarshal failed: %v", err)
			continue
		}

		// Verify data integrity
		if result["id"].(float64) != 12345 {
			b.Errorf("Data integrity check failed")
		}
	}
}

// BenchmarkMemoryAllocation demonstrates memory allocation performance
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Allocate various data structures
		slice := make([]int, 1000)
		for j := 0; j < 1000; j++ {
			slice[j] = j
		}

		// Create a map
		m := make(map[string]int)
		for j := 0; j < 100; j++ {
			m[fmt.Sprintf("key%d", j)] = j
		}

		// Create a struct
		type TestStruct struct {
			ID   int
			Name string
			Data []byte
		}

		s := TestStruct{
			ID:   i,
			Name: fmt.Sprintf("test%d", i),
			Data: make([]byte, 256),
		}

		// Use the data to prevent optimization
		_ = slice[999] + len(m) + s.ID
	}
}

// TestSimpleLoadTest demonstrates a simple load test
func TestSimpleLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate processing time
		time.Sleep(10 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Hello from load test",
			"time":    time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Test configuration
	concurrentUsers := 10
	requestsPerUser := 20
	testDuration := 5 * time.Second

	metrics := runSimpleLoadTest(t, server.URL, concurrentUsers, requestsPerUser, testDuration)

	// Log results
	t.Logf("Load Test Results:")
	t.Logf("  Total Requests: %d", metrics.TotalRequests)
	t.Logf("  Successful: %d", metrics.SuccessfulReqs)
	t.Logf("  Failed: %d", metrics.FailedReqs)
	t.Logf("  Average Latency: %v", metrics.AverageLatency)
	t.Logf("  Max Latency: %v", metrics.MaxLatency)
	t.Logf("  Min Latency: %v", metrics.MinLatency)
	t.Logf("  Requests/Second: %.2f", metrics.RequestsPerSecond)
	t.Logf("  Memory Usage: %.2f MB", metrics.MemoryUsageMB)
	t.Logf("  Test Duration: %v", metrics.TestDuration)

	// Validate results
	if metrics.SuccessfulReqs == 0 {
		t.Error("No successful requests")
	}

	successRate := float64(metrics.SuccessfulReqs) / float64(metrics.TotalRequests)
	if successRate < 0.95 { // 95% success rate
		t.Errorf("Low success rate: %.2f%%", successRate*100)
	}

	if metrics.AverageLatency > 1*time.Second {
		t.Errorf("High average latency: %v", metrics.AverageLatency)
	}
}

// runSimpleLoadTest executes a simple load test
func runSimpleLoadTest(t *testing.T, serverURL string, concurrentUsers, requestsPerUser int, duration time.Duration) SimplePerformanceMetrics {
	var (
		totalRequests  int
		successfulReqs int
		failedReqs     int
		totalLatency   time.Duration
		minLatency     time.Duration = time.Hour // Initialize to high value
		maxLatency     time.Duration
		mutex          sync.Mutex
	)

	startTime := time.Now()
	var wg sync.WaitGroup

	// Record initial memory usage
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// Start concurrent users
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			client := &http.Client{Timeout: 5 * time.Second}
			userStartTime := time.Now()

			for j := 0; j < requestsPerUser; j++ {
				// Check if duration exceeded
				if time.Since(startTime) > duration {
					break
				}

				requestStart := time.Now()
				resp, err := client.Get(serverURL + "/api/test")
				requestLatency := time.Since(requestStart)

				mutex.Lock()
				totalRequests++
				totalLatency += requestLatency

				if requestLatency < minLatency {
					minLatency = requestLatency
				}
				if requestLatency > maxLatency {
					maxLatency = requestLatency
				}

				if err != nil {
					failedReqs++
				} else {
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						successfulReqs++
					} else {
						failedReqs++
					}
				}
				mutex.Unlock()

				// Small delay between requests
				time.Sleep(50 * time.Millisecond)
			}

			t.Logf("User %d completed in %v", userID, time.Since(userStartTime))
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()
	endTime := time.Now()
	testDuration := endTime.Sub(startTime)

	// Record final memory usage
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)
	memoryUsageMB := float64(finalMem.Alloc-initialMem.Alloc) / 1024 / 1024

	// Calculate metrics
	avgLatency := time.Duration(0)
	if totalRequests > 0 {
		avgLatency = totalLatency / time.Duration(totalRequests)
	}

	requestsPerSecond := float64(totalRequests) / testDuration.Seconds()

	return SimplePerformanceMetrics{
		TotalRequests:     totalRequests,
		SuccessfulReqs:    successfulReqs,
		FailedReqs:        failedReqs,
		AverageLatency:    avgLatency,
		MaxLatency:        maxLatency,
		MinLatency:        minLatency,
		RequestsPerSecond: requestsPerSecond,
		MemoryUsageMB:     memoryUsageMB,
		TestDuration:      testDuration,
	}
}

// TestMemoryUsage demonstrates memory usage testing
func TestMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)
	t.Logf("Initial memory usage: %.2f MB", float64(initialMem.Alloc)/1024/1024)

	// Allocate memory in chunks
	var data [][]byte
	chunkSize := 1024 * 1024 // 1MB chunks
	numChunks := 10

	for i := 0; i < numChunks; i++ {
		chunk := make([]byte, chunkSize)
		// Fill with data to prevent optimization
		for j := range chunk {
			chunk[j] = byte(i % 256)
		}
		data = append(data, chunk)

		var currentMem runtime.MemStats
		runtime.ReadMemStats(&currentMem)
		t.Logf("After chunk %d: %.2f MB", i+1, float64(currentMem.Alloc)/1024/1024)
	}

	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)
	memoryIncrease := float64(finalMem.Alloc-initialMem.Alloc) / 1024 / 1024

	t.Logf("Final memory usage: %.2f MB", float64(finalMem.Alloc)/1024/1024)
	t.Logf("Memory increase: %.2f MB", memoryIncrease)

	// Validate memory usage is reasonable
	expectedIncrease := float64(numChunks*chunkSize) / 1024 / 1024
	if memoryIncrease > expectedIncrease*2 { // Allow for 100% overhead
		t.Errorf("Memory usage too high: %.2f MB (expected ~%.2f MB)", memoryIncrease, expectedIncrease)
	}

	// Clean up
	data = nil
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup

	var cleanupMem runtime.MemStats
	runtime.ReadMemStats(&cleanupMem)
	t.Logf("After cleanup: %.2f MB", float64(cleanupMem.Alloc)/1024/1024)
}

// TestConcurrentOperations demonstrates concurrent operation testing
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations test in short mode")
	}

	// Shared data structure
	data := make(map[string]int)
	var mutex sync.RWMutex
	var wg sync.WaitGroup

	numWorkers := 20
	operationsPerWorker := 100

	startTime := time.Now()

	// Start concurrent workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				key := fmt.Sprintf("worker_%d_key_%d", workerID, j)

				// Write operation
				mutex.Lock()
				data[key] = workerID*1000 + j
				mutex.Unlock()

				// Read operation
				mutex.RLock()
				value, exists := data[key]
				mutex.RUnlock()

				if !exists {
					t.Errorf("Key %s not found", key)
				}
				if value != workerID*1000+j {
					t.Errorf("Value mismatch for key %s: expected %d, got %d", key, workerID*1000+j, value)
				}

				// Small delay to simulate processing
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalOperations := numWorkers * operationsPerWorker * 2 // Read + Write
	operationsPerSecond := float64(totalOperations) / duration.Seconds()

	t.Logf("Concurrent Operations Results:")
	t.Logf("  Workers: %d", numWorkers)
	t.Logf("  Operations per worker: %d", operationsPerWorker)
	t.Logf("  Total operations: %d", totalOperations)
	t.Logf("  Duration: %v", duration)
	t.Logf("  Operations/second: %.2f", operationsPerSecond)
	t.Logf("  Final data size: %d entries", len(data))

	// Validate results
	expectedEntries := numWorkers * operationsPerWorker
	if len(data) != expectedEntries {
		t.Errorf("Expected %d entries, got %d", expectedEntries, len(data))
	}

	if operationsPerSecond < 1000 { // Minimum threshold
		t.Errorf("Operations per second too low: %.2f", operationsPerSecond)
	}
}
