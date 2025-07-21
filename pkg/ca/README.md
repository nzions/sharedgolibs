# CA Package

[![License: CC0-1.0](https://img.shields.io/badge/License-CC0--1.0-blue.svg)](http://creativecommons.org/publicdomain/zero/1.0/)

The CA package provides comprehensive Certificate Authority functionality for development and testing environments, enabling dynamic certificate issuance and HTTP transport monkey-patching.

## Version: 1.0.0

## Features

### Certificate Authority (authority.go)
- 🔐 **Complete CA Implementation**: Full Certificate Authority with RSA key generation and X.509 certificate creation
- 🏭 **Dynamic Certificate Generation**: Create certificates for any service or domain on-demand
- 🌐 **Web UI**: User-friendly interface for certificate management
- 📡 **REST API**: Programmatic certificate issuance
- 📋 **Certificate Store**: Track all issued certificates with thread-safe operations
- 🔄 **Concurrent Safe**: Thread-safe operations for multi-service environments

### HTTP Transport Monkey-Patching (monkeypatch.go)
- 🔧 **Zero-code-change**: Modify global HTTP transport without changing application code
- 🛡️ **CA Trust**: Automatically trust certificates signed by custom CA
- 🔄 **Reversible**: Restore original transport when emulation is complete
- ⚙️ **Environment-driven**: Configuration through environment variables

## Quick Start

### Basic CA Usage

```go
package main

import (
    "log"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Create a new Certificate Authority
    certificateAuthority, err := ca.NewCA(nil) // Uses default config
    if err != nil {
        log.Fatal(err)
    }

    // Generate a certificate for a service
    certPEM, keyPEM, err := certificateAuthority.GenerateCertificate(
        "my-service",
        "192.168.1.100", 
        []string{"api.example.com", "service.local"},
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Certificate: %s", certPEM)
    log.Printf("Private Key: %s", keyPEM)
}
```

### HTTP Server Usage

```go
package main

import (
    "log"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Create and start CA server
    server, err := ca.NewServer(nil) // Uses default config
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Starting CA server on port 8090...")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### HTTP Transport Monkey-Patching

```go
package main

import (
    "log"
    "os"
    "github.com/nzions/sharedgolibs/pkg/ca"
)

func main() {
    // Setup HTTP transport to trust custom CA
    cleanup, err := ca.SetupFromCAFile("/path/to/ca.crt")
    if err != nil {
        log.Fatal(err)
    }
    defer cleanup()

    // Now all HTTP clients will trust the custom CA
    // Your existing HTTP code works unchanged
}
```

## API Reference

### Certificate Authority

#### CA Struct
The main Certificate Authority structure for issuing certificates.

**Methods:**
- `NewCA(config *CAConfig) (*CA, error)` - Create new CA
- `Certificate() *x509.Certificate` - Get CA certificate
- `CertificatePEM() []byte` - Get CA certificate in PEM format
- `GenerateCertificate(serviceName, serviceIP string, domains []string) (string, string, error)` - Generate service certificate
- `IssueServiceCertificate(req CertRequest) (*CertResponse, error)` - Issue certificate from request
- `GetIssuedCertificates() []*IssuedCert` - Get all issued certificates
- `GetCertificateBySerial(serial string) (*IssuedCert, bool)` - Get certificate by serial number
- `GetCertificateCount() int` - Get count of issued certificates
- `GetCAInfo() map[string]interface{}` - Get CA information

#### Server Struct
HTTP server wrapper for CA with web UI and REST API.

**Methods:**
- `NewServer(config *ServerConfig) (*Server, error)` - Create new server
- `Start() error` - Start HTTP server
- `GetCA() *CA` - Get underlying CA instance

### Configuration Types

```go
type CAConfig struct {
    Country            []string
    Province           []string
    Locality           []string
    Organization       []string
    OrganizationalUnit []string
    CommonName         string
    ValidityPeriod     time.Duration
    KeySize            int
}

type ServerConfig struct {
    Port     string
    CAConfig *CAConfig
}
```

### Request/Response Types

```go
type CertRequest struct {
    ServiceName string   `json:"service_name"`
    ServiceIP   string   `json:"service_ip"`
    Domains     []string `json:"domains"`
}

type CertResponse struct {
    Certificate string `json:"certificate"`
    PrivateKey  string `json:"private_key"`
    CACert      string `json:"ca_cert"`
}

type IssuedCert struct {
    ServiceName  string    `json:"service_name"`
    Domains      []string  `json:"domains"`
    IssuedAt     time.Time `json:"issued_at"`
    ExpiresAt    time.Time `json:"expires_at"`
    Certificate  string    `json:"certificate"`
    SerialNumber string    `json:"serial_number"`
}
```

## HTTP Transport Monkey-Patching

### Functions

- `SetupFromCAFile(caCertPath string) (func(), error)` - Setup from CA file
- `SetupFromCAService(caServiceURL string) (func(), error)` - Setup from CA service
- `SetupFromCABytes(caCertificate []byte) (func(), error)` - Setup from CA bytes
- `SetupWithDefaults() (func(), error)` - Setup with environment defaults

### Environment Variables

- `CA_SERVICE` - URL of the CA service (e.g., "ca:8090")
- `CA_CERT_PATH` - Path to CA certificate file

### Default CA Certificate Paths

- `/tmp/sharedgolibs-ca/ca.crt`
- `/etc/ssl/certs/sharedgolibs-ca.crt`
- `./ca.crt`

## HTTP API Endpoints

When running the CA server, the following endpoints are available:

### GET /ca
Download the CA certificate in PEM format.

### POST /cert
Request a new service certificate.

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

### GET /health
Health check endpoint.

### Web UI Endpoints
- `GET /` or `GET /ui/` - Dashboard
- `GET /ui/certs` - List all issued certificates
- `GET /ui/generate` - Generate new certificate form
- `GET /ui/download-ca` - Download CA certificate

## Examples

### Generate Certificates for Multiple Services

```go
services := []struct {
    name    string
    ip      string
    domains []string
}{
    {"web-server", "10.0.1.10", []string{"web.example.com", "www.example.com"}},
    {"api-server", "10.0.1.20", []string{"api.example.com", "v1.api.example.com"}},
    {"database", "10.0.1.30", []string{"db.internal", "postgres.internal"}},
}

ca, _ := ca.NewCA(nil)
for _, service := range services {
    cert, key, err := ca.GenerateCertificate(service.name, service.ip, service.domains)
    if err != nil {
        log.Printf("Failed to generate cert for %s: %v", service.name, err)
        continue
    }
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
