# Middleware Package

HTTP middleware functions for Go web applications. This package provides common middleware patterns for CORS, logging, Google Cloud metadata headers, and middleware composition.

## Version

Current version: `v0.3.0`

## Available Middleware

### CORS Middleware

Adds CORS (Cross-Origin Resource Sharing) headers to HTTP responses.

```go
import "github.com/nzions/sharedgolibs/pkg/middleware"

// Basic usage
handler := middleware.WithCORS(myHandler)

// With HTTP server
http.ListenAndServe(":8080", middleware.WithCORS(myMux))
```

**Headers added:**
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-Requested-With, x-client-version, x-firebase-gmpid`
- `Access-Control-Allow-Credentials: true`

**Features:**
- Handles preflight OPTIONS requests automatically
- Returns `204 No Content` for OPTIONS requests

### Logging Middleware

Logs HTTP requests with method, path, remote address, status code, and duration.

```go
import (
    "log"
    "log/slog"
    "github.com/nzions/sharedgolibs/pkg/middleware"
)

// With standard log.Logger
logger := log.New(os.Stdout, "", log.LstdFlags)
handler := middleware.WithLogging(logger)(myHandler)

// With slog.Logger
slogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
handler := middleware.WithLogging(slogger)(myHandler)

// Fallback to default logger
handler := middleware.WithLogging(nil)(myHandler)
```

**Supported logger types:**
- `*log.Logger` - Standard library logger
- `*slog.Logger` - Structured logging (Go 1.21+)
- `nil` or other types - Falls back to `log.Printf`

**Log format:**
- Standard logger: `"GET /api/users 192.168.1.1 200 15.2ms"`
- Structured logger: JSON/text with fields: `method`, `path`, `remote_addr`, `status`, `duration`

### Google Metadata Flavor Middleware

Adds Google Cloud metadata service headers to HTTP responses.

```go
import "github.com/nzions/sharedgolibs/pkg/middleware"

// Add Google metadata headers
handler := middleware.WithGoogleMetadataFlavor(myHandler)
```

**Headers added:**
- `Metadata-Flavor: Google`
- `Server: Metadata Server for Google Compute Engine`

**Use cases:**
- Testing applications that expect Google Cloud metadata headers
- Simulating Google Cloud environment in development
- Applications that validate metadata service responses

### Combined Middleware

Convenience functions for common middleware combinations.

```go
import "github.com/nzions/sharedgolibs/pkg/middleware"

// Combine logging and CORS
logger := slog.Default()
handler := middleware.LogAndCORS(logger)(myHandler)

// Chain multiple middleware
handler := middleware.Chain(
    middleware.WithCORS,
    middleware.WithGoogleMetadataFlavor,
    middleware.WithLogging(logger),
)(myHandler)
```

**Available combinations:**
- `LogAndCORS(logger)` - Applies logging and CORS middleware
- `Chain(middlewares...)` - Applies multiple middleware in order

## Usage Examples

### Basic HTTP Server

```go
package main

import (
    "log/slog"
    "net/http"
    "github.com/nzions/sharedgolibs/pkg/middleware"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", healthHandler)
    
    // Apply middleware
    logger := slog.Default()
    handler := middleware.Chain(
        middleware.WithCORS,
        middleware.WithGoogleMetadataFlavor,
        middleware.WithLogging(logger),
    )(mux)
    
    log.Println("Server starting on :8080")
    http.ListenAndServe(":8080", handler)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### Gin Framework Integration

```go
package main

import (
    "log/slog"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/nzions/sharedgolibs/pkg/middleware"
)

func main() {
    r := gin.New()
    
    // Convert middleware for Gin
    logger := slog.Default()
    r.Use(gin.WrapH(middleware.WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))
    r.Use(gin.WrapH(middleware.WithLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))))
    
    r.GET("/api/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.Run(":8080")
}
```

### Custom Middleware Chain

```go
// Create a reusable middleware stack
func createMiddlewareStack(logger *slog.Logger) func(http.Handler) http.Handler {
    return middleware.Chain(
        middleware.WithCORS,
        middleware.WithGoogleMetadataFlavor,
        middleware.WithLogging(logger),
        // Add custom middleware here
        func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Custom logic
                w.Header().Set("X-Custom-Header", "MyApp")
                next.ServeHTTP(w, r)
            })
        },
    )
}

// Usage
logger := slog.Default()
middlewareStack := createMiddlewareStack(logger)
handler := middlewareStack(myHandler)
```

## Testing

```go
import (
    "net/http/httptest"
    "testing"
    "github.com/nzions/sharedgolibs/pkg/middleware"
)

func TestCORSMiddleware(t *testing.T) {
    handler := middleware.WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    
    if w.Header().Get("Access-Control-Allow-Origin") != "*" {
        t.Error("CORS header not set")
    }
}
```

## Performance Considerations

- **Minimal overhead**: All middleware functions have minimal performance impact
- **Memory efficient**: No memory allocations in hot paths except for logging
- **Request logging**: Logging middleware captures response status without buffering response body
- **Header operations**: Header setting operations are O(1)

## Thread Safety

All middleware functions are thread-safe and can be used concurrently across multiple goroutines.

## Dependencies

This package uses only Go standard library dependencies:
- `net/http` - HTTP handling
- `log` - Standard logging
- `log/slog` - Structured logging (Go 1.21+)
- `time` - Duration measurement for logging

## License

This work is dedicated to the public domain under [CC0 1.0 Universal](../../LICENSE).
