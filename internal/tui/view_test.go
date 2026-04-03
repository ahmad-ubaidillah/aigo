package tui

import (
	"strings"
	"testing"
)

func TestView(t *testing.T) {
	t.Parallel()
	m := NewModel()
	output := m.View()
	if output == "" {
		t.Error("expected non-empty view")
	}
	if !strings.Contains(output, "Aigo") {
		t.Error("expected Aigo in view")
	}
}
