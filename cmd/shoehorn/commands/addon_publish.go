package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/spf13/cobra"
)

var addonPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish addon to the marketplace",
	Long: `Publish the current addon to your Shoehorn instance's marketplace.

Reads manifest.json from the current directory and uploads it.
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

	result, spinErr := tui.RunSpinner("Publishing addon...", func() (any, error) {
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

	return nil
}

func init() {
	addonCmd.AddCommand(addonPublishCmd)
}
