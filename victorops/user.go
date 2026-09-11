package victorops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// User is a user in the VictorOps org.
type User struct {
	FirstName           string `json:"firstName,omitempty"`
	LastName            string `json:"lastName,omitempty"`
	Username            string `json:"username,omitempty"`
	Email               string `json:"email,omitempty"`
	Admin               bool   `json:"admin,omitempty"`
	ExpirationHours     int    `json:"expirationHours,omitempty"`
	CreatedAt           string `json:"createdAt,omitempty"`
	PasswordLastUpdated string `json:"passwordLastUpdated,omitempty"`
	Verified            bool   `json:"verified,omitempty"`
}

// UserList is a list of Users
type UserList struct {
	Users [][]User `json:"users"`
}

// UserListV2 is a list of Users for Version 2
type UserListV2 struct {
	Users []User `json:"users"`
}

const (
	userV1Endpoint = "v1/user"
	userV2Endpoint = "v2/user"
)

func parseUserResponse(response string) (*User, error) {
	// Parse the response and return the user object
	var user User
	err := json.Unmarshal([]byte(response), &user)
	if err != nil {
		return nil, err
	}

	return &user, err
}

// CreateUser creates a user in the victorops organization
func (c *Client) CreateUser(ctx context.Context, user *User) (*User, *RequestDetails, error) {
	jsonUser, err := json.Marshal(user)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall(ctx, "POST", userV1Endpoint, bytes.NewBuffer(jsonUser), nil)
	if err != nil {
		return nil, details, err
	}

	newUser, err := parseUserResponse(details.ResponseBody)
	if err != nil {
		return newUser, details, err
	}

	return newUser, details, nil
}

// AddUserPayload is a single user entry for a batch user creation request.
type AddUserPayload struct {
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Admin           bool   `json:"admin,omitempty"`
	ExpirationHours int    `json:"expirationHours,omitempty"`
}

// BatchUserError describes an error for a single user in a batch create response.
type BatchUserError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// BatchUserResult is the per-user result of a batch create request.
type BatchUserResult struct {
	Username string           `json:"username,omitempty"`
	Errors   []BatchUserError `json:"errors,omitempty"`
}

// CreateUsersBatch adds multiple users to the organization in a single request.
// The response contains a per-user result; check each result's Errors field.
func (c *Client) CreateUsersBatch(ctx context.Context, users []AddUserPayload) ([]BatchUserResult, *RequestDetails, error) {
	body, err := json.Marshal(users)
	if err != nil {
		return nil, nil, err
	}

	details, err := c.makePublicAPICall(ctx, "POST", "v1/user/batch", bytes.NewBuffer(body), nil)
	if err != nil {
		return nil, details, err
	}

	var results []BatchUserResult
	if err := json.Unmarshal([]byte(details.ResponseBody), &results); err != nil {
		return nil, details, err
	}

	return results, details, nil
}

// GetUser returns a specific user within this victorops organization
func (c *Client) GetUser(ctx context.Context, username string) (*User, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall(ctx, "GET", userV1Endpoint+"/"+url.PathEscape(username), bytes.NewBufferString("{}"), nil)

	// Check for errors
	if err != nil {
		return nil, details, err
	}

	user, err := parseUserResponse(details.ResponseBody)
	if err != nil {
		return user, details, err
	}

	return user, details, nil
}

// DeleteUser deletes a user from the victorops org
func (c *Client) DeleteUser(ctx context.Context, username string, replacementUser string) (*RequestDetails, error) {
	body, err := json.Marshal(map[string]string{"replacement": replacementUser})
	if err != nil {
		return nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall(ctx, "DELETE", userV1Endpoint+"/"+url.PathEscape(username), bytes.NewBuffer(body), nil)

	// Check for errors
	if err != nil {
		return details, err
	}

	return details, nil
}

// GetAllUsers returns a list of all of the users in the victorops org
func (c *Client) GetAllUsers(ctx context.Context) (*UserList, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall(ctx, "GET", userV1Endpoint, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, details, err
	}

	var userList UserList
	err = json.Unmarshal([]byte(details.ResponseBody), &userList)
	if err != nil {
		return nil, details, err
	}

	return &userList, details, nil
}

// GetAllUserV2 returns a list of all of the users in the victorops org
func (c *Client) GetAllUserV2(ctx context.Context) (*UserListV2, *RequestDetails, error) {
	return c.getAllUsersV2(ctx, userV2Endpoint, nil)
}

// GetUserByEmail returns a list of all of the user(s) in the victorops org that matches the given email
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*UserListV2, *RequestDetails, error) {
	// This endpoint does not decode an escaped @ in the email query value.
	// Escape the rest of the value normally, but preserve @ for API compatibility.
	escapedEmail := strings.ReplaceAll(url.QueryEscape(email), "%40", "@")
	return c.getAllUsersV2(ctx, userV2Endpoint+"?email="+escapedEmail, nil)
}

func (c *Client) getAllUsersV2(ctx context.Context, endpoint string, queryParams map[string]string) (*UserListV2, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall(ctx, "GET", endpoint, bytes.NewBufferString("{}"), queryParams)
	if err != nil {
		return nil, details, err
	}

	var userList UserListV2
	err = json.Unmarshal([]byte(details.ResponseBody), &userList)
	if err != nil {
		return nil, details, err
	}

	return &userList, details, nil
}

// UpdateUser updates a victorops user
func (c *Client) UpdateUser(ctx context.Context, user *User) (*User, *RequestDetails, error) {
	jsonUser, err := json.Marshal(user)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall(ctx, "PUT", userV1Endpoint+"/"+url.PathEscape(user.Username), bytes.NewBuffer(jsonUser), nil)
	if err != nil {
		return nil, details, err
	}

	newUser, err := parseUserResponse(details.ResponseBody)
	if err != nil {
		return newUser, details, err
	}

	return newUser, details, nil
}

type emailsResponse struct {
	ContactMethods []map[string]interface{} `json:"contactMethods"`
}

// GetUserDefaultEmailContactID returns the id of the default email contact for a user
// TODO: Utilize the contact method methods for this
func (c *Client) GetUserDefaultEmailContactID(ctx context.Context, username string) (float64, *RequestDetails, error) {
	// Make the request
	requestDetails, err := c.makePublicAPICall(ctx, "GET", userV1Endpoint+"/"+url.PathEscape(username)+"/contact-methods/emails", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return 0, requestDetails, err
	}

	var er emailsResponse
	err = json.Unmarshal([]byte(requestDetails.ResponseBody), &er)
	if err != nil {
		return 0, requestDetails, err
	}

	// Crawl through and find the right one. Guard the type assertions so a
	// malformed contact method entry cannot panic the caller.
	for _, cm := range er.ContactMethods {
		labelValue, exists := cm["label"]
		if !exists {
			continue
		}
		label, ok := labelValue.(string)
		if !ok {
			return 0, requestDetails, fmt.Errorf("unexpected email contact label type %T", labelValue)
		}
		if label != "Default" {
			continue
		}
		if id, ok := cm["id"].(float64); ok {
			return id, requestDetails, nil
		}
		return 0, requestDetails, fmt.Errorf("unexpected default email contact id type %T", cm["id"])
	}

	return 0, requestDetails, fmt.Errorf("default email contact not found for user %q", username)
}
