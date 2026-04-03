package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	inactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)
)

func Run() error {
	model := NewModel()
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)
	_, err := p.Run()
	return err
}

func headerView(width int) string {
	title := headerStyle.Render(" Aigo ")
	tabs := strings.Join([]string{
		"  [D]ashboard  ",
		"  [S]kills  ",
		"  [T]asks  ",
		"  [L]ogs  ",
		"  [H]elp  ",
	}, "")
	return title + tabs
}

func sidebarView(m Model) string {
	status := activeStyle.Render("● Running") + " " + m.sessionID
	info := fmt.Sprintf("Tasks: %d | Agents: %d", len(m.taskQueue), len(m.agents))
	return boxStyle.Render(fmt.Sprintf("%s\n%s", status, info))
}

func contentView(m Model, view string) string {
	switch view {
	case "dashboard":
		return dashboardView(m)
	case "skills":
		return skillsView(m)
	case "tasks":
		return tasksView(m)
	case "logs":
		return logsView(m)
	case "help":
		return helpView()
	default:
		return dashboardView(m)
	}
}

func dashboardView(m Model) string {
	output := boxStyle.Render("Sessions\n" + m.sessionID)
	output += "\n" + boxStyle.Render("Messages\n"+fmt.Sprintf("%d messages", len(m.messages)))
	return output
}

func skillsView(m Model) string {
	list := []string{
		"git-master: Git operations",
		"playwright: Browser automation",
		"frontend-ui-ux: Frontend dev",
		"code-review: Code review",
		"web-search: Web search",
		"code-search: Code search",
	}
	return boxStyle.Render("Available Skills\n" + strings.Join(list, "\n"))
}

func tasksView(m Model) string {
	if len(m.taskQueue) == 0 {
		return boxStyle.Render("Tasks\nNo tasks in queue")
	}
	return boxStyle.Render("Tasks\n" + strings.Join(m.taskQueue, "\n"))
}

func logsView(m Model) string {
	if len(m.logs) == 0 {
		return boxStyle.Render("Logs\nNo logs available")
	}
	return boxStyle.Render("Logs\n" + strings.Join(m.logs, "\n"))
}

func helpView() string {
	help := []string{
		"Tab: Switch view",
		"Enter: Submit input",
		"Ctrl+C: Quit",
		"",
		"Views:",
		"  D - Dashboard",
		"  S - Skills",
		"  T - Tasks",
		"  L - Logs",
		"  H - Help",
	}
	return boxStyle.Render("Help\n" + strings.Join(help, "\n"))
}

func inputView() string {
	prompt := activeStyle.Render("> ")
	return inputStyle.Render(prompt + "|")
}

func nextView(current string) string {
	views := []string{"dashboard", "skills", "tasks", "logs", "help"}
	for i, v := range views {
		if v == current {
			return views[(i+1)%len(views)]
		}
	}
	return "dashboard"
}
