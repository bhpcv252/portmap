package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/display"
	"github.com/bhpcv252/portmap/internal/iohelp"
)

var killYes bool

var killCmd = &cobra.Command{
	Use:   "kill <port>",
	Short: "Kill the process running on a port",
	Args:  cobra.ExactArgs(1),
	RunE:  runKill,
}

func init() {
	killCmd.Flags().BoolVarP(&killYes, "yes", "y", false, "Skip confirmation")
}

func runKill(cmd *cobra.Command, args []string) error {
	port := args[0]
	n, err := parsePort(port)
	if err != nil {
		printError(err.Error(), "")
		return errSilent
	}

	active, ap, err := getDetector().IsActive(n)
	if err != nil {
		return err
	}
	if !active {
		printError(
			fmt.Sprintf("nothing is running on port %s", port),
			"nothing to kill",
		)
		return errSilent
	}

	nc := display.NoColorEnabled(noColor)
	ew := &iohelp.ErrWriter{W: out}
	ew.Printf("port %s  %s\n", port, display.RenderRunning(nc))
	ew.Printf("  pid:      %d\n", ap.PID)
	if ap.Process != "" {
		ew.Printf("  process:  %s\n", ap.Process)
	}
	if ap.CWD != "" {
		ew.Printf("  cwd:      %s\n", ap.CWD)
	}
	ew.Println("")
	if ew.Err != nil {
		return ew.Err
	}

	if !killYes {
		yes, err := confirm(fmt.Sprintf("kill process on port %s?", port))
		if err != nil {
			return err
		}
		if !yes {
			return nil
		}
	}

	if err := detector.KillProcess(ap.PID); err != nil {
		return fmt.Errorf("kill pid %d: %w", ap.PID, err)
	}

	ew2 := &iohelp.ErrWriter{W: out}
	ew2.Printf("killed pid %d\n", ap.PID)
	return ew2.Err
}
