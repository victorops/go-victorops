package victorops

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateContact(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{
			"id": 42,
			"label": "Work",
			"value": "john@example.com",
			"rank": 1
		}`))
	})

	contact, _, err := testClient.CreateContact(context.Background(), "johndoe", &Contact{
		Email: "john@example.com",
		Label: "Work",
	})
	if err != nil {
		t.Fatal(err)
	}
	if contact.ID != 42 || contact.Label != "Work" || contact.Value != "john@example.com" {
		t.Errorf("unexpected contact: %#v", contact)
	}
}

func TestGetContact(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods/emails/abc-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"id": 42,
			"label": "Work",
			"value": "john@example.com",
			"extId": "abc-123"
		}`))
	})

	contact, _, err := testClient.GetContact(context.Background(), "johndoe", "abc-123", GetContactTypes().Email)
	if err != nil {
		t.Fatal(err)
	}
	if contact.ExtID != "abc-123" || contact.ID != 42 {
		t.Errorf("unexpected contact: %#v", contact)
	}
}

func TestGetAllContacts(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"emails": { "contactMethods": [ { "id": 1, "value": "john@example.com", "label": "Default" } ] },
			"phones": { "contactMethods": [ { "id": 2, "value": "+15550001111", "label": "Mobile" } ] },
			"devices": { "contactMethods": [] }
		}`))
	})

	all, _, err := testClient.GetAllContacts(context.Background(), "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Emails.ContactMethods) != 1 || all.Emails.ContactMethods[0].ID != 1 {
		t.Errorf("unexpected emails: %#v", all.Emails)
	}
	if len(all.Phones.ContactMethods) != 1 || all.Phones.ContactMethods[0].ID != 2 {
		t.Errorf("unexpected phones: %#v", all.Phones)
	}
}

func TestDeleteContact(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods/emails/abc-123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteContact(context.Background(), "johndoe", "abc-123", GetContactTypes().Email)
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}

func TestGetContactByID(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"contactMethods": [
				{ "id": 1, "value": "a@example.com" },
				{ "id": 7, "value": "b@example.com" }
			]
		}`))
	})

	contact, _, err := testClient.GetContactByID(context.Background(), "johndoe", 7, GetContactTypes().Email)
	if err != nil {
		t.Fatal(err)
	}
	if contact == nil || contact.ID != 7 || contact.Value != "b@example.com" {
		t.Errorf("unexpected contact: %#v", contact)
	}
}

func TestGetContactByIDAllDevices(t *testing.T) {
	setup()
	defer teardown()

	// Device 0 is a special "All Devices" pseudo-contact that must not hit the API.
	contact, _, err := testClient.GetContactByID(context.Background(), "johndoe", 0, GetContactTypes().Device)
	if err != nil {
		t.Fatal(err)
	}
	if contact == nil || contact.ID != 0 || contact.Value != "All Devices" {
		t.Errorf("unexpected all-devices contact: %#v", contact)
	}
}

func TestUpdateDeviceContact(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods/devices/dev-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"deviceType":"phone","label":"Work Phone","extId":"dev-1"}`))
	})

	device, _, err := testClient.UpdateDeviceContact(context.Background(), "johndoe", "dev-1", &ContactDeviceUpdatePayload{
		DeviceLabel: "Work Phone",
	})
	if err != nil {
		t.Fatal(err)
	}
	if device.ExtID != "dev-1" || device.Label != "Work Phone" {
		t.Errorf("unexpected device: %#v", device)
	}
}
