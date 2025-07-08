# Test Runner Quick Reference

## 🚀 Quick Commands

### Basic Test Runner (`test.sh`)
```bash
./test.sh                    # Interactive menu
./test.sh all               # Run all tests
./test.sh auth --verbose    # Auth tests with details
./test.sh --coverage --html # All tests with coverage
```

### Advanced Test Runner (`test-runner.sh`)
```bash
./test-runner.sh                           # Interactive menu
./test-runner.sh all --coverage --html     # Full suite with reports
./test-runner.sh auth --race --verbose     # Auth with race detection
./test-runner.sh unit --watch              # Watch mode
./test-runner.sh all --ci --junit          # CI/CD mode
```

### Frontend Testing Commands
```bash
# Frontend unit tests
npm test                                   # All frontend tests
npm run test:dom                          # DOM manipulation tests
npm run test:websocket                    # WebSocket client tests
npm run test:auth                         # Frontend auth tests
npm run test:spa                          # SPA navigation tests

# E2E tests with Playwright
npx playwright test                       # All E2E tests
npx playwright test auth-flow.spec.js     # Auth flow tests
npx playwright test messaging.spec.js     # Messaging tests
npx playwright test --project=chromium    # Chrome only
npx playwright test --grep responsive     # Responsive tests
```

## 📋 Test Categories

### Backend Tests

| Category | Description | Command |
|----------|-------------|---------|
| `all` | Complete test suite | `./test-runner.sh all` |
| `unit` | Component testing | `./test-runner.sh unit` |
| `integration` | End-to-end workflows | `./test-runner.sh integration` |
| `auth` | Authentication & sessions | `./test-runner.sh auth` |
| `posts` | Content management | `./test-runner.sh posts` |
| `messaging` | Real-time messaging | `./test-runner.sh messaging` |
| `websocket` | WebSocket features | `./test-runner.sh websocket` |
| `database` | Data operations | `./test-runner.sh database` |
| `middleware` | Security & validation | `./test-runner.sh middleware` |
| `api` | HTTP endpoints | `./test-runner.sh api` |

### Frontend Tests

| Category | Description | Command |
|----------|-------------|---------|
| `frontend` | All frontend unit tests | `./test.sh frontend` |
| `frontend-dom` | DOM manipulation tests | `./test.sh frontend-dom` |
| `frontend-websocket` | WebSocket client tests | `./test.sh frontend-websocket` |
| `frontend-auth` | Frontend auth tests | `./test.sh frontend-auth` |
| `frontend-spa` | SPA navigation tests | `./test.sh frontend-spa` |

### End-to-End Tests

| Category | Description | Command |
|----------|-------------|---------|
| `e2e` | All E2E tests | `./test.sh e2e` |
| `e2e-auth` | E2E authentication flow | `./test.sh e2e-auth` |
| `e2e-messaging` | E2E real-time messaging | `./test.sh e2e-messaging` |
| `cross-browser` | Cross-browser testing | `./test.sh cross-browser` |
| `responsive` | Responsive design tests | `./test.sh responsive` |

## ⚙️ Common Options

| Option | Description | Example |
|--------|-------------|---------|
| `--verbose` | Detailed output | `./test-runner.sh auth --verbose` |
| `--coverage` | Coverage analysis | `./test-runner.sh all --coverage` |
| `--html` | HTML reports | `./test-runner.sh all --html` |
| `--parallel` | Parallel execution | `./test-runner.sh all --parallel` |
| `--race` | Race detection | `./test-runner.sh all --race` |
| `--benchmark` | Benchmark tests | `./test-runner.sh all --benchmark` |
| `--watch` | Auto-run on changes | `./test-runner.sh unit --watch` |
| `--ci` | CI/CD mode | `./test-runner.sh all --ci` |
| `--timeout 5m` | Set timeout | `./test-runner.sh all --timeout 5m` |

## 📊 Report Locations

```
test-reports/     # Test execution reports
coverage/         # Coverage analysis
logs/            # Execution logs
```

## 🔧 Development Workflow

### 1. Quick Test During Development
```bash
./test-runner.sh unit --watch --verbose
```

### 2. Pre-Commit Testing
```bash
./test-runner.sh all --coverage --race
```

### 3. CI/CD Pipeline
```bash
./test-runner.sh all --ci --coverage --junit --parallel
```

### 4. Performance Analysis
```bash
./test-runner.sh all --benchmark --profile
```

## 🐛 Debugging

### View Test List
```bash
./test-runner.sh --list
```

### Run Single Test
```bash
go test ./tests/ -run "TestSpecificFunction" -v
```

### Debug Mode
```bash
./test-runner.sh all --debug --verbose
```

### Dry Run
```bash
./test-runner.sh all --dry-run
```

## 📈 Coverage Thresholds

- **Minimum**: 70%
- **Target**: 80%
- **Excellent**: 90%+

## ⏱️ Typical Execution Times

| Category | Time | Tests |
|----------|------|-------|
| `unit` | ~2-3 min | ~150 tests |
| `auth` | ~1-2 min | ~40 tests |
| `posts` | ~2-3 min | ~60 tests |
| `messaging` | ~3-4 min | ~50 tests |
| `websocket` | ~4-5 min | ~40 tests |
| `database` | ~3-4 min | ~70 tests |
| `middleware` | ~2-3 min | ~45 tests |
| `integration` | ~5-7 min | ~30 tests |
| `all` | ~8-12 min | ~400+ tests |

## 🚨 Common Issues & Solutions

### Database Locked
```bash
# Clean test environment
./test-runner.sh --clean
```

### Timeout Errors
```bash
# Increase timeout
./test-runner.sh all --timeout 15m
```

### Memory Issues
```bash
# Reduce parallel workers
./test-runner.sh all --workers 1
```

### Race Conditions
```bash
# Enable race detection
./test-runner.sh all --race
```

## 🎯 Best Practices

### ✅ Do
- Run tests before committing
- Use appropriate test categories
- Enable coverage analysis
- Check race conditions
- Use watch mode during development

### ❌ Don't
- Skip integration tests
- Ignore coverage reports
- Run all tests unnecessarily
- Forget to clean up test data
- Disable parallel execution without reason

## 🔗 Integration Examples

### GitHub Actions
```yaml
- run: ./test-runner.sh all --ci --coverage --junit
```

### GitLab CI
```yaml
script:
  - ./test-runner.sh all --ci --coverage --junit
```

### Local Development
```bash
# Start development session
./test-runner.sh unit --watch &
./run.sh native --test-data
```

## 📞 Getting Help

```bash
./test-runner.sh --help     # Full help
./test-runner.sh --version  # Version info
./test.sh --help           # Basic runner help
```

## 🎨 Interactive Menus

Both test runners provide interactive menus:
- Test category selection
- Configuration options
- Report viewing
- Development tools
- System information

Simply run without arguments:
```bash
./test-runner.sh
# or
./test.sh
```

---

**💡 Tip**: Use the advanced test runner (`test-runner.sh`) for development and CI/CD, and the basic runner (`test.sh`) for quick testing.
