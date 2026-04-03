# Aigo V1.5 — Upgrade Guide

**Version:** 1.5.0  
**Date:** 2026-04-03  
**Theme:** Never-Die Architecture — Multi-provider LLM, auto-fallback, token budget, agent roles

---

## What's New

V1.5 introduces a resilient LLM orchestration layer that ensures Aigo never stops working due to a single provider failure. Key features:

1. **Multi-Provider LLM Router** — Configure multiple LLM providers with automatic fallback
2. **Token Budget Manager** — Track and alert on token usage across all providers
3. **Agent Roles** — Pre-configured specialist agents (Aizen, Atlas, Cody, Nova, Testa)
4. **Progressive Enhancement** — OpenCode auto-install with native tool fallback
5. **Cross-Channel Alerts** — Budget alerts to log, TUI, Web, and messaging gateways

---

## Migration: Single-Provider to Multi-Provider

### Before (V1.4 — Single Provider)

```yaml
llm:
  provider: openai
  api_key: sk-xxx
  base_url: https://api.openai.com/v1
  default_model: gpt-4o
```

### After (V1.5 — Multi-Provider)

```yaml
llm:
  providers:
    - name: openai
      api_key: sk-xxx
      model: gpt-4o
      enabled: true
      priority: 1
      timeout: 30
    - name: anthropic
      api_key: sk-ant-xxx
      model: claude-sonnet-4-20250514
      enabled: true
      priority: 2
      timeout: 30
    - name: local
      model: llama-3
      base_url: http://localhost:8080/v1
      enabled: true
      priority: 3
  fallback:
    - openai
    - anthropic
    - local
```

### Backward Compatibility

Your old single-provider config **still works**. Aigo will automatically convert it to the new format internally. However, you lose fallback benefits until you configure multiple providers.

---

## Configuring Providers

### ProviderConfig Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Provider identifier |
| `api_key` | string | Conditional | API key (not needed for `local`) |
| `base_url` | string | Conditional | Custom API endpoint |
| `model` | string | Yes | Default model name |
| `enabled` | bool | Yes | Whether to use this provider |
| `priority` | int | Yes | Lower = higher priority (1 = first) |
| `timeout` | int | No | Request timeout in seconds (default: 30) |

### Built-in Providers

| Name | Base URL | Notes |
|------|----------|-------|
| `openai` | https://api.openai.com/v1 | Supports custom base_url |
| `anthropic` | https://api.anthropic.com/v1 | |
| `openrouter` | https://openrouter.ai/api/v1 | Supports custom base_url |
| `glm` | https://open.bigmodel.cn/api/paas/v4 | Supports custom base_url |
| `local` | http://localhost:8080/v1 | No API key required |
| `custom` | (user-defined) | Requires base_url + api_key |

### Custom Provider Example

```yaml
llm:
  providers:
    - name: custom
      api_key: my-secret-key
      base_url: https://my-private-llm.example.com/v1
      model: custom-model-v2
      enabled: true
      priority: 1
```

---

## LLM Router Fallback Chain

The router tries providers in this order:

1. **Healthy providers** sorted by priority
2. **Fallback list** providers not in primary set
3. **All remaining providers** as last resort (even unhealthy)

A provider is marked **unhealthy** after 3 consecutive failures. It gets marked healthy again on the next success.

### Health Behavior

```
Provider A (priority 1): fails 3 times → marked unhealthy → skipped
Provider B (priority 2): succeeds → used for all requests
Provider A: next request succeeds → marked healthy again
```

---

## Token Budget Manager

### Configuration

```yaml
token_budget:
  warning_threshold: 0.7    # Alert at 70% usage
  critical_threshold: 0.9   # Alert at 90% usage
  alert_channels:
    - log    # Standard logger
    - tui    # TUI dashboard
    - web    # Web GUI (WebSocket)
    - gateway # Messaging platforms
```

### How It Works

- Every LLM response token count is added to the budget tracker
- When usage crosses the warning threshold, a `WARNING` alert fires
- When usage crosses the critical threshold, a `CRITICAL` alert fires
- Alerts are dispatched to all configured channels simultaneously
- Channel failures don't block other channels

---

## Agent Roles

Five pre-configured specialist agents wrap OpenCode task() with role-specific context:

| Role | Category | Purpose | Max Turns |
|------|----------|---------|-----------|
| **Aizen** | ultrabrain | CEO — decisions, coordination, final approval | 20 |
| **Atlas** | deep | Architect — system design, tech choices, patterns | 15 |
| **Cody** | deep | Developer — implementation, bugs, tests, code review | 25 |
| **Nova** | deep | PM — requirements, backlog, effort estimation | 10 |
| **Testa** | deep | QA — test planning, regression, quality metrics | 15 |

### Using Roles Programmatically

```go
import "github.com/ahmad-ubaidillah/aigo/internal/agents"

// Get a specific role
role, ok := agents.GetRole("cody")

// List all roles
allRoles := agents.ListRoles()

// Execute with role context
exec := agents.NewExecutor(opencodeClient)
result, err := exec.Execute(ctx, "cody", "fix the login bug", "session-123")
// Internally sends: "[Cody: deep]\n\nfix the login bug"
// Session ID becomes: "session-123-cody"
```

---

## New CLI Commands

| Command | Description |
|---------|-------------|
| `aigo providers` | List configured providers with health status |
| `aigo budget` | Show current token usage and per-provider breakdown |
| `aigo agents` | List available agent roles with descriptions |
| `aigo install opencode` | Manually trigger OpenCode auto-installation |

---

## Progressive Enhancement

Aigo now handles coding tasks with a fallback chain:

1. **Check** if OpenCode is available
2. **Auto-install** if missing (downloads from GitHub releases)
3. **Health check** the installation
4. **Execute** via OpenCode
5. **Fallback** to native bash tool if all else fails

Native tools (bash, file ops, web search) are always available. OpenCode is an upgrade, not a requirement.

---

## Troubleshooting

### No providers configured

```
Error: no providers configured
```

**Fix:** Run `aigo setup` to configure at least one LLM provider, or add providers to `.aigo/config.yaml`.

### All providers failing

```
Error: all providers failed (last error: connection refused)
```

**Fix:** Check your API keys and network connectivity. Run `aigo providers` to see health status.

### Token budget exceeded

```
Token budget CRITICAL: 9500/10000 (95%) used via openai
```

**Fix:** Increase your budget in config, or switch to a cheaper provider. Budget resets on restart.

### OpenCode install fails

```
OpenCode not available and install failed: download failed
```

**Fix:** Install OpenCode manually: `aigo install opencode`, or ensure native tools are available as fallback.

### Custom provider not working

**Fix:** Ensure `base_url` is set and points to an OpenAI-compatible API endpoint. The `custom` provider type requires both `base_url` and `api_key`.

---

## Architecture Changes

### New Packages

| Package | Purpose |
|---------|---------|
| `internal/llm/factory.go` | Provider factory — creates clients from config |
| `internal/llm/router.go` | LLM Router — fallback chain with health tracking |
| `internal/budget/manager.go` | Token Budget Manager — usage tracking + alerts |
| `internal/budget/alerts.go` | Alert Dispatcher — multi-channel alert routing |
| `internal/agents/roles.go` | Agent Roles — specialist agent configuration |
| `internal/handlers/coding.go` | Coding Handler — OpenCode with native fallback |

### Modified Packages

| Package | Change |
|---------|--------|
| `pkg/types/types.go` | Added `ProviderConfig`, `TokenBudgetConfig` |
| `internal/llm/client.go` | Added `Chatter`, `ExtendedLLMClient` interfaces |
| `internal/llm/openai.go` | Added `SetBaseURL()` method |
| `internal/llm/openrouter.go` | Added `SetBaseURL()` method |
| `internal/agent/loop.go` | Uses LLMRouter instead of LLMClient |
| `internal/planning/*.go` | All planners use LLMRouter |
| `internal/healing/rootcause.go` | Uses LLMRouter |

---

## Testing

All V1.5 components have comprehensive test coverage:

```bash
go test ./internal/llm/...      # Factory + Router tests
go test ./internal/budget/...   # Manager + Alert tests
go test ./internal/agents/...   # Role tests
go test ./internal/handlers/... # Coding handler tests
```

Run with race detection:

```bash
go test -race ./internal/llm/... ./internal/budget/... ./internal/agents/... ./internal/handlers/...
```
