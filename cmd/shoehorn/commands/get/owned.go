package get

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/imbabamba/shoehorn-cli/cmd/shoehorn/commands"
	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/imbabamba/shoehorn-cli/pkg/ui"
	"github.com/spf13/cobra"
)

var ownedCmd = &cobra.Command{
	Use:   "owned",
	Short: "List entities owned by your teams",
	Long: `List all catalog entities owned by your teams.

Uses your team memberships from 'shoehorn whoami' to find owned entities.
To see entities owned by a specific team: shoehorn get entities --owner <team>

Examples:
  shoehorn get owned                  # entities owned by your teams`,
	Args: cobra.NoArgs,
	RunE: runGetOwned,
}

func init() {
	GetCmd.AddCommand(ownedCmd)
}

func runGetOwned(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()

	me, err := client.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("get current user: %w", err)
	}

	if len(me.Teams) == 0 {
		fmt.Println("You are not a member of any teams.")
		return nil
	}

	seen := map[string]bool{}
	var allEntities []*api.Entity
	for _, team := range me.Teams {
		entities, err := client.ListEntities(ctx, api.ListEntitiesOpts{Owner: team})
		if err != nil {
			continue
		}
		for _, e := range entities {
			if !seen[e.ID] {
				seen[e.ID] = true
				allEntities = append(allEntities, e)
			}
		}
	}

	mode := ui.DetectMode(commands.Interactive(), commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(allEntities)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(allEntities)
	}

	colNames := []string{"Name", "Type", "Owner", "Description"}
	rows := make([][]string, len(allEntities))
	for i, e := range allEntities {
		desc := e.Description
		if len(desc) > 60 {
			desc = desc[:60] + "…"
		}
		rows[i] = []string{e.Name, e.Type, e.Owner, desc}
	}

	if mode == ui.ModeInteractive {
		tuiCols := []table.Column{
			{Title: "Name", Width: 28},
			{Title: "Type", Width: 14},
			{Title: "Owner", Width: 20},
			{Title: "Description", Width: 45},
		}
		tuiRows := make([]table.Row, len(rows))
		for i, r := range rows {
			tuiRows[i] = table.Row(r)
		}
		_, err = tui.RunTable(tui.TableConfig{
			Title:   fmt.Sprintf("My Teams' Entities  (%d)", len(allEntities)),
			Columns: tuiCols,
			Rows:    tuiRows,
		})
		return err
	}

	ui.RenderTable(colNames, rows)
	return nil
}
