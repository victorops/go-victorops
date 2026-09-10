package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// RotationGroup represents a rotation group in v1 API
type RotationGroup struct {
	TeamSlug string `json:"teamSlug,omitempty"`
	Slug     string `json:"slug,omitempty"`
	Label    string `json:"label,omitempty"`
	GroupID  int64  `json:"groupId,omitempty"`
}

// RotationGroupList represents the response from listing rotation groups (v1)
type RotationGroupList struct {
	RotationGroups []RotationGroup `json:"rotationGroups,omitempty"`
}

// Rotation represents a rotation group returned by the v2 rotations API.
type Rotation struct {
	GroupID                int64                  `json:"groupId,omitempty"`
	Label                  string                 `json:"label,omitempty"`
	TotalMembersInRotation int                    `json:"totalMembersInRotation,omitempty"`
	Shifts                 []RotationShiftDetails `json:"shifts,omitempty"`
}

// RotationShiftDetails represents a shift returned by the v2 rotations API.
type RotationShiftDetails struct {
	ShiftID      int64                  `json:"shiftId,omitempty"`
	Label        string                 `json:"label,omitempty"`
	Duration     int                    `json:"duration,omitempty"`
	Current      *RotationOnCallPeriod  `json:"current,omitempty"`
	Next         *RotationOnCallPeriod  `json:"next,omitempty"`
	Periods      []RotationOnCallPeriod `json:"periods,omitempty"`
	ShiftMembers []ShiftMember          `json:"shiftMembers,omitempty"`
	ShiftType    string                 `json:"shifttype,omitempty"`
	Start        string                 `json:"start,omitempty"`
	Timezone     string                 `json:"timezone,omitempty"`
	Mask         *RotationGroupMask     `json:"mask,omitempty"`
	Mask2        *RotationGroupMask     `json:"mask2,omitempty"`
	Mask3        *RotationGroupMask     `json:"mask3,omitempty"`
}

// RotationOnCallPeriod represents an on-call period returned by the v2 API.
// apppublic serializes its timestamps as ISO-8601 strings.
type RotationOnCallPeriod struct {
	Start      string `json:"start,omitempty"`
	End        string `json:"end,omitempty"`
	Username   string `json:"username,omitempty"`
	IsRoll     bool   `json:"isRoll"`
	MemberSlug string `json:"memberSlug,omitempty"`
}

// RotationList represents the response from listing rotations (v2)
type RotationList struct {
	Rotations []Rotation `json:"rotations,omitempty"`
}

// ListRotationsV1 lists all rotation groups for a team (v1 API)
func (c *Client) ListRotationsV1(ctx context.Context, teamSlug string) (*RotationGroupList, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/teams/"+url.PathEscape(teamSlug)+"/rotations", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var rotationList RotationGroupList
	err = json.Unmarshal([]byte(details.ResponseBody), &rotationList)
	if err != nil {
		return nil, details, err
	}

	return &rotationList, details, nil
}

// GetRotationGroupByGroupID finds a v1 rotation group by numeric group ID.
func (c *Client) GetRotationGroupByGroupID(ctx context.Context, teamSlug string, groupID int64) (*RotationGroup, *RequestDetails, error) {
	groups, details, err := c.ListRotationsV1(ctx, teamSlug)
	if err != nil {
		return nil, details, err
	}

	for i := range groups.RotationGroups {
		if groups.RotationGroups[i].GroupID == groupID {
			return &groups.RotationGroups[i], details, nil
		}
	}

	return nil, details, nil
}

// ListRotationsV2 lists all rotations with details for a team (v2 API)
func (c *Client) ListRotationsV2(ctx context.Context, teamSlug string) (*RotationList, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v2/team/"+url.PathEscape(teamSlug)+"/rotations", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var rotationList RotationList
	err = json.Unmarshal([]byte(details.ResponseBody), &rotationList)
	if err != nil {
		return nil, details, err
	}

	return &rotationList, details, nil
}

// GetRotationByGroupID finds a v2 rotation by numeric group ID.
func (c *Client) GetRotationByGroupID(ctx context.Context, teamSlug string, groupID int64) (*Rotation, *RequestDetails, error) {
	rotations, details, err := c.ListRotationsV2(ctx, teamSlug)
	if err != nil {
		return nil, details, err
	}

	for i := range rotations.Rotations {
		if rotations.Rotations[i].GroupID == groupID {
			return &rotations.Rotations[i], details, nil
		}
	}

	return nil, details, nil
}
