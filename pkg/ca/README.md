# CA Package

[![License: CC0-1.0](https://img.shields.io/badge/License-CC0--1.0-blue.svg)](http://creativecommons.org/publicdomain/zero/1.0/)

The CA package provides comprehensive Certificate Authority functionality for development and testing environments, enabling dynamic certificate issuance, persistent storage, thread-safe operations, gRPC support, and HTTP transport integration.

## Version: v1.4.0

## Features

### 🔐 Certificate Authority (authority.go)
- **Complete CA Implementation**: Full Certificate Authority with RSA key generation and X.509 certificate creation
- **Dynamic Certificate Generation**: Create certificates for any service or domain on-demand
- **Persistent Storage**: RAM and disk-based storage with automatic loading
- **Thread-Safe Operations**: Concurrent certificate generation with proper locking
- **Certificate Management**: Track all issued certificates with serial number lookup

### 🌐 Web Interface & API (server.go, gui.go)
- **Web UI**: User-friendly interface for certificate management and generation
- **REST API**: Programmatic certificate issuance with JSON responses
- **API Key Authentication**: Secure access control with optional API keys
- **Health Monitoring**: Built-in health check endpoints

### 🚀 gRPC Support (transport.go)
- **Secure gRPC Servers**: `CreateSecureGRPCServer()` with automatic certificate provisioning
- **gRPC Credentials**: `CreateGRPCCredentials()` for client connections
- **Dial Options**: `UpdateGRPCDialOptions()` for zero-configuration gRPC clients

### 🔧 HTTP Transport Integration (transport.go)
- **Zero-code-change**: Modify global HTTP transport without changing application code  
- **Automatic CA Trust**: Configure HTTP clients to trust CA-issued certificates
- **Environment-driven**: Configuration through `SGL_CA` and `SGL_CA_API_KEY` variables
- **HTTPS Server Creation**: `CreateSecureHTTPSServer()` with automatic certificates

### 💾 Storage Architecture (storage.go)
- **Storage Abstraction**: Pluggable storage backends via `CertStorage` interface
- **RAM Storage**: High-performance in-memory certificate storage
- **Disk Storage**: Persistent JSON-based certificate storage with atomic operations
- **Automatic Loading**: Certificates and CA state restored on startup

## Quick Start

### Basic CA Usage with Persistence

```go
package main

import (
    "log"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Create CA with disk persistence
    config := ca.DefaultCAConfig()
    config.PersistDir = "./ca-data" // Certificates will persist across restarts
    
    certificateAuthority, err := ca.NewCA(config)
    if err != nil {
        log.Fatal(err)
    }

    // Issue a certificate using the new API
    req := ca.CertRequest{
        ServiceName: "my-service",
        ServiceIP:   "192.168.1.100",
        Domains:     []string{"api.example.com", "service.local"},
    }
    
    resp, err := certificateAuthority.IssueServiceCertificate(req)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Certificate: %s", resp.Certificate)
    log.Printf("Private Key: %s", resp.PrivateKey)
    log.Printf("CA Certificate: %s", resp.CACert)
}
```

### HTTP Server with API Key Authentication

```go
package main

import (
    "log"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Create CA server with API key protection
    config := &ca.ServerConfig{
        Port:      "8090",
        CAConfig:  ca.DefaultCAConfig(),
        EnableGUI: true,
        GUIAPIKey: "my-secure-api-key", // Protect with API key
    }
    
    server, err := ca.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Starting CA server on port 8090...")
    log.Println("Access with: http://localhost:8090?api_key=my-secure-api-key")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Environment-based Transport Setup

```go
package main

import (
    "log"
    "os"
    "net/http"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Set environment variables
    os.Setenv("SGL_CA", "http://localhost:8090")
    os.Setenv("SGL_CA_API_KEY", "my-secure-api-key")
    
    // Update HTTP transport to trust CA certificates
    err := ca.UpdateTransport()
    if err != nil {
        log.Fatal(err)
    }

    // Now all HTTP clients will trust CA-issued certificates
    client := &http.Client{}
    resp, err := client.Get("https://my-service.local")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()
    
    log.Printf("Response status: %s", resp.Status)
}
```

### Optional Transport Setup

```go
package main

import (
    "log"
    "net/http"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Only configure CA transport if SGL_CA is set
    // No error if environment variables are not configured
    err := ca.UpdateTransportOnlyIf()
    if err != nil {
        log.Printf("Failed to configure CA transport: %v", err)
        // Continue without CA - will use system certificates
    }

    // This works whether CA is configured or not
    client := &http.Client{}
    resp, err := client.Get("https://my-service.local")
    if err != nil {
        log.Printf("Request failed: %v", err)
        return
    }
    defer resp.Body.Close()
    
    log.Printf("Response status: %s", resp.Status)
}
```

### Create Secure HTTPS Server

```go
package main

import (
    "log"
    "net/http"
    "os"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Set CA service URL
    os.Setenv("SGL_CA", "http://localhost:8090")
    os.Setenv("SGL_CA_API_KEY", "my-secure-api-key")
    
    // Create handler
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from secure server!"))
    })
    
    // Create HTTPS server with automatic certificates
    server, err := ca.CreateSecureHTTPSServer(
        "my-web-service",      // Service name
        "127.0.0.1",          // Service IP
        "8443",               // Port
        []string{"localhost", "my-service.local"}, // Domains
        mux,                  // Handler
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Starting secure HTTPS server on :8443")
    log.Fatal(server.ListenAndServeTLS("", ""))
}
```

### Create Secure gRPC Server

```go
package main

import (
    "log"
    "net"
    "os"
    "google.golang.org/grpc"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Set CA service URL  
    os.Setenv("SGL_CA", "http://localhost:8090")
    os.Setenv("SGL_CA_API_KEY", "my-secure-api-key")
    
    // Create secure gRPC server with automatic certificates
    server, err := ca.CreateSecureGRPCServer(
        "my-grpc-service",    // Service name
        "127.0.0.1",         // Service IP
        []string{"localhost", "grpc.local"}, // Domains
    )
    if err != nil {
        log.Fatal(err)
    }

    // Register your gRPC services here
    // pb.RegisterMyServiceServer(server, &myServiceImpl{})

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Starting secure gRPC server on :50051")
    log.Fatal(server.Serve(lis))
}
```

## API Reference

### Core Types

#### CAConfig
Configuration for Certificate Authority:
```go
type CAConfig struct {
    CertValidDays    int              // Certificate validity period in days (default: 365)
    KeySize         int              // RSA key size in bits (default: 2048)
    Organization    string           // Certificate organization (default: "Default Org")
    Country         string           // Certificate country (default: "US")
    PersistDir      string           // Directory for persistent storage (optional)
    StorageBackend  StorageBackend   // Custom storage backend (optional)
}
```

#### ServerConfig  
Configuration for HTTP server:
```go
type ServerConfig struct {
    Port      string     // Server port (default: "8090")
    CAConfig  *CAConfig  // CA configuration
    EnableGUI bool       // Enable web GUI (default: true)
    GUIAPIKey string     // API key for GUI protection (optional)
}
```

#### CertRequest
Request for certificate issuance:
```go
type CertRequest struct {
    ServiceName string   `json:"service_name"` // Service identifier
    ServiceIP   string   `json:"service_ip"`   // Service IP address
    Domains     []string `json:"domains"`      // Domain names for certificate
}
```

#### CertResponse
Response from certificate issuance:
```go
type CertResponse struct {
    Certificate string `json:"certificate"` // PEM-encoded certificate
    PrivateKey  string `json:"private_key"` // PEM-encoded private key
    CACert      string `json:"ca_cert"`     // PEM-encoded CA certificate
}
```

### Certificate Authority

#### CA Struct
The main Certificate Authority structure for issuing certificates.

**Constructor:**
- `NewCA(config *CAConfig) (*CA, error)` - Create new CA with optional persistence

**Certificate Methods:**
- `IssueServiceCertificate(req CertRequest) (*CertResponse, error)` - Issue certificate from request
- `GetCACertificate() (string, error)` - Get CA certificate in PEM format
- `ListCertificates() ([]CertificateInfo, error)` - List all issued certificates
- `GenerateCertificate(serviceName, serviceIP string, domains []string) (string, string, error)` - Legacy method

**Information Methods:**
- `GetIssuedCertificates() []*IssuedCert` - Get all issued certificates (legacy)
- `GetCertificateBySerial(serial string) (*IssuedCert, bool)` - Get certificate by serial number
- `GetCertificateCount() int` - Get count of issued certificates
- `GetCAInfo() map[string]interface{}` - Get CA information

### Server Functions

#### Server Struct
HTTP server wrapper for CA with web UI and REST API.

**Constructor:**
- `NewServer(config *ServerConfig) (*Server, error)` - Create new server with optional API key protection

**Server Methods:**
- `Start() error` - Start HTTP server (blocking)
- `Stop() error` - Stop HTTP server gracefully
- `GetCA() *CA` - Get underlying CA instance

### Transport Integration

#### UpdateTransport
Configures the default HTTP client to trust CA certificates by fetching the CA certificate from a CA server and adding it to the trusted root CAs.

```go
func UpdateTransport() error
```

**Environment Variables Used:**
- `SGL_CA` (required): CA server URL (must be http:// or https://)
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

**Global Variables Modified:**
- `http.DefaultClient.Transport`: Replaced with custom transport trusting the CA
- `http.DefaultTransport`: Replaced with the same custom transport

This ensures that both direct usage of `http.DefaultClient` and libraries that create HTTP clients based on `http.DefaultTransport` will trust the CA certificate.

Returns an error if `SGL_CA` is not set, invalid, or if the CA certificate cannot be fetched or parsed.

#### UpdateTransportOnlyIf
Configures the default HTTP client to trust CA certificates only if the `SGL_CA` environment variable is set. This is a conditional version of `UpdateTransport` that gracefully handles the case where no CA server is configured.

```go
func UpdateTransportOnlyIf() error
```

**Environment Variables Used:**
- `SGL_CA` (optional): CA server URL (must be http:// or https://) - if not set, function returns nil
- `SGL_CA_API_KEY` (optional): API key for CA server authentication (only used if SGL_CA is set)

**Global Variables Modified (only if SGL_CA is set):**
- `http.DefaultClient.Transport`: Replaced with custom transport trusting the CA
- `http.DefaultTransport`: Replaced with the same custom transport

Returns nil without error if `SGL_CA` is not set (no-op). Returns an error if `SGL_CA` is set but invalid, or if the CA certificate cannot be fetched or parsed.

#### RequestCertificate
Requests a certificate from the CA server for a given service. The certificate includes the service name, IP address, and additional domain names.

```go
func RequestCertificate(serviceName, serviceIP string, domains []string) (*CertResponse, error)
```

**Environment Variables Used:**
- `SGL_CA` (required): CA server URL (must be http:// or https://)
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

**Parameters:**
- `serviceName`: Name of the service requesting the certificate
- `serviceIP`: IP address of the service
- `domains`: Additional domain names to include in the certificate

Returns a `CertResponse` containing the PEM-encoded certificate and private key, or an error if the request fails or authentication is required but invalid.

#### CreateSecureHTTPSServer
Creates an HTTPS server with certificates from the CA. This is a convenience method that requests certificates from the CA server and returns a configured HTTP server ready to serve HTTPS traffic.

```go
func CreateSecureHTTPSServer(serviceName, serviceIP, port string, domains []string, handler http.Handler) (*http.Server, error)
```

**Environment Variables Used:**
- `SGL_CA` (required): CA server URL for certificate requests
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

**Parameters:**
- `serviceName`: Name of the service for certificate generation
- `serviceIP`: IP address of the service
- `port`: Port number the server will listen on (without ":")
- `domains`: Additional domain names to include in the certificate
- `handler`: HTTP handler for the server

Returns a configured `*http.Server` with TLS certificates, ready to call `ListenAndServeTLS()`.

#### CreateSecureGRPCServer
Creates a gRPC server with certificates from the CA. This is a convenience method that requests certificates from the CA server and returns a configured gRPC server with TLS transport credentials.

```go
func CreateSecureGRPCServer(serviceName, serviceIP string, domains []string, opts ...grpc.ServerOption) (*grpc.Server, error)
```

**Environment Variables Used:**
- `SGL_CA` (required): CA server URL for certificate requests
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

**Parameters:**
- `serviceName`: Name of the service for certificate generation
- `serviceIP`: IP address of the service
- `domains`: Additional domain names to include in the certificate
- `opts`: Additional gRPC server options (TLS credentials will be appended)

Returns a configured `*grpc.Server` with TLS credentials, ready to serve.

#### CreateGRPCCredentials
Returns gRPC TLS credentials using CA certificates. This is a convenience method for gRPC clients that need to connect to servers with CA-issued certificates.

```go
func CreateGRPCCredentials() (credentials.TransportCredentials, error)
```

**Environment Variables Used:**
- `SGL_CA` (required): CA server URL to fetch the CA certificate from
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

Returns `credentials.TransportCredentials` that can be used with `grpc.WithTransportCredentials()` for secure gRPC client connections.

#### UpdateGRPCDialOptions
Returns configured gRPC dial options to trust CA certificates. This is a convenience method for gRPC clients that need to dial servers with CA-issued certificates.

```go
func UpdateGRPCDialOptions() ([]grpc.DialOption, error)
```

**Environment Variables Used:**
- `SGL_CA` (required): CA server URL to fetch the CA certificate from
- `SGL_CA_API_KEY` (optional): API key for CA server authentication

Returns a slice of `grpc.DialOption` that can be passed to `grpc.Dial()` or `grpc.NewClient()` to establish secure connections to gRPC servers with CA-issued certificates.

**Example:**
```go
opts, err := ca.UpdateGRPCDialOptions()
if err != nil { 
    return err 
}
conn, err := grpc.Dial("server:443", opts...)
```

#### Legacy Transport Functions
- `SetupFromCAFile(caFilePath string) (func(), error)` - Setup HTTP transport from CA file
- `SetupFromCAService(caServiceURL string) (func(), error)` - Setup from CA service  
- `SetupFromCABytes(caCertificate []byte) (func(), error)` - Setup from CA bytes
- `SetupWithDefaults() (func(), error)` - Setup with environment defaults

### gRPC and HTTPS Integration

#### CreateSecureGRPCServer
Create a gRPC server with automatic certificates:
```go
func CreateSecureGRPCServer(serviceName, serviceIP string, domains []string) (*grpc.Server, error)
```

#### CreateSecureHTTPSServer
Create an HTTPS server with automatic certificates:
```go
func CreateSecureHTTPSServer(serviceName, serviceIP, port string, domains []string, handler http.Handler) (*http.Server, error)
```

### Storage Backends

#### StorageBackend Interface
Pluggable storage interface for certificate persistence:
```go
type StorageBackend interface {
    SaveCertificate(serviceName string, cert, key string) error
    LoadCertificate(serviceName string) (cert, key string, err error)
    SaveCA(cert, key string) error
    LoadCA() (cert, key string, err error)
    ListCertificates() ([]string, error)
    DeleteCertificate(serviceName string) error
}
```

### Utility Functions

#### DefaultCAConfig
Get default CA configuration:
```go
func DefaultCAConfig() *CAConfig
```

### Environment Variables

The package recognizes these environment variables for transport configuration:

- `SGL_CA`: CA service URL (e.g., "http://localhost:8090") - **Required** for transport functions
- `SGL_CA_API_KEY`: API key for CA service authentication - **Optional** for all transport functions

**Transport Functions Using These Variables:**
- `UpdateTransport()` - Requires `SGL_CA`, optionally uses `SGL_CA_API_KEY`
- `UpdateTransportOnlyIf()` - Optional `SGL_CA`, optionally uses `SGL_CA_API_KEY` 
- `RequestCertificate()` - Requires `SGL_CA`, optionally uses `SGL_CA_API_KEY`
- `CreateSecureHTTPSServer()` - Requires `SGL_CA`, optionally uses `SGL_CA_API_KEY`
- `CreateSecureGRPCServer()` - Requires `SGL_CA`, optionally uses `SGL_CA_API_KEY`
- `CreateGRPCCredentials()` - Requires `SGL_CA`, optionally uses `SGL_CA_API_KEY`
- `UpdateGRPCDialOptions()` - Requires `SGL_CA`, optionally uses `SGL_CA_API_KEY`

### Legacy Environment Variables
- `CA_SERVICE`: Legacy CA service URL (use `SGL_CA` instead)
- `CA_CERT_PATH`: Legacy CA certificate path (for file-based setup functions)

## HTTP API Endpoints

When running the CA server, the following endpoints are available:

### GET /ca
Download the CA certificate in PEM format.

**Headers:**
- `X-API-Key`: API key (if configured)

**Response:** PEM-encoded CA certificate

### POST /cert
Request a new service certificate.

**Headers:**
- `Content-Type: application/json`
- `X-API-Key`: API key (if configured)

**Request Body:**
```json
{
    "service_name": "my-service",
    "service_ip": "192.168.1.100", 
    "domains": ["api.example.com", "service.local"]
}
```

**Response:**
```json
{
    "certificate": "-----BEGIN CERTIFICATE-----\n...",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...",
    "ca_cert": "-----BEGIN CERTIFICATE-----\n..."
}
```

### GET /certs
List all issued certificates.

**Headers:**
- `X-API-Key`: API key (if configured)

**Response:**
```json
{
    "certificates": [
        {
            "service_name": "my-service",
            "domains": ["api.example.com"],
            "issued_at": "2024-01-01T00:00:00Z",
            "expires_at": "2025-01-01T00:00:00Z",
            "serial_number": "123456789"
        }
    ]
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
    "status": "healthy",
    "ca_info": {
        "issued_certificates": 5,
        "ca_subject": "CN=Certificate Authority"
    }
}
```

### Web GUI
When `EnableGUI` is true, a web interface is available at the root path (`/`).

**Access:** 
- Without API key: `http://localhost:8090/`
- With API key: `http://localhost:8090/?api_key=your-api-key`

**Features:**
- View CA certificate and information
- Generate new certificates
- List issued certificates
- Download certificates and keys

### Web UI Endpoints
- `GET /` or `GET /ui/` - Dashboard
- `GET /ui/certs` - List all issued certificates
- `GET /ui/generate` - Generate new certificate form
- `GET /ui/download-ca` - Download CA certificate

## Advanced Examples

### Generate Certificates for Multiple Services

```go
package main

import (
    "log"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Create CA with persistence
    config := ca.DefaultCAConfig()
    config.PersistDir = "./certificates"
    certificateAuthority, err := ca.NewCA(config)
    if err != nil {
        log.Fatal(err)
    }

    services := []ca.CertRequest{
        {ServiceName: "web-server", ServiceIP: "10.0.1.10", Domains: []string{"web.example.com", "www.example.com"}},
        {ServiceName: "api-server", ServiceIP: "10.0.1.20", Domains: []string{"api.example.com", "v1.api.example.com"}},
        {ServiceName: "database", ServiceIP: "10.0.1.30", Domains: []string{"db.internal", "postgres.internal"}},
    }

    for _, service := range services {
        resp, err := certificateAuthority.IssueServiceCertificate(service)
        if err != nil {
            log.Printf("Failed to generate cert for %s: %v", service.ServiceName, err)
            continue
        }
        
        log.Printf("Generated certificate for %s with domains: %v", 
            service.ServiceName, service.Domains)
        
        // Save to files (or handle as needed)
        // ioutil.WriteFile(fmt.Sprintf("%s.crt", service.ServiceName), []byte(resp.Certificate), 0644)
        // ioutil.WriteFile(fmt.Sprintf("%s.key", service.ServiceName), []byte(resp.PrivateKey), 0600)
    }
}
```

### Custom Storage Backend

```go
package main

import (
    "fmt"
    "log"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

// Custom storage backend using a database
type DatabaseStorage struct {
    // database connection, etc.
}

func (ds *DatabaseStorage) SaveCertificate(serviceName string, cert, key string) error {
    // Save to database
    fmt.Printf("Saving certificate for %s to database\n", serviceName)
    return nil
}

func (ds *DatabaseStorage) LoadCertificate(serviceName string) (cert, key string, err error) {
    // Load from database
    return "", "", fmt.Errorf("not found")
}

func (ds *DatabaseStorage) SaveCA(cert, key string) error {
    // Save CA to database
    return nil
}

func (ds *DatabaseStorage) LoadCA() (cert, key string, err error) {
    // Load CA from database
    return "", "", fmt.Errorf("not found")
}

func (ds *DatabaseStorage) ListCertificates() ([]string, error) {
    // List from database
    return []string{}, nil
}

func (ds *DatabaseStorage) DeleteCertificate(serviceName string) error {
    // Delete from database
    return nil
}

func main() {
    config := ca.DefaultCAConfig()
    config.StorageBackend = &DatabaseStorage{}
    
    certificateAuthority, err := ca.NewCA(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use CA with custom storage
    req := ca.CertRequest{
        ServiceName: "my-service",
        ServiceIP:   "192.168.1.100",
        Domains:     []string{"service.local"},
    }
    
    resp, err := certificateAuthority.IssueServiceCertificate(req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Certificate issued and saved to custom storage: %s", resp.Certificate[:50])
}
```

### Complete Docker-Compose Integration

```yaml
# docker-compose.yml
version: '3.8'
services:
  ca-server:
    build: .
    ports:
      - "8090:8090"
    environment:
      - CA_API_KEY=secure-api-key-here
    volumes:
      - ca-data:/app/ca-data
    command: ["./ca-server"]

  web-service:
    build: ./web-service
    ports:
      - "8443:8443"
    environment:
      - SGL_CA=http://ca-server:8090
      - SGL_CA_API_KEY=secure-api-key-here
    depends_on:
      - ca-server

volumes:
  ca-data:
```

```go
// main.go for ca-server
package main

import (
    "log"
    "os"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    config := &ca.ServerConfig{
        Port:      "8090",
        CAConfig:  ca.DefaultCAConfig(),
        EnableGUI: true,
        GUIAPIKey: os.Getenv("CA_API_KEY"),
    }
    
    // Enable persistence
    config.CAConfig.PersistDir = "./ca-data"
    
    server, err := ca.NewServer(config)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Starting CA server with persistence and API key protection...")
    log.Fatal(server.Start())
}
```

### Production Configuration

```go
package main

import (
    "log"
    "time"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Production-ready CA configuration
    config := &ca.CAConfig{
        CertValidDays: 90,        // 3 months validity
        KeySize:      4096,       // Strong encryption
        Organization: "My Company",
        Country:      "US",
        PersistDir:   "/var/lib/ca", // Persistent storage
    }
    
    serverConfig := &ca.ServerConfig{
        Port:      "8090",
        CAConfig:  config,
        EnableGUI: true,
        GUIAPIKey: "production-secure-key-change-me", // Use env var in production
    }
    
    server, err := ca.NewServer(serverConfig)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Starting production CA server...")
    log.Println("- Certificate validity: 90 days")
    log.Println("- Key size: 4096 bits")  
    log.Println("- Persistent storage: /var/lib/ca")
    log.Println("- GUI protected with API key")
    
    log.Fatal(server.Start())
}
```
    // Save certificate and key files
    saveToFile(fmt.Sprintf("%s.crt", service.name), cert)
    saveToFile(fmt.Sprintf("%s.key", service.name), key)
}
```

### Using with Docker Compose

```yaml
version: '3.8'
services:
  ca:
    image: my-ca-server
    ports:
      - "8090:8090"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8090/health"]
      interval: 10s
      timeout: 5s
      retries: 3
```

### Client Certificate Request

```bash
# Request certificate via API
curl -X POST http://localhost:8090/cert \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "my-service",
    "service_ip": "192.168.1.100",
    "domains": ["my-service.local", "api.my-service.local"]
  }'
```

### HTTP Transport Setup

```go
import "github.com/nzions/sharedgolibs/pkg/ca"

// Start CA server
server, _ := ca.NewServer(nil)
go server.Start()

// Configure HTTP transport to use CA
cleanup, err := ca.SetupFromCAService("localhost:8090")
if err != nil {
    log.Fatal(err)
}
defer cleanup()

// Now all HTTP clients will use the CA certificate
client := &http.Client{}
resp, err := client.Get("https://my-service.local")
```

## Testing

Run the comprehensive test suite:

```bash
cd pkg/ca
go test -v
```

Tests cover:
- CA creation and initialization
- Certificate generation with various configurations
- Concurrent certificate generation
- API request/response parsing
- Certificate validation and verification
- Thread safety
- HTTP transport monkey-patching

## Thread Safety

All CA operations are thread-safe and can be used concurrently:

```go
// Safe to call from multiple goroutines
go func() {
    cert1, _, _ := ca.GenerateCertificate("service1", "10.0.1.1", []string{"service1.local"})
}()

go func() {
    cert2, _, _ := ca.GenerateCertificate("service2", "10.0.1.2", []string{"service2.local"})
}()
```

## Best Practices

1. **Use Default Configurations**: Start with `DefaultCAConfig()` and customize as needed
2. **Certificate Validity**: Service certificates are valid for 30 days - implement rotation
3. **CA Certificate Storage**: Save the CA certificate for client verification
4. **Health Monitoring**: Use the `/health` endpoint for monitoring
5. **Graceful Shutdown**: Implement proper cleanup in production
6. **Environment Variables**: Use environment-driven configuration for flexibility

## Semantic Versioning

This package follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (x.0.0): Breaking changes that require code updates
- **MINOR** (0.x.0): New features that are backwards-compatible  
- **PATCH** (0.0.x): Bug fixes and backwards-compatible improvements

### Version History

- **1.0.0**: Initial release
  - Complete Certificate Authority implementation
  - HTTP transport monkey-patching
  - HTTP server with REST API and Web UI
  - Thread-safe operations
  - Comprehensive test coverage
