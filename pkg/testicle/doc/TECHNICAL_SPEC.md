# Testicle Technical Specification

## 🏗️ System Architecture

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Client    │    │   Web UI        │    │   VS Code Ext   │
│                 │    │                 │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│                 │    │                 │    │                 │
│  Command Parser │    │  React Frontend │    │  Extension Host │
│  Progress UI    │    │  WebSocket      │    │  Tree Provider  │
│  Report Export  │    │  Real-time UI   │    │  Status Bar     │
│                 │    │                 │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼───────────┐
                    │                         │
                    │    Testicle Core API    │
                    │                         │
                    └─────────────┬───────────┘
                                  │
    ┌─────────────────────────────┼─────────────────────────────┐
    │                             │                             │
    ▼                             ▼                             ▼
┌───────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Discovery   │    │     Runner      │    │    Watcher      │
│               │    │                 │    │                 │
│ AST Parser    │    │ Test Executor   │    │ File Monitor    │
│ Test Finder   │    │ Output Stream   │    │ Change Detect   │
│ Tag Extractor │    │ Status Tracker  │    │ Dependency Map  │
│               │    │                 │    │                 │
└───────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Data Structures

```go
// TestNode represents a test in the hierarchy
type TestNode struct {
    ID          string              `json:"id"`
    Name        string              `json:"name"`
    Type        TestNodeType        `json:"type"` // package, file, test, subtest
    Status      TestStatus          `json:"status"`
    Children    []*TestNode         `json:"children,omitempty"`
    Parent      *TestNode           `json:"-"`
    Metadata    TestMetadata        `json:"metadata"`
    Results     *TestResult         `json:"results,omitempty"`
}

// TestMetadata contains additional test information
type TestMetadata struct {
    Package     string              `json:"package"`
    File        string              `json:"file"`
    Line        int                 `json:"line"`
    Tags        []string            `json:"tags"`
    Timeout     time.Duration       `json:"timeout,omitempty"`
    Parallel    bool                `json:"parallel"`
    Skip        bool                `json:"skip"`
    Dependencies []string           `json:"dependencies"`
}

// TestExecution tracks real-time test execution
type TestExecution struct {
    ID          string              `json:"id"`
    TestID      string              `json:"test_id"`
    Status      TestStatus          `json:"status"`
    StartTime   time.Time           `json:"start_time"`
    EndTime     *time.Time          `json:"end_time,omitempty"`
    Duration    time.Duration       `json:"duration"`
    Output      []OutputLine        `json:"output"`
    Progress    float64             `json:"progress"` // 0.0 to 1.0
    ResourceUsage ResourceUsage     `json:"resource_usage"`
}

// TestSession represents a complete test run session
type TestSession struct {
    ID          string              `json:"id"`
    StartTime   time.Time           `json:"start_time"`
    EndTime     *time.Time          `json:"end_time,omitempty"`
    Config      SessionConfig       `json:"config"`
    TestTree    *TestNode           `json:"test_tree"`
    Executions  []*TestExecution    `json:"executions"`
    Summary     SessionSummary      `json:"summary"`
    Status      SessionStatus       `json:"status"`
}
```

## 🚀 Core Components Deep Dive

### 1. Test Discovery Engine

```go
// TestDiscovery finds and analyzes test functions
type TestDiscovery interface {
    // DiscoverProject scans entire project for tests
    DiscoverProject(projectRoot string) (*TestTree, error)
    
    // DiscoverPackage scans specific package
    DiscoverPackage(packagePath string) (*TestPackage, error)
    
    // WatchForChanges monitors for test file changes
    WatchForChanges(ctx context.Context) (<-chan ChangeEvent, error)
    
    // AnalyzeDependencies builds test dependency graph
    AnalyzeDependencies(testTree *TestTree) (*DependencyGraph, error)
}

// Implementation strategies:
// - Use go/ast for parsing Go source files
// - Extract test functions, benchmarks, examples
// - Parse comments for tags and metadata
// - Build package hierarchy and relationships
// - Track file modification times for incremental updates
```

### 2. Test Runner Engine

```go
// TestRunner executes tests and manages execution lifecycle
type TestRunner interface {
    // RunTests executes selected tests
    RunTests(ctx context.Context, selection TestSelection) (*TestSession, error)
    
    // RunSingle executes a single test
    RunSingle(ctx context.Context, testID string) (*TestExecution, error)
    
    // StreamExecution provides real-time execution updates
    StreamExecution(ctx context.Context) (<-chan ExecutionEvent, error)
    
    // StopExecution cancels running tests
    StopExecution(sessionID string) error
}

// Execution strategies:
// - Wrap 'go test' with custom output parsing
// - Use test2json for structured output
// - Implement custom test runner for advanced features
// - Support parallel execution with worker pools
// - Capture stdout/stderr with real-time streaming
```

### 3. File Watcher System

```go
// FileWatcher monitors filesystem changes
type FileWatcher interface {
    // Watch starts monitoring specified paths
    Watch(ctx context.Context, paths []string) error
    
    // Events returns channel of file change events
    Events() <-chan FileChangeEvent
    
    // AddPath adds new path to watch list
    AddPath(path string) error
    
    // RemovePath removes path from watch list
    RemovePath(path string) error
}

// Change detection strategies:
// - Use fsnotify for efficient file system monitoring
// - Debounce rapid file changes
// - Analyze change impact using dependency graph
// - Smart test selection based on affected code
// - Support for .gitignore-style exclusion patterns
```

### 4. UI Server Architecture

```go
// UIServer provides web interface for test management
type UIServer interface {
    // Start launches the web server
    Start(ctx context.Context, config ServerConfig) error
    
    // BroadcastEvent sends events to all connected clients
    BroadcastEvent(event UIEvent) error
    
    // RegisterHandler adds custom HTTP handlers
    RegisterHandler(pattern string, handler http.Handler)
    
    // GetSessionManager returns session management interface
    GetSessionManager() SessionManager
}

// Server features:
// - WebSocket for real-time communication
// - REST API for test management operations
// - Static file serving for web UI assets
// - Authentication and session management
// - CORS support for development
```

## 📡 Communication Protocols

### WebSocket Events

```typescript
// Client to Server Events
interface ClientEvents {
    // Test execution control
    'test:run': { selection: TestSelection; config?: RunConfig };
    'test:stop': { sessionId: string };
    'test:debug': { testId: string };
    
    // UI interactions
    'ui:subscribe': { eventTypes: string[] };
    'ui:unsubscribe': { eventTypes: string[] };
    'tree:expand': { nodeId: string };
    'tree:collapse': { nodeId: string };
    
    // Configuration
    'config:update': { config: Partial<TestConfig> };
    'filter:apply': { filter: TestFilter };
}

// Server to Client Events
interface ServerEvents {
    // Test execution updates
    'execution:started': TestExecution;
    'execution:progress': { executionId: string; progress: number };
    'execution:output': { executionId: string; output: OutputLine };
    'execution:completed': TestExecution;
    
    // Test discovery updates
    'discovery:started': { projectRoot: string };
    'discovery:progress': { processed: number; total: number };
    'discovery:completed': TestTree;
    
    // File change notifications
    'file:changed': { path: string; changeType: 'add' | 'modify' | 'delete' };
    'tests:affected': { testIds: string[] };
    
    // System notifications
    'error': { message: string; details?: any };
    'warning': { message: string };
    'info': { message: string };
}
```

### REST API Endpoints

```yaml
# Test Management
GET    /api/tests                    # List all tests
GET    /api/tests/{id}               # Get specific test
POST   /api/tests/run                # Run selected tests
DELETE /api/tests/{sessionId}        # Stop test session

# Test Sessions
GET    /api/sessions                 # List test sessions
GET    /api/sessions/{id}            # Get session details
GET    /api/sessions/{id}/results    # Get session results
POST   /api/sessions/{id}/export     # Export session data

# Project Management
GET    /api/project/info             # Get project information
POST   /api/project/discover         # Trigger test discovery
GET    /api/project/config           # Get project configuration
PUT    /api/project/config           # Update project configuration

# Reports and Analytics
GET    /api/reports/summary          # Get test summary
GET    /api/reports/trends           # Get test trends
GET    /api/reports/coverage         # Get coverage data
GET    /api/reports/export/{format}  # Export reports (html, pdf, json)
```

## 🎨 Frontend Architecture

### React Component Hierarchy

```
App
├── Layout
│   ├── Header
│   │   ├── ProjectSelector
│   │   ├── RunControls
│   │   └── SettingsMenu
│   ├── Sidebar
│   │   ├── TestTree
│   │   ├── FilterPanel
│   │   └── TagsPanel
│   └── MainContent
│       ├── TestExecutionView
│       │   ├── ExecutionTimeline
│       │   ├── OutputPanel
│       │   └── ProgressIndicators
│       ├── TestDetailsView
│       │   ├── TestMetadata
│       │   ├── TestHistory
│       │   └── TestCoverage
│       └── ReportsView
│           ├── SummaryDashboard
│           ├── TrendsChart
│           └── CoverageVisualization
```

### State Management

```typescript
// Redux store structure
interface AppState {
    // Test data
    tests: {
        tree: TestNode | null;
        selected: string[];
        filter: TestFilter;
        loading: boolean;
    };
    
    // Execution state
    execution: {
        currentSession: TestSession | null;
        executions: { [id: string]: TestExecution };
        output: { [executionId: string]: OutputLine[] };
    };
    
    // UI state
    ui: {
        sidebarCollapsed: boolean;
        activeView: 'execution' | 'details' | 'reports';
        theme: 'light' | 'dark';
        expandedNodes: Set<string>;
    };
    
    // Configuration
    config: {
        project: ProjectConfig;
        user: UserPreferences;
        runtime: RuntimeConfig;
    };
}
```

## 🔧 CLI Interface Design

### Command Structure

```bash
# Core commands
testicle run [options] [packages...]     # Run tests
testicle watch [options] [packages...]   # Watch mode
testicle ui [options]                    # Launch UI
testicle discover [options]              # Discover tests
testicle report [options]                # Generate reports

# Global options
--config <file>                          # Configuration file
--project-root <path>                    # Project root directory
--verbose, -v                           # Verbose output
--quiet, -q                             # Minimal output
--help, -h                              # Show help
--version                               # Show version

# Run command options
--package <pattern>                      # Package pattern
--test <pattern>                        # Test name pattern
--tag <tag>                             # Test tags
--parallel <count>                      # Parallel workers
--timeout <duration>                    # Test timeout
--coverage                              # Enable coverage
--race                                  # Enable race detection
--failfast                              # Stop on first failure

# Watch command options
--delay <duration>                      # Debounce delay
--ignore <pattern>                      # Ignore patterns
--include <pattern>                     # Include patterns

# UI command options
--port <port>                           # Server port
--host <host>                           # Server host
--open                                  # Open browser automatically
--auth                                  # Enable authentication
```

### Configuration File Format

```yaml
# testicle.yaml
project:
  name: "My Go Project"
  root: "."
  go_mod: "./go.mod"
  
execution:
  parallel: 4
  timeout: "10m"
  race: true
  coverage: true
  fail_fast: false
  
watch:
  enabled: true
  delay: "500ms"
  patterns:
    include:
      - "**/*.go"
      - "**/go.mod"
      - "**/go.sum"
    exclude:
      - "**/vendor/**"
      - "**/.git/**"
      - "**/node_modules/**"
      
ui:
  enabled: true
  port: 8080
  host: "localhost"
  auto_open: true
  theme: "auto"
  
reporting:
  formats: ["html", "json"]
  output_dir: "./test-reports"
  history_limit: 50
  
integrations:
  vscode:
    enabled: true
    port: 8081
  github:
    enabled: false
    token: "${GITHUB_TOKEN}"
```

## 📊 Performance Considerations

### Optimization Strategies

1. **Test Discovery Caching**
   - Cache AST parsing results
   - Incremental discovery for changed files
   - Persistent cache with file modification tracking

2. **Execution Optimization**
   - Smart test scheduling based on execution time
   - Parallel execution with optimal worker allocation
   - Resource-aware test distribution

3. **Memory Management**
   - Streaming output processing
   - Bounded output buffers
   - Garbage collection optimization for long-running sessions

4. **Network Efficiency**
   - WebSocket message batching
   - Compression for large payloads
   - Client-side caching of static data

### Scalability Targets

- **Project Size**: Up to 10,000 test functions
- **Concurrent Tests**: Up to 100 parallel executions
- **Output Volume**: Handle 1GB+ of test output
- **File Watching**: Monitor 10,000+ files efficiently
- **UI Responsiveness**: <100ms UI update latency

## 🔒 Security Considerations

### Input Validation
- Sanitize file paths to prevent directory traversal
- Validate test selection patterns
- Limit resource usage to prevent DoS

### Access Control
- Optional authentication for UI access
- Session-based security
- CORS configuration for development

### Data Protection
- Secure handling of test output (may contain sensitive data)
- Optional output redaction/filtering
- Secure temporary file handling

---

*This technical specification will be refined during implementation based on practical requirements and performance testing.*
