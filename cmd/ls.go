package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/display"
	"github.com/bhpcv252/portmap/internal/registry"
)

var (
	lsProject   string
	lsActive    bool
	lsFree      bool
	lsUnclaimed bool
	lsJSON      bool
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List claimed ports and their status",
	RunE:  runLs,
}

func init() {
	lsCmd.Flags().StringVarP(&lsProject, "project", "p", "", "Filter by project")
	lsCmd.Flags().BoolVar(&lsActive, "active", false, "Show only running ports")
	lsCmd.Flags().BoolVar(&lsFree, "free", false, "Show only stopped ports")
	lsCmd.Flags().BoolVar(&lsUnclaimed, "unclaimed", false, "Show only unclaimed active ports")
	lsCmd.Flags().BoolVar(&lsJSON, "json", false, "Output as JSON")
}

func runLs(cmd *cobra.Command, args []string) error {
	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	// build rows
	rows := make([]display.Row, 0, len(r.Claims))
	for port, c := range r.Claims {
		if lsProject != "" && c.Project != lsProject {
			continue
		}
		rows = append(rows, display.Row{
			Port:        port,
			Project:     c.Project,
			Service:     c.Service,
			Status:      display.StatusStopped, // TODO: populate live status via detector
			Description: c.Description,
		})
	}

	// apply filters
	if lsActive {
		rows = filterRows(
			rows,
			func(r display.Row) bool { return r.Status == display.StatusRunning },
		)
	}
	if lsFree {
		rows = filterRows(
			rows,
			func(r display.Row) bool { return r.Status == display.StatusStopped },
		)
	}

	// --unclaimed shows only active-but-unclaimed ports
	unclaimed := []display.Row{}
	if lsUnclaimed {
		rows = []display.Row{}
	}

	nc := display.NoColorEnabled(noColor)
	if lsJSON {
		return display.RenderJSON(out, rows, unclaimed)
	}
	return display.RenderTable(out, rows, unclaimed, nc)
}

func filterRows(rows []display.Row, keep func(display.Row) bool) []display.Row {
	result := make([]display.Row, 0)
	for _, r := range rows {
		if keep(r) {
			result = append(result, r)
		}
	}
	return result
}
