package victorops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testMethodOverride is a local helper to verify HTTP method
func testMethodOverride(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestListScheduledOverrides(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/overrides", func(w http.ResponseWriter, r *http.Request) {
		testMethodOverride(t, r, "GET")
		w.Write([]byte(`{
			"overrides": [
				{
					"publicId": "override-123",
					"user": {
						"username": "testuser",
						"firstName": "Test",
						"lastName": "User"
					},
					"start": "2025-01-01T00:00:00Z",
					"end": "2025-01-02T00:00:00Z",
					"timezone": "UTC"
				}
			],
			"_selfUrl": "/api-public/v1/overrides"
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	overrides, details, err := client.ListScheduledOverrides(context.Background())

	if err != nil {
		t.Errorf("ListScheduledOverrides returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(overrides.Overrides) != 1 {
		t.Errorf("Expected 1 override, got %d", len(overrides.Overrides))
	}
	if overrides.Overrides[0].PublicID != "override-123" {
		t.Errorf("Expected override ID 'override-123', got '%s'", overrides.Overrides[0].PublicID)
	}
}

func TestCreateScheduledOverride(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/overrides", func(w http.ResponseWriter, r *http.Request) {
		testMethodOverride(t, r, "POST")
		// Note: Real API returns "override" key (not "schedule" as spec says)
		w.Write([]byte(`{
			"override": {
				"publicId": "override-456",
				"user": {
					"username": "testuser",
					"firstName": "Test",
					"lastName": "User"
				},
				"start": "2025-01-01T00:00:00Z",
				"end": "2025-01-02T00:00:00Z",
				"timezone": "UTC"
			},
			"_selfUrl": "/api-public/v1/overrides/override-456"
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	payload := &ScheduledOverridePayload{
		Username: "testuser",
		Start:    "2025-01-01T00:00:00Z",
		End:      "2025-01-02T00:00:00Z",
		Timezone: "UTC",
	}
	override, details, err := client.CreateScheduledOverride(context.Background(), payload)

	if err != nil {
		t.Errorf("CreateScheduledOverride returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if override.PublicID != "override-456" {
		t.Errorf("Expected override ID 'override-456', got '%s'", override.PublicID)
	}
}

func TestGetScheduledOverride(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/overrides/override-123", func(w http.ResponseWriter, r *http.Request) {
		testMethodOverride(t, r, "GET")
		w.Write([]byte(`{
			"override": {
				"publicId": "override-123",
				"user": {
					"username": "testuser",
					"firstName": "Test",
					"lastName": "User"
				},
				"start": "2025-01-01T00:00:00Z",
				"end": "2025-01-02T00:00:00Z",
				"timezone": "UTC"
			},
			"_selfUrl": "/api-public/v1/overrides/override-123"
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	override, details, err := client.GetScheduledOverride(context.Background(), "override-123")

	if err != nil {
		t.Errorf("GetScheduledOverride returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if override.PublicID != "override-123" {
		t.Errorf("Expected override ID 'override-123', got '%s'", override.PublicID)
	}
}

func TestDeleteScheduledOverride(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/overrides/override-123", func(w http.ResponseWriter, r *http.Request) {
		testMethodOverride(t, r, "DELETE")
		w.WriteHeader(200)
	})

	client := NewClient("apiID", "apiKey", server.URL)
	details, err := client.DeleteScheduledOverride(context.Background(), "override-123")

	if err != nil {
		t.Errorf("DeleteScheduledOverride returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
}
