package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/display"
	"github.com/bhpcv252/portmap/internal/iohelp"
)

var checkCmd = &cobra.Command{
	Use:   "check <port>",
	Short: "Show what is running on a port",
	Args:  cobra.ExactArgs(1),
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
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

	ew := &iohelp.ErrWriter{W: out}

	if !active {
		ew.Printf("port %s\n", port)
		ew.Println("  not running")
		return ew.Err // exit 0: port is free
	}

	nc := display.NoColorEnabled(noColor)
	ew.Printf("port %s\n", port)
	ew.Printf("  status:   %s\n", display.RenderRunning(nc))
	ew.Printf("  pid:      %d\n", ap.PID)
	if ap.Process != "" {
		ew.Printf("  process:  %s\n", ap.Process)
	}
	if ap.CWD != "" {
		ew.Printf("  cwd:      %s\n", ap.CWD)
	}
	if ew.Err != nil {
		return ew.Err
	}
	return &ExitError{Code: 1} // exit 1: port is in use
}
