package victorops

import (
	"context"
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

	resp, _, err := testClient.CreateUser(context.Background(), user)
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

	resp, _, err := testClient.CreateUser(context.Background(), user)
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

	_, _, err := testClient.CreateUser(context.Background(), user)

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

	resp, _, err := testClient.GetAllUserV2(context.Background())
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

	resp, _, err := testClient.GetAllUserV2(context.Background())
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

	resp, _, err := testClient.GetUserByEmail(context.Background(), testEmail)
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

func TestGetUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/go_testuser", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"firstName": "test",
			"lastName": "user",
			"username": "go_testuser",
			"email": "go_test@victorops.com"
		}`))
	})

	resp, _, err := testClient.GetUser(context.Background(), "go_testuser")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Username != "go_testuser" || resp.Email != "go_test@victorops.com" {
		t.Errorf("unexpected user: %#v", resp)
	}
}

func TestDeleteUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/go_testuser", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteUser(context.Background(), "go_testuser", "replacement_user")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}

func TestGetAllUsers(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"users": [
				[ { "username": "go_testuser", "email": "go_test@victorops.com" } ]
			]
		}`))
	})

	resp, _, err := testClient.GetAllUsers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Users) != 1 || len(resp.Users[0]) != 1 || resp.Users[0][0].Username != "go_testuser" {
		t.Errorf("unexpected users: %#v", resp)
	}
}

func TestUpdateUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/go_testuser", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{
			"firstName": "updated",
			"lastName": "user",
			"username": "go_testuser",
			"email": "go_test@victorops.com"
		}`))
	})

	resp, _, err := testClient.UpdateUser(context.Background(), &User{
		Username:  "go_testuser",
		FirstName: "updated",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.FirstName != "updated" {
		t.Errorf("unexpected user: %#v", resp)
	}
}

func TestGetUserDefaultEmailContactID(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/go_testuser/contact-methods/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"contactMethods": [
				{ "id": 111, "label": "Work" },
				{ "id": 222, "label": "Default" }
			]
		}`))
	})

	id, _, err := testClient.GetUserDefaultEmailContactID(context.Background(), "go_testuser")
	if err != nil {
		t.Fatal(err)
	}
	if id != 222 {
		t.Errorf("expected default contact id 222, got %v", id)
	}
}

func TestCreateUsersBatch(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/batch", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`[
			{"username":"newuser1"},
			{"username":"newuser2","errors":[{"code":"DUPLICATE","message":"already exists"}]}
		]`))
	})

	results, _, err := testClient.CreateUsersBatch(context.Background(), []AddUserPayload{
		{FirstName: "New", LastName: "One", Username: "newuser1", Email: "one@example.com"},
		{FirstName: "New", LastName: "Two", Username: "newuser2", Email: "two@example.com"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if len(results[0].Errors) != 0 {
		t.Errorf("expected no errors for first user, got %#v", results[0].Errors)
	}
	if len(results[1].Errors) != 1 || results[1].Errors[0].Code != "DUPLICATE" {
		t.Errorf("expected duplicate error for second user, got %#v", results[1])
	}
}
