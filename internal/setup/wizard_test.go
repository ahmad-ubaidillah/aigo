package setup

import (
	"testing"
)

func TestNewSetupWizard(t *testing.T) {
	t.Parallel()
	w := NewSetupWizard()
	if w == nil {
		t.Error("expected wizard")
	}
	if w.mode != ModeCLI {
		t.Errorf("expected CLI mode, got %s", w.mode)
	}
	if w.complete {
		t.Error("expected not complete")
	}
}

func TestSetupWizard_IsComplete(t *testing.T) {
	t.Parallel()
	w := NewSetupWizard()
	if w.IsComplete() {
		t.Error("expected not complete")
	}
	w.complete = true
	if !w.IsComplete() {
		t.Error("expected complete")
	}
}

func TestSetupWizard_GetMode(t *testing.T) {
	t.Parallel()
	w := NewSetupWizard()
	if w.GetMode() != ModeCLI {
		t.Errorf("expected CLI, got %s", w.GetMode())
	}
	w.mode = ModeWeb
	if w.GetMode() != ModeWeb {
		t.Errorf("expected Web, got %s", w.GetMode())
	}
}

func TestModeConstants(t *testing.T) {
	t.Parallel()
	if ModeCLI != "cli" {
		t.Errorf("expected cli, got %s", ModeCLI)
	}
	if ModeWeb != "web" {
		t.Errorf("expected web, got %s", ModeWeb)
	}
}
