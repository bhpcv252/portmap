package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/registry"
)

var killYes bool

var killCmd = &cobra.Command{
	Use:   "kill <port>",
	Short: "Kill the process running on a port",
	Args:  cobra.ExactArgs(1),
	RunE:  runKill,
}

func init() {
	killCmd.Flags().BoolVarP(&killYes, "yes", "y", false, "Skip confirmation prompt")
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

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	// show process and claim info before prompting
	ew := &iohelp.ErrWriter{W: out}
	ew.Printf("port %s is being used by:\n", port)
	ew.Printf("  pid:      %d\n", ap.PID)
	if ap.Process != "" {
		ew.Printf("  process:  %s\n", ap.Process)
	}
	if claim := r.GetClaim(port); claim != nil {
		ew.Printf("  project:  %s/%s (claimed)\n", claim.Project, claim.Service)
	}
	ew.Printf("\n")
	if ew.Err != nil {
		return ew.Err
	}

	if !killYes {
		yes, err := confirm(fmt.Sprintf("kill process %d?", ap.PID))
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
	ew2.Printf("process %d killed. port %s is now free.\n", ap.PID, port)
	return ew2.Err
}
