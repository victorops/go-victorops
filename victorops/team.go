package victorops

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
)

// Team is a struct to hold the data for a victorops Team
type Team struct {
	Name          string `json:"name,omitempty"`
	Slug          string `json:"slug,omitempty"`
	MemberCount   int    `json:"memberCount,omitempty"`
	Version       int    `json:"version,omitempty"`
	IsDefaultTeam bool   `json:"isDefaultTeam,omitempty"`
	Description   string `json:"description,omitempty"`
}

// TeamMembers contains membership details for a team
type TeamMembers struct {
	Members []User `json:"members,omitempty"`
}

// User is a user in the VictorOps org.
type Admin struct {
	Username  string `json:"username,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	SelfUrl   string `json:"_selfUrl,omitempty"`
}

// TeamAdmins contains administrators for a team
type TeamAdmins struct {
	TeamAdmins []Admin `json:"admin,omitempty"`
}

func parseTeamResponse(response string) (*Team, error) {
	// Parse the response and return the user object
	var team Team
	err := json.Unmarshal([]byte(response), &team)
	if err != nil {
		return nil, err
	}

	return &team, err
}

func parseTeamMembersResponse(response string) (*TeamMembers, error) {
	var teamMembers TeamMembers
	err := json.Unmarshal([]byte(response), &teamMembers)
	if err != nil {
		return nil, err
	}

	return &teamMembers, err
}

func parseTeamAdminsResponse(response string) (*TeamAdmins, error) {
	var teamAdmins TeamAdmins
	err := json.Unmarshal([]byte(response), &teamAdmins)
	if err != nil {
		return nil, err
	}

	return &teamAdmins, err
}

// CreateTeam creates a team in the victorops organization
func (c Client) CreateTeam(team *Team) (*Team, *RequestDetails, error) {
	jsonTeam, err := json.Marshal(team)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall("POST", "v1/team", bytes.NewBuffer(jsonTeam), nil)
	if err != nil {
		return nil, details, err
	}

	newTeam, err := parseTeamResponse(details.ResponseBody)
	if err != nil {
		return newTeam, details, err
	}

	return newTeam, details, nil
}

// GetTeam returns a specific team within this victorops organization
func (c Client) GetTeam(teamID string) (*Team, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/team/"+teamID, bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	team, err := parseTeamResponse(details.ResponseBody)
	if err != nil {
		return team, details, err
	}

	return team, details, nil
}

// GetAllTeams returns a list of all team within this victorops organization
func (c Client) GetAllTeams() (*[]Team, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/team", bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	var teams []Team
	err = json.Unmarshal([]byte(details.ResponseBody), &teams)
	if err != nil {
		return nil, details, err
	}

	return &teams, details, nil
}

// GetTeamMembers returns a members on a team within this victorops organization
func (c Client) GetTeamMembers(teamID string) (*TeamMembers, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/team/"+teamID+"/members", bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	teamMembers, err := parseTeamMembersResponse(details.ResponseBody)
	if err != nil {
		return nil, details, err
	}

	return teamMembers, details, err
}

// DeleteTeam deletes a team from this victorops org
func (c Client) DeleteTeam(teamID string) (*RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("DELETE", "v1/team/"+teamID, bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return details, err
	}

	return details, nil
}

// UpdateTeam updates a victorops user
func (c Client) UpdateTeam(team *Team) (*Team, *RequestDetails, error) {
	jsonTeam, err := json.Marshal(team)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall("PUT", "v1/team/"+team.Name, bytes.NewBuffer(jsonTeam), nil)
	if err != nil {
		return nil, nil, err
	}

	newTeam, err := parseTeamResponse(details.ResponseBody)
	if err != nil {
		return newTeam, details, err
	}

	return newTeam, details, nil
}

// AddTeamMember adds a member to a victorops team.
func (c Client) AddTeamMember(teamID string, username string) (*RequestDetails, error) {
	details, err := c.makePublicAPICall("POST", "v1/team/"+teamID+"/members", bytes.NewBufferString("{\"username\": \""+username+"\"}"), nil)
	return details, err
}

// RemoveTeamMember Removes a member from a victorops team
func (c Client) RemoveTeamMember(teamID string, username string, replacement string) (*RequestDetails, error) {
	details, err := c.makePublicAPICall("DELETE", "v1/team/"+teamID+"/members/"+url.QueryEscape(username), bytes.NewBufferString("{\"replacement\":\""+replacement+"\"}"), nil)
	return details, err
}

// IsTeamMember Returns wether or not a user is in a specific victorops team
// TODO: Maybe we should do this using the v1/user/{username}/teams endpoint instead
func (c Client) IsTeamMember(teamID string, username string) (bool, *RequestDetails, error) {
	members, details, err := c.GetTeamMembers(teamID)
	if err != nil {
		return false, details, err
	}

	for _, member := range members.Members {
		if strings.EqualFold(member.Username, username) {
			return true, details, nil
		}
	}
	return false, details, nil
}

// GetTeamAdmins returns a list of admins for this team
func (c Client) GetTeamAdmins(teamID string) (*TeamAdmins, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/team/"+teamID+"/admins", bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	teamAdmins, err := parseTeamAdminsResponse(details.ResponseBody)
	if err != nil {
		return nil, details, err
	}

	return teamAdmins, details, err
}
