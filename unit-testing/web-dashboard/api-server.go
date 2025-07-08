package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// TestResult represents a test execution result
type TestResult struct {
	Category    string    `json:"category"`
	Total       int       `json:"total"`
	Passed      int       `json:"passed"`
	Failed      int       `json:"failed"`
	Duration    int       `json:"duration"`
	LastRun     time.Time `json:"last_run"`
	Status      string    `json:"status"`
	Output      string    `json:"output"`
	CoverageURL string    `json:"coverage_url,omitempty"`
}

// CoverageData represents code coverage information
type CoverageData struct {
	Lines      float64 `json:"lines"`
	Functions  float64 `json:"functions"`
	Branches   float64 `json:"branches"`
	Statements float64 `json:"statements"`
	Total      float64 `json:"total"`
}

// TestExecution represents an ongoing test execution
type TestExecution struct {
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	Progress  int       `json:"progress"`
	Output    []string  `json:"output"`
}

// DashboardServer manages the test dashboard API
type DashboardServer struct {
	testResults   map[string]*TestResult
	coverageData  *CoverageData
	executions    map[string]*TestExecution
	clients       map[*websocket.Conn]bool
	mutex         sync.RWMutex
	upgrader      websocket.Upgrader
	reportsDir    string
	coverageDir   string
}

// NewDashboardServer creates a new dashboard server instance
func NewDashboardServer() *DashboardServer {
	return &DashboardServer{
		testResults:  make(map[string]*TestResult),
		coverageData: &CoverageData{},
		executions:   make(map[string]*TestExecution),
		clients:      make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		reportsDir:  "../test-reports",
		coverageDir: "../coverage",
	}
}

// Start starts the dashboard server
func (ds *DashboardServer) Start(port int) {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/test-results", ds.handleGetTestResults).Methods("GET")
	api.HandleFunc("/coverage", ds.handleGetCoverage).Methods("GET")
	api.HandleFunc("/run-tests", ds.handleRunTests).Methods("POST")
	api.HandleFunc("/status", ds.handleGetStatus).Methods("GET")
	api.HandleFunc("/reports", ds.handleGetReports).Methods("GET")
	api.HandleFunc("/export", ds.handleExportResults).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", ds.handleWebSocket)

	// Static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./"))).Methods("GET")

	// CORS middleware
	router.Use(ds.corsMiddleware)

	// Start background tasks
	go ds.loadExistingData()
	go ds.periodicDataRefresh()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Dashboard server starting on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

// CORS middleware
func (ds *DashboardServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleGetTestResults returns current test results
func (ds *DashboardServer) handleGetTestResults(w http.ResponseWriter, r *http.Request) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ds.testResults)
}

// handleGetCoverage returns coverage data
func (ds *DashboardServer) handleGetCoverage(w http.ResponseWriter, r *http.Request) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ds.coverageData)
}

// handleRunTests starts test execution
func (ds *DashboardServer) handleRunTests(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Category string `json:"category"`
		Options  string `json:"options,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start test execution in background
	go ds.executeTests(request.Category, request.Options)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": fmt.Sprintf("Test execution started for category: %s", request.Category),
	})
}

// handleGetStatus returns current execution status
func (ds *DashboardServer) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ds.executions)
}

// handleGetReports returns available test reports
func (ds *DashboardServer) handleGetReports(w http.ResponseWriter, r *http.Request) {
	reports, err := ds.getAvailableReports()
	if err != nil {
		http.Error(w, "Error reading reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// handleExportResults exports test results in various formats
func (ds *DashboardServer) handleExportResults(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	ds.mutex.RLock()
	data := map[string]interface{}{
		"test_results": ds.testResults,
		"coverage":     ds.coverageData,
		"timestamp":    time.Now(),
	}
	ds.mutex.RUnlock()

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=test-results.json")
		json.NewEncoder(w).Encode(data)
	case "csv":
		ds.exportCSV(w, data)
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
	}
}

// handleWebSocket manages WebSocket connections for real-time updates
func (ds *DashboardServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ds.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	ds.mutex.Lock()
	ds.clients[conn] = true
	ds.mutex.Unlock()

	// Send initial data
	ds.sendToClient(conn, map[string]interface{}{
		"type":         "initial_data",
		"test_results": ds.testResults,
		"coverage":     ds.coverageData,
	})

	// Keep connection alive and handle client messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			ds.mutex.Lock()
			delete(ds.clients, conn)
			ds.mutex.Unlock()
			break
		}
	}
}

// executeTests runs tests for the specified category
func (ds *DashboardServer) executeTests(category, options string) {
	ds.mutex.Lock()
	execution := &TestExecution{
		Category:  category,
		Status:    "running",
		StartTime: time.Now(),
		Progress:  0,
		Output:    []string{},
	}
	ds.executions[category] = execution
	ds.mutex.Unlock()

	// Broadcast start
	ds.broadcastUpdate(map[string]interface{}{
		"type":      "test_started",
		"category":  category,
		"execution": execution,
	})

	// Build test command
	var cmd *exec.Cmd
	switch category {
	case "all":
		cmd = exec.Command("../test.sh", "all", "--json")
	case "unit", "integration", "auth", "messaging", "frontend", "e2e":
		cmd = exec.Command("../test.sh", category, "--json")
	default:
		ds.updateExecutionStatus(category, "failed", "Unknown test category")
		return
	}

	// Add options if provided
	if options != "" {
		cmd.Args = append(cmd.Args, strings.Fields(options)...)
	}

	// Set working directory
	cmd.Dir = ".."

	// Execute command and capture output
	output, err := cmd.CombinedOutput()
	
	ds.mutex.Lock()
	execution.Progress = 100
	if err != nil {
		execution.Status = "failed"
		execution.Output = append(execution.Output, fmt.Sprintf("Error: %v", err))
	} else {
		execution.Status = "completed"
	}
	execution.Output = append(execution.Output, string(output))
	ds.mutex.Unlock()

	// Parse results and update
	result := ds.parseTestOutput(category, string(output))
	ds.mutex.Lock()
	ds.testResults[category] = result
	ds.mutex.Unlock()

	// Broadcast completion
	ds.broadcastUpdate(map[string]interface{}{
		"type":      "test_completed",
		"category":  category,
		"execution": execution,
		"result":    result,
	})

	// Load updated coverage data
	ds.loadCoverageData()
}

// parseTestOutput parses test command output to extract results
func (ds *DashboardServer) parseTestOutput(category, output string) *TestResult {
	result := &TestResult{
		Category: category,
		LastRun:  time.Now(),
		Status:   "completed",
		Output:   output,
	}

	// Parse Go test output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "PASS") || strings.Contains(line, "FAIL") {
			// Extract test counts and timing
			if matches := regexp.MustCompile(`(\d+) passed`).FindStringSubmatch(line); len(matches) > 1 {
				result.Passed, _ = strconv.Atoi(matches[1])
			}
			if matches := regexp.MustCompile(`(\d+) failed`).FindStringSubmatch(line); len(matches) > 1 {
				result.Failed, _ = strconv.Atoi(matches[1])
			}
			if matches := regexp.MustCompile(`(\d+\.\d+)s`).FindStringSubmatch(line); len(matches) > 1 {
				duration, _ := strconv.ParseFloat(matches[1], 64)
				result.Duration = int(duration)
			}
		}
	}

	result.Total = result.Passed + result.Failed

	if result.Failed > 0 {
		result.Status = "failed"
	} else if result.Passed > 0 {
		result.Status = "passed"
	}

	return result
}

// loadExistingData loads existing test reports and coverage data
func (ds *DashboardServer) loadExistingData() {
	ds.loadTestReports()
	ds.loadCoverageData()
}

// loadTestReports loads existing test reports from files
func (ds *DashboardServer) loadTestReports() {
	files, err := ioutil.ReadDir(ds.reportsDir)
	if err != nil {
		log.Printf("Error reading reports directory: %v", err)
		return
	}

	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			category := ds.extractCategoryFromFilename(file.Name())
			if category != "" {
				content, err := ioutil.ReadFile(filepath.Join(ds.reportsDir, file.Name()))
				if err == nil {
					result := ds.parseTestOutput(category, string(content))
					result.LastRun = file.ModTime()
					ds.testResults[category] = result
				}
			}
		}
	}
}

// loadCoverageData loads coverage information
func (ds *DashboardServer) loadCoverageData() {
	// Look for latest coverage file
	files, err := ioutil.ReadDir(ds.coverageDir)
	if err != nil {
		return
	}

	var latestFile os.FileInfo
	for _, file := range files {
		if strings.HasSuffix(file.Name(), "_summary.txt") {
			if latestFile == nil || file.ModTime().After(latestFile.ModTime()) {
				latestFile = file
			}
		}
	}

	if latestFile != nil {
		content, err := ioutil.ReadFile(filepath.Join(ds.coverageDir, latestFile.Name()))
		if err == nil {
			ds.parseCoverageData(string(content))
		}
	}
}

// parseCoverageData parses Go coverage output
func (ds *DashboardServer) parseCoverageData(content string) {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			if matches := regexp.MustCompile(`(\d+\.\d+)%`).FindStringSubmatch(line); len(matches) > 1 {
				coverage, _ := strconv.ParseFloat(matches[1], 64)
				ds.coverageData.Total = coverage
				ds.coverageData.Lines = coverage
				ds.coverageData.Functions = coverage * 0.95  // Estimate
				ds.coverageData.Branches = coverage * 0.85   // Estimate
				ds.coverageData.Statements = coverage * 0.92 // Estimate
			}
		}
	}
}

// Helper functions
func (ds *DashboardServer) extractCategoryFromFilename(filename string) string {
	parts := strings.Split(filename, "_")
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func (ds *DashboardServer) updateExecutionStatus(category, status, message string) {
	ds.mutex.Lock()
	if execution, exists := ds.executions[category]; exists {
		execution.Status = status
		execution.Output = append(execution.Output, message)
	}
	ds.mutex.Unlock()
}

func (ds *DashboardServer) broadcastUpdate(data map[string]interface{}) {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	for client := range ds.clients {
		ds.sendToClient(client, data)
	}
}

func (ds *DashboardServer) sendToClient(conn *websocket.Conn, data map[string]interface{}) {
	if err := conn.WriteJSON(data); err != nil {
		log.Printf("WebSocket write error: %v", err)
		conn.Close()
		delete(ds.clients, conn)
	}
}

func (ds *DashboardServer) getAvailableReports() ([]string, error) {
	files, err := ioutil.ReadDir(ds.reportsDir)
	if err != nil {
		return nil, err
	}

	var reports []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			reports = append(reports, file.Name())
		}
	}

	return reports, nil
}

func (ds *DashboardServer) exportCSV(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=test-results.csv")

	// Simple CSV export
	fmt.Fprintf(w, "Category,Total,Passed,Failed,Duration,Status,LastRun\n")
	
	if testResults, ok := data["test_results"].(map[string]*TestResult); ok {
		for _, result := range testResults {
			fmt.Fprintf(w, "%s,%d,%d,%d,%d,%s,%s\n",
				result.Category, result.Total, result.Passed, result.Failed,
				result.Duration, result.Status, result.LastRun.Format(time.RFC3339))
		}
	}
}

func (ds *DashboardServer) periodicDataRefresh() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ds.loadExistingData()
		
		// Broadcast updates to connected clients
		ds.broadcastUpdate(map[string]interface{}{
			"type":         "data_refresh",
			"test_results": ds.testResults,
			"coverage":     ds.coverageData,
			"timestamp":    time.Now(),
		})
	}
}

func main() {
	server := NewDashboardServer()
	server.Start(8082)
}
