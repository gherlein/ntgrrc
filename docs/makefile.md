# Makefile Documentation for ntgrrc

## Overview

The Makefile provides a comprehensive build system for ntgrrc with targets for building, testing, cross-compilation, and development workflows. It automates common tasks and ensures consistent builds across different environments.

## Quick Start

```bash
# Show all available targets
make help

# Build for current platform
make build

# Run all tests
make test

# Run with race detection and coverage
make ci
```

## Core Targets

### Building

#### `make build`
Builds the binary for the current platform with optimizations and version information.
- Output: `build/ntgrrc`
- Includes version, build time, and git commit in binary
- Uses trimpath for reproducible builds

#### `make build-dev`
Builds a development version with race detection enabled.
- Output: `build/ntgrrc-dev`
- Includes race detection
- No optimizations for easier debugging

#### `make install`
Installs the binary to `$GOPATH/bin`.
- Builds with full optimizations
- Makes ntgrrc available system-wide

### Running

#### `make run ARGS="command args"`
Builds and runs ntgrrc with specified arguments.
```bash
# Show help
make run ARGS="--help"

# Login to switch
make run ARGS="login --address=192.168.1.1 --password=secret"

# Get POE status
make run ARGS="poe status --address=192.168.1.1"
```

#### `make run-dev ARGS="command args"`
Runs in development mode with race detection.
```bash
make run-dev ARGS="--help"
```

## Testing Targets

### Basic Testing

#### `make test`
Runs all tests with verbose output.
- Includes unit tests and integration tests
- Shows detailed test results

#### `make test-short`
Runs tests with the `-short` flag.
- Skips long-running tests
- Useful for quick validation

#### `make test-race`
Runs tests with race condition detection.
- Critical for concurrent code validation
- Required for CI/CD pipelines

### Specialized Testing

#### `make test-unit`
Runs only unit tests, excluding integration tests.
```bash
# Tests individual functions and components
make test-unit
```

#### `make test-integration`
Runs only integration tests.
```bash
# Tests complete workflows
make test-integration
```

#### `make test-models`
Runs model-specific tests.
```bash
# Tests GS305EP, GS316EP functionality
make test-models
```

### Coverage Analysis

#### `make test-coverage`
Generates coverage profile.
- Output: `coverage/coverage.out`
- Required for other coverage targets

#### `make coverage-html`
Creates HTML coverage report.
- Output: `coverage/coverage.html`
- Open in browser for visual coverage analysis

#### `make coverage-func`
Shows function-level coverage in terminal.
```bash
make coverage-func
# Output:
# main.go:32:    detectNetgearModel        85.7%
# main.go:65:    main                      100.0%
# total:         (statements)              87.5%
```

#### `make coverage-check`
Validates coverage meets threshold (80%).
- Fails build if coverage is below threshold
- Used in CI/CD pipelines

## Cross-Platform Building

#### `make cross-build`
Builds for all supported platforms.
- Platforms: Linux, macOS, Windows, FreeBSD
- Architectures: amd64, arm64, 386
- Output: `dist/ntgrrc-{os}-{arch}[.exe]`

Supported platforms:
- `linux/amd64`, `linux/arm64`, `linux/386`
- `darwin/amd64`, `darwin/arm64`
- `windows/amd64`, `windows/386`
- `freebsd/amd64`

#### `make release`
Creates release archives for all platforms.
- `.tar.gz` for Unix-like systems
- `.zip` for Windows
- Includes README.md and LICENSE

## Development Targets

### Code Quality

#### `make fmt`
Formats Go source code using `go fmt`.

#### `make vet`
Runs `go vet` to find potential issues.

#### `make lint`
Runs basic linting (fmt + vet).

#### `make check`
Comprehensive quality check (lint + test-race + coverage-check).

### Dependencies

#### `make deps`
Downloads and verifies Go module dependencies.

#### `make tidy`
Cleans up go.mod and go.sum files.

### Development Environment

#### `make dev`
Sets up complete development environment.
- Downloads dependencies
- Builds development version
- Ready for development

#### `make debug`
Builds with debug symbols for use with Delve debugger.
- Output: `build/ntgrrc-debug`
- Disables optimizations
- Enables debugging symbols

## CI/CD Targets

#### `make ci`
Complete CI pipeline.
1. Clean build artifacts
2. Download dependencies
3. Run linting
4. Run tests with race detection
5. Verify coverage threshold

Perfect for automated testing environments.

## Utility Targets

#### `make clean`
Removes all build artifacts and temporary files.
- Cleans `build/`, `dist/`, `coverage/` directories
- Removes generated binaries

#### `make version`
Shows version information.
```bash
make version
# Output:
# ntgrrc version information:
#   Version: v1.2.3
#   Build Time: 2024-01-15_14:30:22
#   Git Commit: abc1234
#   Go Version: go version go1.23.0 linux/amd64
```

#### `make size`
Shows binary sizes for current and cross-compiled builds.

#### `make help`
Shows all available targets with descriptions.

## Docker Support

#### `make docker-build`
Builds Docker image.
- Tags: `ntgrrc:version` and `ntgrrc:latest`

#### `make docker-run ARGS="command args"`
Runs ntgrrc in Docker container.

## Advanced Features

### Benchmarking

#### `make bench`
Runs benchmark tests.
```bash
make bench
# Shows performance metrics for critical functions
```

### Profiling

#### `make profile`
Builds with profiling support enabled.
- Use with Go profiling tools
- Enables performance analysis

### File Watching

#### `make watch`
Watches for file changes and runs tests automatically.
- Requires `entr` command (`apt install entr` or `brew install entr`)
- Useful during development

### Test Data Validation

#### `make mock-data`
Validates test data files are present.
- Checks `test-data/` directory structure
- Ensures HTML fixtures exist

## Environment Variables

The Makefile supports several environment variables:

- `VERSION`: Override version string (default: git describe)
- `ARGS`: Arguments to pass to `make run` or `make run-dev`
- `COVERAGE_THRESHOLD`: Coverage percentage threshold (default: 80)

## Usage Examples

### Development Workflow
```bash
# Set up development environment
make dev

# Make changes to code...

# Run tests with coverage
make check

# Test specific functionality
make run-dev ARGS="poe status --address=192.168.1.1"

# Build for release
make build
```

### CI/CD Pipeline
```bash
# Complete CI check
make ci

# Cross-compile for release
make cross-build

# Create release packages
make release
```

### Testing Workflow
```bash
# Quick test during development
make test-short

# Full test suite
make test-race

# Check coverage
make coverage-html
# Open coverage/coverage.html in browser

# Integration tests only
make test-integration
```

### Release Workflow
```bash
# Clean previous builds
make clean

# Run full quality checks
make ci

# Cross-compile for all platforms
make cross-build

# Create release archives
make release

# Verify binary sizes
make size
```

## Customization

### Local Customizations
Create `Makefile.local` for project-specific customizations:
```makefile
# Makefile.local
CUSTOM_FLAGS := -tags debug
BUILD_FLAGS += $(CUSTOM_FLAGS)

custom-target:
	@echo "Custom target"
```

### Build Flags
The Makefile uses several build flags:
- `-ldflags`: Injects version information
- `-trimpath`: Removes file system paths for reproducible builds
- `-race`: Enables race detection in development builds

## Dependencies

Required tools:
- Go 1.23+ (specified in go.mod)
- Git (for version information)
- Make
- `bc` (for coverage threshold checking)

Optional tools:
- Docker (for containerized builds)
- `entr` (for file watching)
- `dlv` (for debugging)

The Makefile handles missing optional tools gracefully and provides clear error messages.