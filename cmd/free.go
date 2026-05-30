package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/registry"
)

var freeForce bool

var freeCmd = &cobra.Command{
	Use:   "free <port>",
	Short: "Release a claim on a port",
	Args:  cobra.ExactArgs(1),
	RunE:  runFree,
}

func init() {
	freeCmd.Flags().
		BoolVar(&freeForce, "force", false, "Release without confirmation even if active")
}

func runFree(cmd *cobra.Command, args []string) error {
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

	ew := &iohelp.ErrWriter{W: out}
	claim := r.GetClaim(port)
	if claim == nil {
		ew.Printf("port %s has no registered claim\n", port)
		return ew.Err
	}

	// warn if the port is currently active
	active, ap, err := getDetector().IsActive(n)
	if err != nil {
		return err
	}
	if active && !freeForce {
		pid := 0
		if ap != nil {
			pid = ap.PID
		}
		yes, err := confirm(
			fmt.Sprintf("port %s is still running (pid %d). release anyway?", port, pid),
		)
		if err != nil {
			return err
		}
		if !yes {
			return nil
		}
	}

	r.RemoveClaim(port)
	if err := r.Save(getRegistryPath()); err != nil {
		return err
	}

	ew.Printf("released port %s (%s/%s)\n", port, claim.Project, claim.Service)
	return ew.Err
}
