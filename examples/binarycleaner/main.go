package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nzions/sharedgolibs/pkg/binarycleaner"
)

func main() {
	// Example 1: Basic usage with dry run
	fmt.Println("=== Example 1: Basic Dry Run ===")
	basicExample()

	fmt.Println("\n=== Example 2: Finding Binaries Only ===")
	findOnlyExample()

	fmt.Println("\n=== Example 3: Filtering by Size ===")
	filterBySizeExample()
}

func basicExample() {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	config := binarycleaner.Config{
		Directory: currentDir,
		DryRun:    true, // Safe mode - won't actually remove files
		Verbose:   true,
		Recursive: false, // Only current directory
	}

	cleaner := binarycleaner.New(config)
	err = cleaner.Clean()
	if err != nil {
		log.Printf("Error: %v", err)
	}
}

func findOnlyExample() {
	// Create a temporary directory with some test files
	tmpDir, err := os.MkdirTemp("", "binarycleaner_example")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake ELF binary
	elfPath := filepath.Join(tmpDir, "test_elf")
	elfContent := []byte{0x7f, 'E', 'L', 'F', 0x02, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	err = os.WriteFile(elfPath, elfContent, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Create a fake Mach-O binary
	machoPath := filepath.Join(tmpDir, "test_macho")
	machoContent := []byte{0xfe, 0xed, 0xfa, 0xce, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	err = os.WriteFile(machoPath, machoContent, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Create a text file (should be ignored)
	textPath := filepath.Join(tmpDir, "readme.txt")
	err = os.WriteFile(textPath, []byte("This is a text file"), 0644)
	if err != nil {
		log.Fatal(err)
	}

	config := binarycleaner.Config{
		Directory: tmpDir,
		DryRun:    true,
		Verbose:   false,
		Recursive: false,
	}

	cleaner := binarycleaner.New(config)

	// Find binaries without removing them
	binaries, err := cleaner.FindBinaries()
	if err != nil {
		log.Printf("Error finding binaries: %v", err)
		return
	}

	fmt.Printf("Found %d binaries in %s:\n", len(binaries), tmpDir)
	for _, binary := range binaries {
		fmt.Printf("  - %s (%s, %d bytes)\n",
			filepath.Base(binary.Path), binary.Type, binary.Size)
	}
}

func filterBySizeExample() {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	config := binarycleaner.Config{
		Directory: currentDir,
		DryRun:    true,
		Verbose:   false,
		Recursive: true,
	}

	cleaner := binarycleaner.New(config)

	// Find all binaries first
	binaries, err := cleaner.FindBinaries()
	if err != nil {
		log.Printf("Error finding binaries: %v", err)
		return
	}

	// Filter by size - only show binaries larger than 1MB
	largeBinaries := []binarycleaner.BinaryInfo{}
	for _, binary := range binaries {
		if binary.Size > 1024*1024 { // 1MB
			largeBinaries = append(largeBinaries, binary)
		}
	}

	fmt.Printf("Found %d large binaries (>1MB):\n", len(largeBinaries))
	var totalSize int64
	for _, binary := range largeBinaries {
		fmt.Printf("  - %s (%s, %.2f MB)\n",
			filepath.Base(binary.Path),
			binary.Type,
			float64(binary.Size)/(1024*1024))
		totalSize += binary.Size
	}

	if len(largeBinaries) > 0 {
		fmt.Printf("Total size of large binaries: %.2f MB\n", float64(totalSize)/(1024*1024))

		// Demonstrate manual removal (dry run)
		err = cleaner.RemoveBinaries(largeBinaries)
		if err != nil {
			log.Printf("Error during removal: %v", err)
		}
	}
}
