# Service Manager

A comprehensive Go package for discovering and managing processes, containers, and services running on specific ports in development environments. This package unifies port discovery, Docker container management, and local process management into a single, modular object-oriented interface.

## Features

- **Docker Integration**: Automatic detection of Docker containers with support for standard Docker and Colima
- **Process Discovery**: Local process detection and management using system tools
- **SSH Detection**: Intelligent identification of Docker port forwarding via SSH processes  
- **Service Categorization**: Expected vs unexpected service detection using autoport configuration
- **Port Management**: Comprehensive port scanning and monitoring capabilities
- **Multi-Environment Support**: Works with various Docker installations and development setups
- **Object-Oriented Design**: Clean, modular API with functional options pattern
- **Auto-configuration**: Generates autoport configuration from docker-compose.yml files

## Installation

```bash
go get github.com/nzions/sharedgolibs/pkg/servicemanager
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/nzions/sharedgolibs/pkg/servicemanager"
)

func main() {
    // Create a new service manager with default configuration
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
}
```

### Custom Configuration

```go
// Create service manager with custom options
sm := servicemanager.New(
    servicemanager.WithPortRange(3000, 9000),
    servicemanager.WithKnownService(3000, "My API", "http://localhost:3000/health", false),
    servicemanager.WithMonitoredPort(3001, "Frontend"),
    servicemanager.WithDockerTimeout(10*time.Second),
)
```

### Simple Mode (No Docker)

```go
// Create without Docker integration for simpler use cases
sm := servicemanager.NewSimple(
    servicemanager.WithPortRange(8000, 9000),
    servicemanager.WithMonitoredPort(8080, "My Service"),
)
```

## Core Concepts

### Service Types

- **Docker Container** (`ServiceTypeDockerContainer`): Services running in Docker containers
- **Local Process** (`ServiceTypeLocalProcess`): Services running as local processes
- **Unknown** (`ServiceTypeUnknown`): Services that couldn't be categorized

### Service Discovery

The service manager uses a priority-based discovery system:

1. **Docker containers** (if Docker is available)
2. **Real-time Docker check** for dynamic containers
3. **Local process detection** using `lsof`
4. **SSH process detection** for Docker port forwarding

### Expected vs Unexpected Services

Services are categorized as "expected" or "unexpected" based on autoport configuration:

- **Expected**: Services defined in docker-compose.yml or autoport configuration
- **Unexpected**: Services running on ports not in the expected configuration

## API Reference

### Creating Service Managers

#### `New(options ...ManagerOption) *ServiceManager`

Creates a new ServiceManager with Docker integration and default configurations.

#### `NewSimple(options ...ManagerOption) *ServiceManager`

Creates a ServiceManager without Docker integration for simpler use cases.

### Configuration Options

#### `WithPortRange(start, end int) ManagerOption`

Sets the port scanning range.

```go
sm := servicemanager.New(servicemanager.WithPortRange(3000, 4000))
```

#### `WithKnownService(port int, name, healthURL string, isSecure bool) ManagerOption`

Adds a known service configuration.

```go
sm := servicemanager.New(
    servicemanager.WithKnownService(8080, "API Server", "http://localhost:8080/health", false),
)
```

#### `WithMonitoredPort(port int, description string) ManagerOption`

Adds a port to the monitored ports list.

```go
sm := servicemanager.New(
    servicemanager.WithMonitoredPort(3000, "Development Server"),
)
```

#### `WithDockerTimeout(timeout time.Duration) ManagerOption`

Sets the Docker client timeout.

```go
sm := servicemanager.New(servicemanager.WithDockerTimeout(10*time.Second))
```

### Service Discovery Methods

#### `DiscoverAllServices() ([]ServiceInfo, error)`

Discovers all services running on monitored ports.

#### `DiscoverExpectedServices() ([]ServiceInfo, error)`

Returns only services that are expected according to autoport configuration.

#### `DiscoverUnexpectedServices() ([]ServiceInfo, error)`

Returns only services that are NOT expected according to autoport configuration.

#### `DiscoverDockerServices() ([]ServiceInfo, error)`

Discovers only Docker container services.

#### `DiscoverLocalServices() []ServiceInfo`

Discovers only local process services.

### Port and Service Management

#### `CheckPort(port int) (*ServiceInfo, error)`

Checks if a specific port has a service listening and returns detailed info.

#### `CheckMonitoredPorts() (*ServiceStatus, error)`

Checks all monitored ports and returns their status (legacy processmanager compatibility).

#### `GetServiceStatus() (*ServiceStatus, error)`

Returns comprehensive status of all expected and discovered services.

#### `GetMissingServices() []autoport.ServiceConfig`

Returns expected services that are not currently running.

### Service Control

#### `KillServiceOnPort(port int) error`

Kills the service (container or process) on a specific port.

#### `KillDockerContainer(containerNameOrID string) error`

Stops a Docker container by name or ID.

#### `KillAllServices() []error`

Kills all services listening on monitored ports.

### Configuration Management

#### `AddMonitoredPort(port int, description string)`

Adds a port to the monitored ports list.

#### `RemoveMonitoredPort(port int)`

Removes a port from the monitored ports list.

#### `GetMonitoredPorts() []int`

Returns all monitored ports.

#### `GetPortDescription(port int) string`

Returns the description for a monitored port.

#### `AddKnownService(port int, name, healthURL string, isSecure bool)`

Adds a service configuration for a specific port.

#### `SetPortRange(start, end int)`

Updates the port scanning range.

#### `GetPortRange() PortRange`

Returns the current port scanning range.

### Docker Configuration

#### `IsDockerAvailable() bool`

Returns whether Docker is available.

#### `GetDockerSocketPath() string`

Returns the Docker socket path being used.

#### `GetDockerConfig() *DockerConfig`

Returns the current Docker configuration.

## Data Structures

### ServiceInfo

```go
type ServiceInfo struct {
    Name          string      `json:"name"`
    Type          ServiceType `json:"type"`
    ExternalPort  int         `json:"external_port"`
    InternalPort  int         `json:"internal_port,omitempty"`
    PID           string      `json:"pid,omitempty"`
    Command       string      `json:"command,omitempty"`
    ContainerID   string      `json:"container_id,omitempty"`
    Image         string      `json:"image,omitempty"`
    Status        string      `json:"status"`
    Uptime        string      `json:"uptime,omitempty"`
    IsListening   bool        `json:"is_listening"`
    HealthURL     string      `json:"health_url,omitempty"`
    IsExpected    bool        `json:"is_expected"`
    ExpectedImage string      `json:"expected_image,omitempty"`
    ImageMatches  bool        `json:"image_matches"`
    Description   string      `json:"description,omitempty"`
}
```

### ServiceStatus

```go
type ServiceStatus struct {
    Running       []ServiceInfo            `json:"running"`
    Missing       []autoport.ServiceConfig `json:"missing"`
    Expected      int                      `json:"expected_count"`
    Unexpected    int                      `json:"unexpected_count"`
    ImageMatch    int                      `json:"image_match_count"`
    ImageMismatch int                      `json:"image_mismatch_count"`
    Total         int                      `json:"total_count"`
    Listening     int                      `json:"listening_count"`
}
```

### DockerConfig

```go
type DockerConfig struct {
    Client      *client.Client
    Available   bool
    SocketPath  string
    Timeout     time.Duration
}
```

## Docker Compose Integration

### Generate Autoport Configuration

```go
// Generate autoport configuration from docker-compose.yml
err := sm.GenerateAutoPortConfig("docker-compose.yml", "pkg/autoport/autoport.go")
if err != nil {
    log.Fatal(err)
}
```

This reads your docker-compose.yml file and generates Go code for the autoport package, enabling automatic service detection and categorization.

## Examples

### Monitor Development Environment

```go
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
```

### Kill All Development Services

```go
func killAllServices() {
    sm := servicemanager.New()
    
    errors := sm.KillAllServices()
    if len(errors) > 0 {
        fmt.Println("Errors occurred while killing services:")
        for _, err := range errors {
            fmt.Printf("  - %v\n", err)
        }
    } else {
        fmt.Println("All services killed successfully.")
    }
}
```

### Check Specific Port

```go
func checkSpecificPort(port int) {
    sm := servicemanager.New()
    
    service, err := sm.CheckPort(port)
    if err != nil {
        fmt.Printf("No service on port %d: %v\n", port, err)
        return
    }
    
    fmt.Printf("Port %d: %s (%s)\n", port, service.Name, service.Type)
    if service.Type == servicemanager.ServiceTypeDockerContainer {
        fmt.Printf("  Container: %s\n", service.ContainerID)
        fmt.Printf("  Image: %s\n", service.Image)
        fmt.Printf("  Uptime: %s\n", service.Uptime)
    } else {
        fmt.Printf("  PID: %s\n", service.PID)
        fmt.Printf("  Command: %s\n", service.Command)
    }
}
```

## Migration from portmanager/processmanager

The unified servicemanager provides backward compatibility with both portmanager and processmanager:

### From portmanager

```go
// Old: portmanager.New()
// New: servicemanager.New()

// Old: pm.DiscoverAllServices()
// New: sm.DiscoverAllServices()

// Old: pm.IsDockerAvailable()
// New: sm.IsDockerAvailable()
```

### From processmanager

```go
// Old: processmanager.New()
// New: servicemanager.New()

// Old: pm.CheckAllPorts()
// New: sm.CheckMonitoredPorts()

// Old: pm.AddPort(port, desc)
// New: sm.AddMonitoredPort(port, desc)
```

## Version

Current version: `v0.3.0`

### Recent Changes (v0.3.0)
- Added Docker Compose integration for autoport generation
- Added `GenerateAutoPortConfig()` method for creating autoport configurations from `docker-compose.yml`
- Enhanced template generation with comprehensive service metadata
- Added Makefile target for easy autoport regeneration
- Improved error handling and validation for Docker Compose parsing

This package follows [Semantic Versioning](https://semver.org/).

## Dependencies

- Docker API client: `github.com/docker/docker/client`
- YAML parsing: `gopkg.in/yaml.v3`
- Autoport integration: `github.com/nzions/sharedgolibs/pkg/autoport`

## License

This package is part of the sharedgolibs repository and follows the same license terms.
