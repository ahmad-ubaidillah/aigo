package config

import (
	"testing"
)

func TestProfile_Create(t *testing.T) {
	p := NewProfile("test-profile")
	if p.Name != "test-profile" {
		t.Errorf("Expected 'test-profile', got '%s'", p.Name)
	}
}

func TestProfileStorage_Save(t *testing.T) {
	store := NewProfileStorage()
	err := store.Save("profile1", &Profile{Name: "profile1"})
	if err != nil {
		t.Logf("Save error: %v", err)
	}
}

func TestProfileSwitch(t *testing.T) {
	sw := NewProfileSwitcher()
	sw.AddProfile(&Profile{Name: "p1"})
	sw.AddProfile(&Profile{Name: "p2"})

	current := sw.SwitchTo("p2")
	if current == nil {
		t.Error("SwitchTo should return profile")
	}
}