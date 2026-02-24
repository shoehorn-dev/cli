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

var (
	entityType    string
	entityOwner   string
	showScorecard bool
)

var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "List all catalog entities",
	RunE:  runGetEntities,
}

var entityCmd = &cobra.Command{
	Use:   "entity <id-or-slug>",
	Short: "Get details for a specific entity",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetEntity,
}

func init() {
	entitiesCmd.Flags().StringVar(&entityType, "type", "", "Filter by entity type (service, library, etc.)")
	entitiesCmd.Flags().StringVar(&entityOwner, "owner", "", "Filter by owning team slug")

	entityCmd.Flags().BoolVar(&showScorecard, "scorecard", false, "Include scorecard in output")

	GetCmd.AddCommand(entitiesCmd)
	GetCmd.AddCommand(entityCmd)
}

func runGetEntities(cmd *cobra.Command, args []string) error {
	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	opts := api.ListEntitiesOpts{
		Type:  entityType,
		Owner: entityOwner,
	}

	result, spinErr := tui.RunSpinner("Loading entities...", func() (any, error) {
		return client.ListEntities(context.Background(), opts)
	})
	if spinErr != nil {
		return fmt.Errorf("list entities: %w", spinErr)
	}

	entities := result.([]*api.Entity)

	mode := ui.DetectMode(commands.Interactive(), commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(entities)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(entities)
	}

	colNames := []string{"Name", "Type", "Owner", "Description"}
	rows := make([][]string, len(entities))
	for i, e := range entities {
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
		title := fmt.Sprintf("Entities  (%d)", len(entities))
		if entityType != "" {
			title += fmt.Sprintf("  type=%s", entityType)
		}
		if entityOwner != "" {
			title += fmt.Sprintf("  owner=%s", entityOwner)
		}
		_, err = tui.RunTable(tui.TableConfig{
			Title:   title,
			Columns: tuiCols,
			Rows:    tuiRows,
		})
		return err
	}

	ui.RenderTable(colNames, rows)
	return nil
}

func runGetEntity(cmd *cobra.Command, args []string) error {
	id := args[0]

	client, err := api.NewClientFromConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()

	result, spinErr := tui.RunSpinner("Loading entity...", func() (any, error) {
		e, err := client.GetEntity(ctx, id)
		if err != nil {
			return nil, err
		}
		return e, nil
	})
	if spinErr != nil {
		return fmt.Errorf("get entity: %w", spinErr)
	}

	entity := result.(*api.EntityDetail)

	mode := ui.DetectMode(commands.Interactive(), commands.NoInteractive(), commands.OutputFormat())
	if mode == ui.ModeJSON {
		return ui.RenderJSON(entity)
	}
	if mode == ui.ModeYAML {
		return ui.RenderYAML(entity)
	}

	// Fetch resources and status concurrently
	type fetchResult struct {
		resources []*api.Resource
		status    *api.EntityStatus
		scorecard *api.Scorecard
	}

	fr := fetchResult{}
	resCh := make(chan []*api.Resource, 1)
	statCh := make(chan *api.EntityStatus, 1)
	scCh := make(chan *api.Scorecard, 1)

	go func() {
		r, _ := client.GetEntityResources(ctx, entity.ID)
		resCh <- r
	}()
	go func() {
		s, _ := client.GetEntityStatus(ctx, entity.ID)
		statCh <- s
	}()
	if showScorecard {
		go func() {
			sc, _ := client.GetEntityScorecard(ctx, entity.ID)
			scCh <- sc
		}()
	} else {
		scCh <- nil
	}

	fr.resources = <-resCh
	fr.status = <-statCh
	fr.scorecard = <-scCh

	// Build detail panel
	mainFields := []tui.Field{
		{Label: "Type", Value: entity.Type},
		{Label: "Owner", Value: entity.Owner},
		{Label: "Lifecycle", Value: entity.Lifecycle},
		{Label: "Tier", Value: entity.Tier},
		{Label: "Description", Value: entity.Description},
		{Label: "Tags", Value: tui.RenderTagBadges(entity.Tags)},
	}

	if fr.status != nil {
		health := tui.StatusColor(fr.status.Health).Render("● " + fr.status.Health)
		mainFields = append(mainFields,
			tui.Field{Label: "Status", Value: fmt.Sprintf("%s  (%.2f%% uptime)", health, fr.status.Uptime)},
		)
	}

	// Links
	if len(entity.Links) > 0 {
		linkNames := make([]string, len(entity.Links))
		for i, l := range entity.Links {
			linkNames[i] = l.Title
		}
		mainFields = append(mainFields, tui.Field{Label: "Links", Value: tui.RenderLinkLine(linkNames)})
	}

	sections := []tui.DetailSection{
		{Fields: mainFields},
	}

	// Resources section
	if len(fr.resources) > 0 {
		resFields := make([]tui.Field, len(fr.resources))
		for i, r := range fr.resources {
			resFields[i] = tui.Field{
				Label: r.Name,
				Value: fmt.Sprintf("%s  %s", r.Type, tui.MutedStyle.Render(r.Environment)),
			}
		}
		sections = append(sections, tui.DetailSection{
			Title:  fmt.Sprintf("Resources (%d)", len(fr.resources)),
			Fields: resFields,
		})
	}

	// Scorecard section
	if fr.scorecard != nil {
		gradeStr := tui.GradeColor(fr.scorecard.Grade).Render(fr.scorecard.Grade)
		bar := tui.RenderScoreBar(fr.scorecard.Score, fr.scorecard.MaxScore)
		sections = append(sections, tui.DetailSection{
			Title: "Scorecard",
			Fields: []tui.Field{
				{Label: "Grade", Value: fmt.Sprintf("%s  %s", gradeStr, bar)},
			},
		})
		// Failed checks
		var failed []string
		for _, ch := range fr.scorecard.Checks {
			if !ch.Passed {
				failed = append(failed, "✗ "+ch.Name)
			}
		}
		if len(failed) > 0 {
			sections[len(sections)-1].Fields = append(sections[len(sections)-1].Fields,
				tui.Field{Label: "Failed", Value: strings.Join(failed, "\n"+strings.Repeat(" ", 20))},
			)
		}
	}

	fmt.Println(tui.RenderDetail(entity.Name, sections))
	return nil
}
