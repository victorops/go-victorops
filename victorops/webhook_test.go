package victorops

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testMethodWebhook is a local helper to verify HTTP method
func testMethodWebhook(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.Method != want {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

func TestListWebhooks(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/api-public/v1/webhooks", func(w http.ResponseWriter, r *http.Request) {
		testMethodWebhook(t, r, "GET")
		w.Write([]byte(`{
			"webhooks": [
				{
					"slug": "webhook-abc123",
					"name": "Test Webhook",
					"url": "https://example.com/webhook"
				}
			]
		}`))
	})

	client := NewClient("apiID", "apiKey", server.URL)
	webhooks, details, err := client.ListWebhooks(context.Background())

	if err != nil {
		t.Errorf("ListWebhooks returned error: %v", err)
	}
	if details.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", details.StatusCode)
	}
	if len(webhooks.Webhooks) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(webhooks.Webhooks))
	}
	if webhooks.Webhooks[0].Slug != "webhook-abc123" {
		t.Errorf("Expected webhook slug 'webhook-abc123', got '%s'", webhooks.Webhooks[0].Slug)
	}
}
