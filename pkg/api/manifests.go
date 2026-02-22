package api

import (
	"context"
)

// ManifestValidationResult represents the validation result from the API
type ManifestValidationResult struct {
	Valid  bool                      `json:"valid"`
	Errors []ManifestValidationError `json:"errors"`
}

// ManifestValidationError represents a validation error
type ManifestValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ManifestConversionRequest represents a conversion request
type ManifestConversionRequest struct {
	Content    string `json:"content"`
	TargetType string `json:"targetType"` // "shoehorn", "backstage", or "mold"
	Validate   bool   `json:"validate"`
}

// ManifestConversionResponse represents the conversion response
type ManifestConversionResponse struct {
	Success    bool                      `json:"success"`
	Content    string                    `json:"content,omitempty"` // For shoehorn/backstage
	Mold       map[string]interface{}    `json:"mold,omitempty"`    // For mold
	Format     string                    `json:"format"`
	Validation *ManifestValidationResult `json:"validation,omitempty"`
}

// ValidateManifestRequest represents a validation request
type ValidateManifestRequest struct {
	Content string `json:"content"`
}

// ValidateManifestResponse represents the validation response
type ValidateManifestResponse struct {
	Valid  bool                      `json:"valid"`
	Errors []ManifestValidationError `json:"errors"`
}

// ValidateManifest validates a manifest file via the API
func (c *Client) ValidateManifest(ctx context.Context, content string) (*ValidateManifestResponse, error) {
	req := ValidateManifestRequest{
		Content: content,
	}

	var resp ValidateManifestResponse
	if err := c.Post(ctx, "/api/v1/manifests/validate", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ConvertManifest converts a manifest between formats via the API
func (c *Client) ConvertManifest(ctx context.Context, content string, targetType string, validate bool) (*ManifestConversionResponse, error) {
	req := ManifestConversionRequest{
		Content:    content,
		TargetType: targetType,
		Validate:   validate,
	}

	var resp ManifestConversionResponse
	if err := c.Post(ctx, "/api/v1/manifests/convert", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
