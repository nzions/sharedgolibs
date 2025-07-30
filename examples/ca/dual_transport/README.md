# CA-Integrated Dual Protocol Transport

The CA-integrated dual protocol transport provides a one-shot server that automatically fetches certificates from a CA server and accepts both HTTP and HTTPS connections on the same port.

## Features

- **CA Integration**: Automatically requests certificates from CA server
- **Single Port Operation**: Listen on one port for both HTTP and HTTPS traffic
- **Automatic Protocol Detection**: Detects TLS handshake bytes (0x16) vs plain HTTP
- **Zero Configuration**: Just provide service details, certificates are handled automatically
- **Production Ready**: Includes proper error handling, timeouts, and graceful shutdown
- **Protocol Awareness**: Handlers can detect which protocol was used via `r.TLS != nil`

## Quick Start

### Environment Setup

```bash
# Required: CA server URL
export SGL_CA=http://localhost:8090

# Optional: API key for CA authentication
export SGL_CA_API_KEY=your-api-key
```

### Basic Usage

```go
package main

import (
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Create CA-integrated dual protocol server (one-shot)
    server, err := ca.CreateSecureDualProtocolServer(
        "my-service",           // service name
        "127.0.0.1",           // service IP
        "8443",                // port
        []string{"localhost"}, // additional domains
        nil,                   // handler (nil = default)
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server - automatically handles both HTTP and HTTPS
    server.ListenAndServe()
}
```

### How It Works

1. **Certificate Request**: Automatically calls `RequestCertificate()` with provided details
2. **TLS Setup**: Configures TLS with certificates from CA
3. **Protocol Detection**: First byte examination determines HTTP vs HTTPS
4. **Unified Handling**: Same handlers serve both protocols

## Example

Run the example:

```bash
# Start a CA server first (in another terminal)
cd path/to/ca
go run ./cmd/ca-server

# Then run the dual protocol example
export SGL_CA=http://localhost:8090
go run ./examples/ca/dual_transport/main.go
```

Test both protocols:

```bash
# HTTP request
curl http://localhost:8443/

# HTTPS request (with CA-issued certificate)
curl https://localhost:8443/

# Get detailed info
curl http://localhost:8443/info
curl https://localhost:8443/info
```

## Comparison with Standard Approach

### Standard HTTP/HTTPS Servers
```go
// Separate servers for HTTP and HTTPS
httpServer := &http.Server{Addr: ":80", Handler: handler}
httpsServer := &http.Server{Addr: ":443", Handler: handler, TLSConfig: tlsConfig}

go httpServer.ListenAndServe()
go httpsServer.ListenAndServeTLS("cert.pem", "key.pem")
```

### CA-Integrated Dual Protocol
```go
// Single server handles both protocols with CA certificates
server, err := ca.CreateSecureDualProtocolServer("service", "127.0.0.1", "8443", domains, handler)
server.ListenAndServe() // Handles both HTTP and HTTPS automatically
```

## Advanced Configuration

For more control, use the lower-level functions:

```go
// Manual certificate management
certResp, err := ca.RequestCertificate("service", "127.0.0.1", domains)
cert, err := tls.X509KeyPair([]byte(certResp.Certificate), []byte(certResp.PrivateKey))

tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
server := &http.Server{Addr: ":8443", Handler: handler}
dualServer := ca.NewDualProtocolServer(server, tlsConfig)
```

## Protocol Detection in Handlers

Handlers can detect which protocol was used:

```go
func myHandler(w http.ResponseWriter, r *http.Request) {
    if r.TLS != nil {
        // HTTPS connection
        fmt.Fprintf(w, "Secure connection with TLS %s\n", 
            getTLSVersion(r.TLS.Version))
    } else {
        // HTTP connection  
        fmt.Fprintf(w, "Plain HTTP connection\n")
    }
}
```

## Security Considerations

### Certificate Management
- Certificates are automatically requested from the CA server
- No manual certificate file management required
- Automatic renewal would need to be implemented separately

### Protocol Handling
- HTTP requests are handled normally
- HTTPS requests use CA-issued certificates
- Consider redirecting HTTP to HTTPS for sensitive endpoints:

```go
func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
    if r.TLS == nil {
        httpsURL := "https://" + r.Host + r.RequestURI
        http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
        return
    }
    // Handle HTTPS request normally
}
```

## Error Handling

The `CreateSecureDualProtocolServer` function handles various error conditions:

- **CA Server Unavailable**: Returns error if SGL_CA server is not reachable
- **Authentication Failure**: Returns error if API key is invalid
- **Certificate Issues**: Returns error if certificate parsing fails
- **Invalid Parameters**: Returns error for missing service name or domains

## Monitoring and Logging

The server includes structured logging:

```bash
# Example log output
time=2025-07-30T14:00:00Z level=INFO msg="Request handled" protocol=HTTPS method=GET path=/info
time=2025-07-30T14:00:01Z level=INFO msg="Request handled" protocol=HTTP method=GET path=/health
```

## Testing

### Unit Tests

```bash
go test -v ./pkg/ca/ -run TestCreateSecureDualProtocol
```

### Integration Testing

```bash
# Start CA server
export SGL_CA=http://localhost:8090
go run ./examples/ca/server &

# Start dual protocol server
go run ./examples/ca/dual_transport/main.go &

# Test both protocols
curl http://localhost:8443/health
curl -k https://localhost:8443/health
```

## Production Deployment

### Environment Variables

```bash
# Required
SGL_CA=https://ca.yourdomain.com

# Optional
SGL_CA_API_KEY=production-api-key
```

### Recommended Settings

```go
server, err := ca.CreateSecureDualProtocolServer(
    "production-service",
    "10.0.1.100",  // Internal IP
    "8443",
    []string{
        "api.yourdomain.com",
        "service.internal",
    },
    yourProductionHandler,
)
```

### Health Checks

The server provides health check endpoints accessible via both protocols:

```bash
# Health check via HTTP
curl http://your-server:8443/health

# Health check via HTTPS  
curl https://your-server:8443/health
```

This makes it easy to configure load balancers and monitoring systems that can use either protocol.

## Integration with CA Package

The dual protocol server seamlessly integrates with other CA package functions:

```go
// Update transport to trust the CA
ca.UpdateTransport()

// Create dual protocol server with same CA
server, err := ca.CreateSecureDualProtocolServer(...)

// Create gRPC credentials for the same CA
creds, err := ca.CreateGRPCCredentials()
```

All components work together using the same CA infrastructure.
