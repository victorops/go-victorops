package victorops

import (
	"bytes"
	"context"
	"encoding/json"
)

// ChatMessage represents a chat message to send to VictorOps
type ChatMessage struct {
	Username         string   `json:"username"`
	ExternalUsername string   `json:"externalUsername"`
	Text             string   `json:"text"`
	MonitoringTool   string   `json:"monitoringTool"`
	IncidentID       int      `json:"incidentId,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// SendChatMessage sends a chat message into VictorOps
func (c *Client) SendChatMessage(ctx context.Context, message *ChatMessage) (*RequestDetails, error) {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/chat", bytes.NewBuffer(jsonMessage), nil)
	return details, err
}
