#!/bin/bash

# Real-Time Forum Advanced Test Runner
# Professional-grade testing suite with comprehensive features
# Version: 2.0.0

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$SCRIPT_DIR/test-config.json"
TEST_DIR="."
REPORT_DIR="test-reports"
COVERAGE_DIR="coverage"
LOG_DIR="../logs"
TEMP_DIR="temp"

# Default settings
VERBOSE=false
QUIET=false
GENERATE_HTML=false
RUN_COVERAGE=false
PARALLEL_TESTS=false
RACE_DETECTION=false
BENCHMARK_MODE=false
WATCH_MODE=false
CI_MODE=false
TEST_TIMEOUT="10m"
WORKERS=4

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
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

log_success() {
    if [ "$QUIET" != true ]; then
        echo -e "${GREEN}[SUCCESS]${NC} $1"
    fi
}

log_header() {
    if [ "$QUIET" != true ]; then
        echo -e "${BOLD}${PURPLE}========================================${NC}"
        echo -e "${BOLD}${PURPLE}$1${NC}"
        echo -e "${BOLD}${PURPLE}========================================${NC}"
    fi
}

# Configuration management
load_config() {
    if [ -f "$CONFIG_FILE" ]; then
        log_debug "Loading configuration from $CONFIG_FILE"
        # Parse JSON config (requires jq for advanced parsing)
        if command -v jq &> /dev/null; then
            TEST_DIR=$(jq -r '.directories.tests // "tests"' "$CONFIG_FILE")
            REPORT_DIR=$(jq -r '.directories.reports // "test-reports"' "$CONFIG_FILE")
            COVERAGE_DIR=$(jq -r '.directories.coverage // "coverage"' "$CONFIG_FILE")
            LOG_DIR=$(jq -r '.directories.logs // "logs"' "$CONFIG_FILE")
            TEMP_DIR=$(jq -r '.directories.temp // "temp"' "$CONFIG_FILE")
            WORKERS=$(jq -r '.performance.parallel_workers // 4' "$CONFIG_FILE")
        else
            log_warn "jq not found, using default configuration"
        fi
    else
        log_debug "Configuration file not found, using defaults"
    fi
}

# Help function
show_help() {
    cat << EOF
${CYAN}${BOLD}Real-Time Forum Advanced Test Runner v2.0.0${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS] [TEST_CATEGORY]

${YELLOW}TEST CATEGORIES:${NC}
    all                 Run all tests (default)
    unit                Run unit tests only
    integration         Run integration tests only
    auth                Run authentication tests
    posts               Run post and comment tests
    messaging           Run messaging tests
    websocket           Run WebSocket tests
    database            Run database tests
    middleware          Run middleware and security tests
    api                 Run API endpoint tests

${YELLOW}OPTIONS:${NC}
    ${BOLD}Basic Options:${NC}
    -v, --verbose       Enable verbose output
    -q, --quiet         Disable output (except errors)
    -h, --html          Generate HTML test report
    -c, --coverage      Run with coverage analysis
    -p, --parallel      Run tests in parallel
    -t, --timeout TIME  Set test timeout (default: ${TEST_TIMEOUT})
    
    ${BOLD}Advanced Options:${NC}
    -r, --race          Enable race condition detection
    -b, --benchmark     Run benchmark tests
    -w, --watch         Watch mode (re-run tests on file changes)
    --ci                CI/CD mode (machine-readable output)
    --workers N         Number of parallel workers (default: ${WORKERS})
    --config FILE       Use custom configuration file
    
    ${BOLD}Reporting Options:${NC}
    --junit             Generate JUnit XML reports
    --json              Generate JSON reports
    --artifacts         Save test artifacts
    --no-cleanup        Don't clean up temporary files
    
    ${BOLD}Development Options:${NC}
    --profile           Enable CPU/memory profiling
    --debug             Enable debug mode
    --dry-run           Show what would be executed without running
    --list              List available tests
    
    ${BOLD}Help:${NC}
    --help              Show this help message
    --version           Show version information

${YELLOW}EXAMPLES:${NC}
    $0                                    # Interactive mode
    $0 all --coverage --html --parallel   # Full test suite with reports
    $0 auth --verbose --race              # Auth tests with race detection
    $0 unit --watch                       # Unit tests in watch mode
    $0 integration --ci --junit           # Integration tests for CI/CD
    $0 --list                            # List all available tests

${YELLOW}CONFIGURATION:${NC}
    Configuration file: ${CONFIG_FILE}
    Test directory: ${TEST_DIR}/
    Reports directory: ${REPORT_DIR}/
    Coverage directory: ${COVERAGE_DIR}/
    Logs directory: ${LOG_DIR}/

${YELLOW}REQUIREMENTS:${NC}
    - Go 1.19 or later
    - SQLite3
    - jq (optional, for advanced configuration)
    - All project dependencies installed

${YELLOW}CI/CD INTEGRATION:${NC}
    The test runner supports GitHub Actions, GitLab CI, Jenkins, and other
    CI/CD platforms with machine-readable output formats.

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Version information
show_version() {
    cat << EOF
${CYAN}${BOLD}Real-Time Forum Test Runner${NC}
Version: 2.0.0
Go Version: $(go version 2>/dev/null | awk '{print $3}' || echo "Not installed")
Platform: $(uname -s)/$(uname -m)
Script Location: $SCRIPT_DIR

${YELLOW}Features:${NC}
- Comprehensive test categorization
- Coverage analysis with HTML reports
- Parallel test execution
- Race condition detection
- CI/CD integration
- Watch mode for development
- Benchmark testing
- Advanced reporting formats

${YELLOW}Configuration:${NC}
- Config file: $([ -f "$CONFIG_FILE" ] && echo "Found" || echo "Not found")
- Test directory: $TEST_DIR
- Reports directory: $REPORT_DIR
EOF
}

# Parse command line arguments
parse_arguments() {
    TEST_CATEGORY=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            # Test categories
            all|unit|integration|auth|posts|messaging|websocket|database|middleware|api)
                TEST_CATEGORY="$1"
                shift
                ;;
            # Basic options
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
            # Advanced options
            -r|--race)
                RACE_DETECTION=true
                shift
                ;;
            -b|--benchmark)
                BENCHMARK_MODE=true
                shift
                ;;
            -w|--watch)
                WATCH_MODE=true
                shift
                ;;
            --ci)
                CI_MODE=true
                QUIET=true
                shift
                ;;
            --workers)
                WORKERS="$2"
                shift 2
                ;;
            --config)
                CONFIG_FILE="$2"
                shift 2
                ;;
            # Reporting options
            --junit)
                JUNIT_XML=true
                shift
                ;;
            --json)
                JSON_REPORT=true
                shift
                ;;
            --artifacts)
                SAVE_ARTIFACTS=true
                shift
                ;;
            --no-cleanup)
                NO_CLEANUP=true
                shift
                ;;
            # Development options
            --profile)
                ENABLE_PROFILING=true
                shift
                ;;
            --debug)
                VERBOSE=true
                DEBUG_MODE=true
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --list)
                LIST_TESTS=true
                shift
                ;;
            # Help options
            --help)
                show_help
                exit 0
                ;;
            --version)
                show_version
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
    if [ -z "$TEST_CATEGORY" ] && [ "$LIST_TESTS" != true ]; then
        if [ "$CI_MODE" = true ]; then
            TEST_CATEGORY="all"
        else
            TEST_CATEGORY="interactive"
        fi
    fi
}

# Environment setup
setup_environment() {
    log_debug "Setting up test environment..."
    
    # Load configuration
    load_config
    
    # Create necessary directories
    mkdir -p "$REPORT_DIR" "$COVERAGE_DIR" "$LOG_DIR" "$TEMP_DIR"
    
    # Check prerequisites
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run tests."
        exit 1
    fi
    
    # Check test directory
    if [ ! -d "$TEST_DIR" ]; then
        log_error "Test directory '$TEST_DIR' not found."
        exit 1
    fi
    
    # Set Go environment variables
    export GORACE="halt_on_error=1"
    export CGO_ENABLED=1  # Required for race detection
    
    log_debug "Environment setup completed"
}

# List available tests
list_tests() {
    log_header "Available Tests"
    
    echo -e "${YELLOW}Test Categories:${NC}"
    if command -v jq &> /dev/null && [ -f "$CONFIG_FILE" ]; then
        jq -r '.test_categories | to_entries[] | "  \(.key): \(.value.name) - \(.value.description)"' "$CONFIG_FILE"
    else
        echo "  all: All Tests - Run the complete test suite"
        echo "  unit: Unit Tests - Individual component testing"
        echo "  integration: Integration Tests - End-to-end workflow testing"
        echo "  auth: Authentication Tests - User authentication and sessions"
        echo "  posts: Post & Comment Tests - Content creation and management"
        echo "  messaging: Messaging Tests - Real-time messaging features"
        echo "  websocket: WebSocket Tests - Real-time communication"
        echo "  database: Database Tests - Database operations and integrity"
        echo "  middleware: Middleware Tests - Security and validation"
        echo "  api: API Tests - HTTP endpoint testing"
    fi
    
    echo -e "\n${YELLOW}Available Test Files:${NC}"
    find "$TEST_DIR" -name "*_test.go" | sort | while read -r file; do
        echo "  $file"
    done
    
    echo -e "\n${YELLOW}Test Functions:${NC}"
    find "$TEST_DIR" -name "*_test.go" -exec grep -h "^func Test" {} \; | sort | head -20 | while read -r func; do
        echo "  $func"
    done
    
    if [ $(find "$TEST_DIR" -name "*_test.go" -exec grep -h "^func Test" {} \; | wc -l) -gt 20 ]; then
        echo "  ... and $(( $(find "$TEST_DIR" -name "*_test.go" -exec grep -h "^func Test" {} \; | wc -l) - 20 )) more"
    fi
}

# Advanced test execution
run_advanced_tests() {
    local category="$1"
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/test_report_${category}_${timestamp}.txt"
    local coverage_file="$COVERAGE_DIR/coverage_${category}_${timestamp}"
    local log_file="$LOG_DIR/test_${category}_${timestamp}.log"
    local junit_file="$REPORT_DIR/junit_${category}_${timestamp}.xml"
    local json_file="$REPORT_DIR/test_${category}_${timestamp}.json"

    log_header "Running $category Tests (Advanced Mode)"

    # Get test pattern from config or use default
    local test_pattern=""
    if command -v jq &> /dev/null && [ -f "$CONFIG_FILE" ]; then
        test_pattern=$(jq -r ".test_categories.${category}.pattern // \"./\"" "$CONFIG_FILE")
        local config_timeout=$(jq -r ".test_categories.${category}.timeout // \"$TEST_TIMEOUT\"" "$CONFIG_FILE")
        local config_parallel=$(jq -r ".test_categories.${category}.parallel // false" "$CONFIG_FILE")

        if [ "$config_timeout" != "null" ]; then
            TEST_TIMEOUT="$config_timeout"
        fi

        if [ "$config_parallel" = "true" ] && [ "$PARALLEL_TESTS" != true ]; then
            PARALLEL_TESTS=true
            log_debug "Enabling parallel execution based on configuration"
        fi
    else
        case "$category" in
            "all") test_pattern="./" ;;
            "unit") test_pattern="./ -run 'Test(User|Post|Comment|Database)Repository'" ;;
            "integration") test_pattern="./ -run 'TestIntegration|TestEndToEnd'" ;;
            "auth") test_pattern="./ -run 'TestAuth|TestUser'" ;;
            "posts") test_pattern="./ -run 'TestPost|TestComment'" ;;
            "messaging") test_pattern="./ -run 'TestMessaging|TestConversation'" ;;
            "websocket") test_pattern="./ -run 'TestWebSocket|TestTyping|TestOnline'" ;;
            "database") test_pattern="./ -run 'TestDatabase|TestRepository'" ;;
            "middleware") test_pattern="./ -run 'TestMiddleware|TestSecurity'" ;;
            "api") test_pattern="./ -run 'TestAPI|TestHTTP'" ;;
            *)
                log_error "Unknown test category: $category"
                return 1
                ;;
        esac
    fi

    # Build test command
    local test_cmd="go test $test_pattern"

    # Add basic options
    if [ "$VERBOSE" = true ]; then
        test_cmd="$test_cmd -v"
    fi

    if [ "$PARALLEL_TESTS" = true ]; then
        test_cmd="$test_cmd -parallel $WORKERS"
    fi

    test_cmd="$test_cmd -timeout $TEST_TIMEOUT"

    # Add advanced options
    if [ "$RACE_DETECTION" = true ]; then
        test_cmd="$test_cmd -race"
        log_debug "Race detection enabled"
    fi

    if [ "$BENCHMARK_MODE" = true ]; then
        test_cmd="$test_cmd -bench=."
        log_debug "Benchmark mode enabled"
    fi

    # Add coverage
    if [ "$RUN_COVERAGE" = true ]; then
        test_cmd="$test_cmd -coverprofile=${coverage_file}.out -covermode=atomic"
        log_debug "Coverage analysis enabled"
    fi

    # Add profiling
    if [ "$ENABLE_PROFILING" = true ]; then
        test_cmd="$test_cmd -cpuprofile=${coverage_file}_cpu.prof -memprofile=${coverage_file}_mem.prof"
        log_debug "Profiling enabled"
    fi

    # Add JSON output for CI mode
    if [ "$CI_MODE" = true ] || [ "$JSON_REPORT" = true ]; then
        test_cmd="$test_cmd -json"
    fi

    log_debug "Test command: $test_cmd"

    # Dry run mode
    if [ "$DRY_RUN" = true ]; then
        log_info "DRY RUN - Would execute: $test_cmd"
        return 0
    fi

    # Execute tests
    log_info "Executing tests..."
    local start_time=$(date +%s)

    if [ "$CI_MODE" = true ]; then
        # CI mode with structured output
        eval "$test_cmd" > "$json_file" 2>&1
        local exit_code=$?

        # Convert JSON to human-readable format for logging
        if command -v jq &> /dev/null; then
            jq -r 'select(.Action == "pass" or .Action == "fail") | "\(.Time) \(.Action | ascii_upcase) \(.Package)/\(.Test // "PACKAGE")"' "$json_file" > "$report_file" 2>/dev/null || true
        else
            cp "$json_file" "$report_file"
        fi
    elif [ "$QUIET" = true ]; then
        eval "$test_cmd" > "$report_file" 2>&1
        local exit_code=$?
    else
        eval "$test_cmd" 2>&1 | tee "$report_file"
        local exit_code=${PIPESTATUS[0]}
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Generate additional reports
    generate_advanced_reports "$category" "$report_file" "$coverage_file" "$junit_file" "$json_file" "$duration" "$exit_code"

    return $exit_code
}

# Generate comprehensive reports
generate_advanced_reports() {
    local category="$1"
    local report_file="$2"
    local coverage_file="$3"
    local junit_file="$4"
    local json_file="$5"
    local duration="$6"
    local exit_code="$7"

    log_debug "Generating advanced reports..."

    # Generate coverage reports
    if [ "$RUN_COVERAGE" = true ] && [ -f "${coverage_file}.out" ]; then
        log_info "Generating coverage reports..."

        # Text coverage summary
        go tool cover -func="${coverage_file}.out" > "${coverage_file}_summary.txt"

        # HTML coverage report
        if [ "$GENERATE_HTML" = true ]; then
            go tool cover -html="${coverage_file}.out" -o "${coverage_file}.html"
            log_info "HTML coverage report: ${coverage_file}.html"
        fi

        # Coverage badge data
        local total_coverage=$(go tool cover -func="${coverage_file}.out" | tail -1 | awk '{print $3}' | sed 's/%//')
        echo "{\"coverage\": $total_coverage, \"timestamp\": \"$(date -Iseconds)\"}" > "${coverage_file}_badge.json"
    fi

    # Generate JUnit XML report
    if [ "$JUNIT_XML" = true ]; then
        generate_junit_report "$report_file" "$junit_file" "$category" "$duration" "$exit_code"
    fi

    # Generate test summary
    generate_comprehensive_summary "$category" "$report_file" "$coverage_file" "$duration" "$exit_code"

    # Save artifacts
    if [ "$SAVE_ARTIFACTS" = true ]; then
        local artifact_dir="$REPORT_DIR/artifacts_${category}_$(date +%Y%m%d_%H%M%S)"
        mkdir -p "$artifact_dir"

        # Copy all generated files
        cp "$report_file" "$artifact_dir/" 2>/dev/null || true
        cp "${coverage_file}"* "$artifact_dir/" 2>/dev/null || true
        cp "$junit_file" "$artifact_dir/" 2>/dev/null || true
        cp "$json_file" "$artifact_dir/" 2>/dev/null || true

        log_info "Test artifacts saved to: $artifact_dir"
    fi
}

# Generate JUnit XML report
generate_junit_report() {
    local report_file="$1"
    local junit_file="$2"
    local category="$3"
    local duration="$4"
    local exit_code="$5"

    if [ ! -f "$report_file" ]; then
        return
    fi

    log_debug "Generating JUnit XML report..."

    # Parse test results
    local total_tests=$(grep -c "=== RUN" "$report_file" 2>/dev/null || echo "0")
    local passed_tests=$(grep -c "--- PASS:" "$report_file" 2>/dev/null || echo "0")
    local failed_tests=$(grep -c "--- FAIL:" "$report_file" 2>/dev/null || echo "0")
    local skipped_tests=$(grep -c "--- SKIP:" "$report_file" 2>/dev/null || echo "0")

    # Generate XML
    cat > "$junit_file" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="$category" tests="$total_tests" failures="$failed_tests" skipped="$skipped_tests" time="$duration">
EOF

    # Add individual test cases
    if [ -f "$report_file" ]; then
        grep -E "(=== RUN|--- PASS:|--- FAIL:|--- SKIP:)" "$report_file" | while read -r line; do
            if [[ "$line" =~ "=== RUN" ]]; then
                test_name=$(echo "$line" | awk '{print $3}')
                echo "  <testcase name=\"$test_name\" classname=\"$category\">" >> "$junit_file"
            elif [[ "$line" =~ "--- FAIL:" ]]; then
                echo "    <failure message=\"Test failed\">Test failed</failure>" >> "$junit_file"
                echo "  </testcase>" >> "$junit_file"
            elif [[ "$line" =~ "--- SKIP:" ]]; then
                echo "    <skipped/>" >> "$junit_file"
                echo "  </testcase>" >> "$junit_file"
            elif [[ "$line" =~ "--- PASS:" ]]; then
                echo "  </testcase>" >> "$junit_file"
            fi
        done
    fi

    echo "</testsuite>" >> "$junit_file"
    log_info "JUnit XML report generated: $junit_file"
}

# Generate comprehensive test summary
generate_comprehensive_summary() {
    local category="$1"
    local report_file="$2"
    local coverage_file="$3"
    local duration="$4"
    local exit_code="$5"

    log_header "Test Summary: $category"

    # Parse test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    local skipped_tests=0
    local benchmark_tests=0

    if [ -f "$report_file" ]; then
        total_tests=$(grep -c "=== RUN" "$report_file" 2>/dev/null || echo "0")
        passed_tests=$(grep -c "--- PASS:" "$report_file" 2>/dev/null || echo "0")
        failed_tests=$(grep -c "--- FAIL:" "$report_file" 2>/dev/null || echo "0")
        skipped_tests=$(grep -c "--- SKIP:" "$report_file" 2>/dev/null || echo "0")
        benchmark_tests=$(grep -c "Benchmark" "$report_file" 2>/dev/null || echo "0")
    fi

    # Calculate success rate
    local success_rate=0
    if [ $total_tests -gt 0 ]; then
        success_rate=$(( (passed_tests * 100) / total_tests ))
    fi

    # Display summary
    echo -e "${CYAN}Category:${NC} $category"
    echo -e "${CYAN}Duration:${NC} ${duration}s"
    echo -e "${CYAN}Total Tests:${NC} $total_tests"
    echo -e "${GREEN}Passed:${NC} $passed_tests"
    echo -e "${RED}Failed:${NC} $failed_tests"
    echo -e "${YELLOW}Skipped:${NC} $skipped_tests"

    if [ $benchmark_tests -gt 0 ]; then
        echo -e "${BLUE}Benchmarks:${NC} $benchmark_tests"
    fi

    echo -e "${CYAN}Success Rate:${NC} ${success_rate}%"

    # Overall result
    if [ $exit_code -eq 0 ]; then
        echo -e "${BOLD}${GREEN}Result: ✅ PASSED${NC}"
    else
        echo -e "${BOLD}${RED}Result: ❌ FAILED${NC}"

        # Show failed tests
        if [ -f "$report_file" ] && [ $failed_tests -gt 0 ]; then
            echo -e "\n${RED}Failed Tests:${NC}"
            grep -A 3 "--- FAIL:" "$report_file" | head -15 | while read -r line; do
                echo -e "  ${RED}$line${NC}"
            done

            if [ $failed_tests -gt 5 ]; then
                echo -e "  ${YELLOW}... and $((failed_tests - 5)) more failures${NC}"
            fi
        fi
    fi

    # Show coverage if available
    if [ "$RUN_COVERAGE" = true ] && [ -f "${coverage_file}_summary.txt" ]; then
        local total_coverage=$(tail -1 "${coverage_file}_summary.txt" | awk '{print $3}' | sed 's/%//')
        echo -e "${CYAN}Coverage:${NC} ${total_coverage}%"

        # Coverage status
        local coverage_threshold=80
        if command -v jq &> /dev/null && [ -f "$CONFIG_FILE" ]; then
            coverage_threshold=$(jq -r '.coverage.threshold // 80' "$CONFIG_FILE")
        fi

        if (( $(echo "$total_coverage >= $coverage_threshold" | bc -l 2>/dev/null || echo "0") )); then
            echo -e "${GREEN}Coverage Status: ✅ Above threshold (${coverage_threshold}%)${NC}"
        else
            echo -e "${YELLOW}Coverage Status: ⚠️  Below threshold (${coverage_threshold}%)${NC}"
        fi
    fi

    # Show performance metrics
    if [ "$BENCHMARK_MODE" = true ] && [ $benchmark_tests -gt 0 ]; then
        echo -e "\n${CYAN}Performance Metrics:${NC}"
        grep "Benchmark" "$report_file" | head -5 | while read -r line; do
            echo -e "  $line"
        done
    fi

    # Show race conditions if detected
    if [ "$RACE_DETECTION" = true ]; then
        local race_conditions=$(grep -c "WARNING: DATA RACE" "$report_file" 2>/dev/null || echo "0")
        if [ $race_conditions -gt 0 ]; then
            echo -e "${RED}Race Conditions Detected: $race_conditions${NC}"
        else
            echo -e "${GREEN}Race Conditions: None detected${NC}"
        fi
    fi

    echo -e "\n${CYAN}Reports Generated:${NC}"
    echo -e "  Test Report: $report_file"

    if [ "$RUN_COVERAGE" = true ] && [ -f "${coverage_file}.html" ]; then
        echo -e "  Coverage Report: ${coverage_file}.html"
    fi

    if [ "$JUNIT_XML" = true ]; then
        echo -e "  JUnit XML: ${report_file%.*}_junit.xml"
    fi

    echo ""
}

# Interactive menu system
show_interactive_menu() {
    clear
    log_header "Real-Time Forum Advanced Test Runner"

    echo -e "${CYAN}Test Environment:${NC}"
    echo -e "Configuration: ${YELLOW}$([ -f "$CONFIG_FILE" ] && echo "Loaded" || echo "Default")${NC}"
    echo -e "Test Directory: ${YELLOW}$TEST_DIR${NC}"
    echo -e "Coverage: ${YELLOW}$([ "$RUN_COVERAGE" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Parallel: ${YELLOW}$([ "$PARALLEL_TESTS" = true ] && echo "Enabled ($WORKERS workers)" || echo "Disabled")${NC}"
    echo -e "Race Detection: ${YELLOW}$([ "$RACE_DETECTION" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo ""

    echo -e "${YELLOW}Choose test category:${NC}"
    echo "1.  🔍 All Tests (comprehensive suite)"
    echo "2.  🧪 Unit Tests (component testing)"
    echo "3.  🔗 Integration Tests (end-to-end workflows)"
    echo "4.  🔐 Authentication Tests (login/session)"
    echo "5.  📝 Post & Comment Tests (content management)"
    echo "6.  💬 Messaging Tests (conversations)"
    echo "7.  🌐 WebSocket Tests (real-time features)"
    echo "8.  🗄️  Database Tests (data operations)"
    echo "9.  🛡️  Middleware Tests (security/validation)"
    echo "10. 🔌 API Tests (HTTP endpoints)"
    echo ""
    echo "11. ⚙️  Advanced Configuration"
    echo "12. 📊 View Test Reports"
    echo "13. 🔧 Development Tools"
    echo "14. ℹ️  System Information"
    echo ""
    echo "0.  🚪 Exit"
    echo ""

    read -p "Enter your choice [0-14]: " choice

    case $choice in
        1) TEST_CATEGORY="all" ;;
        2) TEST_CATEGORY="unit" ;;
        3) TEST_CATEGORY="integration" ;;
        4) TEST_CATEGORY="auth" ;;
        5) TEST_CATEGORY="posts" ;;
        6) TEST_CATEGORY="messaging" ;;
        7) TEST_CATEGORY="websocket" ;;
        8) TEST_CATEGORY="database" ;;
        9) TEST_CATEGORY="middleware" ;;
        10) TEST_CATEGORY="api" ;;
        11) show_advanced_configuration; return ;;
        12) show_reports_dashboard; return ;;
        13) show_development_tools; return ;;
        14) show_system_info; return ;;
        0) log_info "Goodbye! 👋"; exit 0 ;;
        *)
            log_error "Invalid choice. Please try again."
            sleep 2
            show_interactive_menu
            return
            ;;
    esac
}

# Advanced configuration menu
show_advanced_configuration() {
    clear
    log_header "Advanced Configuration"

    echo -e "${YELLOW}Current Settings:${NC}"
    echo -e "Verbose Output: ${CYAN}$([ "$VERBOSE" = true ] && echo "✅ Enabled" || echo "❌ Disabled")${NC}"
    echo -e "Coverage Analysis: ${CYAN}$([ "$RUN_COVERAGE" = true ] && echo "✅ Enabled" || echo "❌ Disabled")${NC}"
    echo -e "HTML Reports: ${CYAN}$([ "$GENERATE_HTML" = true ] && echo "✅ Enabled" || echo "❌ Disabled")${NC}"
    echo -e "Parallel Execution: ${CYAN}$([ "$PARALLEL_TESTS" = true ] && echo "✅ Enabled ($WORKERS workers)" || echo "❌ Disabled")${NC}"
    echo -e "Race Detection: ${CYAN}$([ "$RACE_DETECTION" = true ] && echo "✅ Enabled" || echo "❌ Disabled")${NC}"
    echo -e "Benchmark Mode: ${CYAN}$([ "$BENCHMARK_MODE" = true ] && echo "✅ Enabled" || echo "❌ Disabled")${NC}"
    echo -e "Watch Mode: ${CYAN}$([ "$WATCH_MODE" = true ] && echo "✅ Enabled" || echo "❌ Disabled")${NC}"
    echo -e "Test Timeout: ${CYAN}$TEST_TIMEOUT${NC}"
    echo ""

    echo -e "${YELLOW}Configuration Options:${NC}"
    echo "1. Toggle Verbose Output"
    echo "2. Toggle Coverage Analysis"
    echo "3. Toggle HTML Reports"
    echo "4. Toggle Parallel Execution"
    echo "5. Toggle Race Detection"
    echo "6. Toggle Benchmark Mode"
    echo "7. Toggle Watch Mode"
    echo "8. Set Test Timeout"
    echo "9. Set Parallel Workers"
    echo "10. Reset to Defaults"
    echo "11. Save Configuration"
    echo "12. Back to Main Menu"
    echo ""

    read -p "Enter your choice [1-12]: " choice

    case $choice in
        1)
            VERBOSE=$([ "$VERBOSE" = true ] && echo false || echo true)
            log_info "Verbose output $([ "$VERBOSE" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        2)
            RUN_COVERAGE=$([ "$RUN_COVERAGE" = true ] && echo false || echo true)
            log_info "Coverage analysis $([ "$RUN_COVERAGE" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        3)
            GENERATE_HTML=$([ "$GENERATE_HTML" = true ] && echo false || echo true)
            log_info "HTML reports $([ "$GENERATE_HTML" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        4)
            PARALLEL_TESTS=$([ "$PARALLEL_TESTS" = true ] && echo false || echo true)
            log_info "Parallel execution $([ "$PARALLEL_TESTS" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        5)
            RACE_DETECTION=$([ "$RACE_DETECTION" = true ] && echo false || echo true)
            log_info "Race detection $([ "$RACE_DETECTION" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        6)
            BENCHMARK_MODE=$([ "$BENCHMARK_MODE" = true ] && echo false || echo true)
            log_info "Benchmark mode $([ "$BENCHMARK_MODE" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        7)
            WATCH_MODE=$([ "$WATCH_MODE" = true ] && echo false || echo true)
            log_info "Watch mode $([ "$WATCH_MODE" = true ] && echo "enabled" || echo "disabled")"
            sleep 1
            show_advanced_configuration
            ;;
        8)
            read -p "Enter test timeout (e.g., 5m, 30s, 1h): " new_timeout
            if [[ "$new_timeout" =~ ^[0-9]+[smh]$ ]]; then
                TEST_TIMEOUT="$new_timeout"
                log_info "Test timeout set to: $TEST_TIMEOUT"
            else
                log_error "Invalid timeout format. Use: 5m, 30s, 1h"
            fi
            sleep 2
            show_advanced_configuration
            ;;
        9)
            read -p "Enter number of parallel workers (1-16): " new_workers
            if [[ "$new_workers" =~ ^[1-9]$|^1[0-6]$ ]]; then
                WORKERS="$new_workers"
                log_info "Parallel workers set to: $WORKERS"
            else
                log_error "Invalid number. Use 1-16"
            fi
            sleep 2
            show_advanced_configuration
            ;;
        10)
            # Reset to defaults
            VERBOSE=false
            RUN_COVERAGE=false
            GENERATE_HTML=false
            PARALLEL_TESTS=false
            RACE_DETECTION=false
            BENCHMARK_MODE=false
            WATCH_MODE=false
            TEST_TIMEOUT="10m"
            WORKERS=4
            log_info "Configuration reset to defaults"
            sleep 1
            show_advanced_configuration
            ;;
        11)
            save_configuration
            sleep 2
            show_advanced_configuration
            ;;
        12)
            show_interactive_menu
            ;;
        *)
            log_error "Invalid choice. Please try again."
            sleep 2
            show_advanced_configuration
            ;;
    esac
}

# Save configuration to file
save_configuration() {
    local config_backup="$CONFIG_FILE.backup.$(date +%Y%m%d_%H%M%S)"

    if [ -f "$CONFIG_FILE" ]; then
        cp "$CONFIG_FILE" "$config_backup"
        log_info "Configuration backed up to: $config_backup"
    fi

    # Create new configuration with current settings
    cat > "$CONFIG_FILE" << EOF
{
  "test_runner": {
    "version": "2.0.0",
    "last_updated": "$(date -Iseconds)"
  },
  "settings": {
    "verbose": $VERBOSE,
    "coverage": $RUN_COVERAGE,
    "html_reports": $GENERATE_HTML,
    "parallel": $PARALLEL_TESTS,
    "race_detection": $RACE_DETECTION,
    "benchmark_mode": $BENCHMARK_MODE,
    "watch_mode": $WATCH_MODE,
    "timeout": "$TEST_TIMEOUT",
    "workers": $WORKERS
  }
}
EOF

    log_info "Configuration saved to: $CONFIG_FILE"
}

# Reports dashboard
show_reports_dashboard() {
    clear
    log_header "Test Reports Dashboard"

    echo -e "${YELLOW}Recent Test Reports:${NC}"
    if [ -d "$REPORT_DIR" ] && [ "$(ls -A $REPORT_DIR/*.txt 2>/dev/null)" ]; then
        ls -lt "$REPORT_DIR"/*.txt 2>/dev/null | head -10 | while read -r line; do
            echo "  $line"
        done
    else
        echo -e "  ${YELLOW}No test reports found${NC}"
    fi

    echo -e "\n${YELLOW}Coverage Reports:${NC}"
    if [ -d "$COVERAGE_DIR" ] && [ "$(ls -A $COVERAGE_DIR/*.html 2>/dev/null)" ]; then
        ls -lt "$COVERAGE_DIR"/*.html 2>/dev/null | head -5 | while read -r line; do
            echo "  $line"
        done
    else
        echo -e "  ${YELLOW}No coverage reports found${NC}"
    fi

    echo -e "\n${YELLOW}Options:${NC}"
    echo "1. View Latest Test Report"
    echo "2. Open Latest Coverage Report"
    echo "3. Generate Coverage Summary"
    echo "4. Clean Old Reports"
    echo "5. Export Reports Archive"
    echo "6. Back to Main Menu"
    echo ""

    read -p "Enter your choice [1-6]: " choice

    case $choice in
        1)
            local latest_report=$(ls -t "$REPORT_DIR"/*.txt 2>/dev/null | head -1)
            if [ -n "$latest_report" ]; then
                log_info "Latest test report: $latest_report"
                echo -e "${CYAN}========================================${NC}"
                tail -50 "$latest_report"
                echo -e "${CYAN}========================================${NC}"
                read -p "Press Enter to continue..."
            else
                log_warn "No test reports found"
                sleep 2
            fi
            show_reports_dashboard
            ;;
        2)
            local latest_coverage=$(ls -t "$COVERAGE_DIR"/*.html 2>/dev/null | head -1)
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
            show_reports_dashboard
            ;;
        3)
            generate_coverage_summary
            read -p "Press Enter to continue..."
            show_reports_dashboard
            ;;
        4)
            clean_old_reports
            sleep 2
            show_reports_dashboard
            ;;
        5)
            export_reports_archive
            sleep 2
            show_reports_dashboard
            ;;
        6)
            show_interactive_menu
            ;;
        *)
            log_error "Invalid choice"
            sleep 1
            show_reports_dashboard
            ;;
    esac
}

# Development tools menu
show_development_tools() {
    clear
    log_header "Development Tools"

    echo -e "${YELLOW}Available Tools:${NC}"
    echo "1. 🔍 List All Tests"
    echo "2. 🧪 Run Single Test Function"
    echo "3. 📁 Run Tests in Specific File"
    echo "4. 👀 Watch Mode (auto-run on changes)"
    echo "5. 🏃 Benchmark Tests"
    echo "6. 🔍 Race Condition Detection"
    echo "7. 📊 Performance Profiling"
    echo "8. 🧹 Clean Test Environment"
    echo "9. Back to Main Menu"
    echo ""

    read -p "Enter your choice [1-9]: " choice

    case $choice in
        1)
            list_tests
            read -p "Press Enter to continue..."
            show_development_tools
            ;;
        2)
            run_single_test
            show_development_tools
            ;;
        3)
            run_file_tests
            show_development_tools
            ;;
        4)
            start_watch_mode
            show_development_tools
            ;;
        5)
            BENCHMARK_MODE=true
            TEST_CATEGORY="all"
            return
            ;;
        6)
            RACE_DETECTION=true
            TEST_CATEGORY="all"
            return
            ;;
        7)
            ENABLE_PROFILING=true
            TEST_CATEGORY="all"
            return
            ;;
        8)
            clean_test_environment
            sleep 2
            show_development_tools
            ;;
        9)
            show_interactive_menu
            ;;
        *)
            log_error "Invalid choice"
            sleep 1
            show_development_tools
            ;;
    esac
}

# System information
show_system_info() {
    clear
    log_header "System Information"

    echo -e "${YELLOW}Environment:${NC}"
    echo -e "Go Version: ${CYAN}$(go version 2>/dev/null | awk '{print $3}' || echo "Not installed")${NC}"
    echo -e "Platform: ${CYAN}$(uname -s)/$(uname -m)${NC}"
    echo -e "Shell: ${CYAN}$SHELL${NC}"
    echo -e "PWD: ${CYAN}$(pwd)${NC}"

    echo -e "\n${YELLOW}Test Runner:${NC}"
    echo -e "Version: ${CYAN}2.0.0${NC}"
    echo -e "Script: ${CYAN}$0${NC}"
    echo -e "Config: ${CYAN}$([ -f "$CONFIG_FILE" ] && echo "$CONFIG_FILE (found)" || echo "$CONFIG_FILE (not found)")${NC}"

    echo -e "\n${YELLOW}Directories:${NC}"
    echo -e "Tests: ${CYAN}$TEST_DIR $([ -d "$TEST_DIR" ] && echo "(✅)" || echo "(❌)")${NC}"
    echo -e "Reports: ${CYAN}$REPORT_DIR $([ -d "$REPORT_DIR" ] && echo "(✅)" || echo "(❌)")${NC}"
    echo -e "Coverage: ${CYAN}$COVERAGE_DIR $([ -d "$COVERAGE_DIR" ] && echo "(✅)" || echo "(❌)")${NC}"
    echo -e "Logs: ${CYAN}$LOG_DIR $([ -d "$LOG_DIR" ] && echo "(✅)" || echo "(❌)")${NC}"

    echo -e "\n${YELLOW}Dependencies:${NC}"
    echo -e "jq: ${CYAN}$(command -v jq &> /dev/null && echo "✅ Available" || echo "❌ Not found")${NC}"
    echo -e "bc: ${CYAN}$(command -v bc &> /dev/null && echo "✅ Available" || echo "❌ Not found")${NC}"

    echo -e "\n${YELLOW}Test Statistics:${NC}"
    if [ -d "$TEST_DIR" ]; then
        local test_files=$(find "$TEST_DIR" -name "*_test.go" | wc -l)
        local test_functions=$(find "$TEST_DIR" -name "*_test.go" -exec grep -h "^func Test" {} \; | wc -l)
        echo -e "Test Files: ${CYAN}$test_files${NC}"
        echo -e "Test Functions: ${CYAN}$test_functions${NC}"
    fi

    if [ -d "$REPORT_DIR" ]; then
        local report_count=$(ls "$REPORT_DIR"/*.txt 2>/dev/null | wc -l)
        echo -e "Reports: ${CYAN}$report_count${NC}"
    fi

    echo ""
    read -p "Press Enter to continue..."
    show_interactive_menu
}

# Utility functions
run_single_test() {
    echo -e "${YELLOW}Available test functions:${NC}"
    find "$TEST_DIR" -name "*_test.go" -exec grep -h "^func Test" {} \; | head -20 | nl
    echo ""
    read -p "Enter test function name: " test_func

    if [ -n "$test_func" ]; then
        log_info "Running single test: $test_func"
        go test ./tests/... -run "^${test_func}$" -v
        read -p "Press Enter to continue..."
    fi
}

run_file_tests() {
    echo -e "${YELLOW}Available test files:${NC}"
    find "$TEST_DIR" -name "*_test.go" | nl
    echo ""
    read -p "Enter test file path: " test_file

    if [ -n "$test_file" ] && [ -f "$test_file" ]; then
        log_info "Running tests in file: $test_file"
        go test "$test_file" -v
        read -p "Press Enter to continue..."
    fi
}

start_watch_mode() {
    log_info "Starting watch mode... (Press Ctrl+C to stop)"
    log_info "Watching for changes in $TEST_DIR and source files"

    # Simple watch implementation
    while true; do
        if command -v inotifywait &> /dev/null; then
            inotifywait -r -e modify,create,delete "$TEST_DIR" . --exclude '\.git|test-reports|coverage|logs' 2>/dev/null
        else
            sleep 5
        fi

        log_info "Changes detected, running tests..."
        run_advanced_tests "all"
        echo ""
        log_info "Waiting for changes... (Press Ctrl+C to stop)"
    done
}

clean_test_environment() {
    log_info "Cleaning test environment..."

    # Remove temporary files
    find . -name "test_*.db" -delete 2>/dev/null || true
    find . -name "*.test" -delete 2>/dev/null || true
    find . -name "*.prof" -delete 2>/dev/null || true

    # Clean old reports (keep last 10)
    if [ -d "$REPORT_DIR" ]; then
        ls -t "$REPORT_DIR"/* 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
    fi

    # Clean old coverage files (keep last 5)
    if [ -d "$COVERAGE_DIR" ]; then
        ls -t "$COVERAGE_DIR"/* 2>/dev/null | tail -n +6 | xargs rm -f 2>/dev/null || true
    fi

    log_success "Test environment cleaned"
}

generate_coverage_summary() {
    log_info "Generating coverage summary..."

    local latest_coverage=$(ls -t "$COVERAGE_DIR"/*_summary.txt 2>/dev/null | head -1)
    if [ -n "$latest_coverage" ]; then
        echo -e "${CYAN}Latest Coverage Summary:${NC}"
        cat "$latest_coverage"
    else
        log_warn "No coverage data found. Run tests with --coverage first."
    fi
}

clean_old_reports() {
    read -p "Clean reports older than how many days? (default: 7): " days
    days=${days:-7}

    log_info "Cleaning reports older than $days days..."

    find "$REPORT_DIR" -name "*.txt" -mtime +$days -delete 2>/dev/null || true
    find "$COVERAGE_DIR" -name "*" -mtime +$days -delete 2>/dev/null || true
    find "$LOG_DIR" -name "*.log" -mtime +$days -delete 2>/dev/null || true

    log_success "Old reports cleaned"
}

export_reports_archive() {
    local archive_name="test-reports-$(date +%Y%m%d_%H%M%S).tar.gz"

    log_info "Creating reports archive: $archive_name"

    tar -czf "$archive_name" "$REPORT_DIR" "$COVERAGE_DIR" "$LOG_DIR" 2>/dev/null || {
        log_error "Failed to create archive"
        return 1
    }

    log_success "Reports archive created: $archive_name"
}

# Execute test category
execute_test_category() {
    local category="$1"

    case "$category" in
        "interactive")
            show_interactive_menu
            execute_test_category
            ;;
        "all"|"unit"|"integration"|"auth"|"posts"|"messaging"|"websocket"|"database"|"middleware"|"api")
            run_advanced_tests "$category"
            local exit_code=$?

            if [ $exit_code -eq 0 ]; then
                log_success "All tests in category '$category' passed! ✅"
            else
                log_error "Some tests in category '$category' failed! ❌"
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

    # Clean up temporary files if not in no-cleanup mode
    if [ "$NO_CLEANUP" != true ]; then
        find . -name "test_*.db" -delete 2>/dev/null || true
        find . -name "*.test" -delete 2>/dev/null || true

        # Clean up temp directory
        if [ -d "$TEMP_DIR" ]; then
            rm -rf "$TEMP_DIR"/* 2>/dev/null || true
        fi
    fi

    # Kill any background processes
    jobs -p | xargs kill 2>/dev/null || true
}

# Signal handlers
trap cleanup SIGINT SIGTERM EXIT

# Main execution function
main() {
    # Show banner (unless in CI mode)
    if [ "$CI_MODE" != true ] && [ "$QUIET" != true ]; then
        echo -e "${PURPLE}${BOLD}"
        echo "╔══════════════════════════════════════════════════════════════╗"
        echo "║            Real-Time Forum Advanced Test Runner              ║"
        echo "║                Professional Testing Suite v2.0               ║"
        echo "╚══════════════════════════════════════════════════════════════╝"
        echo -e "${NC}"
    fi

    # Parse command line arguments
    parse_arguments "$@"

    # Handle special modes
    if [ "$LIST_TESTS" = true ]; then
        setup_environment
        list_tests
        exit 0
    fi

    # Setup environment
    setup_environment

    # Execute tests
    if [ "$WATCH_MODE" = true ]; then
        start_watch_mode
    else
        execute_test_category "$TEST_CATEGORY"
    fi

    local exit_code=$?

    # Final cleanup
    cleanup

    # Exit with test result code
    exit $exit_code
}

# Run main function with all arguments
main "$@"
