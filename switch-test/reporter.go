package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// TestResult represents the result of a single test operation
type TestResult struct {
	Name      string    `json:"name"`
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// PortOperation represents a single operation on a port
type PortOperation struct {
	PortID    int       `json:"port_id"`
	Operation string    `json:"operation"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// TestSummary represents the overall test summary
type TestSummary struct {
	Switch struct {
		Hostname string `json:"hostname"`
		Model    string `json:"model"`
	} `json:"switch"`
	Tests struct {
		POECycling        *TestResult     `json:"poe_cycling,omitempty"`
		BandwidthLimiting *TestResult     `json:"bandwidth_limiting,omitempty"`
		LEDControl        *TestResult     `json:"led_control,omitempty"`
		PortOperations    []PortOperation `json:"port_operations,omitempty"`
	} `json:"tests"`
	Summary struct {
		TotalOperations int           `json:"total_operations"`
		Successful      int           `json:"successful"`
		Failed          int           `json:"failed"`
		Duration        time.Duration `json:"duration_seconds"`
		FinalState      string        `json:"final_state"`
	} `json:"summary"`
	Errors []TestResult `json:"errors,omitempty"`
}

// Reporter handles test result reporting
type Reporter struct {
	jsonOutput     bool
	verbose        bool
	testResults    []TestResult
	portOperations []PortOperation
	errors         []TestResult
	totalOps       int
	successfulOps  int
	failedOps      int
}

// NewReporter creates a new test reporter
func NewReporter(jsonOutput, verbose bool) *Reporter {
	return &Reporter{
		jsonOutput:     jsonOutput,
		verbose:        verbose,
		testResults:    make([]TestResult, 0),
		portOperations: make([]PortOperation, 0),
		errors:         make([]TestResult, 0),
	}
}

// PrintHeader prints the test program header
func (r *Reporter) PrintHeader(switchAddress string) {
	if r.jsonOutput {
		return
	}

	fmt.Println("Switch Test Program - ntgrrc Library Validation")
	fmt.Println("===============================================")
	fmt.Println()
}

// RecordTestResult records the result of a major test operation
func (r *Reporter) RecordTestResult(testName string, success bool, message string, err error) {
	result := TestResult{
		Name:      testName,
		Success:   success,
		Message:   message,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
	}

	r.testResults = append(r.testResults, result)
	r.totalOps++

	if success {
		r.successfulOps++
	} else {
		r.failedOps++
	}
}

// RecordPortOperation records the result of an operation on a specific port
func (r *Reporter) RecordPortOperation(portID int, operation string, success bool, err error) {
	op := PortOperation{
		PortID:    portID,
		Operation: operation,
		Success:   success,
		Timestamp: time.Now(),
	}

	if err != nil {
		op.Error = err.Error()
	}

	r.portOperations = append(r.portOperations, op)
	r.totalOps++

	if success {
		r.successfulOps++
	} else {
		r.failedOps++
	}

	// Print detailed operation result if verbose
	if r.verbose && !r.jsonOutput {
		status := "✓"
		if !success {
			status = "✗"
		}
		fmt.Printf("    %s Port %d %s", status, portID, operation)
		if err != nil {
			fmt.Printf(": %v", err)
		}
		fmt.Println()
	}
}

// RecordError records a general error
func (r *Reporter) RecordError(operation string, err error) {
	errorResult := TestResult{
		Name:      operation,
		Success:   false,
		Error:     err.Error(),
		Timestamp: time.Now(),
	}

	r.errors = append(r.errors, errorResult)
	r.totalOps++
	r.failedOps++
}

// PrintFinalReport prints the final test summary
func (r *Reporter) PrintFinalReport(duration time.Duration) {
	if r.jsonOutput {
		r.printJSONReport(duration)
	} else {
		r.printTextReport(duration)
	}
}

// printTextReport prints a human-readable final report
func (r *Reporter) printTextReport(duration time.Duration) {
	fmt.Println("\nTest Summary:")
	fmt.Printf("  Total Operations: %d\n", r.totalOps)
	fmt.Printf("  Successful: %d\n", r.successfulOps)  
	fmt.Printf("  Failed: %d\n", r.failedOps)
	fmt.Printf("  Duration: %v\n", duration.Round(time.Second))

	// Print test results summary
	if len(r.testResults) > 0 {
		fmt.Println("\nTest Results:")
		for _, result := range r.testResults {
			status := "✓"
			if !result.Success {
				status = "✗"
			}
			fmt.Printf("  %s %s", status, result.Name)
			if result.Message != "" {
				fmt.Printf(" (%s)", result.Message)
			}
			if result.Error != "" {
				fmt.Printf(" - Error: %s", result.Error)
			}
			fmt.Println()
		}
	}

	// Print error summary if there were errors
	if len(r.errors) > 0 {
		fmt.Println("\nErrors Encountered:")
		for _, err := range r.errors {
			fmt.Printf("  ✗ %s: %s\n", err.Name, err.Error)
		}
	}

	// Print overall result
	fmt.Println()
	if r.failedOps == 0 {
		fmt.Println("🎉 All tests completed successfully!")
	} else {
		fmt.Printf("⚠ %d operations failed out of %d total\n", r.failedOps, r.totalOps)
	}
}

// printJSONReport prints a JSON-formatted report
func (r *Reporter) printJSONReport(duration time.Duration) {
	summary := TestSummary{}
	
	// Find specific test results
	for _, result := range r.testResults {
		switch result.Name {
		case "poe_cycling":
			summary.Tests.POECycling = &result
		case "bandwidth_limiting":
			summary.Tests.BandwidthLimiting = &result
		case "led_control":
			summary.Tests.LEDControl = &result
		}
	}

	summary.Tests.PortOperations = r.portOperations

	summary.Summary.TotalOperations = r.totalOps
	summary.Summary.Successful = r.successfulOps
	summary.Summary.Failed = r.failedOps
	summary.Summary.Duration = duration
	
	if r.failedOps == 0 {
		summary.Summary.FinalState = "all_passed"
	} else {
		summary.Summary.FinalState = "some_failed"
	}

	summary.Errors = r.errors

	// Marshal to JSON and print
	jsonBytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON report: %v\n", err)
		return
	}

	fmt.Println(string(jsonBytes))
}

// GetSuccessCount returns the number of successful operations
func (r *Reporter) GetSuccessCount() int {
	return r.successfulOps
}

// GetFailureCount returns the number of failed operations
func (r *Reporter) GetFailureCount() int {
	return r.failedOps
}

// GetTotalOperations returns the total number of operations
func (r *Reporter) GetTotalOperations() int {
	return r.totalOps
}

// HasErrors returns true if any errors were recorded
func (r *Reporter) HasErrors() bool {
	return r.failedOps > 0 || len(r.errors) > 0
}

// GetTestResults returns all recorded test results
func (r *Reporter) GetTestResults() []TestResult {
	return r.testResults
}

// GetPortOperations returns all recorded port operations
func (r *Reporter) GetPortOperations() []PortOperation {
	return r.portOperations
}

// PrintProgressUpdate prints a progress update (only in non-JSON mode)
func (r *Reporter) PrintProgressUpdate(message string) {
	if !r.jsonOutput && r.verbose {
		fmt.Printf("  %s\n", message)
	}
}