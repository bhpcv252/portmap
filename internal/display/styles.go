package display

import "github.com/charmbracelet/lipgloss"

const symRunning = "● running"

var styleRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

func NoColorEnabled(noColor bool) bool { return noColor }

func RenderRunning(noColor bool) string {
	if noColor {
		return symRunning
	}
	return styleRunning.Render(symRunning)
}
