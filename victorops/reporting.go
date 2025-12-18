package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

// OnCallLogEntry represents a single entry in the on-call log
type OnCallLogEntry struct {
	Username         string `json:"username,omitempty"`
	ChangeType       string `json:"changeType,omitempty"`
	OldOnCallUser    string `json:"oldOnCallUser,omitempty"`
	NewOnCallUser    string `json:"newOnCallUser,omitempty"`
	EscalationPolicy string `json:"escalationPolicy,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
}

// OnCallLog represents the response from the on-call log endpoint
type OnCallLog struct {
	Log []OnCallLogEntry `json:"log,omitempty"`
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

// makeReportingAPICall makes a call to the reporting API (different base path)
func (c *Client) makeReportingAPICall(ctx context.Context, method string, endpoint string, requestBody io.Reader, queryParams map[string]string) (*RequestDetails, error) {
	details := &RequestDetails{}
	// Create the request - reporting API uses a different path
	req, err := http.NewRequestWithContext(ctx, method, c.publicBaseURL+"/"+endpoint, requestBody)
	if err != nil {
		return details, err
	}

	// Set the auth headers needed for the public api
	req.Header.Set("X-VO-Api-Id", c.apiID)
	req.Header.Set("X-VO-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Set the query params
	q := req.URL.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Add the request to the details
	details.RawRequest = req
	requestDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return details, err
	}
	details.RequestBody = string(requestDump)

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		details.ErrorCategory = "rate_limit"
		return details, fmt.Errorf("rate limiter error: %w", err)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return details, err
	}

	// Read the entire response
	responseBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return details, err
	}

	details.StatusCode = resp.StatusCode
	details.ResponseBody = string(responseBody)
	details.RawResponse = resp

	return details, nil
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

	details, err := c.makeReportingAPICall(ctx, "GET", "api-reporting/v1/team/"+url.QueryEscape(teamSlug)+"/oncall/log", bytes.NewBufferString("{}"), queryParams)
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

// OnCallCurrentResponse represents the response from the on-call current endpoint
type OnCallCurrentResponse struct {
	TeamsOnCall []OnCallTeam `json:"teamsOnCall,omitempty"`
}

// OnCallTeam represents a team's on-call information
type OnCallTeam struct {
	Team   TeamInfo           `json:"team,omitempty"`
	OnCall []OnCallPolicyInfo `json:"oncall,omitempty"`
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

// GetUserTeams gets the teams a user belongs to
func (c *Client) GetUserTeams(ctx context.Context, username string) (*UserTeamsResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", fmt.Sprintf("v1/user/%s/teams", url.QueryEscape(username)), bytes.NewBufferString("{}"), nil)
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

// UserTeamsResponse represents the response from getting a user's teams
type UserTeamsResponse struct {
	Teams []Team `json:"teams,omitempty"`
}
