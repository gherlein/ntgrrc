package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"ntgrrc/pkg/netgear"
)

// SwitchState represents the complete state of a switch
type SwitchState struct {
	POEStatus    []netgear.POEPortStatus
	POESettings  []netgear.POEPortSettings
	PortSettings []netgear.PortSettings
	LEDStatus    bool // Whether LEDs are enabled
	Timestamp    time.Time
}

// StateManager handles backup and restoration of switch state
type StateManager struct {
	client       *netgear.Client
	initialState *SwitchState
	debug        bool
}

// NewStateManager creates a new state manager
func NewStateManager(client *netgear.Client, debug bool) *StateManager {
	return &StateManager{
		client: client,
		debug:  debug,
	}
}

// CaptureInitialState captures the current state of the switch
func (sm *StateManager) CaptureInitialState(ctx context.Context) error {
	if sm.debug {
		fmt.Println("Capturing initial switch state...")
	}

	state := &SwitchState{
		Timestamp: time.Now(),
	}

	// Capture POE status
	poeStatus, err := sm.client.POE().GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get POE status: %w", err)
	}
	state.POEStatus = poeStatus

	// Capture POE settings
	poeSettings, err := sm.client.POE().GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get POE settings: %w", err)
	}
	state.POESettings = poeSettings

	// Capture port settings (bandwidth, etc.)
	portSettings, err := sm.client.Ports().GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get port settings: %w", err)
	}
	state.PortSettings = portSettings

	// LED status is not available in the current library implementation
	// Assume LEDs are enabled by default
	state.LEDStatus = true

	sm.initialState = state

	if sm.debug {
		fmt.Printf("Initial state captured: %d POE ports, %d ethernet ports, LEDs %s\n",
			len(state.POEStatus), len(state.PortSettings), 
			map[bool]string{true: "enabled", false: "disabled"}[state.LEDStatus])
	}

	return nil
}

// GetInitialState returns the captured initial state
func (sm *StateManager) GetInitialState() *SwitchState {
	return sm.initialState
}

// RestoreState restores the switch to its initial state
func (sm *StateManager) RestoreState(ctx context.Context) error {
	if sm.initialState == nil {
		return fmt.Errorf("no initial state captured")
	}

	if sm.debug {
		fmt.Println("Restoring switch to initial state...")
	}

	var errors []error

	// Restore POE settings first (this includes enable/disable state)
	for _, setting := range sm.initialState.POESettings {
		// Create POE update to restore settings
		update := netgear.POEPortUpdate{
			PortID:              setting.PortID,
			Enabled:             &setting.Enabled,
			Mode:                &setting.Mode,
			Priority:            &setting.Priority,
			PowerLimitType:      &setting.PowerLimitType,
			PowerLimitW:         &setting.PowerLimitW,
		}
		
		err := sm.client.POE().UpdatePort(ctx, update)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to restore POE settings for port %d: %w", setting.PortID, err))
			continue
		}
		
		// Small delay between operations to avoid overwhelming the switch
		time.Sleep(100 * time.Millisecond)
	}

	// Restore port settings (bandwidth, etc.)
	for _, setting := range sm.initialState.PortSettings {
		// Create port update to restore settings
		update := netgear.PortUpdate{
			PortID:       setting.PortID,
			Name:         &setting.PortName,
			Speed:        &setting.Speed,
			IngressLimit: &setting.IngressLimit,
			EgressLimit:  &setting.EgressLimit,
			FlowControl:  &setting.FlowControl,
		}
		
		err := sm.client.Ports().UpdatePort(ctx, update)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to restore port settings for port %d: %w", setting.PortID, err))
			continue
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	// LED control is not available in the current library implementation
	// Skip LED restoration

	// Return combined errors if any
	if len(errors) > 0 {
		var errMsg string
		for i, err := range errors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return fmt.Errorf("state restoration encountered %d errors: %s", len(errors), errMsg)
	}

	if sm.debug {
		fmt.Println("Switch state restored successfully")
	}

	return nil
}

// ValidateStateRestoration verifies that the current state matches the initial state
func (sm *StateManager) ValidateStateRestoration(ctx context.Context) error {
	if sm.initialState == nil {
		return fmt.Errorf("no initial state to compare against")
	}

	if sm.debug {
		fmt.Println("Validating state restoration...")
	}

	// Get current state
	currentPOESettings, err := sm.client.POE().GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current POE settings: %w", err)
	}

	currentPortSettings, err := sm.client.Ports().GetSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current port settings: %w", err)
	}

	// LED status validation is not available in the current library implementation
	// Skip LED validation
	currentLEDStatus := true // Assume LEDs are enabled

	// Compare POE settings
	if !sm.comparePOESettings(sm.initialState.POESettings, currentPOESettings) {
		return fmt.Errorf("POE settings do not match initial state")
	}

	// Compare port settings
	if !sm.comparePortSettings(sm.initialState.PortSettings, currentPortSettings) {
		return fmt.Errorf("port settings do not match initial state")
	}

	// Compare LED status (if available) - skip since not implemented
	_ = currentLEDStatus // Prevent unused variable warning

	if sm.debug {
		fmt.Println("State validation successful - all settings match initial state")
	}

	return nil
}

// comparePOESettings compares two sets of POE settings for equality
func (sm *StateManager) comparePOESettings(initial, current []netgear.POEPortSettings) bool {
	if len(initial) != len(current) {
		return false
	}

	// Create maps for easier comparison
	initialMap := make(map[int]netgear.POEPortSettings)
	currentMap := make(map[int]netgear.POEPortSettings)

	for _, setting := range initial {
		initialMap[setting.PortID] = setting
	}
	
	for _, setting := range current {
		currentMap[setting.PortID] = setting
	}

	// Compare each port
	for portID, initialSetting := range initialMap {
		currentSetting, exists := currentMap[portID]
		if !exists {
			if sm.debug {
				fmt.Printf("Port %d missing in current POE settings\n", portID)
			}
			return false
		}

		// Compare key fields (ignore transient fields like current power draw)
		if initialSetting.Enabled != currentSetting.Enabled ||
		   initialSetting.Mode != currentSetting.Mode ||
		   initialSetting.Priority != currentSetting.Priority ||
		   initialSetting.PowerLimitW != currentSetting.PowerLimitW {
			if sm.debug {
				fmt.Printf("POE settings mismatch for port %d: initial=%+v, current=%+v\n", 
					portID, initialSetting, currentSetting)
			}
			return false
		}
	}

	return true
}

// comparePortSettings compares two sets of port settings for equality
func (sm *StateManager) comparePortSettings(initial, current []netgear.PortSettings) bool {
	if len(initial) != len(current) {
		return false
	}

	// Create maps for easier comparison
	initialMap := make(map[int]netgear.PortSettings)
	currentMap := make(map[int]netgear.PortSettings)

	for _, setting := range initial {
		initialMap[setting.PortID] = setting
	}
	
	for _, setting := range current {
		currentMap[setting.PortID] = setting
	}

	// Compare each port
	for portID, initialSetting := range initialMap {
		currentSetting, exists := currentMap[portID]
		if !exists {
			if sm.debug {
				fmt.Printf("Port %d missing in current port settings\n", portID)
			}
			return false
		}

		// Compare key fields
		if !reflect.DeepEqual(initialSetting, currentSetting) {
			if sm.debug {
				fmt.Printf("Port settings mismatch for port %d: initial=%+v, current=%+v\n", 
					portID, initialSetting, currentSetting)
			}
			return false
		}
	}

	return true
}

// GetStateSummary returns a human-readable summary of the current state
func (sm *StateManager) GetStateSummary() string {
	if sm.initialState == nil {
		return "No state captured"
	}

	state := sm.initialState
	return fmt.Sprintf("POE Ports: %d, Ethernet Ports: %d, LEDs: %s, Captured: %s",
		len(state.POEStatus), len(state.PortSettings),
		map[bool]string{true: "enabled", false: "disabled"}[state.LEDStatus],
		state.Timestamp.Format("15:04:05"))
}