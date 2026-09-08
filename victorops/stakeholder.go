package victorops

import (
	"bytes"
	"context"
	"encoding/json"
)

// StakeholderMessage represents a message to send to stakeholders. At least one
// recipient is required, via Recipients (usernames) and/or RecipientTeams (team
// slugs).
type StakeholderMessage struct {
	Title          string   `json:"title"`
	Message        string   `json:"message"`
	IncidentID     int      `json:"incidentId"`
	Recipients     []string `json:"recipients"`
	RecipientTeams []string `json:"recipientTeams,omitempty"`
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
