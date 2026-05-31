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
	ports, err := runLsofPorts()
	if err != nil {
		return nil, err
	}
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

func runLsofPorts() ([]ActivePort, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-Fpcn").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil // no listeners, not an error
		}
		return nil, err
	}
	return parseLsofFieldOutput(string(out)), nil
}

func parseLsofFieldOutput(output string) []ActivePort {
	var ports []ActivePort
	seen := make(map[string]bool)

	var currentPID int
	var currentCmd string

	for _, line := range strings.Split(output, "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(line[1:])
			if err == nil {
				currentPID = pid
			}
		case 'c':
			currentCmd = line[1:]
		case 'n':
			port, err := extractPort(line[1:])
			if err != nil {
				continue
			}
			key := fmt.Sprintf("%d:%d", currentPID, port)
			if seen[key] {
				continue
			}
			seen[key] = true
			ports = append(ports, ActivePort{Port: port, PID: currentPID, Process: currentCmd})
		}
	}
	return ports
}

func batchCWDs(pids []int) map[int]string {
	pidStrs := make([]string, len(pids))
	for i, pid := range pids {
		pidStrs[i] = strconv.Itoa(pid)
	}

	cmd := exec.Command("lsof",
		"-p", strings.Join(pidStrs, ","),
		"-a", "-d", "cwd",
		"-Fn",
	)
	out, _ := cmd.Output()
	if len(out) == 0 {
		return nil
	}

	cwds := make(map[int]string)
	var currentPID int

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(line[1:])
			if err == nil {
				currentPID = pid
			}
		case 'n':
			cwds[currentPID] = line[1:]
		}
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
