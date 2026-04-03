package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	kanbanContainerStyle = lipgloss.NewStyle().
				Padding(1, 2)

	kanbanColumnStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#6366F1")).
				Padding(1, 2).
				Width(20)

	kanbanTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#818CF8")).
				MarginBottom(1)

	kanbanItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E7FF")).
			Padding(0, 1).
			MarginBottom(1)
)

func renderColumn(title string, items []string) string {
	var lines []string
	lines = append(lines, kanbanTitleStyle.Render(title))

	if len(items) == 0 {
		lines = append(lines, kanbanItemStyle.Foreground(lipgloss.Color("#64748B")).Render("empty"))
	} else {
		for _, item := range items {
			lines = append(lines, kanbanItemStyle.Render("• "+item))
		}
	}

	return kanbanColumnStyle.Render(strings.Join(lines, "\n"))
}

func KanbanView(columns map[string][]string) string {
	order := []string{"Backlog", "In Progress", "Review", "Done"}

	var cols []string
	for _, name := range order {
		items := columns[name]
		cols = append(cols, renderColumn(name, items))
	}

	return kanbanContainerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, cols...),
	)
}
