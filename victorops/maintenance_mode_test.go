package victorops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testMethodMaint is a local helper to verify HTTP method
func testMethodMaint(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestGetMaintenanceModeState(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/maintenancemode", func(w http.ResponseWriter, r *http.Request) {
		testMethodMaint(t, r, "GET")
		w.Write([]byte(`{
			"activeInstances": [
				{
					"instanceId": "maint-123",
					"isGlobal": false,
					"startedAt": 1735689600000,
					"startedBy": "admin",
					"purpose": "Test maintenance",
					"routingKeys": ["key1", "key2"]
				}
			]
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	state, details, err := client.GetMaintenanceModeState(context.Background())

	if err != nil {
		t.Errorf("GetMaintenanceModeState returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(state.ActiveInstances) != 1 {
		t.Errorf("Expected 1 instance, got %d", len(state.ActiveInstances))
	}
	if state.ActiveInstances[0].InstanceID != "maint-123" {
		t.Errorf("Expected instance ID 'maint-123', got '%s'", state.ActiveInstances[0].InstanceID)
	}
}

func TestStartMaintenanceMode(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/maintenancemode/start", func(w http.ResponseWriter, r *http.Request) {
		testMethodMaint(t, r, "POST")
		w.Write([]byte(`{
			"activeInstances": [
				{
					"instanceId": "maint-456",
					"isGlobal": false,
					"startedAt": 1735689600000,
					"startedBy": "admin",
					"purpose": "API test",
					"routingKeys": ["testkey"]
				}
			]
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	state, details, err := client.StartMaintenanceMode(context.Background(), []string{"testkey"}, "API test")

	if err != nil {
		t.Errorf("StartMaintenanceMode returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(state.ActiveInstances) != 1 {
		t.Errorf("Expected 1 instance, got %d", len(state.ActiveInstances))
	}
}

func TestEndMaintenanceMode(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/maintenancemode/maint-123/end", func(w http.ResponseWriter, r *http.Request) {
		testMethodMaint(t, r, "PUT")
		w.Write([]byte(`{
			"activeInstances": []
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	state, details, err := client.EndMaintenanceMode(context.Background(), "maint-123")

	if err != nil {
		t.Errorf("EndMaintenanceMode returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(state.ActiveInstances) != 0 {
		t.Errorf("Expected 0 instances, got %d", len(state.ActiveInstances))
	}
}
