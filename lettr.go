// Package lettr provides a Go client for the Lettr Email API.
//
// Create a client with your API key, then use the service objects to interact
// with the API:
//
//	client := lettr.NewClient("your-api-key")
//	sent, err := client.Emails.Send(ctx, &lettr.SendEmailRequest{
//	    From:    "sender@example.com",
//	    To:      []string{"recipient@example.com"},
//	    Subject: "Hello from Lettr",
//	    Html:    "<h1>Hello!</h1>",
//	})
package lettr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Version is the current version of this SDK.
	Version = "0.1.0"

	defaultBaseURL = "https://app.lettr.com/api/"
	userAgent      = "lettr-go/" + Version
	contentType    = "application/json"
)

// Client manages communication with the Lettr API.
type Client struct {
	// httpClient is the underlying HTTP client used for API requests.
	httpClient *http.Client

	// apiKey is the bearer token used for authentication.
	apiKey string

	// baseURL is the base URL for API requests.
	baseURL *url.URL

	// userAgent is the User-Agent header sent with each request.
	userAgent string

	// Services for different API resources.
	Emails    *EmailService
	Domains   *DomainService
	Webhooks  *WebhookService
	Templates *TemplateService
}

// NewClient creates a new Lettr API client with the given API key.
// It uses a default HTTP client with a 30-second timeout.
func NewClient(apiKey string) *Client {
	return NewClientWithHTTPClient(apiKey, &http.Client{
		Timeout: 30 * time.Second,
	})
}

// NewClientWithHTTPClient creates a new Lettr API client with a custom HTTP client.
// This is useful for testing or for configuring custom timeouts and transports.
func NewClientWithHTTPClient(apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(apiKey),
		baseURL:    baseURL,
		userAgent:  userAgent,
	}

	c.Emails = &EmailService{client: c}
	c.Domains = &DomainService{client: c}
	c.Webhooks = &WebhookService{client: c}
	c.Templates = &TemplateService{client: c}

	return c
}

// SetBaseURL overrides the default base URL. Useful for testing against
// a mock server.
func (c *Client) SetBaseURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	c.baseURL = u
	return nil
}

// newRequest builds an HTTP request for the Lettr API.
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("lettr: invalid path %q: %w", path, err)
	}

	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("lettr: failed to marshal request body: %w", err)
		}
		buf = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", contentType)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// do sends an HTTP request and decodes the JSON response into v.
// It returns the raw HTTP response and any error encountered.
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lettr: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, parseError(resp)
	}

	if v != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return resp, fmt.Errorf("lettr: failed to decode response: %w", err)
		}
	}

	return resp, nil
}

// HealthCheck verifies that the Lettr API is reachable.
func (c *Client) HealthCheck(ctx context.Context) (*HealthCheckResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "health", nil)
	if err != nil {
		return nil, err
	}

	// Health check doesn't require auth, remove the header.
	req.Header.Del("Authorization")

	var resp HealthCheckResponse
	if _, err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ValidateAPIKey checks whether the configured API key is valid and returns
// the associated team information.
func (c *Client) ValidateAPIKey(ctx context.Context) (*AuthCheckResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "auth/check", nil)
	if err != nil {
		return nil, err
	}

	var resp AuthCheckResponse
	if _, err := c.do(req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// HealthCheckResponse is the response from the health check endpoint.
type HealthCheckResponse struct {
	Message string          `json:"message"`
	Data    HealthCheckData `json:"data"`
}

// HealthCheckData contains the health check status information.
type HealthCheckData struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// AuthCheckResponse is the response from the auth check endpoint.
type AuthCheckResponse struct {
	Message string        `json:"message"`
	Data    AuthCheckData `json:"data"`
}

// AuthCheckData contains the auth check details.
type AuthCheckData struct {
	TeamID    int    `json:"team_id"`
	Timestamp string `json:"timestamp"`
}
