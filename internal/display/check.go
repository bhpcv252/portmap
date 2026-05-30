package display

import (
	"io"
	"time"

	"github.com/bhpcv252/portmap/internal/iohelp"
)

type CheckInfo struct {
	Port        string
	Project     string // empty if not claimed
	Service     string
	Description string
	ClaimedAt   time.Time // zero if not claimed
	Path        string
	PID         int    // 0 if not running
	Process     string // empty if not running
	Status      Status
}

func RenderCheck(w io.Writer, info CheckInfo, noColor bool) error {
	ew := &iohelp.ErrWriter{W: w}

	ew.Printf("port %s\n", info.Port)
	ew.Printf("  status:      %s\n", checkStatus(info, noColor))

	if info.Project != "" {
		ew.Printf("  project:     %s\n", info.Project)
		ew.Printf("  service:     %s\n", info.Service)
		ew.Printf("  description: %s\n", info.Description)
		ew.Printf("  claimed at:  %s\n", info.ClaimedAt.Format("2006-01-02 15:04"))
		if info.Path != "" {
			ew.Printf("  path:        %s\n", info.Path)
		}
	}

	if info.PID > 0 {
		ew.Printf("  pid:         %d\n", info.PID)
		if info.Process != "" {
			ew.Printf("  process:     %s\n", info.Process)
		}
	}

	switch {
	case info.Project == "" && info.PID == 0:
		ew.Println("  no claim registered, nothing running")
	case info.Project == "" && info.PID > 0:
		ew.Println("  no claim registered for this port")
	}

	return ew.Err
}

func checkStatus(info CheckInfo, noColor bool) string {
	switch {
	case info.Status == StatusFree:
		if noColor {
			return symFree
		}
		return styleStopped.Render(symFree)
	case info.PID > 0 && info.Project == "":
		if noColor {
			return symUnclaimed
		}
		return styleRunning.Render(symUnclaimed)
	case info.PID > 0:
		if noColor {
			return symRunning
		}
		return styleRunning.Render(symRunning)
	default:
		if noColor {
			return symStopped
		}
		return styleStopped.Render(symStopped)
	}
}
