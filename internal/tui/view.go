package tui

import (
	"github.com/ahmad-ubaidillah/aigo/internal/tui/components"
	"github.com/ahmad-ubaidillah/aigo/internal/tui/views"
	"github.com/charmbracelet/lipgloss"
)

var (
	dashboardStyle = lipgloss.NewStyle().
			Padding(0, 1)

	leftPanelStyle = lipgloss.NewStyle().
			Width(32)

	rightPanelStyle = lipgloss.NewStyle()
)

func (m Model) View() string {
	if m.activeView != "dashboard" {
		return m.renderAltView()
	}

	header := components.Header("0.1.0", "Dashboard", "Session: "+m.sessionID, m.agentStatus)

	leftTop := components.AgentFleet(m.agents)
	leftBottom := components.ActivityFeed(m.activity)
	leftCol := lipgloss.JoinVertical(lipgloss.Top, leftTop, leftBottom)

	rightTop := m.renderContextTools()
	rightBottom := components.InteractionPanel(m.messages, m.input, m.focused)
	rightCol := lipgloss.JoinVertical(lipgloss.Top, rightTop, rightBottom)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		leftPanelStyle.Render(leftCol),
		rightPanelStyle.Render(rightCol),
	)

	focusText := "true"
	if !m.focused {
		focusText = "false"
	}

	statusBar := components.StatusBar(focusText, "Dashboard", []string{
		"Tab:switch", "Enter:send", "Esc:clear", "F2:kanban", "F3:logs", "F4:tools", "Ctrl+C:quit",
	})

	return dashboardStyle.Render(lipgloss.JoinVertical(lipgloss.Top, header, body, statusBar))
}

func (m Model) renderContextTools() string {
	toolNames := make([]string, 0, len(m.tools))
	for name := range m.tools {
		toolNames = append(toolNames, name)
	}

	if len(toolNames) == 0 {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6366F1")).
			Padding(1, 2).
			Render("No tools configured")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6366F1")).
		Padding(1, 2).
		Render("Tools: " + lipgloss.NewStyle().Foreground(lipgloss.Color("#818CF8")).Render(string(rune(len(toolNames))+'0')))
}

func (m Model) renderAltView() string {
	header := components.Header("0.1.0", m.activeView, "Session: "+m.sessionID, m.agentStatus)

	var body string
	switch m.activeView {
	case "kanban":
		body = views.KanbanView(m.kanban)
	case "logs":
		body = views.LogsView(m.logs)
	case "tools":
		body = views.ToolsView(m.tools)
	default:
		body = "Unknown view"
	}

	focusText := "true"
	if !m.focused {
		focusText = "false"
	}

	statusBar := components.StatusBar(focusText, m.activeView, []string{
		"Esc:back", "Ctrl+C:quit",
	})

	return dashboardStyle.Render(lipgloss.JoinVertical(lipgloss.Top, header, body, statusBar))
}
