package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/spf13/cobra"
)

var addonPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish addon to the marketplace",
	Long: `Publish the current addon to your Shoehorn instance's marketplace.

Reads manifest.json from the current directory and uploads it along with
any built bundles (dist/addon.js, dist/frontend.js).
Run this from the addon project directory.`,
	RunE: runAddonPublish,
}

func runAddonPublish(_ *cobra.Command, _ []string) error {
	// Read manifest.json
	manifestData, err := os.ReadFile("manifest.json")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no manifest.json found - run this from an addon project directory")
		}
		return fmt.Errorf("read manifest.json: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("invalid manifest.json: %w", err)
	}

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	// Step 1: Publish manifest
	result, spinErr := tui.RunSpinner("Publishing manifest...", func() (any, error) {
		return client.PublishAddonManifest(context.Background(), manifest)
	})
	if spinErr != nil {
		return fmt.Errorf("publish addon: %w", spinErr)
	}

	pub := result.(*api.PublishResult)

	action := "updated"
	if pub.Created {
		action = "published"
	}

	fmt.Printf("Addon %q %s successfully.\n", pub.Slug, action)
	fmt.Printf("  Name: %s\n", pub.Name)
	if pub.Installed {
		fmt.Println("  Auto-installed for your tenant.")
	}

	// Step 2: Upload bundles if they exist
	bundles := map[string][]byte{}

	if data, err := os.ReadFile(filepath.Join("dist", "addon.js")); err == nil {
		bundles["backend"] = data
	}
	if data, err := os.ReadFile(filepath.Join("dist", "frontend.js")); err == nil {
		bundles["frontend"] = data
	}

	if len(bundles) > 0 {
		uploadResult, uploadErr := tui.RunSpinner("Uploading bundles...", func() (any, error) {
			return client.UploadAddonBundle(context.Background(), pub.Slug, bundles)
		})
		if uploadErr != nil {
			return fmt.Errorf("upload bundles: %w", uploadErr)
		}

		upload := uploadResult.(*api.BundleUploadResult)
		for name, size := range upload.Uploaded {
			fmt.Printf("  Bundle %s: %d bytes uploaded\n", name, size)
		}
	}

	return nil
}

func init() {
	addonCmd.AddCommand(addonPublishCmd)
}
