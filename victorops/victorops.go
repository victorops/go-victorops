package victorops

import (
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

// Default configuration values
const (
	DefaultRateLimit       = 2              // 2 requests per second (per API spec)
	DefaultMaxRetries      = 3              // Maximum retry attempts
	DefaultInitialBackoff  = 100 * time.Millisecond
	DefaultMaxBackoff      = 30 * time.Second
	DefaultBackoffMultiple = 2.0
)

// RetryConfig configures the retry behavior for transient errors
type RetryConfig struct {
	MaxRetries        int           // Maximum number of retry attempts
	InitialBackoff    time.Duration // Initial backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	BackoffMultiplier float64       // Multiplier for exponential backoff
	RetryableStatus   []int         // HTTP status codes that trigger a retry
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        DefaultMaxRetries,
		InitialBackoff:    DefaultInitialBackoff,
		MaxBackoff:        DefaultMaxBackoff,
		BackoffMultiplier: DefaultBackoffMultiple,
		RetryableStatus:   []int{429, 500, 502, 503, 504},
	}
}

// Client is the main client for interacting with victorops
type Client struct {
	publicBaseURL string
	apiID         string
	apiKey        string
	httpClient    http.Client
	rateLimiter   *rate.Limiter
	retryConfig   RetryConfig
}

// ClientArgs is used to dynamically pass in parameters when instantiating the Client
type ClientArgs struct {
	TimeoutSeconds int
	RateLimit      float64     // Requests per second (0 uses default of 2/sec)
	RetryConfig    RetryConfig // Retry configuration (zero value uses defaults)
}

// RequestDetails contains details from the API response
type RequestDetails struct {
	StatusCode    int
	ResponseBody  string
	RequestBody   string
	RawResponse   *http.Response
	RawRequest    *http.Request
	RetryCount    int           // Number of retries performed
	RateLimited   bool          // Whether the request was rate limited
	RetryAfter    time.Duration // Retry-After duration if provided by server
	ErrorCategory string        // Category: "rate_limit", "server_error", "client_error", "network"
}

func (c Client) String() string {
	return fmt.Sprintf("VictorOps Client: publicBaseURL: %s ", c.publicBaseURL)
}

// isRetryable checks if the status code should trigger a retry
func (c *Client) isRetryable(statusCode int) bool {
	for _, code := range c.retryConfig.RetryableStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}

// calculateBackoff returns the backoff duration for the given attempt
func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.retryConfig.InitialBackoff) * math.Pow(c.retryConfig.BackoffMultiplier, float64(attempt))
	if backoff > float64(c.retryConfig.MaxBackoff) {
		backoff = float64(c.retryConfig.MaxBackoff)
	}
	return time.Duration(backoff)
}

// parseRetryAfter extracts the Retry-After duration from response headers
func parseRetryAfter(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// Try parsing as seconds
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date
	if t, err := http.ParseTime(retryAfter); err == nil {
		return time.Until(t)
	}

	return 0
}

// categorizeError returns the error category based on status code
func categorizeError(statusCode int, err error) string {
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "no such host") ||
			strings.Contains(errStr, "network") {
			return "network"
		}
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

// makePublicAPICall makes an API call with rate limiting and retry logic
func (c *Client) makePublicAPICall(ctx context.Context, method string, endpoint string, requestBody io.Reader, queryParams map[string]string) (*RequestDetails, error) {
	details := &RequestDetails{}

	// Read the request body into a buffer so we can retry
	var bodyBytes []byte
	if requestBody != nil {
		var err error
		bodyBytes, err = io.ReadAll(requestBody)
		if err != nil {
			details.ErrorCategory = "client_error"
			return details, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			details.ErrorCategory = "rate_limit"
			return details, fmt.Errorf("rate limiter error: %w", err)
		}

		// Create a new reader from the body bytes for each attempt
		var body io.Reader
		if bodyBytes != nil {
			body = strings.NewReader(string(bodyBytes))
		}

		// Create the request with context
		req, err := http.NewRequestWithContext(ctx, method, c.publicBaseURL+"/api-public/"+endpoint, body)
		if err != nil {
			details.ErrorCategory = "client_error"
			return details, err
		}

		// Set the auth headers needed for the public api
		req.Header.Set("X-VO-Api-Id", c.apiID)
		req.Header.Set("X-VO-Api-Key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		// Set the query params
		q := req.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		req.URL.RawQuery = q.Encode()

		// Add the request to the details
		details.RawRequest = req
		requestDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			details.ErrorCategory = "client_error"
			return details, err
		}
		details.RequestBody = string(requestDump)

		// Make the request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			details.ErrorCategory = categorizeError(0, err)

			// Check if context was cancelled
			if ctx.Err() != nil {
				return details, ctx.Err()
			}

			// Network error - retry if we have attempts left
			if attempt < c.retryConfig.MaxRetries {
				details.RetryCount = attempt + 1
				backoff := c.calculateBackoff(attempt)
				select {
				case <-ctx.Done():
					return details, ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}
			return details, err
		}

		lastResp = resp

		// Read the entire response
		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			details.ErrorCategory = "network"
			return details, err
		}

		details.StatusCode = resp.StatusCode
		details.ResponseBody = string(responseBody)
		details.RawResponse = resp
		details.ErrorCategory = categorizeError(resp.StatusCode, nil)

		// Check if we should retry
		if c.isRetryable(resp.StatusCode) && attempt < c.retryConfig.MaxRetries {
			details.RetryCount = attempt + 1

			// Check for rate limiting
			if resp.StatusCode == 429 {
				details.RateLimited = true
				details.RetryAfter = parseRetryAfter(resp)
			}

			// Calculate backoff (use Retry-After if provided)
			backoff := c.calculateBackoff(attempt)
			if details.RetryAfter > 0 && details.RetryAfter > backoff {
				backoff = details.RetryAfter
			}

			select {
			case <-ctx.Done():
				return details, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}

		// Success or non-retryable error
		return details, nil
	}

	// All retries exhausted
	if lastErr != nil {
		return details, lastErr
	}
	if lastResp != nil {
		details.StatusCode = lastResp.StatusCode
		details.ErrorCategory = categorizeError(lastResp.StatusCode, nil)
	}
	return details, nil
}

// NewClient creates a new VictorOps client with default configuration
func NewClient(apiID string, apiKey string, publicBaseURL string) *Client {
	return NewConfigurableClient(apiID, apiKey, publicBaseURL, http.Client{Timeout: time.Second * 30})
}

// NewConfigurableClient creates a new VictorOps client with custom http.Client
func NewConfigurableClient(apiID string, apiKey string, publicBaseURL string, httpClient http.Client) *Client {
	return NewClientWithArgs(apiID, apiKey, publicBaseURL, ClientArgs{
		TimeoutSeconds: int(httpClient.Timeout.Seconds()),
		RateLimit:      DefaultRateLimit,
		RetryConfig:    DefaultRetryConfig(),
	})
}

// NewClientWithArgs creates a new VictorOps client with full configuration
func NewClientWithArgs(apiID string, apiKey string, publicBaseURL string, args ClientArgs) *Client {
	// Set defaults
	rateLimit := args.RateLimit
	if rateLimit <= 0 {
		rateLimit = DefaultRateLimit
	}

	retryConfig := args.RetryConfig
	if retryConfig.MaxRetries == 0 && retryConfig.InitialBackoff == 0 {
		retryConfig = DefaultRetryConfig()
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

// GetHTTPClient returns http client for the purpose of test
func (c *Client) GetHTTPClient() *http.Client {
	return &c.httpClient
}

// GetRateLimiter returns the rate limiter for testing purposes
func (c *Client) GetRateLimiter() *rate.Limiter {
	return c.rateLimiter
}

// GetRetryConfig returns the retry configuration for testing purposes
func (c *Client) GetRetryConfig() RetryConfig {
	return c.retryConfig
}
