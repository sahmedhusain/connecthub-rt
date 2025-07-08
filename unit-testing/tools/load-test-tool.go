package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	BaseURL         string        `json:"base_url"`
	ConcurrentUsers int           `json:"concurrent_users"`
	Duration        time.Duration `json:"duration"`
	RequestsPerUser int           `json:"requests_per_user"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	ThinkTime       time.Duration `json:"think_time"`
	TestScenarios   []string      `json:"test_scenarios"`
	OutputFormat    string        `json:"output_format"`
	ReportFile      string        `json:"report_file"`
}

// TestResult holds individual request results
type TestResult struct {
	Scenario     string        `json:"scenario"`
	Method       string        `json:"method"`
	URL          string        `json:"url"`
	StatusCode   int           `json:"status_code"`
	Latency      time.Duration `json:"latency"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	ResponseSize int           `json:"response_size"`
}

// LoadTestResults aggregates all test results
type LoadTestResults struct {
	Config          LoadTestConfig     `json:"config"`
	StartTime       time.Time          `json:"start_time"`
	EndTime         time.Time          `json:"end_time"`
	TotalDuration   time.Duration      `json:"total_duration"`
	TotalRequests   int                `json:"total_requests"`
	SuccessfulReqs  int                `json:"successful_requests"`
	FailedReqs      int                `json:"failed_requests"`
	AverageLatency  time.Duration      `json:"average_latency"`
	MinLatency      time.Duration      `json:"min_latency"`
	MaxLatency      time.Duration      `json:"max_latency"`
	P50Latency      time.Duration      `json:"p50_latency"`
	P95Latency      time.Duration      `json:"p95_latency"`
	P99Latency      time.Duration      `json:"p99_latency"`
	RequestsPerSec  float64            `json:"requests_per_second"`
	ErrorRate       float64            `json:"error_rate"`
	ScenarioResults map[string]Metrics `json:"scenario_results"`
	DetailedResults []TestResult       `json:"detailed_results,omitempty"`
}

// Metrics holds performance metrics for a scenario
type Metrics struct {
	TotalRequests  int           `json:"total_requests"`
	SuccessfulReqs int           `json:"successful_requests"`
	FailedReqs     int           `json:"failed_requests"`
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	RequestsPerSec float64       `json:"requests_per_second"`
	ErrorRate      float64       `json:"error_rate"`
}

// TestScenario defines a test scenario
type TestScenario struct {
	Name        string
	Weight      int
	ExecuteFunc func(*http.Client, string) TestResult
}

var (
	baseURL         = flag.String("url", "http://localhost:8080", "Base URL for load testing")
	concurrentUsers = flag.Int("users", 10, "Number of concurrent users")
	duration        = flag.Duration("duration", 30*time.Second, "Test duration")
	requestsPerUser = flag.Int("requests", 100, "Requests per user")
	rampUpTime      = flag.Duration("rampup", 5*time.Second, "Ramp up time")
	thinkTime       = flag.Duration("think", 100*time.Millisecond, "Think time between requests")
	scenarios       = flag.String("scenarios", "all", "Test scenarios (comma-separated or 'all')")
	outputFormat    = flag.String("format", "json", "Output format (json, csv, html)")
	reportFile      = flag.String("output", "", "Output file (default: stdout)")
	verbose         = flag.Bool("verbose", false, "Verbose output")
	configFile      = flag.String("config", "", "Load test configuration file")
)

func main() {
	flag.Parse()

	var config LoadTestConfig

	// Load configuration from file if provided
	if *configFile != "" {
		if err := loadConfig(*configFile, &config); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else {
		// Use command line flags
		config = LoadTestConfig{
			BaseURL:         *baseURL,
			ConcurrentUsers: *concurrentUsers,
			Duration:        *duration,
			RequestsPerUser: *requestsPerUser,
			RampUpTime:      *rampUpTime,
			ThinkTime:       *thinkTime,
			TestScenarios:   strings.Split(*scenarios, ","),
			OutputFormat:    *outputFormat,
			ReportFile:      *reportFile,
		}
	}

	// Run load test
	results := runLoadTest(config)

	// Output results
	if err := outputResults(results, config.OutputFormat, config.ReportFile); err != nil {
		log.Fatalf("Failed to output results: %v", err)
	}

	// Print summary to console
	printSummary(results)
}

func loadConfig(filename string, config *LoadTestConfig) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

func runLoadTest(config LoadTestConfig) LoadTestResults {
	fmt.Printf("Starting load test with %d concurrent users for %v\n", config.ConcurrentUsers, config.Duration)
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Scenarios: %v\n", config.TestScenarios)

	startTime := time.Now()

	// Initialize test scenarios
	testScenarios := getTestScenarios(config.TestScenarios)

	// Channel to collect results
	resultsChan := make(chan TestResult, config.ConcurrentUsers*config.RequestsPerUser)

	var wg sync.WaitGroup

	// Start users with ramp-up
	userDelay := config.RampUpTime / time.Duration(config.ConcurrentUsers)

	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)

		go func(userID int) {
			defer wg.Done()

			// Ramp-up delay
			time.Sleep(time.Duration(userID) * userDelay)

			// Create HTTP client for this user
			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			// Execute requests
			userStartTime := time.Now()
			requestCount := 0

			for {
				// Check if duration exceeded
				if time.Since(startTime) > config.Duration {
					break
				}

				// Check if max requests per user reached
				if requestCount >= config.RequestsPerUser {
					break
				}

				// Select random scenario
				scenario := selectRandomScenario(testScenarios)

				// Execute scenario
				result := scenario.ExecuteFunc(client, config.BaseURL)
				result.Scenario = scenario.Name
				result.Timestamp = time.Now()

				resultsChan <- result
				requestCount++

				// Think time
				if requestCount < config.RequestsPerUser {
					time.Sleep(config.ThinkTime)
				}
			}

			if *verbose {
				fmt.Printf("User %d completed %d requests in %v\n", userID, requestCount, time.Since(userStartTime))
			}
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()
	close(resultsChan)

	endTime := time.Now()

	// Collect and analyze results
	var allResults []TestResult
	for result := range resultsChan {
		allResults = append(allResults, result)
	}

	return analyzeResults(config, startTime, endTime, allResults)
}

func getTestScenarios(scenarioNames []string) []TestScenario {
	allScenarios := map[string]TestScenario{
		"homepage":      {Name: "Homepage", Weight: 20, ExecuteFunc: testHomepage},
		"login":         {Name: "Login", Weight: 15, ExecuteFunc: testLogin},
		"signup":        {Name: "Signup", Weight: 10, ExecuteFunc: testSignup},
		"posts":         {Name: "View Posts", Weight: 25, ExecuteFunc: testViewPosts},
		"create_post":   {Name: "Create Post", Weight: 10, ExecuteFunc: testCreatePost},
		"view_post":     {Name: "View Single Post", Weight: 15, ExecuteFunc: testViewSinglePost},
		"add_comment":   {Name: "Add Comment", Weight: 5, ExecuteFunc: testAddComment},
		"messaging":     {Name: "Messaging", Weight: 10, ExecuteFunc: testMessaging},
		"conversations": {Name: "View Conversations", Weight: 5, ExecuteFunc: testViewConversations},
	}

	var scenarios []TestScenario

	if len(scenarioNames) == 1 && scenarioNames[0] == "all" {
		for _, scenario := range allScenarios {
			scenarios = append(scenarios, scenario)
		}
	} else {
		for _, name := range scenarioNames {
			if scenario, exists := allScenarios[name]; exists {
				scenarios = append(scenarios, scenario)
			}
		}
	}

	return scenarios
}

func selectRandomScenario(scenarios []TestScenario) TestScenario {
	totalWeight := 0
	for _, scenario := range scenarios {
		totalWeight += scenario.Weight
	}

	randomWeight := rand.Intn(totalWeight)
	currentWeight := 0

	for _, scenario := range scenarios {
		currentWeight += scenario.Weight
		if randomWeight < currentWeight {
			return scenario
		}
	}

	return scenarios[0] // Fallback
}

// Test scenario implementations
func testHomepage(client *http.Client, baseURL string) TestResult {
	return executeRequest(client, "GET", baseURL+"/", nil, nil)
}

func testLogin(client *http.Client, baseURL string) TestResult {
	loginData := map[string]string{
		"identifier": "johndoe",
		"password":   "Aa123456",
	}

	jsonData, _ := json.Marshal(loginData)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return executeRequest(client, "POST", baseURL+"/api/login", bytes.NewBuffer(jsonData), headers)
}

func testSignup(client *http.Client, baseURL string) TestResult {
	timestamp := time.Now().UnixNano()
	signupData := map[string]string{
		"firstName":   "Load",
		"lastName":    "Test",
		"username":    fmt.Sprintf("loadtest%d", timestamp),
		"email":       fmt.Sprintf("loadtest%d@example.com", timestamp),
		"gender":      "other",
		"dateOfBirth": "1990-01-01",
		"password":    "password123",
	}

	jsonData, _ := json.Marshal(signupData)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return executeRequest(client, "POST", baseURL+"/api/signup", bytes.NewBuffer(jsonData), headers)
}

func testViewPosts(client *http.Client, baseURL string) TestResult {
	return executeRequest(client, "GET", baseURL+"/api/posts", nil, nil)
}

func testCreatePost(client *http.Client, baseURL string) TestResult {
	// First login to get session
	loginResult := testLogin(client, baseURL)
	if !loginResult.Success {
		return TestResult{
			Method:     "POST",
			URL:        baseURL + "/api/post/create",
			StatusCode: 401,
			Success:    false,
			Error:      "Login failed",
			Latency:    0,
		}
	}

	timestamp := time.Now().UnixNano()
	postData := map[string]interface{}{
		"title":      fmt.Sprintf("Load Test Post %d", timestamp),
		"content":    fmt.Sprintf("This is a load test post created at %d", timestamp),
		"categories": []string{"Technology"},
	}

	jsonData, _ := json.Marshal(postData)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return executeRequest(client, "POST", baseURL+"/api/post/create", bytes.NewBuffer(jsonData), headers)
}

func testViewSinglePost(client *http.Client, baseURL string) TestResult {
	// Random post ID between 1 and 100
	postID := rand.Intn(100) + 1
	return executeRequest(client, "GET", baseURL+"/api/post?id="+strconv.Itoa(postID), nil, nil)
}

func testAddComment(client *http.Client, baseURL string) TestResult {
	// First login
	loginResult := testLogin(client, baseURL)
	if !loginResult.Success {
		return TestResult{
			Method:     "POST",
			URL:        baseURL + "/addcomment",
			StatusCode: 401,
			Success:    false,
			Error:      "Login failed",
			Latency:    0,
		}
	}

	timestamp := time.Now().UnixNano()
	postID := rand.Intn(100) + 1

	formData := url.Values{}
	formData.Set("post_id", strconv.Itoa(postID))
	formData.Set("content", fmt.Sprintf("Load test comment %d", timestamp))

	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	return executeRequest(client, "POST", baseURL+"/addcomment", bytes.NewBufferString(formData.Encode()), headers)
}

func testMessaging(client *http.Client, baseURL string) TestResult {
	// First login
	loginResult := testLogin(client, baseURL)
	if !loginResult.Success {
		return TestResult{
			Method:     "POST",
			URL:        baseURL + "/api/send-message",
			StatusCode: 401,
			Success:    false,
			Error:      "Login failed",
			Latency:    0,
		}
	}

	timestamp := time.Now().UnixNano()
	messageData := map[string]interface{}{
		"conversation_id": 1, // Assume conversation exists
		"content":         fmt.Sprintf("Load test message %d", timestamp),
	}

	jsonData, _ := json.Marshal(messageData)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	return executeRequest(client, "POST", baseURL+"/api/send-message", bytes.NewBuffer(jsonData), headers)
}

func testViewConversations(client *http.Client, baseURL string) TestResult {
	// First login
	loginResult := testLogin(client, baseURL)
	if !loginResult.Success {
		return TestResult{
			Method:     "GET",
			URL:        baseURL + "/api/conversations",
			StatusCode: 401,
			Success:    false,
			Error:      "Login failed",
			Latency:    0,
		}
	}

	return executeRequest(client, "GET", baseURL+"/api/conversations", nil, nil)
}

func executeRequest(client *http.Client, method, url string, body *bytes.Buffer, headers map[string]string) TestResult {
	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = body
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	startTime := time.Now()

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return TestResult{
			Method:     method,
			URL:        url,
			StatusCode: 0,
			Success:    false,
			Error:      err.Error(),
			Latency:    time.Since(startTime),
		}
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	latency := time.Since(startTime)

	if err != nil {
		return TestResult{
			Method:     method,
			URL:        url,
			StatusCode: 0,
			Success:    false,
			Error:      err.Error(),
			Latency:    latency,
		}
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	success := resp.StatusCode >= 200 && resp.StatusCode < 400
	errorMsg := ""
	if !success {
		errorMsg = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return TestResult{
		Method:       method,
		URL:          url,
		StatusCode:   resp.StatusCode,
		Success:      success,
		Error:        errorMsg,
		Latency:      latency,
		ResponseSize: len(responseBody),
	}
}

func analyzeResults(config LoadTestConfig, startTime, endTime time.Time, results []TestResult) LoadTestResults {
	totalRequests := len(results)
	successfulReqs := 0
	failedReqs := 0

	var latencies []time.Duration
	scenarioMetrics := make(map[string][]TestResult)

	for _, result := range results {
		latencies = append(latencies, result.Latency)

		if result.Success {
			successfulReqs++
		} else {
			failedReqs++
		}

		// Group by scenario
		scenarioMetrics[result.Scenario] = append(scenarioMetrics[result.Scenario], result)
	}

	// Sort latencies for percentile calculations
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// Calculate overall metrics
	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}

	avgLatency := time.Duration(0)
	if len(latencies) > 0 {
		avgLatency = totalLatency / time.Duration(len(latencies))
	}

	totalDuration := endTime.Sub(startTime)
	requestsPerSec := float64(totalRequests) / totalDuration.Seconds()
	errorRate := float64(failedReqs) / float64(totalRequests)

	// Calculate percentiles
	p50Latency := getPercentile(latencies, 0.50)
	p95Latency := getPercentile(latencies, 0.95)
	p99Latency := getPercentile(latencies, 0.99)

	minLatency := time.Duration(0)
	maxLatency := time.Duration(0)
	if len(latencies) > 0 {
		minLatency = latencies[0]
		maxLatency = latencies[len(latencies)-1]
	}

	// Calculate scenario-specific metrics
	scenarioResults := make(map[string]Metrics)
	for scenario, scenarioData := range scenarioMetrics {
		scenarioResults[scenario] = calculateScenarioMetrics(scenarioData)
	}

	return LoadTestResults{
		Config:          config,
		StartTime:       startTime,
		EndTime:         endTime,
		TotalDuration:   totalDuration,
		TotalRequests:   totalRequests,
		SuccessfulReqs:  successfulReqs,
		FailedReqs:      failedReqs,
		AverageLatency:  avgLatency,
		MinLatency:      minLatency,
		MaxLatency:      maxLatency,
		P50Latency:      p50Latency,
		P95Latency:      p95Latency,
		P99Latency:      p99Latency,
		RequestsPerSec:  requestsPerSec,
		ErrorRate:       errorRate,
		ScenarioResults: scenarioResults,
		DetailedResults: results,
	}
}

func getPercentile(sortedLatencies []time.Duration, percentile float64) time.Duration {
	if len(sortedLatencies) == 0 {
		return 0
	}

	index := int(float64(len(sortedLatencies)) * percentile)
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}

	return sortedLatencies[index]
}

func calculateScenarioMetrics(results []TestResult) Metrics {
	totalRequests := len(results)
	successfulReqs := 0
	failedReqs := 0

	var latencies []time.Duration

	for _, result := range results {
		latencies = append(latencies, result.Latency)

		if result.Success {
			successfulReqs++
		} else {
			failedReqs++
		}
	}

	// Sort latencies
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	// Calculate metrics
	var totalLatency time.Duration
	for _, latency := range latencies {
		totalLatency += latency
	}

	avgLatency := time.Duration(0)
	if len(latencies) > 0 {
		avgLatency = totalLatency / time.Duration(len(latencies))
	}

	minLatency := time.Duration(0)
	maxLatency := time.Duration(0)
	if len(latencies) > 0 {
		minLatency = latencies[0]
		maxLatency = latencies[len(latencies)-1]
	}

	p95Latency := getPercentile(latencies, 0.95)
	p99Latency := getPercentile(latencies, 0.99)

	errorRate := float64(failedReqs) / float64(totalRequests)

	return Metrics{
		TotalRequests:  totalRequests,
		SuccessfulReqs: successfulReqs,
		FailedReqs:     failedReqs,
		AverageLatency: avgLatency,
		MinLatency:     minLatency,
		MaxLatency:     maxLatency,
		P95Latency:     p95Latency,
		P99Latency:     p99Latency,
		ErrorRate:      errorRate,
	}
}

func outputResults(results LoadTestResults, format, filename string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(results, "", "  ")
	case "csv":
		output = []byte(generateCSVReport(results))
	case "html":
		output = []byte(generateHTMLReport(results))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return err
	}

	if filename == "" {
		fmt.Print(string(output))
	} else {
		return ioutil.WriteFile(filename, output, 0644)
	}

	return nil
}

func generateCSVReport(results LoadTestResults) string {
	var csv strings.Builder

	// Header
	csv.WriteString("Scenario,Method,URL,StatusCode,Latency(ms),Success,Error,Timestamp,ResponseSize\n")

	// Data rows
	for _, result := range results.DetailedResults {
		csv.WriteString(fmt.Sprintf("%s,%s,%s,%d,%.2f,%t,%s,%s,%d\n",
			result.Scenario,
			result.Method,
			result.URL,
			result.StatusCode,
			float64(result.Latency.Nanoseconds())/1000000, // Convert to milliseconds
			result.Success,
			result.Error,
			result.Timestamp.Format(time.RFC3339),
			result.ResponseSize,
		))
	}

	return csv.String()
}

func generateHTMLReport(results LoadTestResults) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Load Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .summary { background: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background: white; border-radius: 3px; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .success { color: green; }
        .error { color: red; }
    </style>
</head>
<body>
    <h1>Load Test Report</h1>
    
    <div class="summary">
        <h2>Summary</h2>
        <div class="metric"><strong>Total Requests:</strong> %d</div>
        <div class="metric"><strong>Successful:</strong> <span class="success">%d</span></div>
        <div class="metric"><strong>Failed:</strong> <span class="error">%d</span></div>
        <div class="metric"><strong>Average Latency:</strong> %v</div>
        <div class="metric"><strong>P95 Latency:</strong> %v</div>
        <div class="metric"><strong>P99 Latency:</strong> %v</div>
        <div class="metric"><strong>Requests/Second:</strong> %.2f</div>
        <div class="metric"><strong>Error Rate:</strong> %.2f%%</div>
        <div class="metric"><strong>Duration:</strong> %v</div>
    </div>
    
    <h2>Scenario Results</h2>
    <table>
        <tr>
            <th>Scenario</th>
            <th>Total Requests</th>
            <th>Successful</th>
            <th>Failed</th>
            <th>Avg Latency</th>
            <th>P95 Latency</th>
            <th>Error Rate</th>
        </tr>`

	html = fmt.Sprintf(html,
		results.TotalRequests,
		results.SuccessfulReqs,
		results.FailedReqs,
		results.AverageLatency,
		results.P95Latency,
		results.P99Latency,
		results.RequestsPerSec,
		results.ErrorRate*100,
		results.TotalDuration,
	)

	for scenario, metrics := range results.ScenarioResults {
		html += fmt.Sprintf(`
        <tr>
            <td>%s</td>
            <td>%d</td>
            <td class="success">%d</td>
            <td class="error">%d</td>
            <td>%v</td>
            <td>%v</td>
            <td>%.2f%%</td>
        </tr>`,
			scenario,
			metrics.TotalRequests,
			metrics.SuccessfulReqs,
			metrics.FailedReqs,
			metrics.AverageLatency,
			metrics.P95Latency,
			metrics.ErrorRate*100,
		)
	}

	html += `
    </table>
</body>
</html>`

	return html
}

func printSummary(results LoadTestResults) {
	fmt.Printf("\n=== Load Test Summary ===\n")
	fmt.Printf("Duration: %v\n", results.TotalDuration)
	fmt.Printf("Total Requests: %d\n", results.TotalRequests)
	fmt.Printf("Successful: %d (%.1f%%)\n", results.SuccessfulReqs, float64(results.SuccessfulReqs)/float64(results.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.1f%%)\n", results.FailedReqs, results.ErrorRate*100)
	fmt.Printf("Requests/Second: %.2f\n", results.RequestsPerSec)
	fmt.Printf("Average Latency: %v\n", results.AverageLatency)
	fmt.Printf("P50 Latency: %v\n", results.P50Latency)
	fmt.Printf("P95 Latency: %v\n", results.P95Latency)
	fmt.Printf("P99 Latency: %v\n", results.P99Latency)
	fmt.Printf("Min Latency: %v\n", results.MinLatency)
	fmt.Printf("Max Latency: %v\n", results.MaxLatency)

	if results.ErrorRate > 0.05 { // 5% error rate threshold
		fmt.Printf("\n⚠️  High error rate detected: %.1f%%\n", results.ErrorRate*100)
	}

	if results.P95Latency > 2*time.Second {
		fmt.Printf("\n⚠️  High P95 latency detected: %v\n", results.P95Latency)
	}

	fmt.Printf("\n=== Scenario Breakdown ===\n")
	for scenario, metrics := range results.ScenarioResults {
		fmt.Printf("%s: %d requests, %.1f%% success, %v avg latency\n",
			scenario, metrics.TotalRequests,
			float64(metrics.SuccessfulReqs)/float64(metrics.TotalRequests)*100,
			metrics.AverageLatency)
	}
}
