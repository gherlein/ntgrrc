# Switch Test Program

A comprehensive real-world test program for the ntgrrc library that exercises all major POE and port management functionality against actual Netgear switch hardware.

## Overview

This program performs automated testing of switch functionality including:

1. **POE Power Cycling**: Disable → Verify → Enable → Verify for each POE port
2. **Bandwidth Limitation**: Limit to 1Mbps → Verify → Restore → Verify for each ethernet port  
3. **LED Control**: Disable LEDs → Enable LEDs → Verify state restoration

The program includes comprehensive safety features including state backup/restore, dry-run mode, and graceful error handling.

## Features

- ✅ **Real-world validation** against actual hardware
- ✅ **State backup and restoration** to ensure no permanent changes
- ✅ **Dry-run mode** for safe preview of operations
- ✅ **Comprehensive error handling** with retry logic
- ✅ **JSON output** for CI/CD integration
- ✅ **Signal handling** for graceful interruption cleanup
- ✅ **Flexible test selection** with skip options

## Installation

Build the switch-test program:

```bash
make build-switch-test
```

Or build manually:

```bash
cd switch-test
go build -o switch-test .
```

## Usage

### Basic Usage

```bash
# Test all functionality
export NETGEAR_SWITCHES="switch1=password123"
./switch-test switch1

# Preview operations without executing
./switch-test --dry-run switch1

# Test with debug output  
./switch-test --debug switch1
```

### Selective Testing

```bash
# Skip POE tests (only test bandwidth and LEDs)
./switch-test --skip-poe switch1

# Skip bandwidth tests
./switch-test --skip-bandwidth switch1

# Skip LED tests
./switch-test --skip-leds switch1

# Test only POE functionality
./switch-test --skip-bandwidth --skip-leds switch1
```

### Configuration Options

```bash
# Adjust timing
./switch-test --delay 5 --timeout 60 switch1

# JSON output for automation
./switch-test --json switch1 > test-results.json

# Verbose output
./switch-test --verbose switch1
```

## Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--debug, -d` | Enable debug output | false |
| `--dry-run` | Preview operations without executing | false |
| `--skip-poe` | Skip POE power cycling tests | false |
| `--skip-bandwidth` | Skip bandwidth limitation tests | false |
| `--skip-leds` | Skip LED control tests | false |
| `--json` | Output results in JSON format | false |
| `--verbose` | Verbose output with detailed operations | false |
| `--delay <seconds>` | Delay between operations | 2 |
| `--timeout <seconds>` | Operation timeout | 30 |
| `--help, -h` | Show help message | - |

## Environment Variables

The program uses the same authentication system as other ntgrrc tools:

```bash
# Multi-switch configuration
export NETGEAR_SWITCHES="switch1=password123;switch2=password456"

# Host-specific password
export NETGEAR_PASSWORD_SWITCH1=password123

# Example with IP address
export NETGEAR_PASSWORD_192_168_1_10=password123
```

## Test Sequence

### 1. Initial Setup
- Connect and authenticate with switch
- Detect switch model and capabilities
- Capture initial state (POE status, port settings, LED status)
- Display switch configuration summary

### 2. POE Power Cycling Test
For each POE port:
- Disable POE power → Wait → Verify disabled
- Enable POE power → Wait → Verify enabled
- Report success/failure for each port

### 3. Bandwidth Limitation Test  
For each ethernet port:
- Record original bandwidth settings
- Set ingress/egress limits to 1 Mbps → Verify setting
- Restore original bandwidth → Verify restoration
- Report success/failure for each port

### 4. LED Control Test
- Disable all port LEDs → Verify state change
- Enable all port LEDs → Verify state restoration
- Report success/failure

### 5. Final Validation
- Compare current state with initial state
- Verify all settings were properly restored
- Generate comprehensive test report

## Output Examples

### Human-Readable Output

```
Switch Test Program - ntgrrc Library Validation
===============================================

Connecting to switch: 192.168.1.10
✓ Authentication successful (Model: GS308EPP)
✓ Detected 8 POE ports, 8 ethernet ports

POE Power Cycling Test:
  Port 1: ✓ Disabled → ✓ Verified Off → ✓ Enabled → ✓ Verified On
  Port 2: ✓ Disabled → ✓ Verified Off → ✓ Enabled → ✓ Verified On
  Port 3: ✗ Failed to disable (timeout) → ⚠ Skipped remaining steps
  ...

Bandwidth Limitation Test:
  Port 1: ✓ Limited to 1Mbps → ✓ Verified → ✓ Restored → ✓ Verified
  ...

LED Control Test:
  ✓ LEDs disabled → ✓ LEDs enabled → ✓ State restored

Final Validation:
  ✓ All settings restored to original state

Test Summary:
  Total Operations: 45
  Successful: 42
  Failed: 3
  Duration: 2m 15s

⚠ 3 operations failed out of 45 total
```

### JSON Output

```json
{
  "switch": {
    "hostname": "192.168.1.10",
    "model": "GS308EPP"
  },
  "tests": {
    "poe_cycling": {
      "name": "poe_cycling",
      "success": false,
      "message": "failed on ports: [3]",
      "timestamp": "2025-08-01T15:30:45Z"
    },
    "bandwidth_limiting": {
      "name": "bandwidth_limiting", 
      "success": true,
      "message": "all 8 ports",
      "timestamp": "2025-08-01T15:32:12Z"
    },
    "led_control": {
      "name": "led_control",
      "success": true,
      "message": "LED control test completed",
      "timestamp": "2025-08-01T15:33:01Z"
    }
  },
  "summary": {
    "total_operations": 45,
    "successful": 42,
    "failed": 3,
    "duration_seconds": 135,
    "final_state": "some_failed"
  }
}
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All tests passed successfully |
| 1 | Some tests failed but switch state was restored |
| 2 | Critical failure, switch state may not be fully restored |
| 3 | Authentication or connection failure |
| 4 | Invalid arguments or configuration |
| 5 | Interrupted by user (SIGINT/SIGTERM) |

## Safety Features

### State Management
- **Automatic Backup**: Captures complete switch state before any modifications
- **Automatic Restore**: Restores original settings even if program is interrupted
- **State Validation**: Verifies restoration by comparing final state with initial state
- **Graceful Cleanup**: Handles interruption signals (Ctrl+C) with proper cleanup

### Error Handling
- **Operation Isolation**: Failure in one test doesn't prevent others from running
- **Retry Logic**: Automatic retries for transient failures
- **Timeout Protection**: Operations abort if they take too long
- **Detailed Logging**: Comprehensive error reporting with context

### User Protection  
- **Dry-Run Mode**: Preview all operations without executing them
- **Skip Options**: Selectively disable potentially disruptive tests
- **Progress Reporting**: Clear indication of current operation and progress
- **Interruption Support**: Clean way to stop execution with proper cleanup

## Troubleshooting

### Common Issues

**Authentication Failures**
```bash
# Verify environment variables are set
echo $NETGEAR_SWITCHES
echo $NETGEAR_PASSWORD_SWITCH1

# Test authentication with simple example first
./build/examples/poe_status_simple switch1
```

**Timeout Issues**
```bash
# Increase timeout for slower switches
./switch-test --timeout 60 switch1

# Increase delay between operations
./switch-test --delay 5 switch1
```

**Partial Test Failures**
```bash
# Run with verbose output to see detailed errors
./switch-test --verbose switch1

# Skip problematic test categories
./switch-test --skip-poe switch1  # If POE tests fail
```

**State Restoration Problems**
```bash
# Check if switch state was properly restored
./switch-test --dry-run switch1  # Preview expected state

# Manual verification with status examples
./build/examples/poe_status switch1
```

### Getting Help

For additional help:
```bash
./switch-test --help
```

## Development

The program consists of:

- `main.go` - CLI argument parsing and main program flow
- `state_manager.go` - State backup and restoration
- `test_operations.go` - Test sequence implementations  
- `reporter.go` - Test result reporting and output formatting

### Adding New Tests

To add new test sequences:

1. Add test configuration options to `Config` struct in `main.go`
2. Implement test logic in `test_operations.go`
3. Add result reporting calls to `reporter.go`
4. Update command-line help and documentation

### Testing the Test Program

```bash
# Test against mock or development switch
export NETGEAR_SWITCHES="testswitch=password123"
./switch-test --dry-run testswitch

# Verify JSON output format
./switch-test --json --dry-run testswitch | jq .
```