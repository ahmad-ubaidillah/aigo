package tui

import tea "github.com/charmbracelet/bubbletea"

func Run() error {
	model := NewModel()
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)
	_, err := p.Run()
	return err
}
