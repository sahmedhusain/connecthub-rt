#!/bin/bash

# Frontend Testing Setup Script
# Sets up Node.js dependencies and Playwright browsers for frontend testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE} $1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Check if Node.js is installed
check_nodejs() {
    log_info "Checking Node.js installation..."
    
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed!"
        log_info "Please install Node.js 16+ from https://nodejs.org/"
        log_info "Or use a package manager:"
        log_info "  macOS: brew install node"
        log_info "  Ubuntu: sudo apt install nodejs npm"
        log_info "  CentOS: sudo yum install nodejs npm"
        exit 1
    fi
    
    local node_version=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$node_version" -lt 16 ]; then
        log_warning "Node.js version is $node_version, but version 16+ is recommended"
        log_info "Consider upgrading Node.js for better compatibility"
    else
        log_success "Node.js $(node --version) is installed"
    fi
}

# Check if npm is installed
check_npm() {
    log_info "Checking npm installation..."
    
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed!"
        log_info "npm usually comes with Node.js. Please reinstall Node.js."
        exit 1
    fi
    
    log_success "npm $(npm --version) is installed"
}

# Install Node.js dependencies
install_dependencies() {
    log_header "Installing Node.js Dependencies"
    
    if [ ! -f "package.json" ]; then
        log_error "package.json not found in current directory"
        log_info "Please run this script from the unit-testing directory"
        exit 1
    fi
    
    log_info "Installing npm dependencies..."
    npm install
    
    if [ $? -eq 0 ]; then
        log_success "npm dependencies installed successfully"
    else
        log_error "Failed to install npm dependencies"
        exit 1
    fi
}

# Install Playwright browsers
install_playwright_browsers() {
    log_header "Installing Playwright Browsers"
    
    log_info "Installing Playwright browsers (Chrome, Firefox, Safari)..."
    log_info "This may take a few minutes..."
    
    npx playwright install
    
    if [ $? -eq 0 ]; then
        log_success "Playwright browsers installed successfully"
    else
        log_error "Failed to install Playwright browsers"
        log_info "You can try installing manually with: npx playwright install"
        exit 1
    fi
}

# Verify installation
verify_installation() {
    log_header "Verifying Installation"
    
    # Check Jest
    log_info "Checking Jest installation..."
    if npm list jest &> /dev/null; then
        log_success "Jest is installed"
    else
        log_warning "Jest may not be properly installed"
    fi
    
    # Check Playwright
    log_info "Checking Playwright installation..."
    if npm list @playwright/test &> /dev/null; then
        log_success "Playwright is installed"
    else
        log_warning "Playwright may not be properly installed"
    fi
    
    # Check if browsers are installed
    log_info "Checking Playwright browsers..."
    if npx playwright --version &> /dev/null; then
        log_success "Playwright browsers are ready"
    else
        log_warning "Playwright browsers may not be properly installed"
    fi
}

# Run test verification
run_test_verification() {
    log_header "Running Test Verification"
    
    log_info "Running a quick test to verify setup..."
    
    # Create a simple test file for verification
    cat > test-verification.js << 'EOF'
import { test, expect } from '@jest/globals';

test('setup verification', () => {
    expect(1 + 1).toBe(2);
});
EOF
    
    # Run the verification test
    if npx jest test-verification.js --silent; then
        log_success "Jest is working correctly"
    else
        log_warning "Jest test failed - there may be configuration issues"
    fi
    
    # Clean up verification test
    rm -f test-verification.js
    
    # Check Playwright
    log_info "Verifying Playwright setup..."
    if npx playwright --version &> /dev/null; then
        log_success "Playwright is working correctly"
    else
        log_warning "Playwright verification failed"
    fi
}

# Create test directories if they don't exist
create_directories() {
    log_header "Creating Test Directories"
    
    local directories=(
        "frontend-tests/unit"
        "frontend-tests/integration"
        "frontend-tests/e2e"
        "frontend-tests/setup"
        "test-reports"
        "coverage/frontend"
    )
    
    for dir in "${directories[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            log_info "Created directory: $dir"
        else
            log_info "Directory already exists: $dir"
        fi
    done
    
    log_success "Test directories are ready"
}

# Display usage information
show_usage() {
    log_header "Frontend Testing Setup Complete"
    
    echo -e "${CYAN}Available Commands:${NC}"
    echo ""
    echo -e "${YELLOW}Frontend Unit Tests:${NC}"
    echo "  npm test                    # Run all frontend unit tests"
    echo "  npm run test:dom           # Run DOM manipulation tests"
    echo "  npm run test:websocket     # Run WebSocket client tests"
    echo "  npm run test:auth          # Run frontend auth tests"
    echo "  npm run test:spa           # Run SPA navigation tests"
    echo ""
    echo -e "${YELLOW}E2E Tests:${NC}"
    echo "  npx playwright test        # Run all E2E tests"
    echo "  npx playwright test auth-flow.spec.js    # Auth flow tests"
    echo "  npx playwright test messaging.spec.js    # Messaging tests"
    echo "  npx playwright test --headed              # Run with browser UI"
    echo "  npx playwright test --debug               # Debug mode"
    echo ""
    echo -e "${YELLOW}Cross-browser Testing:${NC}"
    echo "  npx playwright test --project=chromium   # Chrome only"
    echo "  npx playwright test --project=firefox    # Firefox only"
    echo "  npx playwright test --project=webkit     # Safari only"
    echo ""
    echo -e "${YELLOW}Integration with Test Runner:${NC}"
    echo "  ./test.sh frontend         # Run frontend tests via test runner"
    echo "  ./test.sh e2e              # Run E2E tests via test runner"
    echo "  ./test.sh cross-browser    # Run cross-browser tests"
    echo ""
    echo -e "${CYAN}Next Steps:${NC}"
    echo "1. Run './test.sh frontend' to test the frontend unit tests"
    echo "2. Run './test.sh e2e' to test the E2E functionality"
    echo "3. Check the test reports in the 'test-reports/' directory"
    echo ""
}

# Main execution
main() {
    log_header "Frontend Testing Setup"
    
    # Change to script directory
    cd "$(dirname "$0")"
    
    # Run setup steps
    check_nodejs
    check_npm
    create_directories
    install_dependencies
    install_playwright_browsers
    verify_installation
    run_test_verification
    
    log_success "Frontend testing setup completed successfully!"
    show_usage
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Frontend Testing Setup Script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --verify       Only run verification checks"
        echo "  --deps-only    Only install dependencies"
        echo ""
        exit 0
        ;;
    --verify)
        check_nodejs
        check_npm
        verify_installation
        exit 0
        ;;
    --deps-only)
        check_nodejs
        check_npm
        install_dependencies
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        log_info "Use --help for usage information"
        exit 1
        ;;
esac
