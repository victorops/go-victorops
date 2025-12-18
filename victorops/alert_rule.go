package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// AlertRule represents an alert rule returned from the VictorOps API
type AlertRule struct {
	ID              int64             `json:"id,omitempty"`
	AlertField      string            `json:"alertField,omitempty"`
	AlertValueMatch string            `json:"alertValueMatch,omitempty"`
	MatchType       string            `json:"matchType,omitempty"` // WILDCARD or REGEX
	Rank            int               `json:"rank,omitempty"`
	StopFlag        bool              `json:"stopFlag"`
	LastUpdated     int64             `json:"lastUpdated,omitempty"`
	LastUpdatedBy   string            `json:"lastUpdatedBy,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	RouteKey        string            `json:"routeKey,omitempty"` // Note: API returns routeKey, not routingKey
	Annotations     []AlertAnnotation `json:"annotations,omitempty"`
}

// AlertAnnotation represents an annotation/transformation in an alert rule response
type AlertAnnotation struct {
	ID             int64  `json:"id,omitempty"`
	AnnotationType string `json:"annotationType,omitempty"` // i=image, u=url, s=notes
	FieldName      string `json:"fieldName,omitempty"`
	FieldValue     string `json:"fieldValue,omitempty"`
	Flags          int    `json:"flags,omitempty"` // 0=annotation, 1=transformation
}

// AlertRulePayload is the payload for creating/updating an alert rule
type AlertRulePayload struct {
	AlertField      string                   `json:"alertField"`
	AlertValueMatch string                   `json:"alertValueMatch"`
	MatchType       string                   `json:"matchType"` // WILDCARD or REGEX
	StopFlag        bool                     `json:"stopFlag"`
	Notes           string                   `json:"notes,omitempty"`
	Rank            int                      `json:"rank,omitempty"`
	RoutingKey      string                   `json:"routingKey,omitempty"`
	Annotations     []AlertAnnotationPayload `json:"annotations"`
}

// AlertAnnotationPayload is the payload for an annotation in create/update requests
type AlertAnnotationPayload struct {
	AnnotationType string `json:"annotationType"` // i=image, u=url, s=notes
	FieldName      string `json:"fieldName"`
	FieldValue     string `json:"fieldValue"`
	Flags          int    `json:"flags"` // 0=annotation, 1=transformation
}

// DeleteAlertRuleResponse represents the response from deleting an alert rule
type DeleteAlertRuleResponse struct {
	ID int64 `json:"id,omitempty"`
}

// ListAlertRules lists all alert rules for the organization
func (c *Client) ListAlertRules(ctx context.Context) ([]AlertRule, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/alertRules", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var rules []AlertRule
	err = json.Unmarshal([]byte(details.ResponseBody), &rules)
	if err != nil {
		return nil, details, err
	}

	return rules, details, nil
}

// CreateAlertRule creates a new alert rule
func (c *Client) CreateAlertRule(ctx context.Context, rule *AlertRulePayload) (*AlertRule, *RequestDetails, error) {
	jsonRule, err := json.Marshal(rule)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/alertRules", bytes.NewBuffer(jsonRule), nil)
	if err != nil {
		return nil, details, err
	}

	var newRule AlertRule
	err = json.Unmarshal([]byte(details.ResponseBody), &newRule)
	if err != nil {
		return nil, details, err
	}

	return &newRule, details, nil
}

// GetAlertRule gets an alert rule by ID
func (c *Client) GetAlertRule(ctx context.Context, ruleID string) (*AlertRule, *RequestDetails, error) {
	// Note: The API doesn't have a GET by ID endpoint, we need to list and filter
	rules, details, err := c.ListAlertRules(ctx)
	if err != nil {
		return nil, details, err
	}

	for i := range rules {
		if ruleID == strconv.FormatInt(rules[i].ID, 10) {
			return &rules[i], details, nil
		}
	}

	return nil, details, nil
}

// UpdateAlertRule updates an existing alert rule
func (c *Client) UpdateAlertRule(ctx context.Context, ruleID string, rule *AlertRulePayload) (*AlertRule, *RequestDetails, error) {
	jsonRule, err := json.Marshal(rule)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PUT", "v1/alertRules/"+url.QueryEscape(ruleID), bytes.NewBuffer(jsonRule), nil)
	if err != nil {
		return nil, details, err
	}

	var updatedRule AlertRule
	err = json.Unmarshal([]byte(details.ResponseBody), &updatedRule)
	if err != nil {
		return nil, details, err
	}

	return &updatedRule, details, nil
}

// DeleteAlertRule deletes an alert rule
func (c *Client) DeleteAlertRule(ctx context.Context, ruleID string) (*DeleteAlertRuleResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "DELETE", "v1/alertRules/"+url.QueryEscape(ruleID), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response DeleteAlertRuleResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}
