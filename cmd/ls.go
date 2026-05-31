package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/display"
)

var lsJSON bool

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all active ports",
	Args:  cobra.NoArgs,
	RunE:  runLs,
}

func init() {
	lsCmd.Flags().BoolVar(&lsJSON, "json", false, "Output as JSON")
}

func runLs(cmd *cobra.Command, args []string) error {
	activePorts, err := getDetector().ActivePorts()
	if err != nil {
		return err
	}

	rows := make([]display.Row, 0, len(activePorts))
	for _, ap := range activePorts {
		rows = append(rows, display.Row{
			Port:    strconv.Itoa(ap.Port),
			PID:     ap.PID,
			Process: ap.Process,
			CWD:     ap.CWD,
		})
	}

	if lsJSON {
		return display.RenderJSON(out, rows)
	}
	return display.RenderTable(out, rows, display.NoColorEnabled(noColor))
}
