package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/project"
	"github.com/bhpcv252/portmap/internal/registry"
)

var (
	claimProject string
	claimService string
	claimDesc    string
	claimForce   bool
)

var claimCmd = &cobra.Command{
	Use:   "claim <port>",
	Short: "Register a port for your project",
	Args:  cobra.ExactArgs(1),
	RunE:  runClaim,
}

func init() {
	claimCmd.Flags().StringVarP(&claimProject, "project", "p", "", "Project name")
	claimCmd.Flags().StringVarP(&claimService, "service", "s", "", "Service name")
	claimCmd.Flags().StringVarP(&claimDesc, "desc", "d", "", "Short description")
	claimCmd.Flags().BoolVar(&claimForce, "force", false, "Overwrite an existing claim")
}

func runClaim(cmd *cobra.Command, args []string) error {
	port := args[0]
	if _, err := parsePort(port); err != nil {
		printError(err.Error(), "")
		return errSilent
	}

	cwd, err := getCwd()
	if err != nil {
		return err
	}

	proj := claimProject
	if proj == "" {
		proj = project.InferName(cwd)
	}

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	// snapshot the existing claim before modification to choose the right success message
	prev := r.GetClaim(port)

	if err := r.AddClaim(port, proj, claimService, claimDesc, cwd, claimForce); err != nil {
		printError(err.Error(),
			fmt.Sprintf("use --force to overwrite, or run `portmap free %s` first", port))
		return errSilent
	}

	if err := r.Save(getRegistryPath()); err != nil {
		return err
	}

	ew := &iohelp.ErrWriter{W: out}
	if prev != nil && prev.Project == proj && prev.Service == claimService {
		ew.Printf("updated port %s (%s/%s)\n", port, proj, claimService)
	} else {
		ew.Printf("claimed port %s for %s/%s\n", port, proj, claimService)
	}
	return ew.Err
}
