package views

import (
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	toolsPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6366F1")).
			Padding(1, 2)

	toolsTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#818CF8")).
			MarginBottom(1)

	toolEnabledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22C55E"))

	toolDisabledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#64748B"))

	toolNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E7FF"))
)

func ToolsView(tools map[string]bool) string {
	if len(tools) == 0 {
		return toolsPanelStyle.Render(toolsTitleStyle.Render("Tools Config") +
			"\n" + toolDisabledStyle.Render("No tools configured"))
	}

	var lines []string
	lines = append(lines, toolsTitleStyle.Render("Tools Config"))

	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		enabled := tools[name]
		var badge string
		if enabled {
			badge = toolEnabledStyle.Render("[ON]")
		} else {
			badge = toolDisabledStyle.Render("[OFF]")
		}

		line := lipgloss.JoinHorizontal(lipgloss.Left,
			badge,
			" ",
			toolNameStyle.Render(name),
		)
		lines = append(lines, line)
	}

	return toolsPanelStyle.Render(strings.Join(lines, "\n"))
}
