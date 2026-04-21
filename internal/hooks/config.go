package hooks

import (
	"os"

	"gopkg.in/yaml.v3"
)

type HookConfig struct {
	Version  string            `yaml:"version"`
	Enabled map[string]bool  `yaml:"enabled"`
	Order   []string          `yaml:"order"`
	Rules   []HookRule        `yaml:"rules,omitempty"`
}

type HookRule struct {
	Name      string   `yaml:"name"`
	Condition string   `yaml:"condition,omitempty"`
	Priority  int      `yaml:"priority"`
	Enabled  bool     `yaml:"enabled"`
}

func LoadHookConfig(path string) (*HookConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config HookConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveHookConfig(path string, config *HookConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func NewDefaultHookConfig() *HookConfig {
	config := &HookConfig{
		Version:  "1.0",
		Enabled:  make(map[string]bool),
		Order:    make([]string, 0),
	}

	for _, hook := range HookTypes {
		config.Enabled[hook] = true
		config.Order = append(config.Order, hook)
	}

	return config
}

func (c *HookConfig) IsEnabled(hook string) bool {
	enabled, ok := c.Enabled[hook]
	if !ok {
		return true
	}
	return enabled
}

func (c *HookConfig) SetEnabled(hook string, enabled bool) {
	c.Enabled[hook] = enabled
}

func (c *HookConfig) GetOrderedHooks() []string {
	ordered := make([]string, 0, len(c.Order))

	for _, hook := range c.Order {
		if c.IsEnabled(hook) {
			ordered = append(ordered, hook)
		}
	}

	for _, hook := range HookTypes {
		found := false
		for _, o := range ordered {
			if o == hook {
				found = true
				break
			}
		}
		if !found && c.IsEnabled(hook) {
			ordered = append(ordered, hook)
		}
	}

	return ordered
}

func (c *HookConfig) EnableAll() {
	for _, hook := range HookTypes {
		c.Enabled[hook] = true
	}
}

func (c *HookConfig) DisableAll() {
	for _, hook := range HookTypes {
		c.Enabled[hook] = false
	}
}