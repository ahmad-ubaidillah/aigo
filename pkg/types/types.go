// Package types defines all shared data structures for Aigo.
package types

import (
	"time"
)

// Message represents a single message in a conversation session.
type Message struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"` // "user", "assistant", "system", "tool"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Intent categories for classification.
const (
	IntentCoding     = "coding"
	IntentWeb        = "web"
	IntentFile       = "file"
	IntentGateway    = "gateway"
	IntentMemory     = "memory"
	IntentAutomation = "automation"
	IntentGeneral    = "general"
	IntentSkill      = "skill"
	IntentResearch   = "research"
	IntentHTTPCall   = "http_call"
	IntentBrowser    = "browser"
	IntentPython     = "python"
)

// Intent represents a classified user intention.
type Intent struct {
	Category    string   `json:"category"`
	Confidence  float64  `json:"confidence"`
	Workspace   string   `json:"workspace"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// Task status constants.
const (
	TaskPending    = "pending"
	TaskInProgress = "in_progress"
	TaskDone       = "done"
	TaskFailed     = "failed"
	TaskCancelled  = "cancelled"
)

// Task priority constants.
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

// Task represents a unit of work to be executed.
type Task struct {
	ID          int64     `json:"id"`
	SessionID   string    `json:"session_id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Workspace   string    `json:"workspace"`
	Result      string    `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// Session represents a conversation session.
type Session struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Workspace  string    `json:"workspace"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	LastActive time.Time `json:"last_active"`
}

const (
	SessionIdle    = "idle"
	SessionRunning = "running"
	SessionPaused  = "paused"
)

// Memory represents a stored piece of context.
type Memory struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Tags      string    `json:"tags"` // comma-separated
	CreatedAt time.Time `json:"created_at"`
}

// ProviderConfig holds configuration for a single LLM provider.
type ProviderConfig struct {
	Name     string `yaml:"name"` // "openai", "anthropic", "openrouter", "glm", "local", "custom"
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"` // For custom providers or local endpoints
	Model    string `yaml:"model"`    // Default model for this provider
	Enabled  bool   `yaml:"enabled"`
	Priority int    `yaml:"priority"` // Lower = higher priority (1 = first in fallback chain)
	Timeout  int    `yaml:"timeout"`  // Seconds, 0 = use default (30s)
}

// TokenBudgetConfig holds token budget alerting configuration.
type TokenBudgetConfig struct {
	TotalBudget       int      `yaml:"total_budget"`       // Total token budget per session (0 = unlimited)
	WarningThreshold  float64  `yaml:"warning_threshold"`  // 0.7 = warn at 70% usage
	CriticalThreshold float64  `yaml:"critical_threshold"` // 0.9 = critical at 90% usage
	AlertChannels     []string `yaml:"alert_channels"`     // "log", "tui", "web", "gateway"
	PerProvider       bool     `yaml:"per_provider"`       // Track budget per provider separately
}

// LLMConfig holds LLM provider configuration.
// Supports both legacy single-provider and new multi-provider modes.
type LLMConfig struct {
	// NEW: Multi-provider configuration
	Providers []ProviderConfig `yaml:"providers"` // Ordered list of providers
	Fallback  []string         `yaml:"fallback"`  // Provider names in fallback order (overrides priority)

	// LEGACY: Single-provider mode (kept for backward compatibility)
	Provider     string `yaml:"provider"` // openai, glm, local, anthropic, openrouter
	APIKey       string `yaml:"api_key"`
	BaseURL      string `yaml:"base_url"` // custom endpoint override
	DefaultModel string `yaml:"default_model"`
}

// Config represents the full application configuration.
type Config struct {
	LLM         LLMConfig         `yaml:"llm"`
	Model       ModelConfig       `yaml:"model"`
	OpenCode    OpenCodeConfig    `yaml:"opencode"`
	Gateway     GatewayConfig     `yaml:"gateway"`
	Memory      MemoryConfig      `yaml:"memory"`
	Web         WebConfig         `yaml:"web"`
	TokenBudget TokenBudgetConfig `yaml:"token_budget"`
	Workspace   string            `yaml:"workspace"` // Default workspace directory
}

// ModelConfig holds model selection settings.
type ModelConfig struct {
	Default string `yaml:"default"`
	Coding  string `yaml:"coding"`
	Intent  string `yaml:"intent"`
}

// OpenCodeConfig holds OpenCode delegation settings.
type OpenCodeConfig struct {
	Binary   string `yaml:"binary"`
	Timeout  int    `yaml:"timeout"`
	MaxTurns int    `yaml:"max_turns"`
}

// GatewayConfig holds messaging platform settings.
type GatewayConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Platforms []string `yaml:"platforms"`

	// Platform-specific configs
	Telegram TelegramConfig `yaml:"telegram"`
	Discord  DiscordConfig  `yaml:"discord"`
	Slack    SlackConfig    `yaml:"slack"`
	WhatsApp WhatsAppConfig `yaml:"whatsapp"`
}

// TelegramConfig holds Telegram bot settings.
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
}

// DiscordConfig holds Discord bot settings.
type DiscordConfig struct {
	BotToken string `yaml:"bot_token"`
}

// SlackConfig holds Slack bot settings.
type SlackConfig struct {
	BotToken      string `yaml:"bot_token"`
	SigningSecret string `yaml:"signing_secret"`
}

// WhatsAppConfig holds WhatsApp business API settings.
type WhatsAppConfig struct {
	PhoneNumberID string `yaml:"phone_number_id"`
	AccessToken   string `yaml:"access_token"`
}

// MemoryConfig holds context memory settings.
type MemoryConfig struct {
	MaxL0Items   int  `yaml:"max_l0_items"`
	MaxL1Items   int  `yaml:"max_l1_items"`
	AutoCompress bool `yaml:"auto_compress"`
	TokenBudget  int  `yaml:"token_budget"`
	SmartPrune   bool `yaml:"smart_prune"`
}

// WebConfig holds web GUI settings.
type WebConfig struct {
	Enabled bool          `yaml:"enabled"`
	Port    string        `yaml:"port"`
	Auth    WebAuthConfig `yaml:"auth"`
}

// WebAuthConfig holds web authentication settings.
type WebAuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// ToolResult represents the outcome of a tool execution.
type ToolResult struct {
	Success  bool              `json:"success"`
	Output   string            `json:"output"`
	Error    string            `json:"error,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Skill represents an executable skill that extends agent capabilities.
type Skill struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Code        string    `json:"code"`    // Go code for the skill
	Command     string    `json:"command"` // Shell command
	Tags        string    `json:"tags"`
	Category    string    `json:"category"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UsageCount  int       `json:"usage_count"`
	Rating      float64   `json:"rating"`
	Enabled     bool      `json:"enabled"`
}

// SkillResult represents the output of executing a skill.
type SkillResult struct {
	Output   string            `json:"output"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Error    string            `json:"error,omitempty"`
}

// MCPConnection represents an MCP server connection.
type MCPConnection struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	ServerType    string     `json:"server_type"`
	Endpoint      string     `json:"endpoint"`
	APIKey        string     `json:"api_key"`
	Tools         []string   `json:"tools"`
	Enabled       bool       `json:"enabled"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastConnected *time.Time `json:"last_connected,omitempty"`
}

// MCPToolResult represents the result of calling an MCP tool.
type MCPToolResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// CronJob represents a scheduled task.
type CronJob struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Expression string    `json:"expression"` // cron expression (e.g., "0 9 * * *")
	Command    string    `json:"command"`
	Platform   string    `json:"platform"` // telegram, discord, slack, email
	Enabled    bool      `json:"enabled"`
	LastRun    time.Time `json:"last_run"`
	NextRun    time.Time `json:"next_run"`
	CreatedAt  time.Time `json:"created_at"`
}

// CronSchedule represents a cron schedule entry.
type CronSchedule struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	JobID       string     `json:"job_id"`
	Schedule    string     `json:"schedule"`
	Command     string     `json:"command"`
	Status      string     `json:"status"`
	Output      string     `json:"output"`
	Enabled     bool       `json:"enabled"`
	LastRun     *time.Time `json:"last_run,omitempty"`
	NextRun     *time.Time `json:"next_run,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// SelfImproveLog represents a learning log entry for self-improvement.
type SelfImproveLog struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	TurnInput  string    `json:"turn_input"`
	TurnOutput string    `json:"turn_output"`
	Outcome    string    `json:"outcome"` // success, failure, partial
	SkillName  string    `json:"skill_name"`
	Action     string    `json:"action"`
	Details    string    `json:"details"`
	Result     string    `json:"result"`
	Success    bool      `json:"success"`
	SkillGen   bool      `json:"skill_gen"`
	CreatedAt  time.Time `json:"created_at"`
}

// ResearchQuery represents a research task.
type ResearchQuery struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Sources   []string  `json:"sources"` // web, code, docs
	Result    string    `json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

// Profile represents a user profile with preferences and configuration.
type Profile struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	DefaultModel     string            `json:"default_model"`
	CodingModel      string            `json:"coding_model"`
	APIKey           string            `json:"api_key"`
	OpenCodeBinary   string            `json:"opencode_binary"`
	OpenCodeTimeout  int               `json:"opencode_timeout"`
	OpenCodeMaxTurns int               `json:"opencode_max_turns"`
	PlatformPrefs    map[string]string `json:"platform_prefs"` // key=platform, value=token/config
	Preferences      map[string]string `json:"preferences"`    // arbitrary key-value prefs
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}
