package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Provider.Default != "openai" {
		t.Errorf("expected default provider 'openai', got '%s'", cfg.Provider.Default)
	}
	if cfg.Provider.Model != "gpt-4o-mini" {
		t.Errorf("expected default model 'gpt-4o-mini', got '%s'", cfg.Provider.Model)
	}
	if cfg.Agent.MaxIterations != 90 {
		t.Errorf("expected max iterations 90, got %d", cfg.Agent.MaxIterations)
	}
	if cfg.Agent.LoopDetection != true {
		t.Error("expected loop detection enabled")
	}
	if cfg.Memory.Enabled != true {
		t.Error("expected memory enabled")
	}
	if cfg.WebUI.Port != 8080 {
		t.Errorf("expected webui port 8080, got %d", cfg.WebUI.Port)
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if path == "" {
		t.Error("config path should not be empty")
	}
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", filepath.Base(path))
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	content := `
provider:
  default: nous
  base_url: https://api.nousresearch.com/v1
  api_key: test-key-123
  model: hermes-3
agent:
  max_iterations: 50
  max_tokens: 2048
  loop_detection: true
memory:
  enabled: true
  use_fts5: false
channels:
  telegram:
    enabled: true
    token: "123456:ABC"
  discord:
    enabled: false
  slack:
    enabled: false
  websocket:
    enabled: true
    port: 9090
webui:
  enabled: true
  port: 3000
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Provider.Default != "nous" {
		t.Errorf("expected provider 'nous', got '%s'", cfg.Provider.Default)
	}
	if cfg.Provider.Model != "hermes-3" {
		t.Errorf("expected model 'hermes-3', got '%s'", cfg.Provider.Model)
	}
	if cfg.Agent.MaxIterations != 50 {
		t.Errorf("expected max_iterations 50, got %d", cfg.Agent.MaxIterations)
	}
	if !cfg.Channels.Telegram.Enabled {
		t.Error("expected telegram enabled")
	}
	if cfg.Channels.Telegram.Token != "123456:ABC" {
		t.Errorf("expected telegram token '123456:ABC', got '%s'", cfg.Channels.Telegram.Token)
	}
	if cfg.WebUI.Port != 3000 {
		t.Errorf("expected webui port 3000, got %d", cfg.WebUI.Port)
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	os.Setenv("AIGO_PROVIDER", "anthropic")
	os.Setenv("AIGO_MODEL", "claude-3")
	os.Setenv("AIGO_API_KEY", "env-key")
	defer func() {
		os.Unsetenv("AIGO_PROVIDER")
		os.Unsetenv("AIGO_MODEL")
		os.Unsetenv("AIGO_API_KEY")
	}()

	cfg, _ := Load("/nonexistent") // Falls back to defaults + env

	if cfg.Provider.Default != "anthropic" {
		t.Errorf("expected provider 'anthropic', got '%s'", cfg.Provider.Default)
	}
	if cfg.Provider.Model != "claude-3" {
		t.Errorf("expected model 'claude-3', got '%s'", cfg.Provider.Model)
	}
	if cfg.Provider.APIKey != "env-key" {
		t.Errorf("expected api_key 'env-key', got '%s'", cfg.Provider.APIKey)
	}
}

func TestGetAPIKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Provider.APIKey = "direct-key"
	if cfg.GetAPIKey() != "direct-key" {
		t.Error("expected direct API key")
	}

	cfg2 := DefaultConfig()
	cfg2.Provider.Providers = map[string]ProviderEntry{
		"openai": {APIKey: "provider-key"},
	}
	if cfg2.GetAPIKey() != "provider-key" {
		t.Error("expected provider entry API key")
	}
}

func TestGetBaseURL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Provider.BaseURL = "http://localhost:8000"
	if cfg.GetBaseURL() != "http://localhost:8000" {
		t.Error("expected custom base URL")
	}

	cfg2 := DefaultConfig()
	cfg2.Provider.Default = "ollama"
	if cfg2.GetBaseURL() != "http://localhost:11434" {
		t.Errorf("expected ollama default URL, got '%s'", cfg2.GetBaseURL())
	}

	cfg3 := DefaultConfig()
	cfg3.Provider.Default = "anthropic"
	if cfg3.GetBaseURL() != "https://api.anthropic.com" {
		t.Errorf("expected anthropic default URL, got '%s'", cfg3.GetBaseURL())
	}
}
