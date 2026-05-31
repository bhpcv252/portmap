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
	ports := parseLsofOutput(out)
	if len(ports) == 0 {
		return ports, nil
	}

	pids := make([]int, len(ports))
	for i, p := range ports {
		pids[i] = p.PID
	}
	cwds := batchCWDs(pids)
	for i := range ports {
		ports[i].CWD = cwds[ports[i].PID]
	}
	return ports, nil
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
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil
		}
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 127 {
			return "", fmt.Errorf("lsof not found in PATH (exit 127)")
		}
		return "", err
	}
	return string(out), nil
}

func parseLsofOutput(output string) []ActivePort {
	var ports []ActivePort
	seen := make(map[string]bool) // "pid:port" dedup key

	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] {
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
		key := fmt.Sprintf("%d:%d", pid, port)
		if seen[key] {
			continue
		}
		seen[key] = true
		ports = append(ports, ActivePort{Port: port, PID: pid, Process: command})
	}
	return ports
}

// batchCWDs fetches the working directory for multiple PIDs in a single lsof call.
//
// lsof field output format (with -Fnt, which always includes p and f fields):
//
//	p<pid>   process ID
//	fcwd     file descriptor = cwd  ← this signals the cwd entry
//	tDIR     type = directory        ← ignored; nextIsCWD stays true
//	n<path>  the actual path
//	ftxt     next file descriptor    ← resets nextIsCWD
//
// The error is intentionally ignored: SIP-protected system processes cause lsof
// to exit non-zero, but stdout still contains valid data for accessible PIDs.
func batchCWDs(pids []int) map[int]string {
	pidStrs := make([]string, len(pids))
	for i, pid := range pids {
		pidStrs[i] = strconv.Itoa(pid)
	}

	cmd := exec.Command("lsof", "-p", strings.Join(pidStrs, ","), "-Fnt")
	out, _ := cmd.Output() // ignore exit code; parse whatever stdout we got
	if len(out) == 0 {
		return nil
	}

	cwds := make(map[int]string)
	var currentPID int
	var nextIsCWD bool

	for _, line := range strings.Split(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "p"):
			pid, err := strconv.Atoi(line[1:])
			if err == nil {
				currentPID = pid
				nextIsCWD = false
			}
		case line == "fcwd":
			// file descriptor is cwd; the n<path> line follows (possibly after tDIR)
			nextIsCWD = true
		case strings.HasPrefix(line, "f"):
			// any other file descriptor entry; cwd is no longer next
			nextIsCWD = false
		case nextIsCWD && strings.HasPrefix(line, "n"):
			cwds[currentPID] = line[1:]
			nextIsCWD = false
		}
		// t<type> and other lines are deliberately not handled so they
		// do not disturb the nextIsCWD flag set by fcwd
	}
	return cwds
}

func extractPort(addr string) (int, error) {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0, fmt.Errorf("no colon in address: %s", addr)
	}
	return strconv.Atoi(addr[idx+1:])
}
