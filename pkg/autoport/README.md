# Auto Port Configuration

Auto-generated port configurations from Docker Compose for consistent service discovery across all applications.

## Overview

This package provides auto-generated Go configuration based on the `docker-compose.yml` file from the googleemu project. It extracts:

- **Service Names**: Docker service names
- **Port Mappings**: External to internal port mappings
- **Health Endpoints**: Automatically determined health check URLs
- **Security Information**: HTTPS/HTTP determination
- **Network Aliases**: Service aliases for internal networking
- **Dependencies**: Service dependency mapping
- **Environment**: Service environment variables

## Generation

This file is **auto-generated** by the `servicemanager` tool. Do not edit manually.

To regenerate:

```bash
# From sharedgolibs directory
./bin/servicemanager -generate=docker-compose.yml
```

### Custom Generation

```bash
# Custom paths
./bin/servicemanager -generate=/path/to/docker-compose.yml
```

## Usage

### Basic Service Discovery

```go
package main

import (
    "fmt"
    "github.com/nzions/sharedgolibs/pkg/autoport"
)

func main() {
    // Get all available ports
    ports := autoport.GetAllPorts()
    for _, port := range ports {
        if service, found := autoport.GetServiceByPort(port); found {
            fmt.Printf("Port %d: %s (%s)\n", port, service.Name, service.Image)
        }
    }
}
```

### Service Lookup

```go
// Look up service by port
if service, found := autoport.GetServiceByPort(8080); found {
    fmt.Printf("Frontend: %s on %d->%d\n", 
        service.Name, service.ExternalPort, service.InternalPort)
    
    // Health check
    fmt.Printf("Health URL: %s\n", service.HealthPath)
    
    // Check if secure
    if service.IsSecure {
        fmt.Println("Uses HTTPS")
    }
}

// Look up service by name
if service, found := autoport.GetServiceByName("ca"); found {
    fmt.Printf("CA Service on port %d\n", service.ExternalPort)
}
```

### Configuration Information

```go
config := autoport.GetConfiguration()
fmt.Printf("Generated: %s\n", config.Generated.Format("2006-01-02 15:04:05"))
fmt.Printf("Total services: %d\n", len(config.Services))
```

## Generated Services

| Port | Service      | Image               | Internal Port | Secure | Health URL               |
| ---- | ------------ | ------------------- | ------------- | ------ | ------------------------ |
| 80   | portdash     | portdash:latest     | 80            | No     | http://localhost/health  |
| 8080 | amt-frontend | amt-frontend:latest | 8080          | No     | http://localhost/health  |
| 8081 | amt-backend  | amt-backend:latest  | 8080          | No     | http://localhost/health  |
| 8082 | gcs          | gcs:latest          | 443           | Yes    | https://localhost/health |
| 8083 | secrets      | secrets:latest      | 443           | Yes    | https://localhost/health |
| 8084 | gmail        | gmail:latest        | 443           | Yes    | https://localhost/health |
| 8086 | gcr          | gcr:latest          | 443           | Yes    | https://localhost/health |
| 8087 | openai       | openai:latest       | 443           | Yes    | https://localhost/health |
| 8088 | metadata     | metadata:latest     | 80            | No     | http://localhost/        |
| 8089 | ca           | ca:latest           | 8090          | No     | http://localhost/health  |
| 8090 | firebase     | firebase:latest     | 443           | Yes    | https://localhost/health |

## Service Dependencies

The configuration includes dependency mapping from the docker-compose.yml:

```go
if service, found := autoport.GetServiceByName("amt-frontend"); found {
    fmt.Printf("Dependencies: %v\n", service.DependsOn)
    // Output: Dependencies: [firebase gcs gmail secrets metadata amt-backend ca]
}
```

## Network Aliases

Services with network aliases can be accessed via multiple hostnames:

```go
if service, found := autoport.GetServiceByName("firebase"); found {
    fmt.Printf("Aliases: %v\n", service.Aliases)
    // Output: Aliases: [firebase.googleapis.com firestore.googleapis.com ...]
}
```

## Integration Examples

### HTTP Client Configuration

```go
// Configure HTTP client based on service security
if service, found := autoport.GetServiceByPort(8082); found {
    baseURL := fmt.Sprintf("http://localhost:%d", service.ExternalPort)
    if service.IsSecure {
        baseURL = fmt.Sprintf("https://localhost:%d", service.ExternalPort)
    }
    
    // Use baseURL for API calls
    fmt.Printf("GCS Emulator: %s\n", baseURL)
}
```

### Health Check Monitoring

```go
// Health check all services
for _, port := range autoport.GetAllPorts() {
    if service, found := autoport.GetServiceByPort(port); found {
        if service.HealthPath != "" {
            // Perform health check
            fmt.Printf("Checking %s at %s\n", service.Name, service.HealthPath)
        }
    }
}
```

### Service Discovery for Load Balancers

```go
// Generate load balancer configuration
services := autoport.GetServiceNames()
for _, name := range services {
    if service, found := autoport.GetServiceByName(name); found {
        fmt.Printf("upstream %s {\n", name)
        fmt.Printf("    server localhost:%d;\n", service.ExternalPort)
        fmt.Printf("}\n")
    }
}
```

## Metadata

- **Version**: v0.1.0
- **Auto-generated**: Updated when `servicemanager -generate` is run
- **Source**: docker-compose.yml from googleemu project
- **Dependencies**: None (pure Go standard library)

## Related Tools

- [`pkg/servicemanager`](../servicemanager/README.md): Unified service management with Docker integration
- [`cmd/servicemanager`](../../cmd/servicemanager/README.md): CLI tool for comprehensive service discovery and management
