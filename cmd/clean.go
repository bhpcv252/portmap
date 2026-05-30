package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/registry"
)

var (
	cleanDryRun bool
	cleanYes    bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove claims whose project path no longer exists",
	Args:  cobra.NoArgs,
	RunE:  runClean,
}

func init() {
	cleanCmd.Flags().
		BoolVar(&cleanDryRun, "dry-run", false, "Show what would be removed without writing")
	cleanCmd.Flags().BoolVarP(&cleanYes, "yes", "y", false, "Skip confirmation")
}

func runClean(cmd *cobra.Command, args []string) error {
	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	type stale struct {
		port  string
		claim registry.Claim
	}
	var stales []stale
	for port, c := range r.Claims {
		if c.Path == "" {
			continue
		}
		if _, statErr := os.Stat(c.Path); os.IsNotExist(statErr) {
			stales = append(stales, stale{port, c})
		}
	}

	ew := &iohelp.ErrWriter{W: out}

	if len(stales) == 0 {
		ew.Println("no stale claims found")
		return ew.Err
	}

	sort.Slice(stales, func(i, j int) bool {
		pi, _ := strconv.Atoi(stales[i].port)
		pj, _ := strconv.Atoi(stales[j].port)
		return pi < pj
	})

	ew.Println("stale claims found:\n")
	for _, s := range stales {
		ew.Printf("  port %s  %s/%s  -> path %s not found\n",
			s.port, s.claim.Project, s.claim.Service, s.claim.Path)
	}
	ew.Printf("\n")
	if ew.Err != nil {
		return ew.Err
	}

	if cleanDryRun {
		return nil
	}

	if !cleanYes {
		yes, confirmErr := confirm(fmt.Sprintf("remove %d stale claim(s)?", len(stales)))
		if confirmErr != nil {
			return confirmErr
		}
		if !yes {
			return nil
		}
	}

	for _, s := range stales {
		r.RemoveClaim(s.port)
	}
	if err := r.Save(getRegistryPath()); err != nil {
		return err
	}

	ew2 := &iohelp.ErrWriter{W: out}
	ew2.Printf("removed %d stale claim(s)\n", len(stales))
	return ew2.Err
}
