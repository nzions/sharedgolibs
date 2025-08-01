# Development Guide 👨‍💻

This guide provides comprehensive information for developing and testing the Testicle test runner.

## Development Environment Setup

### Prerequisites
- Go 1.21 or later
- Git
- Docker (optional, for container testing)

### Local Development
```bash
# Clone the repository
git clone https://github.com/nzions/sharedgolibs.git
cd sharedgolibs/pkg/testicle

# Install dependencies
go mod download

# Build the CLI
go build -o bin/testicle ./cmd/testicle

# Run tests
go test ./...
```

## Development Test Suite

To ensure Testicle handles all test scenarios correctly, we've created a comprehensive test suite in the `tests/` directory. This suite includes various test types that developers commonly encounter.

### Test Directory Structure

```
tests/
├── basic/
│   ├── basic_test.go           # Simple passing tests
│   └── helper.go               # Test helpers
├── failures/
│   ├── assertion_test.go       # Tests with assertion failures
│   ├── panic_test.go           # Tests that panic
│   └── timeout_test.go         # Long-running/hanging tests
├── benchmarks/
│   ├── cpu_test.go             # CPU-intensive benchmarks
│   └── memory_test.go          # Memory allocation benchmarks
├── subtests/
│   ├── table_driven_test.go    # Table-driven tests with subtests
│   └── nested_test.go          # Nested subtests
├── setup/
│   ├── setup_test.go           # Tests with TestMain setup
│   └── fixture_test.go         # Tests requiring file fixtures
├── parallel/
│   ├── parallel_test.go        # Tests that run in parallel
│   └── race_test.go            # Tests that expose race conditions
├── skip/
│   ├── conditional_test.go     # Tests that skip conditionally
│   └── build_tags_test.go      # Tests with build tags
├── external/
│   ├── network_test.go         # Tests requiring network
│   ├── database_test.go        # Tests requiring database
│   └── filesystem_test.go      # Tests requiring file system
└── edge_cases/
    ├── empty_test.go           # Empty test file
    ├── malformed_test.go       # Tests with syntax errors
    └── import_cycle_test.go    # Tests with import issues
```

## Test Categories

### 1. Basic Tests (`tests/basic/`)
Simple, fast-running tests that should always pass.

### 2. Failure Tests (`tests/failures/`)
Tests designed to fail in various ways to verify error handling.

### 3. Performance Tests (`tests/benchmarks/`)
Benchmarks and performance-sensitive tests.

### 4. Complex Structure Tests (`tests/subtests/`)
Tests with subtests, table-driven patterns, and nested structures.

### 5. Setup/Teardown Tests (`tests/setup/`)
Tests requiring initialization, cleanup, or external resources.

### 6. Parallel Tests (`tests/parallel/`)
Tests designed to run concurrently and expose race conditions.

### 7. Conditional Tests (`tests/skip/`)
Tests that may be skipped based on environment or build tags.

### 8. External Dependency Tests (`tests/external/`)
Tests requiring network, database, or filesystem access.

### 9. Edge Cases (`tests/edge_cases/`)
Malformed tests, empty files, and error conditions.

## Running Development Tests

### Full Test Suite
```bash
# Run all development tests
testicle --dir ./tests

# Run with debug output
testicle --debug --dir ./tests

# Run in watch mode
testicle --daemon --dir ./tests
```

### Specific Test Categories
```bash
# Run only basic tests
testicle --dir ./tests/basic

# Run failure tests (expect failures)
testicle --dir ./tests/failures

# Run benchmarks
testicle --dir ./tests/benchmarks --benchmark
```

### Testing Edge Cases
```bash
# Test malformed Go files
testicle --dir ./tests/edge_cases

# Test with compilation errors
testicle --no-build-check --dir ./tests/edge_cases
```

## Performance Testing

### Baseline Measurements
Run performance tests to establish baseline metrics:

```bash
# Measure testicle overhead vs go test
time go test ./tests/basic/...
time testicle --dir ./tests/basic

# Memory profiling
testicle --profile=mem --dir ./tests/basic

# CPU profiling  
testicle --profile=cpu --dir ./tests/basic
```

### Stress Testing
```bash
# Large test suite simulation
testicle --dir ./tests --parallel=10

# Watch mode stability
testicle --daemon --dir ./tests
# (then make rapid file changes)
```

## Development Workflow

### 1. Feature Development
```bash
# Create feature branch
git checkout -b feature/new-feature

# Run development tests frequently
testicle --daemon --dir ./tests

# Check specific scenarios
testicle --dir ./tests/failures  # Test error handling
testicle --dir ./tests/parallel  # Test race conditions
```

### 2. Testing Changes
```bash
# Validate all test types work
./scripts/run-dev-tests.sh

# Test container compatibility
docker build -t testicle:dev .
docker run -v $(pwd)/tests:/tests testicle:dev
```

### 3. Performance Validation
```bash
# Ensure no performance regression
./scripts/benchmark.sh

# Validate memory usage
./scripts/memory-check.sh
```

## CI/CD Integration

### Automated Testing
The development test suite is used in CI to validate:

1. **Functional Correctness**: All test types execute properly
2. **Error Handling**: Failures are handled gracefully
3. **Performance**: No significant regression in execution time
4. **Container Compatibility**: Works in containerized environments

### Example CI Configuration
```yaml
# .github/workflows/test.yml
name: Test Testicle
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build testicle
        run: go build -o bin/testicle ./cmd/testicle
      
      - name: Run development test suite
        run: |
          ./bin/testicle --dir ./tests/basic
          ./bin/testicle --dir ./tests/benchmarks
          ./bin/testicle --dir ./tests/parallel
          
      - name: Test failure handling
        run: |
          # Expect failures but verify graceful handling
          ./bin/testicle --dir ./tests/failures || true
          
      - name: Container test
        run: |
          docker build -t testicle:test .
          docker run -v $(pwd)/tests:/tests testicle:test
```

## Debugging

### Debug Mode
```bash
# Enable verbose debugging
testicle --debug --dir ./tests

# Debug specific components
export TESTICLE_DEBUG=discovery,execution,validation
testicle --dir ./tests
```

### Common Issues

#### Performance Regression
```bash
# Compare with baseline
testicle --benchmark --dir ./tests/benchmarks
testicle --profile=cpu --dir ./tests/basic
```

#### Race Conditions
```bash
# Run with race detection
testicle --race --dir ./tests/parallel
```

#### Memory Leaks
```bash
# Monitor memory usage
testicle --profile=mem --dir ./tests
```

## Contributing

### Adding New Test Cases
1. Identify the test category
2. Create test file in appropriate `tests/` subdirectory
3. Follow Go testing conventions
4. Document expected behavior
5. Test with Testicle to ensure proper handling

### Documentation
- Keep [README.md](./README.md) as the single source for overview
- Update specific guides for detailed information
- Cross-reference related documentation

See [CLI_REFERENCE.md](./CLI_REFERENCE.md) for complete flag documentation and [TECHNICAL_SPEC.md](./TECHNICAL_SPEC.md) for architecture details.
