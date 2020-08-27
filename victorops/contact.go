package victorops

import (
	"bytes"
	"encoding/json"
	"log"
	"net/url"
)

// Define a ContactType enum that is used internally for endpoint generation
// Golang doesn't have real enums, so this is a way to mimic it
type ContactTypes struct {
	Phone  ContactType
	Email  ContactType
	Device ContactType
}

type AllContactResponse struct {
	Phones  ContactGroup `json:"phones,omitempty"`
	Emails  ContactGroup `json:"emails,omitempty"`
	Devices ContactGroup `json:"devices,omitempty"`
}

type ContactGroup struct {
	ContactMethods []Contact `json:"contactMethods"`
}

type ContactType struct {
	endpointNoun string
}

func GetContactTypes() ContactTypes {
	contactTypes := ContactTypes{
		Phone:  ContactType{endpointNoun: "phones"},
		Email:  ContactType{endpointNoun: "emails"},
		Device: ContactType{endpointNoun: "devices"},
	}
	return contactTypes
}

// GetContactTypeFromNotificationType returns a ContactType based on the notificationType string
// returned in notification steps.
func GetContactTypeFromNotificationType(notificationType string) ContactType {
	if notificationType == "push" {
		return GetContactTypes().Device
	} else if notificationType == "email" {
		return GetContactTypes().Email
	} else if notificationType == "phone" || notificationType == "sms" {
		return GetContactTypes().Phone
	}
	return ContactType{}
}

// Contact is a struct to hold the data for a victorops phone or email contact.
// This has Email and PhoneNumber fields for when making a create request
// But on querying later, those values are always returned as "value"
type Contact struct {
	PhoneNumber string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	Label       string `json:"label,omitempty"`
	Rank        int    `json:"rank,omitempty"`
	ExtID       string `json:"extId,omitempty"`
	ID          int    `json:"id,omitempty"`
	Value       string `json:"value,omitempty"`
	Verified    string `json:"verified,omitempty"`
}

func (c Contact) Type() ContactType {
	if c.PhoneNumber != "" {
		return GetContactTypes().Phone
	} else {
		return GetContactTypes().Email
	}
}

func parseContactResponse(response string) (*Contact, error) {
	var contact Contact
	err := json.Unmarshal([]byte(response), &contact)
	if err != nil {
		return nil, err
	}

	return &contact, err
}

// CreateContact creates a new contact for a user
func (c Client) CreateContact(username string, contact *Contact) (*Contact, *RequestDetails, error) {
	jsonContact, err := json.Marshal(contact)
	if err != nil {
		return nil, nil, err
	}

	requestDetails, err := c.makePublicAPICall("POST", "v1/user/"+url.QueryEscape(username)+"/contact-methods/"+contact.Type().endpointNoun, bytes.NewBuffer(jsonContact), nil)
	if err != nil {
		return nil, requestDetails, err
	}

	newContact, err := parseContactResponse(requestDetails.ResponseBody)
	if err != nil {
		return newContact, requestDetails, err
	}

	return newContact, requestDetails, nil
}

// GetContact gets a contact for a user
func (c Client) GetContact(username string, contactExtID string, contactType ContactType) (*Contact, *RequestDetails, error) {
	requestDetails, err := c.makePublicAPICall("GET", "v1/user/"+url.QueryEscape(username)+"/contact-methods/"+contactType.endpointNoun+"/"+contactExtID, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, requestDetails, err
	}

	newContact, err := parseContactResponse(requestDetails.ResponseBody)
	if err != nil {
		return newContact, requestDetails, err
	}

	return newContact, requestDetails, nil
}

// GetAllContacts returns a list of all of the contacts for a user in the victorops org
func (c Client) GetAllContacts(username string) (*AllContactResponse, *RequestDetails, error) {
	// Make the request
	requestDetails, err := c.makePublicAPICall("GET", "v1/user/"+url.QueryEscape(username)+"/contact-methods", bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, requestDetails, err
	}
	allcontacts := AllContactResponse{}
	err = json.Unmarshal([]byte(requestDetails.ResponseBody), &allcontacts)
	if err != nil {
		return nil, requestDetails, err
	}

	return &allcontacts, requestDetails, nil
}

// DeleteContact deletes a contact
func (c Client) DeleteContact(username string, contactExtID string, contactType ContactType) (*RequestDetails, error) {
	requestDetails, err := c.makePublicAPICall("DELETE", "v1/user/"+url.QueryEscape(username)+"/contact-methods/"+contactType.endpointNoun+"/"+contactExtID, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return requestDetails, err
	}

	return requestDetails, nil
}

// Get a contact via it's internal api ID, but using the public API
type GetAllContactResponse struct {
	ContactMethods []Contact `json:"contactMethods,omitempty"`
}

func (c Client) GetContactByID(username string, id int, contactType ContactType) (*Contact, *RequestDetails, error) {
	// Device 0 is a special device for "All devices"
	if contactType == GetContactTypes().Device && id == 0 {
		contact := Contact{
			Value: "All Devices",
			ID:    0,
			Rank:  0,
		}
		return &contact, &RequestDetails{}, nil
	}

	requestDetails, err := c.makePublicAPICall("GET", "v1/user/"+url.QueryEscape(username)+"/contact-methods/"+contactType.endpointNoun, bytes.NewBufferString("{}"), nil)
	if err != nil {
		return nil, requestDetails, err
	}

	contacts := GetAllContactResponse{}
	err = json.Unmarshal([]byte(requestDetails.ResponseBody), &contacts)
	if err != nil {
		log.Println("test")
		return nil, requestDetails, err
	}

	for _, contact := range contacts.ContactMethods {
		if contact.ID == id {
			return &contact, requestDetails, err
		}
	}

	return nil, requestDetails, err
}
