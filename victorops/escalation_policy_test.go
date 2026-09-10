package victorops

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateEscalationPolicy(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`{
			"name": "High Severity",
			"teamSlug": "team-abcd",
			"slug": "pol-abcd",
			"ignoreCustomPagingPolicies": false,
			"steps": []
		}`))
	})

	policy, _, err := testClient.CreateEscalationPolicy(context.Background(), &EscalationPolicy{
		Name:   "High Severity",
		TeamID: "team-abcd",
	})
	if err != nil {
		t.Fatal(err)
	}
	if policy.Name != "High Severity" || policy.ID != "pol-abcd" || policy.TeamID != "team-abcd" {
		t.Errorf("unexpected policy: %#v", policy)
	}
}

func TestGetAllEscalationPolicies(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"policies": [
				{
					"policy": { "name": "High Severity", "slug": "pol-abcd" },
					"team":   { "name": "Infrastructure", "slug": "team-abcd" }
				}
			]
		}`))
	})

	list, _, err := testClient.GetAllEscalationPolicies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(list.Policies))
	}
	if list.Policies[0].Policy.Slug != "pol-abcd" || list.Policies[0].Team.Slug != "team-abcd" {
		t.Errorf("unexpected list element: %#v", list.Policies[0])
	}
}

func TestGetEscalationPolicy(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/pol-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{
			"name": "High Severity",
			"teamSlug": "team-abcd",
			"slug": "pol-abcd",
			"steps": [ { "timeout": 5, "entries": [] } ]
		}`))
	})

	policy, _, err := testClient.GetEscalationPolicy(context.Background(), "pol-abcd")
	if err != nil {
		t.Fatal(err)
	}
	if policy.ID != "pol-abcd" || len(policy.Steps) != 1 || policy.Steps[0].Timeout != 5 {
		t.Errorf("unexpected policy: %#v", policy)
	}
}

func TestDeleteEscalationPolicy(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/pol-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusOK)
	})

	details, err := testClient.DeleteEscalationPolicy(context.Background(), "pol-abcd")
	if err != nil {
		t.Fatal(err)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", details.StatusCode)
	}
}

func TestUpdateEscalationPolicy(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/policies/pol-abcd", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		var payload map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode update payload: %v", err)
		}
		if len(payload) != 2 || payload["steps"] == nil || payload["ignoreCustomPagingPolicies"] == nil {
			t.Errorf("update payload must contain only steps and ignoreCustomPagingPolicies: %#v", payload)
		}
		w.Write([]byte(`{
			"name": "Renamed Severity",
			"teamSlug": "team-abcd",
			"slug": "pol-abcd",
			"steps": []
		}`))
	})

	policy, _, err := testClient.UpdateEscalationPolicy(context.Background(), "pol-abcd", &EscalationPolicyUpdatePayload{
		IgnoreCustomPagingPolicies: true,
		Steps:                      []EscalationPolicySteps{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if policy.Name != "Renamed Severity" || policy.ID != "pol-abcd" {
		t.Errorf("unexpected policy: %#v", policy)
	}
}
