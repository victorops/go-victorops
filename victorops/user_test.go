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

func TestGetAllUsersV2(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"users": [
			  {
				"firstName": "test",
				"lastName": "user",
				"displayName": "test user",
				"username": "go_testuser",
				"email": "go_test@victorops.com",
				"createdAt": "2018-06-16T01:19:39Z",
				"passwordLastUpdated": "2018-07-16T22:48:01Z",
				"verified": true,
				"_selfUrl": "/api-public/v1/user/go_testuser"
			  }
			]
		}`))
	})

	resp, _, err := testClient.GetAllUserV2()
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		FirstName:           "test",
		LastName:            "user",
		Username:            "go_testuser",
		Email:               "go_test@victorops.com",
		CreatedAt:           "2018-06-16T01:19:39Z",
		PasswordLastUpdated: "2018-07-16T22:48:01Z",
		Verified:            true,
	}

	expected := &UserListV2{
		Users: []User{user},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, expected)
	}
}

func TestGetAllUsersV2WrongFormat(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"users": [
				[
					{
						"firstName": "test",
						"lastName": "user",
						"displayName": "test user",
						"username": "go_testuser",
						"email": "go_test@victorops.com",
						"createdAt": "2018-06-16T01:19:39Z",
						"passwordLastUpdated": "2018-07-16T22:48:01Z",
						"verified": true,
						"_selfUrl": "/api-public/v2/user/go_testuser"
					}
				]
			]
		}`))
	})

	resp, _, err := testClient.GetAllUserV2()
	if err == nil {
		t.Fatal(err)
	}
	if resp != nil {
		t.Fatal(err)
	}
}

func TestGetUsersByEmailV2(t *testing.T) {
	setup()
	defer teardown()

	testEmail := "go_test@victorops.com"
	testMux.HandleFunc("/api-public/v2/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"users": [
			  {
				"firstName": "test",
				"lastName": "user",
				"displayName": "test user",
				"username": "go_testuser",
				"email": "go_test@victorops.com",
				"createdAt": "2018-06-16T01:19:39Z",
				"passwordLastUpdated": "2018-07-16T22:48:01Z",
				"verified": true,
				"_selfUrl": "/api-public/v2/user/go_testuser"
			  }
			]
		}`))
	})

	resp, _, err := testClient.GetUserByEmail(testEmail)
	if err != nil {
		t.Fatal(err)
	}
	user := User{
		FirstName:           "test",
		LastName:            "user",
		Username:            "go_testuser",
		Email:               testEmail,
		CreatedAt:           "2018-06-16T01:19:39Z",
		PasswordLastUpdated: "2018-07-16T22:48:01Z",
		Verified:            true,
	}

	expected := &UserListV2{
		Users: []User{user},
	}

	if !reflect.DeepEqual(resp, expected) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, expected)
	}
}
