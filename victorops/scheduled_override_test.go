package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestListScheduledOverrides(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"overrides":[{"publicId":"ov-1","user":{"username":"johndoe"},"timezone":"America/Denver"}]}`))
	})

	list, _, err := testClient.ListScheduledOverrides(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Overrides) != 1 || list.Overrides[0].GetUsername() != "johndoe" {
		t.Errorf("unexpected overrides: %#v", list.Overrides)
	}
}

func TestCreateScheduledOverride(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{"override":{"publicId":"ov-2","user":{"username":"janedoe"},"timezone":"UTC"}}`))
	})

	ov, _, err := testClient.CreateScheduledOverride(context.Background(), &ScheduledOverridePayload{
		Username: "janedoe",
		Start:    "2026-01-01T00:00:00Z",
		End:      "2026-01-02T00:00:00Z",
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ov.PublicID != "ov-2" || ov.GetUsername() != "janedoe" {
		t.Errorf("unexpected override: %#v", ov)
	}
}

func TestGetScheduledOverride(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides/ov-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"override":{"publicId":"ov-1","user":{"username":"johndoe"}}}`))
	})

	ov, _, err := testClient.GetScheduledOverride(context.Background(), "ov-1")
	if err != nil {
		t.Fatal(err)
	}
	if ov.PublicID != "ov-1" {
		t.Errorf("unexpected override: %#v", ov)
	}
}

func TestDeleteScheduledOverride(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides/ov-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteScheduledOverride(context.Background(), "ov-1")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestGetScheduledOverrideAssignments(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides/ov-1/assignments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`[{"policy":"pol-1","user":"johndoe","team":"team-a","assigned":true}]`))
	})

	assignments, _, err := testClient.GetScheduledOverrideAssignments(context.Background(), "ov-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(assignments) != 1 || assignments[0].Policy != "pol-1" || assignments[0].User != "johndoe" {
		t.Errorf("unexpected assignments: %#v", assignments)
	}
}

func TestGetScheduledOverrideAssignment(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides/ov-1/assignments/pol-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"policy":"pol-1","user":"johndoe","team":"team-a","assigned":true,"_selfUrl":"/x"}`))
	})

	assignment, _, err := testClient.GetScheduledOverrideAssignment(context.Background(), "ov-1", "pol-1")
	if err != nil {
		t.Fatal(err)
	}
	if assignment.User != "johndoe" || !assignment.Assigned || assignment.Team != "team-a" {
		t.Errorf("unexpected assignment: %#v", assignment)
	}
}

func TestUpdateScheduledOverrideAssignment(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides/ov-1/assignments/pol-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		// The request body must carry username (+ optional acceptOverlap) per
		// the API spec; the policy is identified by the URL path.
		body, _ := io.ReadAll(r.Body)
		var sent UpdateAssignmentPayload
		if err := json.Unmarshal(body, &sent); err != nil {
			t.Fatalf("bad request body: %v", err)
		}
		if sent.Username != "janedoe" || !sent.AcceptOverlap {
			t.Errorf("unexpected request body: %#v", sent)
		}
		w.Write([]byte(`{"policy":"pol-1","user":"janedoe","team":"team-a","assigned":true}`))
	})

	assignment, _, err := testClient.UpdateScheduledOverrideAssignment(context.Background(), "ov-1", "pol-1", &UpdateAssignmentPayload{
		Username:      "janedoe",
		AcceptOverlap: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if assignment.User != "janedoe" || !assignment.Assigned {
		t.Errorf("unexpected assignment: %#v", assignment)
	}
}

func TestDeleteScheduledOverrideAssignment(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/overrides/ov-1/assignments/pol-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.Write([]byte(`{"policy":"pol-1","user":"johndoe"}`))
	})

	assignment, _, err := testClient.DeleteScheduledOverrideAssignment(context.Background(), "ov-1", "pol-1")
	if err != nil {
		t.Fatal(err)
	}
	if assignment.Policy != "pol-1" {
		t.Errorf("unexpected assignment: %#v", assignment)
	}
}
