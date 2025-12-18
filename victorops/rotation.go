package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
)

// RotationGroup represents a rotation group in v1 API
type RotationGroup struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// RotationGroupList represents the response from listing rotation groups (v1)
type RotationGroupList struct {
	RotationGroups []RotationGroup `json:"rotationGroups,omitempty"`
}

// Rotation represents a rotation in v2 API
type Rotation struct {
	Name            string          `json:"name,omitempty"`
	Slug            string          `json:"slug,omitempty"`
	ShiftLength     int             `json:"shiftLength,omitempty"`
	ShiftLengthUnit string          `json:"shiftLengthUnit,omitempty"`
	HandoffDay      string          `json:"handoffDay,omitempty"`
	HandoffTime     string          `json:"handoffTime,omitempty"`
	TimeZone        string          `json:"timeZone,omitempty"`
	RestrictionType string          `json:"restrictionType,omitempty"`
	Shifts          []RotationShift `json:"shifts,omitempty"`
	Mask1           *RotationMask   `json:"mask1,omitempty"`
	Mask2           *RotationMask   `json:"mask2,omitempty"`
	Mask3           *RotationMask   `json:"mask3,omitempty"`
}

// RotationShift represents a shift within a rotation
type RotationShift struct {
	Name    string   `json:"name,omitempty"`
	Slug    string   `json:"slug,omitempty"`
	Members []string `json:"members,omitempty"`
}

// RotationMask represents restriction masks for rotations
type RotationMask struct {
	Days      []string `json:"days,omitempty"`
	StartTime string   `json:"startTime,omitempty"`
	EndTime   string   `json:"endTime,omitempty"`
}

// RotationList represents the response from listing rotations (v2)
type RotationList struct {
	Rotations []Rotation `json:"rotations,omitempty"`
}

// ListRotationsV1 lists all rotation groups for a team (v1 API)
func (c *Client) ListRotationsV1(ctx context.Context, teamSlug string) (*RotationGroupList, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/teams/"+url.QueryEscape(teamSlug)+"/rotations", bytes.NewBufferString("{}"), nil)
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

// ListRotationsV2 lists all rotations with details for a team (v2 API)
func (c *Client) ListRotationsV2(ctx context.Context, teamSlug string) (*RotationList, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v2/team/"+url.QueryEscape(teamSlug)+"/rotations", bytes.NewBufferString("{}"), nil)
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
