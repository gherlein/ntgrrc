# Real-World Test Example Program Design

## Overview

This document describes the design for a comprehensive test example program that exercises all major POE and port management functionality of the ntgrrc library. The program performs a series of real-world operations on a Netgear switch to validate the library's functionality and can be used for both testing and demonstration purposes.

## Program Purpose

- **Testing**: Validate library functionality against real hardware
- **Demonstration**: Show comprehensive usage of all library features
- **Validation**: Verify that operations actually take effect on the switch
- **Safety**: Include proper error handling and state restoration

## Test Sequence

The program will execute the following test sequence:

### 1. Initial Setup and Discovery
- Connect to switch using environment-based authentication
- Detect switch model and capabilities
- Get initial POE port count and status
- Get initial port settings and bandwidth configurations
- Display summary of switch configuration

### 2. POE Port Power Cycling Test
For each POE-capable port (in sequence):
- **Disable POE**: Turn off power to the port
- **Verify Off**: Check POE status to confirm power is actually off
- **Wait**: Brief delay to ensure state change is registered
- **Enable POE**: Turn power back on to the port
- **Verify On**: Check POE status to confirm power is restored
- **Report**: Log success/failure for each port operation

### 3. Bandwidth Limitation Test  
For each ethernet port (in sequence):
- **Get Current Settings**: Record original bandwidth settings
- **Set Low Bandwidth**: Reduce both ingress and egress to 1 Mbps
- **Verify Setting**: Read back configuration to confirm change
- **Wait**: Brief delay for settings to take effect
- **Restore Bandwidth**: Set back to original/unlimited bandwidth
- **Verify Restoration**: Confirm bandwidth is restored
- **Report**: Log success/failure for each bandwidth operation

### 4. LED Control Test
- **Get LED Status**: Record current LED state (if supported)
- **Disable LEDs**: Turn off all port LEDs
- **Verify Off**: Check that LEDs are actually disabled
- **Wait**: Brief delay for visual confirmation
- **Enable LEDs**: Turn LEDs back on
- **Verify On**: Check that LEDs are restored
- **Report**: Success/failure of LED operations

### 5. Final Validation and Cleanup
- **Status Check**: Get final POE and port status
- **Compare States**: Verify all settings match initial state
- **Generate Report**: Summary of all operations and results
- **Exit**: Clean shutdown with appropriate exit code

## Program Structure

```
switch-test/
├── main.go              # Main program logic
├── test_operations.go   # Test operation implementations
├── state_manager.go     # State backup/restore functionality
├── reporter.go          # Test result reporting
└── README.md           # Usage instructions
```

## Command Line Interface

```bash
switch-test [options] <switch-hostname>

Options:
  --debug, -d           Enable debug output
  --dry-run            Show what would be done without executing
  --skip-poe           Skip POE power cycling tests
  --skip-bandwidth     Skip bandwidth limitation tests  
  --skip-leds          Skip LED control tests
  --delay <seconds>    Delay between operations (default: 2)
  --timeout <seconds>  Operation timeout (default: 30)
  --help, -h           Show help message

Environment:
  NETGEAR_SWITCHES="host=password;..."
  NETGEAR_PASSWORD_<HOST>=password
```

## Safety Features

### State Backup and Restore
- **Initial State Capture**: Record all port settings before any changes
- **Automatic Restore**: Restore original settings if program is interrupted
- **Graceful Shutdown**: Handle SIGINT/SIGTERM to cleanup properly
- **Error Recovery**: Attempt to restore state even after failures

### Operation Validation
- **Pre-flight Checks**: Verify switch is responsive before starting
- **Step Verification**: Confirm each operation actually took effect
- **Timeout Protection**: Abort operations that take too long
- **Error Boundaries**: Isolate failures to prevent cascade effects

### User Safety
- **Dry Run Mode**: Preview operations without executing them
- **Confirmation Prompts**: Optional user confirmation for destructive operations
- **Progress Reporting**: Clear indication of current operation
- **Abort Capability**: Clean way to stop mid-execution

## Error Handling Strategy

### Operation-Level Errors
- **Retry Logic**: Attempt failed operations up to 3 times
- **Graceful Degradation**: Continue with remaining tests if one fails
- **Error Classification**: Distinguish between network, auth, and device errors
- **Context Preservation**: Include operation context in error messages

### Program-Level Errors
- **Signal Handling**: Catch interruption signals for cleanup
- **Resource Cleanup**: Ensure network connections are properly closed
- **State Restoration**: Restore original settings before exit
- **Exit Codes**: Use appropriate exit codes for different failure types

## Test Result Reporting

### Real-Time Output
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

Final Status:
  ✓ All ports restored to original state
  ✓ All settings match initial configuration

Test Summary:
  Total Operations: 45
  Successful: 42
  Failed: 3
  Duration: 2m 15s
```

### JSON Output Mode
```json
{
  "switch": {
    "hostname": "192.168.1.10",
    "model": "GS308EPP",
    "poe_ports": 8,
    "ethernet_ports": 8
  },
  "tests": {
    "poe_cycling": {
      "total_ports": 8,
      "successful": 7,
      "failed": 1,
      "operations": [...]
    },
    "bandwidth_limiting": {
      "total_ports": 8,
      "successful": 8,
      "failed": 0,
      "operations": [...]
    },
    "led_control": {
      "successful": true,
      "operations": [...]
    }
  },
  "summary": {
    "total_operations": 45,
    "successful": 42,
    "failed": 3,
    "duration_seconds": 135,
    "final_state": "restored"
  }
}
```

## Implementation Considerations

### Performance
- **Parallel Operations**: Where safe, perform operations in parallel
- **Caching**: Cache switch state to minimize redundant queries
- **Connection Reuse**: Maintain single connection throughout test
- **Efficient Polling**: Use appropriate intervals for status checking

### Compatibility
- **Model Detection**: Adapt operations based on detected switch model
- **Feature Detection**: Skip unsupported operations gracefully  
- **Version Handling**: Handle different firmware versions appropriately
- **Fallback Logic**: Alternative approaches for different switch types

### Reliability
- **Network Resilience**: Handle temporary network issues
- **Authentication Refresh**: Re-authenticate if session expires
- **State Verification**: Double-check critical state changes
- **Idempotent Operations**: Safe to run multiple times

## Usage Examples

### Basic Test Run
```bash
# Test all functionality with environment authentication
export NETGEAR_SWITCHES="switch1=password123"
./switch-test switch1
```

### Debug Mode with Selective Testing
```bash
# Run only POE tests with detailed output
./switch-test --debug --skip-bandwidth --skip-leds switch1
```

### Dry Run Preview
```bash
# Preview operations without executing them
./switch-test --dry-run --debug switch1
```

### CI/CD Integration
```bash
# JSON output for automated testing
./switch-test --json --timeout 60 switch1 > test-results.json
echo $? # Check exit code
```

## Exit Codes

- **0**: All tests passed successfully
- **1**: Some tests failed but switch state was restored
- **2**: Critical failure, switch state may not be fully restored
- **3**: Authentication or connection failure
- **4**: Invalid arguments or configuration
- **5**: Interrupted by user (SIGINT/SIGTERM)

## Security Considerations

- **Password Handling**: Never log or display passwords
- **Network Security**: Use appropriate timeouts and error handling
- **State Protection**: Prevent unauthorized state changes during test
- **Audit Trail**: Log all operations for security review