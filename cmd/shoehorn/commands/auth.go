package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	serverURL string
)

// authCmd represents the auth command group
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Manage authentication with the Shoehorn platform.`,
}

// loginCmd represents the auth login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Shoehorn",
	Long: `Authenticate with the Shoehorn platform using OAuth2 device flow.

This command initiates a device authorization flow and opens your browser
to complete authentication.`,
	RunE: runLogin,
}

// statusCmd represents the auth status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  `Display current authentication status and profile information.`,
	RunE:  runStatus,
}

// logoutCmd represents the auth logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from Shoehorn",
	Long:  `Clear local credentials. Note: tokens are not revoked on the server.`,
	RunE:  runLogout,
}

func init() {
	// Add login flags
	loginCmd.Flags().StringVar(&serverURL, "server", "http://localhost:8080", "Shoehorn API server URL")

	// Add commands to auth group
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)

	// Add auth group to root
	rootCmd.AddCommand(authCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	fmt.Println("Logging in to Shoehorn...")

	// Normalize server URL
	serverURL = NormalizeServerURL(serverURL)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create API client
	client := api.NewClient(serverURL)

	ctx := context.Background()

	// Step 1: Initiate device flow
	fmt.Println("Initiating device authorization flow...")
	deviceResp, err := client.InitDeviceFlow(ctx)
	if err != nil {
		return fmt.Errorf("failed to initiate device flow: %w", err)
	}

	// Step 2: Display instructions to user
	fmt.Println()
	fmt.Println("To authenticate, visit:")
	if deviceResp.VerificationURIComplete != "" {
		fmt.Printf("  %s\n", deviceResp.VerificationURIComplete)
		fmt.Println()
		fmt.Println("Or manually visit:")
	}
	fmt.Printf("  %s\n", deviceResp.VerificationURI)
	fmt.Printf("  And enter code: %s\n", deviceResp.UserCode)
	fmt.Println()
	fmt.Printf("Code expires in %d seconds\n", deviceResp.ExpiresIn)
	fmt.Println()

	// Step 3: Poll for token
	fmt.Println("Waiting for authentication...")

	var pollResp *api.DevicePollResponse
	interval := time.Duration(deviceResp.Interval) * time.Second
	timeout := time.Duration(deviceResp.ExpiresIn) * time.Second
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("authentication timeout: device code expired")
		}

		pollResp, err = client.PollDeviceFlow(ctx, deviceResp.DeviceCode)
		if err != nil {
			return fmt.Errorf("failed to poll device flow: %w", err)
		}

		// Check if authentication is complete
		if pollResp.AccessToken != "" {
			break
		}

		// Check pending status
		if pollResp.Pending {
			if pollResp.Message == "slow_down" {
				// Double the interval if we're polling too fast
				interval *= 2
			}
			time.Sleep(interval)
			continue
		}

		// Unexpected response
		return fmt.Errorf("unexpected response from server: %+v", pollResp)
	}

	// Step 4: Store credentials
	currentProfile := cfg.Profiles[cfg.CurrentProfile]
	if currentProfile == nil {
		return fmt.Errorf("current profile not found: %s", cfg.CurrentProfile)
	}

	currentProfile.Auth = &config.Auth{
		ProviderType: "api", // API-proxied auth
		Issuer:       serverURL,
		ClientID:     "shoehorn-cli",
		AccessToken:  pollResp.AccessToken,
		RefreshToken: pollResp.RefreshToken,
		TokenType:    pollResp.TokenType,
		ExpiresAt:    pollResp.ExpiresAt,
		User: &config.User{
			Email:    pollResp.User.Email,
			Name:     pollResp.User.Name,
			TenantID: pollResp.User.TenantID,
		},
	}

	currentProfile.Server = serverURL

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Success!
	fmt.Println()
	fmt.Println("Authentication successful!")
	fmt.Printf("Logged in as: %s\n", pollResp.User.Email)
	if pollResp.User.Name != "" {
		fmt.Printf("Name: %s\n", pollResp.User.Name)
	}
	if pollResp.User.TenantID != "" {
		fmt.Printf("Tenant: %s\n", pollResp.User.TenantID)
	}
	fmt.Printf("Profile: %s\n", cfg.CurrentProfile)
	fmt.Printf("Server: %s\n", serverURL)

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current profile
	currentProfile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	fmt.Printf("Profile: %s\n", cfg.CurrentProfile)
	fmt.Printf("Server:  %s\n", currentProfile.Server)

	if !cfg.IsAuthenticated() {
		fmt.Println("Status:  Not authenticated")
		fmt.Println()
		fmt.Println("Run 'shoehorn auth login' to authenticate")
		return nil
	}

	fmt.Println("Status:  Authenticated")

	if currentProfile.Auth.User != nil {
		fmt.Printf("Email:   %s\n", currentProfile.Auth.User.Email)
		if currentProfile.Auth.User.Name != "" {
			fmt.Printf("Name:    %s\n", currentProfile.Auth.User.Name)
		}
		if currentProfile.Auth.User.TenantID != "" {
			fmt.Printf("Tenant:  %s\n", currentProfile.Auth.User.TenantID)
		}
	}

	// Check token expiration
	if cfg.IsTokenExpired() {
		fmt.Println("Token:   Expired (use 'shoehorn auth login' to refresh)")
	} else {
		timeUntilExpiry := time.Until(currentProfile.Auth.ExpiresAt)
		fmt.Printf("Token:   Valid (expires in %s)\n", formatDuration(timeUntilExpiry))
	}

	// Optionally verify with server
	if currentProfile.Auth.AccessToken != "" && !cfg.IsTokenExpired() {
		client := api.NewClient(currentProfile.Server)
		client.SetToken(currentProfile.Auth.AccessToken)

		ctx := context.Background()
		serverStatus, err := client.GetAuthStatus(ctx)
		if err != nil {
			fmt.Printf("Server:  Unable to verify (offline or token invalid)\n")
		} else {
			if serverStatus.Authenticated {
				fmt.Println("Server:  Token verified with server")
			} else {
				fmt.Println("Server:  Token rejected by server")
			}
		}
	}

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Clear auth from current profile
	currentProfile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	currentProfile.Auth = nil

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Logged out from profile: %s\n", cfg.CurrentProfile)
	fmt.Println("Note: Tokens are not revoked on the server. They will expire naturally.")

	return nil
}

// Helper functions
// NormalizeServerURL normalizes a server URL by ensuring it has a scheme and removing trailing slashes
func NormalizeServerURL(url string) string {
	// Add scheme if missing
	if url != "" && !hasScheme(url) {
		url = "https://" + url
	}

	// Remove trailing slashes
	for len(url) > 0 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}

	return url
}

// hasScheme checks if a URL has a scheme (http:// or https://)
func hasScheme(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}


func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
