package victorops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testMethodAlertRule is a helper to verify HTTP method (local to avoid circular deps)
func testMethodAlertRule(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestListAlertRules(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/alertRules", func(w http.ResponseWriter, r *http.Request) {
		testMethodAlertRule(t, r, "GET")
		// Return bare array (matches real API and spec)
		w.Write([]byte(`[
			{
				"id": 123,
				"alertField": "monitoring_tool",
				"alertValueMatch": "test-*",
				"matchType": "WILDCARD",
				"stopFlag": false,
				"rank": 1,
				"routeKey": "test-route"
			}
		]`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	rules, details, err := client.ListAlertRules(context.Background())

	if err != nil {
		t.Errorf("ListAlertRules returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}
	if rules[0].ID != 123 {
		t.Errorf("Expected rule ID 123, got %d", rules[0].ID)
	}
	if rules[0].AlertField != "monitoring_tool" {
		t.Errorf("Expected alertField 'monitoring_tool', got '%s'", rules[0].AlertField)
	}
	if rules[0].MatchType != "WILDCARD" {
		t.Errorf("Expected matchType 'WILDCARD', got '%s'", rules[0].MatchType)
	}
}

func TestCreateAlertRule(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/alertRules", func(w http.ResponseWriter, r *http.Request) {
		testMethodAlertRule(t, r, "POST")
		w.Write([]byte(`{
			"id": 456,
			"alertField": "monitoring_tool",
			"alertValueMatch": "new-*",
			"matchType": "WILDCARD",
			"stopFlag": false,
			"rank": 2
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	payload := &AlertRulePayload{
		AlertField:      "monitoring_tool",
		AlertValueMatch: "new-*",
		MatchType:       "WILDCARD",
		StopFlag:        false,
		Rank:            2,
		Annotations:     []AlertAnnotationPayload{},
	}
	rule, details, err := client.CreateAlertRule(context.Background(), payload)

	if err != nil {
		t.Errorf("CreateAlertRule returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if rule.ID != 456 {
		t.Errorf("Expected rule ID 456, got %d", rule.ID)
	}
	if rule.AlertField != "monitoring_tool" {
		t.Errorf("Expected alertField 'monitoring_tool', got '%s'", rule.AlertField)
	}
}

func TestUpdateAlertRule(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/alertRules/123", func(w http.ResponseWriter, r *http.Request) {
		testMethodAlertRule(t, r, "PUT")
		w.Write([]byte(`{
			"id": 123,
			"alertField": "monitoring_tool",
			"alertValueMatch": "updated-*",
			"matchType": "REGEX",
			"stopFlag": true,
			"rank": 1
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	payload := &AlertRulePayload{
		AlertField:      "monitoring_tool",
		AlertValueMatch: "updated-*",
		MatchType:       "REGEX",
		StopFlag:        true,
		Rank:            1,
		Annotations:     []AlertAnnotationPayload{},
	}
	rule, details, err := client.UpdateAlertRule(context.Background(), "123", payload)

	if err != nil {
		t.Errorf("UpdateAlertRule returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if rule.AlertValueMatch != "updated-*" {
		t.Errorf("Expected alertValueMatch 'updated-*', got '%s'", rule.AlertValueMatch)
	}
	if rule.MatchType != "REGEX" {
		t.Errorf("Expected matchType 'REGEX', got '%s'", rule.MatchType)
	}
	if rule.StopFlag != true {
		t.Errorf("Expected stopFlag true, got %v", rule.StopFlag)
	}
}

func TestDeleteAlertRule(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/alertRules/123", func(w http.ResponseWriter, r *http.Request) {
		testMethodAlertRule(t, r, "DELETE")
		w.Write([]byte(`{"id": 123}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	response, details, err := client.DeleteAlertRule(context.Background(), "123")

	if err != nil {
		t.Errorf("DeleteAlertRule returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if response.ID != 123 {
		t.Errorf("Expected rule ID 123, got %d", response.ID)
	}
}
