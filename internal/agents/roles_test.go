package agents

import (
	"testing"
)

func TestRoles_ContainsAllFive(t *testing.T) {
	t.Parallel()

	expected := []string{"aigo", "atlas", "cody", "nova", "testa"}
	for _, name := range expected {
		if _, ok := Roles[name]; !ok {
			t.Errorf("expected role %q to exist", name)
		}
	}
	if len(Roles) != 5 {
		t.Errorf("expected 5 roles, got %d", len(Roles))
	}
}

func TestRoles_UniqueFields(t *testing.T) {
	t.Parallel()

	names := make(map[string]bool)
	prompts := make(map[string]bool)

	for _, r := range Roles {
		if names[r.Name] {
			t.Errorf("duplicate role name: %s", r.Name)
		}
		names[r.Name] = true

		if prompts[r.SystemPrompt] {
			t.Errorf("duplicate system prompt for role: %s", r.Name)
		}
		prompts[r.SystemPrompt] = true

		if r.MaxTurns <= 0 {
			t.Errorf("role %s has invalid MaxTurns: %d", r.Name, r.MaxTurns)
		}
		if len(r.Skills) == 0 {
			t.Errorf("role %s has no skills", r.Name)
		}
		if r.Category == "" {
			t.Errorf("role %s has empty category", r.Name)
		}
	}
}

func TestRoles_AigoCategory(t *testing.T) {
	t.Parallel()

	r, ok := Roles["aigo"]
	if !ok {
		t.Fatal("aigo role not found")
	}
	if r.Name != "Aigo" {
		t.Errorf("expected name 'Aigo', got %q", r.Name)
	}
	if r.Category != "ultrabrain" {
		t.Errorf("expected category 'ultrabrain', got %q", r.Category)
	}
	if r.MaxTurns != 20 {
		t.Errorf("expected MaxTurns 20, got %d", r.MaxTurns)
	}
}

func TestRoles_DeepCategoryRoles(t *testing.T) {
	t.Parallel()

	deepRoles := []string{"atlas", "cody", "nova", "testa"}
	for _, name := range deepRoles {
		r, ok := Roles[name]
		if !ok {
			t.Fatalf("role %q not found", name)
		}
		if r.Category != "deep" {
			t.Errorf("role %q expected category 'deep', got %q", name, r.Category)
		}
	}
}

func TestGetRole_ValidName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
	}{
		{"aigo", "Aigo"},
		{"atlas", "Atlas"},
		{"cody", "Cody"},
		{"nova", "Nova"},
		{"testa", "Testa"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			role, ok := GetRole(tt.name)
			if !ok {
				t.Fatalf("expected role %q to exist", tt.name)
			}
			if role.Name != tt.expected {
				t.Errorf("expected name %q, got %q", tt.expected, role.Name)
			}
		})
	}
}

func TestGetRole_UnknownName(t *testing.T) {
	t.Parallel()

	_, ok := GetRole("unknown")
	if ok {
		t.Error("expected false for unknown role")
	}
}

func TestListRoles(t *testing.T) {
	t.Parallel()

	roles := ListRoles()
	if len(roles) != 5 {
		t.Errorf("expected 5 roles, got %d", len(roles))
	}

	// Verify all expected names are present
	nameSet := make(map[string]bool)
	for _, r := range roles {
		nameSet[r.Name] = true
	}
	for _, expected := range []string{"Aigo", "Atlas", "Cody", "Nova", "Testa"} {
		if !nameSet[expected] {
			t.Errorf("expected role %q in ListRoles()", expected)
		}
	}
}
