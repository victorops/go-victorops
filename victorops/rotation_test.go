package victorops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testMethodRotation is a local helper to verify HTTP method
func testMethodRotation(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestListRotationsV1(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/teams/team-123/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethodRotation(t, r, "GET")
		w.Write([]byte(`{
			"rotationGroups": [
				{
					"name": "Primary Rotation",
					"slug": "rtg-abc123"
				}
			]
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	rotations, details, err := client.ListRotationsV1(context.Background(), "team-123")

	if err != nil {
		t.Errorf("ListRotationsV1 returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(rotations.RotationGroups) != 1 {
		t.Errorf("Expected 1 rotation, got %d", len(rotations.RotationGroups))
	}
	if rotations.RotationGroups[0].Slug != "rtg-abc123" {
		t.Errorf("Expected slug 'rtg-abc123', got '%s'", rotations.RotationGroups[0].Slug)
	}
}

func TestListRotationsV2(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v2/team/team-123/rotations", func(w http.ResponseWriter, r *http.Request) {
		testMethodRotation(t, r, "GET")
		w.Write([]byte(`{
			"rotations": [
				{
					"name": "Primary Rotation",
					"slug": "rtg-abc123",
					"shiftLength": 7,
					"shiftLengthUnit": "days",
					"timeZone": "America/New_York"
				}
			]
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	rotations, details, err := client.ListRotationsV2(context.Background(), "team-123")

	if err != nil {
		t.Errorf("ListRotationsV2 returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(rotations.Rotations) != 1 {
		t.Errorf("Expected 1 rotation, got %d", len(rotations.Rotations))
	}
	if rotations.Rotations[0].ShiftLength != 7 {
		t.Errorf("Expected shift length 7, got %d", rotations.Rotations[0].ShiftLength)
	}
}
