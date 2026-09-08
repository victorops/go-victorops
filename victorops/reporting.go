package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

// HoursMinutes represents an hours/minutes duration total in the on-call log.
type HoursMinutes struct {
	Hours   float64 `json:"hours,omitempty"`
	Minutes float64 `json:"minutes,omitempty"`
}

// OnCallInterval is a single on-call interval within a user's on-call log.
type OnCallInterval struct {
	Duration         HoursMinutes `json:"duration,omitempty"`
	EscalationPolicy PolicyInfo   `json:"escalationPolicy,omitempty"`
	Off              string       `json:"off,omitempty"`
	On               string       `json:"on,omitempty"`
}

// UserLog holds a single user's on-call intervals and totals for the log window.
type UserLog struct {
	AdjustedTotal HoursMinutes     `json:"adjustedTotal,omitempty"`
	Log           []OnCallInterval `json:"log,omitempty"`
	Total         HoursMinutes     `json:"total,omitempty"`
	UserID        string           `json:"userId,omitempty"`
}

// OnCallLog represents the response from the on-call log endpoint.
type OnCallLog struct {
	TeamSlug string    `json:"teamSlug,omitempty"`
	Start    string    `json:"start,omitempty"`
	End      string    `json:"end,omitempty"`
	UserLogs []UserLog `json:"userLogs,omitempty"`
}

// ReportingIncident represents an incident in the reporting API
type ReportingIncident struct {
	IncidentNumber    string   `json:"incidentNumber,omitempty"`
	EntityID          string   `json:"entityId,omitempty"`
	EntityDisplayName string   `json:"entityDisplayName,omitempty"`
	CurrentPhase      string   `json:"currentPhase,omitempty"`
	StartTime         string   `json:"startTime,omitempty"`
	EndTime           string   `json:"endTime,omitempty"`
	Host              string   `json:"host,omitempty"`
	Service           string   `json:"service,omitempty"`
	RoutingKey        string   `json:"routingKey,omitempty"`
	AlertCount        int      `json:"alertCount,omitempty"`
	PagedUsers        []string `json:"pagedUsers,omitempty"`
	PagedTeams        []string `json:"pagedTeams,omitempty"`
}

// ReportingIncidentList represents the response from the reporting incidents endpoint
type ReportingIncidentList struct {
	Total     int                 `json:"total,omitempty"`
	Offset    int                 `json:"offset,omitempty"`
	Limit     int                 `json:"limit,omitempty"`
	Incidents []ReportingIncident `json:"incidents,omitempty"`
}

// SearchIncidentsOptions contains options for searching incidents
type SearchIncidentsOptions struct {
	Offset         int
	Limit          int
	EntityID       string
	IncidentNumber string
	StartedAfter   string
	StartedBefore  string
	Host           string
	Service        string
	CurrentPhase   string
	RoutingKey     string
}

// GetOnCallLog gets the on-call log for a team
func (c *Client) GetOnCallLog(ctx context.Context, teamSlug string, start string, end string, userName string) (*OnCallLog, *RequestDetails, error) {
	queryParams := make(map[string]string)
	if start != "" {
		queryParams["start"] = start
	}
	if end != "" {
		queryParams["end"] = end
	}
	if userName != "" {
		queryParams["userName"] = userName
	}

	details, err := c.makeReportingAPICall(ctx, "GET", "api-reporting/v1/team/"+url.PathEscape(teamSlug)+"/oncall/log", bytes.NewBufferString("{}"), queryParams)
	if err != nil {
		return nil, details, err
	}

	var log OnCallLog
	err = json.Unmarshal([]byte(details.ResponseBody), &log)
	if err != nil {
		return nil, details, err
	}

	return &log, details, nil
}

// SearchIncidents searches incident history with various filters
func (c *Client) SearchIncidents(ctx context.Context, options *SearchIncidentsOptions) (*ReportingIncidentList, *RequestDetails, error) {
	if options == nil {
		options = &SearchIncidentsOptions{}
	}

	queryParams := make(map[string]string)

	if options.Offset > 0 {
		queryParams["offset"] = strconv.Itoa(options.Offset)
	}
	if options.Limit > 0 {
		queryParams["limit"] = strconv.Itoa(options.Limit)
	}
	if options.EntityID != "" {
		queryParams["entityId"] = options.EntityID
	}
	if options.IncidentNumber != "" {
		queryParams["incidentNumber"] = options.IncidentNumber
	}
	if options.StartedAfter != "" {
		queryParams["startedAfter"] = options.StartedAfter
	}
	if options.StartedBefore != "" {
		queryParams["startedBefore"] = options.StartedBefore
	}
	if options.Host != "" {
		queryParams["host"] = options.Host
	}
	if options.Service != "" {
		queryParams["service"] = options.Service
	}
	if options.CurrentPhase != "" {
		queryParams["currentPhase"] = options.CurrentPhase
	}
	if options.RoutingKey != "" {
		queryParams["routingKey"] = options.RoutingKey
	}

	details, err := c.makeReportingAPICall(ctx, "GET", "api-reporting/v2/incidents", bytes.NewBufferString("{}"), queryParams)
	if err != nil {
		return nil, details, err
	}

	var incidents ReportingIncidentList
	err = json.Unmarshal([]byte(details.ResponseBody), &incidents)
	if err != nil {
		return nil, details, err
	}

	return &incidents, details, nil
}

// OnCallCurrentResponse represents the response from the on-call current endpoint
type OnCallCurrentResponse struct {
	TeamsOnCall []OnCallTeam `json:"teamsOnCall,omitempty"`
}

// OnCallTeam represents a team's on-call information
type OnCallTeam struct {
	Team   TeamInfo           `json:"team,omitempty"`
	OnCall []OnCallPolicyInfo `json:"onCallNow,omitempty"`
}

// TeamInfo represents basic team information
type TeamInfo struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// OnCallPolicyInfo represents on-call information for a policy
type OnCallPolicyInfo struct {
	EscalationPolicy PolicyInfo       `json:"escalationPolicy,omitempty"`
	Users            []OnCallUserInfo `json:"users,omitempty"`
}

// PolicyInfo represents basic policy information
type PolicyInfo struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// OnCallUserInfo represents on-call user information
type OnCallUserInfo struct {
	OnCallUser UserInfo `json:"onCallUser,omitempty"`
}

// UserInfo represents basic user information
type UserInfo struct {
	Username string `json:"username,omitempty"`
}

// GetOnCallCurrent gets all users and teams currently on-call for the organization
func (c *Client) GetOnCallCurrent(ctx context.Context) (*OnCallCurrentResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/oncall/current", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response OnCallCurrentResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// UserTeamsResponse represents the response from getting a user's teams
type UserTeamsResponse struct {
	Teams []Team `json:"teams,omitempty"`
}

// GetUserTeams gets the teams a user belongs to
func (c *Client) GetUserTeams(ctx context.Context, username string) (*UserTeamsResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/user/"+url.PathEscape(username)+"/teams", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response UserTeamsResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}
