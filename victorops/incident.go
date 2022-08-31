package victorops

import (
	"bytes"
	"encoding/json"
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

// IncidentActionRequest represents the payload for a request to modify the state of an incident
type IncidentActionRequest struct {
	UserName      string   `json:"userName,omitempty"`
	IncidentNames []string `json:"incidentNames,omitempty"`
	Message       string   `json:"message,omitempty"`
}

type IncidentActionByUserRequest struct {
	UserName string `json:"userName,omitempty"`
	Message  string `json:"message,omitempty"`
}

// IncidentActionResponse represents the payload for a response to a request to modify the state of an
// incident.
type IncidentActionResponse struct {
	Results []IncidentAction `json:"results,omitempty"`
}

// IncidentAction is the result of the single action on an incident.
type IncidentAction struct {
	IncidentNumber string `json:"incidentNumber,omitempty"`
	EntityID       string `json:"entityId,omitempty"`
	CmdAccepted    bool   `json:"cmdAccepted,omitempty"`
	Message        string `json:"message,omitempty"`
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

func parseIncidentActionResponse(response string) (*IncidentActionResponse, error) {
	var incidentAcks IncidentActionResponse
	JSONResponse := []byte(response)
	err := json.Unmarshal(JSONResponse, &incidentAcks)
	if err != nil {
		return nil, err
	}

	return &incidentAcks, err
}

// GetIncident returns the details of a specific incident
func (c Client) GetIncident(incidentID int) (*Incident, *RequestDetails, error) {
	details, err := c.makePublicAPICall("GET", "v1/incidents/"+strconv.Itoa(incidentID), bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	incident, err := parseIncidentResponse(details.ResponseBody)
	return incident, details, err
}

// GetIncidents gets a list of the currently open, acknowledged and
// recently resolved incidents
func (c Client) GetIncidents() (*IncidentResponse, *RequestDetails, error) {

	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/incidents", bytes.NewBufferString("{}"), nil)

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

func (c Client) actOnIncidents(what string, userName string, incidents []int, message string) (*IncidentActionResponse, *RequestDetails, error) {
	inc := make([]string, len(incidents))
	for i := 0; i < len(incidents); i++ {
		inc[i] = strconv.Itoa(incidents[i])
	}
	request, err := json.Marshal(IncidentActionRequest{UserName: userName, IncidentNames: inc, Message: message})
	if err != nil {
		return nil, nil, err
	}
	details, err := c.makePublicAPICall("PATCH", "v1/incidents/"+what, bytes.NewBuffer(request), nil)
	// Check for errors
	if err != nil {
		return nil, details, err
	}
	incidentAct, err := parseIncidentActionResponse(details.ResponseBody)
	return incidentAct, details, err
}

func (c Client) actOnIncidentsByUser(what string, userName string, message string) (*IncidentActionResponse, *RequestDetails, error) {
	request, err := json.Marshal(IncidentActionByUserRequest{UserName: userName, Message: message})
	if err != nil {
		return nil, nil, err
	}
	details, err := c.makePublicAPICall("PATCH", "v1/incidents/byUser/"+what, bytes.NewBuffer(request), nil)
	// Check for errors
	if err != nil {
		return nil, details, err
	}
	incidentAct, err := parseIncidentActionResponse(details.ResponseBody)
	return incidentAct, details, err
}

// AckIncidents acknowledges a list of incidents
func (c Client) AckIncidents(userName string, incidents []int, message string) (*IncidentActionResponse, *RequestDetails, error) {
	return c.actOnIncidents("ack", userName, incidents, message)
}

// ResolveIncidents resolves a list of incidents
func (c Client) ResolveIncidents(userName string, incidents []int, message string) (*IncidentActionResponse, *RequestDetails, error) {
	return c.actOnIncidents("resolve", userName, incidents, message)
}

// AckIncidentsByUser acknowledges all incidents for the given user
func (c Client) AckIncidentsByUser(userName string, message string) (*IncidentActionResponse, *RequestDetails, error) {
	return c.actOnIncidentsByUser("ack", userName, message)
}

// ResolveIncidentsByUser solves all incidents for the given user
func (c Client) ResolveIncidentsByUser(userName string, message string) (*IncidentActionResponse, *RequestDetails, error) {
	return c.actOnIncidentsByUser("resolve", userName, message)
}
