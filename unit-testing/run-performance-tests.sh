#!/bin/bash

# Performance and Load Testing Runner
# Comprehensive script for running performance, load, and stress tests

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORTS_DIR="$SCRIPT_DIR/performance-reports"
LOAD_TEST_TOOL="$SCRIPT_DIR/load-test-tool"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
DEFAULT_DURATION="30s"
DEFAULT_USERS="50"
DEFAULT_BASE_URL="http://localhost:8080"
VERBOSE=false
GENERATE_REPORTS=true
RUN_BENCHMARKS=true
RUN_LOAD_TESTS=true
RUN_STRESS_TESTS=true

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
}

log_header() {
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}"
}

# Help function
show_help() {
    cat << EOF
${CYAN}Performance and Load Testing Runner${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS]

${YELLOW}OPTIONS:${NC}
    --duration DURATION     Test duration (default: ${DEFAULT_DURATION})
    --users N               Number of concurrent users (default: ${DEFAULT_USERS})
    --url URL               Base URL for testing (default: ${DEFAULT_BASE_URL})
    --no-benchmarks         Skip Go benchmark tests
    --no-load-tests         Skip load tests
    --no-stress-tests       Skip stress tests
    --no-reports            Skip report generation
    --verbose               Enable verbose output
    --help                  Show this help message

${YELLOW}TEST TYPES:${NC}
    - Go Benchmarks: Built-in Go benchmark tests
    - Load Tests: HTTP load testing with multiple scenarios
    - Stress Tests: System limit testing with gradual load increase
    - WebSocket Tests: Real-time communication performance

${YELLOW}EXAMPLES:${NC}
    $0                                    # Run all tests with defaults
    $0 --duration 60s --users 100        # Custom duration and user count
    $0 --no-stress-tests --verbose       # Skip stress tests with verbose output
    $0 --url http://localhost:8081       # Test different server

${YELLOW}REPORTS:${NC}
    Performance reports are saved to: ${REPORTS_DIR}/
    Formats: JSON, CSV, HTML

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --duration)
                DEFAULT_DURATION="$2"
                shift 2
                ;;
            --users)
                DEFAULT_USERS="$2"
                shift 2
                ;;
            --url)
                DEFAULT_BASE_URL="$2"
                shift 2
                ;;
            --no-benchmarks)
                RUN_BENCHMARKS=false
                shift
                ;;
            --no-load-tests)
                RUN_LOAD_TESTS=false
                shift
                ;;
            --no-stress-tests)
                RUN_STRESS_TESTS=false
                shift
                ;;
            --no-reports)
                GENERATE_REPORTS=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use '$0 --help' for usage information."
                exit 1
                ;;
        esac
    done
}

# Setup environment
setup_environment() {
    log_info "Setting up performance testing environment..."
    
    # Create reports directory
    mkdir -p "$REPORTS_DIR"
    
    # Check prerequisites
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run performance tests."
        exit 1
    fi
    
    # Build load test tool if it doesn't exist
    if [ ! -f "$LOAD_TEST_TOOL" ]; then
        log_info "Building load test tool..."
        cd "$SCRIPT_DIR"
        go build -o load-test-tool load-test-tool.go
        if [ $? -ne 0 ]; then
            log_error "Failed to build load test tool"
            exit 1
        fi
    fi
    
    log_success "Environment setup completed"
}

# Run Go benchmark tests
run_benchmark_tests() {
    if [ "$RUN_BENCHMARKS" = false ]; then
        log_info "Skipping benchmark tests"
        return 0
    fi
    
    log_header "Running Go Benchmark Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local benchmark_report="$REPORTS_DIR/benchmarks_${timestamp}.txt"
    local benchmark_json="$REPORTS_DIR/benchmarks_${timestamp}.json"
    
    log_info "Running benchmark tests..."
    
    cd "$SCRIPT_DIR"
    
    # Run benchmarks with different options
    local benchmark_tests=(
        "BenchmarkUserRegistration"
        "BenchmarkUserLogin"
        "BenchmarkPostCreation"
        "BenchmarkPostRetrieval"
        "BenchmarkCommentCreation"
        "BenchmarkMessageSending"
        "BenchmarkDatabaseOperations"
    )
    
    for test in "${benchmark_tests[@]}"; do
        log_info "Running $test..."
        
        if [ "$VERBOSE" = true ]; then
            go test -bench="^${test}$" -benchmem -count=3 -timeout=10m ./... | tee -a "$benchmark_report"
        else
            go test -bench="^${test}$" -benchmem -count=3 -timeout=10m ./... >> "$benchmark_report" 2>&1
        fi
        
        if [ $? -eq 0 ]; then
            log_success "$test completed"
        else
            log_error "$test failed"
        fi
    done
    
    # Generate JSON report from benchmark results
    if command -v benchstat &> /dev/null; then
        log_info "Generating benchmark statistics..."
        benchstat "$benchmark_report" > "${benchmark_report%.txt}_stats.txt"
    fi
    
    log_success "Benchmark tests completed. Report: $benchmark_report"
}

# Run load tests
run_load_tests() {
    if [ "$RUN_LOAD_TESTS" = false ]; then
        log_info "Skipping load tests"
        return 0
    fi
    
    log_header "Running Load Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local load_report="$REPORTS_DIR/load_test_${timestamp}"
    
    log_info "Starting load tests with $DEFAULT_USERS users for $DEFAULT_DURATION..."
    log_info "Target URL: $DEFAULT_BASE_URL"
    
    # Check if server is running
    if ! curl -s "$DEFAULT_BASE_URL" >/dev/null 2>&1; then
        log_warn "Server at $DEFAULT_BASE_URL is not responding. Starting local server..."
        
        # Start server in background
        cd "$PROJECT_ROOT"
        ./run.sh --port 8080 &
        local server_pid=$!
        
        # Wait for server to start
        log_info "Waiting for server to start..."
        for i in {1..30}; do
            if curl -s "$DEFAULT_BASE_URL" >/dev/null 2>&1; then
                log_success "Server started successfully"
                break
            fi
            sleep 1
        done
        
        if ! curl -s "$DEFAULT_BASE_URL" >/dev/null 2>&1; then
            log_error "Failed to start server"
            kill $server_pid 2>/dev/null || true
            return 1
        fi
    fi
    
    # Run different load test scenarios
    local scenarios=(
        "homepage,login,posts"
        "signup,create_post,add_comment"
        "messaging,conversations"
        "all"
    )
    
    for scenario in "${scenarios[@]}"; do
        log_info "Running load test scenario: $scenario"
        
        local scenario_name=$(echo "$scenario" | tr ',' '_')
        local scenario_report="${load_report}_${scenario_name}"
        
        if [ "$VERBOSE" = true ]; then
            "$LOAD_TEST_TOOL" \
                --url "$DEFAULT_BASE_URL" \
                --users "$DEFAULT_USERS" \
                --duration "$DEFAULT_DURATION" \
                --scenarios "$scenario" \
                --format json \
                --output "${scenario_report}.json" \
                --verbose
        else
            "$LOAD_TEST_TOOL" \
                --url "$DEFAULT_BASE_URL" \
                --users "$DEFAULT_USERS" \
                --duration "$DEFAULT_DURATION" \
                --scenarios "$scenario" \
                --format json \
                --output "${scenario_report}.json" \
                > "${scenario_report}.log" 2>&1
        fi
        
        if [ $? -eq 0 ]; then
            log_success "Load test scenario '$scenario' completed"
            
            # Generate additional report formats
            if [ "$GENERATE_REPORTS" = true ]; then
                "$LOAD_TEST_TOOL" \
                    --url "$DEFAULT_BASE_URL" \
                    --users "$DEFAULT_USERS" \
                    --duration "$DEFAULT_DURATION" \
                    --scenarios "$scenario" \
                    --format html \
                    --output "${scenario_report}.html" \
                    > /dev/null 2>&1
                
                "$LOAD_TEST_TOOL" \
                    --url "$DEFAULT_BASE_URL" \
                    --users "$DEFAULT_USERS" \
                    --duration "$DEFAULT_DURATION" \
                    --scenarios "$scenario" \
                    --format csv \
                    --output "${scenario_report}.csv" \
                    > /dev/null 2>&1
            fi
        else
            log_error "Load test scenario '$scenario' failed"
        fi
    done
    
    # Kill server if we started it
    if [ -n "$server_pid" ]; then
        log_info "Stopping test server..."
        kill $server_pid 2>/dev/null || true
        wait $server_pid 2>/dev/null || true
    fi
    
    log_success "Load tests completed. Reports: ${load_report}_*"
}

# Run stress tests
run_stress_tests() {
    if [ "$RUN_STRESS_TESTS" = false ]; then
        log_info "Skipping stress tests"
        return 0
    fi
    
    log_header "Running Stress Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local stress_report="$REPORTS_DIR/stress_test_${timestamp}.txt"
    
    log_info "Running stress tests..."
    log_warn "Stress tests may take longer and use significant system resources"
    
    cd "$SCRIPT_DIR"
    
    # Run stress tests
    local stress_tests=(
        "TestStressUserRegistration"
        "TestStressWebSocketConnections"
        "TestStressDatabaseOperations"
        "TestStressMemoryUsage"
    )
    
    for test in "${stress_tests[@]}"; do
        log_info "Running $test..."
        
        if [ "$VERBOSE" = true ]; then
            go test -run="^${test}$" -timeout=15m -v ./... | tee -a "$stress_report"
        else
            go test -run="^${test}$" -timeout=15m ./... >> "$stress_report" 2>&1
        fi
        
        if [ $? -eq 0 ]; then
            log_success "$test completed"
        else
            log_error "$test failed"
        fi
    done
    
    log_success "Stress tests completed. Report: $stress_report"
}

# Generate summary report
generate_summary_report() {
    if [ "$GENERATE_REPORTS" = false ]; then
        return 0
    fi
    
    log_header "Generating Summary Report"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local summary_report="$REPORTS_DIR/performance_summary_${timestamp}.html"
    
    log_info "Generating comprehensive performance summary..."
    
    # Create HTML summary report
    cat > "$summary_report" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Performance Test Summary</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background: #e9ecef; border-radius: 3px; }
        .success { color: green; }
        .warning { color: orange; }
        .error { color: red; }
        table { border-collapse: collapse; width: 100%; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Real-Time Forum Performance Test Summary</h1>
        <p>Generated: $(date)</p>
        <p>Test Configuration:</p>
        <ul>
            <li>Duration: $DEFAULT_DURATION</li>
            <li>Concurrent Users: $DEFAULT_USERS</li>
            <li>Base URL: $DEFAULT_BASE_URL</li>
        </ul>
    </div>
    
    <div class="section">
        <h2>Test Results Overview</h2>
        <div class="metric"><strong>Benchmarks:</strong> $([ "$RUN_BENCHMARKS" = true ] && echo "✅ Executed" || echo "⏭️ Skipped")</div>
        <div class="metric"><strong>Load Tests:</strong> $([ "$RUN_LOAD_TESTS" = true ] && echo "✅ Executed" || echo "⏭️ Skipped")</div>
        <div class="metric"><strong>Stress Tests:</strong> $([ "$RUN_STRESS_TESTS" = true ] && echo "✅ Executed" || echo "⏭️ Skipped")</div>
    </div>
    
    <div class="section">
        <h2>Report Files</h2>
        <p>All detailed reports are available in: <code>$REPORTS_DIR/</code></p>
        <ul>
EOF

    # List all generated reports
    find "$REPORTS_DIR" -name "*${timestamp%_*}*" -type f | while read -r file; do
        echo "            <li><a href=\"$(basename "$file")\">$(basename "$file")</a></li>" >> "$summary_report"
    done

    cat >> "$summary_report" << EOF
        </ul>
    </div>
    
    <div class="section">
        <h2>Performance Recommendations</h2>
        <ul>
            <li>Monitor response times under load</li>
            <li>Check for memory leaks during stress tests</li>
            <li>Validate error rates remain below 5%</li>
            <li>Ensure P95 latency stays under 2 seconds</li>
            <li>Monitor database connection pool usage</li>
        </ul>
    </div>
</body>
</html>
EOF

    log_success "Summary report generated: $summary_report"
}

# Main execution
main() {
    local start_time=$(date +%s)
    
    log_header "Real-Time Forum Performance Testing Suite"
    
    # Parse arguments
    parse_arguments "$@"
    
    # Setup environment
    setup_environment
    
    # Run tests
    run_benchmark_tests
    run_load_tests
    run_stress_tests
    
    # Generate reports
    generate_summary_report
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_header "Performance Testing Complete"
    log_success "Total execution time: ${duration} seconds"
    log_info "Reports available in: $REPORTS_DIR/"
    
    # Open summary report if available
    local latest_summary=$(find "$REPORTS_DIR" -name "performance_summary_*.html" -type f | sort | tail -1)
    if [ -n "$latest_summary" ] && [ "$GENERATE_REPORTS" = true ]; then
        log_info "Summary report: $latest_summary"
        
        # Try to open in browser
        if command -v open &> /dev/null; then
            open "$latest_summary"
        elif command -v xdg-open &> /dev/null; then
            xdg-open "$latest_summary"
        fi
    fi
}

# Run main function
main "$@"
