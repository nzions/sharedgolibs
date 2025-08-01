package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nzions/sharedgolibs/pkg/testicle"
)

const (
	version = "v1.0.0"
)

type Config struct {
	Debug        bool
	Daemon       bool
	Dir          string
	ConfigFile   string
	Version      bool
	NoVet        bool
	NoBuildCheck bool
	Validate     bool
}

func main() {
	config := parseFlags()

	if config.Version {
		fmt.Printf("testicle %s\n", version)
		os.Exit(0)
	}

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Received interrupt signal, shutting down...")
		cancel()
	}()

	// Initialize testicle runner
	runner, err := testicle.NewRunner(&testicle.Config{
		Debug:        config.Debug,
		Daemon:       config.Daemon,
		Dir:          config.Dir,
		ConfigFile:   config.ConfigFile,
		NoVet:        config.NoVet,
		NoBuildCheck: config.NoBuildCheck,
		Validate:     config.Validate,
	})
	if err != nil {
		log.Fatalf("Failed to initialize testicle: %v", err)
	}

	// Run testicle
	if err := runner.Run(ctx); err != nil {
		if err == context.Canceled {
			fmt.Println("✅ Testicle stopped gracefully")
			os.Exit(0)
		}
		log.Fatalf("❌ Testicle failed: %v", err)
	}
}

func parseFlags() *Config {
	config := &Config{}

	flag.BoolVar(&config.Debug, "debug", false, "Enable debug output for troubleshooting")
	flag.BoolVar(&config.Daemon, "daemon", false, "Watch mode - auto-run tests on file changes")
	flag.BoolVar(&config.Daemon, "d", false, "Watch mode - auto-run tests on file changes (short)")
	flag.StringVar(&config.Dir, "dir", getDefaultTestDir(), "Test directory")
	flag.StringVar(&config.ConfigFile, "config", "testicle.yaml", "Configuration file location")
	flag.BoolVar(&config.Version, "version", false, "Show version information")

	// Validation flags
	flag.BoolVar(&config.NoVet, "no-vet", false, "Skip go vet validation")
	flag.BoolVar(&config.NoBuildCheck, "no-build-check", false, "Skip test compilation validation")
	flag.BoolVar(&config.Validate, "validate", false, "Run validation only (no test execution)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "🧪 Testicle %s - A Playwright-inspired test runner for Go\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: testicle [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Core Flags:\n")
		fmt.Fprintf(os.Stderr, "  --debug         Enable debug output for troubleshooting\n")
		fmt.Fprintf(os.Stderr, "  --daemon, -d    Watch mode - auto-run tests on file changes\n")
		fmt.Fprintf(os.Stderr, "  --dir <path>    Test directory (default: %s)\n", getDefaultTestDir())
		fmt.Fprintf(os.Stderr, "  --config <file> Configuration file location (default: testicle.yaml)\n")
		fmt.Fprintf(os.Stderr, "  --version       Show version information\n\n")
		fmt.Fprintf(os.Stderr, "Validation Flags:\n")
		fmt.Fprintf(os.Stderr, "  --validate      Run validation only (no test execution)\n")
		fmt.Fprintf(os.Stderr, "  --no-vet        Skip go vet validation\n")
		fmt.Fprintf(os.Stderr, "  --no-build-check Skip test compilation validation\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  testicle                           # Run tests once with validation\n")
		fmt.Fprintf(os.Stderr, "  testicle --daemon                  # Watch mode\n")
		fmt.Fprintf(os.Stderr, "  testicle --validate                # Run validation only\n")
		fmt.Fprintf(os.Stderr, "  testicle --no-vet --no-build-check # Skip all validation\n")
		fmt.Fprintf(os.Stderr, "  testicle --debug --dir ./my-tests  # Debug mode with custom directory\n")
		fmt.Fprintf(os.Stderr, "  testicle --config custom.yaml      # Use custom configuration\n\n")
		fmt.Fprintf(os.Stderr, "For complete documentation, see: https://github.com/nzions/sharedgolibs/tree/master/pkg/testicle/doc\n")
	}

	flag.Parse()
	return config
}

// getDefaultTestDir returns the appropriate default test directory
// based on whether we're running in a container or locally
func getDefaultTestDir() string {
	// Check if we're in a container environment
	if isContainer() {
		return "/tests"
	}
	// Local development - use current directory
	return "."
}

// isContainer detects if we're running inside a container
func isContainer() bool {
	// Check for container-specific environment variables
	if os.Getenv("CONTAINER") == "true" {
		return true
	}

	// Check for Docker-specific files
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check if we're running as PID 1 (common in containers)
	if os.Getpid() == 1 {
		return true
	}

	return false
}
