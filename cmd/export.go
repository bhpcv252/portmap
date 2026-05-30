package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/project"
	"github.com/bhpcv252/portmap/internal/registry"
)

var (
	exportProject string
	exportOutput  string
	exportStdout  bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export project claims as a portmap.toml file",
	Args:  cobra.NoArgs,
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportProject, "project", "p", "", "Project to export")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "./portmap.toml", "Output file path")
	exportCmd.Flags().
		BoolVar(&exportStdout, "stdout", false, "Print to stdout instead of writing a file")
}

func runExport(cmd *cobra.Command, args []string) error {
	cwd, err := getCwd()
	if err != nil {
		return err
	}

	proj := exportProject
	if proj == "" {
		proj = project.InferName(cwd)
	}

	r, err := registry.Load(getRegistryPath())
	if err != nil {
		return err
	}

	type entry struct {
		port int
		svc  string
		desc string
	}
	var matches []entry
	for portStr, c := range r.Claims {
		if c.Project != proj {
			continue
		}
		n, parseErr := strconv.Atoi(portStr)
		if parseErr != nil {
			continue
		}
		matches = append(matches, entry{n, c.Service, c.Description})
	}

	if len(matches) == 0 {
		ew := &iohelp.ErrWriter{W: out}
		ew.Printf("no claims found for project %q\n", proj)
		return ew.Err
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].port < matches[j].port })

	cfg := &project.Config{Project: proj}
	for _, m := range matches {
		cfg.Ports = append(cfg.Ports, project.PortEntry{
			Port:        m.port,
			Service:     m.svc,
			Description: m.desc,
		})
	}

	if exportStdout {
		return project.WriteConfigTo(out, cfg)
	}

	if err := project.WriteConfig(exportOutput, cfg); err != nil {
		return fmt.Errorf("write %s: %w", exportOutput, err)
	}

	ew := &iohelp.ErrWriter{W: out}
	ew.Printf("wrote %s with %d port(s)\n", exportOutput, len(cfg.Ports))
	return ew.Err
}
