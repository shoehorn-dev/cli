package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

const maxBundleSize = 2 * 1024 * 1024 // 2MB

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
	// Verify we're in an addon project
	if _, err := os.Stat("manifest.json"); os.IsNotExist(err) {
		return fmt.Errorf("no manifest.json found - run this from an addon project directory")
	}
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		return fmt.Errorf("no package.json found - declarative addons don't need building")
	}

	// Run npm run build (which invokes esbuild)
	cmd := exec.Command("npm", "run", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Verify output exists
	bundlePath := filepath.Join("dist", "addon.js")
	info, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("build output not found at %s", bundlePath)
	}

	// Check bundle size
	if info.Size() > maxBundleSize {
		return fmt.Errorf("bundle size %d bytes exceeds maximum %d bytes (2MB)", info.Size(), maxBundleSize)
	}

	// Compute SHA256
	content, err := os.ReadFile(bundlePath)
	if err != nil {
		return fmt.Errorf("read bundle: %w", err)
	}
	hash := sha256.Sum256(content)
	checksum := hex.EncodeToString(hash[:])

	fmt.Println()
	fmt.Printf("Build complete: %s\n", bundlePath)
	fmt.Printf("  Size:   %s\n", formatBuildSize(info.Size()))
	fmt.Printf("  SHA256: %s\n", checksum)

	return nil
}

func formatBuildSize(bytes int64) string {
	const kb = 1024
	if bytes < kb {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
}

func init() {
	addonCmd.AddCommand(addonBuildCmd)
}
