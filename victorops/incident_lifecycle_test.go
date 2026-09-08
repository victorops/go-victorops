package victorops

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateIncident(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		body, _ := io.ReadAll(r.Body)
		var req CreateIncidentRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("bad body: %v", err)
		}
		if req.Summary != "DB down" || len(req.Targets) != 1 || req.Targets[0].Type != "EscalationPolicy" {
			t.Errorf("unexpected create payload: %#v", req)
		}
		w.Write([]byte(`{"incidentNumber":"123"}`))
	})

	resp, _, err := testClient.CreateIncident(context.Background(), &CreateIncidentRequest{
		Summary:  "DB down",
		UserName: "johndoe",
		Targets:  []Target{{Type: "EscalationPolicy", Slug: "pol-1"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.IncidentNumber != "123" {
		t.Errorf("unexpected response: %#v", resp)
	}
}

func TestAcknowledgeIncidents(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/ack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.Write([]byte(`{"results":[{"incidentNumber":"123","cmdAccepted":true}]}`))
	})

	resp, _, err := testClient.AcknowledgeIncidents(context.Background(), &AckOrResolveRequest{
		UserName:      "johndoe",
		IncidentNames: []string{"123"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 || !resp.Results[0].CmdAccepted {
		t.Errorf("unexpected ack response: %#v", resp)
	}
}

func TestResolveIncidents(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/resolve", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.Write([]byte(`{"results":[{"incidentNumber":"123","cmdAccepted":true}]}`))
	})

	resp, _, err := testClient.ResolveIncidents(context.Background(), &AckOrResolveRequest{
		UserName:      "johndoe",
		IncidentNames: []string{"123"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 {
		t.Errorf("unexpected resolve response: %#v", resp)
	}
}

func TestAcknowledgeIncidentsByUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/byUser/ack", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.Write([]byte(`{"results":[{"incidentNumber":"123"}]}`))
	})

	resp, _, err := testClient.AcknowledgeIncidentsByUser(context.Background(), &AckOrResolveByUserRequest{UserName: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 {
		t.Errorf("unexpected response: %#v", resp)
	}
}

func TestResolveIncidentsByUser(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/byUser/resolve", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PATCH")
		w.Write([]byte(`{"results":[{"incidentNumber":"123"}]}`))
	})

	resp, _, err := testClient.ResolveIncidentsByUser(context.Background(), &AckOrResolveByUserRequest{UserName: "johndoe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 {
		t.Errorf("unexpected response: %#v", resp)
	}
}

func TestRerouteIncidents(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/reroute", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{"statuses":[{"incidentNumber":"123","message":"rerouted","success":true,"targetStatus":[{"slug":"janedoe","success":true,"message":"ok"}]}]}`))
	})

	resp, _, err := testClient.RerouteIncidents(context.Background(), &RerouteRequest{
		UserName: "johndoe",
		Reroutes: []RerouteEntry{{IncidentNumber: "123", Targets: []Target{{Type: "User", Slug: "janedoe"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Statuses) != 1 || !resp.Statuses[0].Success || resp.Statuses[0].IncidentNumber != "123" {
		t.Fatalf("unexpected reroute response: %#v", resp)
	}
	ts := resp.Statuses[0].TargetStatus
	if len(ts) != 1 || ts[0].Slug != "janedoe" || !ts[0].Success {
		t.Errorf("unexpected reroute target status: %#v", ts)
	}
}

func TestGetIncidentNotes(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"name":"note-1","displayName":"Status","json_value":{"text":"looking into it"}}`))
	})

	resp, _, err := testClient.GetIncidentNotes(context.Background(), 123)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Name != "note-1" || resp.DisplayName != "Status" || resp.JSONValue["text"] != "looking into it" {
		t.Errorf("unexpected notes: %#v", resp)
	}
}

func TestCreateIncidentNote(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		body, _ := io.ReadAll(r.Body)
		var payload NotePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("bad body: %v", err)
		}
		// The request must use name/display_name/json_value.
		if payload.Name != "note-2" || payload.DisplayName != "Mitigation" || payload.JSONValue["text"] != "mitigated" {
			t.Errorf("unexpected note payload: %#v", payload)
		}
		w.Write([]byte(`{"name":"note-2","displayName":"Mitigation","json_value":{"text":"mitigated"}}`))
	})

	resp, _, err := testClient.CreateIncidentNote(context.Background(), 123, NotePayload{
		Name:        "note-2",
		DisplayName: "Mitigation",
		JSONValue:   map[string]interface{}{"text": "mitigated"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Name != "note-2" || resp.JSONValue["text"] != "mitigated" {
		t.Errorf("unexpected response: %#v", resp)
	}
}

func TestUpdateIncidentNote(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/123/notes/note-2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		body, _ := io.ReadAll(r.Body)
		var payload NotePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("bad body: %v", err)
		}
		if payload.JSONValue["text"] != "resolved" {
			t.Errorf("unexpected note payload: %#v", payload)
		}
		// A successful update returns an empty 200 body (no response schema).
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.UpdateIncidentNote(context.Background(), 123, "note-2", NotePayload{
		JSONValue: map[string]interface{}{"text": "resolved"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}

func TestDeleteIncidentNote(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/incidents/123/notes/note-2", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteIncidentNote(context.Background(), 123, "note-2")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", details.StatusCode)
	}
}
