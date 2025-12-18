package victorops

import (
	"bytes"
	"context"
	"encoding/json"
)

// Webhook represents a webhook in VictorOps
type Webhook struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
	URL  string `json:"url,omitempty"`
}

// WebhookList represents the response from listing webhooks
type WebhookList struct {
	Webhooks []Webhook `json:"webhooks,omitempty"`
}

// ListWebhooks lists all webhooks for the organization
func (c *Client) ListWebhooks(ctx context.Context) (*WebhookList, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/webhooks", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var webhookList WebhookList
	err = json.Unmarshal([]byte(details.ResponseBody), &webhookList)
	if err != nil {
		return nil, details, err
	}

	return &webhookList, details, nil
}
