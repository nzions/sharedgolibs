# Testicle 🧪

> A Playwright-inspired test runner for Go that enhances the native `go test` experience with interactive controls, watch mode, and real-time feedback.

## Features

- 🧪 **Interactive Test Runner**: Real-time test execution with live progress updates
- 👀 **Watch Mode**: Automatic test re-runs on file changes with intelligent filtering
- 📋 **Pre-execution Validation**: `go vet` and `go test -c` compilation checks
- 📊 **Performance Tracking**: Statistical analysis with regression detection
- 🐳 **Container Ready**: Optimized for both local development and CI/CD environments
- ⌨️ **Interactive Terminal**: Keyboard controls for pause, resume, run, and stop operations

## Quick Start

### Local Development
```bash
# Run tests once
testicle

# Watch mode - auto-run on file changes  
testicle --daemon

# Debug mode with verbose output
testicle --debug

# Custom test directory
testicle --dir ./my-tests

# Use configuration file
testicle --config ./custom-testicle.yaml
```

### Container Usage  
```bash
docker run -v $(pwd):/tests testicle:latest --daemon
```

## Core CLI Flags

| Flag              | Short | Description                                                  |
| ----------------- | ----- | ------------------------------------------------------------ |
| `--debug`         |       | Enable debug output for troubleshooting                      |
| `--daemon`        | `-d`  | Watch mode - auto-run tests on file changes                  |
| `--dir <path>`    |       | Test directory (default: `/tests` in container, `.` locally) |
| `--config <file>` |       | Configuration file location                                  |

> **Complete CLI Reference**: See [CLI_REFERENCE.md](./CLI_REFERENCE.md) for all flags and options.

## Documentation

- 📋 **[CLI Reference](./CLI_REFERENCE.md)** - Complete command-line interface guide and configuration
- 🎮 **[Interactive Terminal](./INTERACTIVE_TERMINAL.md)** - Key bindings and interactive controls
- 🔍 **[Validation & Performance](./VALIDATION_AND_PERFORMANCE.md)** - Pre-execution validation and performance tracking
- 🏗️ **[Technical Specification](./TECHNICAL_SPEC.md)** - Core architecture and components
- 🐳 **[Container Guide](./CONTAINER_GUIDE.md)** - Docker deployment and CI/CD integration
- 📋 **[Implementation Plan](./IMPLEMENTATION_PLAN.md)** - Development roadmap and code examples
- 🛣️ **[Roadmap](./ROADMAP.md)** - Long-term feature roadmap
- 👨‍💻 **[Development Guide](./DEVELOPMENT.md)** - Local development and testing

## Example Output

```bash
🧪 Testicle v1.0 - Running tests in /tests

Validating tests... ████████████████████████████████ 100%
✅ go vet: clean (47 files checked)
✅ go test -c: all tests compile successfully (12 packages)

Discovering tests... ████████████████████████████████ 100%
Found 47 tests in 12 packages [estimated total: ~4m 30s]

Running tests... ████████████████████████████░░░░░ 89% (42/47) [~45s remaining]

✅ pkg/auth/TestPasswordValidation (89ms)  [faster than usual: 92ms ±8ms]
✅ pkg/user/TestProfileUpdate (156ms) [avg: 145ms ±12ms]
❌ pkg/api/TestRateLimiting (2.1s) [SLOW: expected ~150ms]
    --- FAIL: TestRateLimiting (2.10s)
        api_test.go:45: rate limit not enforced correctly

[r] Re-run failed • [a] Run all • [f] Run specific file • [q] Quit
```

## Development Status

**Status**: 🚧 In Development - CLI-focused implementation with core flags: `--debug`, `--daemon/-d`, `--dir`, `--config`

**Next Steps**: Implementation of test discovery and execution engine.