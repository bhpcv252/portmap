package display

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleRunning  = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	styleStopped  = lipgloss.NewStyle().Faint(true)                     // dim
	styleConflict = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	styleHeader   = lipgloss.NewStyle().Bold(true)
)

const (
	symRunning   = "● running"
	symStopped   = "○ stopped"
	symConflict  = "⚠ conflict"
	symUnclaimed = "● running (unclaimed)"
	symFree      = "○ free"
)

// respecting both the --no-color flag and the NO_COLOR env var (https://no-color.org)
func NoColorEnabled(flag bool) bool {
	return flag || os.Getenv("NO_COLOR") != ""
}

func renderStatus(s Status, noColor bool) string {
	plain := plainStatus(s)
	if noColor {
		return plain
	}
	switch s {
	case StatusRunning, StatusUnclaimed:
		return styleRunning.Render(plain)
	case StatusStopped, StatusFree:
		return styleStopped.Render(plain)
	case StatusConflict:
		return styleConflict.Render(plain)
	default:
		return plain
	}
}

func plainStatus(s Status) string {
	switch s {
	case StatusRunning:
		return symRunning
	case StatusStopped:
		return symStopped
	case StatusConflict:
		return symConflict
	case StatusUnclaimed:
		return symUnclaimed
	case StatusFree:
		return symFree
	default:
		return string(s)
	}
}
