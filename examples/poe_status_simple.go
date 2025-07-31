// poe_status_simple.go - Simple example using the ntgrrc library
// This version uses environment variables for the password.
//
// Usage: 
//   export NETGEAR_PASSWORD=your_password
//   go run poe_status_simple.go [--debug|-d] <switch-hostname>

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"ntgrrc/pkg/netgear"
)

func main() {
	// Parse command line flags
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Enable debug output")
	flag.BoolVar(&debug, "d", false, "Enable debug output (shorthand)")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--debug|-d] <switch-hostname>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Set NETGEAR_PASSWORD environment variable\n")
		os.Exit(1)
	}

	switchAddress := args[0]
	password := os.Getenv("NETGEAR_PASSWORD")
	if password == "" {
		log.Fatal("NETGEAR_PASSWORD environment variable not set")
	}

	// Create client with optional debug logging
	var client *netgear.Client
	var err error
	
	if debug {
		client, err = netgear.NewClient(switchAddress, netgear.WithVerbose(true))
	} else {
		client, err = netgear.NewClient(switchAddress)
	}
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Login
	ctx := context.Background()
	if debug {
		fmt.Printf("Debug mode enabled\n")
		fmt.Printf("Connecting to: %s\n", switchAddress)
	}
	
	err = client.Login(ctx, password)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Printf("✓ Logged in to %s (Model: %s)\n\n", switchAddress, client.GetModel())

	// Get POE status
	if debug {
		fmt.Println("Fetching POE status...")
	}
	statuses, err := client.POE().GetStatus(ctx)
	if err != nil {
		log.Fatalf("Failed to get POE status: %v", err)
	}

	if debug {
		fmt.Printf("Retrieved data for %d ports\n", len(statuses))
	}

	// Print results
	fmt.Println("POE Port Status:")
	for _, status := range statuses {
		fmt.Printf("Port %d: %s", status.PortID, status.Status)
		if status.PowerW > 0 {
			fmt.Printf(" (%.1fW)", status.PowerW)
		}
		if status.PortName != "" {
			fmt.Printf(" - %s", status.PortName)
		}
		fmt.Println()
	}
}