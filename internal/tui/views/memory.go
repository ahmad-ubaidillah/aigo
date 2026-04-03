package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	memoryTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#818CF8")).
				Bold(true)

	memoryItemStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("#475569")).
			Padding(0, 1).
			MarginBottom(1)

	memoryTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#22C55E")).
			Padding(0, 1)
)

func MemoryView(memories []map[string]string, search string) string {
	var b strings.Builder

	b.WriteString(memoryTitleStyle.Render("Memory Browser"))
	b.WriteString("\n\n")

	if search != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("Search: " + search))
		b.WriteString("\n\n")
	}

	if len(memories) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render("No memories stored yet."))
		return b.String()
	}

	for i, m := range memories {
		content := m["content"]
		category := m["category"]
		tags := m["tags"]
		created := m["created_at"]

		if len(content) > 100 {
			content = content[:100] + "..."
		}

		var itemParts []string
		itemParts = append(itemParts, content)

		if category != "" {
			itemParts = append(itemParts, memoryTagStyle.Render("["+category+"]"))
		}
		if tags != "" {
			itemParts = append(itemParts, memoryTagStyle.Render("#"+strings.ReplaceAll(tags, ",", " #")))
		}
		if created != "" {
			itemParts = append(itemParts, lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render(created))
		}

		b.WriteString(memoryItemStyle.Render(strings.Join(itemParts, "  ")))

		if i < len(memories)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render(fmt.Sprintf("%d memories", len(memories))))

	return b.String()
}
