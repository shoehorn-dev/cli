package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shoehorn-dev/cli/pkg/addon"
	"github.com/spf13/cobra"
)

var addonBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build addon TypeScript into a JS bundle",
	Long: `Compile the addon TypeScript source into a single JS bundle using esbuild.

Run this from the addon project directory (where package.json is).
Requires esbuild (installed via npm install).

Output: dist/addon.js`,
	RunE: runAddonBuild,
}

func runAddonBuild(_ *cobra.Command, _ []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if err := addon.ValidateBuildPrereqs(workDir); err != nil {
		return err
	}

	// Run npm run build (which invokes esbuild)
	cmd := exec.Command("npm", "run", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Validate and checksum the output
	bundlePath := filepath.Join(workDir, "dist", "addon.js")
	result, err := addon.ValidateBundle(bundlePath)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Build complete: %s\n", result.Path)
	fmt.Printf("  Size:   %s\n", result.SizeFormatted)
	fmt.Printf("  SHA256: %s\n", result.SHA256)

	return nil
}

func init() {
	addonCmd.AddCommand(addonBuildCmd)
}
