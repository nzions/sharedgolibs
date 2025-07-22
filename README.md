# Shared Go Libraries

Common Go packages shared between allmytails and googleemu projects, providing Certificate Authority functionality, HTTP middleware, and utility functions.

## License

This work is dedicated to the public domain under [CC0 1.0 Universal](LICENSE).

## Packages

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

### 🌐 pkg/middleware (v0.3.0)
HTTP middleware for CORS, logging, Google metadata flavor headers, API key authentication, and request handling.

**Key Features:**
- CORS headers with customizable origins
- Structured logging with multiple logger support
- Google Cloud metadata headers
- API key authentication middleware
- Middleware chaining and composition utilities

### 🛠️ pkg/util (v0.1.0)  
Environment variable utilities and common helper functions.

**Key Features:**
- Environment variable handling with fallbacks
- Type-safe environment variable parsing
- Configuration management utilities

## Usage

```go
import (
    "github.com/nzions/sharedgolibs/pkg/ca"
    "github.com/nzions/sharedgolibs/pkg/middleware"
    "github.com/nzions/sharedgolibs/pkg/util"
)
```

## Quick Start

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

- [Certificate Authority (pkg/ca)](pkg/ca/README.md) - Complete CA documentation
- [HTTP Middleware (pkg/middleware)](pkg/middleware/README.md) - Middleware documentation  
- [Utilities (pkg/util)](pkg/util/README.md) - Utility functions documentation

## Examples

See the [examples directory](pkg/ca/examples/) for complete working examples of:
- Certificate persistence and storage
- Thread-safe concurrent operations
- gRPC server and client setup
- HTTP transport integration

## Testing

Run tests for all packages:

```bash
go test ./...
```

## Semantic Versioning

All packages follow [Semantic Versioning](https://semver.org/) for backwards compatibility and clear version progression.
