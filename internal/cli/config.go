package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	"gopkg.in/yaml.v3"
)

type EnvConfig struct {
	ModelDefault       string `yaml:"model_default"`
	ModelCoding        string `yaml:"model_coding"`
	ModelIntent        string `yaml:"model_intent"`
	OpenCodeBinary     string `yaml:"opencode_binary"`
	OpenCodeTimeout    int    `yaml:"opencode_timeout"`
	OpenCodeMaxTurns   int    `yaml:"opencode_max_turns"`
	GatewayEnabled     bool   `yaml:"gateway_enabled"`
	MemoryMaxL0        int    `yaml:"memory_max_l0"`
	MemoryMaxL1        int    `yaml:"memory_max_l1"`
	MemoryAutoCompress bool   `yaml:"memory_auto_compress"`
	WebEnabled         bool   `yaml:"web_enabled"`
	WebPort            string `yaml:"web_port"`
	WebAuthEnabled     bool   `yaml:"web_auth_enabled"`
	WebAuthUser        string `yaml:"web_auth_user"`
	WebAuthPass        string `yaml:"web_auth_pass"`
}

func LoadEnvFile() (map[string]string, error) {
	envVars := make(map[string]string)

	paths := []string{
		".env",
		".aigo/.env",
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		paths = append(paths, filepath.Join(home, ".aigo", ".env"))
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			envVars[key] = value
		}
	}

	return envVars, nil
}

func (e EnvConfig) ApplyToConfig(cfg *types.Config) {
	if e.ModelDefault != "" {
		cfg.Model.Default = e.ModelDefault
	}
	if e.ModelCoding != "" {
		cfg.Model.Coding = e.ModelCoding
	}
	if e.ModelIntent != "" {
		cfg.Model.Intent = e.ModelIntent
	}
	if e.OpenCodeBinary != "" {
		cfg.OpenCode.Binary = e.OpenCodeBinary
	}
	if e.OpenCodeTimeout > 0 {
		cfg.OpenCode.Timeout = e.OpenCodeTimeout
	}
	if e.OpenCodeMaxTurns > 0 {
		cfg.OpenCode.MaxTurns = e.OpenCodeMaxTurns
	}
	if e.GatewayEnabled {
		cfg.Gateway.Enabled = e.GatewayEnabled
	}
	if e.MemoryMaxL0 > 0 {
		cfg.Memory.MaxL0Items = e.MemoryMaxL0
	}
	if e.MemoryMaxL1 > 0 {
		cfg.Memory.MaxL1Items = e.MemoryMaxL1
	}
	cfg.Memory.AutoCompress = e.MemoryAutoCompress
	cfg.Web.Enabled = e.WebEnabled
	if e.WebPort != "" {
		cfg.Web.Port = e.WebPort
	}
	cfg.Web.Auth.Enabled = e.WebAuthEnabled
	if e.WebAuthUser != "" {
		cfg.Web.Auth.Username = e.WebAuthUser
	}
	if e.WebAuthPass != "" {
		cfg.Web.Auth.Password = e.WebAuthPass
	}
}

func (e EnvConfig) ToMap() map[string]string {
	result := make(map[string]string)
	if e.ModelDefault != "" {
		result["AIGO_MODEL_DEFAULT"] = e.ModelDefault
	}
	if e.ModelCoding != "" {
		result["AIGO_MODEL_CODING"] = e.ModelCoding
	}
	if e.ModelIntent != "" {
		result["AIGO_MODEL_INTENT"] = e.ModelIntent
	}
	if e.OpenCodeBinary != "" {
		result["AIGO_OPENCODE_BINARY"] = e.OpenCodeBinary
	}
	if e.OpenCodeTimeout > 0 {
		result["AIGO_OPENCODE_TIMEOUT"] = strconv.Itoa(e.OpenCodeTimeout)
	}
	if e.OpenCodeMaxTurns > 0 {
		result["AIGO_OPENCODE_MAX_TURNS"] = strconv.Itoa(e.OpenCodeMaxTurns)
	}
	if e.GatewayEnabled {
		result["AIGO_GATEWAY_ENABLED"] = "true"
	}
	if e.MemoryMaxL0 > 0 {
		result["AIGO_MEMORY_MAX_L0"] = strconv.Itoa(e.MemoryMaxL0)
	}
	if e.MemoryMaxL1 > 0 {
		result["AIGO_MEMORY_MAX_L1"] = strconv.Itoa(e.MemoryMaxL1)
	}
	result["AIGO_MEMORY_AUTO_COMPRESS"] = strconv.FormatBool(e.MemoryAutoCompress)
	result["AIGO_WEB_ENABLED"] = strconv.FormatBool(e.WebEnabled)
	if e.WebPort != "" {
		result["AIGO_WEB_PORT"] = e.WebPort
	}
	result["AIGO_WEB_AUTH_ENABLED"] = strconv.FormatBool(e.WebAuthEnabled)
	if e.WebAuthUser != "" {
		result["AIGO_WEB_AUTH_USER"] = e.WebAuthUser
	}
	if e.WebAuthPass != "" {
		result["AIGO_WEB_AUTH_PASS"] = e.WebAuthPass
	}
	return result
}

func (e EnvConfig) FromMap(m map[string]string) EnvConfig {
	prefix := "AIGO_"
	for k, v := range m {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		key := strings.TrimPrefix(k, prefix)
		switch key {
		case "MODEL_DEFAULT":
			e.ModelDefault = v
		case "MODEL_CODING":
			e.ModelCoding = v
		case "MODEL_INTENT":
			e.ModelIntent = v
		case "OPENCODE_BINARY":
			e.OpenCodeBinary = v
		case "OPENCODE_TIMEOUT":
			e.OpenCodeTimeout, _ = strconv.Atoi(v)
		case "OPENCODE_MAX_TURNS":
			e.OpenCodeMaxTurns, _ = strconv.Atoi(v)
		case "GATEWAY_ENABLED":
			e.GatewayEnabled, _ = strconv.ParseBool(v)
		case "MEMORY_MAX_L0":
			e.MemoryMaxL0, _ = strconv.Atoi(v)
		case "MEMORY_MAX_L1":
			e.MemoryMaxL1, _ = strconv.Atoi(v)
		case "MEMORY_AUTO_COMPRESS":
			e.MemoryAutoCompress, _ = strconv.ParseBool(v)
		case "WEB_ENABLED":
			e.WebEnabled, _ = strconv.ParseBool(v)
		case "WEB_PORT":
			e.WebPort = v
		case "WEB_AUTH_ENABLED":
			e.WebAuthEnabled, _ = strconv.ParseBool(v)
		case "WEB_AUTH_USER":
			e.WebAuthUser = v
		case "WEB_AUTH_PASS":
			e.WebAuthPass = v
		}
	}
	return e
}

func DefaultConfig() types.Config {
	return types.Config{
		LLM: types.LLMConfig{
			Providers: []types.ProviderConfig{
				{Name: "openai", APIKey: "", Model: "gpt-4o", Enabled: false, Priority: 1, Timeout: 30},
				{Name: "anthropic", APIKey: "", Model: "claude-sonnet-4-20250514", Enabled: false, Priority: 2, Timeout: 30},
				{Name: "openrouter", APIKey: "", Model: "openai/gpt-4o", Enabled: false, Priority: 3, Timeout: 30},
				{Name: "glm", APIKey: "", Model: "glm-4-plus", Enabled: false, Priority: 4, Timeout: 30},
				{Name: "local", BaseURL: "http://localhost:8080/v1", Model: "", Enabled: false, Priority: 5, Timeout: 60},
			},
			Fallback: []string{"openai", "anthropic", "openrouter", "glm", "local"},
		},
		Model: types.ModelConfig{
			Default: "opencode/qwen3.6-plus-free",
			Coding:  "auto",
			Intent:  "gpt-4o-mini",
		},
		OpenCode: types.OpenCodeConfig{
			Binary:   "",
			Timeout:  300,
			MaxTurns: 50,
		},
		Gateway: types.GatewayConfig{
			Enabled:   false,
			Platforms: []string{},
		},
		Memory: types.MemoryConfig{
			MaxL0Items:   20,
			MaxL1Items:   50,
			AutoCompress: true,
			TokenBudget:  8000,
			SmartPrune:   true,
		},
		Web: types.WebConfig{
			Enabled: false,
			Port:    ":8080",
			Auth:    types.WebAuthConfig{Enabled: false},
		},
		TokenBudget: types.TokenBudgetConfig{
			TotalBudget:       100000,
			WarningThreshold:  0.7,
			CriticalThreshold: 0.9,
			AlertChannels:     []string{"log", "tui", "web", "gateway"},
			PerProvider:       false,
		},
	}
}

func LoadConfig(configPath string) (types.Config, error) {
	cfg := DefaultConfig()

	envVars, _ := LoadEnvFile()
	if len(envVars) > 0 {
		envCfg := EnvConfig{}.FromMap(envVars)
		envCfg.ApplyToConfig(&cfg)
	}

	path, err := resolveConfigPath(configPath)
	if err != nil {
		// Config not found - will trigger auto-setup in main.go
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}

	return cfg, nil
}

// ConfigExists checks if a config file exists at any of the standard paths.
func ConfigExists() bool {
	for _, p := range ConfigPaths() {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func SaveConfig(cfg types.Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func ConfigPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{".aigo/config.yaml"}
	}
	return []string{
		".aigo/config.yaml",
		filepath.Join(home, ".aigo", "config.yaml"),
	}
}

func resolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		if _, err := os.Stat(flagPath); err == nil {
			return flagPath, nil
		}
		return "", fmt.Errorf("config not found: %s", flagPath)
	}

	for _, p := range ConfigPaths() {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("no config file found")
}

func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".aigo/config.yaml"
	}
	return filepath.Join(home, ".aigo", "config.yaml")
}
