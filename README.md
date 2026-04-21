# Aigo — AI Coding Assistant

<p align="center">
  <strong>Lightweight, Fast, Autonomous</strong><br>
  AI coding assistant inspired by Aider + Plandex + PicoClaw
</p>

<p align="center">
  <a href="https://github.com/ahmad-ubaidillah/aigo/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/ahmad-ubaidillah/aigo">
    <img src="https://goreportcard.com/badge/github.com/ahmad-ubaidillah/aigo" alt="Go Report"/>
  </a>
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/Binary-6.3MB-brightgreen" alt="Size"/>
</p>

---

## What is Aigo?

**Aigo** is an autonomous AI coding assistant written in Go. Inspired by:
- **Aider** — Git integration, auto-commit, undo
- **Plandex** — Plan/Task system with diff sandbox
- **PicoClaw** — Ultra-lightweight, fast boot

### Core Features

| Feature | Description |
|---------|-------------|
| **Git Integration** | Auto-commit, diff review, undo like Git |
| **Plan/Task System** | Multi-step tasks with state tracking |
| **Diff Sandbox** | Queue changes, review before apply |
| **Code Indexing** | Symbol search, error mapping |
| **Project Memory** | Per-project context |
| **Action Log** | Full audit trail with undo |
| **Vision Pipeline** | Image input for multimodal |
| **Smart Routing** | Route simple queries to cheap models |
| **MCP Support** | Connect MCP servers |

---

## Quick Start

### One-Line Install (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/ahmad-ubaidillah/aigo/main/install.sh | bash
```

### Or Build from Source

```bash
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo
go build -ldflags="-s -w" -o aigo ./cmd/aigo/
upx -9 aigo
./aigo "Hello, help me write a Go function"
```

### Configuration

`~/.aigo/config.yaml`:

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  api_key: "${OPENAI_API_KEY}"
```

---

## Tools

### Git (Aider-like)
- `git_status` — Branch, staged, modified files
- `git_diff` — View changes
- `git_commit` — Commit with message
- `git_commit_auto` — Auto-commit all
- `git_undo` — Undo last N commits
- `git_log` — Commit history
- `git_branch` — Branch management

### Plan/Task (Plandex-like)
- `plan_create` — Create new plan
- `plan_add_task` — Add task to plan
- `plan_list` — List all plans
- `plan_show` — Show plan details

### Diff Sandbox
- `sandbox_add` — Queue change for review
- `sandbox_list` — List pending
- `sandbox_show` — View diff
- `sandbox_apply` — Apply change(s)
- `sandbox_reject` — Reject change(s)

### Memory
- `project_context` — Get project context
- `project_add_fact` — Add fact to memory
- `codex_index` — Index symbols
- `codex_find_symbol` — Find symbol
- `codex_map_error` — Map error to source

### Action Log
- `actionlog_list` — List actions
- `actionlog_undo` — Undo last action
- `actionlog_diff` — Get diff

### Vision
- `vision_encode` — Encode image to base64
- `vision_detect_type` — Detect MIME type

---

## Architecture

```
┌─────────────┐
│    CLI     │
└──────┬──────┘
       │
┌──────▼──────┐
│  Intent    │
│  Gate     │
└──────┬──────┘
       │
┌──────▼──────┐
│  Planning  │
│  System   │
└──────┬──────┘
       │
┌──────▼──────┐
│   Agent    │
└──────┬──────┘
       │
┌──────▼──────┐
│   Tools    │
└──────┬──────┘
       │
┌──────▼──────┐
│  Memory   │
└───────────┘
```

---

## Binary Size

| Version | Size | Method |
|---------|------|--------|
| Normal | 24MB | `go build` |
| Stripped | 16MB | `-ldflags="-s -w"` |
| Compressed | **6.3MB** | UPX |

---

## Providers

60+ AI providers supported. See [Providers](./docs/providers.md) for full list.

### Major
- OpenAI (GPT-4o, o3)
- Anthropic (Claude 3.5 Sonnet)
- Google (Gemini 2.0 Flash)
- DeepSeek (V3, R1)
- xAI (Grok-2)
- Mistral

### Chinese
- Moonshot (Kimi), GLM/Zhipu, MiniMax, Qwen, Xiaomi MiMo, Baidu, iFlytek

### Aggregators
- OpenRouter (200+ models), Together AI, Fireworks, DeepInfra

### Local
- Ollama, vLLM

---

## Documentation

- [Getting Started](./docs/getting-started.md) — Quick start
- [Installation](./docs/installation.md) — Detailed install
- [Features](./docs/features.md) — All features
- [Tools Reference](./docs/tools.md) — Complete tool reference
- [Providers](./docs/providers.md) — AI providers guide

---

## Channels

| Platform | Support |
|----------|---------|
| CLI | ✅ Default |
| Telegram | ✅ |
| Discord | ✅ |
| Slack | ✅ |
| WhatsApp | ✅ |
| WebSocket | ✅ |

---

## License

MIT

---

<p align="center">
  <strong>Built for speed 🚀</strong>
</p>