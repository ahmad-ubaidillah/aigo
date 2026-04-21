package ui

import (
	"testing"
)

func TestCustomTheme(t *testing.T) {
	ct := &CustomTheme{
		Name:   "custom",
		Colors: map[string]string{"bg": "#000", "fg": "#fff"},
	}
	if ct.Name != "custom" {
		t.Error("Custom theme name mismatch")
	}
}

func TestSSE_Connect(t *testing.T) {
	sse := NewSSE()
	ch := make(chan string, 1)
	sse.Add(ch)

	if len(sse.clients) != 1 {
		t.Error("Client should be added")
	}
}

func TestDesignTokens(t *testing.T) {
	tokens := GetDesignTokens()
	if tokens == nil {
		t.Fatal("Design tokens should exist")
	}
}

func TestRichCLI(t *testing.T) {
	cli := NewRichCLI()
	output := cli.Format("test", "bold")

	if output == "" {
		t.Error("Output should not be empty")
	}
}

func TestAnimation(t *testing.T) {
	anim := NewAnimation("spin")
	if anim == nil {
		t.Fatal("Animation should be created")
	}
}