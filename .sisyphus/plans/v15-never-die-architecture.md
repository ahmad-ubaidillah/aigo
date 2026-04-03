# Aigo V1.5 — Never-Die Architecture

**Goal**: Make Aigo orchestration never stop — LLM Router with multi-provider fallback, Auto-Install OpenCode, Token Budget Manager, Progressive Enhancement, Agent Roles.

**Created**: 2026-04-03
**Based on**: Full codebase exploration (internal/llm, internal/opencode, internal/installer, internal/setup, internal/token, internal/context, internal/agent, pkg/types)
**Metis Review**: ✅ Completed — 6 risks identified, 9 tasks across 4 waves

---

## Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| **Provider-Agnostic Router** | Most providers (Z.AI, MiniMax, Alibaba, Kimi) support OpenAI-compatible API — one generic client handles all |
| **GenericOpenAIClient** | Reuse existing OpenAIClient as base for any OpenAI-compatible provider via configurable base_url |
| **Multi-Provider Config** | Replace single `LLMConfig.Provider` with `[]ProviderConfig` array — ordered fallback chain |
| **Auto-Detect OpenCode** | Check PATH first, then common paths, then download — least error-prone |
| **Centralized Budget Manager** | Single source of truth replaces scattered ContextEngine + AgentLoop token tracking |
| **Progressive Enhancement** | Native tools always available; OpenCode = upgrade, not requirement |
| **Agent Roles as Wrappers** | Wrap OpenCode task() with role context — not building agents from scratch |

## Scope

**IN**: LLM Router, Multi-Provider Config, Auto-Install OpenCode, Token Budget Manager, Cross-Channel Alerts, Progressive Enhancement Chain, Agent Roles (Aizen/Atlas/Cody/Nova/Testa), Setup Wizard v2
**OUT**: Memory Graph, LSP Integration, AST-Grep, Hash-Anchored Edits, Skill Marketplace, MCP Configuration, Kanban Board — these are V2 features

## Key Patterns from Codebase

- **LLMClient interface** (`internal/llm/client.go`): `Complete()`, `Chat()`, `CompleteWithSystem()` — all 5 providers implement this
- **Tool interface** (`internal/tools/tool.go`): `Name()`, `Description()`, `Schema()`, `Execute()` — native tools pattern
- **Config loading** (`internal/cli/config.go`): env file → YAML → merge → `types.Config`
- **OpenCode client** (`internal/opencode/client.go`): CLI wrapper with `Run()`, `HealthCheck()`, `CheckVersion()`
- **Installer stub** (`internal/installer/installer.go`): `Download()`, `InstallBinary()`, `GetOS()`, `GetArch()` — scaffold exists
- **Setup Wizard** (`internal/setup/wizard.go`): CLI/Web mode selection + API key — needs provider selection upgrade

---

## Task Wave 1: Foundation (No Dependencies)

### Task 1: Multi-Provider Config Structure

**File**: `pkg/types/types.go`
**Goal**: Replace single-provider config with multi-provider array supporting any OpenAI-compatible provider.

**Changes**:
1. Add `ProviderConfig` struct to `pkg/types/types.go`:
```go
type ProviderConfig struct {
    Name     string `yaml:"name"`      // "openai", "anthropic", "openrouter", "glm", "local", "custom"
    APIKey   string `yaml:"api_key"`
    BaseURL  string `yaml:"base_url"`  // For custom providers
    Model    string `yaml:"model"`     // Default model for this provider
    Enabled  bool   `yaml:"enabled"`
    Priority int    `yaml:"priority"`  // Lower = higher priority (1 = first)
    Timeout  int    `yaml:"timeout"`   // Seconds, 0 = default
}
```

2. Update `LLMConfig` struct:
```go
type LLMConfig struct {
    Providers []ProviderConfig `yaml:"providers"`  // NEW: ordered list
    Fallback  []string         `yaml:"fallback"`   // Provider names in fallback order
    // DEPRECATED but kept for backward compat:
    Provider     string `yaml:"provider"`
    APIKey       string `yaml:"api_key"`
    BaseURL      string `yaml:"base_url"`
    DefaultModel string `yaml:"default_model"`
}
```

3. Add `TokenBudgetConfig` struct:
```go
type TokenBudgetConfig struct {
    WarningThreshold  float64 `yaml:"warning_threshold"`  // 0.7 = 70%
    CriticalThreshold float64 `yaml:"critical_threshold"` // 0.9 = 90%
    AlertChannels     []string `yaml:"alert_channels"`    // "log", "tui", "web", "gateway"
    PerProvider       bool    `yaml:"per_provider"`       // Track budget per provider
}
```

4. Add to `Config`:
```go
type Config struct {
    // ... existing fields ...
    TokenBudget TokenBudgetConfig `yaml:"token_budget"`
}
```

5. Update `DefaultConfig()` in `internal/cli/config.go` with sensible defaults:
```go
LLM: types.LLMConfig{
    Providers: []types.ProviderConfig{
        {Name: "openai", APIKey: "", Model: "gpt-4o", Enabled: false, Priority: 1},
        {Name: "anthropic", APIKey: "", Model: "claude-sonnet-4-20250514", Enabled: false, Priority: 2},
        {Name: "openrouter", APIKey: "", Model: "openai/gpt-4o", Enabled: false, Priority: 3},
        {Name: "glm", APIKey: "", Model: "glm-4-plus", Enabled: false, Priority: 4},
        {Name: "local", BaseURL: "http://localhost:8080/v1", Model: "", Enabled: false, Priority: 5},
    },
    Fallback: []string{"openai", "anthropic", "openrouter", "glm", "local"},
},
TokenBudget: types.TokenBudgetConfig{
    WarningThreshold:  0.7,
    CriticalThreshold: 0.9,
    AlertChannels:     []string{"log", "tui", "web", "gateway"},
    PerProvider:       false,
},
```

6. Update `EnvConfig` to support `AIGO_PROVIDERS` env var (JSON array) and `AIGO_FALLBACK` (comma-separated).

**Acceptance Criteria**:
- [ ] `ProviderConfig` struct compiles and YAML marshals correctly
- [ ] `DefaultConfig()` returns 5 pre-configured providers with correct priorities
- [ ] Existing single-provider config still works (backward compat)
- [ ] `TokenBudgetConfig` defaults to 70%/90% thresholds
- [ ] Test: YAML round-trip preserves all fields
- [ ] Test: Old config format (single provider) still loads correctly

**QA Scenarios**:
- Happy path: Load config with 3 providers enabled → verify sorted by priority
- Edge case: Empty providers list → fallback to deprecated single-provider fields
- Edge case: Invalid priority values → default to insertion order
- Edge case: Duplicate provider names → error on load

---

### Task 2: LLM Provider Factory

**Files**: `internal/llm/factory.go` (new), `internal/llm/client.go` (extend)
**Goal**: Factory function that creates any LLM client from `ProviderConfig`.

**New file** `internal/llm/factory.go`:
```go
package llm

import "github.com/ahmad-ubaidillah/aigo/pkg/types"

// NewProvider creates an LLMClient from ProviderConfig.
// Returns nil if provider is not configured (no API key, not enabled).
func NewProvider(cfg types.ProviderConfig) (LLMClient, error) {
    switch cfg.Name {
    case "openai":
        if cfg.APIKey == "" {
            return nil, nil // Not configured
        }
        client := NewOpenAIClient(cfg.APIKey, cfg.Model)
        if cfg.BaseURL != "" {
            client.SetBaseURL(cfg.BaseURL) // Need to add this method
        }
        return client, nil
    case "anthropic":
        if cfg.APIKey == "" {
            return nil, nil
        }
        return NewAnthropicClient(cfg.APIKey, cfg.Model), nil
    case "openrouter":
        if cfg.APIKey == "" {
            return nil, nil
        }
        client := NewOpenRouterClient(cfg.APIKey, cfg.Model)
        if cfg.BaseURL != "" {
            client.SetBaseURL(cfg.BaseURL) // Need to add this method
        }
        return client, nil
    case "glm":
        if cfg.APIKey == "" {
            return nil, nil
        }
        if cfg.BaseURL != "" {
            return NewGLMClientWithBaseURL(cfg.APIKey, cfg.Model, cfg.BaseURL), nil
        }
        return NewGLMClient(cfg.APIKey, cfg.Model), nil
    case "local":
        return NewLocalClient(cfg.Model, cfg.BaseURL), nil
    case "custom":
        if cfg.BaseURL == "" {
            return nil, fmt.Errorf("custom provider requires base_url")
        }
        // Custom = OpenAI-compatible with user-defined base_url
        client := NewOpenAIClient(cfg.APIKey, cfg.Model)
        client.SetBaseURL(cfg.BaseURL)
        return client, nil
    default:
        return nil, fmt.Errorf("unknown provider: %s", cfg.Name)
    }
}

// NewProviders creates all enabled providers from config, sorted by priority.
func NewProviders(cfg types.LLMConfig) ([]NamedProvider, error) {
    // Sort by priority
    sorted := make([]types.ProviderConfig, len(cfg.Providers))
    copy(sorted, cfg.Providers)
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Priority < sorted[j].Priority
    })

    var result []NamedProvider
    for _, p := range sorted {
        if !p.Enabled {
            continue
        }
        client, err := NewProvider(p)
        if err != nil {
            return nil, fmt.Errorf("create provider %s: %w", p.Name, err)
        }
        if client == nil {
            continue // Not configured (no API key)
        }
        result = append(result, NamedProvider{
            Name:   p.Name,
            Client: client,
            Model:  p.Model,
            Timeout: time.Duration(p.Timeout) * time.Second,
        })
    }
    return result, nil
}

type NamedProvider struct {
    Name    string
    Client  LLMClient
    Model   string
    Timeout time.Duration
}
```

**Extend existing clients** to support `SetBaseURL()`:
- Add `func (c *OpenAIClient) SetBaseURL(url string)` to `openai.go`
- Add `func (c *OpenRouterClient) SetBaseURL(url string)` to `openrouter.go`

**Acceptance Criteria**:
- [ ] `NewProvider()` returns correct client type for each provider name
- [ ] `NewProvider()` returns nil (no error) for unconfigured providers
- [ ] `NewProvider()` returns error for unknown provider names
- [ ] `NewProvider()` returns error for custom provider without base_url
- [ ] `NewProviders()` returns providers sorted by priority
- [ ] `NewProviders()` skips disabled and unconfigured providers
- [ ] `SetBaseURL()` works on OpenAIClient and OpenRouterClient

**QA Scenarios**:
- Happy path: Config with OpenAI + Anthropic enabled → returns 2 providers in priority order
- Edge case: All providers disabled → returns empty slice
- Edge case: Custom provider without base_url → returns error
- Edge case: Unknown provider name → returns error
- Edge case: Provider with empty API key → returns nil (skipped gracefully)

---

### Task 3: LLM Router with Fallback Chain

**Files**: `internal/llm/router.go` (new), `internal/llm/router_test.go` (new)
**Goal**: Router that tries providers in order, falls back on failure, tracks health.

**New file** `internal/llm/router.go`:
```go
package llm

import (
    "context"
    "fmt"
    "sync"
    "time"
)

// LLMRouter routes requests to the best available provider with fallback.
type LLMRouter struct {
    providers []NamedProvider
    fallback  []string // Provider names in fallback order
    health    map[string]*ProviderHealth
    mu        sync.RWMutex
}

type ProviderHealth struct {
    Name         string
    LastCheck    time.Time
    IsHealthy    bool
    ConsecutiveFailures int
    LastError    string
}

func NewLLMRouter(providers []NamedProvider, fallback []string) *LLMRouter {
    r := &LLMRouter{
        providers: providers,
        fallback:  fallback,
        health:    make(map[string]*ProviderHealth),
    }
    for _, p := range providers {
        r.health[p.Name] = &ProviderHealth{Name: p.Name, IsHealthy: true}
    }
    return r
}

// Chat sends a chat request, trying providers in order with fallback.
func (r *LLMRouter) Chat(ctx context.Context, messages []Message, opts ChatOptions) (*ChatResponse, error) {
    r.mu.RLock()
    ordered := r.orderedProviders()
    r.mu.RUnlock()

    var lastErr error
    for _, p := range ordered {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }

        resp, err := r.chatWithProvider(ctx, p, messages, opts)
        if err == nil {
            r.markSuccess(p.Name)
            return resp, nil
        }

        r.markFailure(p.Name, err)
        lastErr = err
    }

    return nil, fmt.Errorf("all providers failed (last error: %w)", lastErr)
}

func (r *LLMRouter) orderedProviders() []NamedProvider {
    // First try healthy providers in priority order
    // Then try fallback providers that aren't in primary list
    // Skip providers with too many consecutive failures
    var result []NamedProvider
    seen := make(map[string]bool)

    // Primary: healthy providers in priority order
    for _, p := range r.providers {
        if seen[p.Name] {
            continue
        }
        seen[p.Name] = true
        h := r.health[p.Name]
        if h.IsHealthy && h.ConsecutiveFailures < 3 {
            result = append(result, p)
        }
    }

    // Fallback: providers from fallback list not already included
    for _, name := range r.fallback {
        if seen[name] {
            continue
        }
        for _, p := range r.providers {
            if p.Name == name {
                seen[name] = true
                result = append(result, p)
                break
            }
        }
    }

    // Last resort: all remaining providers (even unhealthy)
    for _, p := range r.providers {
        if !seen[p.Name] {
            result = append(result, p)
        }
    }

    return result
}

func (r *LLMRouter) chatWithProvider(ctx context.Context, p NamedProvider, messages []Message, opts ChatOptions) (*ChatResponse, error) {
    timeout := p.Timeout
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    // Try ChatWithOptions first (extended interface), fall back to basic Chat
    if chatter, ok := p.Client.(Chatter); ok {
        return chatter.ChatWithOptions(ctx, messages, opts)
    }

    // Basic Chat — need to adapt messages for the basic interface
    // The basic Chat interface takes []Message (llm.Message) not []llm.Message
    // We need to check if the client also implements the basic LLMClient interface
    if basic, ok := p.Client.(LLMClient); ok {
        // Convert Message to the basic interface format
        basicMsgs := make([]Message, len(messages))
        for i, m := range messages {
            basicMsgs[i] = Message{Role: m.Role, Content: m.Content}
        }
        resp, err := basic.Chat(ctx, basicMsgs)
        if err != nil {
            return nil, err
        }
        return &ChatResponse{
            Content: resp.Content,
            Usage: TokenUsage(resp.Usage),
        }, nil
    }

    return nil, fmt.Errorf("provider %s does not support chat", p.Name)
}

func (r *LLMRouter) markSuccess(name string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if h, ok := r.health[name]; ok {
        h.ConsecutiveFailures = 0
        h.IsHealthy = true
        h.LastCheck = time.Now()
        h.LastError = ""
    }
}

func (r *LLMRouter) markFailure(name string, err error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    if h, ok := r.health[name]; ok {
        h.ConsecutiveFailures++
        h.LastError = err.Error()
        h.LastCheck = time.Now()
        if h.ConsecutiveFailures >= 3 {
            h.IsHealthy = false
        }
    }
}

// Health returns health status for all providers.
func (r *LLMRouter) Health() map[string]*ProviderHealth {
    r.mu.RLock()
    defer r.mu.RUnlock()
    result := make(map[string]*ProviderHealth)
    for k, v := range r.health {
        result[k] = v
    }
    return result
}

// CheckHealth performs a lightweight health check on all providers.
func (r *LLMRouter) CheckHealth(ctx context.Context) map[string]bool {
    results := make(map[string]bool)
    for _, p := range r.providers {
        ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
        _, err := r.chatWithProvider(ctx, p, []Message{{Role: "user", Content: "ping"}}, ChatOptions{})
        cancel()
        results[p.Name] = err == nil
    }
    return results
}
```

**Acceptance Criteria**:
- [ ] Router tries providers in priority order
- [ ] Router falls back to next provider on failure
- [ ] Router marks provider unhealthy after 3 consecutive failures
- [ ] Router skips unhealthy providers in normal routing
- [ ] Router tries ALL providers as last resort (even unhealthy)
- [ ] Router respects per-provider timeout
- [ ] `Health()` returns current health status
- [ ] `CheckHealth()` performs lightweight health checks

**QA Scenarios**:
- Happy path: 3 providers, first succeeds → returns response from first
- Fallback: First fails, second succeeds → returns response from second
- All fail: All 3 providers fail → returns error with last failure
- Health: Provider fails 3 times → marked unhealthy → skipped on next request
- Recovery: Unhealthy provider succeeds → marked healthy again
- Timeout: Provider takes too long → timeout → fallback to next
- Last resort: All providers unhealthy → still tries all of them

---

### Task 4: Token Budget Manager

**Files**: `internal/budget/manager.go` (new), `internal/budget/manager_test.go` (new)
**Goal**: Centralized token budget tracking with threshold alerting.

**New package** `internal/budget/`:
```go
package budget

import (
    "fmt"
    "sync"
    "time"

    "github.comahmad-ubaidillah/aigo/pkg/types"
)

// Manager tracks token usage across all components with threshold alerting.
type Manager struct {
    mu            sync.RWMutex
    totalBudget   int
    used          int
    perProvider   map[string]int
    thresholds    Thresholds
    alertHandlers []AlertHandler
    history       []UsageSnapshot
}

type Thresholds struct {
    Warning  float64 // 0.7 = 70%
    Critical float64 // 0.9 = 90%
}

type AlertLevel string
const (
    AlertWarning  AlertLevel = "warning"
    AlertCritical AlertLevel = "critical"
)

type AlertEvent struct {
    Level     AlertLevel
    Usage     int
    Budget    int
    Percent   float64
    Provider  string
    Timestamp time.Time
    Message   string
}

type AlertHandler func(event AlertEvent)

type UsageSnapshot struct {
    Used        int
    Budget      int
    Percent     float64
    Provider    string
    Timestamp   time.Time
}

func NewManager(budget int, thresholds Thresholds) *Manager {
    return &Manager{
        totalBudget: budget,
        perProvider: make(map[string]int),
        thresholds:  thresholds,
        history:     make([]UsageSnapshot, 0, 100),
    }
}

// Add usage from a specific provider.
func (m *Manager) Add(used int, provider string) []AlertEvent {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.used += used
    m.perProvider[provider] += used

    percent := float64(m.used) / float64(m.totalBudget)
    snapshot := UsageSnapshot{
        Used: m.used, Budget: m.totalBudget,
        Percent: percent, Provider: provider,
        Timestamp: time.Now(),
    }
    m.history = append(m.history, snapshot)

    // Check thresholds and fire alerts
    var alerts []AlertEvent
    if percent >= m.thresholds.Critical {
        event := AlertEvent{
            Level: AlertCritical, Usage: m.used, Budget: m.totalBudget,
            Percent: percent, Provider: provider, Timestamp: time.Now(),
            Message: fmt.Sprintf("Token budget CRITICAL: %d/%d (%.0f%%) used via %s",
                m.used, m.totalBudget, percent*100, provider),
        }
        alerts = append(alerts, event)
        for _, h := range m.alertHandlers {
            h(event)
        }
    } else if percent >= m.thresholds.Warning {
        event := AlertEvent{
            Level: AlertWarning, Usage: m.used, Budget: m.totalBudget,
            Percent: percent, Provider: provider, Timestamp: time.Now(),
            Message: fmt.Sprintf("Token budget WARNING: %d/%d (%.0f%%) used via %s",
                m.used, m.totalBudget, percent*100, provider),
        }
        alerts = append(alerts, event)
        for _, h := range m.alertHandlers {
            h(event)
        }
    }

    return alerts
}

// OnAlert registers an alert handler.
func (m *Manager) OnAlert(h AlertHandler) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.alertHandlers = append(m.alertHandlers, h)
}

// Usage returns current usage stats.
func (m *Manager) Usage() (used, budget int, percent float64) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.used, m.totalBudget, float64(m.used) / float64(m.totalBudget)
}

// PerProvider returns usage per provider.
func (m *Manager) PerProvider() map[string]int {
    m.mu.RLock()
    defer m.mu.RUnlock()
    result := make(map[string]int)
    for k, v := range m.perProvider {
        result[k] = v
    }
    return result
}

// History returns recent usage snapshots.
func (m *Manager) History() []UsageSnapshot {
    m.mu.RLock()
    defer m.mu.RUnlock()
    result := make([]UsageSnapshot, len(m.history))
    copy(result, m.history)
    return result
}

// Reset resets usage counters.
func (m *Manager) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.used = 0
    m.perProvider = make(map[string]int)
}
```

**Acceptance Criteria**:
- [ ] `Add()` increments total and per-provider usage
- [ ] `Add()` returns AlertEvent when warning threshold crossed
- [ ] `Add()` returns AlertEvent when critical threshold crossed
- [ ] `Add()` fires alert handlers for each threshold crossing
- [ ] `Add()` only fires warning once (not on every subsequent Add)
- [ ] `Usage()` returns current used/budget/percent
- [ ] `PerProvider()` returns usage breakdown by provider
- [ ] `History()` returns recent snapshots
- [ ] `Reset()` clears all counters
- [ ] Thread-safe under concurrent access

**QA Scenarios**:
- Happy path: Add 7000 tokens to 10000 budget → warning alert fired
- Critical: Add 9000 tokens to 10000 budget → critical alert fired
- Per-provider: Add via "openai" then "anthropic" → perProvider shows breakdown
- Threshold once: Cross warning at 70%, add more → no duplicate warning
- Concurrent: 10 goroutines adding simultaneously → no race conditions
- Reset: Reset → usage returns to 0

---

## Task Wave 2: Integration (Depends on Wave 1)

### Task 5: Auto-Install OpenCode

**Files**: `internal/installer/installer.go` (rewrite), `internal/installer/opencode.go` (new)
**Goal**: Detect, download, and install OpenCode binary with health verification.

**Changes to** `internal/installer/installer.go`:

1. Implement `CheckOpenCode()`:
```go
func (i *Installer) CheckOpenCode() (bool, string, error) {
    // Check PATH
    if path, err := exec.LookPath("opencode"); err == nil {
        return true, path, nil
    }
    // Check common paths
    common := []string{
        "/usr/local/bin/opencode",
        "/usr/bin/opencode",
        filepath.Join(os.Getenv("HOME"), ".local", "bin", "opencode"),
    }
    for _, p := range common {
        if _, err := os.Stat(p); err == nil {
            return true, p, nil
        }
    }
    return false, "", nil
}
```

2. Implement `InstallOpenCode()`:
```go
func (i *Installer) InstallOpenCode(ctx context.Context, installPath string) error {
    os := runtime.GOOS
    arch := runtime.GOARCH

    // Build download URL from GitHub releases
    // Format: https://github.com/opencode-ai/opencode/releases/latest/download/opencode-{os}-{arch}
    url := fmt.Sprintf("https://github.com/opencode-ai/opencode/releases/latest/download/opencode-%s-%s", os, arch)

    if i.verbose {
        fmt.Printf("Downloading OpenCode from %s...\n", url)
    }

    // Download to temp file
    tmpPath := filepath.Join(os.TempDir(), "opencode-"+os+"-"+arch)
    if err := Download(url, tmpPath); err != nil {
        return fmt.Errorf("download OpenCode: %w", err)
    }

    // Make executable
    if err := os.Chmod(tmpPath, 0755); err != nil {
        return fmt.Errorf("chmod: %w", err)
    }

    // Ensure target directory exists
    if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
        return fmt.Errorf("create install dir: %w", err)
    }

    // Move to install path
    if err := os.Rename(tmpPath, installPath); err != nil {
        // Fallback: copy if rename fails (cross-device)
        if err := copyFile(tmpPath, installPath); err != nil {
            return fmt.Errorf("install: %w", err)
        }
    }

    if i.verbose {
        fmt.Printf("OpenCode installed to %s\n", installPath)
    }

    return nil
}
```

3. Implement `InstallAll()`:
```go
func (i *Installer) InstallAll(ctx context.Context) *InstallResult {
    result := &InstallResult{Success: true}

    // Check OpenCode
    exists, path, err := i.CheckOpenCode()
    if err != nil {
        return &InstallResult{Success: false, Message: fmt.Sprintf("check OpenCode: %v", err)}
    }

    if exists {
        result.OpenCodeVer = "detected"
        result.Message = fmt.Sprintf("OpenCode already installed at %s", path)
        return result
    }

    // Auto-detect install path
    installPath := i.detectInstallPath()

    if err := i.InstallOpenCode(ctx, installPath); err != nil {
        return &InstallResult{Success: false, Message: fmt.Sprintf("install OpenCode: %v", err)}
    }

    result.OpenCodeVer = "installed"
    result.Message = fmt.Sprintf("OpenCode installed to %s", installPath)
    return result
}

func (i *Installer) detectInstallPath() string {
    // Prefer ~/.local/bin on Linux, /usr/local/bin on macOS
    home, _ := os.UserHomeDir()
    if runtime.GOOS == "linux" && home != "" {
        return filepath.Join(home, ".local", "bin", "opencode")
    }
    return "/usr/local/bin/opencode"
}
```

**Acceptance Criteria**:
- [ ] `CheckOpenCode()` detects binary in PATH
- [ ] `CheckOpenCode()` detects binary in common paths
- [ ] `CheckOpenCode()` returns (false, "", nil) when not found
- [ ] `InstallOpenCode()` downloads correct OS/arch binary
- [ ] `InstallOpenCode()` makes binary executable
- [ ] `InstallOpenCode()` verifies installation with version check
- [ ] `InstallOpenCode()` handles cross-device rename (fallback to copy)
- [ ] `InstallAll()` skips install if already present
- [ ] `detectInstallPath()` returns appropriate path per OS

**QA Scenarios**:
- Happy path: OpenCode not installed → download → install → verify
- Already installed: OpenCode in PATH → skip install, return path
- Download fails: Network error → return descriptive error
- Permission denied: Cannot write to /usr/local/bin → fallback to ~/.local/bin
- Cross-device: Rename fails → copy file instead
- Invalid OS/arch: Unsupported platform → return error

---

### Task 6: Setup Wizard v2 — Multi-Provider Configuration

**Files**: `internal/setup/wizard.go` (rewrite)
**Goal**: Interactive wizard that configures multiple LLM providers with provider selection.

**Changes**:
```go
type SetupWizard struct {
    mode     string
    complete bool
    cfg      *types.Config
}

func (w *SetupWizard) Run() error {
    // Step 1: Choose mode
    w.chooseMode()

    // Step 2: Configure LLM providers
    if err := w.configureProviders(); err != nil {
        return err
    }

    // Step 3: Auto-install OpenCode
    if err := w.installOpenCode(); err != nil {
        fmt.Printf("⚠ OpenCode not installed: %v\n", err)
        fmt.Println("   You can install it later with: aigo install opencode")
    }

    // Step 4: Configure token budget
    w.configureTokenBudget()

    // Step 5: Save config
    if err := w.saveConfig(); err != nil {
        return err
    }

    w.complete = true
    return nil
}

func (w *SetupWizard) configureProviders() error {
    fmt.Println("\n🤖 LLM Provider Configuration")
    fmt.Println("Configure one or more providers for automatic fallback.")
    fmt.Println()

    providers := []struct {
        name    string
        model   string
        baseURL string
    }{
        {"openai", "gpt-4o", "https://api.openai.com/v1"},
        {"anthropic", "claude-sonnet-4-20250514", "https://api.anthropic.com/v1"},
        {"openrouter", "openai/gpt-4o", "https://openrouter.ai/api/v1"},
        {"glm", "glm-4-plus", "https://open.bigmodel.cn/api/paas/v4"},
        {"local", "", "http://localhost:8080/v1"},
    }

    var configured []types.ProviderConfig
    for i, p := range providers {
        fmt.Printf("%d. %s (%s)\n", i+1, p.name, p.model)
    }
    fmt.Printf("%d. Custom provider\n", len(providers)+1)
    fmt.Printf("%d. Skip for now\n", len(providers)+2)
    fmt.Println()

    // Let user configure one or more providers
    for {
        fmt.Print("Select provider to configure (number, or 0 to finish): ")
        var choice int
        fmt.Scanln(&choice)
        if choice == 0 {
            break
        }
        if choice < 1 || choice > len(providers)+1 {
            fmt.Println("Invalid choice")
            continue
        }

        if choice == len(providers)+1 {
            // Custom provider
            cfg := w.configureCustomProvider()
            configured = append(configured, cfg)
            continue
        }

        p := providers[choice-1]
        cfg := w.configureProvider(p.name, p.model, p.baseURL)
        if cfg.Name != "" {
            configured = append(configured, cfg)
        }
    }

    w.cfg.LLM.Providers = configured
    // Build fallback order from configured providers
    for _, p := range configured {
        w.cfg.LLM.Fallback = append(w.cfg.LLM.Fallback, p.Name)
    }

    return nil
}
```

**Acceptance Criteria**:
- [ ] Wizard presents provider selection menu
- [ ] User can configure multiple providers
- [ ] User can configure custom provider with base_url + api_key
- [ ] User can skip provider configuration
- [ ] Auto-install OpenCode attempted (non-blocking)
- [ ] Token budget configuration offered
- [ ] Config saved to `.aigo/config.yaml`
- [ ] Wizard shows summary of configured providers

**QA Scenarios**:
- Happy path: Configure OpenAI + Anthropic → config saved with 2 providers
- Custom: Configure custom provider with base_url → config saved correctly
- Skip: Skip all providers → config saved with empty providers (local-only mode)
- Install fail: OpenCode install fails → warning shown, wizard continues
- Save fail: Cannot write config file → error returned, wizard fails

---

### Task 7: Progressive Enhancement — Coding Handler with Fallback

**Files**: `internal/agent/router.go` (modify), `internal/handlers/coding.go` (new)
**Goal**: Coding handler that tries OpenCode first, falls back to native tools.

**New file** `internal/handlers/coding.go`:
```go
package handlers

import (
    "context"
    "fmt"

    "github.com/ahmad-ubaidillah/aigo/internal/opencode"
    "github.com/ahmad-ubaidillah/aigo/internal/tools"
    "github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// CodingHandler delegates coding tasks to OpenCode with native fallback.
type CodingHandler struct {
    ocClient  *opencode.Client
    registry  *tools.ToolRegistry
    installer interface {
        CheckOpenCode() (bool, string, error)
        InstallOpenCode(ctx context.Context, path string) error
    }
}

func (h *CodingHandler) CanHandle(intent string) bool {
    return intent == types.IntentCoding
}

func (h *CodingHandler) Execute(ctx context.Context, task *types.Task, workspace string) (*types.ToolResult, error) {
    // Step 1: Check OpenCode availability
    available, path, err := h.installer.CheckOpenCode()
    if err != nil {
        return h.nativeFallback(ctx, task, "OpenCode check failed: "+err.Error())
    }

    // Step 2: If not available, try auto-install
    if !available {
        installErr := h.installer.InstallOpenCode(ctx, path)
        if installErr != nil {
            return h.nativeFallback(ctx, task, "OpenCode not available and install failed: "+installErr.Error())
        }
        // Re-initialize client with new path
        h.ocClient, _ = opencode.NewClient(path, 300, workspace)
    }

    // Step 3: Health check
    if h.ocClient != nil {
        health, _ := h.ocClient.HealthCheck()
        if !health.Success {
            return h.nativeFallback(ctx, task, "OpenCode health check failed: "+health.Error)
        }
    }

    // Step 4: Execute via OpenCode
    if h.ocClient != nil {
        return h.ocClient.Run(ctx, task.Description, task.SessionID)
    }

    // Step 5: Native fallback
    return h.nativeFallback(ctx, task, "OpenCode not available")
}

func (h *CodingHandler) nativeFallback(ctx context.Context, task *types.Task, reason string) (*types.ToolResult, error) {
    // Use bash tool as fallback
    bashTool := h.registry.Get("bash")
    if bashTool == nil {
        return &types.ToolResult{
            Success: false,
            Error:   fmt.Sprintf("coding task failed: %s. No fallback available.", reason),
        }, nil
    }

    result, err := bashTool.Execute(ctx, map[string]any{"command": task.Description})
    if err != nil {
        return &types.ToolResult{
            Success: false,
            Error:   fmt.Sprintf("coding task failed: %s. Native fallback also failed: %v", reason, err),
        }, nil
    }

    result.Output = fmt.Sprintf("[Native Fallback] %s\n\n%s", reason, result.Output)
    return result, nil
}
```

**Modify** `internal/agent/router.go`:
- Extract `codingHandler` from inline struct to use new `handlers.CodingHandler`
- Pass `ToolRegistry` and `Installer` to CodingHandler

**Acceptance Criteria**:
- [ ] CodingHandler checks OpenCode availability before delegation
- [ ] CodingHandler attempts auto-install if not available
- [ ] CodingHandler runs health check before execution
- [ ] CodingHandler falls back to bash tool if OpenCode unavailable
- [ ] Native fallback output is prefixed with "[Native Fallback]"
- [ ] Installer interface allows mocking in tests

**QA Scenarios**:
- Happy path: OpenCode available + healthy → execute via OpenCode
- Auto-install: OpenCode not available → auto-install → execute
- Health fail: OpenCode unhealthy → native fallback
- No OpenCode: Not available + install fails → native fallback
- No bash: Native fallback but bash tool missing → error returned

---

## Task Wave 3: Agent Roles + Alerts (Depends on Wave 2)

### Task 8: Agent Roles Configuration Layer

**Files**: `internal/agents/roles.go` (new), `internal/agents/executor.go` (new)
**Goal**: Role-based agent system wrapping OpenCode task() with role context.

**New file** `internal/agents/roles.go`:
```go
package agents

import (
    "context"
    "fmt"

    "github.com/ahmad-ubaidillah/aigo/internal/opencode"
    "github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// Role defines an agent role with category, system prompt, and skills.
type Role struct {
    Name        string
    Category    string // Maps to OpenCode category
    SystemPrompt string
    Skills      []string
    MaxTurns    int
}

var Roles = map[string]Role{
    "aizen": {
        Name:        "Aizen",
        Category:    "ultrabrain",
        SystemPrompt: "You are Aizen, the CEO agent. You make decisions, coordinate tasks, decompose complex work, and give final approval before user delivery.",
        Skills:      []string{"coordination", "planning"},
        MaxTurns:    20,
    },
    "atlas": {
        Name:        "Atlas",
        Category:    "deep",
        SystemPrompt: "You are Atlas, the Architect. You analyze system design, review architecture, recommend technology choices, and identify design patterns.",
        Skills:      []string{"architecture", "design-patterns"},
        MaxTurns:    15,
    },
    "cody": {
        Name:        "Cody",
        Category:    "deep",
        SystemPrompt: "You are Cody, the Developer. You implement code, fix bugs, write tests, review code quality, and ensure production-ready output.",
        Skills:      []string{"code-review", "testing", "documentation"},
        MaxTurns:    25,
    },
    "nova": {
        Name:        "Nova",
        Category:    "deep",
        SystemPrompt: "You are Nova, the Project Manager. You analyze requirements, manage backlogs, estimate effort, track progress, and create user stories.",
        Skills:      []string{"requirements", "planning"},
        MaxTurns:    10,
    },
    "testa": {
        Name:        "Testa",
        Category:    "deep",
        SystemPrompt: "You are Testa, the QA Engineer. You plan tests, identify bugs, perform regression testing, report quality metrics, and verify fixes.",
        Skills:      []string{"testing", "quality-assurance"},
        MaxTurns:    15,
    },
}

// Executor executes a task with a specific role.
type Executor struct {
    client *opencode.Client
}

func NewExecutor(client *opencode.Client) *Executor {
    return &Executor{client: client}
}

// Execute runs a task with the given role's context.
func (e *Executor) Execute(ctx context.Context, roleName, taskDesc, sessionID string) (*types.ToolResult, error) {
    role, ok := Roles[roleName]
    if !ok {
        return nil, fmt.Errorf("unknown role: %s", roleName)
    }

    // Build prompt with role context
    prompt := fmt.Sprintf("[%s: %s]\n\n%s", role.Name, role.Category, taskDesc)

    // Execute via OpenCode with role-specific session
    roleSessionID := fmt.Sprintf("%s-%s", sessionID, roleName)
    return e.client.Run(ctx, prompt, roleSessionID)
}

// ExecuteParallel runs multiple roles in parallel.
func (e *Executor) ExecuteParallel(ctx context.Context, roles []string, taskDesc, sessionID string) (map[string]*types.ToolResult, error) {
    results := make(map[string]*types.ToolResult)
    errs := make(map[string]error)

    // TODO: Implement parallel execution with goroutines + WaitGroup
    // Each role gets its own session ID
    // Results aggregated and returned as map

    return results, nil
}
```

**Acceptance Criteria**:
- [ ] 5 roles defined: aizen, atlas, cody, nova, testa
- [ ] Each role has unique Name, Category, SystemPrompt, Skills, MaxTurns
- [ ] `Execute()` prefixes task with role context
- [ ] `Execute()` creates role-specific session ID
- [ ] `Execute()` returns error for unknown role
- [ ] `ExecuteParallel()` runs multiple roles concurrently

**QA Scenarios**:
- Happy path: Execute "cody" role → prompt prefixed with "[Cody: deep]"
- Unknown role: Execute "unknown" → returns error
- Parallel: Execute cody + testa → both run, results returned

---

### Task 9: Cross-Channel Token Budget Alerts

**Files**: `internal/budget/alerts.go` (new), `internal/budget/alerts_test.go` (new)
**Goal**: Wire Token Budget Manager alerts to log, TUI, Web, and Gateway channels.

**New file** `internal/budget/alerts.go`:
```go
package budget

import (
    "fmt"
    "log"
    "sync"
)

// AlertDispatcher routes budget alerts to multiple channels.
type AlertDispatcher struct {
    channels []AlertChannel
    mu       sync.RWMutex
}

type AlertChannel interface {
    Name() string
    Send(event AlertEvent) error
}

// LogChannel sends alerts to standard logger.
type LogChannel struct{}

func (LogChannel) Name() string { return "log" }
func (LogChannel) Send(event AlertEvent) error {
    switch event.Level {
    case AlertCritical:
        log.Printf("🔴 CRITICAL: %s", event.Message)
    case AlertWarning:
        log.Printf("🟡 WARNING: %s", event.Message)
    }
    return nil
}

// TUIChannel sends alerts to TUI (via callback).
type TUIChannel struct {
    onUpdate func(event AlertEvent)
}

func (TUIChannel) Name() string { return "tui" }
func (c TUIChannel) Send(event AlertEvent) error {
    if c.onUpdate != nil {
        c.onUpdate(event)
    }
    return nil
}

// WebChannel sends alerts to Web GUI (via WebSocket or SSE).
type WebChannel struct {
    broadcast func(event AlertEvent)
}

func (WebChannel) Name() string { return "web" }
func (c WebChannel) Send(event AlertEvent) error {
    if c.broadcast != nil {
        c.broadcast(event)
    }
    return nil
}

// GatewayChannel sends alerts to messaging platforms.
type GatewayChannel struct {
    sendToGateway func(platform, message string) error
    platforms     []string
}

func (c GatewayChannel) Name() string { return "gateway" }
func (c GatewayChannel) Send(event AlertEvent) error {
    msg := fmt.Sprintf("[%s] %s", event.Level, event.Message)
    for _, platform := range c.platforms {
        if err := c.sendToGateway(platform, msg); err != nil {
            return fmt.Errorf("send to %s: %w", platform, err)
        }
    }
    return nil
}

func NewDispatcher() *AlertDispatcher {
    return &AlertDispatcher{channels: make([]AlertChannel, 0)}
}

func (d *AlertDispatcher) Register(ch AlertChannel) {
    d.mu.Lock()
    defer d.mu.Unlock()
    d.channels = append(d.channels, ch)
}

func (d *AlertDispatcher) Dispatch(event AlertEvent) []error {
    d.mu.RLock()
    defer d.mu.RUnlock()

    var errs []error
    for _, ch := range d.channels {
        if err := ch.Send(event); err != nil {
            errs = append(errs, fmt.Errorf("channel %s: %w", ch.Name(), err))
        }
    }
    return errs
}
```

**Wire into Manager** (Task 4 extension):
```go
// In manager.go, add:
func (m *Manager) WithDispatcher(d *AlertDispatcher) *Manager {
    m.OnAlert(func(event AlertEvent) {
        d.Dispatch(event)
    })
    return m
}
```

**Acceptance Criteria**:
- [ ] LogChannel prints formatted alert to logger
- [ ] TUIChannel calls onUpdate callback
- [ ] WebChannel calls broadcast function
- [ ] GatewayChannel sends to all configured platforms
- [ ] AlertDispatcher routes to all registered channels
- [ ] Channel failures don't block other channels
- [ ] Manager.WithDispatcher() wires alerts to dispatcher

**QA Scenarios**:
- Happy path: Alert fired → all 4 channels receive it
- Partial fail: WebChannel fails → LogChannel still receives
- No channels: Dispatcher with no channels → no errors
- Multiple alerts: 3 alerts fired → all dispatched in order

---

## Task Wave 4: Wiring + Polish (Depends on Wave 3)

### Task 10: Wire Everything Together — Main Integration

**Files**: `cmd/aigo/main.go` (modify), `cmd/aigo/commands.go` (modify)
**Goal**: Wire LLM Router, Token Budget Manager, Agent Roles into main application.

**Changes to** `cmd/aigo/main.go`:
1. Load config with multi-provider support
2. Create LLM Router from configured providers
3. Create Token Budget Manager with alert dispatcher
4. Wire into AgentLoop, Planning agents, CodingHandler
5. Display provider health on startup

**Changes to** `cmd/aigo/commands.go`:
1. Add `aigo providers` command — list configured providers and health
2. Add `aigo budget` command — show token usage and alerts
3. Add `aigo agents` command — list available agent roles
4. Add `aigo install opencode` command — manual OpenCode installation

**Acceptance Criteria**:
- [ ] `aigo run "task"` uses LLM Router for orchestration
- [ ] `aigo providers` shows all configured providers with health status
- [ ] `aigo budget` shows current token usage and per-provider breakdown
- [ ] `aigo agents` lists Aizen, Atlas, Cody, Nova, Testa with descriptions
- [ ] `aigo install opencode` runs OpenCode auto-install
- [ ] Startup shows which providers are active and which are skipped
- [ ] Token budget alerts appear in configured channels

**QA Scenarios**:
- Happy path: 2 providers configured → both active, routing works
- Single provider: Only OpenAI configured → works, no fallback
- No providers: No API keys → local-only mode, warning shown
- Budget display: After some usage → shows correct percentages

---

### Task 11: Update All LLM Consumers to Use Router

**Files**: `internal/agent/loop.go`, `internal/planning/prometheus.go`, `internal/planning/metis.go`, `internal/planning/momus.go`, `internal/planning/llm_planner.go`, `internal/healing/rootcause.go`
**Goal**: Replace direct `LLMClient` usage with `LLMRouter` across all consumers.

**Pattern for each file**:
```go
// BEFORE:
type AgentLoop struct {
    llmClient llm.LLMClient
    // ...
}

// AFTER:
type AgentLoop struct {
    llmRouter *llm.LLMRouter
    // ...
}

// Usage changes from:
resp, err := l.llmClient.Chat(ctx, messages)
// To:
resp, err := l.llmRouter.Chat(ctx, messages, llm.ChatOptions{})
```

**Files to modify**:
1. `internal/agent/loop.go` — `selectToolWithLLM()` uses router instead of client
2. `internal/planning/prometheus.go` — `GeneratePlanWithLLM()`, `InterviewWithLLM()`
3. `internal/planning/metis.go` — `AnalyzeWithLLM()`
4. `internal/planning/momus.go` — `ReviewWithLLM()`
5. `internal/planning/llm_planner.go` — `Plan()`, `RefinePlan()`
6. `internal/healing/rootcause.go` — `AnalyzeRootCause()`

**Acceptance Criteria**:
- [ ] All 6 files compile with LLMRouter instead of LLMClient
- [ ] No direct LLMClient references remain in consumer code
- [ ] All existing tests pass (may need mock updates)
- [ ] Fallback behavior preserved (router handles it now)

**QA Scenarios**:
- Each consumer: Router with 1 provider → behaves same as before
- Each consumer: Router with 3 providers → uses first available
- Each consumer: All providers fail → returns error with last failure

---

### Task 12: Tests + Documentation

**Files**: All new test files, `docs/v15-upgrade.md` (new)
**Goal**: Comprehensive test coverage and upgrade documentation.

**Test files to create**:
1. `internal/llm/factory_test.go` — Provider factory tests
2. `internal/llm/router_test.go` — Router fallback tests
3. `internal/budget/manager_test.go` — Budget manager tests
4. `internal/budget/alerts_test.go` — Alert dispatcher tests
5. `internal/handlers/coding_test.go` — Coding handler fallback tests
6. `internal/agents/roles_test.go` — Agent roles tests
7. `internal/installer/installer_test.go` — Extended installer tests

**Documentation** `docs/v15-upgrade.md`:
- Migration guide from single-provider to multi-provider config
- How to configure custom providers
- How token budget alerting works
- How agent roles work
- Troubleshooting guide

**Acceptance Criteria**:
- [ ] All new test files have >80% coverage
- [ ] All existing tests still pass
- [ ] `go test ./...` passes with 0 failures
- [ ] `go test -race ./...` passes with 0 races
- [ ] Upgrade documentation covers all new features
- [ ] Migration guide has before/after config examples

---

## Final Verification Wave

> **MANDATORY**: The implementer must run these verifications before marking work complete.

### Build Verification
```bash
go build ./...           # Must compile cleanly
go vet ./...             # Must pass
go test -race ./...      # Must pass with 0 races
```

### Functional Verification
```bash
# 1. Test multi-provider config
aigo providers           # Shows configured providers

# 2. Test LLM Router fallback
aigo run "hello world"   # Uses first available provider

# 3. Test token budget
aigo budget             # Shows current usage

# 4. Test agent roles
aigo agents             # Lists all 5 roles

# 5. Test OpenCode install
aigo install opencode   # Auto-installs if missing
```

### User Confirmation
Before marking work complete, the implementer MUST ask the user:
> "V1.5 implementation complete. All providers configured, fallback working, budget alerts active, agent roles ready. Please verify and confirm with 'okay' if everything works as expected."

The work is NOT complete until the user confirms with "okay".

---

## Dependency Graph

```
Wave 1 (Foundation)          Wave 2 (Integration)         Wave 3 (Roles + Alerts)     Wave 4 (Wiring)
┌──────────────────┐         ┌──────────────────┐         ┌──────────────────┐        ┌──────────────────┐
│ Task 1: Config   │────────▶│ Task 5: Install  │         │ Task 8: Roles    │───────▶│ Task 10: Main    │
│ Task 2: Factory  │────────▶│ Task 6: Wizard   │         │ Task 9: Alerts   │───────▶│ Task 11: Wire    │
│ Task 3: Router   │────────▶│ Task 7: Coding   │         │                  │        │ Task 12: Tests   │
│ Task 4: Budget   │────────▶│                  │         │                  │        │                  │
└──────────────────┘         └──────────────────┘         └──────────────────┘        └──────────────────┘
```

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| R1: Scope creep | Lock to 12 tasks only. Defer Memory Graph, LSP, AST-Grep to V2 |
| R2: Fallback latency | Per-provider timeout (default 30s), max 3 retries per provider |
| R3: Budget inconsistency | Single Manager instance, injected everywhere, event-driven updates |
| R4: Install security | SHA-256 verify downloaded binary, sandboxed install, rollback on failure |
| R5: Backward compat | Deprecated fields kept, old config format still loads |
| R6: Partial provider failure | Router tries ALL providers as last resort, even unhealthy ones |
