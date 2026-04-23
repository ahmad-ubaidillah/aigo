// Package config handles Aigo configuration.
// Config precedence: CLI flags > env vars > config file > defaults
package config

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration for Aigo.
type Config struct {
	Provider  ProviderConfig  `yaml:"provider"`
	Agent     AgentConfig     `yaml:"agent"`
	Memory    MemoryConfig    `yaml:"memory"`
	Channels  ChannelsConfig  `yaml:"channels"`
	Security  SecurityConfig  `yaml:"security"`
	WebUI     WebUIConfig     `yaml:"webui"`
	Autonomy  AutonomyConfig  `yaml:"autonomy"`
}

type ProviderConfig struct {
	Default    string            `yaml:"default"`     // e.g. "openai", "anthropic"
	BaseURL    string            `yaml:"base_url"`
	APIKey     string            `yaml:"api_key"`
	Model      string            `yaml:"model"`
	Failover   []string          `yaml:"failover"`    // fallback providers
	Providers  map[string]ProviderEntry `yaml:"providers"`
}

type ProviderEntry struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

type AgentConfig struct {
	MaxIterations   int    `yaml:"max_iterations"`
	MaxTokens       int    `yaml:"max_tokens"`
	SystemPrompt    string `yaml:"system_prompt"`
	LoopDetection   bool   `yaml:"loop_detection"`
	LoopMaxRepeats  int    `yaml:"loop_max_repeats"`
	SandboxMode     bool   `yaml:"sandbox_mode"`
}

type MemoryConfig struct {
	Enabled        bool   `yaml:"enabled"`
	StoragePath    string `yaml:"storage_path"`
	UseFTS5        bool   `yaml:"use_fts5"`         // Use SQLite FTS5 instead of flat files
	VectorDim      int    `yaml:"vector_dim"`
	PyramidEnabled bool   `yaml:"pyramid_enabled"` // Enable 5-tier pyramidal memory
}

type ChannelsConfig struct {
	Telegram  TelegramConfig  `yaml:"telegram"`
	Discord   DiscordConfig   `yaml:"discord"`
	Slack     SlackConfig     `yaml:"slack"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	WhatsApp  WhatsAppConfig  `yaml:"whatsapp"`
}

type TelegramConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`
}

type DiscordConfig struct {
	Enabled bool   `yaml:"enabled"`
	Token   string `yaml:"token"`
}

type SlackConfig struct {
	Enabled   bool   `yaml:"enabled"`
	AppToken string `yaml:"app_token"` // xapp-...
	BotToken string `yaml:"bot_token"` // xoxb-...
}

type WebSocketConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Port      int    `yaml:"port"`
	AuthToken string `yaml:"auth_token"`
}

type WhatsAppConfig struct {
	Enabled    bool   `yaml:"enabled"`
	AccountSid string `yaml:"account_sid"`
	AuthToken  string `yaml:"auth_token"`
	FromNumber string `yaml:"from_number"`
}

type SecurityConfig struct {
	HardBaseline bool `yaml:"hard_baseline"`
	SandboxMode  bool `yaml:"sandbox_mode"`
}

type WebUIConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

type AutonomyConfig struct {
	Enabled           bool     `yaml:"enabled"`
	AwakeMinMinutes   int      `yaml:"awake_min_minutes"`
	AwakeMaxMinutes   int      `yaml:"awake_max_minutes"`
	SleepStart        int      `yaml:"sleep_start"`
	SleepEnd          int      `yaml:"sleep_end"`
	Interests         []string `yaml:"interests"`
	EnableNews        bool     `yaml:"enable_news"`
	EnableReflection  bool     `yaml:"enable_reflection"`
	EnableSpontaneous bool     `yaml:"enable_spontaneous"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	return Config{
		Provider: ProviderConfig{
			Default: "openai",
			Model:   "gpt-4o-mini",
		},
		Agent: AgentConfig{
			MaxIterations:  15,
			MaxTokens:      4096,
			LoopDetection:  true,
			LoopMaxRepeats: 3,
		},
		Memory: MemoryConfig{
			Enabled:     true,
			StoragePath: filepath.Join(home, ".aigo", "memory"),
			VectorDim:   384,
		},
		Channels: ChannelsConfig{
			WebSocket: WebSocketConfig{
				Enabled: true,
				Port:    8765,
			},
		},
		Security: SecurityConfig{
			HardBaseline: true,
		},
		WebUI: WebUIConfig{
			Enabled: true,
			Port:    8080,
		},
		Autonomy: AutonomyConfig{
			Enabled:           false,
			AwakeMinMinutes:   10,
			AwakeMaxMinutes:   60,
			SleepStart:        1,
			SleepEnd:          7,
			Interests:         []string{"teknologi", "AI", "programming"},
			EnableNews:        true,
			EnableReflection:  true,
			EnableSpontaneous: true,
		},
	}
}

// Load reads config from file, then applies env var overrides.
func Load(path string) (Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return cfg, err
		}
		if err == nil {
			// Normalize legacy YAML format (llm.* -> provider.*) for backward compat
			normalized := normalizeLegacyConfig(data)
			if err := yaml.Unmarshal(normalized, &cfg); err != nil {
				return cfg, err
			}
		}
	}

	// Env var overrides
	if v := os.Getenv("AIGO_PROVIDER"); v != "" {
		cfg.Provider.Default = v
	}
	if v := os.Getenv("AIGO_MODEL"); v != "" {
		cfg.Provider.Model = v
	}
	if v := os.Getenv("AIGO_API_KEY"); v != "" {
		cfg.Provider.APIKey = v
	}
	if v := os.Getenv("AIGO_BASE_URL"); v != "" {
		cfg.Provider.BaseURL = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" && cfg.Provider.APIKey == "" {
		cfg.Provider.APIKey = v
	}
	// Also support OPENCODE_API_KEY and OPENCODE_ZEN_API_KEY for OpenCode provider
	if v := os.Getenv("OPENCODE_API_KEY"); v != "" && cfg.Provider.APIKey == "" {
		cfg.Provider.APIKey = v
	}
	if v := os.Getenv("OPENCODE_ZEN_API_KEY"); v != "" && cfg.Provider.APIKey == "" {
		cfg.Provider.APIKey = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		if cfg.Provider.Providers == nil {
			cfg.Provider.Providers = make(map[string]ProviderEntry)
		}
		cfg.Provider.Providers["anthropic"] = ProviderEntry{APIKey: v}
	}

	// Expand ~ and env vars in all paths
	cfg.Memory.StoragePath = ExpandPath(cfg.Memory.StoragePath)

	return cfg, nil
}

// ExpandPath expands ~ to the user's home directory and resolves environment variables.
func ExpandPath(path string) string {
	if path == "" {
		return path
	}
	// Expand $HOME and other env vars
	path = os.ExpandEnv(path)
	// Expand ~
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	return path
}

// normalizeLegacyConfig converts old YAML format to new format for backward compat.
// Old: llm.provider, llm.api_key, llm.default_model, llm.base_url
// New: provider.default, provider.api_key, provider.model, provider.base_url
func normalizeLegacyConfig(data []byte) []byte {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return data
	}

	// Check if this is the legacy "llm" format
	llmRaw, hasLLM := raw["llm"]
	if !hasLLM {
		return data // Already new format
	}

	llmMap, ok := llmRaw.(map[string]interface{})
	if !ok {
		return data
	}

	// Directly apply legacy fields to the raw map so yaml.Unmarshal can read them
	// into the proper Config struct
	provider := make(map[string]interface{})

	if v, ok := llmMap["provider"]; ok {
		if s, ok := v.(string); ok {
			provider["default"] = s
		}
	}
	if v, ok := llmMap["api_key"]; ok {
		if s, ok := v.(string); ok && s != "" {
			provider["api_key"] = s
		}
	}
	if v, ok := llmMap["default_model"]; ok {
		if s, ok := v.(string); ok {
			provider["model"] = s
		}
	}
	if v, ok := llmMap["base_url"]; ok {
		if s, ok := v.(string); ok && s != "" {
			provider["base_url"] = s
		}
	}

	// Only modify if we found legacy fields
	if len(provider) > 0 {
		raw["provider"] = provider
	}

	out, _ := yaml.Marshal(raw)
	return out
}

// GetAPIKey returns the API key for the active provider.
func (c *Config) GetAPIKey() string {
	if c.Provider.APIKey != "" {
		return c.Provider.APIKey
	}
	if entry, ok := c.Provider.Providers[c.Provider.Default]; ok {
		return entry.APIKey
	}
	return ""
}

// GetBaseURL returns the base URL for the active provider.
func (c *Config) GetBaseURL() string {
	if c.Provider.BaseURL != "" {
		return c.Provider.BaseURL
	}
	if entry, ok := c.Provider.Providers[c.Provider.Default]; ok {
		if entry.BaseURL != "" {
			return entry.BaseURL
		}
	}
	// Defaults per provider
	switch strings.ToLower(c.Provider.Default) {
	case "anthropic":
		return "https://api.anthropic.com"
	case "ollama":
		return "http://localhost:11434"
	default:
		return "https://api.openai.com/v1"
	}
}

// ConfigPath returns the default config file path.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".aigo", "config.yaml")
}
