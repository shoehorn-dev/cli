// Package api provides a thin HTTP client for the Shoehorn API
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/imbabamba/shoehorn-cli/pkg/config"
)

// loadConfig is a package-level alias to avoid import cycles
var loadConfig = config.Load

// Client is a thin HTTP client for the Shoehorn API
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the Bearer token for authenticated requests
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken returns the current token
func (c *Client) GetToken() string {
	return c.token
}

// do executes an HTTP request and handles the response
func (c *Client) do(ctx context.Context, method, path string, body, result interface{}) error {
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

	// Read response body for error messages
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err != nil {
			// Couldn't parse error response, return raw status
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error.Message)
	}

	// Decode success response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body, result interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// NewClientFromConfig creates an API client from the current config profile.
// Returns an error if not authenticated.
func NewClientFromConfig() (*Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if !cfg.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated — run: shoehorn auth login --token <PAT>")
	}
	profile, err := cfg.GetCurrentProfile()
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	c := NewClient(profile.Server)
	c.SetToken(profile.Auth.AccessToken)
	return c, nil
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}
