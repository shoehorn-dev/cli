package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/imbabamba/shoehorn-cli/pkg/addon"
	"github.com/spf13/cobra"
)

var addonDevCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start addon development mode with watch and rebuild",
	Long: `Start esbuild in watch mode, rebuilding the addon bundle on every file change.

Run this from the addon project directory (where package.json is).
Requires esbuild (installed via npm install).

Uses "npm run dev" which invokes esbuild with --watch.
Press Ctrl+C to stop.`,
	RunE: runAddonDev,
}

func runAddonDev(_ *cobra.Command, _ []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if err := addon.ValidateBuildPrereqs(workDir); err != nil {
		return err
	}

	fmt.Println("Starting addon dev mode (esbuild --watch)...")
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	cmd := exec.Command("npm", "run", "dev")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Forward interrupt signal to child process for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start dev server: %w", err)
	}

	// Wait for either process exit or signal
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case err := <-doneCh:
		if err != nil {
			return fmt.Errorf("dev server exited: %w", err)
		}
		return nil
	case sig := <-sigCh:
		fmt.Printf("\nReceived %s, stopping dev server...\n", sig)
		if cmd.Process != nil {
			cmd.Process.Signal(sig)
		}
		<-doneCh // Wait for process to exit
		return nil
	}
}

func init() {
	addonCmd.AddCommand(addonDevCmd)
}
