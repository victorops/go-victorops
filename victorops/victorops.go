package victorops

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

// Client is the main client for interacting with victorops
type Client struct {
	publicBaseURL string
	apiID         string
	apiKey        string
	httpClient    http.Client
}

// Client args is used to dynamically pass in parameters when instantiating the Client
type ClientArgs struct {
	timeoutSeconds int
}

// RequestDetails contains details from the API response
type RequestDetails struct {
	StatusCode   int
	ResponseBody string
	RequestBody  string
	RawResponse  *http.Response
	RawRequest   *http.Request
}

func (c Client) String() string {
	return fmt.Sprintf("VictorOps Client: publicBaseURL: %s ", c.publicBaseURL)
}

func (c Client) makePublicAPICall(method string, endpoint string, requestBody io.Reader, queryParams map[string]string) (*RequestDetails, error) {
	details := RequestDetails{}
	// Create the request
	req, err := http.NewRequest(method, c.publicBaseURL+"/api-public/"+endpoint, requestBody)
	if err != nil {
		return &details, err
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
		return &details, err
	}
	details.RequestBody = string(requestDump)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &details, err
	}

	// Read the entire response
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &details, err
	}

	details.StatusCode = resp.StatusCode
	details.ResponseBody = string(responseBody)
	details.RawResponse = resp

	return &details, nil
}

// NewClient creates a new VictorOps client
func NewClient(apiID string, apiKey string, publicBaseURL string) *Client {
	return NewConfigurableClient(apiID, apiKey, publicBaseURL, http.Client{Timeout: 30})
}

// NewConfigurableClient creates a new VictorOps client with ClientArgs struct
func NewConfigurableClient(apiID string, apiKey string, publicBaseURL string, httpClient http.Client) *Client {
	client := Client{
		apiID:         apiID,
		apiKey:        apiKey,
		publicBaseURL: publicBaseURL,
		httpClient:    httpClient,
	}
	return &client
}

// GetHTTPClient returns http client for the purpose of test
func (c Client) GetHTTPClient() *http.Client {
	return &c.httpClient
}
