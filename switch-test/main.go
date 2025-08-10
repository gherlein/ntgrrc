// switch-test - Comprehensive real-world test program for the ntgrrc library
// This program exercises all major POE and port management functionality
// against actual Netgear switch hardware for testing and validation.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ntgrrc/pkg/netgear"
)

// Config holds all program configuration
type Config struct {
	SwitchAddress string
	Debug         bool
	DryRun        bool
	SkipPOE       bool
	SkipBandwidth bool
	SkipLEDs      bool
	JSONOutput    bool
	Delay         time.Duration
	Timeout       time.Duration
	Verbose       bool
}

// TestContext holds the test execution context
type TestContext struct {
	Config       *Config
	Client       *netgear.Client
	StateManager *StateManager
	Reporter     *Reporter
	StartTime    time.Time
	Interrupted  bool
}

func main() {
	config := parseFlags()
	
	if config == nil {
		os.Exit(4) // Invalid arguments
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Create test context
	testCtx := &TestContext{
		Config:    config,
		StartTime: time.Now(),
	}

	// Handle signals in goroutine
	go func() {
		<-signalChan
		fmt.Println("\n⚠ Interrupt received, cleaning up...")
		testCtx.Interrupted = true
		cancel()
	}()

	// Run the test program
	exitCode := runTests(ctx, testCtx)
	
	// Ensure cleanup happens
	if testCtx.StateManager != nil {
		if err := testCtx.StateManager.RestoreState(ctx); err != nil {
			fmt.Printf("⚠ Warning: Failed to restore state: %v\n", err)
			if exitCode == 0 {
				exitCode = 2 // Critical failure
			}
		}
	}

	if testCtx.Interrupted {
		fmt.Println("Program interrupted by user")
		os.Exit(5)
	}

	os.Exit(exitCode)
}

// parseFlags parses command line arguments and returns configuration
func parseFlags() *Config {
	config := &Config{}

	flag.BoolVar(&config.Debug, "debug", false, "Enable debug output")
	flag.BoolVar(&config.Debug, "d", false, "Enable debug output (shorthand)")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be done without executing")
	flag.BoolVar(&config.SkipPOE, "skip-poe", false, "Skip POE power cycling tests")
	flag.BoolVar(&config.SkipBandwidth, "skip-bandwidth", false, "Skip bandwidth limitation tests")
	flag.BoolVar(&config.SkipLEDs, "skip-leds", false, "Skip LED control tests")
	flag.BoolVar(&config.JSONOutput, "json", false, "Output results in JSON format")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output (more details)")

	var delaySeconds int
	var timeoutSeconds int
	flag.IntVar(&delaySeconds, "delay", 2, "Delay between operations in seconds")
	flag.IntVar(&timeoutSeconds, "timeout", 30, "Operation timeout in seconds")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Switch Test Program - ntgrrc Library Validation\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <switch-hostname>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "This program performs comprehensive testing of switch functionality:\n")
		fmt.Fprintf(os.Stderr, "1. POE port power cycling (disable → verify → enable → verify)\n")
		fmt.Fprintf(os.Stderr, "2. Bandwidth limitation testing (limit → verify → restore)\n")
		fmt.Fprintf(os.Stderr, "3. LED control testing (disable → enable → verify)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  NETGEAR_SWITCHES=\"host=password;...\"\n")
		fmt.Fprintf(os.Stderr, "  NETGEAR_PASSWORD_<HOST>=password\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s switch1                    # Test all functionality\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --dry-run switch1          # Preview operations\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --skip-poe --debug switch1 # Test only bandwidth and LEDs\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json switch1 > results.json # JSON output for CI/CD\n", os.Args[0])
	}

	flag.Parse()

	// Validate arguments
	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		return nil
	}

	config.SwitchAddress = args[0]
	config.Delay = time.Duration(delaySeconds) * time.Second
	config.Timeout = time.Duration(timeoutSeconds) * time.Second

	// Validate configuration
	if config.Delay < 0 || config.Delay > 60*time.Second {
		fmt.Fprintf(os.Stderr, "Error: delay must be between 0 and 60 seconds\n")
		return nil
	}

	if config.Timeout < 5*time.Second || config.Timeout > 300*time.Second {
		fmt.Fprintf(os.Stderr, "Error: timeout must be between 5 and 300 seconds\n")
		return nil
	}

	// If JSON output is requested, suppress other output
	if config.JSONOutput {
		config.Debug = false
		config.Verbose = false
	}

	return config
}

// runTests executes the main test sequence
func runTests(ctx context.Context, testCtx *TestContext) int {
	config := testCtx.Config

	// Initialize reporter
	testCtx.Reporter = NewReporter(config.JSONOutput, config.Verbose)
	
	if !config.JSONOutput {
		testCtx.Reporter.PrintHeader(config.SwitchAddress)
	}

	// Step 1: Connect to switch
	if !config.JSONOutput {
		fmt.Printf("Connecting to switch: %s\n", config.SwitchAddress)
	}

	client, err := createClient(config)
	if err != nil {
		testCtx.Reporter.RecordError("connection", err)
		if !config.JSONOutput {
			fmt.Printf("✗ Failed to connect: %v\n", err)
		}
		testCtx.Reporter.PrintFinalReport(time.Since(testCtx.StartTime))
		return 3 // Connection failure
	}
	testCtx.Client = client

	if !config.JSONOutput {
		fmt.Printf("✓ Authentication successful (Model: %s)\n", client.GetModel())
	}

	// Step 2: Initialize state management
	testCtx.StateManager = NewStateManager(client, config.Debug)
	if err := testCtx.StateManager.CaptureInitialState(ctx); err != nil {
		testCtx.Reporter.RecordError("state_capture", err)
		if !config.JSONOutput {
			fmt.Printf("✗ Failed to capture initial state: %v\n", err)
		}
		testCtx.Reporter.PrintFinalReport(time.Since(testCtx.StartTime))
		return 2 // Critical failure
	}

	if !config.JSONOutput {
		state := testCtx.StateManager.GetInitialState()
		fmt.Printf("✓ Detected %d POE ports, %d ethernet ports\n", 
			len(state.POEStatus), len(state.PortSettings))
	}

	// Step 3: Run test operations
	testOps := NewTestOperations(client, testCtx.Reporter, config)
	
	var overallSuccess bool = true

	// POE Power Cycling Test
	if !config.SkipPOE {
		if !config.JSONOutput {
			fmt.Println("\nPOE Power Cycling Test:")
		}
		success := testOps.RunPOECyclingTest(ctx)
		overallSuccess = overallSuccess && success
		
		if testCtx.Interrupted {
			return 5
		}
	}

	// Bandwidth Limitation Test
	if !config.SkipBandwidth {
		if !config.JSONOutput {
			fmt.Println("\nBandwidth Limitation Test:")
		}
		success := testOps.RunBandwidthTest(ctx, testCtx.StateManager.GetInitialState())
		overallSuccess = overallSuccess && success
		
		if testCtx.Interrupted {
			return 5
		}
	}

	// LED Control Test
	if !config.SkipLEDs {
		if !config.JSONOutput {
			fmt.Println("\nLED Control Test:")
		}
		success := testOps.RunLEDTest(ctx)
		overallSuccess = overallSuccess && success
		
		if testCtx.Interrupted {
			return 5
		}
	}

	// Step 4: Final validation
	if !config.JSONOutput {
		fmt.Println("\nFinal Validation:")
	}
	
	if err := testCtx.StateManager.ValidateStateRestoration(ctx); err != nil {
		testCtx.Reporter.RecordError("final_validation", err)
		if !config.JSONOutput {
			fmt.Printf("✗ State validation failed: %v\n", err)
		}
		overallSuccess = false
	} else {
		if !config.JSONOutput {
			fmt.Println("✓ All settings restored to original state")
		}
	}

	// Print final report
	duration := time.Since(testCtx.StartTime)
	testCtx.Reporter.PrintFinalReport(duration)

	// Determine exit code
	if overallSuccess {
		return 0 // All tests passed
	} else {
		return 1 // Some tests failed but state restored
	}
}

// createClient creates and authenticates a netgear client
func createClient(config *Config) (*netgear.Client, error) {
	var clientOpts []netgear.ClientOption
	
	if config.Debug {
		clientOpts = append(clientOpts, netgear.WithVerbose(true))
	}
	
	// Set timeout
	clientOpts = append(clientOpts, netgear.WithTimeout(config.Timeout))

	client, err := netgear.NewClient(config.SwitchAddress, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// If not already authenticated, try LoginAuto
	ctx := context.Background()
	if !client.IsAuthenticated() {
		err = client.LoginAuto(ctx)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w\nEnsure environment variables are set:\n  NETGEAR_SWITCHES=\"host=password;...\"\n  OR NETGEAR_PASSWORD_<HOST>=password", err)
		}
	}

	return client, nil
}