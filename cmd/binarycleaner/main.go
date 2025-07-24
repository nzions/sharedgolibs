package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nzions/sharedgolibs/pkg/binarycleaner"
)

const version = "1.0.0"

func main() {
	var (
		directory   = flag.String("dir", ".", "Directory to search for binaries")
		dryRun      = flag.Bool("dry-run", false, "Show what would be removed without actually removing files")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		recursive   = flag.Bool("recursive", false, "Search subdirectories recursively")
		help        = flag.Bool("help", false, "Show help information")
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

	// Convert directory to absolute path
	absDir, err := filepath.Abs(*directory)
	if err != nil {
		log.Fatalf("Error resolving directory path: %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	config := binarycleaner.Config{
		Directory: absDir,
		DryRun:    *dryRun,
		Verbose:   *verbose,
		Recursive: *recursive,
	}

	cleaner := binarycleaner.New(config)

	if *dryRun {
		fmt.Println("=== DRY RUN MODE - No files will be removed ===")
	}

	if *verbose {
		fmt.Printf("Searching directory: %s\n", absDir)
		fmt.Printf("Recursive: %t\n", *recursive)
		fmt.Printf("Dry run: %t\n", *dryRun)
		fmt.Println()
	}

	err = cleaner.Clean()
	if err != nil {
		log.Fatalf("Error during cleanup: %v", err)
	}

	if *dryRun {
		fmt.Println("\n=== DRY RUN COMPLETE - Run without --dry-run to actually remove files ===")
	}
}

func showHelp() {
	fmt.Printf("Binary Cleaner v%s\n\n", version)
	fmt.Println("A tool to find and remove Mach-O and ELF binary files from directories.")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Printf("  %s [OPTIONS]\n\n", os.Args[0])
	fmt.Println("OPTIONS:")
	fmt.Println("  -dir string")
	fmt.Println("        Directory to search for binaries (default \".\")")
	fmt.Println("  -dry-run")
	fmt.Println("        Show what would be removed without actually removing files")
	fmt.Println("  -recursive")
	fmt.Println("        Search subdirectories recursively")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -version")
	fmt.Println("        Show version information")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Preview what would be removed from current directory")
	fmt.Printf("  %s --dry-run --verbose\n\n", os.Args[0])
	fmt.Println("  # Recursively clean build directory")
	fmt.Printf("  %s --dir ./build --recursive\n\n", os.Args[0])
	fmt.Println("  # Clean specific directory with verbose output")
	fmt.Printf("  %s --dir /tmp --verbose\n\n", os.Args[0])
	fmt.Println("SUPPORTED BINARY FORMATS:")
	fmt.Println("  • Mach-O binaries (macOS): 32-bit, 64-bit, Universal/Fat")
	fmt.Println("  • ELF binaries (Linux/Unix): All variants")
	fmt.Println()
	fmt.Println("SAFETY FEATURES:")
	fmt.Println("  • Only removes files with executable permissions")
	fmt.Println("  • Validates binary headers before removal")
	fmt.Println("  • Skips common text file extensions")
	fmt.Println("  • Dry-run mode for safe testing")
	fmt.Println()
}

func showVersion() {
	fmt.Printf("Binary Cleaner v%s\n", version)
	fmt.Printf("Package Version: %s\n", binarycleaner.Version)
	fmt.Println("Built with Go")
}
