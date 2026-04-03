package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	interactionPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#6366F1")).
				Padding(1, 2)

	interactionFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#22C55E")).
				Padding(1, 2)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E7FF")).
			Padding(0, 1)

	inputFocusedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4F46E5")).
				Padding(0, 1)

	inputNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#94A3B8")).
				Padding(0, 1)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#818CF8")).
				Bold(true)
)

func InteractionPanel(messages []string, input string, focused bool) string {
	var parts []string

	for _, msg := range messages {
		parts = append(parts, messageStyle.Render(msg))
	}

	var inputStyle lipgloss.Style
	if focused {
		inputStyle = inputFocusedStyle
	} else {
		inputStyle = inputNormalStyle
	}

	inputLine := lipgloss.JoinHorizontal(lipgloss.Left,
		inputPromptStyle.Render("❯ "),
		inputStyle.Render(input),
	)

	parts = append(parts, inputLine)

	panelStyle := interactionPanelStyle
	if focused {
		panelStyle = interactionFocusedStyle
	}

	return panelStyle.Render(strings.Join(parts, "\n"))
}
