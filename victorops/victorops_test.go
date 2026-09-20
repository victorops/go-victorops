package victorops

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	testMux *http.ServeMux

	// client is the VictorOps client being tested.
	testClient *Client

	// server is a test HTTP server used to provide mock API responses.
	testServer *httptest.Server
)

func setup() {
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	testMux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result":"success","returnTo":"/client/vo_go_test","username":"username","orgslug":"org"}`))
	})

	testClient = NewClient("apiID", "apiKey", testServer.URL)
	log.Printf("Client instantiated: %s", testClient.publicBaseURL)
}

func teardown() {
	testServer.Close()
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func TestConfigurableClient(t *testing.T) {
	args := http.Client{Timeout: 30 * time.Second}
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	testConfigurableClient := NewConfigurableClient("apiID", "apiKey", testServer.URL, args)
	log.Printf("Client instantiated: %s", testConfigurableClient.publicBaseURL)
	if testConfigurableClient.GetHTTPClient() == nil {
		t.Errorf("http client is nil")
	}
}

func TestConfigurableClientTimeout(t *testing.T) {
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})

	// Disable retries so the idempotent GET fails fast on the single timeout
	// rather than replaying the request across the backoff schedule.
	testConfigurableClient := NewClientWithArgs("apiID", "apiKey", testServer.URL, ClientArgs{
		TimeoutSeconds: 1,
		RetryConfig:    &RetryConfig{MaxRetries: 0},
	})
	log.Printf("Client instantiated: %s", testConfigurableClient.publicBaseURL)
	_, _, err := testConfigurableClient.GetAllUsers(context.Background())

	if err == nil || !strings.Contains(err.Error(), "context deadline exceeded (Client.Timeout exceeded while awaiting headers)") {
		t.Errorf("expected to see timeout error, but saw: %v", err)
	}
}
