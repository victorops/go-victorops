package victorops

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

// The engine turns any non-2xx HTTP response into an *APIError while still
// populating details.StatusCode/ResponseBody for inspection. A malformed body on
// a 2xx response produces a (different) unmarshal error. These tests lock in both
// behaviors for the user resource (the largest API surface).

func TestGetUserNotFound(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/ghost", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"User not found"}`))
	})

	user, details, err := testClient.GetUser(context.Background(), "ghost")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 404, got %v", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected APIError status 404, got %d", apiErr.StatusCode)
	}
	if details.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", details.StatusCode)
	}
	if details.ErrorCategory != "client_error" {
		t.Errorf("expected ErrorCategory client_error, got %q", details.ErrorCategory)
	}
	if user != nil {
		t.Errorf("expected nil user for a 404, got %#v", user)
	}
}

func TestGetUserMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`this-is-not-json`))
	})

	_, details, err := testClient.GetUser(context.Background(), "johndoe")
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected the 200 status to still be reported, got %d", details.StatusCode)
	}
}

func TestCreateUserServerReturnsClientError(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"username already exists"}`))
	})

	_, details, err := testClient.CreateUser(context.Background(), &User{Username: "dupe"})
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
}

func TestGetAllUsersMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.Write([]byte(`{"users": [ broken`))
	})

	_, _, err := testClient.GetAllUsers(context.Background())
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

func TestUpdateUserMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		w.Write([]byte(`{{`))
	})

	_, _, err := testClient.UpdateUser(context.Background(), &User{Username: "johndoe"})
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

func TestUpdateUserReturnsDetailsOnAPIError(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid user"}`))
	})

	_, details, err := testClient.UpdateUser(context.Background(), &User{Username: "johndoe"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if details == nil || details.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected response details for failed update, got %#v", details)
	}
}

func TestGetUserDefaultEmailContactIDRejectsUnexpectedID(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/johndoe/contact-methods/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"contactMethods":[{"id":"not-a-number","label":"Default"}]}`))
	})

	_, details, err := testClient.GetUserDefaultEmailContactID(context.Background(), "johndoe")
	if err == nil || !strings.Contains(err.Error(), "unexpected default email contact id type") {
		t.Fatalf("expected descriptive shape error, got %v", err)
	}
	if details == nil || details.StatusCode != http.StatusOK {
		t.Fatalf("expected response details, got %#v", details)
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v2/user", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	list, details, err := testClient.GetUserByEmail(context.Background(), "nobody@example.com")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 404, got %v", err)
	}
	if details.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", details.StatusCode)
	}
	if list != nil {
		t.Errorf("expected nil list for a 404, got %#v", list)
	}
}

func TestCreateUsersBatchMalformedJSON(t *testing.T) {
	setup()
	defer teardown()

	testMux.HandleFunc("/api-public/v1/user/batch", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.Write([]byte(`not-an-array`))
	})

	_, _, err := testClient.CreateUsersBatch(context.Background(), []AddUserPayload{{Username: "a"}})
	if err == nil {
		t.Fatal("expected an unmarshal error for a malformed body")
	}
}

// TestGetUserServerErrorExhaustsRetries confirms that an idempotent GET against a
// persistently failing (5xx) endpoint retries per config and then surfaces the
// last status as an *APIError. Retries are capped low so the test stays fast.
func TestGetUserServerErrorExhaustsRetries(t *testing.T) {
	setup()
	defer teardown()

	var calls int
	testMux.HandleFunc("/api-public/v1/user/flaky", func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"boom"}`))
	})

	client := NewClientWithArgs("apiID", "apiKey", testServer.URL, ClientArgs{
		RateLimit:   1000,
		RetryConfig: &RetryConfig{MaxRetries: 2, InitialBackoff: 1_000_000, MaxBackoff: 5_000_000, BackoffMultiplier: 2, RetryableStatus: []int{500}},
	})

	_, details, err := client.GetUser(context.Background(), "flaky")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 500 chain, got %v", err)
	}
	if details.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", details.StatusCode)
	}
	if details.ErrorCategory != "server_error" {
		t.Errorf("expected ErrorCategory server_error, got %q", details.ErrorCategory)
	}
	// 1 initial attempt + 2 retries.
	if calls != 3 {
		t.Errorf("expected 3 attempts, got %d", calls)
	}
}
