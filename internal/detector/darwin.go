//go:build darwin

package detector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type darwinDetector struct{}

func New() Detector { return &darwinDetector{} }

func (d *darwinDetector) ActivePorts() ([]ActivePort, error) {
	out, err := runLsof()
	if err != nil {
		return nil, err
	}
	return parseLsofOutput(out), nil
}

func (d *darwinDetector) IsActive(port int) (bool, *ActivePort, error) {
	all, err := d.ActivePorts()
	if err != nil {
		return false, nil, err
	}
	for _, ap := range all {
		if ap.Port == port {
			return true, &ap, nil
		}
	}
	return false, nil, nil
}

func runLsof() (string, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN").Output()
	if err != nil {
		// lsof exits 1 when no results
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}
		// exit code 127 means lsof not found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 127 {
			return "", fmt.Errorf("lsof not found in PATH (exit 127)")
		}
		return "", err
	}
	return string(out), nil
}

// parseLsofOutput parses lsof -nP -iTCP -sTCP:LISTEN output
// The NAME column contains the address in forms: *:3000, 127.0.0.1:3000, [::1]:3000
func parseLsofOutput(output string) []ActivePort {
	var ports []ActivePort
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		command := fields[0]
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		name := fields[8]
		name = strings.TrimSuffix(name, " (LISTEN)")
		port, err := extractPort(name)
		if err != nil {
			continue
		}
		ports = append(ports, ActivePort{Port: port, PID: pid, Process: command})
	}
	return ports
}

// extractPort pulls the numeric port from address strings like *:3000, 127.0.0.1:3000, [::1]:3000
func extractPort(addr string) (int, error) {
	// find the last colon to handle both IPv4 and IPv6 addresses
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0, fmt.Errorf("no colon in address: %s", addr)
	}
	return strconv.Atoi(addr[idx+1:])
}
