package binarycleaner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const Version = "0.1.0"

// BinaryType represents the type of binary file
type BinaryType int

const (
	Unknown BinaryType = iota
	MachO
	ELF
)

func (bt BinaryType) String() string {
	switch bt {
	case MachO:
		return "Mach-O"
	case ELF:
		return "ELF"
	default:
		return "Unknown"
	}
}

// BinaryInfo contains information about a detected binary
type BinaryInfo struct {
	Path string
	Type BinaryType
	Size int64
}

// Config holds configuration for the binary cleaner
type Config struct {
	Directory string
	DryRun    bool
	Verbose   bool
	Recursive bool
}

// BinaryCleaner handles finding and removing binary files
type BinaryCleaner struct {
	config Config
}

// New creates a new BinaryCleaner with the given configuration
func New(config Config) *BinaryCleaner {
	return &BinaryCleaner{config: config}
}

// detectBinaryType checks if a file is a Mach-O or ELF binary
func detectBinaryType(filePath string) (BinaryType, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return Unknown, err
	}
	defer file.Close()

	// Read the first 16 bytes to check magic numbers
	header := make([]byte, 16)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return Unknown, err
	}
	if n < 4 {
		return Unknown, nil
	}

	// Check for ELF magic number (0x7f, 'E', 'L', 'F')
	if n >= 4 && header[0] == 0x7f && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
		return ELF, nil
	}

	// Check for Mach-O magic numbers
	if n >= 4 {
		// 32-bit Mach-O: 0xfeedface (little endian) or 0xcefaedfe (big endian)
		if (header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xce) ||
			(header[0] == 0xce && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) {
			return MachO, nil
		}
		// 64-bit Mach-O: 0xfeedfacf (little endian) or 0xcffaedfe (big endian)
		if (header[0] == 0xfe && header[1] == 0xed && header[2] == 0xfa && header[3] == 0xcf) ||
			(header[0] == 0xcf && header[1] == 0xfa && header[2] == 0xed && header[3] == 0xfe) {
			return MachO, nil
		}
		// Universal/Fat Mach-O: 0xcafebabe (big endian) or 0xbebafeca (little endian)
		if (header[0] == 0xca && header[1] == 0xfe && header[2] == 0xba && header[3] == 0xbe) ||
			(header[0] == 0xbe && header[1] == 0xba && header[2] == 0xfe && header[3] == 0xca) {
			return MachO, nil
		}
	}

	return Unknown, nil
}

// isExecutable checks if a file has executable permissions
func isExecutable(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// FindBinaries searches for Mach-O and ELF binaries in the configured directory
func (bc *BinaryCleaner) FindBinaries() ([]BinaryInfo, error) {
	var binaries []BinaryInfo

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if bc.config.Verbose {
				fmt.Printf("Warning: Error accessing %s: %v\n", path, err)
			}
			return nil // Continue walking despite errors
		}

		// Skip directories
		if info.IsDir() {
			// If not recursive and this is a subdirectory, skip it
			if !bc.config.Recursive && path != bc.config.Directory {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files that are obviously not binaries (common extensions)
		ext := strings.ToLower(filepath.Ext(path))
		skipExtensions := []string{".txt", ".md", ".go", ".py", ".js", ".html", ".css", ".json", ".yml", ".yaml", ".xml", ".log"}
		for _, skipExt := range skipExtensions {
			if ext == skipExt {
				return nil
			}
		}

		// Check if file is executable (optimization)
		if !isExecutable(path) {
			return nil
		}

		// Detect binary type
		binaryType, err := detectBinaryType(path)
		if err != nil {
			if bc.config.Verbose {
				fmt.Printf("Warning: Error reading %s: %v\n", path, err)
			}
			return nil
		}

		if binaryType != Unknown {
			binaries = append(binaries, BinaryInfo{
				Path: path,
				Type: binaryType,
				Size: info.Size(),
			})
		}

		return nil
	}

	err := filepath.Walk(bc.config.Directory, walkFunc)
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return binaries, nil
}

// RemoveBinaries removes the specified binary files
func (bc *BinaryCleaner) RemoveBinaries(binaries []BinaryInfo) error {
	for _, binary := range binaries {
		if bc.config.DryRun {
			fmt.Printf("Would remove: %s (%s, %d bytes)\n", binary.Path, binary.Type, binary.Size)
		} else {
			if bc.config.Verbose {
				fmt.Printf("Removing: %s (%s, %d bytes)\n", binary.Path, binary.Type, binary.Size)
			}

			err := os.Remove(binary.Path)
			if err != nil {
				return fmt.Errorf("failed to remove %s: %w", binary.Path, err)
			}
		}
	}
	return nil
}

// Clean finds and removes all Mach-O and ELF binaries in the configured directory
func (bc *BinaryCleaner) Clean() error {
	binaries, err := bc.FindBinaries()
	if err != nil {
		return err
	}

	if len(binaries) == 0 {
		if bc.config.Verbose {
			fmt.Println("No binaries found.")
		}
		return nil
	}

	fmt.Printf("Found %d binarie(s):\n", len(binaries))

	var totalSize int64
	for _, binary := range binaries {
		fmt.Printf("  %s (%s, %d bytes)\n", binary.Path, binary.Type, binary.Size)
		totalSize += binary.Size
	}

	fmt.Printf("Total size: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/(1024*1024))

	return bc.RemoveBinaries(binaries)
}
