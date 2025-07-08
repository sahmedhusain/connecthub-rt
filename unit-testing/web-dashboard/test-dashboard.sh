#!/bin/bash

# Test Dashboard Functionality
# Verifies that the web dashboard components work correctly

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_PORT=8082
TEST_TIMEOUT=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Test functions
test_prerequisites() {
    log_info "Testing prerequisites..."
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        return 1
    fi
    
    # Check curl
    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed"
        return 1
    fi
    
    # Check files exist
    local required_files=(
        "index.html"
        "dashboard.js"
        "styles.css"
        "api-server.go"
        "go.mod"
        "start-dashboard.sh"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$SCRIPT_DIR/$file" ]; then
            log_error "Required file missing: $file"
            return 1
        fi
    done
    
    log_success "Prerequisites check passed"
    return 0
}

test_go_build() {
    log_info "Testing Go build..."
    
    cd "$SCRIPT_DIR"
    
    # Clean previous builds
    rm -f api-server
    
    # Build
    if go build -o api-server api-server.go; then
        log_success "Go build successful"
        return 0
    else
        log_error "Go build failed"
        return 1
    fi
}

test_api_server() {
    log_info "Testing API server..."
    
    cd "$SCRIPT_DIR"
    
    # Start API server in background
    ./api-server &
    local api_pid=$!
    
    # Wait for server to start
    local attempts=0
    while [ $attempts -lt $TEST_TIMEOUT ]; do
        if curl -s "http://localhost:$API_PORT/api/status" >/dev/null 2>&1; then
            log_success "API server started successfully"
            break
        fi
        sleep 1
        ((attempts++))
    done
    
    if [ $attempts -eq $TEST_TIMEOUT ]; then
        log_error "API server failed to start within $TEST_TIMEOUT seconds"
        kill $api_pid 2>/dev/null || true
        return 1
    fi
    
    # Test API endpoints
    local endpoints=(
        "/api/status"
        "/api/test-results"
        "/api/coverage"
        "/api/reports"
    )
    
    for endpoint in "${endpoints[@]}"; do
        log_info "Testing endpoint: $endpoint"
        if curl -s -f "http://localhost:$API_PORT$endpoint" >/dev/null; then
            log_success "Endpoint $endpoint responded correctly"
        else
            log_error "Endpoint $endpoint failed"
            kill $api_pid 2>/dev/null || true
            return 1
        fi
    done
    
    # Test POST endpoint
    log_info "Testing POST /api/run-tests"
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{"category":"unit"}' \
        "http://localhost:$API_PORT/api/run-tests")
    
    if echo "$response" | grep -q "started"; then
        log_success "POST endpoint responded correctly"
    else
        log_error "POST endpoint failed"
        kill $api_pid 2>/dev/null || true
        return 1
    fi
    
    # Stop server
    kill $api_pid 2>/dev/null || true
    wait $api_pid 2>/dev/null || true
    
    log_success "API server tests passed"
    return 0
}

test_html_structure() {
    log_info "Testing HTML structure..."
    
    local html_file="$SCRIPT_DIR/index.html"
    
    # Check for required elements
    local required_elements=(
        "dashboard-container"
        "dashboard-header"
        "overview-cards"
        "test-categories"
        "test-results-section"
        "coverage-section"
    )
    
    for element in "${required_elements[@]}"; do
        if grep -q "$element" "$html_file"; then
            log_success "Found required element: $element"
        else
            log_error "Missing required element: $element"
            return 1
        fi
    done
    
    # Check for JavaScript includes
    if grep -q "dashboard.js" "$html_file"; then
        log_success "JavaScript file included"
    else
        log_error "JavaScript file not included"
        return 1
    fi
    
    # Check for CSS includes
    if grep -q "styles.css" "$html_file"; then
        log_success "CSS file included"
    else
        log_error "CSS file not included"
        return 1
    fi
    
    log_success "HTML structure tests passed"
    return 0
}

test_javascript_syntax() {
    log_info "Testing JavaScript syntax..."
    
    local js_file="$SCRIPT_DIR/dashboard.js"
    
    # Basic syntax check using Node.js if available
    if command -v node &> /dev/null; then
        if node -c "$js_file" 2>/dev/null; then
            log_success "JavaScript syntax is valid"
        else
            log_error "JavaScript syntax errors found"
            return 1
        fi
    else
        log_warn "Node.js not available, skipping syntax check"
    fi
    
    # Check for required functions
    local required_functions=(
        "TestDashboard"
        "initializeWebSocket"
        "loadTestData"
        "runTests"
        "exportResults"
    )
    
    for func in "${required_functions[@]}"; do
        if grep -q "$func" "$js_file"; then
            log_success "Found required function: $func"
        else
            log_error "Missing required function: $func"
            return 1
        fi
    done
    
    log_success "JavaScript tests passed"
    return 0
}

test_css_structure() {
    log_info "Testing CSS structure..."
    
    local css_file="$SCRIPT_DIR/styles.css"
    
    # Check for required CSS classes
    local required_classes=(
        "dashboard-container"
        "dashboard-header"
        "overview-card"
        "test-category"
        "progress-container"
        "notification"
        "modal-overlay"
    )
    
    for class in "${required_classes[@]}"; do
        if grep -q "\.$class" "$css_file"; then
            log_success "Found required CSS class: $class"
        else
            log_error "Missing required CSS class: $class"
            return 1
        fi
    done
    
    # Check for responsive design
    if grep -q "@media" "$css_file"; then
        log_success "Responsive design rules found"
    else
        log_warn "No responsive design rules found"
    fi
    
    log_success "CSS structure tests passed"
    return 0
}

test_startup_script() {
    log_info "Testing startup script..."
    
    local startup_script="$SCRIPT_DIR/start-dashboard.sh"
    
    # Check if script is executable
    if [ -x "$startup_script" ]; then
        log_success "Startup script is executable"
    else
        log_error "Startup script is not executable"
        return 1
    fi
    
    # Test help option
    if "$startup_script" --help >/dev/null 2>&1; then
        log_success "Startup script help option works"
    else
        log_error "Startup script help option failed"
        return 1
    fi
    
    log_success "Startup script tests passed"
    return 0
}

test_integration() {
    log_info "Testing integration with test runners..."
    
    # Check if test runners exist
    local test_runners=(
        "../test.sh"
        "../test-runner.sh"
        "../unified-test-runner.sh"
    )
    
    local found_runner=false
    for runner in "${test_runners[@]}"; do
        if [ -f "$SCRIPT_DIR/$runner" ]; then
            log_success "Found test runner: $runner"
            found_runner=true
        fi
    done
    
    if [ "$found_runner" = false ]; then
        log_warn "No test runners found - dashboard will work in simulation mode"
    fi
    
    # Check if test reports directory exists
    if [ -d "$SCRIPT_DIR/../test-reports" ]; then
        log_success "Test reports directory exists"
    else
        log_info "Creating test reports directory..."
        mkdir -p "$SCRIPT_DIR/../test-reports"
    fi
    
    # Check if coverage directory exists
    if [ -d "$SCRIPT_DIR/../coverage" ]; then
        log_success "Coverage directory exists"
    else
        log_info "Creating coverage directory..."
        mkdir -p "$SCRIPT_DIR/../coverage"
    fi
    
    log_success "Integration tests passed"
    return 0
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."
    
    # Kill any running API servers
    pkill -f "api-server" 2>/dev/null || true
    
    # Remove test binary
    rm -f "$SCRIPT_DIR/api-server"
    
    log_info "Cleanup completed"
}

# Main test execution
main() {
    log_header "Web Dashboard Test Suite"
    
    local failed_tests=0
    local total_tests=0
    
    # Setup cleanup
    trap cleanup EXIT
    
    # Run tests
    local tests=(
        "test_prerequisites"
        "test_html_structure"
        "test_javascript_syntax"
        "test_css_structure"
        "test_startup_script"
        "test_go_build"
        "test_api_server"
        "test_integration"
    )
    
    for test in "${tests[@]}"; do
        log_header "Running: $test"
        ((total_tests++))
        
        if $test; then
            log_success "$test passed"
        else
            log_error "$test failed"
            ((failed_tests++))
        fi
        
        echo
    done
    
    # Summary
    log_header "Test Summary"
    log_info "Total tests: $total_tests"
    log_info "Passed: $((total_tests - failed_tests))"
    log_info "Failed: $failed_tests"
    
    if [ $failed_tests -eq 0 ]; then
        log_success "All tests passed! Dashboard is ready to use."
        log_info "To start the dashboard, run: ./start-dashboard.sh"
        exit 0
    else
        log_error "Some tests failed. Please fix the issues before using the dashboard."
        exit 1
    fi
}

# Run main function
main "$@"
