package netgear

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"ntgrrc/pkg/netgear/internal"
)

// PortManager handles port-related operations
type PortManager struct {
	client *Client
	parser *internal.PortDataParser
}

// newPortManager creates a new port manager (internal constructor)
func newPortManager(client *Client) *PortManager {
	return &PortManager{
		client: client,
		parser: internal.NewPortDataParser(),
	}
}

// GetSettings retrieves port settings
func (m *PortManager) GetSettings(ctx context.Context) ([]PortSettings, error) {
	if !m.client.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	// Determine the appropriate endpoint based on model
	var endpoint string
	if m.client.model.IsModel30x() {
		endpoint = "/PortStatistics.cgi"
	} else if m.client.model.IsModel316() {
		endpoint = "/iss/specific/interface.html"
	} else {
		return nil, NewOperationError("port settings not supported for this model", nil)
	}

	// Make authenticated request
	response, err := m.client.makeAuthenticatedRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, NewOperationError("failed to get port settings", err)
	}

	// Parse the response
	rawData, err := m.parser.ParsePortSettings(response)
	if err != nil {
		return nil, NewParsingError("failed to parse port settings", err)
	}

	// Convert to strongly typed structures
	var settings []PortSettings
	for _, raw := range rawData {
		setting := PortSettings{}

		if portID, ok := raw["port_id"].(int); ok {
			setting.PortID = portID
		}
		if portName, ok := raw["port_name"].(string); ok {
			setting.PortName = portName
		}
		if speed, ok := raw["speed"].(string); ok {
			setting.Speed = PortSpeed(speed)
		}
		if ingressLimit, ok := raw["ingress_limit"].(string); ok {
			setting.IngressLimit = ingressLimit
		}
		if egressLimit, ok := raw["egress_limit"].(string); ok {
			setting.EgressLimit = egressLimit
		}
		if flowControl, ok := raw["flow_control"].(bool); ok {
			setting.FlowControl = flowControl
		}
		if status, ok := raw["status"].(string); ok {
			setting.Status = PortStatus(status)
		}
		if linkSpeed, ok := raw["link_speed"].(string); ok {
			setting.LinkSpeed = linkSpeed
		}

		settings = append(settings, setting)
	}

	return settings, nil
}

// UpdatePort updates settings for specific ports
func (m *PortManager) UpdatePort(ctx context.Context, updates ...PortUpdate) error {
	if !m.client.IsAuthenticated() {
		return ErrNotAuthenticated
	}

	if len(updates) == 0 {
		return NewOperationError("no updates provided", nil)
	}

	// Determine the appropriate endpoint based on model
	var endpoint string
	if m.client.model.IsModel30x() {
		endpoint = "/PortConfig.cgi"
	} else if m.client.model.IsModel316() {
		endpoint = "/iss/specific/interface.html"
	} else {
		return NewOperationError("port updates not supported for this model", nil)
	}

	// Apply each update
	for _, update := range updates {
		data := url.Values{}

		// Add port identification
		data.Set("port", strconv.Itoa(update.PortID))

		// Add updates based on what's provided
		if update.Name != nil {
			data.Set("name", *update.Name)
		}

		if update.Speed != nil {
			data.Set("speed", string(*update.Speed))
		}

		if update.IngressLimit != nil {
			data.Set("ingress_limit", *update.IngressLimit)
		}

		if update.EgressLimit != nil {
			data.Set("egress_limit", *update.EgressLimit)
		}

		if update.FlowControl != nil {
			if *update.FlowControl {
				data.Set("flow_control", "on")
			} else {
				data.Set("flow_control", "off")
			}
		}

		// Make the update request
		response, err := m.client.makeAuthenticatedRequest(ctx, "POST", endpoint, data)
		if err != nil {
			return NewOperationError(fmt.Sprintf("failed to update port %d", update.PortID), err)
		}

		// Check for errors in response
		if errorMsg := internal.ExtractErrorMessage(response); errorMsg != "" {
			return NewOperationError(fmt.Sprintf("update failed for port %d: %s", update.PortID, errorMsg), nil)
		}
	}

	return nil
}

// SetPortName sets the name for a specific port
func (m *PortManager) SetPortName(ctx context.Context, portID int, name string) error {
	return m.UpdatePort(ctx, PortUpdate{
		PortID: portID,
		Name:   &name,
	})
}

// SetPortSpeed sets the speed for a specific port
func (m *PortManager) SetPortSpeed(ctx context.Context, portID int, speed PortSpeed) error {
	return m.UpdatePort(ctx, PortUpdate{
		PortID: portID,
		Speed:  &speed,
	})
}

// SetPortFlowControl sets the flow control for a specific port
func (m *PortManager) SetPortFlowControl(ctx context.Context, portID int, enabled bool) error {
	return m.UpdatePort(ctx, PortUpdate{
		PortID:      portID,
		FlowControl: &enabled,
	})
}

// SetPortLimits sets the ingress and egress limits for a specific port
func (m *PortManager) SetPortLimits(ctx context.Context, portID int, ingressLimit, egressLimit string) error {
	return m.UpdatePort(ctx, PortUpdate{
		PortID:       portID,
		IngressLimit: &ingressLimit,
		EgressLimit:  &egressLimit,
	})
}

// GetPortSettings gets the settings for a specific port
func (m *PortManager) GetPortSettings(ctx context.Context, portID int) (*PortSettings, error) {
	settings, err := m.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	for _, setting := range settings {
		if setting.PortID == portID {
			return &setting, nil
		}
	}

	return nil, NewOperationError(fmt.Sprintf("port %d not found", portID), nil)
}

// DisablePort disables a specific port
func (m *PortManager) DisablePort(ctx context.Context, portID int) error {
	speed := PortSpeedDisable
	return m.UpdatePort(ctx, PortUpdate{
		PortID: portID,
		Speed:  &speed,
	})
}

// EnablePort enables a specific port with auto speed
func (m *PortManager) EnablePort(ctx context.Context, portID int) error {
	speed := PortSpeedAuto
	return m.UpdatePort(ctx, PortUpdate{
		PortID: portID,
		Speed:  &speed,
	})
}