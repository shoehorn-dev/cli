package commands

import (
	"context"
	"fmt"

	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/imbabamba/shoehorn-cli/pkg/ui"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current user info",
	Long:  `Display full information about the currently authenticated user.`,
	RunE:  runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Fetching user info...", func() (any, error) {
		return client.GetMe(context.Background())
	})
	if spinErr != nil {
		return fmt.Errorf("fetch user: %w", spinErr)
	}

	me, ok := result.(*api.MeResponse)
	if !ok {
		return fmt.Errorf("unexpected response type %T from GetMe", result)
	}

	mode := ui.DetectMode(interactive, noInteractive, outputFormat)
	if mode == ui.ModeJSON {
		return ui.RenderJSON(me)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(me)
	}

	panel := tui.RenderDetail(me.Name, []tui.DetailSection{
		{
			Fields: []tui.Field{
				{Label: "Email", Value: me.Email},
				{Label: "Tenant", Value: me.TenantID},
				{Label: "User ID", Value: me.ID},
			},
		},
	})

	fmt.Println(panel)
	return nil
}
