#!/bin/bash

# Real-Time Forum Application Runner
# Enhanced script with comprehensive options and Docker support
# Version: 2.0.0

set -e  # Exit on any error

# Default configuration
DEFAULT_PORT="8080"
DEFAULT_ENV="development"
DEFAULT_DOCKER_IMAGE="forum"
DEFAULT_CONTAINER_NAME="forum"
LOG_DIR="logs"
TEST_DATA_FLAG=""
RESET_DB_FLAG=""
VERBOSE=false
INTERACTIVE=true
DOCKER_BUILD_ARGS=""
DOCKER_RUN_ARGS=""

# Testing configuration
TEST_MODE=""
TEST_COVERAGE=false
TEST_PARALLEL=false
TEST_RACE=false
TEST_BENCHMARK=false
TEST_WATCH=false
TEST_CATEGORY=""
TEST_TIMEOUT="10m"
TEST_WORKERS=4
TEST_REPORTS=false
TEST_ARTIFACTS=false

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

# Help function
show_help() {
    cat << EOF
${CYAN}Real-Time Forum Application Runner${NC}
${CYAN}Version: 2.0.0${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS] [COMMAND]

${YELLOW}COMMANDS:${NC}
    run                 Run the application (default)
    start               Start the application (alias for run)
    docker              Run with Docker
    native              Run with native Go
    build               Build Docker image only
    clean               Clean Docker containers and images
    logs                Show application logs
    status              Show application status
    stop                Stop running containers
    restart             Restart the application
    test                Run the test suite
    test-terminal       Run tests in terminal mode
    test-web            Run tests with web dashboard
    test-coverage       Run tests with coverage analysis
    test-all            Run comprehensive test suite
    test-unit           Run unit tests only
    test-integration    Run integration tests only
    test-frontend       Run frontend tests only
    test-e2e            Run end-to-end tests only
    test-performance    Run performance and benchmark tests
    help                Show this help message

${YELLOW}OPTIONS:${NC}
    -p, --port PORT     Set server port (default: ${DEFAULT_PORT})
    -e, --env ENV       Set environment (development|testing|production) (default: ${DEFAULT_ENV})
    -t, --test-data     Load test/seed data into database
    -r, --reset         Reset database (clear and recreate)
    -v, --verbose       Enable verbose output
    -q, --quiet         Disable interactive mode
    -h, --help          Show this help message

${YELLOW}TESTING OPTIONS:${NC}
    --test-terminal     Run tests in terminal mode with colored output
    --test-web          Launch web-based test dashboard
    --test-coverage     Enable coverage analysis for tests
    --test-parallel     Run tests in parallel mode
    --test-race         Enable race condition detection
    --test-benchmark    Run benchmark and performance tests
    --test-watch        Enable watch mode (re-run on file changes)
    --test-category CAT Specify test category (unit|integration|frontend|e2e|all)
    --test-timeout TIME Set test timeout (default: 10m)
    --test-workers N    Number of parallel test workers (default: 4)
    --test-reports      Generate HTML test reports
    --test-artifacts    Save test artifacts and logs

${YELLOW}DOCKER OPTIONS:${NC}
    --no-cache          Build Docker image without cache
    --build-args ARGS   Additional Docker build arguments
    --run-args ARGS     Additional Docker run arguments
    --image NAME        Docker image name (default: ${DEFAULT_DOCKER_IMAGE})
    --container NAME    Docker container name (default: ${DEFAULT_CONTAINER_NAME})

${YELLOW}EXAMPLES:${NC}
    $0                                  # Interactive mode
    $0 run --port 3000 --test-data     # Run on port 3000 with test data
    $0 docker --no-cache               # Build and run with Docker (no cache)
    $0 native --reset --verbose        # Run natively, reset DB, verbose output
    $0 build --image my-forum          # Build Docker image with custom name
    $0 clean                           # Clean all Docker resources
    $0 logs                            # Show recent application logs
    $0 test-terminal --test-coverage   # Run tests in terminal with coverage
    $0 test-web                        # Launch web-based test dashboard
    $0 test-all --test-parallel        # Run all tests in parallel
    $0 test-unit --test-watch          # Run unit tests in watch mode
    $0 test-e2e --test-reports         # Run E2E tests with HTML reports

${YELLOW}ENVIRONMENT VARIABLES:${NC}
    FORUM_PORT          Override default port
    FORUM_ENV           Override default environment
    FORUM_DB_PATH       Override database path
    DOCKER_BUILDKIT     Enable Docker BuildKit (recommended)

${YELLOW}CONFIGURATION FILES:${NC}
    .env                Environment variables file
    docker-compose.yml  Docker Compose configuration
    Dockerfile          Docker build configuration

For more information, visit: https://github.com/your-repo/real-time-forum
EOF
}

# Parse command line arguments
parse_arguments() {
    COMMAND=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            run|start|docker|native|build|clean|logs|status|stop|restart|test|test-terminal|test-web|test-coverage|test-all|test-unit|test-integration|test-frontend|test-e2e|test-performance|help)
                COMMAND="$1"
                shift
                ;;
            -p|--port)
                DEFAULT_PORT="$2"
                shift 2
                ;;
            -e|--env)
                DEFAULT_ENV="$2"
                shift 2
                ;;
            -t|--test-data)
                TEST_DATA_FLAG="--test-data"
                shift
                ;;
            -r|--reset)
                RESET_DB_FLAG="--reset"
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -q|--quiet)
                INTERACTIVE=false
                shift
                ;;
            --no-cache)
                DOCKER_BUILD_ARGS="$DOCKER_BUILD_ARGS --no-cache"
                shift
                ;;
            --build-args)
                DOCKER_BUILD_ARGS="$DOCKER_BUILD_ARGS $2"
                shift 2
                ;;
            --run-args)
                DOCKER_RUN_ARGS="$DOCKER_RUN_ARGS $2"
                shift 2
                ;;
            --image)
                DEFAULT_DOCKER_IMAGE="$2"
                shift 2
                ;;
            --container)
                DEFAULT_CONTAINER_NAME="$2"
                shift 2
                ;;
            --test-terminal)
                TEST_MODE="terminal"
                shift
                ;;
            --test-web)
                TEST_MODE="web"
                shift
                ;;
            --test-coverage)
                TEST_COVERAGE=true
                shift
                ;;
            --test-parallel)
                TEST_PARALLEL=true
                shift
                ;;
            --test-race)
                TEST_RACE=true
                shift
                ;;
            --test-benchmark)
                TEST_BENCHMARK=true
                shift
                ;;
            --test-watch)
                TEST_WATCH=true
                shift
                ;;
            --test-category)
                TEST_CATEGORY="$2"
                shift 2
                ;;
            --test-timeout)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            --test-workers)
                TEST_WORKERS="$2"
                shift 2
                ;;
            --test-reports)
                TEST_REPORTS=true
                shift
                ;;
            --test-artifacts)
                TEST_ARTIFACTS=true
                shift
                ;;
            -h|--help)
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

    # Set default command if none provided
    if [ -z "$COMMAND" ]; then
        if [ "$INTERACTIVE" = true ]; then
            COMMAND="interactive"
        else
            COMMAND="run"
        fi
    fi
}

# Check prerequisites
check_prerequisites() {
    log_debug "Checking prerequisites..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run the application natively."
        return 1
    fi

    # Check if Docker is installed (for Docker commands)
    if [[ "$COMMAND" == "docker" || "$COMMAND" == "build" || "$COMMAND" == "clean" ]]; then
        if ! command -v docker &> /dev/null; then
            log_error "Docker is not installed. Please install Docker to use Docker commands."
            return 1
        fi

        # Check if Docker daemon is running
        if ! docker info &> /dev/null; then
            log_error "Docker daemon is not running. Please start Docker."
            return 1
        fi
    fi

    # Create logs directory if it doesn't exist
    if [ ! -d "$LOG_DIR" ]; then
        log_debug "Creating logs directory: $LOG_DIR"
        mkdir -p "$LOG_DIR"
    fi

    log_debug "Prerequisites check completed successfully"
    return 0
}

# Load environment variables
load_environment() {
    log_debug "Loading environment configuration..."

    # Load from .env file if it exists
    if [ -f ".env" ]; then
        log_debug "Loading environment variables from .env file"
        set -a  # automatically export all variables
        source .env
        set +a
    fi

    # Override with environment variables if set
    if [ -n "$FORUM_PORT" ]; then
        DEFAULT_PORT="$FORUM_PORT"
        log_debug "Port overridden by environment variable: $DEFAULT_PORT"
    fi

    if [ -n "$FORUM_ENV" ]; then
        DEFAULT_ENV="$FORUM_ENV"
        log_debug "Environment overridden by environment variable: $DEFAULT_ENV"
    fi

    # Set environment-specific configurations
    case "$DEFAULT_ENV" in
        "production")
            log_debug "Production environment detected"
            DOCKER_BUILD_ARGS="$DOCKER_BUILD_ARGS --target production"
            ;;
        "testing")
            log_debug "Testing environment detected"
            TEST_DATA_FLAG="--test-data"
            ;;
        "development")
            log_debug "Development environment detected"
            ;;
        *)
            log_warn "Unknown environment: $DEFAULT_ENV, using development defaults"
            DEFAULT_ENV="development"
            ;;
    esac
}

# Docker utility functions
docker_cleanup() {
    log_info "Cleaning up Docker resources..."

    # Stop and remove container if it exists
    if docker ps -a --format 'table {{.Names}}' | grep -q "^${DEFAULT_CONTAINER_NAME}$"; then
        log_debug "Stopping container: $DEFAULT_CONTAINER_NAME"
        docker stop "$DEFAULT_CONTAINER_NAME" 2>/dev/null || true
        log_debug "Removing container: $DEFAULT_CONTAINER_NAME"
        docker rm "$DEFAULT_CONTAINER_NAME" 2>/dev/null || true
    fi

    # Remove image if it exists
    if docker images --format 'table {{.Repository}}' | grep -q "^${DEFAULT_DOCKER_IMAGE}$"; then
        log_debug "Removing image: $DEFAULT_DOCKER_IMAGE"
        docker rmi "$DEFAULT_DOCKER_IMAGE" 2>/dev/null || true
    fi

    # Clean up dangling images
    log_debug "Cleaning up dangling Docker images"
    docker image prune -f >/dev/null 2>&1 || true

    log_info "Docker cleanup completed"
}

docker_build() {
    log_info "Building Docker image: $DEFAULT_DOCKER_IMAGE"
    log_debug "Build arguments: $DOCKER_BUILD_ARGS"

    # Build the Docker image
    if [ "$VERBOSE" = true ]; then
        docker build $DOCKER_BUILD_ARGS -t "$DEFAULT_DOCKER_IMAGE" .
    else
        docker build $DOCKER_BUILD_ARGS -t "$DEFAULT_DOCKER_IMAGE" . >/dev/null
    fi

    if [ $? -eq 0 ]; then
        log_info "Docker image built successfully: $DEFAULT_DOCKER_IMAGE"
        return 0
    else
        log_error "Failed to build Docker image"
        return 1
    fi
}

docker_run() {
    log_info "Running Docker container: $DEFAULT_CONTAINER_NAME"
    log_debug "Container port mapping: $DEFAULT_PORT:8080"
    log_debug "Run arguments: $DOCKER_RUN_ARGS"

    # Prepare Docker run command
    DOCKER_CMD="docker run -p ${DEFAULT_PORT}:8080 --name ${DEFAULT_CONTAINER_NAME} ${DOCKER_RUN_ARGS} ${DEFAULT_DOCKER_IMAGE}"

    log_debug "Docker command: $DOCKER_CMD"

    # Run the container
    if [ "$VERBOSE" = true ]; then
        $DOCKER_CMD
    else
        $DOCKER_CMD
    fi
}

# Native Go utility functions
native_run() {
    log_info "Running application natively with Go"
    log_debug "Port: $DEFAULT_PORT"
    log_debug "Environment: $DEFAULT_ENV"
    log_debug "Test data flag: $TEST_DATA_FLAG"
    log_debug "Reset DB flag: $RESET_DB_FLAG"

    # Prepare Go run command
    GO_CMD="go run main.go --port=$DEFAULT_PORT $TEST_DATA_FLAG $RESET_DB_FLAG"

    log_debug "Go command: $GO_CMD"

    # Clear screen for better visibility
    if [ "$INTERACTIVE" = true ]; then
        clear
    fi

    # Run the application
    $GO_CMD
}

# Testing utility functions
run_tests_terminal() {
    log_header "Running Tests in Terminal Mode"

    local test_args="--terminal"

    # Build test arguments
    if [ "$TEST_COVERAGE" = true ]; then
        test_args="$test_args --coverage"
    fi

    if [ "$TEST_PARALLEL" = true ]; then
        test_args="$test_args --parallel"
    fi

    if [ "$TEST_RACE" = true ]; then
        test_args="$test_args --race"
    fi

    if [ "$TEST_BENCHMARK" = true ]; then
        test_args="$test_args --benchmark"
    fi

    if [ "$TEST_WATCH" = true ]; then
        test_args="$test_args --watch"
    fi

    if [ "$TEST_REPORTS" = true ]; then
        test_args="$test_args --html"
    fi

    if [ "$VERBOSE" = true ]; then
        test_args="$test_args --verbose"
    fi

    if [ "$INTERACTIVE" = false ]; then
        test_args="$test_args --quiet"
    fi

    if [ -n "$TEST_TIMEOUT" ]; then
        test_args="$test_args --timeout $TEST_TIMEOUT"
    fi

    if [ -n "$TEST_WORKERS" ]; then
        test_args="$test_args --workers $TEST_WORKERS"
    fi

    # Add category
    local category=${TEST_CATEGORY:-"all"}
    test_args="$test_args $category"

    # Change to testing directory and run tests
    cd unit-testing

    if [ -f "./unified-test-runner.sh" ]; then
        log_info "Using unified test runner..."
        ./unified-test-runner.sh $test_args
    elif [ -f "./test-runner.sh" ]; then
        log_info "Using advanced test runner..."
        # Convert args for legacy runner
        local legacy_args=""
        if [[ "$test_args" == *"--coverage"* ]]; then
            legacy_args="$legacy_args --coverage"
        fi
        if [[ "$test_args" == *"--parallel"* ]]; then
            legacy_args="$legacy_args --parallel"
        fi
        if [[ "$test_args" == *"--verbose"* ]]; then
            legacy_args="$legacy_args --verbose"
        fi
        ./test-runner.sh $category $legacy_args
    elif [ -f "./test.sh" ]; then
        log_info "Using basic test runner..."
        ./test.sh $category
    else
        log_error "No test runner found in unit-testing directory"
        return 1
    fi

    cd ..
}

run_tests_web() {
    log_header "Launching Enhanced Web-based Test Dashboard"

    # Check if we have the enhanced web dashboard
    if [ ! -f "unit-testing/web-dashboard/index.html" ]; then
        log_error "Web dashboard not found. Please ensure the dashboard files are present."
        return 1
    fi

    # Check if we have the Go API server
    if [ ! -f "unit-testing/web-dashboard/api-server.go" ]; then
        log_error "Dashboard API server not found. Please ensure api-server.go is present."
        return 1
    fi

    cd unit-testing/web-dashboard

    # Use the enhanced dashboard startup script if available
    if [ -f "./start-dashboard.sh" ]; then
        log_info "Using enhanced dashboard with Go API server..."

        # Build startup arguments
        local dashboard_args=""

        if [ "$QUIET_MODE" = true ]; then
            dashboard_args="$dashboard_args --no-browser"
        fi

        if [ -n "$TEST_PORT" ]; then
            dashboard_args="$dashboard_args --port $TEST_PORT"
        fi

        # Start the enhanced dashboard
        ./start-dashboard.sh $dashboard_args

    else
        # Fallback to manual setup
        log_info "Enhanced startup script not found, using manual setup..."

        # Check Go dependencies
        if ! command -v go &> /dev/null; then
            log_error "Go is required for the enhanced dashboard. Please install Go."
            cd ../..
            return 1
        fi

        # Setup Go module if needed
        if [ ! -f "go.mod" ]; then
            log_info "Initializing Go module..."
            go mod init dashboard-api
            go mod tidy
        fi

        # Build and start API server
        log_info "Building dashboard API server..."
        go build -o api-server api-server.go

        if [ $? -ne 0 ]; then
            log_error "Failed to build API server"
            cd ../..
            return 1
        fi

        # Start API server
        local api_port=8082
        log_info "Starting dashboard API server on port $api_port..."
        ./api-server &
        local api_pid=$!

        # Wait for API server to start
        log_info "Waiting for API server to start..."
        for i in {1..30}; do
            if curl -s "http://localhost:$api_port/api/status" >/dev/null 2>&1; then
                log_info "API server started successfully"
                break
            fi
            sleep 1
        done

        # Open dashboard in browser
        local dashboard_url="http://localhost:$api_port"
        log_info "Dashboard URL: $dashboard_url"

        if [ "$QUIET_MODE" != true ]; then
            sleep 2
            if command -v open &> /dev/null; then
                open "$dashboard_url"
            elif command -v xdg-open &> /dev/null; then
                xdg-open "$dashboard_url"
            fi
        fi

        log_info "Enhanced web dashboard is running. Press Ctrl+C to stop."
        log_info "Features:"
        log_info "  - Real-time test execution monitoring"
        log_info "  - WebSocket-based live updates"
        log_info "  - Interactive coverage reports"
        log_info "  - Export functionality (JSON, CSV, HTML)"
        log_info "  - Responsive design for mobile/desktop"

        # Cleanup function
        cleanup_dashboard() {
            log_info "Stopping dashboard..."
            kill $api_pid 2>/dev/null
            rm -f api-server
            cd ../..
        }

        # Wait for user to stop
        trap cleanup_dashboard SIGINT SIGTERM
        wait $api_pid
        cleanup_dashboard
    fi

    cd ../..
}

start_test_api_server() {
    local port=$1

    # Create a simple API server for test execution
    cat > test-api-server.py << 'EOF'
#!/usr/bin/env python3
import http.server
import socketserver
import json
import subprocess
import threading
import time
import os
from urllib.parse import urlparse, parse_qs

class TestAPIHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        parsed_path = urlparse(self.path)

        if parsed_path.path == '/api/status':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()

            response = {'status': 'running', 'timestamp': time.time()}
            self.wfile.write(json.dumps(response).encode())

        elif parsed_path.path == '/api/test-results':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()

            # Mock test results
            results = {
                'total': 45,
                'passed': 38,
                'failed': 5,
                'skipped': 2,
                'categories': {
                    'unit': {'total': 15, 'passed': 13, 'failed': 2, 'duration': 12},
                    'integration': {'total': 10, 'passed': 9, 'failed': 1, 'duration': 25},
                    'auth': {'total': 8, 'passed': 7, 'failed': 1, 'duration': 18},
                    'messaging': {'total': 6, 'passed': 5, 'failed': 1, 'duration': 22},
                    'frontend': {'total': 4, 'passed': 4, 'failed': 0, 'duration': 8},
                    'e2e': {'total': 2, 'passed': 0, 'failed': 0, 'duration': 0}
                }
            }
            self.wfile.write(json.dumps(results).encode())
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == '/api/run-tests':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            data = json.loads(post_data.decode('utf-8'))

            category = data.get('category', 'all')

            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()

            # Start test execution in background
            def run_tests():
                try:
                    if os.path.exists('./test-runner.sh'):
                        subprocess.run(['./test-runner.sh', category], check=True)
                    elif os.path.exists('./test.sh'):
                        subprocess.run(['./test.sh', category], check=True)
                except subprocess.CalledProcessError:
                    pass

            threading.Thread(target=run_tests).start()

            response = {'status': 'started', 'category': category}
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.end_headers()

if __name__ == '__main__':
    PORT = int(os.environ.get('PORT', 8082))
    with socketserver.TCPServer(("", PORT), TestAPIHandler) as httpd:
        print(f"Test API server running on port {PORT}")
        httpd.serve_forever()
EOF

    python3 test-api-server.py
}

create_web_dashboard() {
    log_info "Creating web-based test dashboard..."

    # Create dashboard directory
    mkdir -p unit-testing/web-dashboard

    # Check if dashboard files exist
    if [ ! -f "unit-testing/web-dashboard/index.html" ]; then
        log_info "Dashboard files not found. Please ensure the web dashboard is properly set up."
        log_info "Expected files:"
        log_info "  - unit-testing/web-dashboard/index.html"
        log_info "  - unit-testing/web-dashboard/styles.css"
        log_info "  - unit-testing/web-dashboard/dashboard.js"
        return 1
    fi

    log_info "Web dashboard is ready"
}

# Status and monitoring functions
show_status() {
    log_header "Application Status"

    # Check if Docker container is running
    if docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep -q "^${DEFAULT_CONTAINER_NAME}"; then
        echo -e "${GREEN}Docker Container:${NC} Running"
        docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' | grep "^${DEFAULT_CONTAINER_NAME}"
    else
        echo -e "${YELLOW}Docker Container:${NC} Not running"
    fi

    # Check if port is in use
    if command -v lsof &> /dev/null; then
        if lsof -i ":$DEFAULT_PORT" &> /dev/null; then
            echo -e "${GREEN}Port $DEFAULT_PORT:${NC} In use"
            lsof -i ":$DEFAULT_PORT" | head -2
        else
            echo -e "${YELLOW}Port $DEFAULT_PORT:${NC} Available"
        fi
    fi

    # Show recent logs if available
    if [ -d "$LOG_DIR" ] && [ "$(ls -A $LOG_DIR 2>/dev/null)" ]; then
        echo -e "\n${BLUE}Recent Log Files:${NC}"
        ls -la "$LOG_DIR" | tail -5
    fi
}

show_logs() {
    log_header "Application Logs"

    if [ -d "$LOG_DIR" ] && [ "$(ls -A $LOG_DIR 2>/dev/null)" ]; then
        # Show the most recent log file
        LATEST_LOG=$(ls -t "$LOG_DIR"/*.log 2>/dev/null | head -1)
        if [ -n "$LATEST_LOG" ]; then
            log_info "Showing latest log file: $LATEST_LOG"
            echo -e "${CYAN}----------------------------------------${NC}"
            tail -50 "$LATEST_LOG"
            echo -e "${CYAN}----------------------------------------${NC}"
        else
            log_warn "No log files found in $LOG_DIR"
        fi
    else
        log_warn "Log directory $LOG_DIR is empty or doesn't exist"
    fi
}

# Interactive menu
show_interactive_menu() {
    clear
    log_header "Real-Time Forum Application Runner"

    echo -e "${CYAN}Environment:${NC} $DEFAULT_ENV"
    echo -e "${CYAN}Port:${NC} $DEFAULT_PORT"
    echo -e "${CYAN}Docker Image:${NC} $DEFAULT_DOCKER_IMAGE"
    echo -e "${CYAN}Container:${NC} $DEFAULT_CONTAINER_NAME"
    echo ""

    echo -e "${YELLOW}Choose an option:${NC}"
    echo "1. Run with Docker (recommended)"
    echo "2. Run with native Go"
    echo "3. Build Docker image only"
    echo "4. Clean Docker resources"
    echo "5. Show application status"
    echo "6. Show application logs"
    echo "7. Stop running containers"
    echo "8. Restart application"
    echo ""
    echo -e "${CYAN}Testing Options:${NC}"
    echo "9. Run tests (interactive)"
    echo "10. Terminal test mode"
    echo "11. Web test dashboard"
    echo "12. Test with coverage"
    echo "13. Performance tests"
    echo ""
    echo "14. Advanced options"
    echo "0. Exit"
    echo ""

    read -p "Enter your choice [0-14]: " choice

    case $choice in
        1)
            COMMAND="docker"
            ;;
        2)
            COMMAND="native"
            ;;
        3)
            COMMAND="build"
            ;;
        4)
            COMMAND="clean"
            ;;
        5)
            COMMAND="status"
            ;;
        6)
            COMMAND="logs"
            ;;
        7)
            COMMAND="stop"
            ;;
        8)
            COMMAND="restart"
            ;;
        9)
            COMMAND="test"
            ;;
        10)
            COMMAND="test-terminal"
            ;;
        11)
            COMMAND="test-web"
            ;;
        12)
            COMMAND="test-coverage"
            ;;
        13)
            COMMAND="test-performance"
            ;;
        14)
            show_advanced_menu
            return
            ;;
        0)
            log_info "Goodbye!"
            exit 0
            ;;
        *)
            log_error "Invalid choice. Please try again."
            sleep 2
            show_interactive_menu
            return
            ;;
    esac
}

# Advanced options menu
show_advanced_menu() {
    clear
    log_header "Advanced Options"

    echo -e "${YELLOW}Current Configuration:${NC}"
    echo -e "Port: ${CYAN}$DEFAULT_PORT${NC}"
    echo -e "Environment: ${CYAN}$DEFAULT_ENV${NC}"
    echo -e "Test Data: ${CYAN}$([ -n "$TEST_DATA_FLAG" ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Reset DB: ${CYAN}$([ -n "$RESET_DB_FLAG" ] && echo "Enabled" || echo "Disabled")${NC}"
    echo -e "Verbose: ${CYAN}$([ "$VERBOSE" = true ] && echo "Enabled" || echo "Disabled")${NC}"
    echo ""

    echo -e "${YELLOW}Advanced Options:${NC}"
    echo "1. Change port"
    echo "2. Change environment"
    echo "3. Toggle test data loading"
    echo "4. Toggle database reset"
    echo "5. Toggle verbose output"
    echo "6. Set Docker build arguments"
    echo "7. Set Docker run arguments"
    echo "8. Reset to defaults"
    echo "9. Back to main menu"
    echo "0. Exit"
    echo ""

    read -p "Enter your choice [0-9]: " choice

    case $choice in
        1)
            read -p "Enter new port (current: $DEFAULT_PORT): " new_port
            if [[ "$new_port" =~ ^[0-9]+$ ]] && [ "$new_port" -ge 1 ] && [ "$new_port" -le 65535 ]; then
                DEFAULT_PORT="$new_port"
                log_info "Port changed to: $DEFAULT_PORT"
            else
                log_error "Invalid port number"
            fi
            sleep 2
            show_advanced_menu
            ;;
        2)
            echo "Available environments: development, testing, production"
            read -p "Enter environment (current: $DEFAULT_ENV): " new_env
            case "$new_env" in
                development|testing|production)
                    DEFAULT_ENV="$new_env"
                    log_info "Environment changed to: $DEFAULT_ENV"
                    ;;
                *)
                    log_error "Invalid environment"
                    ;;
            esac
            sleep 2
            show_advanced_menu
            ;;
        3)
            if [ -n "$TEST_DATA_FLAG" ]; then
                TEST_DATA_FLAG=""
                log_info "Test data loading disabled"
            else
                TEST_DATA_FLAG="--test-data"
                log_info "Test data loading enabled"
            fi
            sleep 2
            show_advanced_menu
            ;;
        4)
            if [ -n "$RESET_DB_FLAG" ]; then
                RESET_DB_FLAG=""
                log_info "Database reset disabled"
            else
                RESET_DB_FLAG="--reset"
                log_info "Database reset enabled"
            fi
            sleep 2
            show_advanced_menu
            ;;
        5)
            if [ "$VERBOSE" = true ]; then
                VERBOSE=false
                log_info "Verbose output disabled"
            else
                VERBOSE=true
                log_info "Verbose output enabled"
            fi
            sleep 2
            show_advanced_menu
            ;;
        6)
            read -p "Enter Docker build arguments (current: $DOCKER_BUILD_ARGS): " new_args
            DOCKER_BUILD_ARGS="$new_args"
            log_info "Docker build arguments updated"
            sleep 2
            show_advanced_menu
            ;;
        7)
            read -p "Enter Docker run arguments (current: $DOCKER_RUN_ARGS): " new_args
            DOCKER_RUN_ARGS="$new_args"
            log_info "Docker run arguments updated"
            sleep 2
            show_advanced_menu
            ;;
        8)
            DEFAULT_PORT="8080"
            DEFAULT_ENV="development"
            TEST_DATA_FLAG=""
            RESET_DB_FLAG=""
            VERBOSE=false
            DOCKER_BUILD_ARGS=""
            DOCKER_RUN_ARGS=""
            log_info "Configuration reset to defaults"
            sleep 2
            show_advanced_menu
            ;;
        9)
            show_interactive_menu
            ;;
        0)
            log_info "Goodbye!"
            exit 0
            ;;
        *)
            log_error "Invalid choice. Please try again."
            sleep 2
            show_advanced_menu
            ;;
    esac
}

# Command execution functions
execute_command() {
    case "$COMMAND" in
        "interactive")
            show_interactive_menu
            execute_command
            ;;
        "run"|"start")
            # Default run command - prefer Docker if available, otherwise native
            if command -v docker &> /dev/null && docker info &> /dev/null; then
                COMMAND="docker"
            else
                COMMAND="native"
            fi
            execute_command
            ;;
        "restart")
            log_header "Restarting Application"
            # First stop any running instances
            if docker ps --format 'table {{.Names}}' | grep -q "^${DEFAULT_CONTAINER_NAME}$"; then
                log_info "Stopping Docker container: $DEFAULT_CONTAINER_NAME"
                docker stop "$DEFAULT_CONTAINER_NAME"
                log_info "Container stopped successfully"
            fi

            # Also try to kill any process using the port
            if command -v lsof &> /dev/null; then
                PID=$(lsof -ti ":$DEFAULT_PORT" 2>/dev/null)
                if [ -n "$PID" ]; then
                    log_info "Killing process using port $DEFAULT_PORT (PID: $PID)"
                    kill "$PID" 2>/dev/null || true
                fi
            fi

            # Wait a moment for cleanup
            sleep 2

            # Now start the application
            log_info "Starting application..."
            if command -v docker &> /dev/null && docker info &> /dev/null; then
                COMMAND="docker"
            else
                COMMAND="native"
            fi
            execute_command
            ;;
        "docker")
            log_header "Running with Docker"
            docker_cleanup
            if docker_build; then
                docker_run
            else
                log_error "Failed to build Docker image"
                exit 1
            fi
            ;;
        "native")
            log_header "Running with Native Go"
            native_run
            ;;
        "build")
            log_header "Building Docker Image"
            docker_build
            ;;
        "clean")
            log_header "Cleaning Docker Resources"
            docker_cleanup
            log_info "Docker cleanup completed successfully"
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs
            ;;
        "stop")
            log_header "Stopping Application"
            if docker ps --format 'table {{.Names}}' | grep -q "^${DEFAULT_CONTAINER_NAME}$"; then
                log_info "Stopping Docker container: $DEFAULT_CONTAINER_NAME"
                docker stop "$DEFAULT_CONTAINER_NAME"
                log_info "Container stopped successfully"
            else
                log_warn "No running container found with name: $DEFAULT_CONTAINER_NAME"
            fi

            # Also try to kill any process using the port
            if command -v lsof &> /dev/null; then
                PID=$(lsof -ti ":$DEFAULT_PORT" 2>/dev/null)
                if [ -n "$PID" ]; then
                    log_info "Killing process using port $DEFAULT_PORT (PID: $PID)"
                    kill "$PID" 2>/dev/null || true
                fi
            fi
            ;;
        "test")
            log_header "Running Tests"
            if [ -f "./unit-testing/test.sh" ]; then
                log_info "Executing test runner..."
                cd unit-testing && ./test.sh
            elif [ -f "./unit-testing/test-runner.sh" ]; then
                log_info "Executing advanced test runner..."
                cd unit-testing && ./test-runner.sh
            else
                log_error "Test runners not found in unit-testing/ directory"
                log_info "Available test runners:"
                log_info "  - unit-testing/test.sh (basic test runner)"
                log_info "  - unit-testing/test-runner.sh (advanced test runner)"
                exit 1
            fi
            ;;
        "test-terminal")
            run_tests_terminal
            ;;
        "test-web")
            run_tests_web
            ;;
        "test-coverage")
            TEST_COVERAGE=true
            run_tests_terminal
            ;;
        "test-all")
            TEST_CATEGORY="all"
            run_tests_terminal
            ;;
        "test-unit")
            TEST_CATEGORY="unit"
            run_tests_terminal
            ;;
        "test-integration")
            TEST_CATEGORY="integration"
            run_tests_terminal
            ;;
        "test-frontend")
            TEST_CATEGORY="frontend"
            run_tests_terminal
            ;;
        "test-e2e")
            TEST_CATEGORY="e2e"
            run_tests_terminal
            ;;
        "test-performance")
            TEST_BENCHMARK=true
            TEST_CATEGORY="all"
            run_tests_terminal
            ;;
        "help")
            show_help
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Cleanup function for graceful shutdown
cleanup() {
    log_debug "Performing cleanup..."
    # Add any cleanup tasks here
    exit 0
}

# Signal handlers
trap cleanup SIGINT SIGTERM

# Main execution
main() {
    # Show banner
    if [ "$INTERACTIVE" = true ] && [ "$COMMAND" != "help" ]; then
        echo -e "${PURPLE}"
        echo "╔══════════════════════════════════════════════════════════════╗"
        echo "║                Real-Time Forum Application                   ║"
        echo "║                     Enhanced Runner v2.0                     ║"
        echo "╚══════════════════════════════════════════════════════════════╝"
        echo -e "${NC}"
    fi

    # Parse command line arguments
    parse_arguments "$@"

    # Load environment configuration
    load_environment

    # Check prerequisites
    if ! check_prerequisites; then
        log_error "Prerequisites check failed"
        exit 1
    fi

    # Execute the command
    execute_command
}

# Run main function with all arguments
main "$@"