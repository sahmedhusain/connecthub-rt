#!/bin/bash

# Real-Time Forum Test Runner
# Comprehensive testing suite with interactive menu and reporting
# Version: 1.0.0

set -e  # Exit on any error

# Configuration
TEST_DIR="."
REPORT_DIR="test-reports"
COVERAGE_DIR="coverage"
LOG_DIR="../logs"
VERBOSE=false
QUIET=false
GENERATE_HTML=false
RUN_COVERAGE=false
PARALLEL_TESTS=false
TEST_TIMEOUT="10m"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    if [ "$QUIET" != true ]; then
        echo -e "${GREEN}[INFO]${NC} $1"
    fi
}

log_warn() {
    if [ "$QUIET" != true ]; then
        echo -e "${YELLOW}[WARN]${NC} $1"
    fi
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_debug() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

log_header() {
    if [ "$QUIET" != true ]; then
        echo -e "${PURPLE}========================================${NC}"
        echo -e "${PURPLE}$1${NC}"
        echo -e "${PURPLE}========================================${NC}"
    fi
}

# Help function
show_help() {
    cat << EOF
${CYAN}Real-Time Forum Test Runner${NC}
${CYAN}Version: 1.0.0${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS] [TEST_CATEGORY]

${YELLOW}TEST CATEGORIES:${NC}
    ${CYAN}Backend Tests:${NC}
    all                 Run all tests (default)
    unit                Run unit tests only
    integration         Run integration tests only
    auth                Run authentication tests
    posts               Run post and comment tests
    messaging           Run messaging and WebSocket tests
    database            Run database and repository tests
    middleware          Run middleware and security tests
    api                 Run API endpoint tests
    websocket           Run WebSocket functionality tests

    ${CYAN}Frontend Tests:${NC}
    frontend            Run all frontend unit tests
    frontend-dom        Run DOM manipulation tests
    frontend-websocket  Run WebSocket client tests
    frontend-auth       Run frontend authentication tests
    frontend-spa        Run SPA navigation tests

    ${CYAN}E2E Tests:${NC}
    e2e                 Run all end-to-end tests
    e2e-auth            Run E2E authentication flow tests
    e2e-messaging       Run E2E real-time messaging tests
    cross-browser       Run tests across multiple browsers
    responsive          Run responsive design tests

    ${CYAN}Performance Tests:${NC}
    benchmarks          Run Go benchmark tests
    load-tests          Run HTTP load tests
    stress-tests        Run stress and limit tests
    websocket-performance Run WebSocket performance tests
    performance-all     Run all performance tests

${YELLOW}OPTIONS:${NC}
    -v, --verbose       Enable verbose output
    -q, --quiet         Disable output (except errors)
    -h, --html          Generate HTML test report
    -c, --coverage      Run with coverage analysis
    -p, --parallel      Run tests in parallel
    -t, --timeout TIME  Set test timeout (default: ${TEST_TIMEOUT})
    --help              Show this help message

${YELLOW}EXAMPLES:${NC}
    $0                          # Interactive mode
    $0 all --coverage --html    # Run all tests with coverage and HTML report
    $0 auth --verbose           # Run auth tests with verbose output
    $0 unit --parallel          # Run unit tests in parallel
    $0 messaging --quiet        # Run messaging tests quietly
    $0 benchmarks               # Run Go benchmark tests
    $0 load-tests --verbose     # Run load tests with detailed output
    $0 performance-all          # Run all performance tests

${YELLOW}REPORTS:${NC}
    Test reports are saved to: ${REPORT_DIR}/
    Coverage reports are saved to: ${COVERAGE_DIR}/
    Test logs are saved to: ${LOG_DIR}/

${YELLOW}REQUIREMENTS:${NC}
    - Go 1.19 or later (for backend tests)
    - Node.js 16+ and npm (for frontend and E2E tests)
    - SQLite3
    - All project dependencies installed

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Parse command line arguments
parse_arguments() {
    TEST_CATEGORY=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            all|unit|integration|auth|posts|messaging|database|middleware|api|websocket|frontend|frontend-dom|frontend-websocket|frontend-auth|frontend-spa|e2e|e2e-auth|e2e-messaging|cross-browser|responsive|benchmarks|load-tests|stress-tests|websocket-performance|performance-all)
                TEST_CATEGORY="$1"
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -q|--quiet)
                QUIET=true
                shift
                ;;
            -h|--html)
                GENERATE_HTML=true
                shift
                ;;
            -c|--coverage)
                RUN_COVERAGE=true
                shift
                ;;
            -p|--parallel)
                PARALLEL_TESTS=true
                shift
                ;;
            -t|--timeout)
                TEST_TIMEOUT="$2"
                shift 2
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
    
    # Set default category if none provided
    if [ -z "$TEST_CATEGORY" ]; then
        TEST_CATEGORY="interactive"
    fi
}

# Setup test environment
setup_test_environment() {
    log_debug "Setting up test environment..."
    
    # Create necessary directories
    mkdir -p "$REPORT_DIR" "$COVERAGE_DIR" "$LOG_DIR"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run tests."
        exit 1
    fi
    
    # Check if test directory exists
    if [ ! -d "$TEST_DIR" ]; then
        log_error "Test directory '$TEST_DIR' not found."
        exit 1
    fi
    
    # Verify test files exist
    if [ ! -f "$TEST_DIR/test_helper.go" ]; then
        log_error "Test helper file not found. Please ensure test infrastructure is set up."
        exit 1
    fi
    
    log_debug "Test environment setup completed"
}

# Interactive menu
show_interactive_menu() {
    clear
    log_header "Real-Time Forum Test Runner"
    
    echo -e "${CYAN}Test Environment:${NC}"
    echo -e "Test Directory: ${YELLOW}$TEST_DIR${NC}"
    echo -e "Report Directory: ${YELLOW}$REPORT_DIR${NC}"
    echo -e "Coverage: ${YELLOW}$([ "$RUN_COVERAGE" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "HTML Reports: ${YELLOW}$([ "$GENERATE_HTML" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Parallel Execution: ${YELLOW}$([ "$PARALLEL_TESTS" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo ""
    
    echo -e "${YELLOW}Choose test category to run:${NC}"
    echo "1. All Tests (comprehensive)"
    echo "2. Unit Tests"
    echo "3. Integration Tests"
    echo "4. Authentication Tests"
    echo "5. Post & Comment Tests"
    echo "6. Messaging & WebSocket Tests"
    echo "7. Database & Repository Tests"
    echo "8. Middleware & Security Tests"
    echo "9. API Endpoint Tests"
    echo "10. WebSocket Functionality Tests"
    echo ""
    echo -e "${CYAN}Frontend Tests:${NC}"
    echo "11. Frontend Unit Tests"
    echo "12. Frontend DOM Tests"
    echo "13. Frontend WebSocket Tests"
    echo "14. Frontend Auth Tests"
    echo "15. Frontend SPA Tests"
    echo ""
    echo -e "${CYAN}E2E Tests:${NC}"
    echo "16. All E2E Tests"
    echo "17. E2E Authentication Flow"
    echo "18. E2E Real-time Messaging"
    echo "19. Cross-browser Testing"
    echo "20. Responsive Design Tests"
    echo ""
    echo -e "${CYAN}Performance Tests:${NC}"
    echo "21. Go Benchmarks"
    echo "22. Load Tests"
    echo "23. Stress Tests"
    echo "24. WebSocket Performance"
    echo "25. All Performance Tests"
    echo ""
    echo "26. Test Configuration"
    echo "27. View Test Reports"
    echo "0. Exit"
    echo ""

    read -p "Enter your choice [0-27]: " choice
    
    case $choice in
        1) TEST_CATEGORY="all" ;;
        2) TEST_CATEGORY="unit" ;;
        3) TEST_CATEGORY="integration" ;;
        4) TEST_CATEGORY="auth" ;;
        5) TEST_CATEGORY="posts" ;;
        6) TEST_CATEGORY="messaging" ;;
        7) TEST_CATEGORY="database" ;;
        8) TEST_CATEGORY="middleware" ;;
        9) TEST_CATEGORY="api" ;;
        10) TEST_CATEGORY="websocket" ;;
        11) TEST_CATEGORY="frontend" ;;
        12) TEST_CATEGORY="frontend-dom" ;;
        13) TEST_CATEGORY="frontend-websocket" ;;
        14) TEST_CATEGORY="frontend-auth" ;;
        15) TEST_CATEGORY="frontend-spa" ;;
        16) TEST_CATEGORY="e2e" ;;
        17) TEST_CATEGORY="e2e-auth" ;;
        18) TEST_CATEGORY="e2e-messaging" ;;
        19) TEST_CATEGORY="cross-browser" ;;
        20) TEST_CATEGORY="responsive" ;;
        21) TEST_CATEGORY="benchmarks" ;;
        22) TEST_CATEGORY="load-tests" ;;
        23) TEST_CATEGORY="stress-tests" ;;
        24) TEST_CATEGORY="websocket-performance" ;;
        25) TEST_CATEGORY="performance-all" ;;
        26) show_configuration_menu; return ;;
        27) show_reports_menu; return ;;
        0) log_info "Goodbye!"; exit 0 ;;
        *) log_error "Invalid choice. Please try again."; sleep 2; show_interactive_menu; return ;;
    esac
}

# Configuration menu
show_configuration_menu() {
    clear
    log_header "Test Configuration"
    
    echo -e "${YELLOW}Current Configuration:${NC}"
    echo -e "Verbose Output: ${CYAN}$([ "$VERBOSE" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Quiet Mode: ${CYAN}$([ "$QUIET" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "HTML Reports: ${CYAN}$([ "$GENERATE_HTML" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Coverage Analysis: ${CYAN}$([ "$RUN_COVERAGE" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Parallel Execution: ${CYAN}$([ "$PARALLEL_TESTS" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Test Timeout: ${CYAN}$TEST_TIMEOUT${NC}"
    echo ""
    
    echo -e "${YELLOW}Configuration Options:${NC}"
    echo "1. Toggle Verbose Output"
    echo "2. Toggle Quiet Mode"
    echo "3. Toggle HTML Reports"
    echo "4. Toggle Coverage Analysis"
    echo "5. Toggle Parallel Execution"
    echo "6. Set Test Timeout"
    echo "7. Reset to Defaults"
    echo "8. Back to Main Menu"
    echo "0. Exit"
    echo ""
    
    read -p "Enter your choice [0-8]: " choice
    
    case $choice in
        1)
            VERBOSE=$([ "$VERBOSE" = true ] && echo false || echo true)
            log_info "Verbose output $([ "$VERBOSE" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_configuration_menu
            ;;
        2)
            QUIET=$([ "$QUIET" = true ] && echo false || echo true)
            log_info "Quiet mode $([ "$QUIET" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_configuration_menu
            ;;
        3)
            GENERATE_HTML=$([ "$GENERATE_HTML" = true ] && echo false || echo true)
            log_info "HTML reports $([ "$GENERATE_HTML" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_configuration_menu
            ;;
        4)
            RUN_COVERAGE=$([ "$RUN_COVERAGE" = true ] && echo false || echo true)
            log_info "Coverage analysis $([ "$RUN_COVERAGE" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_configuration_menu
            ;;
        5)
            PARALLEL_TESTS=$([ "$PARALLEL_TESTS" = true ] && echo false || echo true)
            log_info "Parallel execution $([ "$PARALLEL_TESTS" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_configuration_menu
            ;;
        6)
            read -p "Enter test timeout (e.g., 5m, 30s): " new_timeout
            if [[ "$new_timeout" =~ ^[0-9]+[smh]$ ]]; then
                TEST_TIMEOUT="$new_timeout"
                log_info "Test timeout set to: $TEST_TIMEOUT"
            else
                log_error "Invalid timeout format. Use format like: 5m, 30s, 1h"
            fi
            sleep 2
            show_configuration_menu
            ;;
        7)
            VERBOSE=false
            QUIET=false
            GENERATE_HTML=false
            RUN_COVERAGE=false
            PARALLEL_TESTS=false
            TEST_TIMEOUT="10m"
            log_info "Configuration reset to defaults"
            sleep 1
            show_configuration_menu
            ;;
        8)
            show_interactive_menu
            ;;
        0)
            log_info "Goodbye!"
            exit 0
            ;;
        *)
            log_error "Invalid choice. Please try again."
            sleep 2
            show_configuration_menu
            ;;
    esac
}

# Reports menu
show_reports_menu() {
    clear
    log_header "Test Reports"
    
    echo -e "${YELLOW}Available Reports:${NC}"
    
    # List test reports
    if [ -d "$REPORT_DIR" ] && [ "$(ls -A $REPORT_DIR 2>/dev/null)" ]; then
        echo -e "\n${CYAN}Test Reports:${NC}"
        ls -la "$REPORT_DIR" | tail -n +2 | while read -r line; do
            echo "  $line"
        done
    else
        echo -e "\n${YELLOW}No test reports found${NC}"
    fi
    
    # List coverage reports
    if [ -d "$COVERAGE_DIR" ] && [ "$(ls -A $COVERAGE_DIR 2>/dev/null)" ]; then
        echo -e "\n${CYAN}Coverage Reports:${NC}"
        ls -la "$COVERAGE_DIR" | tail -n +2 | while read -r line; do
            echo "  $line"
        done
    else
        echo -e "\n${YELLOW}No coverage reports found${NC}"
    fi
    
    # List log files
    if [ -d "$LOG_DIR" ] && [ "$(ls -A $LOG_DIR 2>/dev/null)" ]; then
        echo -e "\n${CYAN}Log Files:${NC}"
        ls -la "$LOG_DIR" | tail -n +2 | while read -r line; do
            echo "  $line"
        done
    else
        echo -e "\n${YELLOW}No log files found${NC}"
    fi
    
    echo ""
    echo -e "${YELLOW}Options:${NC}"
    echo "1. View latest test report"
    echo "2. View latest coverage report"
    echo "3. View latest log file"
    echo "4. Open reports directory"
    echo "5. Clean old reports"
    echo "6. Back to Main Menu"
    echo "0. Exit"
    echo ""
    
    read -p "Enter your choice [0-6]: " choice
    
    case $choice in
        1)
            latest_report=$(ls -t "$REPORT_DIR"/*.txt 2>/dev/null | head -1)
            if [ -n "$latest_report" ]; then
                log_info "Showing latest test report: $latest_report"
                echo -e "${CYAN}----------------------------------------${NC}"
                cat "$latest_report"
                echo -e "${CYAN}----------------------------------------${NC}"
                read -p "Press Enter to continue..."
            else
                log_warn "No test reports found"
                sleep 2
            fi
            show_reports_menu
            ;;
        2)
            latest_coverage=$(ls -t "$COVERAGE_DIR"/*.html 2>/dev/null | head -1)
            if [ -n "$latest_coverage" ]; then
                log_info "Opening coverage report: $latest_coverage"
                if command -v open &> /dev/null; then
                    open "$latest_coverage"
                elif command -v xdg-open &> /dev/null; then
                    xdg-open "$latest_coverage"
                else
                    log_info "Coverage report location: $latest_coverage"
                fi
            else
                log_warn "No coverage reports found"
                sleep 2
            fi
            show_reports_menu
            ;;
        3)
            latest_log=$(ls -t "$LOG_DIR"/*.log 2>/dev/null | head -1)
            if [ -n "$latest_log" ]; then
                log_info "Showing latest log file: $latest_log"
                echo -e "${CYAN}----------------------------------------${NC}"
                tail -50 "$latest_log"
                echo -e "${CYAN}----------------------------------------${NC}"
                read -p "Press Enter to continue..."
            else
                log_warn "No log files found"
                sleep 2
            fi
            show_reports_menu
            ;;
        4)
            if command -v open &> /dev/null; then
                open "$REPORT_DIR"
            elif command -v xdg-open &> /dev/null; then
                xdg-open "$REPORT_DIR"
            else
                log_info "Reports directory: $(pwd)/$REPORT_DIR"
            fi
            show_reports_menu
            ;;
        5)
            read -p "Are you sure you want to clean old reports? (y/N): " confirm
            if [[ "$confirm" =~ ^[Yy]$ ]]; then
                rm -rf "$REPORT_DIR"/* "$COVERAGE_DIR"/* "$LOG_DIR"/*
                log_info "Old reports cleaned"
            fi
            sleep 1
            show_reports_menu
            ;;
        6)
            show_interactive_menu
            ;;
        0)
            log_info "Goodbye!"
            exit 0
            ;;
        *)
            log_error "Invalid choice. Please try again."
            sleep 2
            show_reports_menu
            ;;
    esac
}

# Run frontend tests
run_frontend_tests() {
    local category="$1"

    log_header "Running Frontend Tests: $category"

    # Check if Node.js and npm are available
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed. Please install Node.js and npm to run frontend tests."
        return 1
    fi

    # Check if package.json exists
    if [ ! -f "package.json" ]; then
        log_error "package.json not found. Please ensure frontend dependencies are set up."
        return 1
    fi

    # Install dependencies if node_modules doesn't exist
    if [ ! -d "node_modules" ]; then
        log_info "Installing frontend dependencies..."
        npm install
    fi

    # Create report file
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/frontend_${category}_${timestamp}.txt"

    # Ensure directories exist
    mkdir -p "$REPORT_DIR"

    # Determine test command based on category
    local test_cmd=""
    case "$category" in
        "frontend")
            test_cmd="npm test"
            ;;
        "frontend-dom")
            test_cmd="npm run test:dom"
            ;;
        "frontend-websocket")
            test_cmd="npm run test:websocket"
            ;;
        "frontend-auth")
            test_cmd="npm run test:auth"
            ;;
        "frontend-spa")
            test_cmd="npm run test:spa"
            ;;
        *)
            test_cmd="npm test"
            ;;
    esac

    log_info "Running command: $test_cmd"

    # Run tests
    local start_time=$(date +%s)

    if [ "$QUIET" = true ]; then
        eval "$test_cmd" > "$report_file" 2>&1
        local exit_code=$?
    else
        eval "$test_cmd" 2>&1 | tee "$report_file"
        local exit_code=${PIPESTATUS[0]}
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Generate test summary
    generate_frontend_test_summary "$category" "$report_file" "$duration" "$exit_code"

    return $exit_code
}

# Run E2E tests
run_e2e_tests() {
    local category="$1"

    log_header "Running E2E Tests: $category"

    # Check if Playwright is available
    if ! command -v npx &> /dev/null; then
        log_error "npx is not installed. Please install Node.js to run E2E tests."
        return 1
    fi

    # Check if package.json exists
    if [ ! -f "package.json" ]; then
        log_error "package.json not found. Please ensure E2E dependencies are set up."
        return 1
    fi

    # Install dependencies if node_modules doesn't exist
    if [ ! -d "node_modules" ]; then
        log_info "Installing E2E dependencies..."
        npm install
    fi

    # Install Playwright browsers if needed
    if [ ! -d "node_modules/@playwright" ]; then
        log_info "Installing Playwright browsers..."
        npx playwright install
    fi

    # Create report file
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/e2e_${category}_${timestamp}.txt"

    # Ensure directories exist
    mkdir -p "$REPORT_DIR"

    # Determine test command based on category
    local test_cmd=""
    case "$category" in
        "e2e")
            test_cmd="npx playwright test"
            ;;
        "e2e-auth")
            test_cmd="npx playwright test auth-flow.spec.js"
            ;;
        "e2e-messaging")
            test_cmd="npx playwright test messaging.spec.js"
            ;;
        "cross-browser")
            test_cmd="npx playwright test --project=chromium --project=firefox --project=webkit"
            ;;
        "responsive")
            test_cmd="npx playwright test --grep responsive"
            ;;
        *)
            test_cmd="npx playwright test"
            ;;
    esac

    log_info "Running command: $test_cmd"

    # Run tests
    local start_time=$(date +%s)

    if [ "$QUIET" = true ]; then
        eval "$test_cmd" > "$report_file" 2>&1
        local exit_code=$?
    else
        eval "$test_cmd" 2>&1 | tee "$report_file"
        local exit_code=${PIPESTATUS[0]}
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Generate test summary
    generate_e2e_test_summary "$category" "$report_file" "$duration" "$exit_code"

    return $exit_code
}

# Run performance tests
run_performance_tests() {
    local category="$1"

    log_header "Running Performance Tests: $category"

    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/performance_${category}_${timestamp}.txt"
    local performance_reports_dir="./performance-reports"

    # Create performance reports directory
    mkdir -p "$performance_reports_dir"

    local exit_code=0

    case "$category" in
        "benchmarks")
            log_info "Running Go benchmark tests..."

            # Run all benchmark tests
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

                if [ "$QUIET" = true ]; then
                    go test -bench="^${test}$" -benchmem -count=3 -timeout=10m ./... >> "$report_file" 2>&1
                else
                    go test -bench="^${test}$" -benchmem -count=3 -timeout=10m ./... | tee -a "$report_file"
                fi

                if [ ${PIPESTATUS[0]} -ne 0 ]; then
                    exit_code=1
                    log_error "$test failed"
                else
                    log_success "$test completed"
                fi
            done
            ;;

        "load-tests")
            log_info "Running load tests..."

            # Build load test tool if needed
            if [ ! -f "./tools/load-test-tool" ]; then
                log_info "Building load test tool..."
                cd tools && go build -o load-test-tool load-test-tool.go && cd ..
                if [ $? -ne 0 ]; then
                    log_error "Failed to build load test tool"
                    return 1
                fi
            fi

            # Check if server is running
            local base_url="http://localhost:8080"
            if ! curl -s "$base_url" >/dev/null 2>&1; then
                log_warn "Server not running. Please start the server first."
                log_info "You can start it with: ./run.sh --port 8080"
                return 1
            fi

            # Run load tests with different scenarios
            local scenarios=("homepage,login,posts" "signup,create_post" "messaging" "all")

            for scenario in "${scenarios[@]}"; do
                local scenario_name=$(echo "$scenario" | tr ',' '_')
                local scenario_report="$performance_reports_dir/load_test_${scenario_name}_${timestamp}"

                log_info "Running load test scenario: $scenario"

                if [ "$QUIET" = true ]; then
                    ./tools/load-test-tool \
                        --url "$base_url" \
                        --users 50 \
                        --duration 30s \
                        --scenarios "$scenario" \
                        --format json \
                        --output "${scenario_report}.json" \
                        >> "$report_file" 2>&1
                else
                    ./tools/load-test-tool \
                        --url "$base_url" \
                        --users 50 \
                        --duration 30s \
                        --scenarios "$scenario" \
                        --format json \
                        --output "${scenario_report}.json" \
                        | tee -a "$report_file"
                fi

                if [ ${PIPESTATUS[0]} -ne 0 ]; then
                    exit_code=1
                    log_error "Load test scenario '$scenario' failed"
                else
                    log_success "Load test scenario '$scenario' completed"
                fi
            done
            ;;

        "stress-tests")
            log_info "Running stress tests..."

            local stress_tests=(
                "TestStressUserRegistration"
                "TestStressWebSocketConnections"
                "TestStressDatabaseOperations"
                "TestStressMemoryUsage"
            )

            for test in "${stress_tests[@]}"; do
                log_info "Running $test..."

                if [ "$QUIET" = true ]; then
                    go test -run="^${test}$" -timeout=15m -v ./... >> "$report_file" 2>&1
                else
                    go test -run="^${test}$" -timeout=15m -v ./... | tee -a "$report_file"
                fi

                if [ ${PIPESTATUS[0]} -ne 0 ]; then
                    exit_code=1
                    log_error "$test failed"
                else
                    log_success "$test completed"
                fi
            done
            ;;

        "websocket-performance")
            log_info "Running WebSocket performance tests..."

            local websocket_tests=(
                "TestWebSocketPerformance"
                "TestWebSocketConcurrentConnections"
                "TestWebSocketMessageThroughput"
            )

            for test in "${websocket_tests[@]}"; do
                log_info "Running $test..."

                if [ "$QUIET" = true ]; then
                    go test -run="^${test}$" -timeout=10m -v ./... >> "$report_file" 2>&1
                else
                    go test -run="^${test}$" -timeout=10m -v ./... | tee -a "$report_file"
                fi

                if [ ${PIPESTATUS[0]} -ne 0 ]; then
                    exit_code=1
                    log_error "$test failed"
                else
                    log_success "$test completed"
                fi
            done
            ;;

        "performance-all")
            log_info "Running all performance tests..."

            # Run all performance test categories
            run_performance_tests "benchmarks"
            local bench_exit=$?

            run_performance_tests "load-tests"
            local load_exit=$?

            run_performance_tests "stress-tests"
            local stress_exit=$?

            run_performance_tests "websocket-performance"
            local ws_exit=$?

            # Overall exit code
            if [ $bench_exit -ne 0 ] || [ $load_exit -ne 0 ] || [ $stress_exit -ne 0 ] || [ $ws_exit -ne 0 ]; then
                exit_code=1
            fi
            ;;

        *)
            log_error "Unknown performance test category: $category"
            return 1
            ;;
    esac

    # Generate performance summary
    if [ -f "$report_file" ]; then
        log_info "Performance test report saved to: $report_file"

        # Copy to performance reports directory for web dashboard
        cp "$report_file" "$performance_reports_dir/"
    fi

    return $exit_code
}

# Test execution functions
run_tests() {
    local category="$1"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/test_report_${category}_${timestamp}.txt"
    local coverage_file="$COVERAGE_DIR/coverage_${category}_${timestamp}"
    local log_file="$LOG_DIR/test_${category}_${timestamp}.log"

    log_header "Running $category Tests"

    # Prepare Go test command
    local test_cmd="go test"
    local test_pattern=""

    # Set test pattern based on category
    case "$category" in
        "all")
            test_pattern="./"
            ;;
        "unit")
            test_pattern="./ -run 'Test(User|Post|Comment|Database)'"
            ;;
        "integration")
            test_pattern="./ -run 'TestIntegration|TestEndToEnd'"
            ;;
        "auth")
            test_pattern="./ -run 'TestAuth|TestUser'"
            ;;
        "posts")
            test_pattern="./ -run 'TestPost|TestComment'"
            ;;
        "messaging")
            test_pattern="./ -run 'TestMessaging|TestConversation|TestWebSocket'"
            ;;
        "database")
            test_pattern="./ -run 'TestDatabase|TestRepository'"
            ;;
        "middleware")
            test_pattern="./ -run 'TestMiddleware|TestSecurity|TestAuth'"
            ;;
        "api")
            test_pattern="./ -run 'TestAPI|TestHTTP'"
            ;;
        "websocket")
            test_pattern="./ -run 'TestWebSocket|TestTyping|TestOnline'"
            ;;
        "frontend"|"frontend-dom"|"frontend-websocket"|"frontend-auth"|"frontend-spa")
            # Frontend tests use npm
            test_pattern="frontend"
            ;;
        "e2e"|"e2e-auth"|"e2e-messaging"|"cross-browser"|"responsive")
            # E2E tests use Playwright
            test_pattern="e2e"
            ;;
        *)
            log_error "Unknown test category: $category"
            return 1
            ;;
    esac

    # Add test options
    test_cmd="$test_cmd $test_pattern"

    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi

    if [ "$PARALLEL_TESTS" = true ]; then
        test_cmd="$test_cmd -parallel 4"
    fi

    test_cmd="$test_cmd -timeout $TEST_TIMEOUT"

    # Add coverage if enabled
    if [ "$RUN_COVERAGE" = true ]; then
        test_cmd="$test_cmd -coverprofile=${coverage_file}.out -covermode=atomic"
    fi

    log_debug "Test command: $test_cmd"

    # Run tests
    log_info "Executing tests..."
    local start_time=$(date +%s)

    if [ "$QUIET" = true ]; then
        eval "$test_cmd" > "$report_file" 2>&1
        local exit_code=$?
    else
        eval "$test_cmd" 2>&1 | tee "$report_file"
        local exit_code=${PIPESTATUS[0]}
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Generate coverage report if enabled
    if [ "$RUN_COVERAGE" = true ] && [ -f "${coverage_file}.out" ]; then
        log_info "Generating coverage report..."
        go tool cover -func="${coverage_file}.out" > "${coverage_file}_summary.txt"

        if [ "$GENERATE_HTML" = true ]; then
            go tool cover -html="${coverage_file}.out" -o "${coverage_file}.html"
            log_info "HTML coverage report generated: ${coverage_file}.html"
        fi
    fi

    # Generate test summary
    generate_test_summary "$category" "$report_file" "$duration" "$exit_code"

    return $exit_code
}

# Generate test summary
generate_test_summary() {
    local category="$1"
    local report_file="$2"
    local duration="$3"
    local exit_code="$4"

    log_header "Test Summary: $category"

    # Parse test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    local skipped_tests=0

    if [ -f "$report_file" ]; then
        # Count test results
        total_tests=$(grep -c "=== RUN" "$report_file" 2>/dev/null || echo "0")
        passed_tests=$(grep -c "--- PASS:" "$report_file" 2>/dev/null || echo "0")
        failed_tests=$(grep -c "--- FAIL:" "$report_file" 2>/dev/null || echo "0")
        skipped_tests=$(grep -c "--- SKIP:" "$report_file" 2>/dev/null || echo "0")
    fi

    # Ensure variables are numeric (sanitize)
    total_tests=${total_tests:-0}
    passed_tests=${passed_tests:-0}
    failed_tests=${failed_tests:-0}
    skipped_tests=${skipped_tests:-0}

    # Remove any non-numeric characters
    total_tests=$(echo "$total_tests" | tr -cd '0-9' | head -c 10)
    passed_tests=$(echo "$passed_tests" | tr -cd '0-9' | head -c 10)
    failed_tests=$(echo "$failed_tests" | tr -cd '0-9' | head -c 10)
    skipped_tests=$(echo "$skipped_tests" | tr -cd '0-9' | head -c 10)

    # Set to 0 if empty
    total_tests=${total_tests:-0}
    passed_tests=${passed_tests:-0}
    failed_tests=${failed_tests:-0}
    skipped_tests=${skipped_tests:-0}

    # Display summary
    echo -e "${CYAN}Category:${NC} $category"
    echo -e "${CYAN}Duration:${NC} ${duration}s"
    echo -e "${CYAN}Total Tests:${NC} $total_tests"
    echo -e "${GREEN}Passed:${NC} $passed_tests"
    echo -e "${RED}Failed:${NC} $failed_tests"
    echo -e "${YELLOW}Skipped:${NC} $skipped_tests"

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}Result: PASSED${NC}"
    else
        echo -e "${RED}Result: FAILED${NC}"

        # Show failed tests
        if [ -f "$report_file" ] && [ $failed_tests -gt 0 ]; then
            echo -e "\n${RED}Failed Tests:${NC}"
            grep -A 5 "--- FAIL:" "$report_file" | head -20
        fi
    fi

    # Show coverage if available
    if [ "$RUN_COVERAGE" = true ]; then
        local coverage_summary="${COVERAGE_DIR}/coverage_${category}_$(date +"%Y%m%d")_summary.txt"
        if [ -f "$coverage_summary" ]; then
            local total_coverage=$(tail -1 "$coverage_summary" | awk '{print $3}' | sed 's/%//')
            echo -e "${CYAN}Coverage:${NC} ${total_coverage}%"
        fi
    fi

    echo -e "${CYAN}Report:${NC} $report_file"
    echo ""
}

# Generate frontend test summary
generate_frontend_test_summary() {
    local category="$1"
    local report_file="$2"
    local duration="$3"
    local exit_code="$4"

    log_header "Frontend Test Summary: $category"

    # Parse Jest test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    local skipped_tests=0

    if [ -f "$report_file" ]; then
        # Count Jest test results
        total_tests=$(grep -c "✓\|✗\|○" "$report_file" 2>/dev/null || echo "0")
        passed_tests=$(grep -c "✓" "$report_file" 2>/dev/null || echo "0")
        failed_tests=$(grep -c "✗" "$report_file" 2>/dev/null || echo "0")
        skipped_tests=$(grep -c "○" "$report_file" 2>/dev/null || echo "0")
    fi

    # Ensure variables are numeric (sanitize)
    total_tests=${total_tests:-0}
    passed_tests=${passed_tests:-0}
    failed_tests=${failed_tests:-0}
    skipped_tests=${skipped_tests:-0}

    # Remove any non-numeric characters
    total_tests=$(echo "$total_tests" | tr -cd '0-9' | head -c 10)
    passed_tests=$(echo "$passed_tests" | tr -cd '0-9' | head -c 10)
    failed_tests=$(echo "$failed_tests" | tr -cd '0-9' | head -c 10)
    skipped_tests=$(echo "$skipped_tests" | tr -cd '0-9' | head -c 10)

    # Set to 0 if empty
    total_tests=${total_tests:-0}
    passed_tests=${passed_tests:-0}
    failed_tests=${failed_tests:-0}
    skipped_tests=${skipped_tests:-0}

    # Display summary
    echo -e "${CYAN}Category:${NC} $category"
    echo -e "${CYAN}Duration:${NC} ${duration}s"
    echo -e "${CYAN}Total Tests:${NC} $total_tests"
    echo -e "${GREEN}Passed:${NC} $passed_tests"
    echo -e "${RED}Failed:${NC} $failed_tests"
    echo -e "${YELLOW}Skipped:${NC} $skipped_tests"

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}Result: PASSED${NC}"
    else
        echo -e "${RED}Result: FAILED${NC}"

        # Show failed tests
        if [ -f "$report_file" ] && [ $failed_tests -gt 0 ]; then
            echo -e "\n${RED}Failed Tests:${NC}"
            grep -A 3 "✗" "$report_file" | head -20
        fi
    fi

    echo -e "${CYAN}Report:${NC} $report_file"
    echo ""
}

# Generate E2E test summary
generate_e2e_test_summary() {
    local category="$1"
    local report_file="$2"
    local duration="$3"
    local exit_code="$4"

    log_header "E2E Test Summary: $category"

    # Parse Playwright test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    local skipped_tests=0

    if [ -f "$report_file" ]; then
        # Count Playwright test results
        total_tests=$(grep -c "✓\|✗\|○\|passed\|failed\|skipped" "$report_file" 2>/dev/null || echo "0")
        passed_tests=$(grep -c "✓\|passed" "$report_file" 2>/dev/null || echo "0")
        failed_tests=$(grep -c "✗\|failed" "$report_file" 2>/dev/null || echo "0")
        skipped_tests=$(grep -c "○\|skipped" "$report_file" 2>/dev/null || echo "0")
    fi

    # Ensure variables are numeric (sanitize)
    total_tests=${total_tests:-0}
    passed_tests=${passed_tests:-0}
    failed_tests=${failed_tests:-0}
    skipped_tests=${skipped_tests:-0}

    # Remove any non-numeric characters
    total_tests=$(echo "$total_tests" | tr -cd '0-9' | head -c 10)
    passed_tests=$(echo "$passed_tests" | tr -cd '0-9' | head -c 10)
    failed_tests=$(echo "$failed_tests" | tr -cd '0-9' | head -c 10)
    skipped_tests=$(echo "$skipped_tests" | tr -cd '0-9' | head -c 10)

    # Set to 0 if empty
    total_tests=${total_tests:-0}
    passed_tests=${passed_tests:-0}
    failed_tests=${failed_tests:-0}
    skipped_tests=${skipped_tests:-0}

    # Display summary
    echo -e "${CYAN}Category:${NC} $category"
    echo -e "${CYAN}Duration:${NC} ${duration}s"
    echo -e "${CYAN}Total Tests:${NC} $total_tests"
    echo -e "${GREEN}Passed:${NC} $passed_tests"
    echo -e "${RED}Failed:${NC} $failed_tests"
    echo -e "${YELLOW}Skipped:${NC} $skipped_tests"

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}Result: PASSED${NC}"
    else
        echo -e "${RED}Result: FAILED${NC}"

        # Show failed tests
        if [ -f "$report_file" ] && [ $failed_tests -gt 0 ]; then
            echo -e "\n${RED}Failed Tests:${NC}"
            grep -A 5 "failed\|✗" "$report_file" | head -20
        fi
    fi

    echo -e "${CYAN}Report:${NC} $report_file"
    echo ""
}

# Run specific test category
execute_test_category() {
    local category="$1"

    case "$category" in
        "interactive")
            show_interactive_menu
            execute_test_category "$TEST_CATEGORY"
            ;;
        "all"|"unit"|"integration"|"auth"|"posts"|"messaging"|"database"|"middleware"|"api"|"websocket")
            run_tests "$category"
            local exit_code=$?

            if [ $exit_code -eq 0 ]; then
                log_info "All tests in category '$category' passed!"
            else
                log_error "Some tests in category '$category' failed!"
            fi

            return $exit_code
            ;;
        "frontend"|"frontend-dom"|"frontend-websocket"|"frontend-auth"|"frontend-spa")
            run_frontend_tests "$category"
            local exit_code=$?

            if [ $exit_code -eq 0 ]; then
                log_info "All frontend tests in category '$category' passed!"
            else
                log_error "Some frontend tests in category '$category' failed!"
            fi

            return $exit_code
            ;;
        "e2e"|"e2e-auth"|"e2e-messaging"|"cross-browser"|"responsive")
            run_e2e_tests "$category"
            local exit_code=$?

            if [ $exit_code -eq 0 ]; then
                log_info "All E2E tests in category '$category' passed!"
            else
                log_error "Some E2E tests in category '$category' failed!"
            fi

            return $exit_code
            ;;
        "benchmarks"|"load-tests"|"stress-tests"|"websocket-performance"|"performance-all")
            run_performance_tests "$category"
            local exit_code=$?

            if [ $exit_code -eq 0 ]; then
                log_info "All performance tests in category '$category' passed!"
            else
                log_error "Some performance tests in category '$category' failed!"
            fi

            return $exit_code
            ;;
        *)
            log_error "Unknown test category: $category"
            show_help
            return 1
            ;;
    esac
}

# Cleanup function
cleanup() {
    log_debug "Performing cleanup..."

    # Clean up temporary test databases
    find . -name "test_*.db" -delete 2>/dev/null || true

    # Clean up old log files (keep last 10)
    if [ -d "$LOG_DIR" ]; then
        ls -t "$LOG_DIR"/*.log 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
    fi

    # Clean up old reports (keep last 20)
    if [ -d "$REPORT_DIR" ]; then
        ls -t "$REPORT_DIR"/*.txt 2>/dev/null | tail -n +21 | xargs rm -f 2>/dev/null || true
    fi

    # Clean up old coverage files (keep last 10)
    if [ -d "$COVERAGE_DIR" ]; then
        ls -t "$COVERAGE_DIR"/*.out 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
        ls -t "$COVERAGE_DIR"/*.html 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
    fi
}

# Signal handlers
trap cleanup SIGINT SIGTERM EXIT

# Main execution
main() {
    # Show banner
    if [ "$QUIET" != true ]; then
        echo -e "${PURPLE}"
        echo "╔══════════════════════════════════════════════════════════════╗"
        echo "║                Real-Time Forum Test Runner                  ║"
        echo "║                    Comprehensive Testing Suite              ║"
        echo "╚══════════════════════════════════════════════════════════════╝"
        echo -e "${NC}"
    fi

    # Parse command line arguments
    parse_arguments "$@"

    # Setup test environment
    setup_test_environment

    # Execute tests
    execute_test_category "$TEST_CATEGORY"
    local exit_code=$?

    # Final cleanup
    cleanup

    # Exit with test result code
    exit $exit_code
}

# Run main function with all arguments
main "$@"
