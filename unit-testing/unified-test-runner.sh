#!/bin/bash

# Unified Test Runner
# Supports both terminal and web-based test execution
# Version: 1.0.0

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="."
REPORT_DIR="test-reports"
COVERAGE_DIR="coverage"
HTML_REPORT_DIR="html-reports"
LOG_DIR="../logs"

# Default settings
MODE="terminal"
CATEGORY="all"
COVERAGE=false
PARALLEL=false
RACE=false
BENCHMARK=false
WATCH=false
VERBOSE=false
QUIET=false
TIMEOUT="10m"
WORKERS=4
GENERATE_HTML=false

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
${CYAN}Unified Test Runner${NC}
${CYAN}Version: 1.0.0${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS] [CATEGORY]

${YELLOW}MODES:${NC}
    --terminal          Terminal mode (default)
    --web               Web dashboard mode
    --api               API mode for web integration

${YELLOW}CATEGORIES:${NC}
    all                 Run all tests (default)
    unit                Run unit tests
    integration         Run integration tests
    auth                Run authentication tests
    messaging           Run messaging tests
    frontend            Run frontend tests
    e2e                 Run end-to-end tests

${YELLOW}OPTIONS:${NC}
    --coverage          Enable coverage analysis
    --parallel          Run tests in parallel
    --race              Enable race detection
    --benchmark         Run benchmark tests
    --watch             Watch mode (re-run on changes)
    --verbose           Enable verbose output
    --quiet             Disable output (except errors)
    --timeout TIME      Set test timeout (default: ${TIMEOUT})
    --workers N         Number of parallel workers (default: ${WORKERS})
    --html              Generate HTML reports
    --help              Show this help message

${YELLOW}EXAMPLES:${NC}
    $0                              # Run all tests in terminal mode
    $0 --web                        # Launch web dashboard
    $0 unit --coverage --verbose    # Run unit tests with coverage
    $0 all --parallel --html        # Run all tests in parallel with HTML reports

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            all|unit|integration|auth|messaging|frontend|e2e)
                CATEGORY="$1"
                shift
                ;;
            --terminal)
                MODE="terminal"
                shift
                ;;
            --web)
                MODE="web"
                shift
                ;;
            --api)
                MODE="api"
                shift
                ;;
            --coverage)
                COVERAGE=true
                shift
                ;;
            --parallel)
                PARALLEL=true
                shift
                ;;
            --race)
                RACE=true
                shift
                ;;
            --benchmark)
                BENCHMARK=true
                shift
                ;;
            --watch)
                WATCH=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --quiet)
                QUIET=true
                shift
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --workers)
                WORKERS="$2"
                shift 2
                ;;
            --html)
                GENERATE_HTML=true
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
    
    log_debug "Environment setup completed"
}

# Execute tests based on mode
execute_tests() {
    case "$MODE" in
        "terminal")
            execute_terminal_tests
            ;;
        "web")
            execute_web_tests
            ;;
        "api")
            execute_api_tests
            ;;
        *)
            log_error "Unknown mode: $MODE"
            exit 1
            ;;
    esac
}

# Terminal mode execution
execute_terminal_tests() {
    log_header "Running Tests in Terminal Mode"
    
    # Use existing test runners
    if [ -f "./test-runner.sh" ]; then
        local args=""
        
        if [ "$COVERAGE" = true ]; then
            args="$args --coverage"
        fi
        
        if [ "$PARALLEL" = true ]; then
            args="$args --parallel"
        fi
        
        if [ "$RACE" = true ]; then
            args="$args --race"
        fi
        
        if [ "$BENCHMARK" = true ]; then
            args="$args --benchmark"
        fi
        
        if [ "$VERBOSE" = true ]; then
            args="$args --verbose"
        fi
        
        if [ "$QUIET" = true ]; then
            args="$args --quiet"
        fi
        
        if [ "$GENERATE_HTML" = true ]; then
            args="$args --html"
        fi
        
        args="$args --timeout $TIMEOUT --workers $WORKERS"
        
        log_info "Executing: ./test-runner.sh $CATEGORY $args"
        ./test-runner.sh $CATEGORY $args
        
    elif [ -f "./test.sh" ]; then
        log_info "Using basic test runner..."
        ./test.sh $CATEGORY
    else
        log_error "No test runner found"
        exit 1
    fi
    
    # Generate HTML reports if requested
    if [ "$GENERATE_HTML" = true ] && [ -f "./generate-html-report.sh" ]; then
        log_info "Generating HTML reports..."
        ./generate-html-report.sh all
    fi
}

# Web mode execution
execute_web_tests() {
    log_header "Launching Web Dashboard"
    
    # Generate initial reports
    if [ -f "./generate-html-report.sh" ]; then
        ./generate-html-report.sh all
    fi
    
    # Start web server
    local port=8081
    
    log_info "Starting web server on port $port..."
    log_info "Dashboard URL: http://localhost:$port/web-dashboard/"
    
    if command -v python3 &> /dev/null; then
        python3 -m http.server $port
    elif command -v python &> /dev/null; then
        python -m SimpleHTTPServer $port
    else
        log_error "Python not found. Please install Python to use web mode."
        exit 1
    fi
}

# API mode execution
execute_api_tests() {
    log_header "Running Tests in API Mode"
    
    # This mode is used by the web dashboard to execute tests
    # It runs tests and outputs JSON results
    
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="$REPORT_DIR/api_test_${CATEGORY}_${timestamp}.json"
    
    # Execute tests and capture results
    local start_time=$(date +%s)
    
    if [ -f "./test-runner.sh" ]; then
        ./test-runner.sh $CATEGORY --json > "$report_file" 2>&1
        local exit_code=$?
    else
        # Fallback to basic runner
        ./test.sh $CATEGORY > "$report_file" 2>&1
        local exit_code=$?
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Generate JSON response
    cat > "${report_file}.response" << EOF
{
    "category": "$CATEGORY",
    "status": $([ $exit_code -eq 0 ] && echo '"passed"' || echo '"failed"'),
    "duration": $duration,
    "timestamp": "$timestamp",
    "exit_code": $exit_code,
    "report_file": "$report_file"
}
EOF
    
    # Output the response
    cat "${report_file}.response"
}

# Watch mode implementation
start_watch_mode() {
    log_info "Starting watch mode... (Press Ctrl+C to stop)"
    
    while true; do
        if command -v inotifywait &> /dev/null; then
            inotifywait -r -e modify,create,delete . --exclude '(test-reports|coverage|html-reports|logs)' 2>/dev/null
        else
            sleep 5
        fi
        
        log_info "Changes detected, running tests..."
        execute_tests
        echo ""
        log_info "Waiting for changes... (Press Ctrl+C to stop)"
    done
}

# Main execution
main() {
    # Parse arguments
    parse_arguments "$@"
    
    # Setup environment
    setup_environment
    
    # Execute based on mode
    if [ "$WATCH" = true ]; then
        start_watch_mode
    else
        execute_tests
    fi
}

# Run main function
main "$@"
