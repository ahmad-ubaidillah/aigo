package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	agentPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6366F1")).
			Padding(1, 2).
			Margin(1, 0)

	agentTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#818CF8")).
			MarginBottom(1)

	agentNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E7FF"))

	agentStatusIdleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#94A3B8"))

	agentStatusActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22C55E"))

	agentStatusBusyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F59E0B"))

	agentStatusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444"))
)

func agentStatusDot(status string) string {
	switch status {
	case "idle":
		return agentStatusIdleStyle.Render("○")
	case "active":
		return agentStatusActiveStyle.Render("●")
	case "busy":
		return agentStatusBusyStyle.Render("●")
	case "error":
		return agentStatusErrorStyle.Render("✗")
	default:
		return agentStatusIdleStyle.Render("○")
	}
}

func AgentFleet(agents map[string]string) string {
	if len(agents) == 0 {
		return agentPanelStyle.Width(30).Render("No agents available")
	}

	var lines []string
	lines = append(lines, agentTitleStyle.Render("Agent Fleet"))

	for name, status := range agents {
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			agentStatusDot(status),
			" ",
			agentNameStyle.Render(name),
		)
		lines = append(lines, line)
	}

	return agentPanelStyle.Width(30).Render(strings.Join(lines, "\n"))
}
