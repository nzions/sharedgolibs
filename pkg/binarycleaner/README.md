# Binary Cleaner Package

The `binarycleaner` package provides functionality to find and remove Mach-O and ELF binary files from a directory tree. This is useful for cleaning up build artifacts and compiled binaries from development environments.

## Features

- **Binary Detection**: Automatically detects Mach-O and ELF binaries by analyzing file headers
- **Recursive Search**: Optionally search subdirectories recursively  
- **Dry Run Mode**: Preview what would be removed without actually deleting files
- **Verbose Output**: Detailed logging of operations
- **Safe Operation**: Only removes files with executable permissions and valid binary headers

## Supported Binary Formats

### Mach-O (macOS)
- 32-bit Mach-O binaries
- 64-bit Mach-O binaries  
- Universal/Fat Mach-O binaries

### ELF (Linux/Unix)
- All ELF binary formats (32-bit and 64-bit)

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/nzions/sharedgolibs/pkg/binarycleaner"
)

func main() {
    config := binarycleaner.Config{
        Directory: "/path/to/search",
        DryRun:    true,  // Preview mode
        Verbose:   true,  // Detailed output
        Recursive: true,  // Search subdirectories
    }
    
    cleaner := binarycleaner.New(config)
    
    err := cleaner.Clean()
    if err != nil {
        log.Fatal(err)
    }
}
```

### Finding Binaries Only

```go
cleaner := binarycleaner.New(config)

binaries, err := cleaner.FindBinaries()
if err != nil {
    log.Fatal(err)
}

for _, binary := range binaries {
    fmt.Printf("Found: %s (%s, %d bytes)\n", 
        binary.Path, binary.Type, binary.Size)
}
```

### Manual Removal

```go
cleaner := binarycleaner.New(config)

binaries, err := cleaner.FindBinaries()
if err != nil {
    log.Fatal(err)
}

// Filter or modify the list as needed
filteredBinaries := []binarycleaner.BinaryInfo{}
for _, binary := range binaries {
    if binary.Size > 1000000 { // Only large binaries
        filteredBinaries = append(filteredBinaries, binary)
    }
}

err = cleaner.RemoveBinaries(filteredBinaries)
if err != nil {
    log.Fatal(err)
}
```

## Configuration Options

### Config Struct

```go
type Config struct {
    Directory string  // Directory to search (required)
    DryRun    bool    // If true, only preview operations
    Verbose   bool    // Enable detailed output
    Recursive bool    // Search subdirectories recursively
}
```

### Directory
The root directory to search for binaries. Must be an absolute path.

### DryRun
When enabled, the tool will only show what would be removed without actually deleting any files. Useful for testing and verification.

### Verbose
Enables detailed output including:
- Individual file operations
- Warnings for inaccessible files
- Progress information

### Recursive
When enabled, searches all subdirectories recursively. When disabled, only searches the specified directory.

## Binary Detection Algorithm

The package uses a multi-step process to identify binaries:

1. **File Extension Filter**: Skips common text files (.txt, .md, .go, etc.)
2. **Executable Check**: Only examines files with executable permissions
3. **Magic Number Detection**: Reads file headers to identify binary formats:
   - ELF: `0x7f 'E' 'L' 'F'`
   - Mach-O 32-bit: `0xfeedface` (LE) or `0xcefaedfe` (BE)
   - Mach-O 64-bit: `0xfeedfacf` (LE) or `0xcffaedfe` (BE)
   - Universal Mach-O: `0xcafebabe` (BE) or `0xbebafeca` (LE)

## Error Handling

The package provides robust error handling:

- **File Access Errors**: Warns about inaccessible files but continues operation
- **Directory Walk Errors**: Gracefully handles permission issues
- **Removal Errors**: Reports specific files that couldn't be removed

## Safety Features

- **Extension Whitelist**: Automatically skips common text file extensions
- **Permission Check**: Only examines executable files
- **Header Validation**: Verifies binary format before removal
- **Dry Run Mode**: Always test before actual removal

## Performance Considerations

- Skips non-executable files for better performance
- Uses efficient file header reading (only first 16 bytes)
- Provides progress feedback for large directory trees

## Examples

### Clean Current Directory (Dry Run)
```go
config := binarycleaner.Config{
    Directory: ".",
    DryRun:    true,
    Verbose:   true,
    Recursive: false,
}
```

### Recursively Clean Build Directory
```go
config := binarycleaner.Config{
    Directory: "/path/to/project/build",
    DryRun:    false,
    Verbose:   true,
    Recursive: true,
}
```

### Silent Operation
```go
config := binarycleaner.Config{
    Directory: "/tmp",
    DryRun:    false,
    Verbose:   false,
    Recursive: true,
}
```

## Version

Current version: `v0.1.0`

## Thread Safety

The package is not thread-safe. Each `BinaryCleaner` instance should be used by a single goroutine at a time.
