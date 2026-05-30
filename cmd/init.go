package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/project"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a portmap.toml in the current directory",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := getCwd()
	if err != nil {
		return err
	}

	// single reader for the whole command so sequential prompts work correctly
	r := bufio.NewReader(in)
	readLine := func(label string) (string, error) {
		ew := &iohelp.ErrWriter{W: out}
		ew.Printf("  %s: ", label)
		if ew.Err != nil {
			return "", ew.Err
		}
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}

	ew := &iohelp.ErrWriter{W: out}
	ew.Println("initializing portmap for current directory\n")
	if ew.Err != nil {
		return ew.Err
	}

	projName, err := readLine("project name")
	if err != nil {
		return err
	}
	if projName == "" {
		projName = project.InferName(cwd)
	}

	ew2 := &iohelp.ErrWriter{W: out}
	ew2.Println("\nadd ports? (enter to finish)\n")
	if ew2.Err != nil {
		return ew2.Err
	}

	var ports []project.PortEntry
	for {
		portStr, err := readLine("port")
		if err != nil {
			return err
		}
		if portStr == "" {
			break
		}
		n, err := strconv.Atoi(portStr)
		if err != nil || n < 1 || n > 65535 {
			ew3 := &iohelp.ErrWriter{W: out}
			ew3.Printf("  invalid port %q, skipping\n\n", portStr)
			if ew3.Err != nil {
				return ew3.Err
			}
			continue
		}

		svc, err := readLine("service")
		if err != nil {
			return err
		}
		desc, err := readLine("description")
		if err != nil {
			return err
		}

		ports = append(ports, project.PortEntry{Port: n, Service: svc, Description: desc})

		ew4 := &iohelp.ErrWriter{W: out}
		ew4.Println("  -> added\n")
		if ew4.Err != nil {
			return ew4.Err
		}
	}

	outPath := filepath.Join(cwd, "portmap.toml")

	if _, statErr := os.Stat(outPath); statErr == nil {
		ew5 := &iohelp.ErrWriter{W: out}
		ew5.Printf("portmap.toml already exists. overwrite? [y/N] ")
		if ew5.Err != nil {
			return ew5.Err
		}
		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if strings.TrimSpace(strings.ToLower(line)) != "y" {
			return nil
		}
	}

	cfg := &project.Config{Project: projName, Ports: ports}
	if err := project.WriteConfig(outPath, cfg); err != nil {
		return fmt.Errorf("write portmap.toml: %w", err)
	}

	ew6 := &iohelp.ErrWriter{W: out}
	ew6.Printf("\nwrote portmap.toml with %d port(s)\n", len(ports))
	ew6.Println("run `portmap sync` to register them in your local registry")
	return ew6.Err
}
