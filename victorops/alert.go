package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// Alert represents an alert in VictorOps
type Alert struct {
	UUID              string                 `json:"uuid,omitempty"`
	MessageType       string                 `json:"messageType,omitempty"`
	EntityID          string                 `json:"entityId,omitempty"`
	EntityDisplayName string                 `json:"entityDisplayName,omitempty"`
	StateMessage      string                 `json:"stateMessage,omitempty"`
	RoutingKey        string                 `json:"routingKey,omitempty"`
	Timestamp         string                 `json:"timestamp,omitempty"`
	RawPayload        map[string]interface{} `json:"raw,omitempty"`
}

// GetAlertResponse represents the response from getting an alert
type GetAlertResponse struct {
	Alert Alert `json:"alert,omitempty"`
}

// GetAlert gets details of a specific alert by UUID
func (c *Client) GetAlert(ctx context.Context, uuid string) (*GetAlertResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/alerts/"+url.QueryEscape(uuid), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response GetAlertResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		// Try direct unmarshal to Alert
		var alert Alert
		err = json.Unmarshal([]byte(details.ResponseBody), &alert)
		if err != nil {
			return nil, details, err
		}
		response.Alert = alert
	}

	return &response, details, nil
}
