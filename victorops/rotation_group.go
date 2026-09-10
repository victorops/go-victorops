package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// RotationCreateQueryOptions contains the optional paging controls accepted by
// the rotation-group create endpoint.
type RotationCreateQueryOptions struct {
	Skip  *int
	Count *int
}

func (o *RotationCreateQueryOptions) queryParams() map[string]string {
	if o == nil {
		return nil
	}
	params := make(map[string]string, 2)
	if o.Skip != nil {
		params["skip"] = strconv.Itoa(*o.Skip)
	}
	if o.Count != nil {
		params["count"] = strconv.Itoa(*o.Count)
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

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
	// ShiftMembers is retained for gateways that accept the member list under
	// this legacy field when the Joda-date request format is required.
	ShiftMembers []string `json:"shiftMembers,omitempty"`
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
	// Slug is resolved from the v1 list endpoint because create responses do not
	// consistently include it.
	Slug string `json:"-"`
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

// CreateRotationGroup creates a new rotation group under a team, optionally
// with shifts. It preserves the established create workflow while exposing the
// simpler method used by callers that do not need paging query options.
func (c *Client) CreateRotationGroup(ctx context.Context, teamSlug string, payload *RotationGroupCreatePayload) (*RotationGroupResponse, *RequestDetails, error) {
	return c.CreateRotation(ctx, teamSlug, payload, nil)
}

// CreateRotation creates a rotation group and resolves the numeric group ID and
// slug needed by stateful consumers. Some public API gateways accept epoch
// milliseconds for shift start while others require an ISO-8601 Joda date; a
// validation failure for the former is retried with the latter representation.
func (c *Client) CreateRotation(ctx context.Context, teamSlug string, payload *RotationGroupCreatePayload, options *RotationCreateQueryOptions) (*RotationGroupResponse, *RequestDetails, error) {
	if payload == nil {
		return nil, nil, fmt.Errorf("rotation create payload cannot be nil")
	}

	before, beforeDetails, err := c.ListRotationsV1(ctx, teamSlug)
	if err != nil {
		return nil, beforeDetails, err
	}
	beforeIDs := make(map[int64]struct{}, len(before.RotationGroups))
	for _, group := range before.RotationGroups {
		beforeIDs[group.GroupID] = struct{}{}
	}

	created, details, err := c.createRotationRequest(ctx, teamSlug, payload, options.queryParams(), false)
	if isJodaDateValidationError(details, err) {
		created, details, err = c.createRotationRequest(ctx, teamSlug, payload, options.queryParams(), true)
	}
	if err != nil {
		return created, details, err
	}

	groupID := created.ID
	if groupID == 0 && len(created.Shifts) > 0 {
		groupID = created.Shifts[0].GroupID
	}

	resolveLabel := created.Label
	if resolveLabel == "" {
		resolveLabel = payload.Label
	}
	resolved, resolveErr := c.resolveCreatedRotation(ctx, teamSlug, groupID, resolveLabel, beforeIDs)
	if resolveErr == nil {
		created.ID = resolved.GroupID
		created.Slug = resolved.Slug
		if created.Label == "" {
			created.Label = resolved.Label
		}
	} else {
		return created, details, resolveErr
	}

	return created, details, nil
}

func (c *Client) createRotationRequest(ctx context.Context, teamSlug string, payload *RotationGroupCreatePayload, queryParams map[string]string, jodaDate bool) (*RotationGroupResponse, *RequestDetails, error) {
	body, err := marshalRotationGroupCreatePayload(payload, jodaDate)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", rotationsBase(teamSlug), bytes.NewReader(body), queryParams)
	if err != nil {
		return nil, details, err
	}

	response := &RotationGroupResponse{}
	if strings.TrimSpace(details.ResponseBody) != "" {
		if err := json.Unmarshal([]byte(details.ResponseBody), response); err != nil {
			return nil, details, err
		}
	}
	return response, details, nil
}

func marshalRotationGroupCreatePayload(payload *RotationGroupCreatePayload, jodaDate bool) ([]byte, error) {
	if !jodaDate {
		return json.Marshal(payload)
	}

	result := map[string]interface{}{"label": payload.Label}
	if len(payload.Shifts) > 0 {
		shifts := make([]map[string]interface{}, len(payload.Shifts))
		for i := range payload.Shifts {
			shifts[i] = rotationShiftPayloadMap(&payload.Shifts[i], true)
		}
		result["shifts"] = shifts
	}
	return json.Marshal(result)
}

func rotationShiftPayloadMap(payload *RotationShiftCreatePayload, jodaDate bool) map[string]interface{} {
	shift := map[string]interface{}{
		"label":     payload.Label,
		"timezone":  payload.Timezone,
		"duration":  payload.Duration,
		"shifttype": payload.ShiftType,
	}
	if payload.Start != 0 {
		if jodaDate {
			shift["start"] = time.UnixMilli(payload.Start).UTC().Format("2006-01-02T15:04:05.000Z")
		} else {
			shift["start"] = payload.Start
		}
	}
	if payload.Mask != nil {
		shift["mask"] = payload.Mask
	}
	if payload.Mask2 != nil {
		shift["mask2"] = payload.Mask2
	}
	if payload.Mask3 != nil {
		shift["mask3"] = payload.Mask3
	}
	if len(payload.Usernames) > 0 {
		shift["usernames"] = payload.Usernames
		shift["shiftMembers"] = payload.Usernames
	}
	if len(payload.ShiftMembers) > 0 {
		shift["shiftMembers"] = payload.ShiftMembers
	}
	return shift
}

func isJodaDateValidationError(details *RequestDetails, err error) bool {
	if details == nil || !strings.Contains(details.ResponseBody, "error.expected.jodadate.format") {
		return false
	}
	var apiErr *APIError
	return details.StatusCode == 400 && (err == nil || errors.As(err, &apiErr))
}

func (c *Client) resolveCreatedRotation(ctx context.Context, teamSlug string, groupID int64, label string, beforeIDs map[int64]struct{}) (*RotationGroup, error) {
	const maxAttempts = 8
	const delay = 500 * time.Millisecond

	for attempt := 0; attempt < maxAttempts; attempt++ {
		groups, _, err := c.ListRotationsV1(ctx, teamSlug)
		if err != nil {
			return nil, err
		}

		for i := range groups.RotationGroups {
			group := &groups.RotationGroups[i]
			if groupID != 0 && group.GroupID == groupID && group.Slug != "" {
				return group, nil
			}
		}

		var candidates []*RotationGroup
		for i := range groups.RotationGroups {
			group := &groups.RotationGroups[i]
			_, existed := beforeIDs[group.GroupID]
			if !existed && (label == "" || group.Label == label) {
				candidates = append(candidates, group)
			}
		}
		if len(candidates) == 1 && candidates[0].Slug != "" {
			return candidates[0], nil
		}

		if attempt < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return nil, fmt.Errorf("rotation was created but its group id/slug could not be resolved for team %s", teamSlug)
}

// GetRotationGroup gets a single rotation group by ID.
func (c *Client) GetRotationGroup(ctx context.Context, teamSlug string, groupID int64) (*RotationGroupResponse, *RequestDetails, error) {
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
func (c *Client) UpdateRotationGroup(ctx context.Context, teamSlug string, groupID int64, payload *RotationGroupUpdatePayload) (*RotationGroupResponse, *RequestDetails, error) {
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
func (c *Client) DeleteRotationGroup(ctx context.Context, teamSlug string, groupID int64) (*RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d", rotationsBase(teamSlug), groupID)
	details, err := c.makePublicAPICall(ctx, "DELETE", endpoint, bytes.NewBufferString("{}"), nil)
	return details, err
}

// DeleteRotation is the established numeric-ID deletion entry point.
func (c *Client) DeleteRotation(ctx context.Context, teamSlug string, groupID int64) (*RequestDetails, error) {
	return c.DeleteRotationGroup(ctx, teamSlug, groupID)
}

// CreateRotationShift adds a new shift to an existing rotation group.
func (c *Client) CreateRotationShift(ctx context.Context, teamSlug string, groupID int64, payload *RotationShiftCreatePayload) (*RotationShiftResource, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("%s/%d", rotationsBase(teamSlug), groupID)
	details, err := c.makePublicAPICall(ctx, "POST", endpoint, bytes.NewBuffer(body), nil)
	if isJodaDateValidationError(details, err) {
		body, marshalErr := json.Marshal(rotationShiftPayloadMap(payload, true))
		if marshalErr != nil {
			return nil, nil, marshalErr
		}
		details, err = c.makePublicAPICall(ctx, "POST", endpoint, bytes.NewBuffer(body), nil)
	}
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
func (c *Client) GetRotationShift(ctx context.Context, teamSlug string, groupID int64, shiftID int64) (*RotationShiftResource, *RequestDetails, error) {
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
func (c *Client) UpdateRotationShift(ctx context.Context, teamSlug string, groupID int64, shiftID int64, payload *RotationShiftCreatePayload) (*RotationShiftResource, *RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	if isJodaDateValidationError(details, err) {
		body, marshalErr := json.Marshal(rotationShiftPayloadMap(payload, true))
		if marshalErr != nil {
			return nil, nil, marshalErr
		}
		details, err = c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	}
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
func (c *Client) DeleteRotationShift(ctx context.Context, teamSlug string, groupID int64, shiftID int64) (*RequestDetails, error) {
	endpoint := fmt.Sprintf("%s/%d/%d", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "DELETE", endpoint, bytes.NewBufferString("{}"), nil)
	return details, err
}

// AddRotationShiftMember adds a member to a rotation shift.
func (c *Client) AddRotationShiftMember(ctx context.Context, teamSlug string, groupID int64, shiftID int64, payload *RotationMemberAddPayload) (*RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/members", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "POST", endpoint, bytes.NewBuffer(body), nil)
	return details, err
}

// RemoveRotationShiftMember removes a member from a rotation shift, optionally naming a replacement.
func (c *Client) RemoveRotationShiftMember(ctx context.Context, teamSlug string, groupID int64, shiftID int64, payload *RotationMemberRemovePayload) (*RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/members", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "DELETE", endpoint, bytes.NewBuffer(body), nil)
	return details, err
}

// UpdateRotationShiftMemberPosition updates a member's position within a rotation shift.
func (c *Client) UpdateRotationShiftMemberPosition(ctx context.Context, teamSlug string, groupID int64, shiftID int64, payload *RotationMemberPositionPayload) (*RequestDetails, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%d/%d/members", rotationsBase(teamSlug), groupID, shiftID)
	details, err := c.makePublicAPICall(ctx, "PUT", endpoint, bytes.NewBuffer(body), nil)
	return details, err
}

// GetScheduledShiftUser gets the currently scheduled user for a rotation shift.
func (c *Client) GetScheduledShiftUser(ctx context.Context, teamSlug string, groupID int64, shiftID int64) (*ScheduledShiftResponse, *RequestDetails, error) {
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
func (c *Client) SetScheduledShiftUser(ctx context.Context, teamSlug string, groupID int64, shiftID int64, payload *SetScheduledShiftPayload) (*ScheduledShiftResponse, *RequestDetails, error) {
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
