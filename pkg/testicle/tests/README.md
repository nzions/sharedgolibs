# Testicle Development Test Suite

This directory contains a comprehensive test suite designed to validate that Testicle handles all types of Go tests correctly.

## Test Categories

### `basic/` - Basic Test Patterns
- **Purpose**: Simple, fast-running tests that should always pass
- **Contents**: Basic assertions, string operations, slice operations, map operations, struct operations, channel operations, error handling
- **Usage**: `testicle --dir ./tests/basic`

### `failures/` - Failure Scenarios  
- **Purpose**: Tests designed to fail in various ways to verify error handling
- **Contents**: Assertion failures, multiple failures, string mismatches, slice mismatches, fatal errors
- **Usage**: `testicle --dir ./tests/failures` (expect failures)

### `timeout_test.go` - Long-running and Hanging Tests
- **Purpose**: Tests that take a long time or hang to verify timeout handling
- **Contents**: Long-running tests, hanging tests, deadlocks, context timeouts, busy loops, memory leaks
- **Usage**: `testicle --timeout 30s --dir ./tests/failures`

### `benchmarks/` - Performance Tests
- **Purpose**: Benchmarks and performance-sensitive tests
- **Contents**: String concatenation, integer conversion, slice operations, map operations, sorting, hashing, memory allocation
- **Usage**: `testicle --benchmark --dir ./tests/benchmarks`

### `subtests/` - Complex Test Structures
- **Purpose**: Tests with subtests, table-driven patterns, and nested structures
- **Contents**: Table-driven tests, nested subtests, complex structures, dynamic subtests, parallel subtests
- **Usage**: `testicle --dir ./tests/subtests`

### `edge_cases/` - Edge Cases and Special Scenarios
- **Purpose**: Malformed tests, empty files, and unusual conditions
- **Contents**: Empty tests, skipped tests, special characters in names, very long test names
- **Usage**: `testicle --dir ./tests/edge_cases`

## Running Tests

### Full Development Test Suite
```bash
# Run all tests
testicle --dir ./tests

# Run with debug output
testicle --debug --dir ./tests

# Run in watch mode for development
testicle --daemon --dir ./tests
```

### Category-specific Testing
```bash
# Test basic functionality
testicle --dir ./tests/basic

# Test error handling (expect failures)
testicle --dir ./tests/failures

# Run benchmarks
testicle --benchmark --dir ./tests/benchmarks

# Test complex structures
testicle --dir ./tests/subtests

# Test edge cases
testicle --dir ./tests/edge_cases
```

### Validation Testing
```bash
# Test with validation enabled (default)
testicle --dir ./tests/basic

# Test without go vet
testicle --no-vet --dir ./tests/basic

# Test without compilation check
testicle --no-build-check --dir ./tests/basic

# Test with both validations disabled
testicle --no-vet --no-build-check --dir ./tests/basic
```

### Performance Testing
```bash
# Profile memory usage
testicle --profile=mem --dir ./tests/basic

# Profile CPU usage
testicle --profile=cpu --dir ./tests/benchmarks

# Compare with native go test
time go test ./tests/basic/...
time testicle --dir ./tests/basic
```

## Expected Outcomes

### `basic/` Tests
- ✅ All tests should pass
- ✅ Fast execution (< 1 second)
- ✅ No validation errors

### `failures/` Tests  
- ❌ Multiple test failures expected
- ⚠️ Some tests may panic or hang
- 🔄 Graceful error handling verification

### `benchmarks/` Tests
- 📊 Benchmark results displayed
- ⏱️ Performance metrics collected
- 📈 Baseline establishment

### `subtests/` Tests
- 🌳 Nested test structure displayed
- ✅ All subtests should pass
- 🔢 Multiple test counts

### `edge_cases/` Tests
- ⏭️ Some tests skipped
- 📝 Log-only tests
- 🏷️ Special character handling

## Development Workflow

1. **Make changes** to Testicle code
2. **Run basic tests** to verify core functionality:
   ```bash
   testicle --dir ./tests/basic
   ```
3. **Test error handling** with failure tests:
   ```bash
   testicle --dir ./tests/failures
   ```
4. **Validate performance** with benchmarks:
   ```bash
   testicle --benchmark --dir ./tests/benchmarks
   ```
5. **Test edge cases** to ensure robustness:
   ```bash
   testicle --dir ./tests/edge_cases
   ```

## Adding New Test Cases

When adding new test scenarios:

1. **Identify the category** that best fits your test case
2. **Create test file** in the appropriate subdirectory
3. **Follow Go conventions** for test function names and structure
4. **Document expected behavior** in test comments
5. **Test with Testicle** to ensure proper handling

## CI/CD Integration

This test suite is designed for use in automated testing:

```yaml
# Example CI configuration
- name: Run Testicle Development Tests
  run: |
    ./bin/testicle --dir ./tests/basic
    ./bin/testicle --dir ./tests/benchmarks
    ./bin/testicle --dir ./tests/subtests
    ./bin/testicle --dir ./tests/edge_cases
    # Expect failures but verify graceful handling
    ./bin/testicle --dir ./tests/failures || true
```

The test suite validates:
- ✅ **Functional correctness** - All test types execute properly
- ✅ **Error handling** - Failures are handled gracefully  
- ✅ **Performance** - No significant regression in execution time
- ✅ **Robustness** - Edge cases and special conditions work correctly
