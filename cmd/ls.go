package cmd

import (
	"strconv"

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

	// build a set of active ports
	activePorts, err := getDetector().ActivePorts()
	if err != nil {
		return err
	}
	activeMap := make(map[int]struct{}, len(activePorts))
	for _, ap := range activePorts {
		activeMap[ap.Port] = struct{}{}
	}

	// build claimed rows
	rows := make([]display.Row, 0, len(r.Claims))
	claimedPorts := make(map[int]bool)
	for portStr, c := range r.Claims {
		if lsProject != "" && c.Project != lsProject {
			continue
		}
		port, _ := strconv.Atoi(portStr)
		claimedPorts[port] = true

		status := display.StatusStopped
		if _, ok := activeMap[port]; ok {
			status = display.StatusRunning
		}
		rows = append(rows, display.Row{
			Port:        portStr,
			Project:     c.Project,
			Service:     c.Service,
			Status:      status,
			Description: c.Description,
		})
	}

	// build unclaimed rows: active ports not in the registry
	unclaimed := []display.Row{}
	for _, ap := range activePorts {
		if !claimedPorts[ap.Port] {
			unclaimed = append(unclaimed, display.Row{
				Port: strconv.Itoa(ap.Port),
			})
		}
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
	if lsUnclaimed {
		rows = []display.Row{}
	}
	if !lsUnclaimed {
		unclaimed = func() []display.Row {
			if lsUnclaimed {
				return unclaimed
			}
			if lsActive || lsFree {
				return []display.Row{} // hide unclaimed when filtering by claimed status
			}
			return unclaimed
		}()
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
