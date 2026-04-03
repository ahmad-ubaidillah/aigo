package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDoctor(t *testing.T) {
	t.Parallel()

	results := RunDoctor()
	if len(results) == 0 {
		t.Fatal("expected results")
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
		if r.Status == "" {
			t.Errorf("%s has empty status", r.Name)
		}
		if r.Detail == "" {
			t.Errorf("%s has empty detail", r.Name)
		}
	}

	expected := []string{"Go Runtime", "Config", "OpenCode", "Database", "Skills", "Environment"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing doctor check: %s", name)
		}
	}
}

func TestGenerateShellCompletion_Bash(t *testing.T) {
	t.Parallel()

	out := GenerateShellCompletion("bash")
	if !strings.Contains(out, "_aigo") {
		t.Error("expected bash completion to contain _aigo")
	}
}

func TestGenerateShellCompletion_Zsh(t *testing.T) {
	t.Parallel()

	out := GenerateShellCompletion("zsh")
	if !strings.Contains(out, "#compdef aigo") {
		t.Error("expected zsh completion header")
	}
}

func TestGenerateShellCompletion_Fish(t *testing.T) {
	t.Parallel()

	out := GenerateShellCompletion("fish")
	if !strings.Contains(out, "complete -c aigo") {
		t.Error("expected fish completion")
	}
}

func TestGenerateShellCompletion_Unsupported(t *testing.T) {
	t.Parallel()

	out := GenerateShellCompletion("powershell")
	if !strings.Contains(out, "unsupported") {
		t.Errorf("expected unsupported message, got %s", out)
	}
}

func TestGenerateExampleConfig(t *testing.T) {
	t.Parallel()

	out := GenerateExampleConfig()
	if !strings.Contains(out, "model:") {
		t.Error("expected model section")
	}
	if !strings.Contains(out, "memory:") {
		t.Error("expected memory section")
	}
	if !strings.Contains(out, "web:") {
		t.Error("expected web section")
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	if cfg.Model.Default == "" {
		t.Error("expected default model")
	}
	if cfg.OpenCode.Timeout <= 0 {
		t.Error("expected positive timeout")
	}
	if cfg.OpenCode.MaxTurns <= 0 {
		t.Error("expected positive max turns")
	}
	if cfg.Memory.MaxL0Items <= 0 {
		t.Error("expected positive max L0 items")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Model.Default = "test-model"

	err := SaveConfig(cfg, path)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Model.Default != "test-model" {
		t.Errorf("expected test-model, got %s", loaded.Model.Default)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig("/nonexistent/config.yaml")
	if err != nil {
		t.Logf("got error (expected): %v", err)
	}
	if cfg.Model.Default == "" {
		t.Error("expected default config returned")
	}
}

func TestConfigPaths(t *testing.T) {
	t.Parallel()

	paths := ConfigPaths()
	if len(paths) == 0 {
		t.Error("expected at least one config path")
	}
}

func TestEnvConfig_ApplyToConfig(t *testing.T) {
	t.Parallel()

	envCfg := EnvConfig{
		ModelDefault:     "env-model",
		OpenCodeTimeout:  60,
		OpenCodeMaxTurns: 100,
	}

	cfg := DefaultConfig()
	envCfg.ApplyToConfig(&cfg)

	if cfg.Model.Default != "env-model" {
		t.Errorf("expected env-model, got %s", cfg.Model.Default)
	}
	if cfg.OpenCode.Timeout != 60 {
		t.Errorf("expected 60, got %d", cfg.OpenCode.Timeout)
	}
	if cfg.OpenCode.MaxTurns != 100 {
		t.Errorf("expected 100, got %d", cfg.OpenCode.MaxTurns)
	}
}

func TestEnvConfig_ToMap(t *testing.T) {
	t.Parallel()

	envCfg := EnvConfig{
		ModelDefault: "test-model",
		WebPort:      ":9090",
	}

	m := envCfg.ToMap()
	if m["AIGO_MODEL_DEFAULT"] != "test-model" {
		t.Errorf("expected test-model, got %s", m["AIGO_MODEL_DEFAULT"])
	}
	if m["AIGO_WEB_PORT"] != ":9090" {
		t.Errorf("expected :9090, got %s", m["AIGO_WEB_PORT"])
	}
}

func TestEnvConfig_FromMap(t *testing.T) {
	t.Parallel()

	m := map[string]string{
		"AIGO_MODEL_DEFAULT":    "map-model",
		"AIGO_OPENCODE_TIMEOUT": "45",
		"AIGO_WEB_ENABLED":      "true",
	}

	envCfg := EnvConfig{}.FromMap(m)
	if envCfg.ModelDefault != "map-model" {
		t.Errorf("expected map-model, got %s", envCfg.ModelDefault)
	}
	if envCfg.OpenCodeTimeout != 45 {
		t.Errorf("expected 45, got %d", envCfg.OpenCodeTimeout)
	}
	if !envCfg.WebEnabled {
		t.Error("expected web enabled")
	}
}

func TestLoadEnvFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("AIGO_MODEL_DEFAULT=envfile-model\nAIGO_OPENCODE_TIMEOUT=120\n"), 0644)

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	vars, err := LoadEnvFile()
	if err != nil {
		t.Fatal(err)
	}
	if vars["AIGO_MODEL_DEFAULT"] != "envfile-model" {
		t.Errorf("expected envfile-model, got %s", vars["AIGO_MODEL_DEFAULT"])
	}
}

func TestLoadEnvFile_Invalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("INVALID_LINE_NO_EQUALS\n# comment\n\nKEY=value\n"), 0644)

	origCwd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origCwd)

	vars, err := LoadEnvFile()
	if err != nil {
		t.Fatal(err)
	}
	if vars["KEY"] != "value" {
		t.Errorf("expected value, got %s", vars["KEY"])
	}
	if _, exists := vars["INVALID_LINE_NO_EQUALS"]; exists {
		t.Error("expected invalid line to be skipped")
	}
}

func TestPrintDoctorResults(t *testing.T) {
	t.Parallel()

	results := []DoctorResult{
		{Name: "Test", Status: "ok", Detail: "all good"},
		{Name: "Warn", Status: "warn", Detail: "warning"},
		{Name: "Info", Status: "info", Detail: "info"},
	}
	PrintDoctorResults(results)
}

func TestPrintVersion(t *testing.T) {
	t.Parallel()

	PrintVersion()
}

func TestResolveConfigPath_FlagPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("model:\n  default: flag-model\n"), 0644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Model.Default != "flag-model" {
		t.Errorf("expected flag-model, got %s", cfg.Model.Default)
	}
}

func TestResolveConfigPath_FlagPathNotFound(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Logf("got error (expected): %v", err)
	}
	if cfg.Model.Default == "" {
		t.Error("expected default config")
	}
}

func TestEnvConfig_EmptyValues(t *testing.T) {
	t.Parallel()

	envCfg := EnvConfig{}
	cfg := DefaultConfig()
	original := cfg.Model.Default
	envCfg.ApplyToConfig(&cfg)
	if cfg.Model.Default != original {
		t.Errorf("expected %s, got %s", original, cfg.Model.Default)
	}
}
