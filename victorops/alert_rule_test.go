package victorops

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestListAlertRules(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alertRules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		if got := r.URL.Query().Get("routing_key"); got != "*" {
			t.Errorf("expected routing_key=*, got %q", got)
		}
		w.Write([]byte(`[{"id":10,"alertField":"host","alertValueMatch":"web*","matchType":"WILDCARD","stopFlag":true}]`))
	})

	rules, _, err := testClient.ListAlertRules(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].ID != 10 || rules[0].MatchType != "WILDCARD" {
		t.Errorf("unexpected rules: %#v", rules)
	}
}

func TestCreateAlertRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alertRules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		var payload map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if payload["routeKey"] == nil || payload["routingKey"] != nil {
			t.Errorf("expected routeKey and no routingKey in payload: %#v", payload)
		}
		w.Write([]byte(`{"id":11,"alertField":"service","alertValueMatch":"db","matchType":"WILDCARD"}`))
	})

	rule, _, err := testClient.CreateAlertRule(context.Background(), &AlertRulePayload{
		AlertField:      "service",
		AlertValueMatch: "db",
		MatchType:       "WILDCARD",
		RoutingKey:      "database",
		Annotations:     []AlertAnnotationPayload{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rule.ID != 11 || rule.AlertField != "service" {
		t.Errorf("unexpected rule: %#v", rule)
	}
}

func TestGetAlertRuleByUpdate(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alertRules/22", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"id":22,"alertField":"service","routeKey":"database"}`))
	})

	rule, _, err := testClient.GetAlertRuleByUpdate(context.Background(), "22", &AlertRulePayload{
		AlertField:      "service",
		AlertValueMatch: "db",
		MatchType:       "WILDCARD",
		Rank:            1,
		RoutingKey:      "database",
		Annotations:     []AlertAnnotationPayload{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rule == nil || rule.ID != 22 || rule.RouteKey != "database" {
		t.Errorf("unexpected rule: %#v", rule)
	}
}

func TestGetAlertRule(t *testing.T) {
	setup()
	defer teardown()

	// GetAlertRule lists all rules and filters by ID client-side.
	testMux.HandleFunc("/api-public/v1/alertRules", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`[{"id":10,"alertField":"host"},{"id":22,"alertField":"service"}]`))
	})

	rule, _, err := testClient.GetAlertRule(context.Background(), "22")
	if err != nil {
		t.Fatal(err)
	}
	if rule == nil || rule.ID != 22 || rule.AlertField != "service" {
		t.Errorf("unexpected rule: %#v", rule)
	}
}

func TestGetAlertRuleNotFound(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alertRules", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":10}]`))
	})

	rule, _, err := testClient.GetAlertRule(context.Background(), "999")
	if err != nil {
		t.Fatal(err)
	}
	if rule != nil {
		t.Errorf("expected nil rule for missing ID, got %#v", rule)
	}
}

func TestUpdateAlertRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alertRules/11", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{"id":11,"alertField":"service","alertValueMatch":"db2","matchType":"REGEX"}`))
	})

	rule, _, err := testClient.UpdateAlertRule(context.Background(), "11", &AlertRulePayload{
		AlertField:      "service",
		AlertValueMatch: "db2",
		MatchType:       "REGEX",
		Annotations:     []AlertAnnotationPayload{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rule.MatchType != "REGEX" || rule.AlertValueMatch != "db2" {
		t.Errorf("unexpected rule: %#v", rule)
	}
}

func TestDeleteAlertRule(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/alertRules/11", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.Write([]byte(`{"id":11}`))
	})

	resp, _, err := testClient.DeleteAlertRule(context.Background(), "11")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 11 {
		t.Errorf("unexpected delete response: %#v", resp)
	}
}
