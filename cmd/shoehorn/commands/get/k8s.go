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

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "List Kubernetes agents",
	Long:  `Display all registered Kubernetes agents and their connection status.`,
	RunE:  runGetK8s,
}

func init() {
	GetCmd.AddCommand(k8sCmd)
}

func runGetK8s(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	result, spinErr := tui.RunSpinner("Loading K8s agents...", func() (any, error) {
		return client.ListK8sAgents(context.Background())
	})
	if spinErr != nil {
		return fmt.Errorf("list k8s agents: %w", spinErr)
	}

	agents := result.([]*api.K8sAgent)

	mode := ui.DetectMode(commands.Interactive(), commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(agents)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(agents)
	}

	colNames := []string{"Cluster", "Status", "Version", "Namespace", "Last Seen"}
	rows := make([][]string, len(agents))
	for i, a := range agents {
		rows[i] = []string{a.ClusterName, a.Status, a.Version, a.Namespace, a.LastSeen}
	}

	if mode == ui.ModeInteractive {
		tuiCols := []table.Column{
			{Title: "Cluster", Width: 30},
			{Title: "Status", Width: 14},
			{Title: "Version", Width: 14},
			{Title: "Namespace", Width: 20},
			{Title: "Last Seen", Width: 20},
		}
		tuiRows := make([]table.Row, len(agents))
		for j, a := range agents {
			status := tui.StatusColor(a.Status).Render(a.Status)
			tuiRows[j] = table.Row{a.ClusterName, status, a.Version, a.Namespace, a.LastSeen}
		}
		_, err = tui.RunTable(tui.TableConfig{
			Title:   fmt.Sprintf("Kubernetes Agents  (%d)", len(agents)),
			Columns: tuiCols,
			Rows:    tuiRows,
		})
		return err
	}

	ui.RenderTable(colNames, rows)
	return nil
}
