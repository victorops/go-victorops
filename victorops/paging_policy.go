package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// PagingPolicyStep represents a step in a paging policy
type PagingPolicyStep struct {
	Step    int                `json:"step,omitempty"`
	Timeout int                `json:"timeout,omitempty"`
	Rules   []PagingPolicyRule `json:"rules,omitempty"`
}

// PagingPolicyRule represents a rule within a paging policy step
type PagingPolicyRule struct {
	Rule             int    `json:"rule,omitempty"`
	NotificationType string `json:"notificationType,omitempty"`
	ContactType      string `json:"contactType,omitempty"`
	ContactID        int    `json:"contactId,omitempty"`
	ContactLabel     string `json:"contactLabel,omitempty"`
}

// UserPagingPolicy represents a user's paging policy (v2)
type UserPagingPolicy struct {
	PolicyType string             `json:"policyType,omitempty"`
	Steps      []PagingPolicyStep `json:"steps,omitempty"`
}

// UserPagingPoliciesResponse represents the response from v2 policies endpoint
type UserPagingPoliciesResponse struct {
	Policies []UserPagingPolicy `json:"policies,omitempty"`
}

// PagingPolicyStepsResponse represents the response from getting paging policy steps
type PagingPolicyStepsResponse struct {
	Steps   []PagingPolicyStep `json:"steps,omitempty"`
	SelfURL string             `json:"_selfUrl,omitempty"`
}

// PagingPolicyStepResponse represents the response for a single step
type PagingPolicyStepResponse struct {
	Step    PagingPolicyStep `json:"step,omitempty"`
	SelfURL string           `json:"_selfUrl,omitempty"`
}

// PagingPolicyRuleResponse represents the response for a single rule
type PagingPolicyRuleResponse struct {
	StepRule PagingPolicyRule `json:"stepRule,omitempty"`
	SelfURL  string           `json:"_selfUrl,omitempty"`
}

// NotificationType represents a notification type available for paging policies
type NotificationType struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// NotificationTypesResponse represents the response from getting notification types
type NotificationTypesResponse struct {
	NotificationTypes []NotificationType `json:"notificationTypes,omitempty"`
	SelfURL           string             `json:"_selfUrl,omitempty"`
}

// PagingContactType represents a contact type available for paging policies
type PagingContactType struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// ContactTypesResponse represents the response from getting contact types
type ContactTypesResponse struct {
	ContactTypes []PagingContactType `json:"contactTypes,omitempty"`
	SelfURL      string              `json:"_selfUrl,omitempty"`
}

// TimeoutType represents a timeout type available for paging policies
type TimeoutType struct {
	Description string `json:"description,omitempty"`
	Type        int    `json:"type,omitempty"`
}

// TimeoutTypesResponse represents the response from getting timeout types
type TimeoutTypesResponse struct {
	TimeoutTypes []TimeoutType `json:"timeoutTypes,omitempty"`
	SelfURL      string        `json:"_selfUrl,omitempty"`
}

// AddStepPayload represents the payload for adding a paging policy step
type AddStepPayload struct {
	Timeout int `json:"timeout,omitempty"`
}

// AddRulePayload represents the payload for adding a rule to a step
type AddRulePayload struct {
	NotificationType string `json:"notificationType"`
	ContactID        int    `json:"contactId"`
}

// GetNotificationTypes gets the available notification types for paging policies
func (c *Client) GetNotificationTypes(ctx context.Context) (*NotificationTypesResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/policies/types/notifications", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response NotificationTypesResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetContactTypes gets the available contact types for paging policies
func (c *Client) GetPagingContactTypes(ctx context.Context) (*ContactTypesResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/policies/types/contacts", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response ContactTypesResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetTimeoutTypes gets the available timeout types for paging policies
func (c *Client) GetTimeoutTypes(ctx context.Context) (*TimeoutTypesResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/policies/types/timeouts", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response TimeoutTypesResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetUserPagingPolicies gets all paging policies for a user (v1 primary policy)
func (c *Client) GetUserPagingPolicies(ctx context.Context, username string) (*PagingPolicyStepsResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/profile/"+url.QueryEscape(username)+"/policies", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyStepsResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetUserPagingPoliciesV2 gets all paging policies for a user (v2 - includes all policy types)
func (c *Client) GetUserPagingPoliciesV2(ctx context.Context, username string) (*UserPagingPoliciesResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v2/profile/"+url.QueryEscape(username)+"/policies", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response UserPagingPoliciesResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// CreatePagingPolicyStep creates a new paging policy step for a user
func (c *Client) CreatePagingPolicyStep(ctx context.Context, username string, timeout int) (*PagingPolicyStepResponse, *RequestDetails, error) {
	payload := AddStepPayload{Timeout: timeout}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/profile/"+url.QueryEscape(username)+"/policies", bytes.NewBuffer(jsonPayload), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyStepResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetPagingPolicyStep gets a specific paging policy step
func (c *Client) GetPagingPolicyStep(ctx context.Context, username string, step int) (*PagingPolicyStepResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", fmt.Sprintf("v1/profile/%s/policies/%d", url.QueryEscape(username), step), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyStepResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// UpdatePagingPolicyStep updates a paging policy step
func (c *Client) UpdatePagingPolicyStep(ctx context.Context, username string, step int, timeout int) (*PagingPolicyStepResponse, *RequestDetails, error) {
	payload := AddStepPayload{Timeout: timeout}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PUT", fmt.Sprintf("v1/profile/%s/policies/%d", url.QueryEscape(username), step), bytes.NewBuffer(jsonPayload), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyStepResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// CreatePagingPolicyRule creates a new rule in a paging policy step
func (c *Client) CreatePagingPolicyRule(ctx context.Context, username string, step int, notificationType string, contactID int) (*PagingPolicyRuleResponse, *RequestDetails, error) {
	payload := AddRulePayload{
		NotificationType: notificationType,
		ContactID:        contactID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", fmt.Sprintf("v1/profile/%s/policies/%d", url.QueryEscape(username), step), bytes.NewBuffer(jsonPayload), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyRuleResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetPagingPolicyRule gets a specific rule from a paging policy step
func (c *Client) GetPagingPolicyRule(ctx context.Context, username string, step int, rule int) (*PagingPolicyRuleResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", fmt.Sprintf("v1/profile/%s/policies/%d/%d", url.QueryEscape(username), step, rule), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyRuleResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// UpdatePagingPolicyRule updates a rule in a paging policy step
func (c *Client) UpdatePagingPolicyRule(ctx context.Context, username string, step int, rule int, notificationType string, contactID int) (*PagingPolicyRuleResponse, *RequestDetails, error) {
	payload := AddRulePayload{
		NotificationType: notificationType,
		ContactID:        contactID,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PUT", fmt.Sprintf("v1/profile/%s/policies/%d/%d", url.QueryEscape(username), step, rule), bytes.NewBuffer(jsonPayload), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyRuleResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// DeletePagingPolicyRule deletes a rule from a paging policy step
func (c *Client) DeletePagingPolicyRule(ctx context.Context, username string, step int, rule int) (*PagingPolicyRuleResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "DELETE", fmt.Sprintf("v1/profile/%s/policies/%d/%d", url.QueryEscape(username), step, rule), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response PagingPolicyRuleResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}
