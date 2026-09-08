package victorops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetOnCallLog(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-reporting/v1/team/team-a/oncall/log", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		if got := r.URL.Query().Get("start"); got != "2026-01-01" {
			t.Errorf("expected start=2026-01-01, got %q", got)
		}
		w.Write([]byte(`{
			"teamSlug": "team-a",
			"start": "2026-01-01",
			"end": "2026-01-31",
			"userLogs": [
				{
					"userId": "johndoe",
					"total": {"hours": 40, "minutes": 30},
					"adjustedTotal": {"hours": 38, "minutes": 0},
					"log": [
						{
							"duration": {"hours": 8, "minutes": 0},
							"escalationPolicy": {"name": "P1", "slug": "p1"},
							"on": "2026-01-01T00:00:00Z",
							"off": "2026-01-01T08:00:00Z"
						}
					]
				}
			]
		}`))
	})

	log, _, err := testClient.GetOnCallLog(context.Background(), "team-a", "2026-01-01", "2026-01-31", "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if log.TeamSlug != "team-a" || len(log.UserLogs) != 1 {
		t.Fatalf("unexpected log: %#v", log)
	}
	ul := log.UserLogs[0]
	if ul.UserID != "johndoe" || ul.Total.Hours != 40 {
		t.Errorf("unexpected user log totals: %#v", ul)
	}
	if len(ul.Log) != 1 || ul.Log[0].EscalationPolicy.Slug != "p1" || ul.Log[0].On == "" {
		t.Errorf("unexpected user log intervals: %#v", ul.Log)
	}
}

func TestSearchIncidents(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-reporting/v2/incidents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		q := r.URL.Query()
		if q.Get("limit") != "20" || q.Get("host") != "web01" {
			t.Errorf("unexpected query params: %v", q)
		}
		w.Write([]byte(`{"total":1,"limit":20,"incidents":[{"incidentNumber":"5","host":"web01"}]}`))
	})

	list, _, err := testClient.SearchIncidents(context.Background(), &SearchIncidentsOptions{
		Limit: 20,
		Host:  "web01",
	})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 1 || len(list.Incidents) != 1 || list.Incidents[0].Host != "web01" {
		t.Errorf("unexpected incidents: %#v", list)
	}
}

// TestSearchIncidentsNilOptions ensures passing nil options does not panic and
// simply issues an unfiltered search.
func TestSearchIncidentsNilOptions(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-reporting/v2/incidents", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		if len(r.URL.Query()) != 0 {
			t.Errorf("expected no query params for nil options, got %v", r.URL.Query())
		}
		w.Write([]byte(`{"total":0,"incidents":[]}`))
	})

	list, _, err := testClient.SearchIncidents(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 0 || len(list.Incidents) != 0 {
		t.Errorf("unexpected incidents: %#v", list)
	}
}

func TestGetOnCallCurrent(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/oncall/current", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"teamsOnCall":[{"team":{"name":"Team A","slug":"team-a"},"onCallNow":[{"escalationPolicy":{"name":"P1","slug":"p1"},"users":[{"onCallUser":{"username":"johndoe"}}]}]}]}`))
	})

	resp, _, err := testClient.GetOnCallCurrent(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.TeamsOnCall) != 1 || resp.TeamsOnCall[0].Team.Slug != "team-a" {
		t.Errorf("unexpected on-call current: %#v", resp)
	}
	oc := resp.TeamsOnCall[0].OnCall
	if len(oc) != 1 || oc[0].EscalationPolicy.Slug != "p1" || len(oc[0].Users) != 1 {
		t.Errorf("expected onCallNow to populate OnCall, got: %#v", oc)
	}
}

func TestGetUserTeams(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"teams":[{"name":"Team A","slug":"team-a"}]}`))
	})

	resp, _, err := testClient.GetUserTeams(context.Background(), "johndoe")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Teams) != 1 || resp.Teams[0].Slug != "team-a" {
		t.Errorf("unexpected user teams: %#v", resp)
	}
}

// TestReportingPathSharedEngine confirms that reporting-path calls run through the
// same hardened engine as public calls: idempotent GETs are retried on transient
// server errors with the configured backoff.
func TestReportingPathSharedEngine(t *testing.T) {
	var calls int32
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	mux.HandleFunc("/api-reporting/v1/team/team-a/oncall/log", func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"userLogs":[]}`))
	})

	client := NewClientWithArgs("secret-id", "secret-key", server.URL, ClientArgs{
		RateLimit: 1000,
		RetryConfig: &RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond,
			MaxBackoff:        5 * time.Millisecond,
			BackoffMultiplier: 2,
			RetryableStatus:   []int{429, 500, 502, 503, 504},
		},
	})

	_, details, err := client.GetOnCallLog(context.Background(), "team-a", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 calls (2 failures + 1 success) via shared engine, got %d", got)
	}
	if details.RetryCount != 2 {
		t.Errorf("expected RetryCount 2 on reporting path, got %d", details.RetryCount)
	}
}

// TestReportingPathRedactsCredentials confirms reporting calls also redact the
// API id/key from the captured request dump.
func TestReportingPathRedactsCredentials(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	mux.HandleFunc("/api-reporting/v2/incidents", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"incidents":[]}`))
	})

	client := NewClientWithArgs("secret-id", "secret-key", server.URL, ClientArgs{RateLimit: 1000})

	_, details, err := client.SearchIncidents(context.Background(), &SearchIncidentsOptions{Limit: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(details.RequestBody, "secret-id") || strings.Contains(details.RequestBody, "secret-key") {
		t.Errorf("reporting request dump leaked credentials:\n%s", details.RequestBody)
	}
}
