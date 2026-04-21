package ui

import (
	"testing"
)

func TestTheme_Dark(t *testing.T) {
	t.Log("Testing dark theme")
	dark := GetTheme("dark")
	if dark == nil {
		t.Fatal("Dark theme should exist")
	}
}

func TestTheme_Light(t *testing.T) {
	t.Log("Testing light theme")
	light := GetTheme("light")
	if light == nil {
		t.Fatal("Light theme should exist")
	}
}

func TestThemeSwitch(t *testing.T) {
	ts := NewThemeSwitcher()
	ts.SetTheme("dark")

	if ts.Current() != "dark" {
		t.Error("Theme should be dark")
	}
}