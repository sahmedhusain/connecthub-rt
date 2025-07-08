# Real-Time Forum Web Test Dashboard

## Overview

The Web Test Dashboard is a comprehensive, real-time testing interface for the Real-Time Forum project. It provides an intuitive web-based interface for running tests, monitoring progress, viewing results, and analyzing coverage data.

## Features

### 🚀 **Real-Time Test Execution**
- **Live Progress Monitoring**: Real-time progress bars and status updates
- **WebSocket Integration**: Instant updates without page refresh
- **Parallel Test Execution**: Support for concurrent test runs
- **Interactive Controls**: Start, stop, and monitor tests from the web interface

### 📊 **Comprehensive Reporting**
- **Visual Coverage Reports**: Interactive charts and graphs
- **Test Result Visualization**: Color-coded status indicators
- **Export Functionality**: JSON, CSV, and HTML export options
- **Historical Data**: Track test results over time

### 🎯 **Advanced Features**
- **Filter and Search**: Find specific test categories or results
- **Responsive Design**: Works on desktop, tablet, and mobile devices
- **Keyboard Shortcuts**: Quick access to common actions
- **Notification System**: Real-time alerts and status updates

### 🔧 **Developer Tools**
- **API Integration**: RESTful API for programmatic access
- **WebSocket Events**: Real-time event streaming
- **Coverage Analysis**: Detailed code coverage metrics
- **Performance Metrics**: Test execution timing and resource usage

## Quick Start

### Prerequisites
- Go 1.19 or higher
- Modern web browser (Chrome, Firefox, Safari, Edge)
- Network access to localhost

### Launch Dashboard

#### Option 1: Using the Main Runner (Recommended)
```bash
# From project root
./run.sh test-web
```

#### Option 2: Direct Launch
```bash
# From project root
cd unit-testing/web-dashboard
./start-dashboard.sh
```

#### Option 3: Manual Setup
```bash
# From project root
cd unit-testing/web-dashboard

# Build and start API server
go mod tidy
go build -o api-server api-server.go
./api-server

# Open browser to http://localhost:8082
```

### Dashboard URL
Once started, the dashboard is available at:
```
http://localhost:8082
```

## Architecture

### Components

#### **Frontend (HTML/CSS/JavaScript)**
- **index.html**: Main dashboard interface
- **dashboard.js**: Core JavaScript functionality with WebSocket support
- **styles.css**: Responsive CSS with modern design
- **Chart.js**: Interactive coverage visualization

#### **Backend (Go API Server)**
- **api-server.go**: RESTful API server with WebSocket support
- **Real-time Updates**: WebSocket-based live data streaming
- **Test Execution**: Direct integration with test runners
- **Data Management**: Test results and coverage data handling

#### **Integration**
- **Test Runners**: Integration with existing test infrastructure
- **File System**: Automatic detection of test reports and coverage data
- **Export System**: Multiple format support for data export

### API Endpoints

#### **REST API**
```
GET  /api/test-results     # Get current test results
GET  /api/coverage         # Get coverage data
POST /api/run-tests        # Start test execution
GET  /api/status           # Get server status
GET  /api/reports          # List available reports
GET  /api/export           # Export results (JSON/CSV/HTML)
```

#### **WebSocket Events**
```
test_started     # Test execution started
test_progress    # Progress updates
test_completed   # Test execution finished
data_refresh     # Data updated
initial_data     # Initial connection data
```

## Usage Guide

### Running Tests

#### **Individual Categories**
1. Click the "Run" button next to any test category
2. Monitor progress in real-time via progress bars
3. View results in the output panel
4. Check updated coverage statistics

#### **All Tests**
1. Click "Run All Tests" in the header
2. Watch as each category executes sequentially
3. Monitor overall progress and individual category status
4. Review comprehensive results and coverage

### Viewing Results

#### **Overview Cards**
- **Total Tests**: Combined count across all categories
- **Passed**: Number of successful tests
- **Failed**: Number of failed tests
- **Coverage**: Overall code coverage percentage

#### **Category Details**
- **Test Count**: Number of tests in each category
- **Duration**: Execution time for each category
- **Last Run**: Timestamp of most recent execution
- **Status**: Current state (Idle, Running, Passed, Failed)

#### **Test Output**
- **Real-time Logs**: Live test execution output
- **Formatted Results**: Color-coded success/failure indicators
- **Error Details**: Detailed failure information
- **Export Options**: Save results in multiple formats

### Coverage Analysis

#### **Visual Charts**
- **Doughnut Chart**: Overall coverage visualization
- **Interactive Elements**: Hover for detailed information
- **Real-time Updates**: Automatic refresh after test runs

#### **Detailed Metrics**
- **Lines**: Line coverage percentage
- **Functions**: Function coverage percentage
- **Branches**: Branch coverage percentage
- **Statements**: Statement coverage percentage

### Export and Reporting

#### **Export Formats**
- **JSON**: Machine-readable data format
- **CSV**: Spreadsheet-compatible format
- **HTML**: Standalone report with styling

#### **Export Process**
1. Click "Export" button in test results section
2. Select desired format from modal dialog
3. File downloads automatically
4. Reports include timestamp and comprehensive data

## Configuration

### Environment Variables
```bash
# API server port (default: 8082)
export DASHBOARD_PORT=8082

# Test timeout (default: 15m)
export TEST_TIMEOUT=15m

# Number of parallel workers (default: 4)
export TEST_WORKERS=4

# Enable debug logging
export DEBUG=true
```

### Command Line Options
```bash
./start-dashboard.sh [OPTIONS]

Options:
  --port PORT         Set API server port (default: 8082)
  --no-browser        Don't open browser automatically
  --background        Run in background mode
  --help              Show help message
```

## Keyboard Shortcuts

- **Ctrl/Cmd + R**: Refresh data
- **Ctrl/Cmd + E**: Open export modal
- **Ctrl/Cmd + Shift + Enter**: Run all tests
- **Escape**: Close modals

## Troubleshooting

### Common Issues

#### **Port Already in Use**
```bash
# Check what's using the port
lsof -i :8082

# Kill the process
kill -9 <PID>

# Or use a different port
./start-dashboard.sh --port 8083
```

#### **Go Build Errors**
```bash
# Update Go modules
go mod tidy

# Clean module cache
go clean -modcache

# Rebuild
go build -o api-server api-server.go
```

#### **WebSocket Connection Issues**
- Check firewall settings
- Verify port accessibility
- Try refreshing the browser
- Check browser console for errors

#### **Test Execution Failures**
- Verify test runners are available
- Check file permissions
- Ensure database is accessible
- Review test logs for specific errors

### Debug Mode

Enable debug logging for detailed troubleshooting:
```bash
export DEBUG=true
./start-dashboard.sh
```

## Development

### Adding New Features

#### **Frontend Enhancements**
1. Modify `dashboard.js` for new functionality
2. Update `styles.css` for styling changes
3. Extend `index.html` for new UI elements
4. Test across different browsers and devices

#### **Backend Extensions**
1. Add new endpoints to `api-server.go`
2. Implement WebSocket event handlers
3. Update data structures as needed
4. Add comprehensive error handling

#### **Integration Points**
1. Extend test runner integration
2. Add new export formats
3. Implement additional metrics
4. Create custom notification types

### Testing the Dashboard

```bash
# Test API endpoints
curl http://localhost:8082/api/status
curl http://localhost:8082/api/test-results

# Test WebSocket connection
# Use browser developer tools or WebSocket testing tools

# Test export functionality
curl "http://localhost:8082/api/export?format=json" -o results.json
```

## Performance

### Optimization Tips
- **WebSocket Connections**: Limited to prevent resource exhaustion
- **Data Caching**: Results cached for improved response times
- **Efficient Updates**: Only changed data transmitted via WebSocket
- **Resource Management**: Automatic cleanup of completed test processes

### Monitoring
- **Memory Usage**: Monitor Go process memory consumption
- **Connection Count**: Track active WebSocket connections
- **Response Times**: Monitor API endpoint performance
- **Test Execution**: Track test runner resource usage

## Security

### Considerations
- **Local Access Only**: Dashboard designed for localhost access
- **No Authentication**: Suitable for development environments only
- **File System Access**: Limited to test directories
- **Process Execution**: Restricted to test commands only

### Production Deployment
For production use, consider:
- Adding authentication and authorization
- Implementing HTTPS/WSS
- Restricting file system access
- Adding rate limiting and input validation

## Contributing

### Guidelines
1. Follow existing code style and patterns
2. Add comprehensive error handling
3. Include responsive design considerations
4. Test across multiple browsers
5. Update documentation for new features

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Update documentation
5. Submit pull request with detailed description

## License

This project is part of the Real-Time Forum and follows the same licensing terms.

---

For more information, see the [main project README](../../README.md) or the [testing documentation](../TESTING.md).
