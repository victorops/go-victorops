package victorops

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// newFastEngineClient returns a server+client wired for engine-level tests. The
// rate limit is raised and the backoff shrunk so retry paths execute quickly.
func newFastEngineClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	mux.HandleFunc("/api-public/", handler)

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
	return server, client
}

func TestEngineRetriesIdempotentServerError(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 server calls (2 failures + 1 success), got %d", got)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected final status 200, got %d", details.StatusCode)
	}
	if details.RetryCount != 2 {
		t.Errorf("expected RetryCount 2, got %d", details.RetryCount)
	}
}

func TestEngineRetriesExhausted(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError after exhausting retries, got %v", err)
	}
	// MaxRetries=3 means the initial attempt plus 3 retries.
	if got := atomic.LoadInt32(&calls); got != 4 {
		t.Errorf("expected 4 server calls, got %d", got)
	}
	if details.StatusCode != http.StatusBadGateway {
		t.Errorf("expected final status 502, got %d", details.StatusCode)
	}
	if details.ErrorCategory != "server_error" {
		t.Errorf("expected ErrorCategory server_error, got %q", details.ErrorCategory)
	}
}

func TestEngineRetryAfterHeader(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !details.RateLimited {
		t.Errorf("expected RateLimited to be true")
	}
	if details.RetryAfter != time.Second {
		t.Errorf("expected RetryAfter 1s, got %v", details.RetryAfter)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected final status 200, got %d", details.StatusCode)
	}
}

// TestEngineRecordsRateLimitOnPost verifies that a non-idempotent method (POST)
// receiving a 429 still surfaces RateLimited/RetryAfter, even though it is not
// retried.
func TestEngineRecordsRateLimitOnPost(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "POST", "v1/team", bytes.NewBufferString("{}"), nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 429 POST, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 call for non-idempotent POST, got %d", got)
	}
	if !details.RateLimited {
		t.Errorf("expected RateLimited to be true for a 429 POST")
	}
	if details.RetryAfter != 2*time.Second {
		t.Errorf("expected RetryAfter 2s, got %v", details.RetryAfter)
	}
	if details.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", details.StatusCode)
	}
}

// TestEngineRecordsRateLimitAfterExhaustion verifies that a final 429 returned
// after retries are exhausted still surfaces RateLimited.
func TestEngineRecordsRateLimitAfterExhaustion(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError after retry exhaustion, got %v", err)
	}
	// 1 initial attempt + 3 retries = 4 calls.
	if got := atomic.LoadInt32(&calls); got != 4 {
		t.Errorf("expected 4 calls (1 + 3 retries), got %d", got)
	}
	if !details.RateLimited {
		t.Errorf("expected RateLimited to be true after retry exhaustion")
	}
	if details.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected final status 429, got %d", details.StatusCode)
	}
}

// TestEngineRetryAfterDoesNotLeakToLaterAttempts verifies that a Retry-After
// value from an earlier 429 does not inflate the backoff of a later, unrelated
// retry (e.g. a subsequent 500 that carries no Retry-After header).
func TestEngineRetryAfterDoesNotLeakToLaterAttempts(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&calls, 1) {
		case 1:
			// 1s is large relative to the millisecond backoffs; if it leaked
			// into the next attempt's delay, the test would take ~1s.
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
		case 2:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}
	})
	defer server.Close()

	start := time.Now()
	details, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("expected 3 calls (429 -> 500 -> 200), got %d", got)
	}
	if details.StatusCode != http.StatusOK {
		t.Errorf("expected final status 200, got %d", details.StatusCode)
	}
	// The only long wait should be the single 1s Retry-After from attempt 1.
	// If the value leaked into the 500 retry, total elapsed would be ~2s.
	if elapsed >= 1500*time.Millisecond {
		t.Errorf("Retry-After appears to have leaked into a later retry; elapsed %v", elapsed)
	}
	if elapsed < time.Second {
		t.Errorf("expected the 1s Retry-After to be honored once, elapsed %v", elapsed)
	}
}

func TestEngineDoesNotRetryPost(t *testing.T) {
	var calls int32
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "POST", "v1/team", bytes.NewBufferString("{}"), nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError for a 500 POST, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected exactly 1 server call for non-idempotent POST, got %d", got)
	}
	if details.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", details.StatusCode)
	}
	if details.RetryCount != 0 {
		t.Errorf("expected RetryCount 0 for POST, got %d", details.RetryCount)
	}
}

func TestEngineRedactsCredentials(t *testing.T) {
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
	defer server.Close()

	details, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(details.RequestBody, redactedValue) {
		t.Errorf("expected redacted request dump to contain %q, got:\n%s", redactedValue, details.RequestBody)
	}
	if strings.Contains(details.RequestBody, "secret-id") || strings.Contains(details.RequestBody, "secret-key") {
		t.Errorf("request dump leaked credentials:\n%s", details.RequestBody)
	}
	if got := details.RawRequest.Header.Get("X-VO-Api-Id"); got != redactedValue {
		t.Errorf("RawRequest API ID was not redacted: %q", got)
	}
	if got := details.RawRequest.Header.Get("X-VO-Api-Key"); got != redactedValue {
		t.Errorf("RawRequest API key was not redacted: %q", got)
	}
	if details.RawResponse == nil || details.RawResponse.Request == nil {
		t.Fatal("expected RawResponse.Request diagnostics")
	}
	if got := details.RawResponse.Request.Header.Get("X-VO-Api-Key"); got != redactedValue {
		t.Errorf("RawResponse.Request API key was not redacted: %q", got)
	}
}

type failingReadCloser struct{}

func (failingReadCloser) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (failingReadCloser) Close() error             { return nil }

type bodyFailureTransport struct{ calls *int32 }

func (t bodyFailureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(t.calls, 1)
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       failingReadCloser{},
		Request:    req,
	}, nil
}

func TestEngineRetriesResponseBodyReadFailure(t *testing.T) {
	var calls int32
	client := NewConfigurableClient("id", "key", "http://example.invalid", http.Client{
		Transport: bodyFailureTransport{calls: &calls},
	})
	client.rateLimiter = rate.NewLimiter(rate.Inf, 1)
	client.retryConfig = RetryConfig{
		MaxRetries:        1,
		InitialBackoff:    time.Millisecond,
		MaxBackoff:        time.Millisecond,
		BackoffMultiplier: 1,
	}

	_, err := client.makePublicAPICall(context.Background(), http.MethodGet, "v1/incidents", nil, nil)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected response-body read error, got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("expected one retry after response-body failure, got %d calls", got)
	}
}

type flaggingTransport struct {
	used *int32
	base http.RoundTripper
}

func (t *flaggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.StoreInt32(t.used, 1)
	return t.base.RoundTrip(req)
}

func TestEnginePreservesCustomTransport(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	mux.HandleFunc("/api-public/v1/incidents", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})

	var used int32
	custom := http.Client{Transport: &flaggingTransport{used: &used, base: http.DefaultTransport}}
	client := NewConfigurableClient("id", "key", server.URL, custom)

	if _, err := client.makePublicAPICall(context.Background(), "GET", "v1/incidents", bytes.NewBufferString("{}"), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&used) != 1 {
		t.Errorf("expected the caller-provided Transport to be used")
	}
}

func TestEngineContextCancellation(t *testing.T) {
	server, client := newFastEngineClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
	})
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.makePublicAPICall(ctx, "GET", "v1/incidents", bytes.NewBufferString("{}"), nil)
	if err == nil {
		t.Fatalf("expected an error for a cancelled context")
	}
	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}

func TestEngineConfigGetters(t *testing.T) {
	client := NewClientWithArgs("id", "key", "http://example.invalid", ClientArgs{
		RateLimit:   5,
		RetryConfig: &RetryConfig{MaxRetries: 7},
	})
	if client.GetRateLimiter() == nil {
		t.Errorf("expected a non-nil rate limiter")
	}
	if client.GetRetryConfig().MaxRetries != 7 {
		t.Errorf("expected MaxRetries 7, got %d", client.GetRetryConfig().MaxRetries)
	}
	if client.GetHTTPClient() == nil {
		t.Errorf("expected a non-nil http client")
	}
}
