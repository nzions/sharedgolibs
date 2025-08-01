# Testicle Validation and Performance Tracking

## 🔍 Pre-Execution Validation

### Go Vet Integration

Testicle automatically runs `go vet` on all packages before executing tests to catch potential issues early.

#### Default Behavior
```bash
# Automatic vet validation (default)
testicle

# Skip vet validation
testicle --no-vet
```

#### Vet Process
1. **Package Discovery**: Identify all Go packages in the test directory
2. **Parallel Execution**: Run `go vet` on multiple packages simultaneously
3. **Result Aggregation**: Collect and categorize vet findings
4. **Error Reporting**: Display vet issues with context and suggestions

#### Example Output
```
🧪 Testicle v1.0 - Validating and running tests in /tests

Validating packages... ████████████████████████████████ 100%
✅ go vet: clean (47 files checked)

Discovering tests... ████████████████████████████████ 100%
Found 47 tests in 12 packages
```

#### Vet Error Handling
```
🧪 Testicle v1.0 - Validating and running tests in /tests

Validating packages... ████████████████████████░░░░░░░ 75% (9/12)
❌ go vet: found issues in 2 packages

pkg/auth/user.go:45:2: printf: Printf format %d has arg pass.ID of wrong type string
pkg/db/connection.go:123:1: unreachable: unreachable code

⚠️  Found 2 vet issues. Fix these issues or use --no-vet to proceed anyway.

Options:
  [f] Fix issues and retry
  [c] Continue anyway (equivalent to --no-vet)
  [q] Quit
```

### Build Validation

Testicle ensures all test files compile successfully before execution.

#### Compilation Check Process
1. **Test File Discovery**: Find all `*_test.go` files
2. **Go Vet Analysis**: Run `go vet` for static analysis and best practice checks
3. **Test Compilation**: Use `go test -c` to compile tests without execution
4. **Dependency Analysis**: Map test files to their dependencies
5. **Incremental Validation**: Check only changed files when possible

#### Example Output
```
🧪 Testicle v1.0 - Validating and running tests in /tests

Validating tests... ████████████████████████████████ 100%
✅ go vet: clean (47 files checked)
✅ go test -c: all tests compile successfully (12 packages)

Discovering tests... ████████████████████████████████ 100%
```

#### Build Error Handling
```
Validating tests... ████████████████████████░░░░░░░ 75% (9/12)
❌ go test -c: compilation errors found

pkg/auth/user_test.go:45:15: undefined: UnknownFunction
pkg/api/handler_test.go:123:2: syntax error: unexpected '}', expecting expression

⚠️  Found compilation errors in 2 test files. Fix these issues or use --no-build-check to proceed anyway.

Options:
  [f] Fix issues and retry  
  [c] Continue anyway (equivalent to --no-build-check)
  [q] Quit
```

## ⏱️ Performance Tracking System

### Historical Execution Time Learning

Testicle learns from each test execution to build a performance profile and provide intelligent time estimates.

#### Data Collection
- **Test Execution Duration**: Precise timing for each test
- **Success/Failure Status**: Track performance vs outcome correlation
- **Environment Context**: Consider container vs local execution
- **System Load**: Factor in concurrent test execution impact
- **Test Dependencies**: Track setup/teardown time impact

#### Storage Schema
```json
// ~/.local/testicle/data/2f55736572732f757365722f70726f6a656374732f6d792d676f2d70726f6a656374/metrics.json OR .testicle/metrics.json
{
  "project_info": {
    "name": "my-go-project",
    "path": "/Users/user/projects/my-go-project",
    "path_encoded": "2f55736572732f757365722f70726f6a656374732f6d792d676f2d70726f6a656374",
    "go_version": "1.21.0",
    "created_at": "2025-07-31T14:30:22Z",
    "last_updated": "2025-07-31T15:45:10Z"
  },
  "executions": [
    {
      "id": "exec_20250731_143022_001",
      "test_name": "TestUserLogin",
      "package_name": "pkg/auth",
      "duration_ms": 150,
      "success": true,
      "timestamp": "2025-07-31T14:30:22Z",
      "environment": "local",
      "parallel_workers": 4,
      "go_version": "1.21.0",
      "tags": ["unit", "auth"]
    },
    {
      "id": "exec_20250731_143023_002", 
      "test_name": "TestUserAuth",
      "package_name": "pkg/auth", 
      "duration_ms": 1200,
      "success": false,
      "timestamp": "2025-07-31T14:30:23Z",
      "environment": "local",
      "parallel_workers": 4,
      "error_message": "authentication failed",
      "tags": ["unit", "auth"]
    }
  ],
  "computed_metrics": {
    "TestUserLogin": {
      "average_duration_ms": 145,
      "median_duration_ms": 148,
      "std_deviation_ms": 15,
      "min_duration_ms": 120,
      "max_duration_ms": 180,
      "execution_count": 25,
      "success_rate": 0.96,
      "trend": "stable",
      "last_execution": "2025-07-31T14:30:22Z",
      "confidence_level": 0.95
    },
    "TestUserAuth": {
      "average_duration_ms": 800,
      "median_duration_ms": 750,
      "std_deviation_ms": 200,
      "min_duration_ms": 600,
      "max_duration_ms": 1200,
      "execution_count": 18,
      "success_rate": 0.89,
      "trend": "degrading",
      "last_execution": "2025-07-31T14:30:23Z",
      "confidence_level": 0.87
    }
  },
  "retention_policy": {
    "max_age_days": 90,
    "last_cleanup": "2025-07-30T00:00:00Z",
    "records_retained": 500,
    "records_purged": 125
  }
}
```

### Data Storage Strategy

#### Storage Location Decision Tree
```
Is Container Environment?
├── Yes → Use .testicle/ in test directory
└── No → User specified --local-data?
    ├── Yes → Use .testicle/ in test directory  
    └── No → Use ~/.local/testicle/ (default)
```

#### Storage Structure
```
# Default User Directory (~/.local/testicle/)
~/.local/testicle/
├── config.yaml              # User-wide default configuration
├── data/
│   ├── 2f55736572732f757365722f70726f6a656374732f6d792d676f2d70726f6a656374/    # /Users/user/projects/my-go-project
│   │   ├── metrics.json     # Performance data for this project
│   │   ├── test-discovery.json  # Test discovery cache
│   │   └── config.json      # Project-specific config overrides
│   ├── 2f55736572732f757365722f776f726b2f636f6d70616e792d6261636b656e64/        # /Users/user/work/company-backend
│   │   ├── metrics.json
│   │   ├── test-discovery.json
│   │   └── config.json
│   └── 433a5c50726f6a656374735c4d794170705c5465737473/                    # C:\Projects\MyApp\Tests
│       ├── metrics.json
│       ├── test-discovery.json
│       └── config.json
└── logs/
    └── testicle.log

# Local Project Directory (.testicle/)
.testicle/
├── metrics.json                # Performance data for this project only
├── test-discovery.json         # Test discovery cache
├── config.yaml                 # Project-specific overrides
└── logs/
    └── debug.log
```

#### Project Isolation
Projects are isolated using:
1. **Hex Encoding**: Hexadecimal encoding of absolute test directory path
2. **Git Repository**: Git remote URL (if available)
3. **Go Module**: Module name from go.mod (if available)

**Path Encoding Examples:**
```
/Users/user/projects/my-go-project → 2f55736572732f757365722f70726f6a656374732f6d792d676f2d70726f6a656374
/home/dev/work/backend-service    → 2f686f6d652f6465762f776f726b2f6261636b656e642d73657276696365
C:\Projects\MyApp\Tests           → 433a5c50726f6a656374735c4d794170705c5465737473
```

**Encoding Benefits:**
- **Fully Reversible**: Perfect 1:1 mapping between path and encoded filename
- **Filesystem Safe**: Only uses 0-9 and a-f characters
- **Cross-Platform**: Works on all filesystems (no special characters)
- **No Length Issues**: Hex encoding doesn't add problematic padding
- **Easy Debugging**: Can be decoded with any hex-to-text tool

**Utility Commands:**
```bash
# Show data location for current directory
testicle --data-info

# List all tracked projects
testicle --list-projects

# Decode a project path from hex filename
testicle --decode-path "2f55736572732f757365722f70726f6a656374732f6d792d676f2d70726f6a656374"
# Output: /Users/user/projects/my-go-project

# Clean up orphaned data files
testicle --cleanup-data
```

#### Data Migration
When switching storage modes:
```bash
# Migrate from user directory to local
testicle --migrate-to-local

# Migrate from local to user directory  
testicle --migrate-to-user

# Show current data location
testicle --data-info
```
#### Statistical Analysis
```go
type TestMetrics struct {
    TestName           string        `json:"test_name"`
    AverageDuration    time.Duration `json:"average_duration"`
    MedianDuration     time.Duration `json:"median_duration"`
    StandardDeviation  time.Duration `json:"standard_deviation"`
    MinDuration        time.Duration `json:"min_duration"`
    MaxDuration        time.Duration `json:"max_duration"`
    ExecutionCount     int           `json:"execution_count"`
    SuccessRate        float64       `json:"success_rate"`
    TrendDirection     string        `json:"trend"` // "improving", "stable", "degrading"
    ConfidenceLevel    float64       `json:"confidence_level"`
    LastExecution      time.Time     `json:"last_execution"`
}

type PerformanceStore struct {
    ProjectInfo     ProjectInfo         `json:"project_info"`
    Executions      []TestExecution     `json:"executions"`
    ComputedMetrics map[string]*TestMetrics `json:"computed_metrics"`
    RetentionPolicy RetentionPolicy     `json:"retention_policy"`
}

type ProjectInfo struct {
    Name        string    `json:"name"`
    Path        string    `json:"path"`
    PathEncoded string    `json:"path_encoded"`  // Hex-encoded path
    GoVersion   string    `json:"go_version"`
    GoModule    string    `json:"go_module"`     // from go.mod
    GitRemote   string    `json:"git_remote"`    // for additional isolation
    CreatedAt   time.Time `json:"created_at"`
    LastUpdated time.Time `json:"last_updated"`
}

type TestExecution struct {
    ID              string    `json:"id"`
    TestName        string    `json:"test_name"`
    PackageName     string    `json:"package_name"`
    DurationMs      int64     `json:"duration_ms"`
    Success         bool      `json:"success"`
    Timestamp       time.Time `json:"timestamp"`
    Environment     string    `json:"environment"`
    ParallelWorkers int       `json:"parallel_workers"`
    GoVersion       string    `json:"go_version"`
    Tags            []string  `json:"tags,omitempty"`
    ErrorMessage    string    `json:"error_message,omitempty"`
}
```

### Real-Time Estimation

#### Progress Indicators with Time Estimates
```
Running tests... ████████████████████████████░░░░░ 89% (42/47) [~2m 30s remaining]

✅ pkg/auth/TestUserLogin          (150ms) [avg: 145ms ±15ms]
✅ pkg/auth/TestPasswordValidation (89ms)  [avg: 92ms ±8ms]
🏃 pkg/api/TestCreateUser          [████████░░] 80% [~200ms remaining]
⏳ pkg/db/TestConnection           (queued) [~1.2s based on history]
❌ pkg/cache/TestRedisConnection   (2.1s) - connection failed [expected: ~800ms]
```

#### Estimation Algorithm
```go
type TimeEstimator struct {
    metrics     map[string]*TestMetrics
    queuedTests []string
    runningTests map[string]time.Time
}

func (e *TimeEstimator) EstimateRemainingTime() time.Duration {
    var totalEstimated time.Duration
    
    // Add time for queued tests
    for _, testName := range e.queuedTests {
        if metric, exists := e.metrics[testName]; exists {
            // Use 95th percentile for conservative estimate
            totalEstimated += metric.MedianDuration + metric.StandardDeviation
        } else {
            // New test - use package average or global average
            totalEstimated += e.getDefaultEstimate(testName)
        }
    }
    
    // Add remaining time for running tests
    for testName, startTime := range e.runningTests {
        elapsed := time.Since(startTime)
        if metric, exists := e.metrics[testName]; exists {
            remaining := metric.MedianDuration - elapsed
            if remaining > 0 {
                totalEstimated += remaining
            }
        }
    }
    
    return totalEstimated
}
```

### Performance Analysis and Alerts

#### Regression Detection
```
Performance Summary:
• 34 tests faster than average
• 8 tests within normal range  
• 3 tests slower than expected (potential regression)
• 2 new tests (no historical data)

⚠️  Performance Regressions Detected:
  pkg/db/TestConnection: 2.1s (expected: ~800ms) - 262% slower
  pkg/auth/TestEncryption: 450ms (expected: ~200ms) - 225% slower
  
🔍 Recommendations:
  • Check for new dependencies or external service delays
  • Review recent changes to pkg/db and pkg/auth packages
  • Consider profiling these tests with go tool pprof
```

#### Automatic Timeout Adjustment
```
⏰ Timeout Recommendations:
  pkg/slow/TestLongOperation: Current timeout 30s, suggest 45s (based on 95th percentile)
  pkg/integration/TestAPI: Current timeout 10s, suggest 15s (recent trend shows slower execution)
  
To apply these recommendations:
  testicle --update-timeouts
```

### Performance Insights Dashboard

#### CLI Performance Summary
```bash
testicle --performance-report
```

```
🧪 Testicle Performance Report - Last 30 Days

Test Suite Overview:
• Total Executions: 1,247
• Average Suite Duration: 3m 42s
• Fastest Suite: 2m 15s
• Slowest Suite: 8m 31s
• Success Rate: 94.7%

Top 10 Slowest Tests:
1. pkg/integration/TestFullWorkflow     (avg: 12.4s ±2.1s)
2. pkg/db/TestMigrations               (avg: 8.7s ±1.4s)
3. pkg/external/TestAPIIntegration     (avg: 6.2s ±3.8s)
4. pkg/encryption/TestLargeFile        (avg: 4.1s ±0.7s)
5. pkg/cache/TestWarmup                (avg: 3.9s ±0.5s)

Performance Trends:
• 🟢 12 tests showing improvement (>10% faster)
• 🟡 3 tests showing degradation (>20% slower)
• 🔴 1 test showing significant regression (>50% slower)

Flaky Tests (high variance):
• pkg/network/TestTimeout (CV: 45%) - Consider reviewing test stability
• pkg/concurrent/TestRaceCondition (CV: 38%) - Potential race conditions

Recommendations:
• Consider parallelizing pkg/integration/TestFullWorkflow
• Review pkg/external/TestAPIIntegration for external dependency issues
• Investigate pkg/network/TestTimeout for non-deterministic behavior
```

### Configuration Options

#### Performance Tracking Configuration
```yaml
# testicle.yaml
performance:
  track_execution_times: true
  metrics_file: ".testicle/metrics.db"
  history_retention: "90d"
  
  # Statistical thresholds
  variance_threshold: 2.0       # Flag tests >2x standard deviation
  regression_threshold: 1.5     # Flag tests >1.5x their average
  flaky_test_threshold: 0.3     # Flag tests with CV >30%
  
  # Time estimation
  estimation:
    enabled: true
    confidence_interval: 0.95
    minimum_samples: 3          # Need 3+ runs for reliable estimates
    use_percentile: 75          # Use 75th percentile for estimates
  
  # Performance alerts
  alerts:
    regression_detection: true
    timeout_suggestions: true
    flaky_test_detection: true
    performance_trends: true
```

#### Validation Configuration
```yaml
# testicle.yaml  
validation:
  # Go vet settings
  run_vet: true
  vet_flags: ["-composites=false"]  # Additional go vet flags
  vet_timeout: "30s"
  
  # Test compilation settings
  compile_check: true
  compile_timeout: "60s"
  compile_flags: ["-race"]        # Additional go test -c flags
  
  # Error handling
  continue_on_vet_errors: false
  continue_on_compile_errors: false
  interactive_error_handling: true  # Prompt user on errors
```

## 🚀 Implementation Architecture

### Validation Pipeline
```go
type ValidationPipeline struct {
    vetRunner     *VetRunner
    testCompiler  *TestCompiler
    config        *ValidationConfig
}

func (p *ValidationPipeline) Validate(ctx context.Context, packages []string) error {
    // Run go vet validation
    if p.config.RunVet {
        if err := p.vetRunner.ValidatePackages(ctx, packages); err != nil {
            return fmt.Errorf("vet validation failed: %w", err)
        }
    }
    
    // Run test compilation check
    if p.config.CompileCheck {
        if err := p.testCompiler.CompileTests(ctx, packages); err != nil {
            return fmt.Errorf("test compilation failed: %w", err)
        }
    }
    
    return nil
}
```

### Test Compiler
```go
type TestCompiler struct {
    tempDir string
    verbose bool
}

func NewTestCompiler(verbose bool) *TestCompiler {
    return &TestCompiler{
        tempDir: os.TempDir(),
        verbose: verbose,
    }
}

func (tc *TestCompiler) CompileTests(ctx context.Context, packages []string) error {
    for _, pkg := range packages {
        if err := tc.compilePackageTests(ctx, pkg); err != nil {
            return err
        }
    }
    return nil
}

func (tc *TestCompiler) compilePackageTests(ctx context.Context, pkg string) error {
    // Create temporary output file
    tempFile := filepath.Join(tc.tempDir, fmt.Sprintf("testicle-test-%d", time.Now().UnixNano()))
    defer os.Remove(tempFile)
    
    // Run go test -c to compile tests without running them
    cmd := exec.CommandContext(ctx, "go", "test", "-c", "-o", tempFile, pkg)
    
    if tc.verbose {
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
    } else {
        // Capture output for error reporting
        var stderr bytes.Buffer
        cmd.Stderr = &stderr
        
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("compilation failed for %s: %s", pkg, stderr.String())
        }
    }
    
    return cmd.Run()
}
```

### Performance Tracking Engine
```go
type PerformanceTracker struct {
    store     *JSONPerformanceStore
    config    *PerformanceConfig
    estimator *TimeEstimator
    analyzer  *PerformanceAnalyzer
}

type JSONPerformanceStore struct {
    filePath    string
    data        *PerformanceStore
    mutex       sync.RWMutex
    autoSave    bool
    saveInterval time.Duration
}

func NewPerformanceTracker(config *PerformanceConfig) (*PerformanceTracker, error) {
    storePath := determineStoragePath(config)
    
    store, err := NewJSONPerformanceStore(storePath)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize performance store: %w", err)
    }
    
    tracker := &PerformanceTracker{
        store:  store,
        config: config,
    }
    
    tracker.estimator = NewTimeEstimator(tracker)
    tracker.analyzer = NewPerformanceAnalyzer(tracker)
    
    return tracker, nil
}

func determineStoragePath(config *PerformanceConfig) string {
    switch config.StorageLocation {
    case "local":
        return filepath.Join(config.TestDirectory, ".testicle", "metrics.json")
    case "user":
        return filepath.Join(getUserDataDir(), "testicle", "metrics", getProjectHash(config.TestDirectory)+".json")
    case "auto":
        if isContainerEnvironment() || config.ForceLocalData {
            return filepath.Join(config.TestDirectory, ".testicle", "metrics.json")
        }
        return filepath.Join(getUserDataDir(), "testicle", "metrics", getProjectHash(config.TestDirectory)+".json")
    default:
        return config.MetricsFile
    }
}

func (p *PerformanceTracker) RecordExecution(execution *TestExecution) error {
    return p.store.AddExecution(execution)
}

func (p *PerformanceTracker) GetTestMetrics(testName string) (*TestMetrics, error) {
    return p.store.GetTestMetrics(testName)
}

// JSONPerformanceStore implementation
func NewJSONPerformanceStore(filePath string) (*JSONPerformanceStore, error) {
    store := &JSONPerformanceStore{
        filePath:     filePath,
        autoSave:     true,
        saveInterval: 30 * time.Second,
    }
    
    if err := store.load(); err != nil {
        if os.IsNotExist(err) {
            // Initialize new store
            store.data = &PerformanceStore{
                ProjectInfo: ProjectInfo{
                    Name:        filepath.Base(filepath.Dir(filePath)),
                    Path:        filepath.Dir(filePath),
                    Hash:        getProjectHash(filepath.Dir(filePath)),
                    GoVersion:   runtime.Version(),
                    CreatedAt:   time.Now(),
                    LastUpdated: time.Now(),
                },
                Executions:      make([]TestExecution, 0),
                ComputedMetrics: make(map[string]*TestMetrics),
                RetentionPolicy: RetentionPolicy{
                    MaxAgeDays: 90,
                    LastCleanup: time.Now(),
                },
            }
            return store, store.save()
        }
        return nil, err
    }
    
    return store, nil
}

func (s *JSONPerformanceStore) load() error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    data, err := os.ReadFile(s.filePath)
    if err != nil {
        return err
    }
    
    s.data = &PerformanceStore{}
    return json.Unmarshal(data, s.data)
}

func (s *JSONPerformanceStore) save() error {
    if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
        return err
    }
    
    data, err := json.MarshalIndent(s.data, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(s.filePath, data, 0644)
}

func (s *JSONPerformanceStore) AddExecution(execution *TestExecution) error {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    
    // Add execution to history
    s.data.Executions = append(s.data.Executions, *execution)
    s.data.ProjectInfo.LastUpdated = time.Now()
    
    // Recompute metrics for this test
    s.recomputeMetrics(execution.TestName)
    
    // Clean old executions if needed
    s.cleanupOldExecutions()
    
    if s.autoSave {
        return s.save()
    }
    
    return nil
}

func (s *JSONPerformanceStore) GetTestMetrics(testName string) (*TestMetrics, error) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    
    metrics, exists := s.data.ComputedMetrics[testName]
    if !exists {
        return nil, nil
    }
    
    return metrics, nil
}

func (s *JSONPerformanceStore) recomputeMetrics(testName string) {
    var executions []TestExecution
    
    // Filter executions for this test
    for _, exec := range s.data.Executions {
        if exec.TestName == testName {
            executions = append(executions, exec)
        }
    }
    
    if len(executions) == 0 {
        return
    }
    
    // Calculate metrics
    metrics := &TestMetrics{
        TestName:       testName,
        ExecutionCount: len(executions),
    }
    
    // Calculate durations and success rate
    var durations []time.Duration
    var successCount int
    
    for _, exec := range executions {
        duration := time.Duration(exec.DurationMs) * time.Millisecond
        durations = append(durations, duration)
        
        if exec.Success {
            successCount++
        }
        
        if metrics.LastExecution.IsZero() || exec.Timestamp.After(metrics.LastExecution) {
            metrics.LastExecution = exec.Timestamp
        }
    }
    
    // Calculate statistical measures
    metrics.SuccessRate = float64(successCount) / float64(len(executions))
    metrics.MinDuration = minDuration(durations)
    metrics.MaxDuration = maxDuration(durations)
    metrics.AverageDuration = averageDuration(durations)
    metrics.MedianDuration = medianDuration(durations)
    metrics.StandardDeviation = standardDeviation(durations, metrics.AverageDuration)
    
    // Calculate trend and confidence
    metrics.TrendDirection = calculateTrend(executions)
    metrics.ConfidenceLevel = calculateConfidence(len(executions), metrics.StandardDeviation, metrics.AverageDuration)
    
    s.data.ComputedMetrics[testName] = metrics
}
```
```go
type PerformanceTracker struct {
    db          *sql.DB
    estimator   *TimeEstimator
    analyzer    *PerformanceAnalyzer
    config      *PerformanceConfig
}

func (p *PerformanceTracker) RecordExecution(result *TestResult) error {
    return p.db.Exec(`
        INSERT INTO test_executions 
        (test_name, package_name, duration_ms, success, environment, parallel_workers)
        VALUES (?, ?, ?, ?, ?, ?)
    `, result.TestName, result.Package, result.Duration.Milliseconds(), 
       result.Success, result.Environment, result.ParallelWorkers)
}

func (p *PerformanceTracker) GetTestMetrics(testName string) (*TestMetrics, error) {
    // Retrieve and calculate statistics for a specific test
    // Implementation includes average, median, std dev, trend analysis
}
```

This comprehensive validation and performance tracking system ensures that testicle not only runs tests effectively but also helps developers understand and optimize their test suite performance over time.
