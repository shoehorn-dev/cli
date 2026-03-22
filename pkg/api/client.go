// Package api provides a thin HTTP client for the Shoehorn API
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shoehorn-dev/cli/pkg/config"
)

// maxRedirects is the maximum number of HTTP redirects the client will follow.
const maxRedirects = 3

// loadConfig is a package-level alias to avoid import cycles
var loadConfig = config.Load

// maxResponseSize is the maximum allowed API response body size (10 MB).
// Prevents denial-of-service from malicious or compromised servers.
const maxResponseSize = 10 * 1024 * 1024

// Client is a thin HTTP client for the Shoehorn API. It handles JSON
// serialization, Bearer-token authentication, and structured error responses.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new API client pointed at baseURL with a 30-second
// HTTP timeout, TLS 1.2 minimum, and redirect protection that strips the
// Authorization header on cross-origin redirects to prevent credential leaks (S1, S3).
func NewClient(baseURL string) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= maxRedirects {
					return fmt.Errorf("stopped after %d redirects", maxRedirects)
				}
				// Strip Authorization header on cross-origin redirects to prevent
				// credential leaks to third-party servers (security finding S1).
				// Same-origin redirects preserve the header.
				if req.URL.Host != via[0].URL.Host {
					req.Header.Del("Authorization")
				}
				return nil
			},
		},
	}
}

// SetToken sets the Bearer token sent in the Authorization header of
// every subsequent request. Pass an empty string to clear the token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken returns the current Bearer token, or an empty string if none is set.
func (c *Client) GetToken() string {
	return c.token
}

// do executes an HTTP request and handles the response
func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body with size limit to prevent DoS from oversized responses
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if int64(len(respBody)) > maxResponseSize {
		return fmt.Errorf("response too large (>%d bytes)", maxResponseSize)
	}

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			// Couldn't parse error response -- truncate raw body to avoid
			// leaking server internals (stack traces, paths, etc). S8.
			body := string(respBody)
			if len(body) > 200 {
				body = body[:200] + "... (truncated)"
			}
			return NewAPIError(resp.StatusCode, body, "")
		}
		return NewAPIError(resp.StatusCode, errResp.Error.Message, errResp.Error.Code)
	}

	// Decode success response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// doIgnoreStatus performs an HTTP request and decodes the body into result
// regardless of HTTP status code. Returns the status code alongside any error.
func (c *Client) doIgnoreStatus(ctx context.Context, method, path string, body, result any) (int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	if int64(len(respBody)) > maxResponseSize {
		return resp.StatusCode, fmt.Errorf("response too large (>%d bytes)", maxResponseSize)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return resp.StatusCode, fmt.Errorf("decode response: %w", err)
		}
	}

	return resp.StatusCode, nil
}

// Get performs a GET request to the given path and decodes the JSON response
// into result. It returns an *APIError for non-2xx status codes.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request to the given path, encoding body as JSON and
// decoding the JSON response into result. It returns an *APIError for non-2xx
// status codes.
func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request to the given path, encoding body as JSON and
// decoding the JSON response into result. It returns an *APIError for non-2xx
// status codes.
func (c *Client) Put(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request to the given path. It returns an *APIError
// for non-2xx status codes.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// NewClientFromConfig creates an API client from the active configuration
// profile, loading the server URL and access token automatically. It returns
// ErrNotAuthenticated (wrapped) when no valid credentials are found in the
// profile, prompting the caller to run "shoehorn auth login".
func NewClientFromConfig() (*Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if !cfg.IsAuthenticated() {
		return nil, fmt.Errorf("%w — run: shoehorn auth login --token <PAT>", ErrNotAuthenticated)
	}
	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	c := NewClient(profile.Server)
	c.SetToken(profile.Auth.AccessToken)
	return c, nil
}

// ErrorResponse represents the standard JSON error envelope returned by the
// Shoehorn API on non-2xx responses.
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}
