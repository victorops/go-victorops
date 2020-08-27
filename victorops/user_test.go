package victorops

import (
	"net/http"
	"reflect"
	"testing"
)

func TestCreateUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`
		{
			"firstName": "test",
			"lastName": "user",
			"username": "go_testuser",
			"email": "go_test@victorops.com",
			"createdAt": "2020-03-25T17:49:01Z",
			"passwordLastUpdated": "2020-03-25T17:49:01Z",
			"verified": false,
			"_selfUrl": "/api-public/v1/user/go_testuser"
		  }
		`))
	})

	user := &User{
		FirstName:       "test",
		LastName:        "user",
		Username:        "go_testuser",
		Email:           "go_test@victorops.com",
		Admin:           true,
		ExpirationHours: 24,
	}

	resp, _, err := testClient.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}

	// Note, the response here doesn't reflect some of the values we would really expect
	// this is because they are not part of the response from the API and in parsing the JSON
	// if a value is missing golang will use the "empty value".
	// The actual Use
	want := &User{
		FirstName:           "test",
		LastName:            "user",
		Username:            "go_testuser",
		Email:               "go_test@victorops.com",
		CreatedAt:           "2020-03-25T17:49:01Z",
		PasswordLastUpdated: "2020-03-25T17:49:01Z",
		Verified:            false,
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestCreateUserUnavailableUsername(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`
		{
		  "error": "User name go_testuser is unavailable"
		}
		`))
	})

	user := &User{
		FirstName:       "test",
		LastName:        "user",
		Username:        "go_testuser",
		Email:           "go_test@victorops.com",
		Admin:           true,
		ExpirationHours: 24,
	}

	resp, _, err := testClient.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}

	want := &User{
		FirstName:           "",
		LastName:            "",
		Username:            "",
		Email:               "",
		Admin:               false,
		ExpirationHours:     0,
		CreatedAt:           "",
		PasswordLastUpdated: "",
		Verified:            false,
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestCreateUserInvalidResponse(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`
		Cloudflare is not avaialble.
		`))
	})

	user := &User{
		FirstName:       "test",
		LastName:        "user",
		Username:        "go_testuser",
		Email:           "go_test@victorops.com",
		Admin:           true,
		ExpirationHours: 24,
	}

	_, _, err := testClient.CreateUser(user)

	if err.Error() != "invalid character 'C' looking for beginning of value" {
		t.Errorf("expected CreateUser to error out on an invalid response from the server")
	}
}
