package main

import (
	"fmt"

	"github.com/nzions/sharedgolibs/pkg/autoport"
)

func main() {
	fmt.Println("Testing autoport package:")
	fmt.Println("========================")

	config := autoport.GetConfiguration()
	fmt.Printf("Version: %s\n", config.Version)
	fmt.Printf("Generated: %s\n", config.Generated.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("Source: %s\n", config.Source)
	fmt.Printf("Total Services: %d\n", len(config.Services))
	fmt.Printf("Total Port Mappings: %d\n", len(config.PortMappings))

	fmt.Println("\nAll ports:")
	ports := autoport.GetAllPorts()
	for _, port := range ports {
		if service, found := autoport.GetServiceByPort(port); found {
			fmt.Printf("  %d -> %s (%s:%d)\n", port, service.Name, service.Image, service.InternalPort)
		}
	}

	fmt.Println("\nService details:")
	serviceNames := autoport.GetServiceNames()
	for _, name := range serviceNames {
		if service, found := autoport.GetServiceByName(name); found {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    External Port: %d\n", service.ExternalPort)
			fmt.Printf("    Internal Port: %d\n", service.InternalPort)
			fmt.Printf("    Health Path: %s\n", service.HealthPath)
			fmt.Printf("    Is Secure: %t\n", service.IsSecure)
			if len(service.Aliases) > 0 {
				fmt.Printf("    Aliases: %v\n", service.Aliases)
			}
			if len(service.DependsOn) > 0 {
				fmt.Printf("    Dependencies: %v\n", service.DependsOn)
			}
			fmt.Println()
		}
	}

	// Test specific lookups
	fmt.Println("Testing specific lookups:")
	if service, found := autoport.GetServiceByPort(8080); found {
		fmt.Printf("Port 8080: %s (%s)\n", service.Name, service.Image)
	}

	if service, found := autoport.GetServiceByName("ca"); found {
		fmt.Printf("CA Service: Port %d -> %d (%s)\n", service.ExternalPort, service.InternalPort, service.HealthPath)
	}
}
