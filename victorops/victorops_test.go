package victorops

import (
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
	args := ClientArgs{timeoutSeconds: 30}
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	testConfigurableClient := NewConfigurableClient("apiID", "apiKey", testServer.URL, args)
	log.Printf("Client instantiated: %s", testConfigurableClient.publicBaseURL)
	if testConfigurableClient.GetHTTPClient() == nil {
		t.Errorf("http client is nil")
	}
}

func TestConfigurableClientTimeout(t *testing.T) {
	args := ClientArgs{timeoutSeconds: 1}
	testMux = http.NewServeMux()
	testServer = httptest.NewServer(testMux)

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	})

	testConfigurableClient := NewConfigurableClient("apiID", "apiKey", testServer.URL, args)
	log.Printf("Client instantiated: %s", testConfigurableClient.publicBaseURL)
	_, _, err := testConfigurableClient.GetAllUsers()

	if !strings.Contains(err.Error(), "context deadline exceeded (Client.Timeout exceeded while awaiting headers)") {
		t.Errorf("expected to to see timeout error, but saw: %s", err.Error())
	}
}
