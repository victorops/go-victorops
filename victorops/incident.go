package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

// PagedEntity holds references for a parsed paged policy or team for an incident
type PagedEntity struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// PagedPolicy to hold references for parsing an incident
type PagedPolicy struct {
	Policy PagedEntity `json:"policy,omitempty"`
	Team   PagedEntity `json:"team,omitempty"`
}

// Transition represents a state changes of an incident
type Transition struct {
	Name     string    `json:",omitempty"`
	At       time.Time `json:",omitempty"`
	Message  string    `json:",omitempty"`
	By       string    `json:",omitempty"`
	Manually bool
	AlertID  string `json:"alertId,omitempty"`
	AlertURL string `json:"alertUrl,omitempty"`
}

// Incident represents an incident on victorops
type Incident struct {
	AlertCount        int           `json:"alertCount,omitempty"`
	CurrentPhase      string        `json:"currentPhase,omitempty"`
	EntityDisplayName string        `json:"entityDisplayName,omitempty"`
	EntityID          string        `json:"entityId,omitempty"`
	EntityState       string        `json:"entityState,omitempty"`
	EntityType        string        `json:"entityType,omitempty"`
	Host              string        `json:"host,omitempty"`
	IncidentNumber    string        `json:"incidentNumber,omitempty"`
	LastAlertID       string        `json:"lastAlertId,omitempty"`
	LastAlertTime     time.Time     `json:"lastAlertTime,omitempty"`
	Service           string        `json:"service,omitempty"`
	StartTime         time.Time     `json:"startTime,omitempty"`
	PagedTeams        []string      `json:"pagedTeams,omitempty"`
	PagedUsers        []string      `json:"pagedUsers,omitempty"`
	PagedPolicies     []PagedPolicy `json:"pagedPolicies,omitempty"`
	Transitions       []Transition  `json:",omitempty"`
}

// IncidentResponse holds just the list of incidents from the api response
type IncidentResponse struct {
	Incidents []Incident `json:"incidents,omitempty"`
}

func parseIncidentsResponse(response string) (*IncidentResponse, error) {

	var incidentList IncidentResponse
	JSONResponse := []byte(response)
	err := json.Unmarshal(JSONResponse, &incidentList)
	if err != nil {
		return nil, err
	}

	return &incidentList, err
}

func parseIncidentResponse(response string) (*Incident, error) {

	var incident Incident
	JSONResponse := []byte(response)
	err := json.Unmarshal(JSONResponse, &incident)
	if err != nil {
		return nil, err
	}

	return &incident, err
}

// GetIncident returns the details of a specific incident
func (c *Client) GetIncident(ctx context.Context, incidentID int) (*Incident, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/incidents/"+strconv.Itoa(incidentID), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	incident, err := parseIncidentResponse(details.ResponseBody)
	return incident, details, err
}

// GetIncidents gets a list of the currently open, acknowledged and
// recently resolved incidents
func (c *Client) GetIncidents(ctx context.Context) (*IncidentResponse, *RequestDetails, error) {

	// Make the request
	details, err := c.makePublicAPICall(ctx, "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	incidentList, err := parseIncidentsResponse(details.ResponseBody)
	if err != nil {
		return incidentList, details, err
	}

	return incidentList, details, nil
}

// CreateIncidentRequest represents the payload for creating an incident
type CreateIncidentRequest struct {
	Summary          string   `json:"summary"`
	Details          string   `json:"details,omitempty"`
	UserName         string   `json:"userName"`
	Targets          []Target `json:"targets"`
	IsMultiResponder bool     `json:"isMultiResponder,omitempty"`
}

// Target represents a target for incident routing
type Target struct {
	Type string `json:"type"` // "User" or "EscalationPolicy"
	Slug string `json:"slug"`
}

// CreateIncidentResponse represents the response from creating an incident
type CreateIncidentResponse struct {
	IncidentNumber string `json:"incidentNumber,omitempty"`
	Error          string `json:"error,omitempty"`
}

// AckOrResolveRequest represents the payload for acknowledging or resolving incidents
type AckOrResolveRequest struct {
	UserName      string   `json:"userName"`
	IncidentNames []string `json:"incidentNames"`
	Message       string   `json:"message,omitempty"`
}

// AckOrResolveByUserRequest represents the payload for acking/resolving all incidents for a user
type AckOrResolveByUserRequest struct {
	UserName string `json:"userName"`
	Message  string `json:"message,omitempty"`
}

// AckOrResolveResponse represents the response from ack/resolve operations
type AckOrResolveResponse struct {
	Results []AckOrResolveResult `json:"results,omitempty"`
}

// AckOrResolveResult represents a single result from ack/resolve
type AckOrResolveResult struct {
	IncidentNumber string `json:"incidentNumber,omitempty"`
	EntityID       string `json:"entityId,omitempty"`
	CmdAccepted    bool   `json:"cmdAccepted,omitempty"`
	Message        string `json:"message,omitempty"`
}

// RerouteRequest represents the payload for rerouting incidents
type RerouteRequest struct {
	UserName   string         `json:"userName"`
	Reroutes   []RerouteEntry `json:"reroutes"`
	AddTargets bool           `json:"addTargets,omitempty"`
}

// RerouteEntry represents a single reroute configuration
type RerouteEntry struct {
	IncidentNumber string   `json:"incidentNumber"`
	Targets        []Target `json:"targets"`
}

// RerouteResponse represents the response from rerouting
type RerouteResponse struct {
	Data []RerouteResult `json:"data,omitempty"`
}

// RerouteResult represents a single reroute result
type RerouteResult struct {
	IncidentNumber string `json:"incidentNumber,omitempty"`
	Status         string `json:"status,omitempty"`
	Message        string `json:"message,omitempty"`
}

// Note represents a note attached to an incident
type Note struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
	Author  string `json:"author,omitempty"`
	Created string `json:"created,omitempty"`
}

// NotesResponse represents the response from getting notes
type NotesResponse struct {
	Notes []Note `json:"notes,omitempty"`
}

// NotePayload represents the payload for creating/updating a note
type NotePayload struct {
	Content string `json:"content"`
}

// CreateIncident creates a new incident
func (c *Client) CreateIncident(ctx context.Context, request *CreateIncidentRequest) (*CreateIncidentResponse, *RequestDetails, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/incidents", bytes.NewBuffer(jsonRequest), nil)
	if err != nil {
		return nil, details, err
	}

	var response CreateIncidentResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// AcknowledgeIncidents acknowledges one or more incidents
func (c *Client) AcknowledgeIncidents(ctx context.Context, request *AckOrResolveRequest) (*AckOrResolveResponse, *RequestDetails, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PATCH", "v1/incidents/ack", bytes.NewBuffer(jsonRequest), nil)
	if err != nil {
		return nil, details, err
	}

	var response AckOrResolveResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// ResolveIncidents resolves one or more incidents
func (c *Client) ResolveIncidents(ctx context.Context, request *AckOrResolveRequest) (*AckOrResolveResponse, *RequestDetails, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PATCH", "v1/incidents/resolve", bytes.NewBuffer(jsonRequest), nil)
	if err != nil {
		return nil, details, err
	}

	var response AckOrResolveResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// AcknowledgeIncidentsByUser acknowledges all incidents for a user
func (c *Client) AcknowledgeIncidentsByUser(ctx context.Context, request *AckOrResolveByUserRequest) (*AckOrResolveResponse, *RequestDetails, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PATCH", "v1/incidents/byUser/ack", bytes.NewBuffer(jsonRequest), nil)
	if err != nil {
		return nil, details, err
	}

	var response AckOrResolveResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// ResolveIncidentsByUser resolves all incidents for a user
func (c *Client) ResolveIncidentsByUser(ctx context.Context, request *AckOrResolveByUserRequest) (*AckOrResolveResponse, *RequestDetails, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PATCH", "v1/incidents/byUser/resolve", bytes.NewBuffer(jsonRequest), nil)
	if err != nil {
		return nil, details, err
	}

	var response AckOrResolveResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// RerouteIncidents reroutes one or more incidents to different targets
func (c *Client) RerouteIncidents(ctx context.Context, request *RerouteRequest) (*RerouteResponse, *RequestDetails, error) {
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/incidents/reroute", bytes.NewBuffer(jsonRequest), nil)
	if err != nil {
		return nil, details, err
	}

	var response RerouteResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// GetIncidentNotes gets all notes for an incident
func (c *Client) GetIncidentNotes(ctx context.Context, incidentNumber int) (*NotesResponse, *RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "GET", "v1/incidents/"+strconv.Itoa(incidentNumber)+"/notes", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var response NotesResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// CreateIncidentNote creates a new note on an incident
func (c *Client) CreateIncidentNote(ctx context.Context, incidentNumber int, content string) (*NotesResponse, *RequestDetails, error) {
	payload := NotePayload{Content: content}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/incidents/"+strconv.Itoa(incidentNumber)+"/notes", bytes.NewBuffer(jsonPayload), nil)
	if err != nil {
		return nil, details, err
	}

	var response NotesResponse
	err = json.Unmarshal([]byte(details.ResponseBody), &response)
	if err != nil {
		return nil, details, err
	}

	return &response, details, nil
}

// UpdateIncidentNote updates an existing note on an incident
func (c *Client) UpdateIncidentNote(ctx context.Context, incidentNumber int, noteName string, content string) (*RequestDetails, error) {
	payload := NotePayload{Content: content}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	details, err := c.makePublicAPICall(ctx, "PUT", "v1/incidents/"+strconv.Itoa(incidentNumber)+"/notes/"+url.QueryEscape(noteName), bytes.NewBuffer(jsonPayload), nil)
	return details, err
}

// DeleteIncidentNote deletes a note from an incident
func (c *Client) DeleteIncidentNote(ctx context.Context, incidentNumber int, noteName string) (*RequestDetails, error) {
	details, err := c.makePublicAPICall(ctx, "DELETE", "v1/incidents/"+strconv.Itoa(incidentNumber)+"/notes/"+url.QueryEscape(noteName), bytes.NewBufferString("{}"), nil)
	return details, err
}
