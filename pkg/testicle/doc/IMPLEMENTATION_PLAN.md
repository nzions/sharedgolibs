# Testicle Implementation Plan v1.0

## 🎯 Core Requirements

### Primary Goals
- **Dual Environment Support**: Run locally or in container with `/tests` mount
- **File Watching**: Monitor test directory for changes
- **Test Discovery**: Automatic discovery on file changes
- **Visual Feedback**: Rich CLI progress indicators and status
- **Test Timeouts**: Configurable timeouts per test
- **Result Persistence**: Write test results to structured files
- **CLI First**: Command-line interface with web-ready architecture

### Design Constraints
- CLI-only for v1.0 (web interface ready for v2.0)
- Container-friendly architecture
- Minimal dependencies
- Fast startup and execution
- Structured output for tooling integration

## 🏗️ Architecture Design

### High-Level Component Structure

```
testicle/
├── cmd/testicle/           # CLI entry point
│   └── main.go            # Application bootstrap
├── internal/              # Private implementation
│   ├── app/               # Application core
│   │   ├── app.go         # Main application logic
│   │   └── config.go      # Configuration management
│   ├── discovery/         # Test discovery engine
│   │   ├── scanner.go     # File system scanning
│   │   ├── parser.go      # Go AST parsing
│   │   └── watcher.go     # File change monitoring
│   ├── runner/            # Test execution engine
│   │   ├── executor.go    # Test execution logic
│   │   ├── timeout.go     # Timeout management
│   │   └── output.go      # Output parsing
│   ├── ui/                # CLI user interface
│   │   ├── renderer.go    # Progress rendering
│   │   ├── colors.go      # Color schemes
│   │   └── spinner.go     # Loading animations
│   └── storage/           # Result persistence
│       ├── writer.go      # File output
│       └── formats.go     # Output formats (JSON, XML, etc.)
└── pkg/testicle/          # Public API (for future web interface)
    ├── types.go           # Core data types
    ├── client.go          # Client interface
    └── events.go          # Event system
```

### Data Flow Architecture

```
File Changes → Watcher → Discovery → Test Queue → Executor → Results → Storage
     ↓              ↓           ↓           ↓           ↓          ↓
   Events      File Events  Test Tree   Execution   Status    Output Files
     ↓              ↓           ↓           ↓           ↓          ↓
   UI Updates  Auto-Reload  Visual Tree  Progress   Results   Reporting
```

## 📋 Phase 1: Core Foundation (Week 1-2)

### 1.1 Project Setup and CLI Framework

```go
// cmd/testicle/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/nzions/sharedgolibs/internal/app"
)

func main() {
    var (
        testDir     = flag.String("dir", "/tests", "Test directory to monitor")
        configFile  = flag.String("config", "testicle.yaml", "Configuration file")
        watch       = flag.Bool("watch", true, "Enable file watching")
        timeout     = flag.Duration("timeout", 30*time.Second, "Default test timeout")
        outputFile  = flag.String("output", "test-results.json", "Output file for results")
        verbose     = flag.Bool("verbose", false, "Verbose output")
    )
    flag.Parse()
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        cancel()
    }()
    
    // Initialize application
    config := &app.Config{
        TestDir:    *testDir,
        ConfigFile: *configFile,
        Watch:      *watch,
        Timeout:    *timeout,
        OutputFile: *outputFile,
        Verbose:    *verbose,
    }
    
    application := app.New(config)
    if err := application.Run(ctx); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### 1.2 Configuration System

```yaml
# testicle.yaml
test_directory: "/tests"
watch_enabled: true
timeout:
  default: "30s"
  per_test: 
    slow_test: "2m"
    integration_test: "5m"
    
# Validation settings
validation:
  run_vet: true
  build_check: true
  vet_flags: []
  build_timeout: "30s"
  continue_on_errors: false

# Performance tracking
performance:
  track_execution_times: true
  storage_location: "auto"      # "auto", "user", "local"
  metrics_file: "metrics.json"  # JSON file name
  history_retention: "90d"
  variance_threshold: 2.0
  regression_threshold: 1.5
  project_isolation: true       # Separate metrics per project
  force_local_data: false       # Force .testicle/ in test dir
  estimation:
    enabled: true
    confidence_interval: 0.95
    minimum_samples: 3
    
output:
  file: "test-results.json"
  format: "json"  # json, xml, junit
  include_output: true
  include_timing: true
discovery:
  patterns:
    - "**/*_test.go"
  exclude:
    - "**/vendor/**"
    - "**/.git/**"
  tags:
    enabled: true
    default: ["unit"]
ui:
  colors: true
  progress: "bar"  # bar, spinner, dots
  update_interval: "100ms"
container:
  mode: false  # auto-detected
  mount_point: "/tests"
```

```go
// internal/app/config.go
type Config struct {
    TestDir      string           `yaml:"test_directory"`
    Watch        bool             `yaml:"watch_enabled"`
    Timeout      TimeoutConfig    `yaml:"timeout"`
    Validation   ValidationConfig `yaml:"validation"`
    Performance  PerformanceConfig `yaml:"performance"`
    Output       OutputConfig     `yaml:"output"`
    Discovery    DiscoveryConfig  `yaml:"discovery"`
    UI           UIConfig         `yaml:"ui"`
    Container    ContainerConfig  `yaml:"container"`
}

type ValidationConfig struct {
    RunVet           bool          `yaml:"run_vet"`
    BuildCheck       bool          `yaml:"build_check"`
    VetFlags         []string      `yaml:"vet_flags"`
    BuildTimeout     time.Duration `yaml:"build_timeout"`
    ContinueOnErrors bool          `yaml:"continue_on_errors"`
}

type PerformanceConfig struct {
    TrackExecutionTimes bool          `yaml:"track_execution_times"`
    StorageLocation     string        `yaml:"storage_location"`     // "auto", "user", "local"
    MetricsFile         string        `yaml:"metrics_file"`         // JSON filename
    HistoryRetention    time.Duration `yaml:"history_retention"`
    VarianceThreshold   float64       `yaml:"variance_threshold"`
    RegressionThreshold float64       `yaml:"regression_threshold"`
    ProjectIsolation    bool          `yaml:"project_isolation"`
    ForceLocalData      bool          `yaml:"force_local_data"`
    Estimation          EstimationConfig `yaml:"estimation"`
}

type EstimationConfig struct {
    Enabled            bool    `yaml:"enabled"`
    ConfidenceInterval float64 `yaml:"confidence_interval"`
    MinimumSamples     int     `yaml:"minimum_samples"`
}

type TimeoutConfig struct {
    Default time.Duration            `yaml:"default"`
    PerTest map[string]time.Duration `yaml:"per_test"`
}

type OutputConfig struct {
    File          string `yaml:"file"`
    Format        string `yaml:"format"`
    IncludeOutput bool   `yaml:"include_output"`
    IncludeTiming bool   `yaml:"include_timing"`
}
```

### 1.3 Container Detection and Environment Setup

```go
// internal/app/environment.go
type Environment struct {
    IsContainer bool
    TestDir     string
    WorkingDir  string
    GoPath      string
}

func DetectEnvironment() (*Environment, error) {
    env := &Environment{}
    
    // Detect if running in container
    if _, err := os.Stat("/.dockerenv"); err == nil {
        env.IsContainer = true
        env.TestDir = "/tests"
    } else {
        env.IsContainer = false
        env.TestDir = "."
    }
    
    // Validate test directory
    if _, err := os.Stat(env.TestDir); os.IsNotExist(err) {
        return nil, fmt.Errorf("test directory %s does not exist", env.TestDir)
    }
    
    // Find Go binary
    goPath, err := exec.LookPath("go")
    if err != nil {
        return nil, fmt.Errorf("go binary not found: %w", err)
    }
    env.GoPath = goPath
    
    return env, nil
}
```

### 1.4 Validation Pipeline

```go
// internal/validation/validator.go
type Validator struct {
    config     *ValidationConfig
    vetRunner  *VetRunner
    buildChecker *BuildChecker
    ui         UIReporter
}

type ValidationResult struct {
    VetPassed    bool          `json:"vet_passed"`
    BuildPassed  bool          `json:"build_passed"`
    VetIssues    []VetIssue    `json:"vet_issues,omitempty"`
    BuildErrors  []BuildError  `json:"build_errors,omitempty"`
    Duration     time.Duration `json:"duration"`
}

type VetIssue struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Column   int    `json:"column"`
    Message  string `json:"message"`
    Category string `json:"category"`
}

type BuildError struct {
    File    string `json:"file"`
    Line    int    `json:"line"`
    Column  int    `json:"column"`
    Message string `json:"message"`
}

func (v *Validator) ValidatePackages(ctx context.Context, packages []string) (*ValidationResult, error) {
    result := &ValidationResult{}
    start := time.Now()
    
    // Run go vet validation
    if v.config.RunVet {
        v.ui.ShowProgress("Running go vet validation...")
        vetResult, err := v.vetRunner.ValidatePackages(ctx, packages)
        if err != nil {
            return nil, fmt.Errorf("vet validation failed: %w", err)
        }
        result.VetPassed = vetResult.Passed
        result.VetIssues = vetResult.Issues
    }
    
    // Run build validation
    if v.config.BuildCheck {
        v.ui.ShowProgress("Validating test compilation...")
        buildResult, err := v.buildChecker.ValidateTests(ctx, packages)
        if err != nil {
            return nil, fmt.Errorf("build validation failed: %w", err)
        }
        result.BuildPassed = buildResult.Passed
        result.BuildErrors = buildResult.Errors
    }
    
    result.Duration = time.Since(start)
    return result, nil
}

// internal/validation/vet.go
type VetRunner struct {
    timeout   time.Duration
    vetFlags  []string
    goCommand string
}

func (v *VetRunner) ValidatePackages(ctx context.Context, packages []string) (*VetResult, error) {
    var allIssues []VetIssue
    
    for _, pkg := range packages {
        cmd := exec.CommandContext(ctx, v.goCommand, append([]string{"vet"}, v.vetFlags...)...)
        cmd.Args = append(cmd.Args, pkg)
        
        output, err := cmd.CombinedOutput()
        if err != nil {
            issues := v.parseVetOutput(string(output))
            allIssues = append(allIssues, issues...)
        }
    }
    
    return &VetResult{
        Passed: len(allIssues) == 0,
        Issues: allIssues,
    }, nil
}

// internal/validation/build.go
type BuildChecker struct {
    timeout     time.Duration
    buildFlags  []string
    goCommand   string
}

func (b *BuildChecker) ValidateTests(ctx context.Context, packages []string) (*BuildResult, error) {
    var allErrors []BuildError
    
    for _, pkg := range packages {
        cmd := exec.CommandContext(ctx, b.goCommand, "build", "-o", "/dev/null")
        cmd.Args = append(cmd.Args, b.buildFlags...)
        cmd.Args = append(cmd.Args, pkg)
        
        output, err := cmd.CombinedOutput()
        if err != nil {
            errors := b.parseBuildOutput(string(output))
            allErrors = append(allErrors, errors...)
        }
    }
    
    return &BuildResult{
        Passed: len(allErrors) == 0,
        Errors: allErrors,
    }, nil
}
```

### 1.5 Performance Tracking System

```go
// internal/performance/tracker.go
type PerformanceTracker struct {
    store     *JSONPerformanceStore
    config    *PerformanceConfig
    estimator *TimeEstimator
    analyzer  *PerformanceAnalyzer
}

type JSONPerformanceStore struct {
    filePath     string
    data         *PerformanceStore
    mutex        sync.RWMutex
    autoSave     bool
    saveInterval time.Duration
}

type PerformanceStore struct {
    ProjectInfo     ProjectInfo                `json:"project_info"`
    Executions      []TestExecution           `json:"executions"`
    ComputedMetrics map[string]*TestMetrics   `json:"computed_metrics"`
    RetentionPolicy RetentionPolicy           `json:"retention_policy"`
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
    Environment     string    `json:"environment"`  // "local", "container"
    ParallelWorkers int       `json:"parallel_workers"`
    GoVersion       string    `json:"go_version"`
    Tags            []string  `json:"tags,omitempty"`
    ErrorMessage    string    `json:"error_message,omitempty"`
}

func NewPerformanceTracker(config *PerformanceConfig, testDir string) (*PerformanceTracker, error) {
    storePath := determineStoragePath(config, testDir)
    
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

func determineStoragePath(config *PerformanceConfig, testDir string) string {
    switch config.StorageLocation {
    case "local":
        return filepath.Join(testDir, ".testicle", config.MetricsFile)
    case "user":
        userDir, _ := os.UserHomeDir()
        projectEncoded := hexEncodeProjectPath(testDir)
        return filepath.Join(userDir, ".local", "testicle", "data", projectEncoded, config.MetricsFile)
    case "auto":
        if isContainerEnvironment() || config.ForceLocalData {
            return filepath.Join(testDir, ".testicle", config.MetricsFile)
        }
        userDir, _ := os.UserHomeDir()
        projectEncoded := hexEncodeProjectPath(testDir)
        return filepath.Join(userDir, ".local", "testicle", "data", projectEncoded, config.MetricsFile)
    default:
        // Explicit file path
        if filepath.IsAbs(config.MetricsFile) {
            return config.MetricsFile
        }
        return filepath.Join(testDir, config.MetricsFile)
    }
}

func hexEncodeProjectPath(testDir string) string {
    absPath, _ := filepath.Abs(testDir)
    return hex.EncodeToString([]byte(absPath))
}

func hexDecodeProjectPath(encoded string) (string, error) {
    decoded, err := hex.DecodeString(encoded)
    if err != nil {
        return "", fmt.Errorf("invalid hex encoding: %w", err)
    }
    return string(decoded), nil
}

func isContainerEnvironment() bool {
    // Check for common container indicators
    if _, err := os.Stat("/.dockerenv"); err == nil {
        return true
    }
    if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
        return true
    }
    if os.Getenv("CONTAINER") == "true" {
        return true
    }
    return false
}

func (p *PerformanceTracker) RecordExecution(execution *TestExecution) error {
    execution.ID = generateExecutionID()
    return p.store.AddExecution(execution)
}

func (p *PerformanceTracker) GetTestMetrics(testName string) (*TestMetrics, error) {
    return p.store.GetTestMetrics(testName)
}

func generateExecutionID() string {
    now := time.Now()
    return fmt.Sprintf("exec_%s_%d", 
        now.Format("20060102_150405"), 
        now.UnixNano()%1000)
}

// internal/performance/store.go
func NewJSONPerformanceStore(filePath string) (*JSONPerformanceStore, error) {
    store := &JSONPerformanceStore{
        filePath:     filePath,
        autoSave:     true,
        saveInterval: 30 * time.Second,
    }
    
    if err := store.load(); err != nil {
        if os.IsNotExist(err) {
            // Initialize new store
            testDir := filepath.Dir(filepath.Dir(filePath))
            store.data = &PerformanceStore{
                ProjectInfo: ProjectInfo{
                    Name:        detectProjectName(testDir),
                    Path:        testDir,
                    PathEncoded: hexEncodeProjectPath(testDir),
                    GoVersion:   runtime.Version(),
                    GoModule:    detectGoModule(testDir),
                    GitRemote:   detectGitRemote(testDir),
                    CreatedAt:   time.Now(),
                    LastUpdated: time.Now(),
                },
                Executions:      make([]TestExecution, 0),
                ComputedMetrics: make(map[string]*TestMetrics),
                RetentionPolicy: RetentionPolicy{
                    MaxAgeDays:      90,
                    LastCleanup:     time.Now(),
                    RecordsRetained: 0,
                    RecordsPurged:   0,
                },
            }
            return store, store.save()
        }
        return nil, err
    }
    
    return store, nil
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
```
```

## 📋 Phase 2: File Watching and Discovery (Week 2-3)

### 2.1 File System Watcher

```go
// internal/discovery/watcher.go
type Watcher struct {
    fsWatcher   *fsnotify.Watcher
    testDir     string
    patterns    []string
    excludes    []string
    debouncer   *Debouncer
    eventChan   chan DiscoveryEvent
}

type DiscoveryEvent struct {
    Type      EventType       `json:"type"`
    Path      string          `json:"path"`
    Timestamp time.Time       `json:"timestamp"`
    Tests     []*TestFunction `json:"tests,omitempty"`
}

type EventType string

const (
    EventFileAdded    EventType = "file_added"
    EventFileModified EventType = "file_modified"
    EventFileDeleted  EventType = "file_deleted"
    EventTestsFound   EventType = "tests_discovered"
)

func NewWatcher(testDir string, config *WatcherConfig) (*Watcher, error) {
    fsWatcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    
    return &Watcher{
        fsWatcher: fsWatcher,
        testDir:   testDir,
        patterns:  config.Patterns,
        excludes:  config.Excludes,
        debouncer: NewDebouncer(config.DebounceDelay),
        eventChan: make(chan DiscoveryEvent, 100),
    }, nil
}

func (w *Watcher) Start(ctx context.Context) error {
    if err := w.addWatchPaths(); err != nil {
        return err
    }
    
    go w.eventLoop(ctx)
    return nil
}

func (w *Watcher) eventLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case event := <-w.fsWatcher.Events:
            if w.shouldProcessEvent(event) {
                w.debouncer.Trigger(event.Name, func() {
                    w.processFileChange(event)
                })
            }
        case err := <-w.fsWatcher.Errors:
            // Log error and continue
            log.Printf("Watcher error: %v", err)
        }
    }
}
```

### 2.2 Test Discovery Engine

```go
// internal/discovery/scanner.go
type Scanner struct {
    testDir   string
    patterns  []string
    excludes  []string
    parser    *Parser
}

type TestFunction struct {
    Name        string            `json:"name"`
    Package     string            `json:"package"`
    File        string            `json:"file"`
    Line        int               `json:"line"`
    Type        TestType          `json:"type"`
    Tags        []string          `json:"tags"`
    Timeout     *time.Duration    `json:"timeout,omitempty"`
    Parallel    bool              `json:"parallel"`
    Skip        bool              `json:"skip"`
    Metadata    map[string]string `json:"metadata"`
}

type TestType string

const (
    TestTypeUnit        TestType = "unit"
    TestTypeIntegration TestType = "integration"
    TestTypeBenchmark   TestType = "benchmark"
    TestTypeExample     TestType = "example"
)

func (s *Scanner) DiscoverTests() ([]*TestFunction, error) {
    var allTests []*TestFunction
    
    err := filepath.Walk(s.testDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !s.shouldScanFile(path, info) {
            return nil
        }
        
        tests, err := s.parser.ParseFile(path)
        if err != nil {
            return fmt.Errorf("parsing %s: %w", path, err)
        }
        
        allTests = append(allTests, tests...)
        return nil
    })
    
    return allTests, err
}
```

### 2.3 AST-Based Go Parser

```go
// internal/discovery/parser.go
type Parser struct {
    fileSet *token.FileSet
}

func (p *Parser) ParseFile(filePath string) ([]*TestFunction, error) {
    src, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }
    
    file, err := parser.ParseFile(p.fileSet, filePath, src, parser.ParseComments)
    if err != nil {
        return nil, err
    }
    
    var tests []*TestFunction
    
    ast.Inspect(file, func(n ast.Node) bool {
        if fn, ok := n.(*ast.FuncDecl); ok {
            if test := p.extractTestFunction(fn, filePath, file); test != nil {
                tests = append(tests, test)
            }
        }
        return true
    })
    
    return tests, nil
}

func (p *Parser) extractTestFunction(fn *ast.FuncDecl, filePath string, file *ast.File) *TestFunction {
    name := fn.Name.Name
    
    // Check if it's a test function
    if !strings.HasPrefix(name, "Test") && 
       !strings.HasPrefix(name, "Benchmark") && 
       !strings.HasPrefix(name, "Example") {
        return nil
    }
    
    pos := p.fileSet.Position(fn.Pos())
    
    test := &TestFunction{
        Name:     name,
        Package:  file.Name.Name,
        File:     filePath,
        Line:     pos.Line,
        Type:     p.determineTestType(name),
        Tags:     p.extractTags(fn.Doc),
        Parallel: p.hasParallelCall(fn),
        Skip:     p.hasSkipCall(fn),
        Metadata: make(map[string]string),
    }
    
    // Extract timeout from comments
    if timeout := p.extractTimeout(fn.Doc); timeout != nil {
        test.Timeout = timeout
    }
    
    return test
}
```

## 📋 Phase 3: Test Execution Engine (Week 3-4)

### 3.1 Test Executor with Timeout Management

```go
// internal/runner/executor.go
type Executor struct {
    testDir     string
    goPath      string
    defaultTimeout time.Duration
    maxParallel int
    resultChan  chan *TestResult
}

type TestResult struct {
    Test        *TestFunction `json:"test"`
    Status      TestStatus    `json:"status"`
    Duration    time.Duration `json:"duration"`
    StartTime   time.Time     `json:"start_time"`
    EndTime     time.Time     `json:"end_time"`
    Output      string        `json:"output"`
    Error       string        `json:"error,omitempty"`
    ExitCode    int           `json:"exit_code"`
    TimedOut    bool          `json:"timed_out"`
}

type TestStatus string

const (
    StatusPending TestStatus = "pending"
    StatusRunning TestStatus = "running"
    StatusPassed  TestStatus = "passed"
    StatusFailed  TestStatus = "failed"
    StatusSkipped TestStatus = "skipped"
    StatusTimeout TestStatus = "timeout"
)

func (e *Executor) RunTest(ctx context.Context, test *TestFunction) *TestResult {
    result := &TestResult{
        Test:      test,
        Status:    StatusRunning,
        StartTime: time.Now(),
    }
    
    // Determine timeout
    timeout := e.defaultTimeout
    if test.Timeout != nil {
        timeout = *test.Timeout
    }
    
    // Create test context with timeout
    testCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    // Build and execute command
    cmd := e.buildTestCommand(test)
    cmd.Dir = e.testDir
    
    output, err := e.runWithTimeout(testCtx, cmd)
    result.EndTime = time.Now()
    result.Duration = result.EndTime.Sub(result.StartTime)
    result.Output = string(output)
    
    // Determine result status
    if testCtx.Err() == context.DeadlineExceeded {
        result.Status = StatusTimeout
        result.TimedOut = true
        result.Error = fmt.Sprintf("Test timed out after %v", timeout)
    } else if err != nil {
        result.Status = StatusFailed
        result.Error = err.Error()
        if exitErr, ok := err.(*exec.ExitError); ok {
            result.ExitCode = exitErr.ExitCode()
        }
    } else {
        result.Status = StatusPassed
    }
    
    return result
}

func (e *Executor) buildTestCommand(test *TestFunction) *exec.Cmd {
    args := []string{
        "test",
        "-v",
        "-run", fmt.Sprintf("^%s$", test.Name),
        fmt.Sprintf("./%s", test.Package),
    }
    
    return exec.Command(e.goPath, args...)
}
```

### 3.2 Parallel Execution Manager

```go
// internal/runner/manager.go
type Manager struct {
    executor    *Executor
    maxParallel int
    semaphore   chan struct{}
    resultChan  chan *TestResult
    statusChan  chan *TestStatus
}

func (m *Manager) RunTests(ctx context.Context, tests []*TestFunction) <-chan *TestResult {
    results := make(chan *TestResult, len(tests))
    
    // Start workers
    var wg sync.WaitGroup
    for _, test := range tests {
        wg.Add(1)
        go func(t *TestFunction) {
            defer wg.Done()
            
            // Acquire semaphore for parallel execution limit
            select {
            case m.semaphore <- struct{}{}:
                defer func() { <-m.semaphore }()
            case <-ctx.Done():
                return
            }
            
            result := m.executor.RunTest(ctx, t)
            select {
            case results <- result:
            case <-ctx.Done():
            }
        }(test)
    }
    
    // Close results channel when all tests complete
    go func() {
        wg.Wait()
        close(results)
    }()
    
    return results
}
```

## 📋 Phase 4: CLI User Interface (Week 4-5)

### 4.1 Progress Renderer

```go
// internal/ui/renderer.go
type Renderer struct {
    width       int
    height      int
    colorScheme *ColorScheme
    lastUpdate  time.Time
    testStates  map[string]*TestState
}

type TestState struct {
    Name      string
    Status    TestStatus
    Duration  time.Duration
    Progress  float64
    Output    []string
}

func (r *Renderer) RenderTestProgress(tests []*TestFunction, results <-chan *TestResult) {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case result := <-results:
            r.updateTestState(result)
        case <-ticker.C:
            r.refresh()
        }
    }
}

func (r *Renderer) refresh() {
    // Clear screen
    fmt.Print("\033[2J\033[H")
    
    // Render header
    r.renderHeader()
    
    // Render test list with status
    r.renderTestList()
    
    // Render summary
    r.renderSummary()
    
    // Render footer with controls
    r.renderFooter()
}

func (r *Renderer) renderTestList() {
    for _, state := range r.testStates {
        statusIcon := r.getStatusIcon(state.Status)
        statusColor := r.getStatusColor(state.Status)
        
        duration := ""
        if state.Duration > 0 {
            duration = fmt.Sprintf("(%v)", state.Duration.Truncate(time.Millisecond))
        }
        
        fmt.Printf("%s %s%s%s %s\n",
            statusIcon,
            statusColor,
            state.Name,
            colorReset,
            duration,
        )
        
        // Show progress bar for running tests
        if state.Status == StatusRunning {
            r.renderProgressBar(state.Progress)
        }
    }
}
```

### 4.2 Color Schemes and Visual Feedback

```go
// internal/ui/colors.go
type ColorScheme struct {
    Passed  string
    Failed  string
    Running string
    Pending string
    Timeout string
    Skipped string
    Reset   string
}

var DefaultColors = &ColorScheme{
    Passed:  "\033[32m", // Green
    Failed:  "\033[31m", // Red
    Running: "\033[33m", // Yellow
    Pending: "\033[37m", // White
    Timeout: "\033[35m", // Magenta
    Skipped: "\033[36m", // Cyan
    Reset:   "\033[0m",  // Reset
}

func (r *Renderer) getStatusIcon(status TestStatus) string {
    switch status {
    case StatusPassed:
        return "✅"
    case StatusFailed:
        return "❌"
    case StatusRunning:
        return "🏃"
    case StatusPending:
        return "⏳"
    case StatusTimeout:
        return "⏰"
    case StatusSkipped:
        return "⏭️"
    default:
        return "❓"
    }
}

func (r *Renderer) renderProgressBar(progress float64) {
    width := 40
    filled := int(progress * float64(width))
    
    bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
    fmt.Printf("    [%s] %.1f%%\n", bar, progress*100)
}
```

## 📋 Phase 5: Result Persistence (Week 5-6)

### 5.1 Structured Output Writers

```go
// internal/storage/writer.go
type Writer interface {
    WriteResults(results []*TestResult) error
    WriteSession(session *TestSession) error
}

type FileWriter struct {
    outputPath string
    format     OutputFormat
}

type OutputFormat string

const (
    FormatJSON  OutputFormat = "json"
    FormatXML   OutputFormat = "xml"
    FormatJUnit OutputFormat = "junit"
)

func (w *FileWriter) WriteResults(results []*TestResult) error {
    switch w.format {
    case FormatJSON:
        return w.writeJSON(results)
    case FormatXML:
        return w.writeXML(results)
    case FormatJUnit:
        return w.writeJUnit(results)
    default:
        return fmt.Errorf("unsupported format: %s", w.format)
    }
}

func (w *FileWriter) writeJSON(results []*TestResult) error {
    session := &TestSession{
        ID:        generateSessionID(),
        StartTime: time.Now(),
        Results:   results,
        Summary:   calculateSummary(results),
    }
    
    data, err := json.MarshalIndent(session, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(w.outputPath, data, 0644)
}
```

### 5.2 Output Formats

```json
// Example JSON output format
{
  "id": "session-20250731-143022",
  "start_time": "2025-07-31T14:30:22Z",
  "end_time": "2025-07-31T14:32:15Z",
  "duration": "1m53s",
  "summary": {
    "total": 47,
    "passed": 42,
    "failed": 3,
    "skipped": 1,
    "timeout": 1,
    "success_rate": 89.4
  },
  "results": [
    {
      "test": {
        "name": "TestUserAuthentication",
        "package": "auth",
        "file": "/tests/pkg/auth/auth_test.go",
        "line": 15,
        "type": "unit",
        "tags": ["unit", "auth"]
      },
      "status": "passed",
      "duration": "150ms",
      "start_time": "2025-07-31T14:30:22.100Z",
      "end_time": "2025-07-31T14:30:22.250Z",
      "output": "=== RUN   TestUserAuthentication\n--- PASS: TestUserAuthentication (0.15s)\n"
    }
  ]
}
```

## 📋 Phase 6: Integration and Polish (Week 6-7)

### 6.1 Command-Line Interface

```bash
# Basic usage
testicle                              # Run with defaults
testicle --dir ./my-tests             # Specify test directory
testicle --watch=false                # Disable watching
testicle --timeout 1m                 # Set default timeout
testicle --output results.json        # Specify output file

# Advanced usage
testicle --config testicle.yaml       # Use config file
testicle --parallel 8                 # Set parallel workers
testicle --format junit               # JUnit XML output
testicle --tags integration          # Run only integration tests
testicle --verbose                    # Verbose output

# Container usage
docker run -v $(pwd):/tests testicle:latest
docker run -v $(pwd)/tests:/tests testicle:latest --timeout 2m
```

### 6.2 Docker Support

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o testicle cmd/testicle/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates git
WORKDIR /app

# Install Go for test execution
RUN apk add go

COPY --from=builder /app/testicle /usr/local/bin/testicle

# Default test directory mount point
VOLUME ["/tests"]

# Default command
ENTRYPOINT ["testicle"]
CMD ["--dir", "/tests"]
```

### 6.3 Future Web Interface Preparation

```go
// pkg/testicle/client.go - Public API for future web interface
type Client interface {
    // Test management
    DiscoverTests() ([]*TestFunction, error)
    RunTests(selection TestSelection) (<-chan *TestResult, error)
    
    // Real-time events
    Subscribe(eventTypes []EventType) (<-chan Event, error)
    
    // Configuration
    GetConfig() (*Config, error)
    UpdateConfig(config *Config) error
    
    // Results and reporting
    GetSession(id string) (*TestSession, error)
    ListSessions() ([]*TestSession, error)
    ExportResults(format OutputFormat) ([]byte, error)
}

// This interface will be implemented by:
// 1. DirectClient (current CLI implementation)
// 2. HTTPClient (future web interface)
// 3. GRPCClient (future remote execution)
```

## 🚀 Implementation Timeline

### Week 1-2: Foundation
- [ ] CLI framework and configuration
- [ ] Container detection and environment setup
- [ ] Basic project structure

### Week 2-3: Discovery
- [ ] File watcher implementation
- [ ] AST-based test discovery
- [ ] Debounced change detection

### Week 3-4: Execution
- [ ] Test execution engine
- [ ] Timeout management
- [ ] Parallel execution

### Week 4-5: UI
- [ ] CLI progress rendering
- [ ] Color schemes and visual feedback
- [ ] Real-time status updates

### Week 5-6: Storage
- [ ] Result persistence
- [ ] Multiple output formats
- [ ] Session management

### Week 6-7: Polish
- [ ] Docker support
- [ ] Documentation
- [ ] Testing and bug fixes

## 🎯 Success Criteria

### Functional Requirements
- ✅ Detect file changes and re-run affected tests
- ✅ Support both local and container environments
- ✅ Provide clear visual feedback during test execution
- ✅ Handle test timeouts gracefully
- ✅ Generate structured output files
- ✅ Support standard Go test patterns

### Performance Requirements
- ✅ Start up in < 2 seconds
- ✅ Detect file changes within 500ms
- ✅ Process test discovery in < 5 seconds for 1000+ tests
- ✅ Minimal memory footprint (< 50MB for CLI)

### Quality Requirements
- ✅ Cross-platform compatibility (Linux, macOS, Windows)
- ✅ Graceful error handling and recovery
- ✅ Comprehensive test coverage (dogfooding)
- ✅ Clean, extensible architecture for web interface

This implementation plan provides a solid foundation for Testicle v1.0 while preparing for future web interface capabilities.
