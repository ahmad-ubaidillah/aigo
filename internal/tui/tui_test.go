package tui

import (
	"testing"
)

func TestModel(t *testing.T) {
	t.Parallel()
	m := NewModel()
	if m.sessionID != "default" {
		t.Errorf("expected default sessionID, got %s", m.sessionID)
	}
}

func TestModel_SetSessionID(t *testing.T) {
	t.Parallel()
	m := NewModel()
	m.sessionID = "test-session"
	if m.sessionID != "test-session" {
		t.Errorf("expected test-session, got %s", m.sessionID)
	}
}

func TestModel_SetInput(t *testing.T) {
	t.Parallel()
	m := NewModel()
	m.input = "hello"
	if m.input != "hello" {
		t.Errorf("expected hello, got %s", m.input)
	}
}
