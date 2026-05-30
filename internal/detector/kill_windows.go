//go:build windows

package detector

import (
	"os/exec"
	"strconv"
)

// KillProcess forcefully terminates a process by PID using taskkill
func KillProcess(pid int) error {
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F").Run()
}
