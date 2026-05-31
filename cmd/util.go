package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/bhpcv252/portmap/internal/detector"
	"github.com/bhpcv252/portmap/internal/iohelp"
)

var (
	out    io.Writer = os.Stdout
	errOut io.Writer = os.Stderr
	in     io.Reader = os.Stdin
)

var det detector.Detector

func getDetector() detector.Detector {
	if det != nil {
		return det
	}
	return detector.New()
}

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
}

func parsePort(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("invalid port %q: must be a number between 1 and 65535", s)
	}
	return n, nil
}

func confirm(prompt string) (bool, error) {
	ew := &iohelp.ErrWriter{W: out}
	ew.Printf("%s [y/N] ", prompt)
	if ew.Err != nil {
		return false, ew.Err
	}
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(strings.ToLower(line)) == "y", nil
}
