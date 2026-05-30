package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bhpcv252/portmap/internal/iohelp"
	"github.com/bhpcv252/portmap/internal/registry"
)

var (
	out    io.Writer = os.Stdout
	errOut io.Writer = os.Stderr
)

var errSilent = errors.New("")

type ExitError struct {
	Code int
}

func (e *ExitError) Error() string { return "" }

func printError(msg, hint string) {
	ew := &iohelp.ErrWriter{W: errOut}
	ew.Printf("error: %s\n", msg)
	if hint != "" {
		ew.Printf("hint:  %s\n", hint)
	}
	// ew.Err is intentionally dropped, we are already in an error path
}

var registryPathOverride string

func getRegistryPath() string {
	if registryPathOverride != "" {
		return registryPathOverride
	}
	return registry.DefaultPath()
}

var cwdOverride string

func getCwd() (string, error) {
	if cwdOverride != "" {
		return cwdOverride, nil
	}
	return os.Getwd()
}

func parsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("invalid port %q: must be a number between 1 and 65535", s)
	}
	return n, nil
}
