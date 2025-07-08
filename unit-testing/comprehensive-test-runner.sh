#!/bin/bash

# Comprehensive Test Runner
# Runs all test categories with detailed reporting and coverage analysis
# Version: 1.0.0

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
REPORT_DIR="$SCRIPT_DIR/test-reports"
COVERAGE_DIR="$SCRIPT_DIR/coverage"
HTML_REPORT_DIR="$SCRIPT_DIR/html-reports"
LOG_DIR="$PROJECT_ROOT/logs"

# Test configuration
TIMEOUT="15m"
PARALLEL_WORKERS=4
COVERAGE_THRESHOLD=75
VERBOSE=false
GENERATE_HTML=true
RUN_E2E=true
RUN_FRONTEND=true

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
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

log_header() {
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_failure() {
    echo -e "${RED}❌ $1${NC}"
}

# Help function
show_help() {
    cat << EOF
${CYAN}Comprehensive Test Runner${NC}
${CYAN}Version: 1.0.0${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS]

${YELLOW}OPTIONS:${NC}
    --timeout DURATION      Set test timeout (default: ${TIMEOUT})
    --workers N             Number of parallel workers (default: ${PARALLEL_WORKERS})
    --coverage-threshold N  Minimum coverage percentage (default: ${COVERAGE_THRESHOLD})
    --no-html              Skip HTML report generation
    --no-e2e               Skip E2E tests
    --no-frontend          Skip frontend tests
    --verbose              Enable verbose output
    --help                 Show this help message

${YELLOW}TEST CATEGORIES:${NC}
    - Unit Tests (Go)
    - Integration Tests (Go)
    - E2E Scenarios (Go)
    - Frontend Unit Tests (JavaScript)
    - Frontend E2E Tests (Playwright)
    - Performance Tests
    - Accessibility Tests

${YELLOW}REPORTS:${NC}
    - Test results: ${REPORT_DIR}/
    - Coverage data: ${COVERAGE_DIR}/
    - HTML reports: ${HTML_REPORT_DIR}/

${YELLOW}EXAMPLES:${NC}
    $0                              # Run all tests with default settings
    $0 --no-e2e --verbose          # Skip E2E tests with verbose output
    $0 --coverage-threshold 80     # Require 80% coverage
    $0 --workers 8 --timeout 20m   # Use 8 workers with 20 minute timeout

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --workers)
                PARALLEL_WORKERS="$2"
                shift 2
                ;;
            --coverage-threshold)
                COVERAGE_THRESHOLD="$2"
                shift 2
                ;;
            --no-html)
                GENERATE_HTML=false
                shift
                ;;
            --no-e2e)
                RUN_E2E=false
                shift
                ;;
            --no-frontend)
                RUN_FRONTEND=false
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
    log_debug "Setting up test environment..."
    
    # Create necessary directories
    mkdir -p "$REPORT_DIR" "$COVERAGE_DIR" "$HTML_REPORT_DIR" "$LOG_DIR"
    
    # Check prerequisites
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run tests."
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found. Please run from the project root directory."
        exit 1
    fi
    
    log_debug "Environment setup completed"
}

# Run Go unit tests
run_unit_tests() {
    log_header "Running Go Unit Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/unit_tests_${timestamp}.txt"
    local coverage_file="$COVERAGE_DIR/unit_coverage_${timestamp}.out"
    
    log_info "Running unit tests with coverage..."
    
    if go test -v -timeout="$TIMEOUT" -coverprofile="$coverage_file" \
        -covermode=atomic ./unit-testing/... > "$report_file" 2>&1; then
        log_success "Unit tests passed"
        
        # Generate coverage summary
        local coverage_summary="$COVERAGE_DIR/unit_coverage_${timestamp}_summary.txt"
        go tool cover -func="$coverage_file" > "$coverage_summary"
        
        # Extract coverage percentage
        local coverage_percent=$(tail -1 "$coverage_summary" | awk '{print $3}' | sed 's/%//')
        log_info "Unit test coverage: ${coverage_percent}%"
        
        if (( $(echo "$coverage_percent >= $COVERAGE_THRESHOLD" | bc -l) )); then
            log_success "Coverage threshold met (${coverage_percent}% >= ${COVERAGE_THRESHOLD}%)"
        else
            log_warn "Coverage below threshold (${coverage_percent}% < ${COVERAGE_THRESHOLD}%)"
        fi
        
        return 0
    else
        log_failure "Unit tests failed"
        cat "$report_file"
        return 1
    fi
}

# Run integration tests
run_integration_tests() {
    log_header "Running Integration Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/integration_tests_${timestamp}.txt"
    
    log_info "Running integration tests..."
    
    if go test -v -timeout="$TIMEOUT" -tags=integration \
        ./unit-testing/integration_test.go ./unit-testing/test_helpers.go \
        ./unit-testing/test_setup.go > "$report_file" 2>&1; then
        log_success "Integration tests passed"
        return 0
    else
        log_failure "Integration tests failed"
        cat "$report_file"
        return 1
    fi
}

# Run E2E scenario tests
run_e2e_tests() {
    if [ "$RUN_E2E" = false ]; then
        log_info "Skipping E2E tests (--no-e2e flag)"
        return 0
    fi
    
    log_header "Running E2E Scenario Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/e2e_tests_${timestamp}.txt"
    
    log_info "Running E2E scenario tests..."
    
    if go test -v -timeout="$TIMEOUT" \
        ./unit-testing/e2e_scenarios_test.go ./unit-testing/test_helpers.go \
        ./unit-testing/test_setup.go > "$report_file" 2>&1; then
        log_success "E2E tests passed"
        return 0
    else
        log_failure "E2E tests failed"
        cat "$report_file"
        return 1
    fi
}

# Run frontend tests
run_frontend_tests() {
    if [ "$RUN_FRONTEND" = false ]; then
        log_info "Skipping frontend tests (--no-frontend flag)"
        return 0
    fi
    
    log_header "Running Frontend Tests"
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    
    # Check if Node.js and npm are available
    if ! command -v node &> /dev/null || ! command -v npm &> /dev/null; then
        log_warn "Node.js or npm not found. Skipping frontend tests."
        return 0
    fi
    
    # Check if package.json exists
    if [ ! -f "package.json" ]; then
        log_warn "package.json not found. Skipping frontend tests."
        return 0
    fi
    
    log_info "Installing frontend dependencies..."
    npm install > /dev/null 2>&1
    
    # Run Jest unit tests
    log_info "Running frontend unit tests..."
    local jest_report="$REPORT_DIR/frontend_unit_${timestamp}.txt"
    
    if npm run test:unit > "$jest_report" 2>&1; then
        log_success "Frontend unit tests passed"
    else
        log_failure "Frontend unit tests failed"
        cat "$jest_report"
        return 1
    fi
    
    # Run Playwright E2E tests if available
    if command -v npx &> /dev/null && [ -f "playwright.config.js" ]; then
        log_info "Running frontend E2E tests..."
        local playwright_report="$REPORT_DIR/frontend_e2e_${timestamp}.txt"
        
        if npx playwright test > "$playwright_report" 2>&1; then
            log_success "Frontend E2E tests passed"
        else
            log_failure "Frontend E2E tests failed"
            cat "$playwright_report"
            return 1
        fi
    fi
    
    return 0
}

# Generate HTML reports
generate_html_reports() {
    if [ "$GENERATE_HTML" = false ]; then
        log_info "Skipping HTML report generation (--no-html flag)"
        return 0
    fi
    
    log_header "Generating HTML Reports"
    
    if [ -f "./generate-html-report.sh" ]; then
        log_info "Generating comprehensive HTML reports..."
        ./generate-html-report.sh all
        log_success "HTML reports generated in $HTML_REPORT_DIR/"
    else
        log_warn "HTML report generator not found. Skipping HTML generation."
    fi
}

# Generate test summary
generate_summary() {
    log_header "Test Execution Summary"
    
    local timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    local summary_file="$REPORT_DIR/test_summary_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$summary_file" << EOF
Real-Time Forum Test Execution Summary
Generated: $timestamp

Configuration:
- Timeout: $TIMEOUT
- Parallel Workers: $PARALLEL_WORKERS
- Coverage Threshold: $COVERAGE_THRESHOLD%
- HTML Reports: $GENERATE_HTML
- E2E Tests: $RUN_E2E
- Frontend Tests: $RUN_FRONTEND

Test Results:
EOF

    # Count test files and results
    local total_reports=$(ls -1 "$REPORT_DIR"/*.txt 2>/dev/null | wc -l)
    local unit_reports=$(ls -1 "$REPORT_DIR"/unit_tests_*.txt 2>/dev/null | wc -l)
    local integration_reports=$(ls -1 "$REPORT_DIR"/integration_tests_*.txt 2>/dev/null | wc -l)
    local e2e_reports=$(ls -1 "$REPORT_DIR"/e2e_tests_*.txt 2>/dev/null | wc -l)
    local frontend_reports=$(ls -1 "$REPORT_DIR"/frontend_*.txt 2>/dev/null | wc -l)
    
    echo "- Total Test Reports: $total_reports" >> "$summary_file"
    echo "- Unit Test Reports: $unit_reports" >> "$summary_file"
    echo "- Integration Test Reports: $integration_reports" >> "$summary_file"
    echo "- E2E Test Reports: $e2e_reports" >> "$summary_file"
    echo "- Frontend Test Reports: $frontend_reports" >> "$summary_file"
    
    # Coverage information
    local latest_coverage=$(ls -1t "$COVERAGE_DIR"/*_summary.txt 2>/dev/null | head -1)
    if [ -f "$latest_coverage" ]; then
        local coverage_percent=$(tail -1 "$latest_coverage" | awk '{print $3}' | sed 's/%//')
        echo "- Latest Coverage: ${coverage_percent}%" >> "$summary_file"
    fi
    
    echo "" >> "$summary_file"
    echo "Reports Location: $REPORT_DIR/" >> "$summary_file"
    echo "Coverage Data: $COVERAGE_DIR/" >> "$summary_file"
    echo "HTML Reports: $HTML_REPORT_DIR/" >> "$summary_file"
    
    log_info "Test summary saved to: $summary_file"
    
    # Display summary
    cat "$summary_file"
}

# Main execution
main() {
    local start_time=$(date +%s)
    local exit_code=0
    
    log_header "Comprehensive Test Runner"
    log_info "Starting comprehensive test execution..."
    
    # Parse arguments
    parse_arguments "$@"
    
    # Setup environment
    setup_environment
    
    # Run all test categories
    run_unit_tests || exit_code=1
    run_integration_tests || exit_code=1
    run_e2e_tests || exit_code=1
    run_frontend_tests || exit_code=1
    
    # Generate reports
    generate_html_reports
    generate_summary
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_header "Test Execution Complete"
    log_info "Total execution time: ${duration} seconds"
    
    if [ $exit_code -eq 0 ]; then
        log_success "All tests completed successfully!"
        log_info "View HTML reports at: $HTML_REPORT_DIR/index.html"
    else
        log_failure "Some tests failed. Check the reports for details."
        log_info "Reports available at: $REPORT_DIR/"
    fi
    
    exit $exit_code
}

# Run main function
main "$@"
