package victorops

import (
	"bytes"
	"context"
	"encoding/json"
)

// StakeholderMessage represents a message to send to stakeholders
type StakeholderMessage struct {
	Summary    string   `json:"summary"`
	Details    string   `json:"details,omitempty"`
	Users      []string `json:"users,omitempty"`
	Teams      []string `json:"teams,omitempty"`
	IncidentID int      `json:"incidentId,omitempty"`
}

// SendStakeholderMessage sends a message to stakeholders
func (c *Client) SendStakeholderMessage(ctx context.Context, message *StakeholderMessage) (*RequestDetails, error) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/stakeholders/sendMessage", bytes.NewBuffer(jsonMessage), nil)
	return details, err
}
