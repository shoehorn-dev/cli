package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/config"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/spf13/cobra"
)

var (
	serverURL string
	patToken  string
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
	Long: `Authenticate with the Shoehorn platform using a Personal Access Token.

Create a PAT in the Shoehorn UI under Settings > API Keys, then run:
  shoehorn auth login --server http://localhost:8080 --token shp_your_token`,
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
	loginCmd.Flags().StringVar(&serverURL, "server", "http://localhost:8080", "Shoehorn API server URL")
	loginCmd.Flags().StringVar(&patToken, "token", "", "Personal Access Token (shp_xxx)")

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)

	rootCmd.AddCommand(authCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	if patToken == "" {
		return fmt.Errorf("a Personal Access Token is required\n\nUsage:\n  shoehorn auth login --server <url> --token <PAT>")
	}
	serverURL = NormalizeServerURL(serverURL)
	return runLoginWithPAT(serverURL, patToken)
}

// runLoginWithPAT authenticates using a Personal Access Token
func runLoginWithPAT(server, token string) error {
	client := api.NewClient(server)
	client.SetToken(token)

	ctx := context.Background()

	// Verify token by calling /me
	result, err := tui.RunSpinner("Verifying token...", func() (any, error) {
		return client.GetMe(ctx)
	})
	if err != nil {
		fmt.Println(tui.ErrorBox("Authentication Failed", err.Error()))
		return nil
	}

	me := result.(*api.MeResponse)

	// Save config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	currentProfile := cfg.Profiles[cfg.CurrentProfile]
	if currentProfile == nil {
		currentProfile = &config.Profile{Name: cfg.CurrentProfile}
		cfg.Profiles[cfg.CurrentProfile] = currentProfile
	}

	currentProfile.Server = server
	currentProfile.Auth = &config.Auth{
		ProviderType: "pat",
		Issuer:       server,
		AccessToken:  token,
		User: &config.User{
			Email:    me.Email,
			Name:     me.Name,
			TenantID: me.TenantID,
		},
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Success panel
	body := fmt.Sprintf(
		"%s  %s\n%s  %s\n%s  %s\n%s  %s",
		tui.LabelStyle.Render("Name"), me.Name,
		tui.LabelStyle.Render("Email"), me.Email,
		tui.LabelStyle.Render("Tenant"), me.TenantID,
		tui.LabelStyle.Render("Server"), server,
	)
	fmt.Println(tui.SuccessBox("Authenticated with PAT", body))
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	currentProfile, err := cfg.GetCurrentProfile()
	if err != nil {
		return err
	}

	fmt.Printf("Profile: %s\n", cfg.CurrentProfile)
	fmt.Printf("Server:  %s\n", currentProfile.Server)

	if !cfg.IsAuthenticated() {
		fmt.Println("Status:  Not authenticated")
		fmt.Println()
		fmt.Println("Run 'shoehorn auth login --token <PAT>' to authenticate")
		return nil
	}

	if cfg.IsPATAuth() {
		fmt.Println("Status:  Authenticated (PAT)")
	} else {
		fmt.Println("Status:  Authenticated")
	}

	if currentProfile.Auth.User != nil {
		fmt.Printf("Email:   %s\n", currentProfile.Auth.User.Email)
		if currentProfile.Auth.User.Name != "" {
			fmt.Printf("Name:    %s\n", currentProfile.Auth.User.Name)
		}
		if currentProfile.Auth.User.TenantID != "" {
			fmt.Printf("Tenant:  %s\n", currentProfile.Auth.User.TenantID)
		}
	}

	if cfg.IsTokenExpired() {
		fmt.Println("Token:   Expired (use 'shoehorn auth login' to refresh)")
	} else if cfg.IsPATAuth() {
		fmt.Println("Token:   Valid (PAT, no expiry)")
	} else {
		timeUntilExpiry := time.Until(currentProfile.Auth.ExpiresAt)
		fmt.Printf("Token:   Valid (expires in %s)\n", formatDuration(timeUntilExpiry))
	}

	// Verify with server
	if currentProfile.Auth.AccessToken != "" && !cfg.IsTokenExpired() {
		client := api.NewClient(currentProfile.Server)
		client.SetToken(currentProfile.Auth.AccessToken)
		ctx := context.Background()
		serverStatus, err := client.GetAuthStatus(ctx)
		if err != nil {
			fmt.Println("Server:  Unable to verify (offline or token invalid)")
		} else if serverStatus.Authenticated {
			fmt.Println("Server:  Token verified with server")
		} else {
			fmt.Println("Server:  Token rejected by server")
		}
	}

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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

// NormalizeServerURL normalizes a server URL
func NormalizeServerURL(url string) string {
	if url != "" && !hasScheme(url) {
		url = "https://" + url
	}
	for len(url) > 0 && url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	return url
}

func hasScheme(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || (len(url) > 8 && url[:8] == "https://"))
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
