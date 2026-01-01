package api

import (
	"context"
	"time"
)

// DeviceInitResponse contains device flow initialization data
type DeviceInitResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DevicePollRequest is the request for polling device flow status
type DevicePollRequest struct {
	DeviceCode string `json:"device_code"`
}

// DevicePollResponse contains device flow poll result
type DevicePollResponse struct {
	// Success response
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresIn    int       `json:"expires_in,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	User         *UserInfo `json:"user,omitempty"`

	// Pending response
	Pending bool   `json:"pending,omitempty"`
	Message string `json:"message,omitempty"`
}

// UserInfo contains authenticated user information
type UserInfo struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
}

// RefreshTokenRequest is the request for refreshing an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse contains refreshed token data
type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthStatusResponse contains current authentication status
type AuthStatusResponse struct {
	Authenticated bool      `json:"authenticated"`
	User          *UserInfo `json:"user,omitempty"`
	ExpiresAt     time.Time `json:"expires_at,omitempty"`
}

// InitDeviceFlow initiates the OAuth2 device authorization flow
func (c *Client) InitDeviceFlow(ctx context.Context) (*DeviceInitResponse, error) {
	var resp DeviceInitResponse
	err := c.Post(ctx, "/api/v1/auth/cli/device-init", nil, &resp)
	return &resp, err
}

// PollDeviceFlow polls for device flow completion
func (c *Client) PollDeviceFlow(ctx context.Context, deviceCode string) (*DevicePollResponse, error) {
	req := DevicePollRequest{DeviceCode: deviceCode}
	var resp DevicePollResponse
	err := c.Post(ctx, "/api/v1/auth/cli/device-poll", req, &resp)
	return &resp, err
}

// RefreshToken refreshes an access token
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	req := RefreshTokenRequest{RefreshToken: refreshToken}
	var resp RefreshTokenResponse
	err := c.Post(ctx, "/api/v1/auth/cli/refresh", req, &resp)
	return &resp, err
}

// GetAuthStatus returns current authentication status (requires valid Bearer token)
func (c *Client) GetAuthStatus(ctx context.Context) (*AuthStatusResponse, error) {
	var resp AuthStatusResponse
	err := c.Get(ctx, "/api/v1/auth/cli/status", &resp)
	return &resp, err
}
