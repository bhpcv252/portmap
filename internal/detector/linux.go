//go:build linux

package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type linuxDetector struct{}

func New() Detector { return &linuxDetector{} }

func (d *linuxDetector) ActivePorts() ([]ActivePort, error) {
	var ports []ActivePort
	seen := make(map[int]bool) // deduplicate ports that appear in both tcp and tcp6

	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		entries := parseProcNetTCP(string(data))
		for _, e := range entries {
			if seen[e.port] {
				continue
			}
			seen[e.port] = true

			pid, _ := mapInodeToPID(e.inode, "/proc")
			proc := ""
			cwd := ""
			if pid > 0 {
				proc = readComm(pid)
				cwd = readCWD(pid)
			}
			ports = append(ports, ActivePort{Port: e.port, PID: pid, Process: proc, CWD: cwd})
		}
	}
	return ports, nil
}

func (d *linuxDetector) IsActive(port int) (bool, *ActivePort, error) {
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

type netEntry struct {
	port  int
	inode string
}

func parseProcNetTCP(content string) []netEntry {
	var entries []netEntry
	lines := strings.Split(content, "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		if fields[3] != "0A" {
			continue
		}
		parts := strings.Split(fields[1], ":")
		if len(parts) < 2 {
			continue
		}
		portHex := parts[len(parts)-1]
		port, err := strconv.ParseInt(portHex, 16, 32)
		if err != nil {
			continue
		}
		entries = append(entries, netEntry{port: int(port), inode: fields[9]})
	}
	return entries
}

func mapInodeToPID(inode, procRoot string) (int, error) {
	target := fmt.Sprintf("socket:[%s]", inode)

	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return 0, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fdDir := filepath.Join(procRoot, e.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				return pid, nil
			}
		}
	}
	return 0, nil
}

func readComm(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func readCWD(pid int) string {
	cwd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return ""
	}
	return cwd
}
