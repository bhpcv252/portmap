//go:build linux || darwin

package detector

import (
	"os"
	"syscall"
	"time"
)

// KillProcess sends SIGTERM and waits up to 2 seconds before sending SIGKILL
func KillProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return err
	}
	// poll every 100ms to see if the process exited gracefully
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		// Signal(0) tests process existence without sending a real signal
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			return nil // process is gone
		}
	}
	return proc.Signal(syscall.SIGKILL)
}
