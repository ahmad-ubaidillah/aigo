package components

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#6366F1")).
			Padding(0, 2).
			Width(80)

	headerLeftStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	headerRightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E0E7FF"))

	statusIdleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Bold(true)

	statusActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22C55E")).
				Bold(true)

	statusBusyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EF4444")).
				Bold(true)
)

func statusBadge(status string) string {
	switch status {
	case "idle":
		return statusIdleStyle.Render("[IDLE]")
	case "active":
		return statusActiveStyle.Render("[ACTIVE]")
	case "busy":
		return statusBusyStyle.Render("[BUSY]")
	case "error":
		return statusErrorStyle.Render("[ERROR]")
	default:
		return statusIdleStyle.Render("[IDLE]")
	}
}

func Header(version, mode, session, status string) string {
	left := headerLeftStyle.Render("Aigo " + version)
	center := headerRightStyle.Render(mode)
	right := headerRightStyle.Render(session + " " + statusBadge(status))

	return headerStyle.Width(80).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, left, "  ", center, "  ", right),
	)
}
