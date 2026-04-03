package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	activityPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#6366F1")).
				Padding(1, 2).
				Margin(1, 0)

	activityTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#818CF8")).
				MarginBottom(1)

	activityEntryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CBD5E1")).
				Padding(0, 1)
)

func ActivityFeed(entries []string) string {
	if len(entries) == 0 {
		return activityPanelStyle.Render(activityTitleStyle.Render("Live Activity") +
			"\n" + activityEntryStyle.Render("No activity yet"))
	}

	var lines []string
	lines = append(lines, activityTitleStyle.Render("Live Activity"))

	maxEntries := 10
	if len(entries) > maxEntries {
		entries = entries[len(entries)-maxEntries:]
	}

	for _, entry := range entries {
		lines = append(lines, activityEntryStyle.Render("▸ "+entry))
	}

	return activityPanelStyle.Render(strings.Join(lines, "\n"))
}
