package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1B4B")).
			Foreground(lipgloss.Color("#CBD5E1")).
			Padding(0, 2)

	focusIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22C55E")).
				Bold(true)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8"))

	keyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#818CF8")).
			Bold(true)
)

func StatusBar(focused, mode string, keyHints []string) string {
	focusText := "unfocused"
	if focused == "true" {
		focusText = "focused"
	}

	left := focusIndicatorStyle.Render("["+focusText+"]") + " " + mode

	var hints []string
	for _, hint := range keyHints {
		hints = append(hints, keyStyle.Render(hint))
	}

	right := keyHintStyle.Render(strings.Join(hints, "  "))

	return statusBarStyle.Width(80).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, left, "  ", right),
	)
}
