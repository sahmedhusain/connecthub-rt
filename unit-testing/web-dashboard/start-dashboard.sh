#!/bin/bash

# Web Dashboard Startup Script
# Starts the API server and opens the dashboard in the browser

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_PORT=8082
DASHBOARD_URL="http://localhost:${API_PORT}"

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

log_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run the dashboard API server."
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    # Check if port is available
    if lsof -Pi :$API_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warn "Port $API_PORT is already in use. Attempting to stop existing process..."
        pkill -f "api-server" || true
        sleep 2
        
        if lsof -Pi :$API_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_error "Port $API_PORT is still in use. Please stop the process manually."
            exit 1
        fi
    fi
    
    log_info "Prerequisites check completed"
}

# Setup Go dependencies
setup_dependencies() {
    log_info "Setting up Go dependencies..."
    
    cd "$SCRIPT_DIR"
    
    # Initialize go.mod if it doesn't exist
    if [ ! -f "go.mod" ]; then
        log_info "Initializing Go module..."
        go mod init dashboard-api
    fi
    
    # Download dependencies
    log_info "Downloading dependencies..."
    go mod tidy
    
    log_info "Dependencies setup completed"
}

# Build and start API server
start_api_server() {
    log_info "Building and starting API server..."
    
    cd "$SCRIPT_DIR"
    
    # Build the API server
    log_info "Building API server..."
    go build -o api-server api-server.go
    
    # Start the API server in background
    log_info "Starting API server on port $API_PORT..."
    ./api-server &
    API_PID=$!
    
    # Save PID for cleanup
    echo $API_PID > api-server.pid
    
    # Wait for server to start
    log_info "Waiting for API server to start..."
    for i in {1..30}; do
        if curl -s "http://localhost:$API_PORT/api/status" >/dev/null 2>&1; then
            log_info "API server started successfully"
            return 0
        fi
        sleep 1
    done
    
    log_error "API server failed to start within 30 seconds"
    return 1
}

# Open dashboard in browser
open_dashboard() {
    log_info "Opening dashboard in browser..."
    
    # Detect OS and open browser accordingly
    case "$(uname -s)" in
        Darwin)
            open "$DASHBOARD_URL"
            ;;
        Linux)
            if command -v xdg-open &> /dev/null; then
                xdg-open "$DASHBOARD_URL"
            elif command -v gnome-open &> /dev/null; then
                gnome-open "$DASHBOARD_URL"
            else
                log_warn "Could not detect browser. Please open $DASHBOARD_URL manually."
            fi
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            start "$DASHBOARD_URL"
            ;;
        *)
            log_warn "Unknown OS. Please open $DASHBOARD_URL manually."
            ;;
    esac
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    
    if [ -f "$SCRIPT_DIR/api-server.pid" ]; then
        API_PID=$(cat "$SCRIPT_DIR/api-server.pid")
        if kill -0 $API_PID 2>/dev/null; then
            log_info "Stopping API server (PID: $API_PID)..."
            kill $API_PID
            wait $API_PID 2>/dev/null || true
        fi
        rm -f "$SCRIPT_DIR/api-server.pid"
    fi
    
    # Clean up binary
    if [ -f "$SCRIPT_DIR/api-server" ]; then
        rm -f "$SCRIPT_DIR/api-server"
    fi
    
    log_info "Cleanup completed"
}

# Signal handlers
trap cleanup EXIT
trap 'log_info "Received interrupt signal"; exit 0' INT TERM

# Show help
show_help() {
    cat << EOF
${BLUE}Web Dashboard Startup Script${NC}

${YELLOW}USAGE:${NC}
    $0 [OPTIONS]

${YELLOW}OPTIONS:${NC}
    --port PORT         Set API server port (default: $API_PORT)
    --no-browser        Don't open browser automatically
    --background        Run in background mode
    --help              Show this help message

${YELLOW}EXAMPLES:${NC}
    $0                  # Start dashboard with default settings
    $0 --port 8083      # Start on custom port
    $0 --no-browser     # Start without opening browser
    $0 --background     # Run in background

${YELLOW}DASHBOARD URL:${NC}
    http://localhost:$API_PORT

${YELLOW}FEATURES:${NC}
    - Real-time test execution monitoring
    - WebSocket-based live updates
    - Interactive test result visualization
    - Coverage analysis and reporting
    - Export functionality (JSON, CSV, HTML)
    - Responsive design for mobile/desktop

For more information, see the documentation at:
https://github.com/your-repo/real-time-forum/blob/main/unit-testing/README.md
EOF
}

# Parse command line arguments
parse_arguments() {
    OPEN_BROWSER=true
    BACKGROUND_MODE=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --port)
                API_PORT="$2"
                DASHBOARD_URL="http://localhost:${API_PORT}"
                shift 2
                ;;
            --no-browser)
                OPEN_BROWSER=false
                shift
                ;;
            --background)
                BACKGROUND_MODE=true
                OPEN_BROWSER=false
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

# Main execution
main() {
    log_header "Real-Time Forum Test Dashboard"
    
    # Parse arguments
    parse_arguments "$@"
    
    log_info "Starting dashboard on port $API_PORT..."
    
    # Setup and start
    check_prerequisites
    setup_dependencies
    
    if start_api_server; then
        log_info "Dashboard API server is running on $DASHBOARD_URL"
        
        if [ "$OPEN_BROWSER" = true ]; then
            sleep 2  # Give server a moment to fully start
            open_dashboard
        fi
        
        if [ "$BACKGROUND_MODE" = true ]; then
            log_info "Running in background mode. Use 'kill $(cat api-server.pid)' to stop."
            exit 0
        else
            log_info "Dashboard is ready! Press Ctrl+C to stop."
            log_info "Dashboard URL: $DASHBOARD_URL"
            
            # Keep script running
            while true; do
                sleep 1
                # Check if API server is still running
                if [ -f "$SCRIPT_DIR/api-server.pid" ]; then
                    API_PID=$(cat "$SCRIPT_DIR/api-server.pid")
                    if ! kill -0 $API_PID 2>/dev/null; then
                        log_error "API server stopped unexpectedly"
                        exit 1
                    fi
                else
                    log_error "API server PID file not found"
                    exit 1
                fi
            done
        fi
    else
        log_error "Failed to start dashboard"
        exit 1
    fi
}

# Run main function
main "$@"
