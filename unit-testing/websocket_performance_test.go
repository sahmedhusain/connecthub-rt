package unit_testing

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMetrics holds WebSocket performance metrics
type WebSocketMetrics struct {
	TotalConnections      int64         `json:"total_connections"`
	SuccessfulConnections int64         `json:"successful_connections"`
	FailedConnections     int64         `json:"failed_connections"`
	TotalMessages         int64         `json:"total_messages"`
	MessagesSent          int64         `json:"messages_sent"`
	MessagesReceived      int64         `json:"messages_received"`
	AverageLatency        time.Duration `json:"average_latency"`
	MinLatency            time.Duration `json:"min_latency"`
	MaxLatency            time.Duration `json:"max_latency"`
	P95Latency            time.Duration `json:"p95_latency"`
	ConnectionTime        time.Duration `json:"average_connection_time"`
	MessagesPerSecond     float64       `json:"messages_per_second"`
	ConnectionErrors      int64         `json:"connection_errors"`
	MessageErrors         int64         `json:"message_errors"`
	TestDuration          time.Duration `json:"test_duration"`
}

// MockWebSocketManager provides a mock WebSocket manager for testing
type MockWebSocketManager struct {
	upgrader websocket.Upgrader
}

// NewMockWebSocketManager creates a new mock WebSocket manager
func NewMockWebSocketManager() *MockWebSocketManager {
	return &MockWebSocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for testing
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections for testing
func (m *MockWebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Simple echo server for testing
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Echo the message back
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			break
		}
	}
}

// WebSocketTestConfig defines WebSocket test parameters
type WebSocketTestConfig struct {
	ConcurrentConnections int           `json:"concurrent_connections"`
	MessagesPerConnection int           `json:"messages_per_connection"`
	MessageInterval       time.Duration `json:"message_interval"`
	TestDuration          time.Duration `json:"test_duration"`
	ConnectionTimeout     time.Duration `json:"connection_timeout"`
	MessageTimeout        time.Duration `json:"message_timeout"`
}

// TestWebSocketPerformance tests WebSocket performance under load
func TestWebSocketPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket performance test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Close()

	// Setup test users
	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		t.Fatalf("Failed to setup test users: %v", err)
	}

	// Create WebSocket manager
	wsManager := NewMockWebSocketManager()

	// Create test server with WebSocket support
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			wsManager.HandleWebSocket(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := WebSocketTestConfig{
		ConcurrentConnections: 50,
		MessagesPerConnection: 100,
		MessageInterval:       100 * time.Millisecond,
		TestDuration:          30 * time.Second,
		ConnectionTimeout:     5 * time.Second,
		MessageTimeout:        2 * time.Second,
	}

	metrics := runWebSocketPerformanceTest(t, server.URL, userIDs, config)
	validateWebSocketMetrics(t, metrics, config)
}

// TestWebSocketConcurrentConnections tests maximum concurrent connections
func TestWebSocketConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket concurrent connections test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Close()

	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		t.Fatalf("Failed to setup test users: %v", err)
	}

	wsManager := NewMockWebSocketManager()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			wsManager.HandleWebSocket(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test increasing connection counts
	connectionCounts := []int{10, 25, 50, 100, 200}

	for _, count := range connectionCounts {
		t.Run(fmt.Sprintf("Connections_%d", count), func(t *testing.T) {
			config := WebSocketTestConfig{
				ConcurrentConnections: count,
				MessagesPerConnection: 10,
				MessageInterval:       500 * time.Millisecond,
				TestDuration:          15 * time.Second,
				ConnectionTimeout:     10 * time.Second,
				MessageTimeout:        5 * time.Second,
			}

			metrics := runWebSocketPerformanceTest(t, server.URL, userIDs, config)

			t.Logf("Connections: %d, Success Rate: %.2f%%, Avg Latency: %v",
				count,
				float64(metrics.SuccessfulConnections)/float64(metrics.TotalConnections)*100,
				metrics.AverageLatency)

			// Validate that most connections succeed
			successRate := float64(metrics.SuccessfulConnections) / float64(metrics.TotalConnections)
			if successRate < 0.9 { // 90% success rate threshold
				t.Errorf("Low success rate for %d connections: %.2f%%", count, successRate*100)
			}
		})
	}
}

// TestWebSocketMessageThroughput tests message throughput
func TestWebSocketMessageThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket message throughput test in short mode")
	}

	testDB := TestSetup(t)
	defer testDB.Close()

	userIDs, err := SetupTestUsers(testDB.DB)
	if err != nil {
		t.Fatalf("Failed to setup test users: %v", err)
	}

	wsManager := NewMockWebSocketManager()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			wsManager.HandleWebSocket(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test different message rates
	messageIntervals := []time.Duration{
		10 * time.Millisecond,  // 100 msg/sec per connection
		50 * time.Millisecond,  // 20 msg/sec per connection
		100 * time.Millisecond, // 10 msg/sec per connection
		500 * time.Millisecond, // 2 msg/sec per connection
	}

	for _, interval := range messageIntervals {
		t.Run(fmt.Sprintf("Interval_%v", interval), func(t *testing.T) {
			config := WebSocketTestConfig{
				ConcurrentConnections: 20,
				MessagesPerConnection: 50,
				MessageInterval:       interval,
				TestDuration:          20 * time.Second,
				ConnectionTimeout:     5 * time.Second,
				MessageTimeout:        2 * time.Second,
			}

			metrics := runWebSocketPerformanceTest(t, server.URL, userIDs, config)

			expectedMsgPerSec := float64(config.ConcurrentConnections) / interval.Seconds()
			actualMsgPerSec := metrics.MessagesPerSecond

			t.Logf("Interval: %v, Expected: %.2f msg/sec, Actual: %.2f msg/sec",
				interval, expectedMsgPerSec, actualMsgPerSec)

			// Allow for some variance in message rate
			if actualMsgPerSec < expectedMsgPerSec*0.8 {
				t.Errorf("Message throughput too low: %.2f < %.2f", actualMsgPerSec, expectedMsgPerSec*0.8)
			}
		})
	}
}

// runWebSocketPerformanceTest executes a WebSocket performance test
func runWebSocketPerformanceTest(t *testing.T, serverURL string, userIDs []int, config WebSocketTestConfig) WebSocketMetrics {
	var (
		totalConnections      int64
		successfulConnections int64
		failedConnections     int64
		totalMessages         int64
		messagesSent          int64
		messagesReceived      int64
		connectionErrors      int64
		messageErrors         int64
		totalLatency          int64
		minLatency            int64 = int64(time.Hour) // Initialize to high value
		maxLatency            int64
		totalConnectionTime   int64
	)

	latencies := make([]time.Duration, 0)
	var latencyMutex sync.Mutex

	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	defer cancel()

	var wg sync.WaitGroup

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + serverURL[4:] + "/ws"

	// Start concurrent connections
	for i := 0; i < config.ConcurrentConnections; i++ {
		wg.Add(1)

		go func(connectionID int) {
			defer wg.Done()

			userID := userIDs[connectionID%len(userIDs)]

			// Attempt to connect
			atomic.AddInt64(&totalConnections, 1)

			connectionStart := time.Now()
			conn, err := connectWebSocket(wsURL, userID, config.ConnectionTimeout)
			connectionTime := time.Since(connectionStart)
			atomic.AddInt64(&totalConnectionTime, int64(connectionTime))

			if err != nil {
				atomic.AddInt64(&failedConnections, 1)
				atomic.AddInt64(&connectionErrors, 1)
				t.Logf("Connection %d failed: %v", connectionID, err)
				return
			}
			defer conn.Close()

			atomic.AddInt64(&successfulConnections, 1)

			// Send and receive messages
			messageCount := 0
			ticker := time.NewTicker(config.MessageInterval)
			defer ticker.Stop()

			// Start message receiver
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						conn.SetReadDeadline(time.Now().Add(config.MessageTimeout))
						_, _, err := conn.ReadMessage()
						if err != nil {
							if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
								atomic.AddInt64(&messageErrors, 1)
							}
							return
						}
						atomic.AddInt64(&messagesReceived, 1)
					}
				}
			}()

			// Send messages
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if messageCount >= config.MessagesPerConnection {
						return
					}

					messageStart := time.Now()

					message := map[string]interface{}{
						"type":            "message",
						"conversation_id": 1,
						"content":         fmt.Sprintf("Performance test message %d from connection %d", messageCount, connectionID),
						"timestamp":       messageStart.UnixNano(),
					}

					conn.SetWriteDeadline(time.Now().Add(config.MessageTimeout))
					err := conn.WriteJSON(message)

					if err != nil {
						atomic.AddInt64(&messageErrors, 1)
						continue
					}

					atomic.AddInt64(&messagesSent, 1)
					atomic.AddInt64(&totalMessages, 1)

					// Calculate latency (simplified - in real scenario you'd wait for response)
					latency := time.Since(messageStart)

					latencyMutex.Lock()
					latencies = append(latencies, latency)
					latencyMutex.Unlock()

					atomic.AddInt64(&totalLatency, int64(latency))

					// Update min/max latency
					for {
						currentMin := atomic.LoadInt64(&minLatency)
						if int64(latency) >= currentMin {
							break
						}
						if atomic.CompareAndSwapInt64(&minLatency, currentMin, int64(latency)) {
							break
						}
					}

					for {
						currentMax := atomic.LoadInt64(&maxLatency)
						if int64(latency) <= currentMax {
							break
						}
						if atomic.CompareAndSwapInt64(&maxLatency, currentMax, int64(latency)) {
							break
						}
					}

					messageCount++
				}
			}
		}(i)
	}

	// Wait for all connections to complete
	wg.Wait()
	endTime := time.Now()
	testDuration := endTime.Sub(startTime)

	// Calculate metrics
	totalConns := atomic.LoadInt64(&totalConnections)
	successConns := atomic.LoadInt64(&successfulConnections)
	totalMsgs := atomic.LoadInt64(&totalMessages)
	sentMsgs := atomic.LoadInt64(&messagesSent)
	receivedMsgs := atomic.LoadInt64(&messagesReceived)

	avgLatency := time.Duration(0)
	if totalMsgs > 0 {
		avgLatency = time.Duration(atomic.LoadInt64(&totalLatency) / totalMsgs)
	}

	avgConnectionTime := time.Duration(0)
	if totalConns > 0 {
		avgConnectionTime = time.Duration(atomic.LoadInt64(&totalConnectionTime) / totalConns)
	}

	messagesPerSecond := float64(totalMsgs) / testDuration.Seconds()

	// Calculate P95 latency
	p95Latency := calculateP95Latency(latencies)

	return WebSocketMetrics{
		TotalConnections:      totalConns,
		SuccessfulConnections: successConns,
		FailedConnections:     atomic.LoadInt64(&failedConnections),
		TotalMessages:         totalMsgs,
		MessagesSent:          sentMsgs,
		MessagesReceived:      receivedMsgs,
		AverageLatency:        avgLatency,
		MinLatency:            time.Duration(atomic.LoadInt64(&minLatency)),
		MaxLatency:            time.Duration(atomic.LoadInt64(&maxLatency)),
		P95Latency:            p95Latency,
		ConnectionTime:        avgConnectionTime,
		MessagesPerSecond:     messagesPerSecond,
		ConnectionErrors:      atomic.LoadInt64(&connectionErrors),
		MessageErrors:         atomic.LoadInt64(&messageErrors),
		TestDuration:          testDuration,
	}
}

// connectWebSocket establishes a WebSocket connection
func connectWebSocket(wsURL string, userID int, timeout time.Duration) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: timeout,
	}

	// Add authentication headers or cookies if needed
	headers := http.Header{}
	headers.Add("User-Agent", fmt.Sprintf("LoadTest-User-%d", userID))

	u, err := url.Parse(wsURL)
	if err != nil {
		return nil, err
	}

	conn, _, err := dialer.Dial(u.String(), headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// calculateP95Latency calculates the 95th percentile latency
func calculateP95Latency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Simple sorting for P95 calculation
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Bubble sort (sufficient for test data)
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}

	return sorted[p95Index]
}

// validateWebSocketMetrics validates WebSocket performance metrics
func validateWebSocketMetrics(t *testing.T, metrics WebSocketMetrics, config WebSocketTestConfig) {
	t.Logf("WebSocket Performance Results:")
	t.Logf("  Total Connections: %d", metrics.TotalConnections)
	t.Logf("  Successful: %d (%.1f%%)", metrics.SuccessfulConnections,
		float64(metrics.SuccessfulConnections)/float64(metrics.TotalConnections)*100)
	t.Logf("  Failed: %d", metrics.FailedConnections)
	t.Logf("  Total Messages: %d", metrics.TotalMessages)
	t.Logf("  Messages Sent: %d", metrics.MessagesSent)
	t.Logf("  Messages Received: %d", metrics.MessagesReceived)
	t.Logf("  Average Latency: %v", metrics.AverageLatency)
	t.Logf("  P95 Latency: %v", metrics.P95Latency)
	t.Logf("  Connection Time: %v", metrics.ConnectionTime)
	t.Logf("  Messages/Second: %.2f", metrics.MessagesPerSecond)
	t.Logf("  Connection Errors: %d", metrics.ConnectionErrors)
	t.Logf("  Message Errors: %d", metrics.MessageErrors)
	t.Logf("  Test Duration: %v", metrics.TestDuration)

	// Validate performance requirements
	connectionSuccessRate := float64(metrics.SuccessfulConnections) / float64(metrics.TotalConnections)
	if connectionSuccessRate < 0.95 { // 95% success rate
		t.Errorf("Connection success rate too low: %.2f%%", connectionSuccessRate*100)
	}

	if metrics.AverageLatency > 500*time.Millisecond {
		t.Errorf("Average latency too high: %v", metrics.AverageLatency)
	}

	if metrics.P95Latency > 2*time.Second {
		t.Errorf("P95 latency too high: %v", metrics.P95Latency)
	}

	if metrics.ConnectionTime > 5*time.Second {
		t.Errorf("Connection time too high: %v", metrics.ConnectionTime)
	}

	// Check for reasonable message throughput
	expectedMinThroughput := float64(config.ConcurrentConnections) * 0.5 // At least 0.5 msg/sec per connection
	if metrics.MessagesPerSecond < expectedMinThroughput {
		t.Errorf("Message throughput too low: %.2f < %.2f", metrics.MessagesPerSecond, expectedMinThroughput)
	}

	// Check error rates
	if metrics.ConnectionErrors > metrics.TotalConnections/10 { // Max 10% connection errors
		t.Errorf("Too many connection errors: %d", metrics.ConnectionErrors)
	}

	if metrics.MessageErrors > metrics.TotalMessages/20 { // Max 5% message errors
		t.Errorf("Too many message errors: %d", metrics.MessageErrors)
	}
}
