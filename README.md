# Shared Go Libraries

Production-ready Go packages for allmytails and googleemu projects, featuring comprehensive service management, Certificate Authority functionality, HTTP middleware, and utility functions.

## License

This work is licensed under [CC0 1.0 Universal](LICENSE).

## 🚀 Featured Package

### 🎯 Service Manager (v0.3.0) - **NEW UNIFIED LIBRARY**
Revolutionary unified service management combining port discovery, Docker integration, and process management into a single, powerful OO-style interface.

**🔥 Key Features:**
- **Docker Integration**: Multi-environment support (Docker Desktop + Colima)
- **Intelligent Discovery**: Automatic service categorization (expected vs unexpected)
- **SSH Detection**: Smart Docker port forwarding identification
- **Process Management**: Kill services by port or container
- **Object-Oriented Design**: Clean, modular API with functional options
- **External Library Ready**: Perfect for integration with functional options pattern
- **Autoport Generation**: Creates configuration from docker-compose.yml
- **Multi-Format Output**: JSON and human-readable formats
- **Comprehensive CLI**: Full-featured command-line interface

**🎯 Perfect For:**
- Development environment management
- Container orchestration monitoring
- Service health checking
- Port conflict resolution
- Multi-project service coordination

## Core Packages

### 🔐 Certificate Authority (ca) - v1.4.0
Complete Certificate Authority implementation with dynamic certificate issuance, persistent storage, thread-safe operations, gRPC support, and HTTP transport integration.

**Key Features:**
- Dynamic certificate generation for any service or domain
- Persistent storage with RAM and disk backends
- Thread-safe concurrent operations  
- Web UI for certificate management
- REST API for programmatic access
- gRPC server and client support
- HTTP transport monkey-patching for zero-code-change integration
- Optional transport configuration with `UpdateTransportOnlyIf()`
- Environment-driven configuration

### 🌐 HTTP Middleware (v0.3.0)
Production-grade HTTP middleware for CORS, logging, Google metadata flavor headers, API key authentication, and request handling.

**Key Features:**
- CORS headers with customizable origins
- Structured logging with multiple logger support
- Google Cloud metadata headers
- API key authentication middleware
- Middleware chaining and composition utilities

### 🛠️ Utilities (v0.1.0)  
Environment variable utilities and common helper functions.

**Key Features:**
- Environment variable handling with fallbacks
- Type-safe environment variable parsing
- Configuration management utilities

### 🧹 Binary Cleaner (v0.1.0)
Intelligent binary file detection and removal tool for cleaning up build artifacts and compiled binaries.

**Key Features:**
- **Smart Detection**: Identifies Mach-O and ELF binaries by analyzing file headers
- **Safety First**: Only examines executable files, validates headers before removal
- **Flexible Search**: Recursive directory scanning with configurable depth
- **Dry Run Mode**: Preview operations without actually removing files
- **Format Support**: Mach-O (macOS) and ELF (Linux/Unix) binaries
- **CLI Tool**: Full-featured command-line interface with verbose output
- **VS Code Integration**: Built-in tasks for development workflow

###  Auto Port (v0.1.0)
Auto-generated port configurations from Docker Compose for consistent service discovery.

**Key Features:**
- Auto-generated from docker-compose.yml
- Service name and port mapping discovery
- Health endpoint URL generation
- Security configuration detection (HTTP/HTTPS)
- Network alias mapping
- Service dependency tracking
- Environment variable extraction
- Pure Go with no external dependencies

## 🎯 Service Manager - Quick Start

### Basic Usage

```go
import "github.com/nzions/sharedgolibs/pkg/servicemanager"

// Create service manager with Docker integration
sm := servicemanager.New()

// Discover all services
services, err := sm.DiscoverAllServices()
if err != nil {
    log.Fatal(err)
}

for _, service := range services {
    fmt.Printf("%s on port %d (%s) - %s\n", 
        service.Name, service.ExternalPort, service.Type, service.Status)
}
```

### Binary Cleaner - Quick Start

```go
import "github.com/nzions/sharedgolibs/pkg/binarycleaner"

// Create cleaner with safe defaults
config := binarycleaner.Config{
    Directory: "/path/to/clean",
    DryRun:    true,  // Preview mode
    Verbose:   true,  // Detailed output
    Recursive: true,  // Search subdirectories
}

cleaner := binarycleaner.New(config)

// Find and preview what would be removed
err := cleaner.Clean()
if err != nil {
    log.Fatal(err)
}
```

**CLI Usage:**
```bash
# Build the tool
go build -o bin/binarycleaner ./cmd/binarycleaner/

# Preview cleanup (safe)
./bin/binarycleaner --dry-run --verbose --recursive

# Actually remove binaries
./bin/binarycleaner --recursive --dir ./build
```

### Advanced Configuration

```go
// Custom configuration with options
sm := servicemanager.New(
    servicemanager.WithPortRange(3000, 9000),
    servicemanager.WithKnownService(3000, "My API", "http://localhost:3000/health", false),
    servicemanager.WithMonitoredPort(3001, "Frontend"),
    servicemanager.WithDockerTimeout(10*time.Second),
)

// Get comprehensive service status
status, err := sm.GetServiceStatus()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Services: %d running, %d expected, %d unexpected\n", 
    status.Listening, status.Expected, status.Unexpected)
```

### Service Management

```go
// Kill specific service
err := sm.KillServiceOnPort(8080)
if err != nil {
    log.Printf("Failed to kill service: %v", err)
}

// Kill all monitored services
errors := sm.KillAllServices()
for _, err := range errors {
    log.Printf("Error: %v", err)
}
```

## Command Line Tools

### `servicemanager` - **NEW UNIFIED CLI**
The ultimate service management tool combining all previous functionality:

```bash
go build -o bin/servicemanager ./cmd/servicemanager/

# Service Discovery
./bin/servicemanager                    # Check all services
./bin/servicemanager -expected          # Show only expected services
./bin/servicemanager -docker            # Show only Docker containers
./bin/servicemanager -status            # Comprehensive status
./bin/servicemanager -missing           # Show missing services

# Service Control
./bin/servicemanager -k                 # Kill all monitored services
./bin/servicemanager -kill-port=8080    # Kill service on specific port

# Configuration
./bin/servicemanager -range=3000-4000   # Custom port range
./bin/servicemanager -generate=docker-compose.yml  # Generate autoport config

# Output Formats
./bin/servicemanager -json              # JSON output
./bin/servicemanager -port=8080 -json   # Check specific port as JSON
```

## Usage

```go
import (
    "github.com/nzions/sharedgolibs/pkg/servicemanager"
    "github.com/nzions/sharedgolibs/pkg/ca"
    "github.com/nzions/sharedgolibs/pkg/middleware"
    "github.com/nzions/sharedgolibs/pkg/util"
)
```

## Real-World Examples

### Service Manager

```go
// Monitor development environment
func monitorDevelopmentEnvironment() {
    sm := servicemanager.New()
    
    // Get comprehensive service status
    status, err := sm.GetServiceStatus()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Services: %d running, %d expected, %d unexpected\n", 
        status.Listening, status.Expected, status.Unexpected)
    fmt.Printf("Docker: %d image matches, %d mismatches\n", 
        status.ImageMatch, status.ImageMismatch)
    
    // List missing services
    missing := sm.GetMissingServices()
    if len(missing) > 0 {
        fmt.Println("Missing services:")
        for _, service := range missing {
            fmt.Printf("  - %s (port %d)\n", service.Name, service.ExternalPort)
        }
    }
}

// External library integration
func integrateWithExternalLib() {
    // Simple mode for external libraries
    sm := servicemanager.NewSimple(
        servicemanager.WithPortRange(8000, 9000),
        servicemanager.WithMonitoredPort(8080, "My Service"),
    )
    
    services, _ := sm.DiscoverAllServices()
    for _, service := range services {
        if service.IsListening {
            fmt.Printf("Found: %s on %d\n", service.Name, service.ExternalPort)
        }
    }
}
```

### Certificate Authority

```go
// Create CA and issue certificate
ca, _ := ca.NewCA(nil)
cert, key, _ := ca.GenerateCertificate("my-service", "127.0.0.1", []string{"service.local"})

// Start CA server with web UI
server, _ := ca.NewServer(nil)
server.Start() // http://localhost:8090
```

### HTTP Middleware

```go
// Apply middleware to HTTP handler  
logger := slog.Default()
handler := middleware.Chain(
    middleware.WithCORS,
    middleware.WithLogging(logger),
    middleware.WithAPIKey("your-api-key"),
)(myHandler)
```

### Environment Variables

```go
// Get environment variables with fallbacks
dbURL := util.MustGetEnv("DATABASE_URL", "localhost:5432")
port := util.MustGetEnv("PORT", "8080")
```

## Installation

```bash
go get github.com/nzions/sharedgolibs
```

## Documentation

- [Service Manager (pkg/servicemanager)](pkg/servicemanager/README.md) - **NEW** Complete unified service management
- [Certificate Authority (pkg/ca)](pkg/ca/README.md) - Complete CA documentation
- [HTTP Middleware (pkg/middleware)](pkg/middleware/README.md) - Middleware documentation  
- [Utilities (pkg/util)](pkg/util/README.md) - Utility functions documentation
- [Auto Port (pkg/autoport)](pkg/autoport/) - Auto-generated port configurations

## Examples

See the [examples directory](examples/) for complete working examples of:
- [Service Manager examples](examples/servicemanager/) - Comprehensive service management demos
- [Certificate Authority examples](pkg/ca/examples/) - CA usage and integration
- Certificate persistence and storage
- Thread-safe concurrent operations
- gRPC server and client setup
- HTTP transport integration

## Migration Guide

### From portmanager/processmanager to servicemanager

```go
// Old portmanager usage
pm := portmanager.New()
services, _ := pm.DiscoverAllServices()

// New servicemanager usage (drop-in replacement)
sm := servicemanager.New()
services, _ := sm.DiscoverAllServices()

// Old processmanager usage  
pm := processmanager.New()
status := pm.CheckAllPorts()

// New servicemanager usage
sm := servicemanager.New()
status, _ := sm.CheckMonitoredPorts()
```

The new `servicemanager` provides full backward compatibility while adding powerful new features!

## Testing

Run tests for all packages:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Test the CLI tools:

```bash
go build -o bin/servicemanager ./cmd/servicemanager/
./bin/servicemanager --version
./bin/servicemanager --help
```

## VS Code Integration

The project includes VS Code tasks for streamlined development. Access via `Cmd+Shift+P` → "Tasks: Run Task":

### Available Tasks

- **Build All CLI Tools**: Builds both servicemanager and binarycleaner tools
- **Build Binary Cleaner**: Builds only the binary cleaner tool
- **Build Service Manager**: Builds only the service manager tool  
- **Test Binary Cleaner**: Runs unit tests for the binary cleaner package
- **Run Binary Cleaner (Dry Run)**: Safely previews binary cleanup (auto-builds first)

### Quick Development Workflow

1. Make code changes
2. Press `Cmd+Shift+P` 
3. Type "Tasks: Run Task"
4. Select "Test Binary Cleaner" to verify changes
5. Select "Run Binary Cleaner (Dry Run)" to test functionality

The tasks provide integrated development without leaving VS Code!

### Setting Up VS Code Tasks

The VS Code tasks are defined in `.vscode/tasks.json`. To add or modify tasks:

1. **Create the directory** (if it doesn't exist):
   ```bash
   mkdir -p .vscode
   ```

2. **Edit tasks.json**:
   ```json
   {
       "version": "2.0.0",
       "tasks": [
           {
               "label": "Your Custom Task",
               "type": "shell",
               "command": "your-command-here",
               "group": "build",
               "isBackground": false
           }
       ]
   }
   ```

3. **Task Properties**:
   - `label`: Name shown in VS Code task picker
   - `command`: Shell command to execute
   - `group`: Task category (`build`, `test`, etc.)
   - `isBackground`: Whether task runs continuously
   - `dependsOn`: Run another task first

4. **Access tasks**: `Cmd+Shift+P` → "Tasks: Configure Task" to edit

### Docker Compose Integration

The servicemanager now includes built-in support for generating autoport configurations from `docker-compose.yml` files:

```bash
# Using the provided Makefile
make regen-autoport

# Or manually
go build -o bin/servicemanager ./cmd/servicemanager/
./bin/servicemanager -generate=docker-compose.yml
```

This will regenerate the `pkg/autoport/autoport.go` file with the latest service configurations from your Docker Compose setup.

## Semantic Versioning

All packages follow [Semantic Versioning](https://semver.org/) for backwards compatibility and clear version progression.

## 🌟 Why Choose sharedgolibs?

- **Production Ready**: Battle-tested in allmytails and googleemu projects
- **Modern Design**: Object-oriented patterns with functional options
- **Docker Native**: Built for containerized development environments  
- **Zero Dependencies**: Minimal external dependencies for core functionality
- **Comprehensive Testing**: Full test coverage with real-world scenarios
- **Great Documentation**: Complete examples and migration guides
- **Active Development**: Continuously improved based on real project needs
