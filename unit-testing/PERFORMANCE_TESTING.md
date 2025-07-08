# Performance and Load Testing Guide

## Overview

This guide covers the comprehensive performance and load testing infrastructure for the Real-Time Forum project. The testing suite includes benchmarks, load tests, stress tests, and WebSocket performance tests.

## Table of Contents

1. [Test Types](#test-types)
2. [Quick Start](#quick-start)
3. [Test Configuration](#test-configuration)
4. [Running Tests](#running-tests)
5. [Performance Metrics](#performance-metrics)
6. [Interpreting Results](#interpreting-results)
7. [Performance Thresholds](#performance-thresholds)
8. [Troubleshooting](#troubleshooting)

## Test Types

### 1. Go Benchmark Tests (`performance_test.go`)

**Purpose**: Measure performance of individual functions and operations

**Tests Include**:
- `BenchmarkUserRegistration`: User registration performance
- `BenchmarkUserLogin`: Login authentication performance
- `BenchmarkPostCreation`: Post creation performance
- `BenchmarkPostRetrieval`: Post retrieval performance
- `BenchmarkCommentCreation`: Comment creation performance
- `BenchmarkMessageSending`: Messaging performance
- `BenchmarkDatabaseOperations`: Database operation performance

**Usage**:
```bash
# Run all benchmarks
go test -bench=. -benchmem ./unit-testing/

# Run specific benchmark
go test -bench=BenchmarkUserRegistration -benchmem ./unit-testing/

# Run with multiple iterations
go test -bench=. -benchmem -count=5 ./unit-testing/
```

### 2. Load Tests (`load-test-tool.go`)

**Purpose**: Test system behavior under expected load conditions

**Features**:
- Multiple concurrent users
- Realistic user scenarios
- Configurable test duration
- Multiple output formats (JSON, CSV, HTML)
- Detailed performance metrics

**Scenarios**:
- Homepage browsing
- User authentication (login/signup)
- Content viewing (posts, comments)
- Content creation (posts, comments)
- Real-time messaging
- API endpoint testing

**Usage**:
```bash
# Build load test tool
go build -o load-test-tool load-test-tool.go

# Run basic load test
./load-test-tool --url http://localhost:8080 --users 50 --duration 30s

# Run with specific scenarios
./load-test-tool --scenarios "login,posts,messaging" --users 100

# Generate HTML report
./load-test-tool --format html --output load-test-report.html
```

### 3. Stress Tests (`stress_test.go`)

**Purpose**: Determine system limits and breaking points

**Tests Include**:
- `TestStressUserRegistration`: Registration under extreme load
- `TestStressWebSocketConnections`: WebSocket connection limits
- `TestStressDatabaseOperations`: Database performance limits
- `TestStressMemoryUsage`: Memory usage under load

**Features**:
- Gradual load increase
- System resource monitoring
- Failure threshold detection
- Memory leak detection

**Usage**:
```bash
# Run all stress tests
go test -run="TestStress" -timeout=15m ./unit-testing/

# Run specific stress test
go test -run="TestStressUserRegistration" -v ./unit-testing/
```

### 4. WebSocket Performance Tests (`websocket_performance_test.go`)

**Purpose**: Test real-time communication performance

**Tests Include**:
- `TestWebSocketPerformance`: Overall WebSocket performance
- `TestWebSocketConcurrentConnections`: Connection scaling
- `TestWebSocketMessageThroughput`: Message throughput testing

**Metrics**:
- Connection success rate
- Message latency
- Throughput (messages/second)
- Connection establishment time
- Error rates

**Usage**:
```bash
# Run WebSocket performance tests
go test -run="TestWebSocket" -v ./unit-testing/
```

## Quick Start

### 1. Run All Performance Tests

```bash
# Use the comprehensive runner
./run-performance-tests.sh

# Or run specific test types
./run-performance-tests.sh --no-stress-tests --duration 60s
```

### 2. Run Individual Test Categories

```bash
# Benchmarks only
go test -bench=. -benchmem ./unit-testing/

# Load tests only
./load-test-tool --users 50 --duration 30s

# Stress tests only
go test -run="TestStress" ./unit-testing/

# WebSocket tests only
go test -run="TestWebSocket" ./unit-testing/
```

## Test Configuration

### Load Test Configuration (`load-test-config.json`)

```json
{
  "base_url": "http://localhost:8080",
  "concurrent_users": 50,
  "duration": "30s",
  "requests_per_user": 100,
  "ramp_up_time": "5s",
  "think_time": "100ms",
  "test_scenarios": ["all"],
  "output_format": "json"
}
```

### Performance Profiles

#### Light Load
- **Users**: 10
- **Duration**: 15s
- **Use Case**: Development testing

#### Medium Load
- **Users**: 50
- **Duration**: 30s
- **Use Case**: Staging environment testing

#### Heavy Load
- **Users**: 200
- **Duration**: 60s
- **Use Case**: Production capacity testing

#### Stress Load
- **Users**: 500+
- **Duration**: 120s
- **Use Case**: Breaking point analysis

## Running Tests

### Prerequisites

1. **Go 1.19+** installed
2. **Server running** on target URL
3. **Test database** available
4. **Sufficient system resources**

### Environment Setup

```bash
# Set environment variables
export TEST_DB_PATH=":memory:"
export PERFORMANCE_TEST_DURATION="30s"
export PERFORMANCE_TEST_USERS="50"

# Create reports directory
mkdir -p performance-reports
```

### Automated Testing

```bash
# Run comprehensive performance test suite
./run-performance-tests.sh

# Custom configuration
./run-performance-tests.sh \
  --duration 60s \
  --users 100 \
  --url http://localhost:8080 \
  --verbose
```

### Manual Testing

```bash
# 1. Start the server
./run.sh --port 8080 &

# 2. Run load tests
./load-test-tool \
  --url http://localhost:8080 \
  --users 50 \
  --duration 30s \
  --scenarios all \
  --format json \
  --output results.json

# 3. Run benchmarks
go test -bench=. -benchmem -count=3 ./unit-testing/

# 4. Run stress tests
go test -run="TestStress" -timeout=15m ./unit-testing/
```

## Performance Metrics

### Key Metrics

#### Response Time Metrics
- **Average Latency**: Mean response time
- **P50 Latency**: 50th percentile (median)
- **P95 Latency**: 95th percentile
- **P99 Latency**: 99th percentile
- **Min/Max Latency**: Fastest and slowest responses

#### Throughput Metrics
- **Requests per Second (RPS)**: Request handling capacity
- **Messages per Second**: WebSocket message throughput
- **Concurrent Users**: Maximum simultaneous users

#### Reliability Metrics
- **Success Rate**: Percentage of successful requests
- **Error Rate**: Percentage of failed requests
- **Connection Success Rate**: WebSocket connection success
- **Uptime**: System availability during tests

#### Resource Metrics
- **Memory Usage**: RAM consumption
- **CPU Usage**: Processor utilization
- **Connection Pool**: Database connection usage
- **Goroutine Count**: Go routine monitoring

### Benchmark Output Example

```
BenchmarkUserRegistration-8     1000    1205847 ns/op    2048 B/op    15 allocs/op
BenchmarkUserLogin-8            2000     856234 ns/op    1024 B/op    10 allocs/op
BenchmarkPostCreation-8          500    2456789 ns/op    4096 B/op    25 allocs/op
```

**Interpretation**:
- `1000`: Number of iterations
- `1205847 ns/op`: Nanoseconds per operation
- `2048 B/op`: Bytes allocated per operation
- `15 allocs/op`: Memory allocations per operation

## Interpreting Results

### Good Performance Indicators

✅ **Response Times**:
- Average latency < 500ms
- P95 latency < 2s
- P99 latency < 5s

✅ **Throughput**:
- RPS > 100 for typical workloads
- Linear scaling with user increase
- Stable performance over time

✅ **Reliability**:
- Success rate > 99%
- Error rate < 1%
- No memory leaks

✅ **Resource Usage**:
- Memory usage stable
- CPU usage < 80%
- Efficient resource utilization

### Warning Signs

⚠️ **Performance Issues**:
- Increasing latency over time
- Decreasing throughput
- High error rates
- Memory leaks

⚠️ **Scalability Issues**:
- Non-linear performance degradation
- Connection failures under load
- Database connection pool exhaustion

## Performance Thresholds

### Response Time Thresholds

| Metric | Excellent | Good | Acceptable | Poor |
|--------|-----------|------|------------|------|
| Average Latency | < 200ms | < 500ms | < 1s | > 1s |
| P95 Latency | < 500ms | < 1s | < 2s | > 2s |
| P99 Latency | < 1s | < 2s | < 5s | > 5s |

### Throughput Thresholds

| Metric | Excellent | Good | Acceptable | Poor |
|--------|-----------|------|------------|------|
| RPS | > 500 | > 200 | > 100 | < 100 |
| Concurrent Users | > 1000 | > 500 | > 200 | < 200 |
| Messages/Second | > 1000 | > 500 | > 200 | < 200 |

### Reliability Thresholds

| Metric | Excellent | Good | Acceptable | Poor |
|--------|-----------|------|------------|------|
| Success Rate | > 99.9% | > 99% | > 95% | < 95% |
| Error Rate | < 0.1% | < 1% | < 5% | > 5% |
| Uptime | > 99.9% | > 99% | > 95% | < 95% |

## Troubleshooting

### Common Issues

#### High Latency
**Symptoms**: Slow response times, timeouts
**Causes**: Database queries, network issues, resource contention
**Solutions**:
- Optimize database queries
- Add database indexes
- Increase connection pool size
- Check network configuration

#### Low Throughput
**Symptoms**: Low RPS, poor scalability
**Causes**: Blocking operations, resource limits, inefficient algorithms
**Solutions**:
- Use async operations
- Optimize critical paths
- Increase worker pools
- Profile CPU usage

#### Memory Leaks
**Symptoms**: Increasing memory usage over time
**Causes**: Unclosed connections, goroutine leaks, large object retention
**Solutions**:
- Profile memory usage
- Check connection cleanup
- Monitor goroutine count
- Use memory profiling tools

#### Connection Failures
**Symptoms**: WebSocket connection errors, database connection errors
**Causes**: Connection limits, network issues, resource exhaustion
**Solutions**:
- Increase connection limits
- Implement connection pooling
- Add retry logic
- Monitor resource usage

### Debugging Tools

#### Go Profiling
```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./unit-testing/
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./unit-testing/
go tool pprof mem.prof
```

#### Load Test Debugging
```bash
# Verbose output
./load-test-tool --verbose --users 10 --duration 10s

# Detailed logging
export DEBUG=true
./load-test-tool --users 50 --duration 30s
```

#### System Monitoring
```bash
# Monitor system resources
top -p $(pgrep forum)
htop
iostat -x 1
netstat -an | grep :8080
```

### Performance Optimization Tips

1. **Database Optimization**:
   - Add appropriate indexes
   - Optimize query patterns
   - Use connection pooling
   - Consider read replicas

2. **Application Optimization**:
   - Profile critical paths
   - Optimize memory allocations
   - Use efficient data structures
   - Implement caching

3. **WebSocket Optimization**:
   - Limit concurrent connections
   - Implement message queuing
   - Use connection pooling
   - Add rate limiting

4. **Infrastructure Optimization**:
   - Scale horizontally
   - Use load balancers
   - Optimize network configuration
   - Monitor resource usage

## Reports and Analysis

Performance test results are saved in multiple formats:

- **JSON**: Machine-readable data for analysis
- **CSV**: Spreadsheet-compatible format
- **HTML**: Visual reports with charts and graphs
- **Text**: Console output and logs

Reports include:
- Executive summary
- Detailed metrics
- Performance trends
- Recommendations
- Comparison with thresholds

For more information, see the [main testing documentation](TESTING.md) or [API documentation](../documentations/api-documentation.md).
