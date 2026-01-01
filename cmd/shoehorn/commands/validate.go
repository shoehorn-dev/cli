package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	validateInput  string
	validateFormat string
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate Shoehorn or Backstage manifest files",
	Long: `Validate manifest files and output structured validation errors.

Supports both Shoehorn and Backstage manifest formats with automatic detection.

Examples:
  # Validate a manifest file (text output)
  shoehorn validate catalog-info.yaml

  # Validate with JSON output
  shoehorn validate .shoehorn/service.yml --format json

  # Validate from stdin
  cat catalog-info.yaml | shoehorn validate -`,
	Args: cobra.MaximumNArgs(1),
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().StringVar(&validateFormat, "format", "text", "output format: text or json")
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Determine input source
	var content string
	var filename string

	if len(args) == 0 || args[0] == "-" {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = string(data)
		filename = "stdin"
	} else {
		// Read from file
		filename = args[0]
		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content = string(data)
	}

	// Load config and create API client
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current profile
	currentProfile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	// Create API client
	client := api.NewClient(currentProfile.Server)

	// Set auth token if available
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'shoehorn auth login' first")
	}
	client.SetToken(currentProfile.Auth.AccessToken)

	// Call API
	ctx := context.Background()
	result, err := client.ValidateManifest(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to validate manifest: %w", err)
	}

	// Output based on format
	if validateFormat == "json" {
		output := map[string]interface{}{
			"file":   filename,
			"valid":  result.Valid,
			"errors": result.Errors,
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// Text output
		if result.Valid {
			fmt.Printf("✓ %s is valid\n", filename)
		} else {
			fmt.Printf("✗ %s has validation errors:\n\n", filename)
			for _, err := range result.Errors {
				if err.Field != "" {
					fmt.Printf("  - %s: %s\n", err.Field, err.Message)
				} else {
					fmt.Printf("  - %s\n", err.Message)
				}
			}
			return fmt.Errorf("validation failed")
		}
	}

	return nil
}
