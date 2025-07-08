# Frontend Testing Infrastructure Guide

## Overview

This document provides a comprehensive guide to the frontend testing infrastructure for the Real-Time Forum project. The testing suite includes unit tests for JavaScript components, integration tests for UI interactions, and end-to-end tests for complete user workflows.

## Architecture

### Testing Framework Stack

- **Jest**: JavaScript unit testing framework with JSDOM environment
- **Playwright**: End-to-end testing framework for cross-browser testing
- **JSDOM**: DOM simulation for unit tests
- **Node.js**: Runtime environment for test execution

### Directory Structure

```
unit-testing/
├── package.json                    # Frontend dependencies and scripts
├── playwright.config.js            # Playwright configuration
├── setup-frontend-tests.sh         # Setup script for dependencies
├── frontend-tests/
│   ├── setup/
│   │   ├── jest.setup.js           # Jest environment setup
│   │   ├── global-setup.js         # Playwright global setup
│   │   └── global-teardown.js      # Playwright global teardown
│   ├── unit/                       # Jest unit tests
│   │   ├── dom.test.js             # DOM manipulation tests
│   │   ├── websocket.test.js       # WebSocket client tests
│   │   ├── auth.test.js            # Authentication flow tests
│   │   ├── spa.test.js             # SPA navigation tests
│   │   └── validation.test.js      # Form validation tests
│   └── e2e/                        # Playwright E2E tests
│       ├── auth-flow.spec.js       # Authentication workflows
│       ├── messaging.spec.js       # Real-time messaging
│       └── responsive.spec.js      # Responsive design
├── test-reports/                   # Test execution reports
└── coverage/                       # Code coverage reports
```

## Setup Instructions

### Prerequisites

- Node.js 16+ and npm
- Go 1.19+ (for backend integration)
- Modern web browser (Chrome, Firefox, Safari)

### Quick Setup

```bash
# Navigate to testing directory
cd unit-testing

# Run automated setup
./setup-frontend-tests.sh

# Or manual setup
npm install
npx playwright install
```

### Verification

```bash
# Verify frontend unit tests
npm test

# Verify E2E tests
npx playwright test --headed

# Verify integration with test runner
./test.sh frontend-dom --verbose
```

## Test Categories

### 1. Frontend Unit Tests (`frontend`)

**Purpose**: Test JavaScript components in isolation
**Framework**: Jest with JSDOM
**Command**: `npm test` or `./test.sh frontend`

**Coverage**:
- DOM manipulation functions
- Form validation logic
- Utility functions
- Component interactions
- State management

### 2. DOM Manipulation Tests (`frontend-dom`)

**Purpose**: Test UI interactions and DOM operations
**Framework**: Jest with JSDOM
**Command**: `npm run test:dom` or `./test.sh frontend-dom`

**Coverage**:
- Dynamic content rendering
- Form interactions
- CSS class manipulation
- Event handling
- UI state changes

### 3. WebSocket Client Tests (`frontend-websocket`)

**Purpose**: Test real-time communication features
**Framework**: Jest with WebSocket mocks
**Command**: `npm run test:websocket` or `./test.sh frontend-websocket`

**Coverage**:
- WebSocket connection management
- Message sending/receiving
- Real-time UI updates
- Connection recovery
- Typing indicators

### 4. Authentication Tests (`frontend-auth`)

**Purpose**: Test authentication flows and session management
**Framework**: Jest with API mocks
**Command**: `npm run test:auth` or `./test.sh frontend-auth`

**Coverage**:
- Login/signup forms
- Session management
- Authentication state
- Form validation
- Error handling

### 5. SPA Navigation Tests (`frontend-spa`)

**Purpose**: Test single-page application routing
**Framework**: Jest with history mocks
**Command**: `npm run test:spa` or `./test.sh frontend-spa`

**Coverage**:
- Route matching
- Navigation functions
- Link interception
- State management
- Error handling

### 6. End-to-End Tests (`e2e`)

**Purpose**: Test complete user workflows
**Framework**: Playwright
**Command**: `npx playwright test` or `./test.sh e2e`

**Coverage**:
- Authentication workflows
- Real-time messaging
- Cross-browser compatibility
- Responsive design
- User interactions

## Running Tests

### Individual Test Categories

```bash
# Frontend unit tests
npm test
./test.sh frontend

# Specific categories
npm run test:dom
npm run test:websocket
npm run test:auth
npm run test:spa

# E2E tests
npx playwright test
./test.sh e2e
./test.sh e2e-auth
./test.sh e2e-messaging
```

### Cross-browser Testing

```bash
# All browsers
npx playwright test

# Specific browsers
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit

# Via test runner
./test.sh cross-browser
```

### Responsive Design Testing

```bash
# Responsive tests
npx playwright test --grep responsive
./test.sh responsive

# Specific viewports
npx playwright test responsive.spec.js
```

### Development Mode

```bash
# Watch mode for unit tests
npm run test:watch

# Debug mode for E2E tests
npx playwright test --debug
npx playwright test --headed

# Coverage reports
npm run test:coverage
```

## Integration with Existing Test Suite

### Test Runner Integration

The frontend tests are fully integrated with the existing Go test runner:

```bash
# Interactive menu includes frontend options
./test.sh

# Direct execution
./test.sh frontend-dom --verbose
./test.sh e2e-auth --coverage
```

### Configuration

Frontend test categories are defined in `test-config.json`:

```json
{
  "test_categories": {
    "frontend": {
      "name": "Frontend Unit Tests",
      "pattern": "npm test",
      "timeout": "5m",
      "type": "frontend"
    },
    "e2e": {
      "name": "End-to-End Tests",
      "pattern": "npx playwright test",
      "timeout": "15m",
      "type": "e2e"
    }
  }
}
```

### Reporting

- **HTML Reports**: Generated in `test-reports/`
- **Coverage Reports**: Available in `coverage/frontend/`
- **JUnit XML**: For CI/CD integration
- **Screenshots**: Captured on E2E test failures

## Best Practices

### Writing Unit Tests

```javascript
// Use descriptive test names
test('should validate email format correctly', () => {
  // Test implementation
});

// Group related tests
describe('Form Validation', () => {
  describe('Email Validation', () => {
    // Related tests
  });
});

// Use setup and teardown
beforeEach(() => {
  global.testUtils.cleanupDOM();
});
```

### Writing E2E Tests

```javascript
// Use page object patterns
test('should login successfully', async ({ page }) => {
  await page.goto('/');
  await page.fill('input[name="identifier"]', 'testuser');
  await page.fill('input[name="password"]', 'password');
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/.*\/home/);
});

// Handle dynamic content
await page.waitForSelector('.message');
await expect(page.locator('.message')).toBeVisible();
```

### Performance Considerations

- Use `--parallel` for faster execution
- Implement proper cleanup in teardown
- Mock external dependencies
- Use selective test execution during development

## Troubleshooting

### Common Issues

1. **Jest Configuration Errors**
   - Ensure `package.json` Jest config is correct
   - Check for conflicting global imports

2. **Playwright Browser Issues**
   - Run `npx playwright install` to update browsers
   - Check system requirements for browser support

3. **Test Timeouts**
   - Increase timeout values in configuration
   - Optimize test performance

4. **DOM Simulation Issues**
   - Verify JSDOM environment setup
   - Check for proper DOM cleanup

### Debug Commands

```bash
# Verbose test output
npm test -- --verbose

# Debug specific test
npm test -- --testNamePattern="should validate email"

# Playwright debug mode
npx playwright test --debug

# Generate test reports
npm run test:coverage
```

## CI/CD Integration

The frontend testing infrastructure is designed for CI/CD environments:

```bash
# CI mode
./test.sh all --ci --junit

# Headless E2E tests
npx playwright test --reporter=junit

# Coverage reports
npm run test:coverage
```

## Maintenance

### Updating Dependencies

```bash
# Update npm packages
npm update

# Update Playwright browsers
npx playwright install

# Verify after updates
./setup-frontend-tests.sh --verify
```

### Adding New Tests

1. Create test files in appropriate directories
2. Follow naming conventions (`*.test.js` for Jest, `*.spec.js` for Playwright)
3. Update test categories in `test-config.json` if needed
4. Document new test patterns in this guide

This comprehensive frontend testing infrastructure ensures robust testing coverage for all client-side functionality while maintaining seamless integration with the existing backend testing suite.
