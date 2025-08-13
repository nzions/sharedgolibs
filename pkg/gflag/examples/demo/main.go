// SPDX-License-Identifier: CC0-1.0

// Package main provides a simple demonstration of gflag capabilities.
package main

import (
	"fmt"

	"github.com/nzions/sharedgolibs/pkg/gflag"
)

func main() {
	// Define flags demonstrating various types and formats
	var (
		verbose = gflag.BoolP("verbose", "v", false, "enable verbose output")
		debug   = gflag.BoolP("debug", "d", false, "enable debug mode")
		quiet   = gflag.BoolP("quiet", "q", false, "suppress output")
		port    = gflag.IntP("port", "p", 8080, "server port")
		host    = gflag.StringP("host", "h", "localhost", "server hostname")
		config  = gflag.String("config", "/etc/app.conf", "config file path")
		workers = gflag.Int("workers", 4, "number of worker threads")
		help    = gflag.BoolP("help", "?", false, "show help message")
	)

	// Parse command line
	gflag.Parse()

	// Show help if requested
	if *help {
		fmt.Printf("gflag-demo - Demonstration of gflag capabilities\n\n")
		fmt.Printf("USAGE:\n")
		fmt.Printf("    gflag-demo [OPTIONS] [arguments...]\n\n")
		fmt.Printf("OPTIONS:\n")
		gflag.CommandLine.PrintDefaults()
		fmt.Printf("\nEXAMPLES:\n")
		fmt.Printf("    gflag-demo -v -p 9000 -h example.com\n")
		fmt.Printf("    gflag-demo --verbose --port=9000 --host=example.com\n")
		fmt.Printf("    gflag-demo -vdq --port=9000 arg1 arg2\n")
		fmt.Printf("    gflag-demo -v --config=/custom/path --workers=8\n")
		return
	}

	// Display parsed configuration
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Verbose: %t\n", *verbose)
	fmt.Printf("  Debug: %t\n", *debug)
	fmt.Printf("  Quiet: %t\n", *quiet)
	fmt.Printf("  Port: %d\n", *port)
	fmt.Printf("  Host: %s\n", *host)
	fmt.Printf("  Config: %s\n", *config)
	fmt.Printf("  Workers: %d\n", *workers)

	// Show arguments
	args := gflag.Args()
	if len(args) > 0 {
		fmt.Printf("  Arguments: %v\n", args)
	}

	// Demonstrate some conditional logic based on flags
	if *verbose && !*quiet {
		fmt.Printf("\nStarting application in verbose mode...\n")
		fmt.Printf("Server will bind to %s:%d\n", *host, *port)
		fmt.Printf("Using config file: %s\n", *config)
		fmt.Printf("Spawning %d worker threads\n", *workers)
	}

	if *debug {
		fmt.Printf("\nDEBUG: Command line arguments processed\n")
		fmt.Printf("DEBUG: Total non-flag arguments: %d\n", gflag.NArg())
		for i := 0; i < gflag.NArg(); i++ {
			fmt.Printf("DEBUG: Arg[%d] = %q\n", i, gflag.Arg(i))
		}
	}

	if !*quiet {
		fmt.Printf("\nApplication configured successfully!\n")
		if len(args) > 0 {
			fmt.Printf("Processing %d arguments...\n", len(args))
		}
	}
}
