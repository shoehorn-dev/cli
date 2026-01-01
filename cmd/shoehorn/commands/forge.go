package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/config"
	"github.com/imbabamba/shoehorn-cli/pkg/ui"
	"github.com/spf13/cobra"
)

// forgeCmd represents the forge command group
var forgeCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge workflow commands",
	Long:  `Manage and execute Forge workflows.`,
}

// runCmd represents the forge run command group
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage workflow runs",
	Long:  `List, create, and inspect workflow runs.`,
}

// runListCmd represents the forge run list command
var runListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workflow runs",
	Long:  `List all workflow runs for the current tenant.`,
	RunE:  runListRuns,
}

// runGetCmd represents the forge run get command
var runGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get workflow run details",
	Long:  `Get detailed information about a specific workflow run.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGetRun,
}

func init() {
	// Build command hierarchy
	runCmd.AddCommand(runListCmd)
	runCmd.AddCommand(runGetCmd)
	forgeCmd.AddCommand(runCmd)
	rootCmd.AddCommand(forgeCmd)
}

// ForgeRun represents a workflow run
type ForgeRun struct {
	ID             string     `json:"id"`
	MoldID         string     `json:"mold_id"`
	TenantID       string     `json:"tenant_id"`
	Status         string     `json:"status"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Error          *string    `json:"error,omitempty"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
}

// ForgeRunsResponse represents the API response for listing runs
type ForgeRunsResponse struct {
	Runs       []ForgeRun `json:"runs"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

func runListRuns(cmd *cobra.Command, args []string) error {
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

	// Create API client with server URL
	client := api.NewClient(currentProfile.Server)

	// Set auth token if available
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'shoehorn auth login' first")
	}
	client.SetToken(currentProfile.Auth.AccessToken)

	// Make API request
	ctx := context.Background()
	var response ForgeRunsResponse

	if err := client.Get(ctx, "/api/v1/forge/runs", &response); err != nil {
		return fmt.Errorf("failed to list runs: %w", err)
	}

	// Detect output mode
	mode := ui.DetectMode(noInteractive, outputFormat)

	// Render based on mode
	switch mode {
	case ui.ModeJSON:
		return ui.RenderJSON(response)
	case ui.ModeYAML:
		return ui.RenderYAML(response)
	case ui.ModeInteractive:
		// TODO: Implement interactive TUI in future
		// For now, fall through to plain table
		fallthrough
	case ui.ModePlain:
		return outputRunsTable(response.Runs)
	default:
		return outputRunsTable(response.Runs)
	}
}

func runGetRun(cmd *cobra.Command, args []string) error {
	runID := args[0]

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

	// Create API client with server URL
	client := api.NewClient(currentProfile.Server)

	// Set auth token if available
	if !cfg.IsAuthenticated() {
		return fmt.Errorf("not authenticated - run 'shoehorn auth login' first")
	}
	client.SetToken(currentProfile.Auth.AccessToken)

	// Make API request
	ctx := context.Background()
	var run ForgeRun

	if err := client.Get(ctx, "/api/v1/forge/runs/"+runID, &run); err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	// Detect output mode
	mode := ui.DetectMode(noInteractive, outputFormat)

	// Render based on mode
	switch mode {
	case ui.ModeJSON:
		return ui.RenderJSON(run)
	case ui.ModeYAML:
		return ui.RenderYAML(run)
	case ui.ModeInteractive:
		// TODO: Implement interactive TUI in future
		// For now, fall through to plain table
		fallthrough
	case ui.ModePlain:
		return outputRunDetails(run)
	default:
		return outputRunDetails(run)
	}
}

func outputRunsTable(runs []ForgeRun) error {
	if len(runs) == 0 {
		fmt.Println("No runs found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "ID\tMOLD ID\tSTATUS\tCREATED BY\tCREATED AT")

	// Rows
	for _, run := range runs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			run.ID,
			run.MoldID,
			formatStatus(run.Status),
			run.CreatedBy,
			formatTime(run.CreatedAt),
		)
	}

	return nil
}

func outputRunDetails(run ForgeRun) error {
	fmt.Printf("ID:              %s\n", run.ID)
	fmt.Printf("Mold ID:         %s\n", run.MoldID)
	fmt.Printf("Status:          %s\n", formatStatus(run.Status))
	fmt.Printf("Created By:      %s\n", run.CreatedBy)
	fmt.Printf("Created At:      %s\n", run.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated At:      %s\n", run.UpdatedAt.Format(time.RFC3339))

	if run.CompletedAt != nil {
		fmt.Printf("Completed At:    %s\n", run.CompletedAt.Format(time.RFC3339))
		duration := run.CompletedAt.Sub(run.CreatedAt)
		fmt.Printf("Duration:        %s\n", formatDuration(duration))
	}

	if run.Error != nil {
		fmt.Printf("Error:           %s\n", *run.Error)
	}

	if run.IdempotencyKey != nil {
		fmt.Printf("Idempotency Key: %s\n", *run.IdempotencyKey)
	}

	return nil
}

func formatStatus(status string) string {
	statusIcons := map[string]string{
		"pending":     "⏳",
		"executing":   "▶️",
		"completed":   "✓",
		"failed":      "✗",
		"cancelled":   "⊘",
		"rolled_back": "↩️",
	}

	icon, ok := statusIcons[status]
	if !ok {
		icon = "•"
	}

	return fmt.Sprintf("%s %s", icon, status)
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}

	return t.Format("2006-01-02")
}
