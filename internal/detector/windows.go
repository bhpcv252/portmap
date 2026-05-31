//go:build windows

package detector

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type windowsDetector struct{}

func New() Detector { return &windowsDetector{} }

func (d *windowsDetector) ActivePorts() ([]ActivePort, error) {
	netOut, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat: %w", err)
	}
	entries := parseNetstatOutput(string(netOut))

	pids := make(map[int]bool)
	for _, e := range entries {
		pids[e.pid] = true
	}
	names := tasklistNames(pids)

	var ports []ActivePort
	for _, e := range entries {
		ports = append(ports, ActivePort{
			Port:    e.port,
			PID:     e.pid,
			Process: names[e.pid],
			// CWD on Windows requires WMI or P/Invoke; not implemented, sorry
			// please feel free to open a PR
		})
	}
	return ports, nil
}

func (d *windowsDetector) IsActive(port int) (bool, *ActivePort, error) {
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

type netstatEntry struct {
	port int
	pid  int
}

func parseNetstatOutput(output string) []netstatEntry {
	var entries []netstatEntry
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if !strings.EqualFold(fields[3], "LISTENING") {
			continue
		}
		port, err := extractPort(fields[1])
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}
		entries = append(entries, netstatEntry{port: port, pid: pid})
	}
	return entries
}

func tasklistNames(pids map[int]bool) map[int]string {
	names := make(map[int]string)
	out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil {
		return names
	}
	return parseTasklist(string(out), pids)
}

func parseTasklist(output string, pids map[int]bool) map[int]string {
	names := make(map[int]string)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		name := strings.Trim(parts[0], `"`)
		pidStr := strings.Trim(parts[1], `"`)
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		if pids[pid] {
			names[pid] = name
		}
	}
	return names
}

func extractPort(addr string) (int, error) {
	idx := strings.LastIndex(addr, ":")
	if idx < 0 {
		return 0, fmt.Errorf("no colon in address: %s", addr)
	}
	return strconv.Atoi(addr[idx+1:])
}
