package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// MaintenanceModeState represents the current maintenance mode state
type MaintenanceModeState struct {
	ActiveInstances []MaintenanceModeInstance `json:"activeInstances,omitempty"`
}

// MaintenanceModeInstance represents an active maintenance mode instance
type MaintenanceModeInstance struct {
	InstanceID    string   `json:"instanceId,omitempty"`
	IsGlobal      bool     `json:"isGlobal,omitempty"`
	StartedAt     int64    `json:"startedAt,omitempty"` // Unix timestamp (number from API)
	StartedBy     string   `json:"startedBy,omitempty"`
	Purpose       string   `json:"purpose,omitempty"`
	RoutingKeys   []string `json:"routingKeys,omitempty"`
	AffectedTeams []string `json:"affectedTeams,omitempty"`
}

// StartMaintenanceModePayload is the payload for starting maintenance mode
type StartMaintenanceModePayload struct {
	Type    string   `json:"type"`
	Names   []string `json:"names"` // Always include names, even if empty array
	Purpose string   `json:"purpose,omitempty"`
}

// GetMaintenanceModeState gets the current maintenance mode state for the organization
func (c *Client) GetMaintenanceModeState(ctx context.Context) (*MaintenanceModeState, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/maintenancemode", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var state MaintenanceModeState
	err = json.Unmarshal([]byte(details.ResponseBody), &state)
	if err != nil {
		return nil, details, err
	}

	return &state, details, nil
}

// StartMaintenanceMode starts maintenance mode for the specified routing keys
// If names is empty, this starts global maintenance mode
func (c *Client) StartMaintenanceMode(ctx context.Context, routingKeys []string, purpose string) (*MaintenanceModeState, *RequestDetails, error) {
	// Ensure Names is an empty slice (not nil) for proper JSON marshalling
	names := routingKeys
	if names == nil {
		names = []string{}
	}
	payload := StartMaintenanceModePayload{
		Type:    "RoutingKeys",
		Names:   names,
		Purpose: purpose,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/maintenancemode/start", bytes.NewBuffer(jsonPayload), nil)
	if err != nil {
		return nil, details, err
	}

	var state MaintenanceModeState
	err = json.Unmarshal([]byte(details.ResponseBody), &state)
	if err != nil {
		return nil, details, err
	}

	return &state, details, nil
}

// EndMaintenanceMode ends a specific maintenance mode instance
func (c *Client) EndMaintenanceMode(ctx context.Context, maintenanceModeID string) (*MaintenanceModeState, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "PUT", "v1/maintenancemode/"+url.QueryEscape(maintenanceModeID)+"/end", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var state MaintenanceModeState
	err = json.Unmarshal([]byte(details.ResponseBody), &state)
	if err != nil {
		return nil, details, err
	}

	return &state, details, nil
}
