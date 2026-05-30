package cmd

import (
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
	if _, err := parsePort(port); err != nil {
		printError(err.Error(), "")
		return errSilent
	}

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	claim := r.GetClaim(port)
	ew := &iohelp.ErrWriter{W: out}

	if claim == nil {
		ew.Printf("port %s has no registered claim\n", port)
		return ew.Err
	}

	// TODO: will add an active-port check here: if the port is running,
	// warn and prompt for confirmation unless --force is set

	r.RemoveClaim(port)
	if err := r.Save(getRegistryPath()); err != nil {
		return err
	}

	ew.Printf("released port %s (%s/%s)\n", port, claim.Project, claim.Service)
	return ew.Err
}
