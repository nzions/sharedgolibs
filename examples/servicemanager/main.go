package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nzions/sharedgolibs/pkg/servicemanager"
)

func main() {
	fmt.Println("=== Service Manager Demo ===")
	fmt.Println()

	// Example 1: Basic service discovery
	fmt.Println("1. Basic Service Discovery")
	basicDiscovery()
	fmt.Println()

	// Example 2: Custom configuration
	fmt.Println("2. Custom Configuration")
	customConfiguration()
	fmt.Println()

	// Example 3: Docker-specific operations
	fmt.Println("3. Docker Integration")
	dockerIntegration()
	fmt.Println()

	// Example 4: Service categorization
	fmt.Println("4. Service Categorization")
	serviceCategorization()
	fmt.Println()

	// Example 5: Port monitoring (legacy processmanager style)
	fmt.Println("5. Port Monitoring")
	portMonitoring()
	fmt.Println()

	// Example 6: Service status analysis
	fmt.Println("6. Service Status Analysis")
	serviceStatusAnalysis()
}

func basicDiscovery() {
	// Create a new service manager with default configuration
	sm := servicemanager.New()

	// Discover all services
	services, err := sm.DiscoverAllServices()
	if err != nil {
		log.Printf("Error discovering services: %v", err)
		return
	}

	fmt.Printf("Found %d services:\n", len(services))
	for _, service := range services {
		fmt.Printf("  - %s on port %d (%s)\n", service.Name, service.ExternalPort, service.Type)
	}
}

func customConfiguration() {
	// Create service manager with custom options
	sm := servicemanager.New(
		servicemanager.WithPortRange(3000, 4000),
		servicemanager.WithKnownService(3000, "My API", "http://localhost:3000/health", false),
		servicemanager.WithMonitoredPort(3001, "Frontend Dev Server"),
		servicemanager.WithDockerTimeout(10*time.Second),
	)

	portRange := sm.GetPortRange()
	fmt.Printf("Custom port range: %d-%d\n", portRange.Start, portRange.End)

	monitoredPorts := sm.GetMonitoredPorts()
	fmt.Printf("Monitored ports: %v\n", monitoredPorts)

	// Check a specific port
	if service, err := sm.CheckPort(3000); err == nil {
		fmt.Printf("Service on port 3000: %s\n", service.Name)
	} else {
		fmt.Printf("No service on port 3000: %v\n", err)
	}
}

func dockerIntegration() {
	sm := servicemanager.New()

	// Check Docker availability
	if sm.IsDockerAvailable() {
		fmt.Printf("Docker available at: %s\n", sm.GetDockerSocketPath())

		// Get Docker-specific services
		dockerServices, err := sm.DiscoverDockerServices()
		if err != nil {
			log.Printf("Error discovering Docker services: %v", err)
			return
		}

		fmt.Printf("Docker services: %d\n", len(dockerServices))
		for _, service := range dockerServices {
			fmt.Printf("  - %s (%s) - %s\n", service.Name, service.Image, service.Status)
			if service.Uptime != "" {
				fmt.Printf("    Uptime: %s\n", service.Uptime)
			}
		}
	} else {
		fmt.Println("Docker not available")

		// Still can discover local services
		localServices := sm.DiscoverLocalServices()
		fmt.Printf("Local services: %d\n", len(localServices))
		for _, service := range localServices {
			fmt.Printf("  - %s (PID: %s)\n", service.Name, service.PID)
		}
	}
}

func serviceCategorization() {
	sm := servicemanager.New()

	// Get expected services (from autoport configuration)
	expectedServices, err := sm.DiscoverExpectedServices()
	if err != nil {
		log.Printf("Error discovering expected services: %v", err)
		return
	}
	fmt.Printf("Expected services running: %d\n", len(expectedServices))

	// Get unexpected services
	unexpectedServices, err := sm.DiscoverUnexpectedServices()
	if err != nil {
		log.Printf("Error discovering unexpected services: %v", err)
		return
	}
	fmt.Printf("Unexpected services running: %d\n", len(unexpectedServices))

	// Get missing services
	missingServices := sm.GetMissingServices()
	fmt.Printf("Missing expected services: %d\n", len(missingServices))

	if len(missingServices) > 0 {
		fmt.Println("Missing services:")
		for _, service := range missingServices {
			fmt.Printf("  - %s on port %d\n", service.Name, service.ExternalPort)
		}
	}
}

func portMonitoring() {
	// Create a simple service manager for monitoring specific ports
	sm := servicemanager.NewSimple(
		servicemanager.WithMonitoredPort(8080, "Backend API"),
		servicemanager.WithMonitoredPort(8081, "Frontend"),
		servicemanager.WithMonitoredPort(8082, "Database"),
	)

	// Check monitored ports (legacy processmanager compatibility)
	status, err := sm.CheckMonitoredPorts()
	if err != nil {
		log.Printf("Error checking monitored ports: %v", err)
		return
	}

	fmt.Printf("Monitored ports: %d total, %d listening\n", status.Total, status.Listening)
	for _, service := range status.Running {
		listening := "not listening"
		if service.IsListening {
			listening = "listening"
		}
		fmt.Printf("  Port %d (%s): %s\n", service.ExternalPort, service.Description, listening)
	}
}

func serviceStatusAnalysis() {
	sm := servicemanager.New()

	// Get comprehensive service status
	status, err := sm.GetServiceStatus()
	if err != nil {
		log.Printf("Error getting service status: %v", err)
		return
	}

	fmt.Printf("Service Status Summary:\n")
	fmt.Printf("  Total services: %d\n", status.Total)
	fmt.Printf("  Listening: %d\n", status.Listening)
	fmt.Printf("  Expected: %d\n", status.Expected)
	fmt.Printf("  Unexpected: %d\n", status.Unexpected)
	fmt.Printf("  Missing: %d\n", len(status.Missing))

	if status.Expected > 0 {
		fmt.Printf("  Docker image matches: %d\n", status.ImageMatch)
		fmt.Printf("  Docker image mismatches: %d\n", status.ImageMismatch)
	}

	// Show detailed info for interesting services
	fmt.Println("\nDetailed Service Information:")
	for _, service := range status.Running {
		if service.IsExpected {
			expectedStatus := "✓"
			if service.Type == servicemanager.ServiceTypeDockerContainer && !service.ImageMatches {
				expectedStatus = "⚠ (image mismatch)"
			}
			fmt.Printf("  %s %s on port %d (%s)\n", expectedStatus, service.Name, service.ExternalPort, service.Type)
		}
	}
}

// Example: Service Management Operations
func serviceManagementExample() {
	sm := servicemanager.New()

	// Example: Kill a specific service
	// err := sm.KillServiceOnPort(8080)
	// if err != nil {
	//     log.Printf("Failed to kill service on port 8080: %v", err)
	// }

	// Example: Kill all monitored services
	// errors := sm.KillAllServices()
	// if len(errors) > 0 {
	//     fmt.Println("Errors occurred while killing services:")
	//     for _, err := range errors {
	//         fmt.Printf("  - %v\n", err)
	//     }
	// }

	// Example: Add/remove monitored ports dynamically
	sm.AddMonitoredPort(9999, "Test Service")
	fmt.Printf("Added port 9999: %s\n", sm.GetPortDescription(9999))

	sm.RemoveMonitoredPort(9999)
	fmt.Printf("Removed port 9999: %s\n", sm.GetPortDescription(9999))
}

// Example: Autoport Configuration Generation
func autoPortGenerationExample() {
	// sm := servicemanager.New()

	// Generate autoport configuration from docker-compose.yml
	// This would read docker-compose.yml and generate Go code
	// err := sm.GenerateAutoPortConfig("docker-compose.yml", "pkg/autoport/autoport.go")
	// if err != nil {
	//     log.Printf("Failed to generate autoport config: %v", err)
	//     return
	// }
	// fmt.Println("Autoport configuration generated successfully")

	fmt.Println("Autoport generation example (commented out - requires docker-compose.yml)")
}
