# Testicle CLI Reference

## 🚀 Command Line Interface

### Basic Usage

```bash
testicle [flags] [command]
```

### Core Flags

#### `--debug`
Enable debug output for troubleshooting and development.

```bash
testicle --debug
```

**Debug Output Includes:**
- Detailed file change events
- Test discovery process details
- Execution timing and resource usage
- Event bus messages
- Configuration loading details
- Container detection information

**Example Debug Output:**
```
[DEBUG] 2025-07-31T14:30:22Z Container environment detected
[DEBUG] 2025-07-31T14:30:22Z Test directory: /tests
[DEBUG] 2025-07-31T14:30:22Z Config loaded from: /config/testicle.yaml
[DEBUG] 2025-07-31T14:30:23Z File watcher started for: /tests/**/*_test.go
[DEBUG] 2025-07-31T14:30:24Z Discovered 47 test functions in 12 packages
[DEBUG] 2025-07-31T14:30:24Z Starting test execution with 4 workers
```

#### `--daemon`, `-d`
Run in daemon mode - continuously watch test directory and execute tests as files change.

```bash
testicle --daemon
# or
testicle -d
```

**Daemon Mode Features:**
- **Continuous Watching** - Monitors test directory for file changes
- **Automatic Test Execution** - Runs affected tests when files change
- **Live Console Output** - Real-time streaming of test results
- **Intelligent Debouncing** - Avoids running tests on rapid file changes
- **Graceful Shutdown** - Handles SIGINT/SIGTERM properly

**Daemon Mode Behavior:**
```bash
🧪 Testicle Daemon - Watching /tests for changes...

[14:30:22] File changed: pkg/auth/user.go
[14:30:22] Discovering affected tests...
[14:30:23] Running 3 affected tests...
[14:30:23] ✅ TestUserValidation (89ms)
[14:30:24] ✅ TestUserSerialization (142ms) 
[14:30:25] ❌ TestUserAuth (1.2s) - authentication failed

[14:32:15] File changed: pkg/auth/user_test.go
[14:32:15] Running 1 test...
[14:32:16] ✅ TestUserAuth (156ms)

Ready - watching for changes... (Ctrl+C to stop)
[r] Run Now | [s] Stop | [p] Pause | [ESC] Resume | [q] Quit
```

**Interactive Controls in Daemon Mode:**
- **`r`** - Run tests immediately (bypass file change trigger)
- **`s`** - Stop currently running tests gracefully
- **`p`** - Pause file watching (tests won't auto-run)
- **`ESC`** - Resume file watching when paused
- **`q`** - Quit Testicle
- **`d`** - Toggle debug output on/off
- **`v`** - Toggle verbose test output
- **`c`** - Clear screen and refresh display
- **`h`** - Show help with all key bindings

#### `--dir <path>`
Specify the test directory to monitor (default: `/tests` in container, `.` locally).

```bash
testicle --dir ./my-tests
testicle --dir /custom/test/path
```

**Directory Detection:**
- **Container Mode**: Defaults to `/tests` (auto-detected)
- **Local Mode**: Defaults to current directory `.`
- **Validation**: Ensures directory exists and is readable
- **Recursive**: Scans subdirectories for test files

#### `--config <location>`
Specify configuration file location for tuning settings.

```bash
testicle --config ./custom-testicle.yaml
testicle --config /config/testicle.yaml
```

**Configuration Priority (highest to lowest):**
1. Command-line flags
2. Configuration file specified by `--config`
3. Local `testicle.yaml` (if exists)
4. Environment variables
5. Built-in defaults

## 📋 Complete Flag Reference

### Primary Flags

| Flag               | Short | Default                             | Description                          |
| ------------------ | ----- | ----------------------------------- | ------------------------------------ |
| `--debug`          |       | `false`                             | Enable debug output                  |
| `--daemon`         | `-d`  | `false`                             | Run in daemon/watch mode             |
| `--dir`            |       | `/tests` (container)<br>`.` (local) | Test directory path                  |
| `--config`         |       | `testicle.yaml`                     | Configuration file location          |
| `--no-vet`         |       | `false`                             | Skip `go vet` validation             |
| `--no-build-check` |       | `false`                             | Skip test compilation validation     |
| `--reset-metrics`  |       | `false`                             | Clear historical execution time data |
| `--local-data`     |       | `false`                             | Force local data storage in test dir |

### Execution Control

| Flag         | Default | Description                     |
| ------------ | ------- | ------------------------------- |
| `--timeout`  | `30s`   | Default test timeout            |
| `--parallel` | `4`     | Number of parallel test workers |
| `--verbose`  | `false` | Verbose test output             |
| `--quiet`    | `false` | Minimal output mode             |

### Output Control

| Flag         | Default             | Description                      |
| ------------ | ------------------- | -------------------------------- |
| `--output`   | `test-results.json` | Output file path                 |
| `--format`   | `json`              | Output format (json, xml, junit) |
| `--no-color` | `false`             | Disable colored output           |

### Filtering

| Flag                  | Description                        |
| --------------------- | ---------------------------------- |
| `--package <pattern>` | Run tests matching package pattern |
| `--test <pattern>`    | Run tests matching name pattern    |
| `--tag <tag>`         | Run tests with specific tag        |
| `--exclude <pattern>` | Exclude tests matching pattern     |

### Utility Flags

| Flag        | Description              |
| ----------- | ------------------------ |
| `--version` | Show version information |
| `--help`    | Show help message        |

### Validation and Performance Flags

#### `--no-vet`
Skip `go vet` validation before running tests. By default, testicle runs `go vet` on all packages to catch potential issues.

```bash
testicle --no-vet
```

**Use Cases:**
- When you know vet issues exist but want to run tests anyway
- In CI environments where vet is run separately
- Performance optimization for trusted code

#### `--no-build-check`
Skip test compilation validation before running tests. By default, testicle uses `go test -c` to ensure all test files compile before execution.

```bash
testicle --no-build-check
```

**Use Cases:**
- When compilation is verified elsewhere
- Performance optimization for large codebases
- Emergency debugging situations

#### `--reset-metrics`
Clear all historical test execution time data and start fresh performance tracking.

```bash
testicle --reset-metrics
```

#### `--local-data`
Force data storage in the test directory (`.testicle/`) instead of the default user directory (`~/.local/testicle/`).

```bash
testicle --local-data
```

**Data Storage Locations:**
- **Default (Local)**: `~/.local/testicle/` - User-wide data directory
- **Container Mode**: `.testicle/` in test directory (automatic)
- **Forced Local**: `.testicle/` in test directory (with `--local-data`)

**Storage Structure:**
```
~/.local/testicle/           # Default location (local runs)
├── config.yaml              # User-wide default configuration
├── data/
│   ├── 2f55736572732f757365722f70726f6a656374732f6d792d676f2d70726f6a656374/
│   │   ├── metrics.json     # Performance data for this project
│   │   ├── test-discovery.json  # Test discovery cache
│   │   └── config.json      # Project-specific overrides
│   └── 2f55736572732f757365722f776f726b2f636f6d70616e792d6261636b656e64/
│       ├── metrics.json
│       ├── test-discovery.json
│       └── config.json
└── logs/
    └── testicle.log

.testicle/                   # Test directory storage (containers + --local-data)
├── metrics.json             # Performance data for this project
├── cache.json               # Test discovery cache
└── config.yaml              # Project-specific config
```

**Performance Tracking Features:**
- **Historical Averages**: Tracks execution time over multiple runs
- **Variance Analysis**: Identifies tests with inconsistent performance
- **Regression Detection**: Flags tests that are significantly slower than usual
- **Time Estimation**: Predicts total test suite duration based on history
- **Automatic Timeout Adjustment**: Suggests optimal timeout values

**Metrics Storage:**
- Data stored in JSON format for easy inspection and portability
- Tracks: test name, duration, timestamp, success/failure
- Calculates: average duration, standard deviation, recent trends
- Project isolation using hex path encoding (fully reversible, filesystem-safe)

**Example Output with Metrics:**
```
✅ pkg/auth/TestUserLogin          (150ms) [avg: 145ms ±15ms]
⚠️  pkg/slow/TestConnection        (2.1s) [regression: was ~800ms]
🆕 pkg/new/TestFeature             (234ms) [no historical data]
```

## 🔧 Configuration File

### Default Configuration (`testicle.yaml`)

```yaml
# Test execution settings
execution:
  test_directory: "/tests"  # Auto-detected based on environment
  timeout:
    default: "30s"
    per_test: {}  # Override timeouts for specific tests
  parallel: 4
  verbose: false

# Daemon/watch mode settings
daemon:
  enabled: false
  debounce_delay: "500ms"
  patterns:
    include:
      - "**/*_test.go"
      - "**/*.go"  # Watch source files too
    exclude:
      - "**/vendor/**"
      - "**/.git/**"
      - "**/node_modules/**"
  auto_discover: true

# Output settings
output:
  file: "test-results.json"
  format: "json"  # json, xml, junit
  include_output: true
  include_timing: true
  directory: "/output"  # Container output directory

# Validation settings
validation:
  run_vet: true            # Run go vet before tests
  compile_check: true      # Verify test compilation with go test -c
  vet_flags: []            # Additional flags for go vet
  compile_timeout: "30s"   # Max time for test compilation validation

# Performance tracking settings
performance:
  track_execution_times: true
  storage_location: "auto"     # "auto", "user", "local" 
  metrics_file: "metrics.json" # JSON file for performance data
  history_retention: "90d"     # Keep 90 days of performance data
  variance_threshold: 2.0      # Flag tests >2x standard deviation
  regression_threshold: 1.5    # Flag tests >1.5x their average
  project_isolation: true      # Separate metrics per project
  estimation:
    enabled: true              # Show time estimates during execution
    confidence_interval: 0.95  # Statistical confidence for estimates
    minimum_samples: 3         # Need 3+ runs for reliable estimates

# Discovery settings
discovery:
  patterns:
    - "**/*_test.go"
  exclude:
    - "**/vendor/**"
    - "**/.git/**"
  tags:
    enabled: true
    default: ["unit"]

# UI settings
ui:
  colors: true  # Auto-disabled in non-TTY
  progress_style: "bar"  # bar, spinner, dots, percent
  update_interval: "100ms"
  show_package_summary: true

# Debug settings
debug:
  enabled: false
  log_file: ""  # Empty = stdout
  log_level: "info"  # debug, info, warn, error
  trace_events: false

# Container-specific settings
container:
  auto_detect: true
  mount_point: "/tests"
  config_mount: "/config"
  output_mount: "/output"
```

### Configuration Examples

#### Minimal Configuration
```yaml
# testicle-minimal.yaml
execution:
  timeout:
    default: "1m"
daemon:
  enabled: true
```

#### CI/CD Configuration
```yaml
# testicle-ci.yaml
execution:
  parallel: 8
  timeout:
    default: "10m"
daemon:
  enabled: false  # Single run in CI
output:
  format: "junit"
  file: "/output/test-results.xml"
ui:
  colors: false
debug:
  enabled: true
```

#### Development Configuration
```yaml
# testicle-dev.yaml
daemon:
  enabled: true
  debounce_delay: "200ms"  # Faster feedback
ui:
  progress_style: "spinner"
debug:
  enabled: true
  trace_events: true
```

## 🚀 Usage Examples

### Basic Test Execution

```bash
# Run all tests once
testicle

# Run with custom directory
testicle --dir ./integration-tests

# Run with debug output
testicle --debug
```

### Daemon Mode

```bash
# Start daemon with defaults
testicle --daemon

# Daemon with custom config
testicle -d --config ./testicle-dev.yaml

# Daemon with specific directory
testicle -d --dir ./my-tests --debug
```

### Container Usage

```bash
# Basic container run
docker run -v $(pwd):/tests testicle:latest

# Daemon mode in container
docker run -it -v $(pwd):/tests testicle:latest --daemon

# With custom config
docker run -v $(pwd):/tests -v $(pwd)/config:/config testicle:latest \
  --config /config/testicle.yaml

# With output persistence
docker run -v $(pwd):/tests -v $(pwd)/reports:/output testicle:latest \
  --output /output/results.json
```

### CI/CD Integration

```bash
# GitHub Actions
testicle --config .github/testicle-ci.yaml --format junit --output test-results.xml

# GitLab CI
testicle --daemon=false --parallel 8 --timeout 10m --format xml

# Jenkins
testicle --no-color --verbose --output junit-results.xml --format junit
```

### Advanced Usage

```bash
# Run specific package with timeout
testicle --package "./pkg/auth" --timeout 2m

# Run with tags and exclusions
testicle --tag integration --exclude "**/slow_test.go"

# Debug daemon mode with custom config
testicle -d --debug --config ./debug-config.yaml --dir ./tests
```

## 🔍 Environment Variables

Environment variables can override configuration settings:

```bash
# Core settings
export TESTICLE_DEBUG=true
export TESTICLE_DAEMON=true
export TESTICLE_TEST_DIR=/custom/tests
export TESTICLE_CONFIG_FILE=/path/to/config.yaml
export TESTICLE_LOCAL_DATA=true

# Execution settings
export TESTICLE_TIMEOUT=60s
export TESTICLE_PARALLEL=8
export TESTICLE_VERBOSE=true

# Output settings
export TESTICLE_OUTPUT_FILE=results.json
export TESTICLE_OUTPUT_FORMAT=junit
export TESTICLE_NO_COLOR=true

# Container settings
export TESTICLE_CONTAINER_MODE=true
export TESTICLE_MOUNT_POINT=/tests
```

## 🎯 Exit Codes

| Code  | Meaning                           |
| ----- | --------------------------------- |
| `0`   | Success - all tests passed        |
| `1`   | Test failures - some tests failed |
| `2`   | Configuration error               |
| `3`   | Test discovery error              |
| `4`   | Execution error                   |
| `5`   | Timeout error                     |
| `130` | Interrupted (Ctrl+C)              |

## 📊 Output Formats

### JSON Output (Default)
```json
{
  "session_id": "20250731-143022",
  "status": "completed",
  "summary": {
    "total": 47,
    "passed": 42,
    "failed": 3,
    "skipped": 1,
    "timeout": 1
  }
}
```

### JUnit XML Output
```xml
<testsuite name="testicle" tests="47" failures="3" errors="1" time="120.5">
  <testcase classname="pkg.auth" name="TestUserLogin" time="0.15"/>
  <testcase classname="pkg.auth" name="TestUserAuth" time="1.2">
    <failure message="authentication failed">...</failure>
  </testcase>
</testsuite>
```

This CLI reference provides comprehensive documentation for the four core flags you specified, along with additional flags that support the core functionality.
