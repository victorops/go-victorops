package victorops

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Default configuration values for the client.
const (
	DefaultRateLimit       = 2 // requests per second (per VictorOps API guidance)
	DefaultMaxRetries      = 3
	DefaultInitialBackoff  = 100 * time.Millisecond
	DefaultMaxBackoff      = 30 * time.Second
	DefaultBackoffMultiple = 2.0

	redactedValue = "REDACTED"
)

// RetryConfig configures retry behavior for transient failures.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts. Set to 0 to disable retries.
	MaxRetries int
	// InitialBackoff is the backoff duration before the first retry.
	InitialBackoff time.Duration
	// MaxBackoff caps the computed exponential backoff.
	MaxBackoff time.Duration
	// BackoffMultiplier is the exponential growth factor between attempts.
	BackoffMultiplier float64
	// RetryableStatus lists HTTP status codes that trigger a retry.
	RetryableStatus []int
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        DefaultMaxRetries,
		InitialBackoff:    DefaultInitialBackoff,
		MaxBackoff:        DefaultMaxBackoff,
		BackoffMultiplier: DefaultBackoffMultiple,
		RetryableStatus:   []int{429, 500, 502, 503, 504},
	}
}

// Client is the main client for interacting with victorops.
type Client struct {
	publicBaseURL string
	apiID         string
	apiKey        string
	httpClient    http.Client
	rateLimiter   *rate.Limiter
	retryConfig   RetryConfig
}

// ClientArgs is used to dynamically pass in parameters when instantiating the Client.
type ClientArgs struct {
	// TimeoutSeconds is the HTTP client timeout. Values <= 0 use a 30 second default.
	TimeoutSeconds int
	// RateLimit is the client-side request rate in requests per second. Values <= 0 use DefaultRateLimit.
	RateLimit float64
	// RetryConfig is the retry configuration. A nil value uses DefaultRetryConfig; a
	// non-nil value is used verbatim (set MaxRetries to 0 to explicitly disable retries).
	RetryConfig *RetryConfig
}

// RequestDetails contains details from the API response.
type RequestDetails struct {
	StatusCode   int
	ResponseBody string
	RequestBody  string
	RawResponse  *http.Response
	RawRequest   *http.Request
	// RetryCount is the number of retries performed for this call.
	RetryCount int
	// RateLimited is true when the server returned HTTP 429 at least once.
	RateLimited bool
	// RetryAfter is the server-provided Retry-After duration, when present.
	RetryAfter time.Duration
	// ErrorCategory classifies the outcome: "rate_limit", "server_error", "client_error", or "network".
	ErrorCategory string
}

func (c *Client) String() string {
	return fmt.Sprintf("VictorOps Client: publicBaseURL: %s ", c.publicBaseURL)
}

// isIdempotentMethod reports whether an HTTP method is safe to retry automatically.
// POST and PATCH are excluded because retrying them can duplicate side effects when
// the server applied the change but the response was not received.
func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

// isRetryableStatus reports whether the status code is configured as retryable.
func (c *Client) isRetryableStatus(statusCode int) bool {
	for _, code := range c.retryConfig.RetryableStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}

// calculateBackoff returns the backoff duration for the given zero-based attempt.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.retryConfig.InitialBackoff) * math.Pow(c.retryConfig.BackoffMultiplier, float64(attempt))
	if backoff > float64(c.retryConfig.MaxBackoff) {
		backoff = float64(c.retryConfig.MaxBackoff)
	}
	return time.Duration(backoff)
}

// parseRetryAfter extracts the Retry-After duration from response headers, supporting
// both delay-seconds and HTTP-date formats.
func parseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(retryAfter); err == nil {
		return time.Until(t)
	}
	return 0
}

// categorizeError classifies an outcome based on status code and transport error.
func categorizeError(statusCode int, err error) string {
	if err != nil {
		// Errors returned by http.Client.Do are transport failures. Avoid
		// classifying them by message text, which misses EOF, TLS, connection
		// reset, and many wrapped net errors.
		return "network"
	}
	switch {
	case statusCode == 429:
		return "rate_limit"
	case statusCode >= 500:
		return "server_error"
	case statusCode >= 400:
		return "client_error"
	default:
		return ""
	}
}

// dumpRequestRedacted returns an httputil request dump with the VictorOps auth
// headers masked so credentials are never captured in diagnostics or logs.
func dumpRequestRedacted(req *http.Request, bodyBytes []byte) (string, error) {
	dumpReq := cloneRequestRedacted(req, bodyBytes)

	dump, err := httputil.DumpRequestOut(dumpReq, bodyBytes != nil)
	if err != nil {
		return "", err
	}
	return string(dump), nil
}

// cloneRequestRedacted returns a diagnostic copy of req whose authentication
// headers cannot expose the caller's credentials through RequestDetails.
func cloneRequestRedacted(req *http.Request, bodyBytes []byte) *http.Request {
	dumpReq := req.Clone(req.Context())
	dumpReq.Header.Set("X-VO-Api-Id", redactedValue)
	dumpReq.Header.Set("X-VO-Api-Key", redactedValue)
	if bodyBytes != nil {
		dumpReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		dumpReq.Body = nil
	}
	return dumpReq
}

// makePublicAPICall performs a request against the VictorOps public API
// (paths under /api-public/) using the shared, hardened request engine.
func (c *Client) makePublicAPICall(ctx context.Context, method string, endpoint string, requestBody io.Reader, queryParams map[string]string) (*RequestDetails, error) {
	return c.doAPICall(ctx, method, c.publicBaseURL+"/api-public/"+endpoint, requestBody, queryParams)
}

// makeReportingAPICall performs a request against the VictorOps reporting API,
// which lives under a different path prefix (/api-reporting/). It shares the same
// hardened engine as public calls, so reporting requests get the same rate
// limiting, idempotent-only retries, Retry-After handling, and credential
// redaction. The endpoint argument already includes the api-reporting/... prefix.
func (c *Client) makeReportingAPICall(ctx context.Context, method string, endpoint string, requestBody io.Reader, queryParams map[string]string) (*RequestDetails, error) {
	return c.doAPICall(ctx, method, c.publicBaseURL+"/"+endpoint, requestBody, queryParams)
}

// APIError is returned when the VictorOps API responds with a non-2xx status.
// The full RequestDetails (status code, response body, headers) is still
// returned alongside this error so callers can inspect it; APIError guarantees a
// failed HTTP call is never mistaken for success or decoded as a success payload.
type APIError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface, including a truncated response body for
// diagnostics without dumping arbitrarily large payloads into logs.
func (e *APIError) Error() string {
	body := e.Body
	const maxBody = 512
	if len(body) > maxBody {
		body = body[:maxBody] + "…(truncated)"
	}
	return fmt.Sprintf("victorops: HTTP %d: %s", e.StatusCode, body)
}

// isSuccessStatus reports whether an HTTP status code is in the 2xx range.
func isSuccessStatus(code int) bool {
	return code >= 200 && code < 300
}

// doAPICall performs a context-aware request against the given fully-qualified URL
// with client-side rate limiting and bounded exponential-backoff retries for
// transient failures. Retries are only attempted for idempotent HTTP methods.
//
// A 2xx response returns a nil error. Any other final status is returned as an
// *APIError; the RequestDetails is still returned so callers can inspect the
// status code and body.
func (c *Client) doAPICall(ctx context.Context, method string, fullURL string, requestBody io.Reader, queryParams map[string]string) (*RequestDetails, error) {
	details := &RequestDetails{}

	// Buffer the body once so it can be safely replayed across retries.
	var bodyBytes []byte
	if requestBody != nil {
		var err error
		bodyBytes, err = io.ReadAll(requestBody)
		if err != nil {
			details.ErrorCategory = "client_error"
			return details, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	idempotent := isIdempotentMethod(method)

	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Respect the client-side rate limit.
		if err := c.rateLimiter.Wait(ctx); err != nil {
			details.ErrorCategory = "rate_limit"
			return details, fmt.Errorf("rate limiter error: %w", err)
		}

		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
		if err != nil {
			details.ErrorCategory = "client_error"
			return details, err
		}

		req.Header.Set("X-VO-Api-Id", c.apiID)
		req.Header.Set("X-VO-Api-Key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		q := req.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()

		details.RawRequest = cloneRequestRedacted(req, bodyBytes)
		requestDump, err := dumpRequestRedacted(req, bodyBytes)
		if err != nil {
			details.ErrorCategory = "client_error"
			return details, err
		}
		details.RequestBody = requestDump

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			details.ErrorCategory = categorizeError(0, err)

			// Never continue once the context is done.
			if ctx.Err() != nil {
				return details, ctx.Err()
			}

			// Only retry idempotent methods on transport failures.
			if idempotent && attempt < c.retryConfig.MaxRetries {
				select {
				case <-ctx.Done():
					return details, ctx.Err()
				case <-time.After(c.calculateBackoff(attempt)):
					details.RetryCount = attempt + 1
					continue
				}
			}
			return details, err
		}

		lastResp = resp

		responseBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			details.ErrorCategory = "network"
			if ctx.Err() != nil {
				return details, ctx.Err()
			}
			if idempotent && attempt < c.retryConfig.MaxRetries {
				select {
				case <-ctx.Done():
					return details, ctx.Err()
				case <-time.After(c.calculateBackoff(attempt)):
					details.RetryCount = attempt + 1
					continue
				}
			}
			return details, readErr
		}

		details.StatusCode = resp.StatusCode
		details.ResponseBody = string(responseBody)
		diagnosticResp := new(http.Response)
		*diagnosticResp = *resp
		diagnosticResp.Request = details.RawRequest
		details.RawResponse = diagnosticResp
		details.ErrorCategory = categorizeError(resp.StatusCode, nil)

		// Parse this response's Retry-After locally so a stale value from an
		// earlier attempt never influences a later, unrelated retry delay
		// (e.g. a 429 Retry-After:30 followed by a 500 with no header).
		responseRetryAfter := parseRetryAfter(resp)

		// Record rate-limit signals for any 429, regardless of whether the
		// request is retried (non-idempotent method, retries disabled, or a
		// final 429 after exhausting retries all still surface RateLimited).
		if resp.StatusCode == http.StatusTooManyRequests {
			details.RateLimited = true
			details.RetryAfter = responseRetryAfter
		}

		// Retry only idempotent methods on retryable status codes.
		if idempotent && c.isRetryableStatus(resp.StatusCode) && attempt < c.retryConfig.MaxRetries {
			backoff := c.calculateBackoff(attempt)
			if responseRetryAfter > backoff {
				backoff = responseRetryAfter
			}

			select {
			case <-ctx.Done():
				return details, ctx.Err()
			case <-time.After(backoff):
				details.RetryCount = attempt + 1
				continue
			}
		}

		// A 2xx response is success. Any other final status is surfaced as an
		// *APIError so callers never decode an error body as a success payload;
		// details still carries the StatusCode and ResponseBody for inspection.
		if isSuccessStatus(resp.StatusCode) {
			return details, nil
		}
		return details, &APIError{StatusCode: resp.StatusCode, Body: details.ResponseBody}
	}

	// All retries exhausted.
	if lastErr != nil {
		return details, lastErr
	}
	if lastResp != nil {
		details.StatusCode = lastResp.StatusCode
		details.ErrorCategory = categorizeError(lastResp.StatusCode, nil)
		if !isSuccessStatus(lastResp.StatusCode) {
			return details, &APIError{StatusCode: lastResp.StatusCode, Body: details.ResponseBody}
		}
	}
	return details, nil
}

// NewClient creates a new VictorOps client with default configuration and a 30 second timeout.
func NewClient(apiID string, apiKey string, publicBaseURL string) *Client {
	return NewConfigurableClient(apiID, apiKey, publicBaseURL, http.Client{Timeout: time.Second * 30})
}

// NewConfigurableClient creates a new VictorOps client using the caller-supplied http.Client.
// The provided client is preserved in full, including any custom Transport, proxy, TLS
// configuration, redirect policy, cookie jar, and connection pooling. Rate limiting and
// retries use default configuration.
func NewConfigurableClient(apiID string, apiKey string, publicBaseURL string, httpClient http.Client) *Client {
	client := &Client{
		apiID:         apiID,
		apiKey:        apiKey,
		publicBaseURL: publicBaseURL,
		httpClient:    httpClient,
		rateLimiter:   rate.NewLimiter(rate.Limit(DefaultRateLimit), 1),
		retryConfig:   DefaultRetryConfig(),
	}
	return client
}

// NewClientWithArgs creates a new VictorOps client with full configuration control over
// timeout, rate limit, and retry behavior.
func NewClientWithArgs(apiID string, apiKey string, publicBaseURL string, args ClientArgs) *Client {
	rateLimit := args.RateLimit
	if rateLimit <= 0 {
		rateLimit = DefaultRateLimit
	}

	retryConfig := DefaultRetryConfig()
	if args.RetryConfig != nil {
		retryConfig = *args.RetryConfig
	}
	if retryConfig.MaxRetries < 0 {
		retryConfig.MaxRetries = 0
	}

	timeout := time.Duration(args.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client := &Client{
		apiID:         apiID,
		apiKey:        apiKey,
		publicBaseURL: publicBaseURL,
		httpClient:    http.Client{Timeout: timeout},
		rateLimiter:   rate.NewLimiter(rate.Limit(rateLimit), 1),
		retryConfig:   retryConfig,
	}
	return client
}

// GetHTTPClient returns the underlying http client, primarily for tests.
func (c *Client) GetHTTPClient() *http.Client {
	return &c.httpClient
}

// GetRateLimiter returns the client rate limiter, primarily for tests.
func (c *Client) GetRateLimiter() *rate.Limiter {
	return c.rateLimiter
}

// GetRetryConfig returns the client retry configuration, primarily for tests.
func (c *Client) GetRetryConfig() RetryConfig {
	return c.retryConfig
}
