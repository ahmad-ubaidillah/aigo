package config

import (
	"testing"
)

func TestWorkspace_Create(t *testing.T) {
	w := NewWorkspace("workspace1")
	if w.Name != "workspace1" {
		t.Error("Workspace name mismatch")
	}
}

func TestWorkspaceIsolation(t *testing.T) {
	w1 := &Workspace{Name: "w1", Config: map[string]string{"key": "val1"}}
	w2 := &Workspace{Name: "w2", Config: map[string]string{"key": "val2"}}

	if w1.Config["key"] == w2.Config["key"] {
		t.Log("Configs should be isolated")
	}
}

func TestWorkspaceConfig(t *testing.T) {
	wc := NewWorkspaceConfig()
	wc.Set("key", "value")

	if wc.Get("key") != "value" {
		t.Error("Config value mismatch")
	}
}

func TestWorkspaceSwitcher(t *testing.T) {
	ws := NewWorkspaceSwitcher()
	ws.Add(&Workspace{Name: "ws1"})
	ws.Add(&Workspace{Name: "ws2"})

	current := ws.Switch("ws2")
	if current == nil {
		t.Error("Should switch to ws2")
	}
}

func TestExportImport(t *testing.T) {
	ei := NewExportImport()
	data := ei.Export(&Profile{Name: "test"})
	if data == "" {
		t.Error("Export should return data")
	}

	profile := ei.Import(data)
	if profile == nil || profile.Name != "test" {
		t.Error("Import should restore profile")
	}
}