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
