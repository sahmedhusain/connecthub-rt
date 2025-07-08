# Real-Time Forum Testing Guide

## Overview

The Real-Time Forum application includes a comprehensive testing suite with professional-grade test runners and extensive coverage across all application components.

## Test Infrastructure

### Test Runners

1. **`test.sh`** - Basic interactive test runner
2. **`test-runner.sh`** - Advanced test runner with professional features
3. **`run.sh`** - Enhanced application runner with testing integration

### Test Organization

```
tests/
├── test_helper.go          # Core testing utilities
├── fixtures.go             # Test data and fixtures
├── http_helper.go          # HTTP testing utilities
├── websocket_helper.go     # WebSocket testing framework
├── auth_test.go            # Authentication tests
├── post_test.go            # Post and content tests
├── comment_test.go         # Comment system tests
├── messaging_test.go       # Messaging and conversations
├── websocket_test.go       # Real-time features
├── database_test.go        # Database operations
├── middleware_test.go      # Security and middleware
└── integration_test.go     # End-to-end workflows
```

## Test Categories

### 1. Authentication Tests (`auth`)
- User registration and validation
- Login/logout functionality
- Session management
- Password security
- User data validation

### 2. Post & Comment Tests (`posts`)
- Post creation and validation
- Content filtering and categorization
- Comment system functionality
- User interactions

### 3. Messaging Tests (`messaging`)
- Conversation creation
- Message sending and retrieval
- Read status management
- User participation validation

### 4. WebSocket Tests (`websocket`)
- Real-time connections
- Typing indicators
- Online status updates
- Message broadcasting

### 5. Database Tests (`database`)
- CRUD operations
- Data integrity
- Foreign key constraints
- Transaction handling

### 6. Middleware Tests (`middleware`)
- Authentication middleware
- Security validation
- Input sanitization
- Error handling

### 7. Integration Tests (`integration`)
- Complete user workflows
- End-to-end scenarios
- API integration
- Cross-component testing

## Frontend Testing

### 8. Frontend Unit Tests (`frontend`)
- JavaScript unit tests for all frontend components
- DOM manipulation testing
- Utility function testing
- Component interaction testing

### 9. DOM Manipulation Tests (`frontend-dom`)
- UI interaction testing
- Form validation and submission
- Dynamic content rendering
- CSS class manipulation

### 10. WebSocket Client Tests (`frontend-websocket`)
- WebSocket connectivity testing
- Real-time message handling
- Connection recovery testing
- Client-side event management

### 11. Frontend Authentication Tests (`frontend-auth`)
- Login/signup form testing
- Session management testing
- Authentication state handling
- Form validation testing

### 12. SPA Navigation Tests (`frontend-spa`)
- Single-page application routing
- Navigation state management
- Link interception testing
- Browser history management

## End-to-End Testing

### 13. All E2E Tests (`e2e`)
- Complete user workflow testing
- Cross-browser compatibility
- Real user interaction simulation
- Full application integration

### 14. E2E Authentication Flow (`e2e-auth`)
- Complete login/signup workflows
- Session persistence testing
- Authentication error handling
- User state management

### 15. E2E Real-time Messaging (`e2e-messaging`)
- Real-time chat functionality
- Multi-user interaction testing
- WebSocket connection testing
- Message delivery verification

### 16. Cross-browser Testing (`cross-browser`)
- Chrome, Firefox, Safari testing
- Browser compatibility verification
- Feature support testing
- Performance across browsers

### 17. Responsive Design Tests (`responsive`)
- Mobile viewport testing
- Tablet viewport testing
- Desktop viewport testing
- Responsive layout verification

## Quick Start

### Basic Usage

```bash
# Interactive mode
./test.sh

# Run all tests
./test.sh all

# Run specific category
./test.sh auth --verbose

# Run with coverage
./test.sh all --coverage --html

# Frontend tests
./test.sh frontend
./test.sh frontend-dom --verbose
./test.sh frontend-websocket

# E2E tests
./test.sh e2e
./test.sh e2e-auth
./test.sh cross-browser
```

### Frontend Testing Setup

Before running frontend tests, ensure you have the required dependencies:

```bash
# Install Node.js dependencies
cd unit-testing
npm install

# Install Playwright browsers for E2E tests
npx playwright install
```

### Frontend Test Commands

```bash
# All frontend unit tests
npm test

# Specific frontend test categories
npm run test:dom
npm run test:websocket
npm run test:auth
npm run test:spa

# E2E tests with Playwright
npx playwright test
npx playwright test auth-flow.spec.js
npx playwright test messaging.spec.js

# Cross-browser testing
npx playwright test --project=chromium --project=firefox --project=webkit

# Responsive design testing
npx playwright test --grep responsive
```

### Advanced Usage

```bash
# Advanced test runner
./test-runner.sh

# Full test suite with all features
./test-runner.sh all --coverage --html --parallel --race

# CI/CD mode
./test-runner.sh all --ci --junit --coverage

# Development mode with watch
./test-runner.sh unit --watch --verbose

# Benchmark testing
./test-runner.sh all --benchmark --profile
```

## Configuration

### test-config.json

The test runner uses a JSON configuration file for advanced settings:

```json
{
  "test_categories": {
    "auth": {
      "name": "Authentication Tests",
      "pattern": "./tests/ -run 'TestAuth|TestUser'",
      "timeout": "3m",
      "parallel": true
    }
  },
  "coverage": {
    "threshold": 80,
    "exclude_patterns": ["*/vendor/*", "*_test.go"]
  },
  "performance": {
    "parallel_workers": 4,
    "timeout_multiplier": 1.5
  }
}
```

### Environment Variables

```bash
export TEST_VERBOSE=true
export TEST_COVERAGE=true
export TEST_PARALLEL=true
export TEST_TIMEOUT=10m
```

## Test Features

### Coverage Analysis
- Line coverage reporting
- HTML coverage reports
- Coverage thresholds
- Exclude/include patterns

### Parallel Execution
- Configurable worker count
- Category-specific parallel settings
- Race condition detection
- Resource management

### CI/CD Integration
- JUnit XML output
- JSON reporting
- GitHub Actions support
- GitLab CI compatibility

### Development Tools
- Watch mode for auto-testing
- Single test execution
- File-specific testing
- Benchmark testing

## Writing Tests

### Test Structure

```go
func TestUserAuthentication(t *testing.T) {
    testDB := TestSetup(t)
    
    t.Run("ValidLogin", func(t *testing.T) {
        // Test implementation
        user, err := authenticateUser("username", "password")
        AssertNoError(t, err, "Authentication should succeed")
        AssertEqual(t, user.Username, "username", "Username should match")
    })
}
```

### Test Helpers

```go
// Database setup
testDB := TestSetup(t)

// Create test users
userIDs, err := SetupTestUsers(testDB.DB)

// HTTP testing
httpHelper := NewHTTPTestHelper(handler)
resp, err := httpHelper.GET("/api/users", nil)

// WebSocket testing
wsHelper := NewWebSocketTestHelper()
conn, err := wsHelper.ConnectUser(userID, sessionToken)
```

### Assertions

```go
AssertNoError(t, err, "Operation should succeed")
AssertError(t, err, "Operation should fail")
AssertEqual(t, expected, actual, "Values should match")
AssertTrue(t, condition, "Condition should be true")
AssertStatusCode(t, resp, http.StatusOK)
```

## Reports and Analysis

### Generated Reports

1. **Test Reports** (`test-reports/`)
   - Execution summaries
   - Pass/fail statistics
   - Error details

2. **Coverage Reports** (`coverage/`)
   - HTML coverage visualization
   - Function-level coverage
   - Package summaries

3. **Performance Reports**
   - Benchmark results
   - Memory profiling
   - CPU profiling

### Report Formats

- **Text**: Human-readable summaries
- **JSON**: Machine-readable data
- **HTML**: Interactive visualizations
- **JUnit XML**: CI/CD integration

## Best Practices

### Test Organization
- Group related tests in the same file
- Use descriptive test names
- Implement table-driven tests for multiple scenarios
- Separate unit and integration tests

### Test Data
- Use fixtures for consistent test data
- Clean up after each test
- Isolate test databases
- Use realistic test scenarios

### Performance
- Run tests in parallel when possible
- Use appropriate timeouts
- Monitor resource usage
- Profile slow tests

### CI/CD Integration
- Use machine-readable output formats
- Set appropriate coverage thresholds
- Cache dependencies
- Parallelize test execution

## Troubleshooting

### Common Issues

1. **Database Lock Errors**
   ```bash
   # Solution: Ensure proper test isolation
   ./test-runner.sh database --verbose
   ```

2. **Timeout Errors**
   ```bash
   # Solution: Increase timeout
   ./test-runner.sh all --timeout 15m
   ```

3. **Race Conditions**
   ```bash
   # Solution: Enable race detection
   ./test-runner.sh all --race
   ```

4. **Memory Issues**
   ```bash
   # Solution: Run tests sequentially
   ./test-runner.sh all --workers 1
   ```

### Debug Mode

```bash
# Enable verbose debugging
./test-runner.sh all --debug --verbose

# Dry run to see commands
./test-runner.sh all --dry-run

# List available tests
./test-runner.sh --list
```

## Advanced Features

### Watch Mode
Automatically re-run tests when files change:
```bash
./test-runner.sh unit --watch
```

### Profiling
Generate performance profiles:
```bash
./test-runner.sh all --profile
```

### Custom Patterns
Run specific test patterns:
```bash
go test ./tests/ -run "TestUser.*Creation"
```

### Benchmark Comparison
Compare benchmark results:
```bash
./test-runner.sh all --benchmark > old.txt
# Make changes
./test-runner.sh all --benchmark > new.txt
benchcmp old.txt new.txt
```

## Integration with IDEs

### VS Code
Add to `.vscode/tasks.json`:
```json
{
  "label": "Run Tests",
  "type": "shell",
  "command": "./test-runner.sh",
  "args": ["all", "--coverage", "--html"],
  "group": "test"
}
```

### GoLand
Configure test runner in Run/Debug Configurations with custom arguments.

## Continuous Integration

### GitHub Actions
```yaml
- name: Run Tests
  run: ./test-runner.sh all --ci --coverage --junit
  
- name: Upload Coverage
  uses: codecov/codecov-action@v1
  with:
    file: ./coverage/coverage.out
```

### GitLab CI
```yaml
test:
  script:
    - ./test-runner.sh all --ci --coverage --junit
  artifacts:
    reports:
      junit: test-reports/junit_*.xml
      coverage_report:
        coverage_format: cobertura
        path: coverage/coverage.xml
```

This comprehensive testing infrastructure ensures high code quality, reliability, and maintainability of the Real-Time Forum application.
