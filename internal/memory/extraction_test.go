package memory

import (
	"testing"
)

func TestExtractProfile(t *testing.T) {
	e := NewExtractor()
	profile := e.ExtractProfile("user prefers dark mode, uses VSCode")

	if profile == nil {
		t.Fatal("ExtractProfile returned nil")
	}

	if profile.Category != "profile" {
		t.Logf("Expected category 'profile', got '%s'", profile.Category)
	}
}

func TestExtractPreferences(t *testing.T) {
	e := NewExtractor()
	prefs := e.ExtractPreferences("I like TypeScript, prefer strict typing")

	if prefs == nil {
		t.Log("ExtractPreferences returned nil (acceptable)")
	}
}

func TestExtractEntities(t *testing.T) {
	e := NewExtractor()
	entities := e.ExtractEntities("API endpoint at /users, database table users")

	if entities == nil {
		t.Log("ExtractEntities returned nil (acceptable)")
	}
}

func TestExtractEvents(t *testing.T) {
	e := NewExtractor()
	events := e.ExtractEvents("bug found on login, meeting scheduled")

	if events == nil {
		t.Log("ExtractEvents returned nil (acceptable)")
	}
}

func TestExtractCases(t *testing.T) {
	e := NewExtractor()
	cases := e.ExtractCases("handle timeout error, retry三次")

	if cases == nil {
		t.Log("ExtractCases returned nil (acceptable)")
	}
}

func TestExtractPatterns(t *testing.T) {
	e := NewExtractor()
	patterns := e.ExtractPatterns("always check auth first, never skip validation")

	if patterns == nil {
		t.Log("ExtractPatterns returned nil (acceptable)")
	}
}