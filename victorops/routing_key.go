package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// RoutingKey is a struct to hold the data for a victorops routing key.
type RoutingKey struct {
	RoutingKey       string   `json:"routingKey,omitempty"`
	Targets          []string `json:"targets,omitempty"`
	IsMultiResponder bool     `json:"isMultiResponder,omitempty"`
}

// RoutingKeyUpdatePayload is the request body for updating a routing key. The
// routingKey value replaces the key named in the path; targets is the list of
// escalation policy slugs the key targets.
type RoutingKeyUpdatePayload struct {
	RoutingKey       string   `json:"routingKey"`
	Targets          []string `json:"targets"`
	IsMultiResponder bool     `json:"isMultiResponder,omitempty"`
}

func parseRoutingKeyResponse(response string) (*RoutingKey, error) {
	// Parse the response and return the user object
	var rk RoutingKey
	err := json.Unmarshal([]byte(response), &rk)
	if err != nil {
		return nil, err
	}

	return &rk, err
}

// In the request to create a routing key, we supply a list of strings which are
// target escalaption policy slugs
// In the response while getting routing keys, we get a different type of result which is
// a map of values like so: {"policyName":"Moderate Severity","policySlug":"pol-tq09wTVkG7BzuMY0","_teamUrl":"/api-public/v1/team/team-Iei67wjVsD14Pe4O"}
// So these structs exist to represent read responses rather that create requests.
type RoutingKeyResponse struct {
	RoutingKey       string                      `json:"routingKey,omitempty"`
	Targets          []RoutingKeyResponseTargets `json:"targets,omitempty"`
	IsDefault        bool                        `json:"isDefault,omitempty"`
	IsMultiResponder bool                        `json:"isMultiResponder,omitempty"`
}
type RoutingKeyResponseList struct {
	RoutingKeys []RoutingKeyResponse `json:"routingKeys,omitempty"`
}
type RoutingKeyResponseTargets struct {
	PolicySlug string `json:"policySlug,omitempty"`
}

func parseRoutingKeyListResponse(response string) (*RoutingKeyResponseList, error) {
	// Parse the response and return the user object
	var rkrl RoutingKeyResponseList
	err := json.Unmarshal([]byte(response), &rkrl)
	if err != nil {
		return nil, err
	}

	return &rkrl, err
}

// CreateRoutingKey creates a routingkey in the victorops organization
func (c *Client) CreateRoutingKey(ctx context.Context, routingKey *RoutingKey) (*RoutingKey, *RequestDetails, error) {
	jsonRk, err := json.Marshal(routingKey)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall(ctx, "POST", "v1/org/routing-keys", bytes.NewBuffer(jsonRk), nil)
	if err != nil {
		return nil, details, err
	}

	newKey, err := parseRoutingKeyResponse(details.ResponseBody)
	if err != nil {
		return newKey, details, err
	}

	return newKey, details, nil
}

// GetRoutingKey returns a specific routingkey within this victorops organization
func (c *Client) GetRoutingKey(ctx context.Context, keyname string) (*RoutingKeyResponse, *RequestDetails, error) {

	rkList, details, err := c.GetAllRoutingKeys(ctx)
	// Check for errors
	if err != nil {
		return nil, details, err
	}

	for _, key := range rkList.RoutingKeys {
		if key.RoutingKey == keyname {
			return &key, details, err
		}
	}

	return nil, details, nil
}

// UpdateRoutingKey updates an existing routing key's target escalation policies.
// The default routing key cannot be updated.
func (c *Client) UpdateRoutingKey(ctx context.Context, routingKey string, payload *RoutingKeyUpdatePayload) (*RoutingKey, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PUT", "v1/org/routing-keys/"+url.PathEscape(routingKey), bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	updated, err := parseRoutingKeyResponse(details.ResponseBody)
	if err != nil {
		return updated, details, err
	}

	return updated, details, nil
}

// DeleteRoutingKey deletes a routing key by its name. The default routing key cannot be deleted.
func (c *Client) DeleteRoutingKey(ctx context.Context, routingKey string) (*RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "DELETE", "v1/org/routing-keys/"+url.PathEscape(routingKey), bytes.NewBufferString("{}"), nil)
	return details, err
}

// GetAllRoutingKeys returns a list of all of the routing keys for an account
func (c *Client) GetAllRoutingKeys(ctx context.Context) (*RoutingKeyResponseList, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall(ctx, "GET", "v1/org/routing-keys", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	rkList, err := parseRoutingKeyListResponse(details.ResponseBody)
	if err != nil {
		return nil, details, err
	}

	return rkList, details, nil
}
