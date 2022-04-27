package victorops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type ApiTeam struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type ApiEscalationPolicy struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

type ApiUser struct {
	Username string `json:"username,omitempty"`
}

type ApiOnCallOverride struct {
	OrigOnCallUser     ApiUser             `json:"origOnCallUser,omitempty"`
	OverrideOnCallUser ApiUser             `json:"overrideOnCallUser,omitempty"`
	Start              time.Time           `json:"start,omitempty"`
	End                time.Time           `json:"end,omitempty"`
	Policy             ApiEscalationPolicy `json:"policy,omitempty"`
}

type ApiOnCallRoll struct {
	Start      time.Time `json:"start,omitempty"`
	End        time.Time `json:"end,omitempty"`
	OnCallUser ApiUser   `json:"onCallUser,omitempty"`
	IsRoll     bool      `json:"isRoll,omitempty"`
}

type ApiOnCallEntry struct {
	OnCallUser         ApiUser         `json:"onCallUser,omitempty"`
	OverrideOnCallUser ApiUser         `json:"overrideOnCallUser,omitempty"`
	OnCallType         string          `json:"onCallType,omitempty"`
	RotationName       string          `json:"rotationName,omitempty"`
	ShiftName          string          `json:"shiftName,omitempty"`
	ShiftRoll          time.Time       `json:"shiftRoll,omitempty"`
	Rolls              []ApiOnCallRoll `json:"rolls,omitempty"`
}

type ApiEscalationPolicySchedule struct {
	Policy    ApiEscalationPolicy `json:"policy,omitempty"`
	Schedule  []ApiOnCallEntry    `json:"schedule,omitempty"`
	Overrides []ApiOnCallOverride `json:"overrides,omitempty"`
}

type ApiTeamSchedule struct {
	Team      ApiTeam                       `json:"team,omitempty"`
	Schedules []ApiEscalationPolicySchedule `json:"schedules,omitempty"`
}

type ApiUserSchedule struct {
	Schedules []ApiTeamSchedule `json:"teamSchedules,omitempty"`
}

type ApiOnCallUser struct {
	OnCallUser ApiUser `json:"onCallUser,omitempty"`
}

type ApiOnCallNow struct {
	EscalationPolicy ApiEscalationPolicy `json:"escalationPolicy,omitempty"`
	Users            []ApiOnCallUser     `json:"users,omitempty"`
}

type ApiTeamOnCall struct {
	Team      ApiTeam        `json:"team,omitempty"`
	OnCallNow []ApiOnCallNow `json:"onCallNow,omitempty"`
}

type ApiTeamsOnCall struct {
	TeamsOnCall []ApiTeamOnCall `json:"teamsOnCall,omitempty"`
}

type TakeRequest struct {
	FromUser string `json:"fromUser,omitempty"`
	ToUser   string `json:"toUser,omitempty"`
}

type TakeResponse struct {
	Result string `json:"result,omitempty"`
}

func parseApiTeamScheduleResponse(response string) (*ApiTeamSchedule, error) {
	// Parse the response and return the user object
	var schedule ApiTeamSchedule
	err := json.Unmarshal([]byte(response), &schedule)
	if err != nil {
		return nil, err
	}

	return &schedule, err
}

func parseApiUserScheduleResponse(response string) (*ApiUserSchedule, error) {
	// Parse the response and return the user object
	var schedule ApiUserSchedule
	err := json.Unmarshal([]byte(response), &schedule)
	if err != nil {
		return nil, err
	}

	return &schedule, err
}

func parseGetOnCallCurrentResponse(response string) (*ApiTeamsOnCall, error) {
	// Parse the response and return the object
	var oncall ApiTeamsOnCall
	err := json.Unmarshal([]byte(response), &oncall)
	if err != nil {
		return nil, err
	}

	return &oncall, err
}

func parseTakeResponse(response string) (*TakeResponse, error) {
	// Parse the response and return the user object
	var take TakeResponse
	err := json.Unmarshal([]byte(response), &take)
	if err != nil {
		return nil, err
	}

	return &take, err
}

// Get all current on-call personnel
func (c Client) GetOnCallCurrent() (*ApiTeamsOnCall, *RequestDetails, error) {
	details, err := c.makePublicAPICall("GET", "v1/oncall/current", bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	oncall, err := parseGetOnCallCurrentResponse(details.ResponseBody)
	if err != nil {
		return oncall, details, err
	}

	return oncall, details, nil
}

func (c Client) GetApiTeamSchedule(teamSlug string, daysForward int, daysSkip int, step int) (*ApiTeamSchedule, *RequestDetails, error) {
	details, err := c.makePublicAPICall("GET", fmt.Sprintf("v2/team/%s/oncall/schedule?daysForward=%v&daysSkip=%v&step=%v", teamSlug, daysForward, daysSkip, step), bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	schedule, err := parseApiTeamScheduleResponse(details.ResponseBody)
	if err != nil {
		return schedule, details, err
	}

	return schedule, details, nil
}

func (c Client) GetUserOnCallSchedule(userName string, daysForward int, daysSkip int, step int) (*ApiUserSchedule, *RequestDetails, error) {
	details, err := c.makePublicAPICall("GET", fmt.Sprintf("v2/user/%s/oncall/schedule?daysForward=%v&daysSkip=%v&step=%v", userName, daysForward, daysSkip, step), bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	schedule, err := parseApiUserScheduleResponse(details.ResponseBody)
	if err != nil {
		return schedule, details, err
	}

	return schedule, details, nil
}

func (c Client) TakeOnCallForTeam(teamSlug string, req *TakeRequest) (*TakeResponse, *RequestDetails, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall("PATCH", fmt.Sprintf("v1/team/%s/oncall/user", teamSlug), bytes.NewBuffer(jsonReq), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	take, err := parseTakeResponse(details.ResponseBody)
	if err != nil {
		return take, details, err
	}

	return take, details, nil
}

func (c Client) TakeOnCallForPolicy(policySlug string, req *TakeRequest) (*TakeResponse, *RequestDetails, error) {
	jsonReq, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall("PATCH", fmt.Sprintf("v1/policies/%s/oncall/user", policySlug), bytes.NewBuffer(jsonReq), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	take, err := parseTakeResponse(details.ResponseBody)
	if err != nil {
		return take, details, err
	}

	return take, details, nil
}
