package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestCreateTeam(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`
		  {
			"_selfUrl": "/api-public/v1/team/go_testteam",
			"_membersUrl": "/api-public/v1/team/go_testteam/members",
			"_policiesUrl": "/api-public/v1/policies",
			"_adminsUrl": "/api-public/v1/team/go_testteam/admins",
			"name": "Go Testteam",
			"slug": "go_testteam",
			"memberCount": 0,
			"version": 0,
			"isDefaultTeam": false
		  }
		`))
	})

	team := &Team{
		Name:          "Go Testteam",
		Slug:          "go_testteam",
		IsDefaultTeam: false,
	}

	resp, _, err := testClient.CreateTeam(context.Background(), team)
	if err != nil {
		t.Fatal(err)
	}

	want := &Team{
		Name:          "Go Testteam",
		Slug:          "go_testteam",
		IsDefaultTeam: false,
		MemberCount:   0,
		Version:       0,
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestCreateTeamUnavailableTeamname(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`
		{
		  "error": "Team name go_testteam is unavailable"
		}
		`))
	})

	team := &Team{
		Name:          "Go Testteam",
		Slug:          "go_testteam",
		IsDefaultTeam: false,
	}

	resp, _, err := testClient.CreateTeam(context.Background(), team)
	if err != nil {
		t.Fatal(err)
	}

	want := &Team{
		Name:          "",
		Slug:          "",
		IsDefaultTeam: false,
		MemberCount:   0,
		Version:       0,
	}

	if !reflect.DeepEqual(resp, want) {
		t.Errorf("returned \n\n%#v want \n\n%#v", resp, want)
	}
}

func TestCreateTeamInvalidResponse(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`
		Cloudflare is not avaialble.
		`))
	})

	team := &Team{
		Name:          "Go Testteam",
		Slug:          "go_testteam",
		IsDefaultTeam: false,
	}

	_, _, err := testClient.CreateTeam(context.Background(), team)

	if err.Error() != "invalid character 'C' looking for beginning of value" {
		t.Errorf("expected CreateUser to error out on an invalid response from the server")
	}
}

func TestGetTeam(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{ "name": "Infrastructure", "slug": "team-abcd", "memberCount": 3 }`))
	})

	resp, _, err := testClient.GetTeam(context.Background(), "team-abcd")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Slug != "team-abcd" || resp.MemberCount != 3 {
		t.Errorf("unexpected team: %#v", resp)
	}
}

func TestGetAllTeams(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`[ { "name": "Infrastructure", "slug": "team-abcd" }, { "name": "Ops", "slug": "team-efgh" } ]`))
	})

	resp, _, err := testClient.GetAllTeams(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(*resp) != 2 || (*resp)[1].Slug != "team-efgh" {
		t.Errorf("unexpected teams: %#v", resp)
	}
}

func TestGetTeamMembers(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{ "members": [ { "username": "janedoe" }, { "username": "johndoe" } ] }`))
	})

	resp, _, err := testClient.GetTeamMembers(context.Background(), "team-abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Members) != 2 || resp.Members[0].Username != "janedoe" {
		t.Errorf("unexpected members: %#v", resp)
	}
}

func TestDeleteTeam(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteTeam(context.Background(), "team-abcd")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}

func TestUpdateTeam(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{ "name": "go_testteam", "slug": "team-abcd" }`))
	})

	resp, _, err := testClient.UpdateTeam(context.Background(), &Team{Slug: "team-abcd", Name: "go_testteam"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Slug != "team-abcd" {
		t.Errorf("unexpected team: %#v", resp)
	}
}

func TestAddTeamMember(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.AddTeamMember(context.Background(), "team-abcd", "janedoe")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}

func TestRemoveTeamMember(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd/members/janedoe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.RemoveTeamMember(context.Background(), "team-abcd", "janedoe", "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}

func TestIsTeamMember(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{ "members": [ { "username": "janedoe" } ] }`))
	})

	isMember, _, err := testClient.IsTeamMember(context.Background(), "team-abcd", "JaneDoe")
	if err != nil {
		t.Fatal(err)
	}
	if !isMember {
		t.Errorf("expected janedoe to be a member (case-insensitive)")
	}

	notMember, _, err := testClient.IsTeamMember(context.Background(), "team-abcd", "someoneelse")
	if err != nil {
		t.Fatal(err)
	}
	if notMember {
		t.Errorf("expected someoneelse to not be a member")
	}
}

func TestGetTeamAdmins(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/team/team-abcd/admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{ "admin": [ { "username": "janedoe", "firstName": "Jane" } ] }`))
	})

	resp, _, err := testClient.GetTeamAdmins(context.Background(), "team-abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.TeamAdmins) != 1 || resp.TeamAdmins[0].Username != "janedoe" {
		t.Errorf("unexpected admins: %#v", resp)
	}
}

// TestTeamDescriptionRoundTrip verifies that the Team.Description field is
// serialized into create/update request bodies and populated from responses.
func TestTeamDescriptionRoundTrip(t *testing.T) {
	setup()
	defer teardown()

	const description = "Handles production incidents"

	// Create: request body must carry the description, and the response must
	// unmarshal it back onto the Team.
	testMux.HandleFunc("/api-public/v1/team", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var sent Team
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		if sent.Description != description {
			t.Errorf("create request body description = %q, want %q", sent.Description, description)
		}
		w.Write([]byte(`{ "name": "Go Testteam", "slug": "team-abcd", "description": "` + description + `" }`))
	})

	created, _, err := testClient.CreateTeam(context.Background(), &Team{Name: "Go Testteam", Description: description})
	if err != nil {
		t.Fatal(err)
	}
	if created.Description != description {
		t.Errorf("created.Description = %q, want %q", created.Description, description)
	}

	// Update: request body must also carry the description.
	testMux.HandleFunc("/api-public/v1/team/team-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		var sent Team
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}
		if sent.Description != description {
			t.Errorf("update request body description = %q, want %q", sent.Description, description)
		}
		w.Write([]byte(`{ "name": "Go Testteam", "slug": "team-abcd", "description": "` + description + `" }`))
	})

	updated, _, err := testClient.UpdateTeam(context.Background(), &Team{Slug: "team-abcd", Name: "Go Testteam", Description: description})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Description != description {
		t.Errorf("updated.Description = %q, want %q", updated.Description, description)
	}
}
