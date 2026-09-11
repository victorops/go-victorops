package victorops

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// Error-path coverage for the rotation surface (rotation.go + rotation_group.go),
// the second-largest API area. Mirrors the user error tests: a non-2xx status
// surfaces as an *APIError (details.StatusCode still populated); a malformed body
// on a 2xx response produces an unmarshal error.

func TestCreateRotationGroupClientError(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Write([]byte(`{"rotationGroups":[]}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid label"}`))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	resp, details, err := testClient.CreateRotationGroup(context.Background(), "team-a", &RotationGroupCreatePayload{Label: ""})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 400, got %v", err)
	}
	if details.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", details.StatusCode)
	}
	if details.ErrorCategory != "client_error" {
		t.Errorf("expected ErrorCategory client_error, got %q", details.ErrorCategory)
	}
	if resp != nil {
		t.Errorf("expected nil group on error, got %#v", resp)
	}
}

func TestGetRotationGroupNotFound(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/999", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	resp, details, err := testClient.GetRotationGroup(context.Background(), "team-a", 999)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 404, got %v", err)
	}
	if details.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", details.StatusCode)
	}
	if resp != nil {
		t.Errorf("expected nil group for a 404, got %#v", resp)
	}
}

func TestGetRotationGroupMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`<html>oops</html>`))
	})

	_, _, err := testClient.GetRotationGroup(context.Background(), "team-a", 100)
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

func TestUpdateRotationGroupMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{`))
	})

	_, _, err := testClient.UpdateRotationGroup(context.Background(), "team-a", 100, &RotationGroupUpdatePayload{Label: "x"})
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

func TestDeleteRotationGroupNotFound(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/999", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNotFound)
	})

	details, err := testClient.DeleteRotationGroup(context.Background(), "team-a", 999)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 404, got %v", err)
	}
	if details.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", details.StatusCode)
	}
}

func TestAddRotationShiftMemberClientError(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/members", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error":"unknown user"}`))
	})

	details, err := testClient.AddRotationShiftMember(context.Background(), "team-a", 100, 200, &RotationMemberAddPayload{Username: "ghost"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 422, got %v", err)
	}
	if details.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", details.StatusCode)
	}
	if details.ErrorCategory != "client_error" {
		t.Errorf("expected ErrorCategory client_error, got %q", details.ErrorCategory)
	}
}

func TestGetScheduledShiftUserMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200/scheduled", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`nope`))
	})

	_, _, err := testClient.GetScheduledShiftUser(context.Background(), "team-a", 100, 200)
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

func TestListRotationsV2MalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/team/team-a/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{ "rotations": [ }`))
	})

	_, _, err := testClient.ListRotationsV2(context.Background(), "team-a")
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

// TestGetRotationShiftRetriesThenSucceeds confirms idempotent GETs on the rotation
// surface get the shared retry engine: a transient 503 is retried and the eventual
// 200 is parsed. Backoff is tiny to keep the test fast.
func TestGetRotationShiftRetriesThenSucceeds(t *testing.T) {
	setup()
	defer teardown()

	var calls int
	testMux.HandleFunc("/api-public/v1/teams/team-a/rotations/100/200", func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte(`{"rot_id":200,"label":"Week"}`))
	})

	client := NewClientWithArgs("apiID", "apiKey", testServer.URL, ClientArgs{
		RateLimit:   1000,
		RetryConfig: &RetryConfig{MaxRetries: 3, InitialBackoff: 1_000_000, MaxBackoff: 5_000_000, BackoffMultiplier: 2, RetryableStatus: []int{503}},
	})

	shift, details, err := client.GetRotationShift(context.Background(), "team-a", 100, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shift.RotID != 200 {
		t.Errorf("unexpected shift: %#v", shift)
	}
	if details.RetryCount != 1 {
		t.Errorf("expected RetryCount 1, got %d", details.RetryCount)
	}
	if calls != 2 {
		t.Errorf("expected 2 attempts, got %d", calls)
	}
}
