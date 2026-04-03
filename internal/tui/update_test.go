package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestUpdate_WindowSizeMsg(t *testing.T) {
	t.Parallel()
	m := NewModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	newModel := result.(Model)
	if newModel.width != 80 {
		t.Errorf("expected width 80, got %d", newModel.width)
	}
	if newModel.height != 24 {
		t.Errorf("expected height 24, got %d", newModel.height)
	}
}
