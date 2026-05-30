package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/display"
	"github.com/bhpcv252/portmap/internal/registry"
)

var checkCmd = &cobra.Command{
	Use:   "check <port>",
	Short: "Show the status of a single port",
	Args:  cobra.ExactArgs(1),
	RunE:  runCheck,
}

func runCheck(cmd *cobra.Command, args []string) error {
	port := args[0]
	if _, err := parsePort(port); err != nil {
		printError(err.Error(), "")
		return errSilent
	}

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	nc := display.NoColorEnabled(noColor)
	claim := r.GetClaim(port)

	// TODO: will add active-port detection here, enabling the running,
	// unclaimed-active, and conflict states

	if claim == nil {
		if err := display.RenderCheck(out, display.CheckInfo{
			Port:   port,
			Status: display.StatusFree,
		}, nc); err != nil {
			return err
		}
		return nil // exit 0: port is free
	}

	if err := display.RenderCheck(out, display.CheckInfo{
		Port:        port,
		Project:     claim.Project,
		Service:     claim.Service,
		Description: claim.Description,
		ClaimedAt:   claim.ClaimedAt,
		Path:        claim.Path,
		Status:      display.StatusStopped,
	}, nc); err != nil {
		return err
	}

	return &ExitError{Code: 1} // exit 1: port is claimed
}
