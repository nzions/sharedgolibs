# Dual Protocol Server

The dual protocol server allows you to handle both HTTP and HTTPS connections on the same port by detecting the protocol from the first bytes of the connection.

## Overview

The dual protocol functionality is split into two main components:

1. **Main CA Integration** (`pkg/ca/dual_transport.go`): Provides the high-level `CreateSecureDualProtocolServer` function that integrates with the CA package
2. **Sub-module Implementation** (`pkg/ca/dualprotocol/`): Contains the low-level dual protocol detection and handling logic

## Quick Start

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/nzions/sharedgolibs/pkg/ca"
    "github.com/nzions/sharedgolibs/pkg/logi"
)

func main() {
    // Create logger
    logger := logi.NewDemonLogger("my-dual-service")

    // Create dual protocol server with automatic certificate management
    server, err := ca.CreateSecureDualProtocolServer(
        "my-service",              // service name
        "127.0.0.1",              // service IP
        "8443",                   // port
        []string{"localhost"},    // additional domains
        nil,                      // handler (nil = default)
        logger,                   // logger (nil = creates default)
    )
    if err != nil {
        panic(err)
    }

    // Start the server
    go server.ListenAndServe()

    // Graceful shutdown example
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    server.Shutdown(ctx)
}
```

## Environment Variables

The dual protocol server requires the following environment variables for CA integration:

- `SGL_CA` (required): URL of the CA server (e.g., `http://localhost:8080`)
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

## Features

### Protocol Detection

The server automatically detects HTTP vs HTTPS by examining the first byte of incoming connections:
- `0x16` indicates a TLS handshake (HTTPS)
- Any other value indicates HTTP

### Connection Information

The server injects connection details into the HTTP request context:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    // Get connection info from context
    connInfo, ok := dualprotocol.GetConnectionInfo(r)
    if ok {
        fmt.Printf("Protocol: %s\n", connInfo.Protocol)
        fmt.Printf("Is TLS: %t\n", connInfo.IsTLS)
        fmt.Printf("TLS Version: %s\n", connInfo.TLSVersion)
        fmt.Printf("Cipher Suite: %s\n", connInfo.CipherSuite)
        fmt.Printf("Remote Addr: %s\n", connInfo.RemoteAddr)
        fmt.Printf("Detected At: %s\n", connInfo.DetectedAt)
    }
}

// Wrap your handler to enable connection info injection
wrappedHandler := dualprotocol.WrapHandlerWithConnectionInfo(myHandler)
```

### Logging Integration

The server integrates with the `logi` package for structured logging:

```go
logger := logi.NewDemonLogger("my-service")
server, err := ca.CreateSecureDualProtocolServer(
    "my-service", "127.0.0.1", "8443", 
    []string{"localhost"}, handler, logger,
)
```

## Usage Examples

### Basic Usage

```go
// Create server with default handler
server, err := ca.CreateSecureDualProtocolServer(
    "basic-service", "127.0.0.1", "8443", 
    []string{"localhost"}, nil, nil,
)
```

### Custom Handler

```go
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    connInfo, _ := dualprotocol.GetConnectionInfo(r)
    fmt.Fprintf(w, "Hello from %s!\n", connInfo.Protocol)
})

server, err := ca.CreateSecureDualProtocolServer(
    "custom-service", "127.0.0.1", "8443", 
    []string{"localhost"}, handler, nil,
)
```

### Advanced Configuration

For more control, you can use the sub-module directly:

```go
import "github.com/nzions/sharedgolibs/pkg/ca/dualprotocol"

// Create your own HTTP server
httpServer := &http.Server{
    Addr:    ":8443",
    Handler: myHandler,
}

// Create TLS config
tlsConfig := &tls.Config{
    // Your TLS configuration
}

// Create logger
logger := logi.NewDemonLogger("advanced-service")

// Create dual protocol server
server := dualprotocol.NewServer(httpServer, tlsConfig, logger)

// Start listening
server.ListenAndServe()
```

## Testing

The server can be tested with both HTTP and HTTPS clients:

```bash
# Test HTTP
curl http://localhost:8443/

# Test HTTPS (with self-signed cert warning)
curl -k https://localhost:8443/
```

## Architecture

The dual protocol server works by:

1. **Listening** on a single port for all connections
2. **Peeking** at the first byte of each connection
3. **Detecting** protocol based on the first byte (0x16 = TLS)
4. **Upgrading** to TLS if needed using Go's built-in TLS functionality
5. **Injecting** connection details into the HTTP request context
6. **Logging** protocol detection and connection details

## Error Handling

The server handles various error conditions:

- **Protocol detection timeout**: Configurable timeout for protocol detection
- **TLS handshake failures**: Proper error reporting and connection cleanup
- **Certificate errors**: Integration with CA package error handling
- **Graceful shutdown**: Context-based shutdown with configurable timeout

## Performance Considerations

- **Minimal overhead**: Protocol detection adds only a single byte peek operation
- **Connection pooling**: Uses Go's built-in HTTP server connection pooling
- **TLS optimization**: Leverages Go's optimized TLS implementation
- **Memory efficient**: Minimal memory overhead per connection

## Security

- **TLS best practices**: Uses secure TLS configuration defaults
- **Certificate validation**: Integrates with CA package for proper certificate management
- **Connection limits**: Inherits HTTP server connection limiting
- **Timeout handling**: Proper timeouts to prevent resource exhaustion
