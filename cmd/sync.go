package cmd

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/project"
	"github.com/bhpcv252/portmap/internal/registry"
)

var (
	syncDryRun bool
	syncForce  bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Register ports from portmap.toml into your local registry",
	Args:  cobra.NoArgs,
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().
		BoolVar(&syncDryRun, "dry-run", false, "Show what would be registered without writing")
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Overwrite existing conflicting claims")
}

func runSync(cmd *cobra.Command, args []string) error {
	cwd, err := getCwd()
	if err != nil {
		return err
	}

	_, cfg, err := project.FindConfig(cwd)
	if err != nil {
		return err
	}
	if cfg == nil {
		printError(
			"no portmap.toml found in current directory or any parent directory",
			"run `portmap init` to create one",
		)
		return errSilent
	}

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	ew := &iohelp.ErrWriter{W: out}
	ew.Printf("syncing from portmap.toml (project: %s)\n\n", cfg.Project)

	claimed, skipped := 0, 0
	for _, pe := range cfg.Ports {
		portStr := strconv.Itoa(pe.Port)
		existing := r.GetClaim(portStr)
		if existing != nil && !syncForce {
			ew.Printf("  port %d  %s  -> already claimed by %s/%s, skipping\n",
				pe.Port, pe.Service, existing.Project, existing.Service)
			skipped++
			continue
		}
		if !syncDryRun {
			if addErr := r.AddClaim(portStr, cfg.Project, pe.Service, pe.Description, cwd, syncForce); addErr != nil {
				ew.Printf("  port %d  %s  -> error: %s\n", pe.Port, pe.Service, addErr.Error())
				skipped++
				continue
			}
		}
		action := "claimed"
		if syncDryRun {
			action = "would claim"
		}
		ew.Printf("  port %d  %s  -> %s\n", pe.Port, pe.Service, action)
		claimed++
	}

	ew.Printf("\n")
	if syncDryRun {
		ew.Printf("dry run: %d port(s) would be synced. %d skipped.\n", claimed, skipped)
		return ew.Err
	}
	ew.Printf("synced %d port(s). %d skipped.\n", claimed, skipped)
	if ew.Err != nil {
		return ew.Err
	}
	return r.Save(getRegistryPath())
}
