package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// RotationGroupCreatePayload is the request body for creating a rotation group.
type RotationGroupCreatePayload struct {
	Label  string                       `json:"label"`
	Shifts []RotationShiftCreatePayload `json:"shifts,omitempty"`
}

// RotationGroupUpdatePayload is the request body for updating a rotation group's label.
type RotationGroupUpdatePayload struct {
	Label string `json:"label"`
}

// RotationShiftCreatePayload is the request body for creating (inline or standalone)
// or fully replacing a rotation shift. shifttype must be one of std, fts, cstm, pho.
// Usernames is only honored when creating shifts inline via CreateRotationGroup; it is
// ignored by CreateRotationShift/UpdateRotationShift (use the shift member endpoints).
type RotationShiftCreatePayload struct {
	Label     string             `json:"label"`
	Timezone  string             `json:"timezone"`
	Start     int64              `json:"start"` // epoch milliseconds
	Duration  int                `json:"duration"`
	ShiftType string             `json:"shifttype"`
	Mask      *RotationGroupMask `json:"mask,omitempty"`
	Mask2     *RotationGroupMask `json:"mask2,omitempty"`
	Mask3     *RotationGroupMask `json:"mask3,omitempty"`
	Usernames []string           `json:"usernames,omitempty"`
}

// RotationGroupMask represents a restriction mask (day selection + time ranges) for a shift.
type RotationGroupMask struct {
	Day  *RotationDayMask `json:"day,omitempty"`
	Time []MaskTimeRange  `json:"time,omitempty"`
}

// RotationDayMask selects which days of the week the mask applies to.
type RotationDayMask struct {
	Su bool `json:"su"`
	M  bool `json:"m"`
	T  bool `json:"t"`
	W  bool `json:"w"`
	Th bool `json:"th"`
	F  bool `json:"f"`
	Sa bool `json:"sa"`
}

// MaskTimeRange is a start/end time range within a day mask.
type MaskTimeRange struct {
	Start *MaskTime `json:"start,omitempty"`
	End   *MaskTime `json:"end,omitempty"`
}

// MaskTime is an hour/minute pair used in a MaskTimeRange.
type MaskTime struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// RotationGroupResponse is returned by rotation group create/get/update.
type RotationGroupResponse struct {
	OrgSlug  string                  `json:"orgslug,omitempty"`
	TeamSlug string                  `json:"teamslug,omitempty"`
	ID       int64                   `json:"id,omitempty"`
	Label    string                  `json:"label,omitempty"`
	Shifts   []RotationShiftResource `json:"shifts,omitempty"`
}

// RotationShiftResource is returned by shift create/get/update.
type RotationShiftResource struct {
	OrgSlug      string                `json:"org_slug,omitempty"`
	TeamSlug     string                `json:"team_slug,omitempty"`
	RotID        int64                 `json:"rot_id,omitempty"`
	GroupID      int64                 `json:"group_id,omitempty"`
	Label        string                `json:"label,omitempty"`
	Timezone     string                `json:"timezone,omitempty"`
	Start        int64                 `json:"start,omitempty"`
	Duration     int                   `json:"duration,omitempty"`
	ShiftType    string                `json:"shifttype,omitempty"`
	Mask         *RotationGroupMask    `json:"mask,omitempty"`
	Mask2        *RotationGroupMask    `json:"mask2,omitempty"`
	Mask3        *RotationGroupMask    `json:"mask3,omitempty"`
	Periods      []RotationShiftPeriod `json:"periods,omitempty"`
	OnCall       string                `json:"oncall,omitempty"`
	Current      *RotationShiftPeriod  `json:"current,omitempty"`
	Next         *RotationShiftPeriod  `json:"next,omitempty"`
	Usernames    []string              `json:"usernames,omitempty"`
	ShiftMembers []ShiftMember         `json:"shiftMembers,omitempty"`
}

// RotationShiftPeriod is a single on-call period within a shift.
type RotationShiftPeriod struct {
	Start      int64  `json:"start,omitempty"`
	End        int64  `json:"end,omitempty"`
	Username   string `json:"username,omitempty"`
	IsRoll     bool   `json:"isRoll,omitempty"`
	MemberSlug string `json:"memberSlug,omitempty"`
}

// ShiftMember is a member of a rotation shift.
type ShiftMember struct {
	Slug     string `json:"slug,omitempty"`
	Username string `json:"username,omitempty"`
}

// RotationMemberAddPayload is the request body for adding a member to a shift.
type RotationMemberAddPayload struct {
	Username string `json:"username"`
	Position *int   `json:"position,omitempty"`
}

// RotationMemberRemovePayload is the request body for removing a member from a shift.
type RotationMemberRemovePayload struct {
	Username    string `json:"username"`
	Replacement string `json:"replacement,omitempty"`
	MemberSlug  string `json:"memberSlug,omitempty"`
}

// RotationMemberPositionPayload is the request body for repositioning a shift member.
type RotationMemberPositionPayload struct {
	Username   string `json:"username"`
	Position   int    `json:"position"`
	MemberSlug string `json:"memberSlug,omitempty"`
}

// ScheduledShiftResponse is the currently scheduled user for a shift.
type ScheduledShiftResponse struct {
	Username    string `json:"username,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Version     int    `json:"version,omitempty"`
	Verified    bool   `json:"verified,omitempty"`
}

// SetScheduledShiftPayload is the request body for setting the scheduled user for a shift.
type SetScheduledShiftPayload struct {
	Username   string `json:"username"`
	Position   *int   `json:"position,omitempty"`
	MemberSlug string `json:"memberSlug,omitempty"`
}

func rotationsBase(teamSlug string) string {
	return "v1/teams/" + url.PathEscape(teamSlug) + "/rotations"
}

// CreateRotationGroup creates a new rotation group under a team, optionally with shifts.
func (c *Client) CreateRotationGroup(ctx context.Context, teamSlug string, payload *RotationGroupCreatePayload) (*RotationGroupResponse, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", rotationsBase(teamSlug), bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	var response RotationGroupResponse
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetRotationGroup gets a single rotation group by ID.
func (c *Client) GetRotationGroup(ctx context.Context, teamSlug string, groupID int) (*RotationGroupResponse, *RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d", rotationsBase(teamSlug), groupID)
	details, err := c.makePublicAPICall(ctx, "GET", endpoint, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response RotationGroupResponse
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// UpdateRotationGroup updates a rotation group's label.
func (c *Client) UpdateRotationGroup(ctx context.Context, teamSlug string, groupID int, payload *RotationGroupUpdatePayload) (*RotationGroupResponse, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("%s/%d", rotationsBase(teamSlug), groupID)
	details, err := c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	var response RotationGroupResponse
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// DeleteRotationGroup deletes a rotation group by ID.
func (c *Client) DeleteRotationGroup(ctx context.Context, teamSlug string, groupID int) (*RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d", rotationsBase(teamSlug), groupID)
	details, err := c.makePublicAPICall(ctx, "DELETE", endpoint, bytes.NewBufferString("{}"), nil)
	return details, err
}

// CreateRotationShift adds a new shift to an existing rotation group.
func (c *Client) CreateRotationShift(ctx context.Context, teamSlug string, groupID int, payload *RotationShiftCreatePayload) (*RotationShiftResource, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("%s/%d", rotationsBase(teamSlug), groupID)
	details, err := c.makePublicAPICall(ctx, "POST", endpoint, bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	var response RotationShiftResource
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetRotationShift gets a single shift within a rotation group.
func (c *Client) GetRotationShift(ctx context.Context, teamSlug string, groupID int, shiftID int) (*RotationShiftResource, *RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d/%d", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "GET", endpoint, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response RotationShiftResource
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// UpdateRotationShift fully replaces an existing shift within a rotation group.
func (c *Client) UpdateRotationShift(ctx context.Context, teamSlug string, groupID int, shiftID int, payload *RotationShiftCreatePayload) (*RotationShiftResource, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	var response RotationShiftResource
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// DeleteRotationShift deletes a shift within a rotation group.
func (c *Client) DeleteRotationShift(ctx context.Context, teamSlug string, groupID int, shiftID int) (*RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d/%d", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "DELETE", endpoint, bytes.NewBufferString("{}"), nil)
	return details, err
}

// AddRotationShiftMember adds a member to a rotation shift.
func (c *Client) AddRotationShiftMember(ctx context.Context, teamSlug string, groupID int, shiftID int, payload *RotationMemberAddPayload) (*RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/members", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "POST", endpoint, bytes.NewBuffer(body), nil)
	return details, err
}

// RemoveRotationShiftMember removes a member from a rotation shift, optionally naming a replacement.
func (c *Client) RemoveRotationShiftMember(ctx context.Context, teamSlug string, groupID int, shiftID int, payload *RotationMemberRemovePayload) (*RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/members", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "DELETE", endpoint, bytes.NewBuffer(body), nil)
	return details, err
}

// UpdateRotationShiftMemberPosition updates a member's position within a rotation shift.
func (c *Client) UpdateRotationShiftMemberPosition(ctx context.Context, teamSlug string, groupID int, shiftID int, payload *RotationMemberPositionPayload) (*RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/members", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	return details, err
}

// GetScheduledShiftUser gets the currently scheduled user for a rotation shift.
func (c *Client) GetScheduledShiftUser(ctx context.Context, teamSlug string, groupID int, shiftID int) (*ScheduledShiftResponse, *RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d/%d/scheduled", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "GET", endpoint, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response ScheduledShiftResponse
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// SetScheduledShiftUser sets the scheduled user for a rotation shift.
func (c *Client) SetScheduledShiftUser(ctx context.Context, teamSlug string, groupID int, shiftID int, payload *SetScheduledShiftPayload) (*ScheduledShiftResponse, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/scheduled", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	var response ScheduledShiftResponse
	if err := json.Unmarshal([]byte(details.ResponseBody), &response); err != nil {
		return nil, details, err
	}

	return &response, details, nil
}
