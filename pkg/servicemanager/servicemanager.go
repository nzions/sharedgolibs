// Package servicemanager provides comprehensive utilities for discovering and managing
// processes, containers, and services running on specific ports in development environments.
//
// This package unifies port discovery, Docker container management, and local process
// management into a single, modular OO-style interface. It provides detailed information
// about services running in development environments with Docker API integration,
// SSH detection, and intelligent service categorization.
//
// Example:
//
//	sm := servicemanager.New()
//	services, err := sm.DiscoverAllServices()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, service := range services {
//	    fmt.Printf("%s on port %d (%s)\n", service.Name, service.ExternalPort, service.Type)
//	}
package servicemanager

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/nzions/sharedgolibs/pkg/autoport"
	"gopkg.in/yaml.v3"
)

const Version = "v0.3.0"

// ServiceType represents the type of service discovered
type ServiceType string

const (
	ServiceTypeDockerContainer ServiceType = "docker"
	ServiceTypeLocalProcess    ServiceType = "local"
	ServiceTypeUnknown         ServiceType = "unknown"
)

// ServiceInfo contains comprehensive information about a discovered service
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

// ServiceConfig holds configuration for known services
type ServiceConfig struct {
	Name      string `json:"name"`
	HealthURL string `json:"health_url,omitempty"`
	IsSecure  bool   `json:"is_secure"`
}

// PortRange defines the range of ports to scan
type PortRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// ServiceStatus represents the overall status of services
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

// DockerConfig holds Docker client configuration
type DockerConfig struct {
	Client     *client.Client
	Available  bool
	SocketPath string
	Timeout    time.Duration
}

// ServiceManager manages comprehensive service discovery and management
type ServiceManager struct {
	dockerConfig     *DockerConfig
	portRange        PortRange
	knownServices    map[int]ServiceConfig
	monitoredPorts   []int
	portDescriptions map[int]string
}

// ManagerOption defines a functional option for ServiceManager configuration
type ManagerOption func(*ServiceManager)

// WithPortRange sets a custom port scanning range
func WithPortRange(start, end int) ManagerOption {
	return func(sm *ServiceManager) {
		sm.portRange = PortRange{Start: start, End: end}
	}
}

// WithDockerTimeout sets the Docker client timeout
func WithDockerTimeout(timeout time.Duration) ManagerOption {
	return func(sm *ServiceManager) {
		if sm.dockerConfig != nil {
			sm.dockerConfig.Timeout = timeout
		}
	}
}

// WithKnownService adds a known service configuration
func WithKnownService(port int, name, healthURL string, isSecure bool) ManagerOption {
	return func(sm *ServiceManager) {
		sm.knownServices[port] = ServiceConfig{
			Name:      name,
			HealthURL: healthURL,
			IsSecure:  isSecure,
		}
	}
}

// WithMonitoredPort adds a port to the monitored ports list
func WithMonitoredPort(port int, description string) ManagerOption {
	return func(sm *ServiceManager) {
		sm.monitoredPorts = append(sm.monitoredPorts, port)
		sm.portDescriptions[port] = description
	}
}

// New creates a new ServiceManager with Docker integration and default configurations
func New(options ...ManagerOption) *ServiceManager {
	sm := &ServiceManager{
		portRange:        PortRange{Start: 80, End: 9099}, // Common development port range
		knownServices:    make(map[int]ServiceConfig),
		monitoredPorts:   make([]int, 0),
		portDescriptions: make(map[int]string),
		dockerConfig: &DockerConfig{
			Timeout: 5 * time.Second,
		},
	}

	// Apply options first
	for _, option := range options {
		option(sm)
	}

	// Initialize default configurations
	sm.initializeDefaultServices()
	sm.initializeDefaultMonitoredPorts()
	sm.initializeDockerClient()

	return sm
}

// NewSimple creates a ServiceManager with minimal configuration (no Docker, custom ports only)
func NewSimple(options ...ManagerOption) *ServiceManager {
	sm := &ServiceManager{
		portRange:        PortRange{Start: 80, End: 9099},
		knownServices:    make(map[int]ServiceConfig),
		monitoredPorts:   make([]int, 0),
		portDescriptions: make(map[int]string),
		dockerConfig:     nil, // No Docker integration
	}

	// Apply options
	for _, option := range options {
		option(sm)
	}

	return sm
}

// initializeDefaultServices sets up the known service configurations from autoport and legacy services
func (sm *ServiceManager) initializeDefaultServices() {
	// Primary services from docker-compose.yml
	services := map[int]ServiceConfig{
		80:   {Name: "Port Dashboard", HealthURL: "http://localhost:80/health"},
		8080: {Name: "AMT Frontend", HealthURL: "http://localhost:8080"},
		8081: {Name: "AMT Backend", HealthURL: "http://localhost:8081/health"},
		8082: {Name: "GCS Emulator", HealthURL: "https://localhost:8082", IsSecure: true},
		8083: {Name: "Secrets Manager", HealthURL: "https://localhost:8083", IsSecure: true},
		8084: {Name: "Gmail Emulator", HealthURL: "https://localhost:8084", IsSecure: true},
		8086: {Name: "GCR Emulator", HealthURL: "https://localhost:8086", IsSecure: true},
		8087: {Name: "OpenAI Emulator", HealthURL: "https://localhost:8087", IsSecure: true},
		8088: {Name: "Metadata Service", HealthURL: "http://localhost:8088"},
		8089: {Name: "CA Service", HealthURL: "http://localhost:8089/health"},
		8090: {Name: "Firebase Emulator", HealthURL: "https://localhost:8090", IsSecure: true},

		// Legacy/standalone ports
		8025: {Name: "Gmail Emulator (Legacy)", HealthURL: "http://localhost:8025"},
		8050: {Name: "GCR Emulator (Legacy)", HealthURL: "http://localhost:8050"},
		4443: {Name: "GCS Emulator (Legacy)", HealthURL: "http://localhost:4443"},
		5005: {Name: "AI Emulator (Legacy)", HealthURL: "http://localhost:5005"},
		9000: {Name: "Secrets Manager gRPC (Legacy)", HealthURL: "http://localhost:9000"},
		9001: {Name: "Secrets Manager HTTP (Legacy)", HealthURL: "http://localhost:9001"},
		9099: {Name: "Firebase Emulator (Legacy)", HealthURL: "http://localhost:9099"},
	}

	for port, config := range services {
		sm.knownServices[port] = config
	}
}

// initializeDefaultMonitoredPorts sets up default monitored ports (from legacy processmanager)
func (sm *ServiceManager) initializeDefaultMonitoredPorts() {
	defaultPorts := map[int]string{
		8025: "Gmail Emulator",
		8050: "GCR Emulator",
		8080: "Backend (amt-backend)",
		8081: "Frontend (amt-frontend)",
		4443: "GCS Emulator",
		5005: "AI Emulator",
		9000: "Secrets Manager gRPC",
		9001: "Secrets Manager HTTP",
		9099: "Firebase Emulator",
	}

	for port, description := range defaultPorts {
		sm.monitoredPorts = append(sm.monitoredPorts, port)
		sm.portDescriptions[port] = description
	}
}

// initializeDockerClient attempts to create a Docker client with multi-environment support
func (sm *ServiceManager) initializeDockerClient() {
	if sm.dockerConfig == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), sm.dockerConfig.Timeout)
	defer cancel()

	// Try standard Docker client from environment first
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		// Test Docker connectivity
		_, err = dockerClient.Ping(ctx)
		if err == nil {
			sm.dockerConfig.Client = dockerClient
			sm.dockerConfig.Available = true
			sm.dockerConfig.SocketPath = "docker-env" // Standard Docker from environment
			return
		}
	}

	// If standard Docker isn't available, try Colima's default socket location
	homeDir, err := os.UserHomeDir()
	if err == nil {
		colimaSocketPath := fmt.Sprintf("unix://%s/.colima/default/docker.sock", homeDir)
		dockerClient, err = client.NewClientWithOpts(
			client.WithHost(colimaSocketPath),
			client.WithAPIVersionNegotiation(),
		)
		if err == nil {
			// Test Colima Docker connectivity
			_, err = dockerClient.Ping(ctx)
			if err == nil {
				sm.dockerConfig.Client = dockerClient
				sm.dockerConfig.Available = true
				sm.dockerConfig.SocketPath = colimaSocketPath
				return
			}
		}
	}

	// If both fail, Docker is not available
	sm.dockerConfig.Available = false
	sm.dockerConfig.SocketPath = ""
}

// Docker Configuration Methods

// IsDockerAvailable returns whether Docker is available
func (sm *ServiceManager) IsDockerAvailable() bool {
	return sm.dockerConfig != nil && sm.dockerConfig.Available
}

// GetDockerSocketPath returns the Docker socket path being used
func (sm *ServiceManager) GetDockerSocketPath() string {
	if sm.dockerConfig == nil {
		return ""
	}
	return sm.dockerConfig.SocketPath
}

// GetDockerConfig returns the current Docker configuration
func (sm *ServiceManager) GetDockerConfig() *DockerConfig {
	return sm.dockerConfig
}

// Service Discovery Methods

// DiscoverAllServices discovers all services running on monitored ports
func (sm *ServiceManager) DiscoverAllServices() ([]ServiceInfo, error) {
	var services []ServiceInfo

	// Get Docker containers if available
	containersByPort := make(map[int]ServiceInfo)
	dockerPorts := make(map[int]bool)
	if sm.IsDockerAvailable() {
		containers, err := sm.getDockerContainers()
		if err == nil {
			for _, container := range containers {
				if container.ExternalPort > 0 {
					containersByPort[container.ExternalPort] = container
					dockerPorts[container.ExternalPort] = true
				}
			}
		}
	}

	// Get expected services from autoport
	expectedPorts := autoport.GetAllPorts()
	expectedPortMap := make(map[int]bool)
	for _, port := range expectedPorts {
		expectedPortMap[port] = true
	}

	// Scan port range
	for port := sm.portRange.Start; port <= sm.portRange.End; port++ {
		// Check if port is listening
		if !sm.isPortListening(port) {
			continue
		}

		var service ServiceInfo

		// Priority 1: Check if we have Docker container info for this port
		if containerInfo, exists := containersByPort[port]; exists {
			service = containerInfo
		} else if sm.IsDockerAvailable() {
			// Priority 2: If Docker is available but no container found in initial scan,
			// do a real-time check (handles dynamic containers)
			if dockerService := sm.checkDockerForPort(port); dockerService != nil {
				service = *dockerService
			} else {
				// Priority 3: Fall back to process detection
				service = sm.getLocalProcessInfo(port)
			}
		} else {
			// Priority 4: Docker not available, use process detection
			service = sm.getLocalProcessInfo(port)
		}

		// Enhance with autoport configuration and monitored port descriptions
		service = sm.enhanceServiceInfo(service, expectedPortMap[port])

		services = append(services, service)
	}

	return services, nil
}

// DiscoverExpectedServices returns only services that are expected according to autoport
func (sm *ServiceManager) DiscoverExpectedServices() ([]ServiceInfo, error) {
	allServices, err := sm.DiscoverAllServices()
	if err != nil {
		return nil, err
	}

	var expectedServices []ServiceInfo
	for _, service := range allServices {
		if service.IsExpected {
			expectedServices = append(expectedServices, service)
		}
	}

	return expectedServices, nil
}

// DiscoverUnexpectedServices returns only services that are NOT expected according to autoport
func (sm *ServiceManager) DiscoverUnexpectedServices() ([]ServiceInfo, error) {
	allServices, err := sm.DiscoverAllServices()
	if err != nil {
		return nil, err
	}

	var unexpectedServices []ServiceInfo
	for _, service := range allServices {
		if !service.IsExpected {
			unexpectedServices = append(unexpectedServices, service)
		}
	}

	return unexpectedServices, nil
}

// DiscoverDockerServices discovers only Docker container services
func (sm *ServiceManager) DiscoverDockerServices() ([]ServiceInfo, error) {
	if !sm.IsDockerAvailable() {
		return nil, fmt.Errorf("docker is not available")
	}

	return sm.getDockerContainers()
}

// DiscoverLocalServices discovers only local process services
func (sm *ServiceManager) DiscoverLocalServices() []ServiceInfo {
	var services []ServiceInfo

	for port := sm.portRange.Start; port <= sm.portRange.End; port++ {
		if sm.isPortListening(port) {
			service := sm.getLocalProcessInfo(port)
			if service.Type == ServiceTypeLocalProcess {
				services = append(services, service)
			}
		}
	}

	return services
}

// Port and Service Management Methods

// CheckPort checks if a specific port has a process listening and returns detailed info
func (sm *ServiceManager) CheckPort(port int) (*ServiceInfo, error) {
	if !sm.isPortListening(port) {
		return nil, fmt.Errorf("no service listening on port %d", port)
	}

	// Try Docker first if available
	if sm.IsDockerAvailable() {
		if dockerService := sm.checkDockerForPort(port); dockerService != nil {
			enhanced := sm.enhanceServiceInfo(*dockerService, false)
			return &enhanced, nil
		}
	}

	// Fall back to local process detection
	service := sm.getLocalProcessInfo(port)
	enhanced := sm.enhanceServiceInfo(service, false)
	return &enhanced, nil
}

// CheckMonitoredPorts checks all monitored ports and returns their status (legacy processmanager compatibility)
func (sm *ServiceManager) CheckMonitoredPorts() (*ServiceStatus, error) {
	var services []ServiceInfo
	listeningCount := 0

	for _, port := range sm.monitoredPorts {
		if sm.isPortListening(port) {
			if service, err := sm.CheckPort(port); err == nil {
				// Add description from monitored ports
				if desc, exists := sm.portDescriptions[port]; exists && service.Description == "" {
					service.Description = desc
				}
				services = append(services, *service)
				listeningCount++
			}
		} else {
			// Create placeholder for non-listening monitored port
			service := ServiceInfo{
				ExternalPort: port,
				IsListening:  false,
				Status:       "not listening",
				Type:         ServiceTypeUnknown,
			}
			if desc, exists := sm.portDescriptions[port]; exists {
				service.Description = desc
				service.Name = desc
			}
			services = append(services, service)
		}
	}

	return &ServiceStatus{
		Running:   services,
		Total:     len(sm.monitoredPorts),
		Listening: listeningCount,
	}, nil
}

// GetMissingServices returns expected services that are not currently running
func (sm *ServiceManager) GetMissingServices() []autoport.ServiceConfig {
	var missingServices []autoport.ServiceConfig
	expectedPorts := autoport.GetAllPorts()

	for _, port := range expectedPorts {
		if !sm.isPortListening(port) {
			if expectedService, found := autoport.GetServiceByPort(port); found {
				missingServices = append(missingServices, expectedService)
			}
		}
	}

	return missingServices
}

// GetServiceStatus returns a comprehensive status of all expected and discovered services
func (sm *ServiceManager) GetServiceStatus() (*ServiceStatus, error) {
	allServices, err := sm.DiscoverAllServices()
	if err != nil {
		return nil, err
	}

	missingServices := sm.GetMissingServices()

	status := &ServiceStatus{
		Running:       allServices,
		Missing:       missingServices,
		Expected:      0,
		Unexpected:    0,
		ImageMatch:    0,
		ImageMismatch: 0,
		Total:         len(allServices),
		Listening:     0,
	}

	for _, service := range allServices {
		if service.IsListening {
			status.Listening++
		}
		if service.IsExpected {
			status.Expected++
			if service.ImageMatches {
				status.ImageMatch++
			} else {
				status.ImageMismatch++
			}
		} else {
			status.Unexpected++
		}
	}

	return status, nil
}

// Service Control Methods

// KillDockerContainer stops a Docker container by name or ID
func (sm *ServiceManager) KillDockerContainer(containerNameOrID string) error {
	if !sm.IsDockerAvailable() {
		return fmt.Errorf("docker is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return sm.dockerConfig.Client.ContainerKill(ctx, containerNameOrID, "SIGTERM")
}

// KillServiceOnPort kills the service (container or process) on a specific port
func (sm *ServiceManager) KillServiceOnPort(port int) error {
	service, err := sm.CheckPort(port)
	if err != nil {
		return err
	}

	switch service.Type {
	case ServiceTypeDockerContainer:
		return sm.KillDockerContainer(service.ContainerID)
	case ServiceTypeLocalProcess:
		if service.PID == "" {
			return fmt.Errorf("no PID available for process on port %d", port)
		}
		return sm.killProcess(service.PID)
	default:
		return fmt.Errorf("unknown service type: %s", service.Type)
	}
}

// KillAllServices kills all services listening on monitored ports
func (sm *ServiceManager) KillAllServices() []error {
	var errors []error

	for _, port := range sm.monitoredPorts {
		if sm.isPortListening(port) {
			if err := sm.KillServiceOnPort(port); err != nil {
				errors = append(errors, fmt.Errorf("failed to kill service on port %d: %w", port, err))
			}
		}
	}

	return errors
}

// Configuration Methods

// AddMonitoredPort adds a port to the monitored ports list
func (sm *ServiceManager) AddMonitoredPort(port int, description string) {
	sm.monitoredPorts = append(sm.monitoredPorts, port)
	sm.portDescriptions[port] = description
}

// RemoveMonitoredPort removes a port from the monitored ports list
func (sm *ServiceManager) RemoveMonitoredPort(port int) {
	// Remove from monitored ports slice
	for i, p := range sm.monitoredPorts {
		if p == port {
			sm.monitoredPorts = append(sm.monitoredPorts[:i], sm.monitoredPorts[i+1:]...)
			break
		}
	}
	// Remove from descriptions
	delete(sm.portDescriptions, port)
}

// GetMonitoredPorts returns all monitored ports
func (sm *ServiceManager) GetMonitoredPorts() []int {
	result := make([]int, len(sm.monitoredPorts))
	copy(result, sm.monitoredPorts)
	return result
}

// GetPortDescription returns the description for a monitored port
func (sm *ServiceManager) GetPortDescription(port int) string {
	if desc, exists := sm.portDescriptions[port]; exists {
		return desc
	}
	if config, exists := sm.knownServices[port]; exists {
		return config.Name
	}
	return "Unknown Service"
}

// AddKnownService adds a service configuration for a specific port
func (sm *ServiceManager) AddKnownService(port int, name, healthURL string, isSecure bool) {
	sm.knownServices[port] = ServiceConfig{
		Name:      name,
		HealthURL: healthURL,
		IsSecure:  isSecure,
	}
}

// SetPortRange updates the port scanning range
func (sm *ServiceManager) SetPortRange(start, end int) {
	sm.portRange = PortRange{Start: start, End: end}
}

// GetPortRange returns the current port scanning range
func (sm *ServiceManager) GetPortRange() PortRange {
	return sm.portRange
}

// Internal helper methods

// enhanceServiceInfo enhances service info with autoport configuration and known service data
func (sm *ServiceManager) enhanceServiceInfo(service ServiceInfo, isExpected bool) ServiceInfo {
	service.IsExpected = isExpected

	// Add description from monitored ports if available
	if desc, exists := sm.portDescriptions[service.ExternalPort]; exists && service.Description == "" {
		service.Description = desc
	}

	if isExpected {
		if expectedService, found := autoport.GetServiceByPort(service.ExternalPort); found {
			// Use autoport name if service name is generic
			if service.Name == "" || service.Name == "Unknown Service" {
				service.Name = expectedService.Name
			}

			// Set health URL from autoport
			if service.HealthURL == "" {
				service.HealthURL = expectedService.HealthPath
			}

			// Store expected image for comparison
			service.ExpectedImage = expectedService.Image

			// Check if image matches (for Docker containers)
			if service.Type == ServiceTypeDockerContainer {
				service.ImageMatches = sm.imagesMatch(service.Image, expectedService.Image)
			} else {
				// For local processes, we can't check image match
				service.ImageMatches = true
			}
		}
	} else {
		// For unexpected services, check if it might be a known service on wrong port
		service.ImageMatches = false

		// Try to identify what this might be
		if service.Type == ServiceTypeDockerContainer {
			service = sm.identifyUnexpectedService(service)
		}

		// Use known service name if available
		if config, exists := sm.knownServices[service.ExternalPort]; exists {
			if service.Name == "" || service.Name == "Unknown Service" {
				service.Name = config.Name
			}
			if service.HealthURL == "" {
				service.HealthURL = config.HealthURL
			}
		}
	}

	return service
}

// getDockerContainers retrieves information about Docker containers
func (sm *ServiceManager) getDockerContainers() ([]ServiceInfo, error) {
	if !sm.IsDockerAvailable() {
		return nil, fmt.Errorf("docker is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := sm.dockerConfig.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list Docker containers: %w", err)
	}

	var services []ServiceInfo

	for _, c := range containers {
		for _, port := range c.Ports {
			if port.PublicPort == 0 {
				continue
			}

			service := ServiceInfo{
				Type:         ServiceTypeDockerContainer,
				ExternalPort: int(port.PublicPort),
				InternalPort: int(port.PrivatePort),
				ContainerID:  c.ID[:12], // Short container ID
				Image:        c.Image,
				Status:       c.State,
				IsListening:  c.State == "running",
			}

			// Get container name
			if len(c.Names) > 0 {
				service.Name = strings.TrimPrefix(c.Names[0], "/")
			}

			// Calculate uptime for running containers
			if c.State == "running" {
				created := time.Unix(c.Created, 0)
				uptime := time.Since(created)
				service.Uptime = formatUptime(uptime)
			}

			services = append(services, service)
		}
	}

	return services, nil
}

// getLocalProcessInfo gets information about a local process on a port
func (sm *ServiceManager) getLocalProcessInfo(port int) ServiceInfo {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	output, err := cmd.CombinedOutput()

	service := ServiceInfo{
		Type:         ServiceTypeLocalProcess,
		ExternalPort: port,
		InternalPort: port,
		IsListening:  false,
		Status:       "not listening",
	}

	lines := strings.Split(string(output), "\n")
	if err == nil && len(lines) > 1 {
		service.IsListening = true
		service.Status = "running"

		// Parse first process found (skip header line)
		for _, line := range lines[1:] {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				service.PID = fields[1]
				service.Command = fields[0]
				break
			}
		}
	}

	// Check if this is an SSH process that might be Docker port forwarding
	if service.Command != "" && sm.isSSHProcess(service.Command) {
		// If Docker is available, always check for actual container info first
		if sm.IsDockerAvailable() {
			if dockerService := sm.checkDockerForPort(port); dockerService != nil {
				// Found actual Docker container - use that info instead
				dockerService.Type = ServiceTypeDockerContainer
				dockerService.ExternalPort = port
				return *dockerService
			}
		}
		// If no Docker container found or Docker not available, mark as SSH forwarding
		service.Name = "SSH (Possible Docker Forward)"
		service.Type = ServiceTypeLocalProcess
	} else {
		// Set name from known services or use command
		if config, exists := sm.knownServices[port]; exists {
			service.Name = config.Name
		} else if service.Command != "" {
			service.Name = service.Command
		} else {
			service.Name = "Unknown Service"
		}
	}

	return service
}

// isSSHProcess checks if a command is an SSH-related process
func (sm *ServiceManager) isSSHProcess(command string) bool {
	sshCommands := []string{"ssh", "sshd", "ssh-agent", "ssh-keygen", "ssh-add"}
	commandLower := strings.ToLower(command)

	for _, sshCmd := range sshCommands {
		if strings.Contains(commandLower, sshCmd) {
			return true
		}
	}

	// Also check for Docker's own SSH forwarding processes
	if strings.Contains(commandLower, "docker") && strings.Contains(commandLower, "ssh") {
		return true
	}

	return false
}

// checkDockerForPort checks if there's a Docker container that might be using this port
func (sm *ServiceManager) checkDockerForPort(port int) *ServiceInfo {
	if !sm.IsDockerAvailable() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := sm.dockerConfig.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil
	}

	// Look for containers that have this port mapped
	for _, c := range containers {
		for _, portMapping := range c.Ports {
			if int(portMapping.PublicPort) == port {
				service := &ServiceInfo{
					Type:         ServiceTypeDockerContainer,
					ExternalPort: int(portMapping.PublicPort),
					InternalPort: int(portMapping.PrivatePort),
					ContainerID:  c.ID[:12],
					Image:        c.Image,
					Status:       c.State,
					IsListening:  c.State == "running",
				}

				// Get container name (prefer container name over image name)
				if len(c.Names) > 0 {
					service.Name = strings.TrimPrefix(c.Names[0], "/")
				} else if service.Image != "" {
					// Fallback to image name if no container name
					parts := strings.Split(service.Image, ":")
					service.Name = parts[0]
				} else {
					service.Name = "Unknown Container"
				}

				// Calculate uptime for running containers
				if c.State == "running" {
					created := time.Unix(c.Created, 0)
					uptime := time.Since(created)
					service.Uptime = formatUptime(uptime)
				} else {
					service.Uptime = "Not Running"
				}

				return service
			}
		}
	}

	return nil
}

// isPortListening checks if a port is listening using a TCP connection attempt
func (sm *ServiceManager) isPortListening(port int) bool {
	timeout := 100 * time.Millisecond
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// killProcess kills a specific process by PID
func (sm *ServiceManager) killProcess(pid string) error {
	if pid == "" {
		return fmt.Errorf("empty PID provided")
	}

	// Validate PID is numeric
	if _, err := strconv.Atoi(pid); err != nil {
		return fmt.Errorf("invalid PID format: %s", pid)
	}

	cmd := exec.Command("kill", "-9", pid)
	return cmd.Run()
}

// imagesMatch checks if two Docker images match (handles tag variations)
func (sm *ServiceManager) imagesMatch(actualImage, expectedImage string) bool {
	// Simple exact match
	if actualImage == expectedImage {
		return true
	}

	// Remove tags for comparison (e.g., "nginx:latest" -> "nginx")
	actualBase := strings.Split(actualImage, ":")[0]
	expectedBase := strings.Split(expectedImage, ":")[0]

	return actualBase == expectedBase
}

// identifyUnexpectedService tries to identify what an unexpected service might be
func (sm *ServiceManager) identifyUnexpectedService(service ServiceInfo) ServiceInfo {
	// Check if this image is expected on a different port
	allServices := autoport.GetServiceNames()
	for _, serviceName := range allServices {
		if expectedService, found := autoport.GetServiceByName(serviceName); found {
			if sm.imagesMatch(service.Image, expectedService.Image) {
				service.Name = fmt.Sprintf("%s (Wrong Port - Expected: %d)",
					expectedService.Name, expectedService.ExternalPort)
				service.ExpectedImage = expectedService.Image
				service.ImageMatches = true
				break
			}
		}
	}

	return service
}

// formatUptime formats a duration into a human-readable string
func formatUptime(d time.Duration) string {
	if d > 24*time.Hour {
		return fmt.Sprintf("%.1f days", d.Hours()/24)
	} else if d > time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	} else if d > time.Minute {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	} else {
		return fmt.Sprintf("%.0f seconds", d.Seconds())
	}
}

// Autoport Generation Methods (Docker Compose integration)

// DockerComposeService represents a service in docker-compose.yml
type DockerComposeService struct {
	Image       string                 `yaml:"image"`
	Ports       []string               `yaml:"ports"`
	Environment []string               `yaml:"environment"`
	DependsOn   interface{}            `yaml:"depends_on"`
	Networks    interface{}            `yaml:"networks"`
	Healthcheck map[string]interface{} `yaml:"healthcheck"`
}

// DockerCompose represents the structure of docker-compose.yml
type DockerCompose struct {
	Version  string                          `yaml:"version"`
	Services map[string]DockerComposeService `yaml:"services"`
	Networks map[string]interface{}          `yaml:"networks"`
}

// AutoPortConfig represents the configuration for autoport generation
type AutoPortConfig struct {
	Name         string
	Image        string
	ExternalPort int
	InternalPort int
	Protocol     string
	HealthPath   string
	IsSecure     bool
	IPAddress    string
	Aliases      []string
	Environment  []string
	DependsOn    []string
}

// GenerateAutoPortConfig reads docker-compose.yml and generates autoport configuration
func (sm *ServiceManager) GenerateAutoPortConfig(composeFilePath, outputPath string) error {
	// Read and parse docker-compose.yml
	yamlFile, err := os.Open(composeFilePath)
	if err != nil {
		return fmt.Errorf("failed to open docker-compose.yml: %w", err)
	}
	defer yamlFile.Close()

	yamlData, err := io.ReadAll(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read docker-compose.yml: %w", err)
	}

	var compose DockerCompose
	err = yaml.Unmarshal(yamlData, &compose)
	if err != nil {
		return fmt.Errorf("failed to parse docker-compose.yml: %w", err)
	}

	// Parse services and extract port configurations
	configs := make(map[string]AutoPortConfig)
	portMappings := make(map[int]string)

	for serviceName, service := range compose.Services {
		config := AutoPortConfig{
			Name:        serviceName,
			Image:       service.Image,
			Environment: service.Environment,
		}

		// Parse port mappings
		for _, portMapping := range service.Ports {
			external, internal, protocol := parsePortMapping(portMapping)
			if external > 0 && internal > 0 {
				config.ExternalPort = external
				config.InternalPort = internal
				config.Protocol = protocol
				portMappings[external] = serviceName
				break // Take first port mapping
			}
		}

		// Determine if service is secure (HTTPS)
		config.IsSecure = (config.InternalPort == 443)

		// Generate health path
		config.HealthPath = generateHealthPath(serviceName, config.IsSecure)

		// Parse dependencies
		config.DependsOn = parseDependsOn(service.DependsOn)

		// Parse network configuration for IP and aliases
		config.IPAddress, config.Aliases = parseNetworkConfig(service.Networks)

		configs[serviceName] = config
	}

	// Generate Go file
	return sm.generateAutoPortGoFile(configs, portMappings, outputPath)
}

// parsePortMapping parses a Docker port mapping string like "8080:80" or "8080:80/tcp"
func parsePortMapping(portStr string) (external, internal int, protocol string) {
	protocol = "tcp" // default

	// Handle protocol suffix
	if strings.Contains(portStr, "/") {
		parts := strings.Split(portStr, "/")
		portStr = parts[0]
		if len(parts) > 1 {
			protocol = parts[1]
		}
	}

	// Split external:internal
	parts := strings.Split(portStr, ":")
	if len(parts) != 2 {
		return 0, 0, protocol
	}

	external, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, protocol
	}

	internal, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, protocol
	}

	return external, internal, protocol
}

// generateHealthPath generates a health check path for a service
func generateHealthPath(serviceName string, isSecure bool) string {
	scheme := "http"
	if isSecure {
		scheme = "https"
	}

	// Special cases for known services
	switch serviceName {
	case "ca", "amt-backend", "amt-frontend", "portdash":
		return fmt.Sprintf("%s://localhost/health", scheme)
	case "firebase", "gcs", "gmail", "secrets", "gcr", "openai":
		return fmt.Sprintf("%s://localhost/health", scheme)
	case "metadata":
		return fmt.Sprintf("%s://localhost/", scheme)
	default:
		return fmt.Sprintf("%s://localhost/health", scheme)
	}
}

// parseDependsOn extracts dependency list from depends_on configuration
func parseDependsOn(dependsOn interface{}) []string {
	var deps []string

	switch v := dependsOn.(type) {
	case []interface{}:
		for _, dep := range v {
			if depStr, ok := dep.(string); ok {
				deps = append(deps, depStr)
			}
		}
	case map[string]interface{}:
		for dep := range v {
			deps = append(deps, dep)
		}
	}

	return deps
}

// parseNetworkConfig extracts IP address and aliases from network configuration
func parseNetworkConfig(networks interface{}) (string, []string) {
	var ipAddress string
	var aliases []string

	switch v := networks.(type) {
	case map[string]interface{}:
		for _, networkConfig := range v {
			if networkMap, ok := networkConfig.(map[string]interface{}); ok {
				if ip, exists := networkMap["ipv4_address"]; exists {
					if ipStr, ok := ip.(string); ok {
						ipAddress = ipStr
					}
				}
				if aliasesInt, exists := networkMap["aliases"]; exists {
					if aliasesList, ok := aliasesInt.([]interface{}); ok {
						for _, alias := range aliasesList {
							if aliasStr, ok := alias.(string); ok {
								aliases = append(aliases, aliasStr)
							}
						}
					}
				}
			}
		}
	}

	return ipAddress, aliases
}

// generateAutoPortGoFile generates the Go source file for autoport package
func (sm *ServiceManager) generateAutoPortGoFile(configs map[string]AutoPortConfig, portMappings map[int]string, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate Go file content
	tmplStr := `// Package autoport provides auto-generated port configurations from Docker Compose
// This file is generated by servicemanager - do not edit manually
// Generated: {{.Generated}}
// Source: docker-compose.yml
package autoport

import "time"

// ServiceConfig represents a service configuration with port mappings
type ServiceConfig struct {
	Name         string    ` + "`json:\"name\"`" + `
	Image        string    ` + "`json:\"image\"`" + `
	ExternalPort int       ` + "`json:\"external_port\"`" + `
	InternalPort int       ` + "`json:\"internal_port\"`" + `
	Protocol     string    ` + "`json:\"protocol\"`" + `
	HealthPath   string    ` + "`json:\"health_path,omitempty\"`" + `
	IsSecure     bool      ` + "`json:\"is_secure\"`" + `
	IPAddress    string    ` + "`json:\"ip_address,omitempty\"`" + `
	Aliases      []string  ` + "`json:\"aliases,omitempty\"`" + `
	Environment  []string  ` + "`json:\"environment,omitempty\"`" + `
	DependsOn    []string  ` + "`json:\"depends_on,omitempty\"`" + `
}

// PortMapping represents a simple port mapping
type PortMapping struct {
	External int    ` + "`json:\"external\"`" + `
	Internal int    ` + "`json:\"internal\"`" + `
	Service  string ` + "`json:\"service\"`" + `
}

// Configuration holds the complete auto-generated port configuration
type Configuration struct {
	Version          string                   ` + "`json:\"version\"`" + `
	Generated        time.Time                ` + "`json:\"generated\"`" + `
	Source           string                   ` + "`json:\"source\"`" + `
	DockerSocketPath string                   ` + "`json:\"docker_socket_path,omitempty\"`" + `
	DockerAvailable  bool                     ` + "`json:\"docker_available\"`" + `
	Services         map[string]ServiceConfig ` + "`json:\"services\"`" + `
	PortMappings     map[int]PortMapping      ` + "`json:\"port_mappings\"`" + `
}

// GetConfiguration returns the auto-generated configuration
func GetConfiguration() *Configuration {
	return &defaultConfig
}

// GetServiceByPort returns the service configuration for a given external port
func GetServiceByPort(port int) (ServiceConfig, bool) {
	if mapping, exists := defaultConfig.PortMappings[port]; exists {
		if service, exists := defaultConfig.Services[mapping.Service]; exists {
			return service, true
		}
	}
	return ServiceConfig{}, false
}

// GetServiceByName returns the service configuration by service name
func GetServiceByName(name string) (ServiceConfig, bool) {
	service, exists := defaultConfig.Services[name]
	return service, exists
}

// GetAllPorts returns all external ports in use
func GetAllPorts() []int {
	ports := make([]int, 0, len(defaultConfig.PortMappings))
	for port := range defaultConfig.PortMappings {
		ports = append(ports, port)
	}
	return ports
}

// GetServiceNames returns all service names
func GetServiceNames() []string {
	names := make([]string, 0, len(defaultConfig.Services))
	for name := range defaultConfig.Services {
		names = append(names, name)
	}
	return names
}

// Auto-generated configuration data
var defaultConfig = Configuration{
	Version:          Version,
	Generated:        time.Date({{.Year}}, {{.Month}}, {{.Day}}, {{.Hour}}, {{.Minute}}, {{.Second}}, 0, time.UTC),
	Source:           "docker-compose.yml",
	DockerSocketPath: "{{.DockerSocketPath}}",
	DockerAvailable:  {{.DockerAvailable}},
	Services: map[string]ServiceConfig{
{{range $name, $config := .Services}}		"{{$name}}": {
			Name:         "{{$config.Name}}",
			Image:        "{{$config.Image}}",
			ExternalPort: {{$config.ExternalPort}},
			InternalPort: {{$config.InternalPort}},
			Protocol:     "{{$config.Protocol}}",
			HealthPath:   "{{$config.HealthPath}}",
			IsSecure:     {{$config.IsSecure}},
			IPAddress:    "{{$config.IPAddress}}",
			Aliases:      {{printf "%#v" $config.Aliases}},
			Environment:  {{printf "%#v" $config.Environment}},
			DependsOn:    {{printf "%#v" $config.DependsOn}},
		},
{{end}}	},
	PortMappings: map[int]PortMapping{
{{range $port, $service := .PortMappings}}		{{$port}}: {
			External: {{$port}},
			Internal: {{(index $.Services $service).InternalPort}},
			Service:  "{{$service}}",
		},
{{end}}	},
}
`

	tmpl, err := template.New("autoport").Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	now := time.Now()
	dockerSocketPath := ""
	dockerAvailable := false
	if sm.dockerConfig != nil {
		dockerSocketPath = sm.dockerConfig.SocketPath
		dockerAvailable = sm.dockerConfig.Available
	}

	data := struct {
		Generated                              string
		Year, Month, Day, Hour, Minute, Second int
		DockerSocketPath                       string
		DockerAvailable                        bool
		Services                               map[string]AutoPortConfig
		PortMappings                           map[int]string
	}{
		Generated:        now.Format("2006-01-02 15:04:05 UTC"),
		Year:             now.Year(),
		Month:            int(now.Month()),
		Day:              now.Day(),
		Hour:             now.Hour(),
		Minute:           now.Minute(),
		Second:           now.Second(),
		DockerSocketPath: dockerSocketPath,
		DockerAvailable:  dockerAvailable,
		Services:         configs,
		PortMappings:     portMappings,
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Execute template
	err = tmpl.Execute(outFile, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
