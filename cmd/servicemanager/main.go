package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nzions/sharedgolibs/pkg/servicemanager"
)

const version = "3.0.0"

func main() {
	var (
		kill        = flag.Bool("k", false, "Kill services listening on monitored ports")
		killPort    = flag.Int("kill-port", 0, "Kill service on specific port")
		check       = flag.Bool("check", false, "Check status of all services")
		port        = flag.Int("port", 0, "Check specific port")
		expected    = flag.Bool("expected", false, "Show only expected services")
		unexpected  = flag.Bool("unexpected", false, "Show only unexpected services")
		docker      = flag.Bool("docker", false, "Show only Docker services")
		local       = flag.Bool("local", false, "Show only local process services")
		missing     = flag.Bool("missing", false, "Show missing expected services")
		status      = flag.Bool("status", false, "Show comprehensive service status")
		jsonOutput  = flag.Bool("json", false, "Output in JSON format")
		portRange   = flag.String("range", "", "Port range to scan (e.g., '3000-4000')")
		generate    = flag.String("generate", "", "Generate autoport config from docker-compose.yml")
		help        = flag.Bool("help", false, "Show help")
		versionFlag = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *versionFlag {
		showVersion()
		return
	}

	// Create service manager with custom port range if specified
	var sm *servicemanager.ServiceManager
	if *portRange != "" {
		start, end, err := parsePortRange(*portRange)
		if err != nil {
			log.Fatalf("Invalid port range: %v", err)
		}
		sm = servicemanager.New(servicemanager.WithPortRange(start, end))
	} else {
		sm = servicemanager.New()
	}

	// Handle autoport generation
	if *generate != "" {
		err := sm.GenerateAutoPortConfig(*generate, "pkg/autoport/autoport.go")
		if err != nil {
			log.Fatalf("Failed to generate autoport config: %v", err)
		}
		fmt.Println("Autoport configuration generated successfully")
		return
	}

	// Handle specific port checking
	if *port > 0 {
		checkSpecificPort(sm, *port, *jsonOutput)
		return
	}

	// Handle specific port killing
	if *killPort > 0 {
		killSpecificPort(sm, *killPort)
		return
	}

	// Handle service discovery with filters
	if *expected || *unexpected || *docker || *local {
		showFilteredServices(sm, *expected, *unexpected, *docker, *local, *jsonOutput)
		return
	}

	// Handle missing services
	if *missing {
		showMissingServices(sm, *jsonOutput)
		return
	}

	// Handle comprehensive status
	if *status {
		showServiceStatus(sm, *jsonOutput)
		return
	}

	// Handle check (default behavior)
	if *check || (!*kill && flag.NFlag() == 0) {
		showAllServices(sm, *jsonOutput)
		return
	}

	// Handle kill services
	if *kill {
		killAllServices(sm)
		return
	}
}

func showHelp() {
	fmt.Printf("servicemanager version %s\n", version)
	fmt.Println("A comprehensive tool for managing development services and ports")
	fmt.Println()
	fmt.Println("Usage: servicemanager [options]")
	fmt.Println()
	fmt.Println("Service Discovery:")
	fmt.Println("  -check          Check status of all services (default)")
	fmt.Println("  -port=N         Check specific port N")
	fmt.Println("  -expected       Show only expected services")
	fmt.Println("  -unexpected     Show only unexpected services")
	fmt.Println("  -docker         Show only Docker container services")
	fmt.Println("  -local          Show only local process services")
	fmt.Println("  -missing        Show missing expected services")
	fmt.Println("  -status         Show comprehensive service status")
	fmt.Println()
	fmt.Println("Service Control:")
	fmt.Println("  -k              Kill services listening on monitored ports")
	fmt.Println("  -kill-port=N    Kill service on specific port N")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  -range=START-END Port range to scan (e.g., '3000-4000')")
	fmt.Println("  -generate=FILE   Generate autoport config from docker-compose.yml")
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  -json           Output in JSON format")
	fmt.Println("  -version        Show version information")
	fmt.Println("  -help           Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  servicemanager                    # Check all services")
	fmt.Println("  servicemanager -port=8080         # Check port 8080")
	fmt.Println("  servicemanager -expected -json    # Expected services as JSON")
	fmt.Println("  servicemanager -k                 # Kill all monitored services")
	fmt.Println("  servicemanager -missing           # Show missing services")
	fmt.Println("  servicemanager -range=3000-4000   # Scan ports 3000-4000")
	fmt.Println("  servicemanager -generate=docker-compose.yml  # Generate autoport config")
}

func showVersion() {
	fmt.Printf("servicemanager version %s\n", version)
	fmt.Printf("servicemanager package version %s\n", servicemanager.Version)
	fmt.Println()

	// Create service manager to show Docker info
	sm := servicemanager.New()

	fmt.Println("Environment:")
	if sm.IsDockerAvailable() {
		fmt.Printf("  Docker: Available (%s)\n", sm.GetDockerSocketPath())
	} else {
		fmt.Println("  Docker: Not Available")
	}

	portRange := sm.GetPortRange()
	fmt.Printf("  Default Port Range: %d-%d\n", portRange.Start, portRange.End)

	monitoredPorts := sm.GetMonitoredPorts()
	fmt.Printf("  Monitored Ports: %d\n", len(monitoredPorts))
}

func parsePortRange(rangeStr string) (int, int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format, expected 'start-end'")
	}

	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start port: %v", err)
	}

	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end port: %v", err)
	}

	if start > end {
		return 0, 0, fmt.Errorf("start port cannot be greater than end port")
	}

	return start, end, nil
}

func checkSpecificPort(sm *servicemanager.ServiceManager, port int, jsonOutput bool) {
	service, err := sm.CheckPort(port)
	if err != nil {
		if jsonOutput {
			result := map[string]interface{}{
				"port":      port,
				"error":     err.Error(),
				"listening": false,
			}
			json.NewEncoder(os.Stdout).Encode(result)
		} else {
			fmt.Printf("Port %d: %v\n", port, err)
		}
		return
	}

	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(service)
	} else {
		printServiceInfo(*service)
	}
}

func killSpecificPort(sm *servicemanager.ServiceManager, port int) {
	fmt.Printf("Killing service on port %d...\n", port)
	err := sm.KillServiceOnPort(port)
	if err != nil {
		fmt.Printf("Failed to kill service on port %d: %v\n", port, err)
		os.Exit(1)
	}
	fmt.Printf("Service on port %d killed successfully\n", port)
}

func showFilteredServices(sm *servicemanager.ServiceManager, expected, unexpected, docker, local bool, jsonOutput bool) {
	var services []servicemanager.ServiceInfo
	var err error

	if expected {
		services, err = sm.DiscoverExpectedServices()
	} else if unexpected {
		services, err = sm.DiscoverUnexpectedServices()
	} else if docker {
		services, err = sm.DiscoverDockerServices()
	} else if local {
		services = sm.DiscoverLocalServices()
	}

	if err != nil {
		log.Fatalf("Failed to discover services: %v", err)
	}

	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(services)
	} else {
		if len(services) == 0 {
			fmt.Println("No services found matching the filter criteria")
			return
		}

		var filterType string
		if expected {
			filterType = "Expected"
		} else if unexpected {
			filterType = "Unexpected"
		} else if docker {
			filterType = "Docker"
		} else if local {
			filterType = "Local"
		}

		fmt.Printf("%s Services (%d found):\n", filterType, len(services))
		for _, service := range services {
			printServiceInfo(service)
		}
	}
}

func showMissingServices(sm *servicemanager.ServiceManager, jsonOutput bool) {
	missing := sm.GetMissingServices()

	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(missing)
	} else {
		if len(missing) == 0 {
			fmt.Println("All expected services are running")
			return
		}

		fmt.Printf("Missing Services (%d):\n", len(missing))
		for _, service := range missing {
			fmt.Printf("  Port %d: %s\n", service.ExternalPort, service.Name)
			if service.Image != "" {
				fmt.Printf("    Image: %s\n", service.Image)
			}
			if service.HealthPath != "" {
				fmt.Printf("    Health: %s\n", service.HealthPath)
			}
		}
	}
}

func showServiceStatus(sm *servicemanager.ServiceManager, jsonOutput bool) {
	status, err := sm.GetServiceStatus()
	if err != nil {
		log.Fatalf("Failed to get service status: %v", err)
	}

	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(status)
	} else {
		fmt.Printf("Service Status Summary:\n")
		fmt.Printf("  Total Services: %d\n", status.Total)
		fmt.Printf("  Listening: %d\n", status.Listening)
		fmt.Printf("  Expected: %d\n", status.Expected)
		fmt.Printf("  Unexpected: %d\n", status.Unexpected)
		fmt.Printf("  Missing: %d\n", len(status.Missing))

		if status.Expected > 0 {
			fmt.Printf("  Image Matches: %d\n", status.ImageMatch)
			fmt.Printf("  Image Mismatches: %d\n", status.ImageMismatch)
		}

		fmt.Println()

		if len(status.Running) > 0 {
			fmt.Printf("Running Services:\n")
			for _, service := range status.Running {
				printServiceInfo(service)
			}
		}

		if len(status.Missing) > 0 {
			fmt.Printf("\nMissing Services:\n")
			for _, service := range status.Missing {
				fmt.Printf("  Port %d: %s (%s)\n", service.ExternalPort, service.Name, service.Image)
			}
		}
	}
}

func showAllServices(sm *servicemanager.ServiceManager, jsonOutput bool) {
	services, err := sm.DiscoverAllServices()
	if err != nil {
		log.Fatalf("Failed to discover services: %v", err)
	}

	if jsonOutput {
		json.NewEncoder(os.Stdout).Encode(services)
	} else {
		fmt.Printf("Service Manager %s\n", version)

		if sm.IsDockerAvailable() {
			fmt.Printf("Docker: Available (%s)\n", sm.GetDockerSocketPath())
		} else {
			fmt.Println("Docker: Not Available")
		}

		portRange := sm.GetPortRange()
		fmt.Printf("Port Range: %d-%d\n", portRange.Start, portRange.End)
		fmt.Println()

		if len(services) == 0 {
			fmt.Println("No services found")
			return
		}

		fmt.Printf("Discovered Services (%d):\n", len(services))
		for _, service := range services {
			printServiceInfo(service)
		}
	}
}

func killAllServices(sm *servicemanager.ServiceManager) {
	fmt.Println("Killing all monitored services...")

	errors := sm.KillAllServices()
	if len(errors) > 0 {
		fmt.Printf("Errors occurred while killing services:\n")
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		os.Exit(1)
	} else {
		fmt.Println("All services killed successfully")
	}
}

func printServiceInfo(service servicemanager.ServiceInfo) {
	status := "●"
	if !service.IsListening {
		status = "○"
	}

	expected := ""
	if service.IsExpected {
		expected = " [EXPECTED]"
		if service.Type == servicemanager.ServiceTypeDockerContainer && !service.ImageMatches {
			expected = " [EXPECTED - IMAGE MISMATCH]"
		}
	} else {
		expected = " [UNEXPECTED]"
	}

	fmt.Printf("  %s Port %d: %s (%s)%s\n", status, service.ExternalPort, service.Name, service.Type, expected)

	if service.Type == servicemanager.ServiceTypeDockerContainer {
		if service.ContainerID != "" {
			fmt.Printf("    Container: %s\n", service.ContainerID)
		}
		if service.Image != "" {
			fmt.Printf("    Image: %s\n", service.Image)
		}
		if service.ExpectedImage != "" && service.Image != service.ExpectedImage {
			fmt.Printf("    Expected Image: %s\n", service.ExpectedImage)
		}
		if service.Uptime != "" {
			fmt.Printf("    Uptime: %s\n", service.Uptime)
		}
	} else if service.Type == servicemanager.ServiceTypeLocalProcess {
		if service.PID != "" {
			fmt.Printf("    PID: %s\n", service.PID)
		}
		if service.Command != "" {
			fmt.Printf("    Command: %s\n", service.Command)
		}
	}

	if service.Description != "" && service.Description != service.Name {
		fmt.Printf("    Description: %s\n", service.Description)
	}

	if service.HealthURL != "" {
		fmt.Printf("    Health: %s\n", service.HealthURL)
	}

	fmt.Printf("    Status: %s\n", service.Status)
}
