package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PermissionConfig struct {
	DefaultAction string           `yaml:"default_action"`
	Rules         []PermissionRule `yaml:"rules"`
}

type PermissionRule struct {
	Tool    string `yaml:"tool"`
	Pattern string `yaml:"pattern"`
	Action  string `yaml:"action"`
}

func DefaultPermissionConfig() PermissionConfig {
	return PermissionConfig{
		DefaultAction: "allow",
		Rules:         []PermissionRule{},
	}
}

func LoadPermissionConfig() (PermissionConfig, error) {
	home, _ := os.UserHomeDir()
	if home == "" {
		return DefaultPermissionConfig(), nil
	}
	path := filepath.Join(home, ".aigo", "permissions.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultPermissionConfig(), nil
		}
		return DefaultPermissionConfig(), fmt.Errorf("read permissions: %w", err)
	}
	var cfg PermissionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultPermissionConfig(), fmt.Errorf("parse permissions: %w", err)
	}
	return cfg, nil
}

func SavePermissionConfig(cfg PermissionConfig) error {
	home, _ := os.UserHomeDir()
	if home == "" {
		return fmt.Errorf("home directory not found")
	}
	dir := filepath.Join(home, ".aigo")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	path := filepath.Join(dir, "permissions.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal permissions: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func AddPermissionRule(tool, pattern, action string) error {
	cfg, err := LoadPermissionConfig()
	if err != nil {
		return err
	}
	cfg.Rules = append(cfg.Rules, PermissionRule{
		Tool:    tool,
		Pattern: pattern,
		Action:  action,
	})
	return SavePermissionConfig(cfg)
}

func RemovePermissionRule(tool string) error {
	cfg, err := LoadPermissionConfig()
	if err != nil {
		return err
	}
	var filtered []PermissionRule
	for _, r := range cfg.Rules {
		if r.Tool != tool {
			filtered = append(filtered, r)
		}
	}
	cfg.Rules = filtered
	return SavePermissionConfig(cfg)
}

func ListPermissionRules() string {
	cfg, err := LoadPermissionConfig()
	if err != nil {
		return "Error loading permissions: " + err.Error()
	}
	if len(cfg.Rules) == 0 {
		return "No permission rules configured. Default: " + cfg.DefaultAction
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Default: %s\n\n", cfg.DefaultAction))
	for i, r := range cfg.Rules {
		b.WriteString(fmt.Sprintf("%d. tool=%s pattern=%s action=%s\n", i+1, r.Tool, r.Pattern, r.Action))
	}
	return b.String()
}
