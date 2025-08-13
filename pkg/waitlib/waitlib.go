// SPDX-License-Identifier: CC0-1.0

// Package waitlib provides a simple wait utility that can be used as a command-line tool
// or imported as a library. It displays version information and uptime while running.
package waitlib

import (
	"flag"
	"fmt"
	"time"
)

// Version is the current version of the waitlib package
const Version = "v0.1.0"

// WaitConfig holds configuration for the wait functionality
type WaitConfig struct {
	Version     string
	ShowHelp    bool
	ShowVersion bool
}

// Run executes the waitlib functionality with the given version string.
// This is the main entry point that handles command-line arguments and starts the wait process.
//
// Example usage:
//
//	waitlib.Run("v1.0.0")
func Run(version string) {
	config := parseFlags()

	if config.ShowHelp {
		showHelp()
		return
	}

	if config.ShowVersion {
		showVersion(version)
		return
	}

	// Start the wait process
	startWait(version)
}

// parseFlags parses command-line flags and returns a WaitConfig
func parseFlags() WaitConfig {
	config := WaitConfig{}

	flag.BoolVar(&config.ShowHelp, "help", false, "Show help information")
	flag.BoolVar(&config.ShowVersion, "version", false, "Show version information")
	flag.Parse()

	return config
}

// showHelp displays help information
func showHelp() {
	fmt.Printf(`waitlib - A simple wait utility

Usage:
  waitlib [options]

Options:
  --help     Show this help message
  --version  Show version information

Description:
  waitlib is a utility that runs indefinitely, updating its process name
  to show the current version and uptime. This is useful for container
  health checks and process monitoring.

  When running, the process will appear in 'docker ps' as:
    wait <version> <uptime>

Examples:
  waitlib --help
  waitlib --version
  waitlib

`)
}

// showVersion displays version information
func showVersion(version string) {
	fmt.Printf("waitlib %s (library version: %s)\n", version, Version)
}

// startWait begins the main wait loop, updating the process title periodically
func startWait(version string) {
	startTime := time.Now()

	// Update process title immediately
	updateProcessTitle(version, startTime)

	// Create a ticker to update the process title every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	fmt.Printf("waitlib %s started - waiting indefinitely...\n", version)
	fmt.Printf("Process will show as: wait %s <uptime>\n", version)

	// Main wait loop
	for range ticker.C {
		updateProcessTitle(version, startTime)
	}
}

// updateProcessTitle updates the process title to show version and uptime
func updateProcessTitle(version string, startTime time.Time) {
	uptime := time.Since(startTime)

	// Format uptime as human-readable string
	uptimeStr := formatUptime(uptime)

	// Update process title using OS-specific method
	newTitle := fmt.Sprintf("wait %s %s", version, uptimeStr)

	// Try to update process title (platform-specific implementation)
	if err := setProcessTitle(newTitle); err != nil {
		// If setting process title fails, we'll just continue silently
		// This is common on some systems where process title changes aren't supported
	}
}

// formatUptime formats a duration as a human-readable uptime string
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}
