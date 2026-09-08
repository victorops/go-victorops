package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// Alert represents the details of an alert as returned by the alerts endpoint.
// All fields are optional. Per the API, timestamp/stateStartTime are epoch
// seconds (numeric) and raw is the full original alert payload as a JSON string.
type Alert struct {
	AckAuthor         string  `json:"ackAuthor,omitempty"`
	AckMsg            string  `json:"ackMsg,omitempty"`
	EntityDisplayName string  `json:"entityDisplayName,omitempty"`
	EntityID          string  `json:"entityId,omitempty"`
	MessageType       string  `json:"messageType,omitempty"`
	MonitoringTool    string  `json:"monitoringTool,omitempty"`
	Raw               string  `json:"raw,omitempty"`
	StateMessage      string  `json:"stateMessage,omitempty"`
	StateStartTime    float64 `json:"stateStartTime,omitempty"`
	Timestamp         float64 `json:"timestamp,omitempty"`
}

// GetAlert gets details of a specific alert by UUID.
func (c *Client) GetAlert(ctx context.Context, uuid string) (*Alert, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/alerts/"+url.PathEscape(uuid), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var alert Alert
	if err := json.Unmarshal([]byte(details.ResponseBody), &alert); err != nil {
		return nil, details, err
	}

	return &alert, details, nil
}
