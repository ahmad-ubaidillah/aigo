package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	logsPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6366F1")).
			Padding(1, 2)

	logsTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#818CF8")).
			MarginBottom(1)

	logInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#60A5FA"))

	logWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FBBF24"))

	logErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171"))

	logDebugStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8"))

	logDefaultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CBD5E1"))
)

func colorizeLog(entry string) string {
	lower := strings.ToLower(entry)

	switch {
	case strings.HasPrefix(lower, "error"):
		return logErrorStyle.Render(entry)
	case strings.HasPrefix(lower, "warn"):
		return logWarnStyle.Render(entry)
	case strings.HasPrefix(lower, "debug"):
		return logDebugStyle.Render(entry)
	case strings.HasPrefix(lower, "info"):
		return logInfoStyle.Render(entry)
	default:
		return logDefaultStyle.Render(entry)
	}
}

func LogsView(entries []string) string {
	if len(entries) == 0 {
		return logsPanelStyle.Render(logsTitleStyle.Render("Logs") +
			"\n" + logDefaultStyle.Render("No log entries"))
	}

	var lines []string
	lines = append(lines, logsTitleStyle.Render("Logs"))

	for _, entry := range entries {
		lines = append(lines, colorizeLog(entry))
	}

	return logsPanelStyle.Render(strings.Join(lines, "\n"))
}
