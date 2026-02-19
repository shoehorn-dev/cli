package commands

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/imbabamba/shoehorn-cli/pkg/ui"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search entities",
	Long:  `Search across all catalog entities by name, description, or tags.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner(fmt.Sprintf("Searching for %q...", query), func() (any, error) {
		return client.Search(context.Background(), query)
	})
	if spinErr != nil {
		return fmt.Errorf("search: %w", spinErr)
	}

	sr := result.(*api.SearchResult)

	mode := ui.DetectMode(noInteractive, outputFormat)
	if mode == ui.ModeJSON {
		return ui.RenderJSON(sr)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(sr)
	}

	if len(sr.Hits) == 0 {
		fmt.Printf("No results for %q\n", query)
		return nil
	}

	cols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Type", Width: 14},
		{Title: "Owner", Width: 20},
		{Title: "Description", Width: 50},
	}

	rows := make([]table.Row, len(sr.Hits))
	for i, h := range sr.Hits {
		desc := h.Description
		if len(desc) > 48 {
			desc = desc[:48] + "…"
		}
		rows[i] = table.Row{h.Name, h.Type, h.Owner, desc}
	}

	selected, err := tui.RunTable(tui.TableConfig{
		Title:   fmt.Sprintf("Search Results — %q  (%d hits)", query, sr.TotalCount),
		Columns: cols,
		Rows:    rows,
	})
	if err != nil {
		return err
	}

	if selected != nil {
		// Find the matching hit and show details
		for _, h := range sr.Hits {
			if h.Name == selected[0] {
				fmt.Println(tui.RenderDetail(h.Name, []tui.DetailSection{
					{
						Fields: []tui.Field{
							{Label: "ID", Value: h.ID},
							{Label: "Type", Value: h.Type},
							{Label: "Owner", Value: h.Owner},
							{Label: "Description", Value: h.Description},
						},
					},
				}))
				break
			}
		}
	}

	return nil
}
