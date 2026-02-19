package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
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
	RunE:  runListRuns,
}

// runGetCmd represents the forge run get command
var runGetCmd = &cobra.Command{
	Use:   "get <run-id>",
	Short: "Get workflow run details",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetRun,
}

// runCreateCmd creates a new forge run
var runCreateCmd = &cobra.Command{
	Use:   "create <mold-slug>",
	Short: "Create a new workflow run",
	Long: `Start a new Forge workflow run from a mold slug.

Optionally pass input values as JSON:
  shoehorn forge run create my-mold --inputs '{"env":"staging"}'`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateRun,
}

// moldsCmd is the forge molds subcommand group
var moldsCmd = &cobra.Command{
	Use:   "molds",
	Short: "Manage workflow molds (templates)",
}

// moldsListCmd lists all molds
var moldsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all molds",
	RunE:  runMoldsList,
}

// moldsGetCmd gets a single mold
var moldsGetCmd = &cobra.Command{
	Use:   "get <slug>",
	Short: "Get details for a specific mold",
	Args:  cobra.ExactArgs(1),
	RunE:  runMoldsGet,
}

var runInputsJSON string

func init() {
	runCreateCmd.Flags().StringVar(&runInputsJSON, "inputs", "", "Input values as JSON object")

	runCmd.AddCommand(runListCmd)
	runCmd.AddCommand(runGetCmd)
	runCmd.AddCommand(runCreateCmd)

	moldsCmd.AddCommand(moldsListCmd)
	moldsCmd.AddCommand(moldsGetCmd)

	forgeCmd.AddCommand(runCmd)
	forgeCmd.AddCommand(moldsCmd)
	rootCmd.AddCommand(forgeCmd)
}

// ─── Runs ────────────────────────────────────────────────────────────────────

// legacyForgeRun mirrors the existing fields from the old API shape
type legacyForgeRun struct {
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

type legacyRunsResponse struct {
	Runs       []legacyForgeRun `json:"runs"`
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

func runListRuns(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	var response legacyRunsResponse
	if err := client.Get(ctx, "/api/v1/forge/runs", &response); err != nil {
		return fmt.Errorf("failed to list runs: %w", err)
	}

	mode := ui.DetectMode(noInteractive, outputFormat)
	switch mode {
	case ui.ModeJSON:
		return ui.RenderJSON(response)
	case ui.ModeYAML:
		return ui.RenderYAML(response)
	default:
		return outputRunsTable(response.Runs)
	}
}

func runGetRun(cmd *cobra.Command, args []string) error {
	runID := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	var run legacyForgeRun
	if err := client.Get(ctx, "/api/v1/forge/runs/"+runID, &run); err != nil {
		return fmt.Errorf("failed to get run: %w", err)
	}

	mode := ui.DetectMode(noInteractive, outputFormat)
	switch mode {
	case ui.ModeJSON:
		return ui.RenderJSON(run)
	case ui.ModeYAML:
		return ui.RenderYAML(run)
	default:
		return outputRunDetails(run)
	}
}

func runCreateRun(cmd *cobra.Command, args []string) error {
	moldSlug := args[0]

	var inputs map[string]any
	if runInputsJSON != "" {
		if err := json.Unmarshal([]byte(runInputsJSON), &inputs); err != nil {
			return fmt.Errorf("parse --inputs JSON: %w", err)
		}
	}

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner(fmt.Sprintf("Starting run for mold %q...", moldSlug), func() (any, error) {
		return client.CreateRun(context.Background(), moldSlug, inputs)
	})
	if spinErr != nil {
		fmt.Println(tui.ErrorBox("Run Failed", spinErr.Error()))
		return nil
	}

	run := result.(*api.ForgeRun)

	body := fmt.Sprintf(
		"%s  %s\n%s  %s\n%s  %s",
		tui.LabelStyle.Render("Run ID"),    run.ID,
		tui.LabelStyle.Render("Mold"),      moldSlug,
		tui.LabelStyle.Render("Status"),    tui.StatusColor(run.Status).Render(run.Status),
	)
	fmt.Println(tui.SuccessBox("Run Created", body))
	return nil
}

// ─── Molds ───────────────────────────────────────────────────────────────────

func runMoldsList(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Loading molds...", func() (any, error) {
		return client.ListMolds(context.Background())
	})
	if spinErr != nil {
		return fmt.Errorf("list molds: %w", spinErr)
	}

	molds := result.([]*api.Mold)

	mode := ui.DetectMode(noInteractive, outputFormat)
	if mode == ui.ModeJSON {
		return ui.RenderJSON(molds)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(molds)
	}

	cols := []table.Column{
		{Title: "Name", Width: 28},
		{Title: "Slug", Width: 24},
		{Title: "Version", Width: 10},
		{Title: "Description", Width: 40},
	}

	rows := make([]table.Row, len(molds))
	for i, m := range molds {
		desc := m.Description
		if len(desc) > 38 {
			desc = desc[:38] + "…"
		}
		rows[i] = table.Row{m.Name, m.Slug, m.Version, desc}
	}

	_, err = tui.RunTable(tui.TableConfig{
		Title:   fmt.Sprintf("Molds  (%d)", len(molds)),
		Columns: cols,
		Rows:    rows,
	})
	return err
}

func runMoldsGet(cmd *cobra.Command, args []string) error {
	slug := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner(fmt.Sprintf("Loading mold %q...", slug), func() (any, error) {
		return client.GetMold(context.Background(), slug)
	})
	if spinErr != nil {
		return fmt.Errorf("get mold: %w", spinErr)
	}

	mold := result.(*api.MoldDetail)

	mode := ui.DetectMode(noInteractive, outputFormat)
	if mode == ui.ModeJSON {
		return ui.RenderJSON(mold)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(mold)
	}

	// Build inputs section
	inputFields := make([]tui.Field, len(mold.Inputs))
	for i, inp := range mold.Inputs {
		req := ""
		if inp.Required {
			req = " (required)"
		}
		def := ""
		if inp.Default != "" {
			def = fmt.Sprintf("  default: %s", tui.MutedStyle.Render(inp.Default))
		}
		inputFields[i] = tui.Field{
			Label: inp.Name,
			Value: fmt.Sprintf("%s%s%s  %s", inp.Type, req, def, tui.MutedStyle.Render(inp.Description)),
		}
	}

	// Build steps section
	stepFields := make([]tui.Field, len(mold.Steps))
	for i, s := range mold.Steps {
		stepFields[i] = tui.Field{
			Label: fmt.Sprintf("Step %d", i+1),
			Value: fmt.Sprintf("%s  %s", s.Name, tui.MutedStyle.Render(s.Action)),
		}
	}

	sections := []tui.DetailSection{
		{
			Fields: []tui.Field{
				{Label: "Name", Value: mold.Name},
				{Label: "Slug", Value: mold.Slug},
				{Label: "Version", Value: mold.Version},
				{Label: "Description", Value: mold.Description},
			},
		},
	}

	if len(inputFields) > 0 {
		sections = append(sections, tui.DetailSection{
			Title:  fmt.Sprintf("Inputs (%d)", len(mold.Inputs)),
			Fields: inputFields,
		})
	}

	if len(stepFields) > 0 {
		sections = append(sections, tui.DetailSection{
			Title:  fmt.Sprintf("Steps (%d)", len(mold.Steps)),
			Fields: stepFields,
		})
	}

	fmt.Println(tui.RenderDetail(mold.Name, sections))
	return nil
}

// ─── Legacy table helpers ────────────────────────────────────────────────────

func outputRunsTable(runs []legacyForgeRun) error {
	if len(runs) == 0 {
		fmt.Println("No runs found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tMOLD ID\tSTATUS\tCREATED BY\tCREATED AT")
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

func outputRunDetails(run legacyForgeRun) error {
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
		"executing":   "▶",
		"completed":   "✓",
		"failed":      "✗",
		"cancelled":   "⊘",
		"rolled_back": "↩",
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
