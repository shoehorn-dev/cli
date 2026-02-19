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

var ownerBy string

var ownedCmd = &cobra.Command{
	Use:   "owned <name>",
	Short: "List entities owned by a team or user",
	Long: `List all catalog entities owned by a specific team or user.

Examples:
  shoehorn get owned --by team platform-team
  shoehorn get owned --by user user-id-123`,
	Args: cobra.ExactArgs(1),
	RunE: runGetOwned,
}

func init() {
	ownedCmd.Flags().StringVar(&ownerBy, "by", "team", "Owner type: team or user")
	GetCmd.AddCommand(ownedCmd)
}

func runGetOwned(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner(fmt.Sprintf("Loading entities owned by %s %q...", ownerBy, name), func() (any, error) {
		return client.ListEntities(context.Background(), api.ListEntitiesOpts{Owner: name})
	})
	if spinErr != nil {
		return fmt.Errorf("list entities: %w", spinErr)
	}

	entities := result.([]*api.Entity)

	mode := ui.DetectMode(commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(entities)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(entities)
	}

	cols := []table.Column{
		{Title: "Name", Width: 28},
		{Title: "Type", Width: 14},
		{Title: "Description", Width: 50},
	}

	rows := make([]table.Row, len(entities))
	for i, e := range entities {
		desc := e.Description
		if len(desc) > 48 {
			desc = desc[:48] + "…"
		}
		rows[i] = table.Row{e.Name, e.Type, desc}
	}

	_, err = tui.RunTable(tui.TableConfig{
		Title:   fmt.Sprintf("Owned by %s %q  (%d entities)", ownerBy, name, len(entities)),
		Columns: cols,
		Rows:    rows,
	})
	return err
}
