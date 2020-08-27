package victorops

import (
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

	resp, _, err := testClient.CreateTeam(team)
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

	resp, _, err := testClient.CreateTeam(team)
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

	_, _, err := testClient.CreateTeam(team)

	if err.Error() != "invalid character 'C' looking for beginning of value" {
		t.Errorf("expected CreateUser to error out on an invalid response from the server")
	}
}
