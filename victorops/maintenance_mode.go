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
	CompanyID       string                    `json:"companyId,omitempty"`
}

// MaintenanceModeInstance represents an active maintenance mode instance
type MaintenanceModeInstance struct {
	InstanceID string                  `json:"instanceId,omitempty"`
	IsGlobal   bool                    `json:"isGlobal,omitempty"`
	StartedAt  int64                   `json:"startedAt,omitempty"` // millis from epoch (number from API)
	StartedBy  string                  `json:"startedBy,omitempty"`
	Targets    []MaintenanceModeTarget `json:"targets,omitempty"`
}

// MaintenanceModeTarget represents the routing keys covered by a maintenance
// mode instance. An empty Names list indicates global maintenance mode.
type MaintenanceModeTarget struct {
	Names []string `json:"names,omitempty"`
	Type  string   `json:"type,omitempty"` // "RoutingKeys"
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

// StartMaintenanceMode starts maintenance mode for the specified routing keys.
// If routingKeys is empty, this starts global maintenance mode.
func (c *Client) StartMaintenanceMode(ctx context.Context, routingKeys []string, purpose string) (*MaintenanceModeState, *RequestDetails, error) {
	// Ensure Names is an empty slice (not nil) for proper JSON marshalling.
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
	details, err := c.makePublicAPICall(ctx, "PUT", "v1/maintenancemode/"+url.PathEscape(maintenanceModeID)+"/end", bytes.NewBufferString("{}"), nil)
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
