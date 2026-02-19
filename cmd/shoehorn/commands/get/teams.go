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

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "List all teams",
	RunE:  runGetTeams,
}

var teamCmd = &cobra.Command{
	Use:   "team <slug>",
	Short: "Get details for a specific team",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetTeam,
}

func init() {
	GetCmd.AddCommand(teamsCmd)
	GetCmd.AddCommand(teamCmd)
}

func runGetTeams(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Loading teams...", func() (any, error) {
		return client.ListTeams(context.Background())
	})
	if spinErr != nil {
		return fmt.Errorf("list teams: %w", spinErr)
	}

	teams := result.([]*api.Team)

	mode := ui.DetectMode(commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(teams)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(teams)
	}

	cols := []table.Column{
		{Title: "Name", Width: 28},
		{Title: "Slug", Width: 24},
		{Title: "Members", Width: 10},
		{Title: "Description", Width: 40},
	}

	rows := make([]table.Row, len(teams))
	for i, t := range teams {
		desc := t.Description
		if len(desc) > 38 {
			desc = desc[:38] + "…"
		}
		rows[i] = table.Row{t.Name, t.Slug, fmt.Sprintf("%d", t.MemberCount), desc}
	}

	_, err = tui.RunTable(tui.TableConfig{
		Title:   fmt.Sprintf("Teams  (%d)", len(teams)),
		Columns: cols,
		Rows:    rows,
	})
	return err
}

func runGetTeam(cmd *cobra.Command, args []string) error {
	slug := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Loading team...", func() (any, error) {
		return client.GetTeam(context.Background(), slug)
	})
	if spinErr != nil {
		return fmt.Errorf("get team: %w", spinErr)
	}

	team := result.(*api.TeamDetail)

	mode := ui.DetectMode(commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(team)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(team)
	}

	sections := []tui.DetailSection{
		{
			Fields: []tui.Field{
				{Label: "Name", Value: team.Name},
				{Label: "Slug", Value: team.Slug},
				{Label: "Description", Value: team.Description},
				{Label: "Members", Value: fmt.Sprintf("%d", len(team.Members))},
			},
		},
	}

	if len(team.Members) > 0 {
		memberFields := make([]tui.Field, len(team.Members))
		for i, m := range team.Members {
			name := m.Name
			if name == "" {
				name = m.Email
			}
			memberFields[i] = tui.Field{
				Label: name,
				Value: fmt.Sprintf("%s  %s", m.Email, tui.MutedStyle.Render(m.Role)),
			}
		}
		sections = append(sections, tui.DetailSection{
			Title:  "Members",
			Fields: memberFields,
		})
	}

	fmt.Println(tui.RenderDetail(team.Name, sections))
	return nil
}
