//go:build !linux && !darwin && !windows

package detector

import "fmt"

func KillProcess(pid int) error {
	return fmt.Errorf("kill not supported on this platform")
}
