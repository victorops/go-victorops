package victorops

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// EscalationPolicyStepEntry is a struct to store escalation policy step entries
type EscalationPolicyStepEntry struct {
	ExecutionType string            `json:"executionType"`
	User          map[string]string `json:"user"`
	RotationGroup map[string]string `json:"rotationGroup"`
	Webhook       map[string]string `json:"webhook"`
	Email         map[string]string `json:"email"`
	TargetPolicy  map[string]string `json:"targetPolicy"`
}

// EscalationPolicySteps is a struct to store escalation policy steps
type EscalationPolicySteps struct {
	Timeout int                         `json:"timeout"`
	Entries []EscalationPolicyStepEntry `json:"entries"`
}

// EscalationPolicy is a struct to hold an escalation policy
type EscalationPolicy struct {
	Name                       string                  `json:"name"`
	TeamID                     string                  `json:"teamSlug"`
	IgnoreCustomPagingPolicies bool                    `json:"ignoreCustomPagingPolicies"`
	Steps                      []EscalationPolicySteps `json:"steps"`
	ID                         string                  `json:"slug"`
}

// EscalationPolicyListDetail is a struct to hold the details of a team or policy returned in
// the list all escalation policies API call
type EscalationPolicyListDetail struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// EscalationPolicyListElement is a struct to hold a single policy/team combination element
// returned in the list all escalation policies API call
type EscalationPolicyListElement struct {
	Policy EscalationPolicyListDetail `json:"policy"`
	Team   EscalationPolicyListDetail `json:"team"`
}

// EscalationPolicyList is a struct to hold the response from the list all escalation policies API call
type EscalationPolicyList struct {
	Policies []EscalationPolicyListElement `json:"policies"`
}

func parseEscalationPoliciesRepsonse(response string) (*EscalationPolicyList, error) {
	var escalationPolicyList EscalationPolicyList
	err := json.Unmarshal([]byte(response), &escalationPolicyList)
	return &escalationPolicyList, err
}

func parseEscalationPolicyRepsonse(response string) (*EscalationPolicy, error) {
	var escalationPolicy EscalationPolicy
	err := json.Unmarshal([]byte(response), &escalationPolicy)
	return &escalationPolicy, err
}

// CreateEscalationPolicy creates a new eslacation policy
func (c Client) CreateEscalationPolicy(escalationPolicy *EscalationPolicy) (*EscalationPolicy, *RequestDetails, error) {
	jsonEp, err := json.Marshal(escalationPolicy)
	if err != nil {
		return nil, nil, err

	}
	details, err := c.makePublicAPICall("POST", "v1/policies", bytes.NewBuffer(jsonEp), nil)
	if err != nil {
		return nil, details, err
	}

	newEscalationPolicy, err := parseEscalationPolicyRepsonse(details.ResponseBody)
	if err != nil {
		return newEscalationPolicy, details, err
	}

	return newEscalationPolicy, details, nil
}

// GetAllEscalationPolicies lists all escalation policies for the org
func (c Client) GetAllEscalationPolicies() (*EscalationPolicyList, *RequestDetails, error) {
	details, err := c.makePublicAPICall("GET", "v1/policies", http.NoBody, nil)
	if err != nil {
		return nil, details, err
	}

	policyList, err := parseEscalationPoliciesRepsonse(details.ResponseBody)
	if err != nil {
		return policyList, details, err
	}

	return policyList, details, nil
}

// GetEscalationPolicy gets an escalation policy by ID
func (c Client) GetEscalationPolicy(escalationPolicyID string) (*EscalationPolicy, *RequestDetails, error) {
	details, err := c.makePublicAPICall("GET", "v1/policies/"+escalationPolicyID, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	newEscalationPolicy, err := parseEscalationPolicyRepsonse(details.ResponseBody)
	if err != nil {
		return newEscalationPolicy, details, err
	}

	return newEscalationPolicy, details, nil
}

// DeleteEscalationPolicy deletes an escalation policy by ID
func (c Client) DeleteEscalationPolicy(escalationPolicyID string) (*RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("DELETE", "v1/policies/"+escalationPolicyID, bytes.NewBufferString("{}"), nil)
	return details, err
}
