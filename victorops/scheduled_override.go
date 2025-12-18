package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// ScheduledUser represents the user info in a scheduled override response
type ScheduledUser struct {
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

// ScheduledOverride represents a scheduled override in VictorOps
type ScheduledOverride struct {
	PublicID    string                        `json:"publicId,omitempty"`
	User        *ScheduledUser                `json:"user,omitempty"` // API returns nested user object
	Start       string                        `json:"start,omitempty"`
	End         string                        `json:"end,omitempty"`
	Timezone    string                        `json:"timezone,omitempty"`
	Assignments []ScheduledOverrideAssignment `json:"assignments,omitempty"`
}

// GetUsername returns the username from the nested user object
func (s *ScheduledOverride) GetUsername() string {
	if s.User != nil {
		return s.User.Username
	}
	return ""
}

// ScheduledOverrideAssignment represents a policy assignment for a scheduled override
type ScheduledOverrideAssignment struct {
	PolicySlug string `json:"policySlug,omitempty"`
	PolicyName string `json:"policyName,omitempty"`
}

// ScheduledOverrideList represents a list of scheduled overrides
type ScheduledOverrideList struct {
	Overrides []ScheduledOverride `json:"overrides,omitempty"`
	SelfURL   string              `json:"_selfUrl,omitempty"`
}

// ScheduledOverrideResponse represents the response when getting/creating a scheduled override
type ScheduledOverrideResponse struct {
	Override ScheduledOverride `json:"override,omitempty"`
	Schedule ScheduledOverride `json:"schedule,omitempty"`
	SelfURL  string            `json:"_selfUrl,omitempty"`
}

// ScheduledOverridePayload is the payload for creating a scheduled override
type ScheduledOverridePayload struct {
	Username string `json:"username"`
	Start    string `json:"start"`
	End      string `json:"end"`
	Timezone string `json:"timezone"`
}

// Assignment represents a policy assignment for override updates
type Assignment struct {
	Policy   string `json:"policy,omitempty"`
	Username string `json:"username,omitempty"`
}

// ListScheduledOverrides lists all scheduled overrides for the organization
func (c *Client) ListScheduledOverrides(ctx context.Context) (*ScheduledOverrideList, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/overrides", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var overrideList ScheduledOverrideList
	err = json.Unmarshal([]byte(details.ResponseBody), &overrideList)
	if err != nil {
		return nil, details, err
	}

	return &overrideList, details, nil
}

// CreateScheduledOverride creates a new scheduled override
func (c *Client) CreateScheduledOverride(ctx context.Context, override *ScheduledOverridePayload) (*ScheduledOverride, *RequestDetails, error) {
	jsonOverride, err := json.Marshal(override)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/overrides", bytes.NewBuffer(jsonOverride), nil)
	if err != nil {
		return nil, details, err
	}

	var response ScheduledOverrideResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	// The API returns the created override in the "override" field
	// (Note: API spec says "schedule" but actual API uses "override")
	if response.Override.PublicID != "" {
		return &response.Override, details, nil
	}
	// Fallback to "schedule" for spec compatibility
	return &response.Schedule, details, nil
}

// GetScheduledOverride gets a specific scheduled override by public ID
func (c *Client) GetScheduledOverride(ctx context.Context, publicID string) (*ScheduledOverride, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/overrides/"+url.QueryEscape(publicID), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response ScheduledOverrideResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response.Override, details, nil
}

// DeleteScheduledOverride deletes a scheduled override by public ID
func (c *Client) DeleteScheduledOverride(ctx context.Context, publicID string) (*RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "DELETE", "v1/overrides/"+url.QueryEscape(publicID), bytes.NewBufferString("{}"), nil)
	return details, err
}

// GetScheduledOverrideAssignments gets all assignments for a scheduled override
func (c *Client) GetScheduledOverrideAssignments(ctx context.Context, publicID string) ([]Assignment, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/overrides/"+url.QueryEscape(publicID)+"/assignments", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var assignments []Assignment
	err = json.Unmarshal([]byte(details.ResponseBody), &assignments)
	if err != nil {
		return nil, details, err
	}

	return assignments, details, nil
}

// GetScheduledOverrideAssignment gets a specific assignment for a scheduled override
func (c *Client) GetScheduledOverrideAssignment(ctx context.Context, publicID string, policySlug string) (*Assignment, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/overrides/"+url.QueryEscape(publicID)+"/assignments/"+url.QueryEscape(policySlug), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var assignment Assignment
	err = json.Unmarshal([]byte(details.ResponseBody), &assignment)
	if err != nil {
		return nil, details, err
	}

	return &assignment, details, nil
}

// UpdateScheduledOverrideAssignment updates an assignment for a scheduled override
func (c *Client) UpdateScheduledOverrideAssignment(ctx context.Context, publicID string, policySlug string, assignment *Assignment) (*Assignment, *RequestDetails, error) {
	jsonAssignment, err := json.Marshal(assignment)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PUT", "v1/overrides/"+url.QueryEscape(publicID)+"/assignments/"+url.QueryEscape(policySlug), bytes.NewBuffer(jsonAssignment), nil)
	if err != nil {
		return nil, details, err
	}

	var updatedAssignment Assignment
	err = json.Unmarshal([]byte(details.ResponseBody), &updatedAssignment)
	if err != nil {
		return nil, details, err
	}

	return &updatedAssignment, details, nil
}

// DeleteScheduledOverrideAssignment deletes an assignment from a scheduled override
func (c *Client) DeleteScheduledOverrideAssignment(ctx context.Context, publicID string, policySlug string) (*Assignment, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "DELETE", "v1/overrides/"+url.QueryEscape(publicID)+"/assignments/"+url.QueryEscape(policySlug), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var assignment Assignment
	err = json.Unmarshal([]byte(details.ResponseBody), &assignment)
	if err != nil {
		return nil, details, err
	}

	return &assignment, details, nil
}
