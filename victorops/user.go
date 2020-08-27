package victorops

import (
	"bytes"
	"encoding/json"
	"net/url"
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
func (c Client) CreateUser(user *User) (*User, *RequestDetails, error) {
	jsonUser, err := json.Marshal(user)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall("POST", "v1/user", bytes.NewBuffer(jsonUser), nil)
	if err != nil {
		return nil, details, err
	}

	newUser, err := parseUserResponse(details.ResponseBody)
	if err != nil {
		return newUser, details, err
	}

	return newUser, details, nil
}

// GetUser returns a specific user within this victorops organization
func (c Client) GetUser(username string) (*User, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/user/"+url.QueryEscape(username), bytes.NewBufferString("{}"), nil)

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
func (c Client) DeleteUser(username string, replacementUser string) (*RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("DELETE", "v1/user/"+url.QueryEscape(username), bytes.NewBufferString("{\"replacement\": \""+replacementUser+"\"}"), nil)

	// Check for errors
	if err != nil {
		return details, err
	}

	return details, nil
}

// GetAllUsers returns a list of all of the users in the victorops org
func (c Client) GetAllUsers() (*UserList, *RequestDetails, error) {
	// Make the request
	details, err := c.makePublicAPICall("GET", "v1/user", bytes.NewBufferString("{}"), nil)
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

// UpdateUser updates a victorops user
func (c Client) UpdateUser(user *User) (*User, *RequestDetails, error) {
	jsonUser, err := json.Marshal(user)
	if err != nil {
		return nil, nil, err
	}

	// Make the request
	details, err := c.makePublicAPICall("PUT", "v1/user/"+url.QueryEscape(user.Username), bytes.NewBuffer(jsonUser), nil)
	if err != nil {
		return nil, nil, err
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
func (c Client) GetUserDefaultEmailContactID(username string) (float64, *RequestDetails, error) {
	// Make the request
	requestDetails, err := c.makePublicAPICall("GET", "v1/user/"+url.QueryEscape(username)+"/contact-methods/emails", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return 0, requestDetails, err
	}

	var er emailsResponse
	err = json.Unmarshal([]byte(requestDetails.ResponseBody), &er)
	if err != nil {
		return 0, requestDetails, err
	}

	// Crawl through and find the right one
	for _, cm := range er.ContactMethods {
		if cm["label"].(string) == "Default" {
			return cm["id"].(float64), requestDetails, err
		}
	}

	return 0, requestDetails, err
}
