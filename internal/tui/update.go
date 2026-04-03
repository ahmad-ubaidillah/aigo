package tui

import (
	"github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "enter":
		if m.input != "" && m.focused {
			m.messages = append(m.messages, "> "+m.input)
			m.input = ""
		}
		return m, nil

	case "esc":
		if m.activeView != "dashboard" {
			m.activeView = "dashboard"
			return m, nil
		}
		m.input = ""
		return m, nil

	case "tab":
		m.focused = !m.focused
		return m, nil

	case "f2":
		m.activeView = "kanban"
		return m, nil

	case "f3":
		m.activeView = "logs"
		return m, nil

	case "f4":
		m.activeView = "tools"
		return m, nil

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil

	default:
		if m.focused && len(msg.String()) == 1 {
			m.input += msg.String()
		}
		return m, nil
	}
}
