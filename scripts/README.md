# Scripts Directory

This directory contains utility scripts for the ntgrrc project.

## Available Scripts

### `ci.sh` - Comprehensive CI Test Suite

A complete continuous integration script that runs all quality checks, tests, and validations.

#### Usage

```bash
# Run the full CI suite
./scripts/ci.sh

# Set custom coverage threshold
COVERAGE_THRESHOLD=85 ./scripts/ci.sh

# Run in CI environment (suppresses interactive features)
CI=true ./scripts/ci.sh
```

#### What it does

The CI script performs the following steps in order:

1. **Environment Check**
   - Validates required tools (Go, Git, Make, bc)
   - Checks Go version compatibility
   - Verifies project structure

2. **Dependency Management**
   - Cleans previous builds
   - Downloads and verifies dependencies
   - Checks for known vulnerabilities (if govulncheck available)

3. **Code Quality**
   - Checks code formatting
   - Runs `go vet`
   - Runs static analysis (if staticcheck available)
   - Checks for ineffectual assignments

4. **Build Tests**
   - Builds for current platform
   - Tests basic functionality
   - Builds development version with race detection

5. **Unit Tests**
   - Runs all unit tests
   - Excludes integration tests for speed

6. **Integration Tests**
   - Runs end-to-end workflow tests
   - Tests complete user scenarios

7. **Race Detection**
   - Runs all tests with race condition detection
   - Critical for concurrent code validation

8. **Coverage Analysis**
   - Generates coverage reports
   - Validates coverage threshold (default: 80%)
   - Shows low-coverage functions

9. **Cross-Compilation**
   - Builds for all supported platforms
   - Verifies all binaries are created
   - Shows binary sizes

10. **Security Checks**
    - Scans for hardcoded secrets
    - Checks file permissions
    - Basic security validation

11. **Performance Tests**
    - Runs benchmark tests (if available)
    - Performance validation

12. **Documentation Check**
    - Verifies required documentation exists
    - Checks documentation currency

13. **Final Validation**
    - Smoke tests of built binaries
    - Git status validation
    - Final readiness check

#### Configuration

Environment variables:
- `COVERAGE_THRESHOLD`: Minimum coverage percentage (default: 80)
- `CI`: Set to "true" to enable CI mode (suppresses interactive features)

#### Exit Codes

- `0`: All checks passed
- `1`: One or more checks failed

#### Output

The script provides colored, detailed output showing:
- Progress indicators
- Step-by-step results
- Error details when failures occur
- Final summary with execution time

#### Requirements

Required tools:
- Go 1.23+
- Git
- Make
- bc (for coverage calculations)

Optional tools (enhance functionality):
- govulncheck (vulnerability scanning)
- staticcheck (static analysis)
- ineffassign (ineffectual assignments check)

#### Example Output

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                           ntgrrc CI Test Suite                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[INFO] Starting CI pipeline at Mon Jan 15 14:30:22 UTC 2024
[INFO] Project: ntgrrc - Netgear Remote Control CLI
[INFO] Coverage threshold: 80%

Progress: [1/13] 8% - Environment Check
==> Environment Check
[INFO] Go version: 1.23.0 ✓
[SUCCESS] Environment check passed

Progress: [2/13] 15% - Dependency Management
==> Dependency Management
[INFO] Cleaning previous build artifacts...
[INFO] Downloading dependencies...
[SUCCESS] Dependencies ready

... (continues for all steps)

╔══════════════════════════════════════════════════════════════════════════════╗
║                             CI PIPELINE PASSED                              ║
╚══════════════════════════════════════════════════════════════════════════════╝

[SUCCESS] All CI checks completed successfully!
[SUCCESS] Total execution time: 2m 34s
[SUCCESS] Ready for deployment/release
```

## Adding New Scripts

When adding new scripts to this directory:

1. Make them executable: `chmod +x scripts/new-script.sh`
2. Add proper error handling and logging
3. Use consistent exit codes (0 for success, non-zero for failure)
4. Document the script purpose and usage in this README
5. Follow the existing code style and patterns