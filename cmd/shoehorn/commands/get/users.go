package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/imbabamba/shoehorn-cli/cmd/shoehorn/commands"
	"github.com/imbabamba/shoehorn-cli/pkg/api"
	"github.com/imbabamba/shoehorn-cli/pkg/tui"
	"github.com/imbabamba/shoehorn-cli/pkg/ui"
	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all users in the directory",
	RunE:  runGetUsers,
}

var userCmd = &cobra.Command{
	Use:   "user <id>",
	Short: "Get details for a specific user",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetUser,
}

func init() {
	GetCmd.AddCommand(usersCmd)
	GetCmd.AddCommand(userCmd)
}

func runGetUsers(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Loading users...", func() (any, error) {
		return client.ListUsers(context.Background())
	})
	if spinErr != nil {
		return fmt.Errorf("list users: %w", spinErr)
	}

	users := result.([]*api.User)

	mode := ui.DetectMode(commands.Interactive(), commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(users)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(users)
	}

	colNames := []string{"Name", "Email", "ID"}
	rows := make([][]string, len(users))
	for i, u := range users {
		rows[i] = []string{u.Name, u.Email, u.ID}
	}

	if mode == ui.ModeInteractive {
		tuiCols := []table.Column{
			{Title: "Name", Width: 28},
			{Title: "Email", Width: 36},
			{Title: "ID", Width: 36},
		}
		tuiRows := make([]table.Row, len(rows))
		for i, r := range rows {
			tuiRows[i] = table.Row(r)
		}
		_, err = tui.RunTable(tui.TableConfig{
			Title:   fmt.Sprintf("Users  (%d)", len(users)),
			Columns: tuiCols,
			Rows:    tuiRows,
		})
		return err
	}

	ui.RenderTable(colNames, rows)
	return nil
}

func runGetUser(cmd *cobra.Command, args []string) error {
	id := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Loading user...", func() (any, error) {
		return client.GetUser(context.Background(), id)
	})
	if spinErr != nil {
		return fmt.Errorf("get user: %w", spinErr)
	}

	user := result.(*api.UserDetail)

	mode := ui.DetectMode(commands.Interactive(), commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(user)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(user)
	}

	groups := strings.Join(user.Groups, ", ")
	teams := strings.Join(user.Teams, ", ")
	roles := strings.Join(user.Roles, ", ")

	fmt.Println(tui.RenderDetail(user.Name, []tui.DetailSection{
		{
			Fields: []tui.Field{
				{Label: "ID", Value: user.ID},
				{Label: "Email", Value: user.Email},
			},
		},
		{
			Title: "Access",
			Fields: []tui.Field{
				{Label: "Groups", Value: groups},
				{Label: "Teams", Value: teams},
				{Label: "Roles", Value: roles},
			},
		},
	}))
	return nil
}
