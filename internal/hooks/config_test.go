package hooks

import (
	"os"
	"testing"
)

func TestHookConfig_New(t *testing.T) {
	config := NewDefaultHookConfig()
	if config.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", config.Version)
	}

	if !config.IsEnabled("on_start") {
		t.Error("on_start should be enabled by default")
	}
}

func TestHookConfig_IsEnabled(t *testing.T) {
	config := NewDefaultHookConfig()
	config.SetEnabled("on_start", false)

	if config.IsEnabled("on_start") {
		t.Error("on_start should be disabled")
	}
}

func TestHookConfig_GetOrderedHooks(t *testing.T) {
	config := NewDefaultHookConfig()
	hooks := config.GetOrderedHooks()

	if len(hooks) == 0 {
		t.Error("Should have ordered hooks")
	}
}

func TestHookConfig_EnableAll(t *testing.T) {
	config := NewDefaultHookConfig()
	config.DisableAll()

	config.EnableAll()

	if !config.IsEnabled("on_start") {
		t.Error("on_start should be enabled after EnableAll")
	}
}

func TestHookConfig_DisableAll(t *testing.T) {
	config := NewDefaultHookConfig()

	config.DisableAll()

	if config.IsEnabled("on_start") {
		t.Error("on_start should be disabled")
	}
}

func TestHookConfig_LoadSave(t *testing.T) {
	tmpFile := "/tmp/hooks_test.yaml"
	os.Remove(tmpFile)

	config := NewDefaultHookConfig()
	config.SetEnabled("on_error", false)

	err := SaveHookConfig(tmpFile, config)
	if err != nil {
		t.Fatalf("SaveHookConfig failed: %v", err)
	}
	defer os.Remove(tmpFile)

	loaded, err := LoadHookConfig(tmpFile)
	if err != nil {
		t.Fatalf("LoadHookConfig failed: %v", err)
	}

	if loaded.IsEnabled("on_error") {
		t.Error("on_error should be disabled in loaded config")
	}
}