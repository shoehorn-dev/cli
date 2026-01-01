package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	convertOutput     string
	convertOutputType string
	convertValidate   bool
	convertRecursive  bool
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert [file]",
	Short: "Convert between Backstage and Shoehorn manifest formats",
	Long: `Convert manifest files between Backstage and Shoehorn formats.

Examples:
  # Convert a Backstage manifest to Shoehorn format
  shoehorn convert catalog-info.yaml

  # Convert and save to file
  shoehorn convert catalog-info.yaml -o .shoehorn/my-service.yml

  # Convert Shoehorn to Backstage format
  shoehorn convert .shoehorn/my-service.yml --to backstage

  # Convert Backstage Template to Shoehorn Mold (outputs JSON)
  shoehorn convert template.yaml --to mold -o mold.json

  # Convert all manifests in a directory
  shoehorn convert ./manifests -r

  # Validate during conversion
  shoehorn convert catalog-info.yaml --validate`,
	Args: cobra.ExactArgs(1),
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().StringVarP(&convertOutput, "output", "o", "", "output file (default: stdout)")
	convertCmd.Flags().StringVar(&convertOutputType, "to", "shoehorn", "output format: shoehorn, backstage, or mold")
	convertCmd.Flags().BoolVar(&convertValidate, "validate", false, "validate manifest after conversion")
	convertCmd.Flags().BoolVarP(&convertRecursive, "recursive", "r", false, "recursively process directories")
	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) error {
	inputPath := args[0]

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

	// Check if input is a directory
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat input: %w", err)
	}

	if fileInfo.IsDir() {
		if !convertRecursive {
			return fmt.Errorf("input is a directory - use -r flag for recursive processing")
		}
		return convertDirectory(client, inputPath)
	}

	// Single file conversion
	return convertFile(client, inputPath, convertOutput)
}

func convertDirectory(client *api.Client, dirPath string) error {
	var filesProcessed int
	var filesFailed int

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process YAML files
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		// Determine output path
		var outputPath string
		if convertOutput != "" {
			// If output is specified, use it as base directory
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}
			outputPath = filepath.Join(convertOutput, relPath)
		}

		fmt.Printf("Converting %s...\n", path)
		if err := convertFile(client, path, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Failed: %v\n", err)
			filesFailed++
			return nil // Continue processing other files
		}

		filesProcessed++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	fmt.Printf("\nProcessed %d files, %d failed\n", filesProcessed, filesFailed)
	if filesFailed > 0 {
		return fmt.Errorf("%d files failed to convert", filesFailed)
	}

	return nil
}

func convertFile(client *api.Client, inputPath string, outputPath string) error {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Call API
	ctx := context.Background()
	result, err := client.ConvertManifest(ctx, string(data), convertOutputType, convertValidate)
	if err != nil {
		return fmt.Errorf("failed to convert manifest: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("conversion failed - manifest is invalid")
	}

	// Prepare output
	var outputData []byte
	if convertOutputType == "mold" {
		// Mold format is JSON
		outputData, err = json.MarshalIndent(result.Mold, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal mold: %w", err)
		}
	} else {
		// Shoehorn and Backstage formats are YAML
		outputData = []byte(result.Content)
	}

	// Write output
	if outputPath != "" {
		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("  ✓ Wrote to %s\n", outputPath)
	} else {
		// Write to stdout
		fmt.Println(string(outputData))
	}

	// Show validation results if requested
	if convertValidate && result.Validation != nil {
		if !result.Validation.Valid {
			fmt.Println("\nValidation errors:")
			for _, err := range result.Validation.Errors {
				if err.Field != "" {
					fmt.Printf("  - %s: %s\n", err.Field, err.Message)
				} else {
					fmt.Printf("  - %s\n", err.Message)
				}
			}
		} else {
			fmt.Println("  ✓ Validation passed")
		}
	}

	return nil
}
