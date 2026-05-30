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
	n, err := parsePort(port)
	if err != nil {
		printError(err.Error(), "")
		return errSilent
	}

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	active, ap, err := getDetector().IsActive(n)
	if err != nil {
		return err
	}

	nc := display.NoColorEnabled(noColor)
	claim := r.GetClaim(port)

	switch {
	case claim == nil && !active:
		// free: not claimed, nothing running
		if err := display.RenderCheck(out, display.CheckInfo{
			Port:   port,
			Status: display.StatusFree,
		}, nc); err != nil {
			return err
		}
		return nil // exit 0

	case claim == nil && active:
		// running but not in registry
		if err := display.RenderCheck(out, display.CheckInfo{
			Port:    port,
			PID:     ap.PID,
			Process: ap.Process,
			Status:  display.StatusUnclaimed,
		}, nc); err != nil {
			return err
		}
		return &ExitError{Code: 1}

	case claim != nil && active:
		// claimed and running
		// TODO: conflict detection requires stored PID; deferred to a future phase
		if err := display.RenderCheck(out, display.CheckInfo{
			Port:        port,
			Project:     claim.Project,
			Service:     claim.Service,
			Description: claim.Description,
			ClaimedAt:   claim.ClaimedAt,
			Path:        claim.Path,
			PID:         ap.PID,
			Process:     ap.Process,
			Status:      display.StatusRunning,
		}, nc); err != nil {
			return err
		}
		return &ExitError{Code: 1}

	default:
		// claimed but not running.
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
		return &ExitError{Code: 1}
	}
}
