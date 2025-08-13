# waitlib

A simple Go library that provides a wait utility with version and uptime display functionality.

## Features

- Command-line argument parsing (`--help`, `--version`)
- Process title updating to show version and uptime
- Human-readable uptime formatting
- Docker-friendly process naming for container monitoring

## Usage

### As a Library

```go
package main

import "github.com/nzions/sharedgolibs/pkg/waitlib"

func main() {
    waitlib.Run("v1.2.3")
}
```

### Command Line

```bash
# Show help
./myapp --help

# Show version
./myapp --version

# Run and wait indefinitely
./myapp
```

### Docker Container Usage

When running in a Docker container, the process will appear in `docker ps` output as:
```
wait v1.2.3 2d4h15m
```

This makes it easy to monitor container versions and uptime at a glance.

## API Reference

### Functions

#### `Run(version string)`
Main entry point that handles command-line arguments and starts the wait process.

#### `formatUptime(d time.Duration) string`
Formats a duration as a human-readable uptime string (e.g., "2d4h15m").

### Constants

#### `Version`
Current version of the waitlib package.

## Process Title Updates

The library attempts to update the process title using platform-specific methods:

**Linux:**
- Uses `prctl(PR_SET_NAME)` syscall for the comm field (15 char limit)
- Falls back to writing `/proc/self/comm` 
- Also attempts argv[0] manipulation for longer titles

**macOS:**
- Uses argv[0] manipulation for ps output

**Other Platforms:**
- No-op implementation (returns success but doesn't change title)

The process title follows the format: `wait <version> <uptime>`

### Platform-Specific Implementation

The code uses Go build constraints to provide platform-specific implementations:
- `waitlib_linux.go` - Linux-specific process title setting
- `waitlib_darwin.go` - macOS-specific process title setting  
- `waitlib_other.go` - Fallback for unsupported platforms

This ensures optimal functionality on each platform while maintaining portability.

### Technical Details
- **No External Dependencies**: Pure Go implementation using syscalls
- **Memory Safe**: Respects original argv bounds to prevent crashes
- **Graceful Degradation**: Continues running even if title updates fail
- **Container Friendly**: Designed for minimal/scratch container environments

## Examples

See the `examples/waitlib/` directory for a complete working example.

## Testing

Run tests with:
```bash
go test ./pkg/waitlib/
```

Run benchmarks with:
```bash
go test -bench=. ./pkg/waitlib/
```

## Version History

- **v0.1.0**: Initial release with basic wait functionality and process title updates
