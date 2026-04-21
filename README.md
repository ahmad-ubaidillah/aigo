# Aigo — Autonomous AI Agent Platform

<p align="center">
  <strong>Execute with Zen</strong><br>
  A minimal, fast, token-efficient autonomous AI agent in Go
</p>

<p align="center">
  <a href="https://github.com/ahmad-ubaidillah/aigo/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/ahmad-ubaidillah/aigo">
    <img src="https://goreportcard.com/badge/github.com/ahmad-ubaidillah/aigo" alt="Go Report"/>
  </a>
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version"/>
  <img src="https://img.shields.io/badge/Binary-6.3MB-blue" alt="Binary Size"/>
</p>

---

## What is Aigo?

**Aigo** (pronounced "eye-go") is a local-first, autonomous AI agent platform written in Go. It combines intelligent orchestration with powerful coding capabilities to help developers automate complex tasks with minimal intervention.

### Core Philosophy

> **"Execute with Zen"** — Aigo handles the complexity so you can focus on what matters.

Aigo is built on three pillars:
- **Intelligence** — Smart intent classification and context understanding
- **Efficiency** — Token-optimized processing (60-90% reduction via distillation)
- **Autonomy** — Plans → Executes → Reports → Learns

---

## Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo
go build -o aigo ./cmd/aigo/

# Or install directly
go install github.com/ahmad-ubaidillah/aigo/cmd/aigo@latest

# Compress binary (optional, reduces from 16MB to 6MB)
upx -9 aigo
```

### First Run

```bash
# Interactive chat mode
./aigo chat

# Single task mode
./aigo "Fix the authentication bug in login.go"

# Start gateway server
./aigo start
```

### Configuration

Create `~/.aigo/config.yaml`:

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  api_key: "${OPENAI_API_KEY}"

agent:
  max_iterations: 90
  max_tokens: 4096

memory:
  enabled: true
  storage_path: ~/.aigo/memory
  use_fts5: true
  pyramid_enabled: true

channels:
  telegram:
    enabled: false
    token: ""
  discord:
    enabled: false
    token: ""
  slack:
    enabled: false
    app_token: ""
    bot_token: ""
  websocket:
    enabled: false
    port: 8765
  whatsapp:
    enabled: false
    account_sid: ""
    auth_token: ""
    from_number: ""

webui:
  enabled: true
  port: 9090
```

---

## Features

| Feature | Description |
|---------|-------------|
| **Prometheus Planner** | Plans before executing with intent classification |
| **Metis Gap Analyzer** | Identifies ambiguities and missing information |
| **Momus Reviewer** | Reviews plans for completeness and verifiability |
| **6-Tier Memory** | L0-L2 context levels with SQLite persistence |
| **Vector Search** | SimHash embeddings with sqlite-vec |
| **48 Lifecycle Hooks** | Fine-grained event triggers |
| **Self-Healing** | Error pattern learning and automatic retry |
| **Multi-Agent** | Sisyphus/Hephaestus/Oracle/Explore agents |
| **Gateway** | Telegram, Discord, Slack, WhatsApp, WebSocket |
| **Skill Hub** | 500+ skills from Smithery, Anthropic, GitHub |
| **Token Distillation** | 60-90% token reduction pipeline |
| **MCP Server** | Expose tools via Model Context Protocol |
| **Prompt Caching** | Support for o1, Claude caching |
| **SSE Streaming** | Real-time token streaming in WebUI |

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      USER INTERFACE                     │
│   CLI  ·  TUI (Bubble Tea)  ·  Web UI  ·  Gateway      │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                    INTENT GATE                          │
│         TF-IDF Classification → Handler Routing        │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                   PLANNING LAYER                        │
│   Prometheus → Metis (Gap) → Momus (Review) → Resolver │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                     AGENT LOOP                           │
│      ReAct: Think → Tool → Observe → Loop until done    │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                    TOOL SYSTEM                          │
│   100+ tools: read, write, edit, grep, bash, LSP...    │
└──────────────────────────┬──────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────┐
│                    MEMORY ENGINE                        │
│   SQLite + Vector Store + Session + Facts + Hooks      │
└─────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
aigo/
├── cmd/aigo/              # CLI entry point
├── internal/
│   ├── agent/            # Core agent loop with ReAct
│   ├── planning/         # Prometheus/Metis/Momus + LLM
│   ├── memory/           # SQLite + FTS5 + Vector
│   ├── hooks/            # 48 lifecycle hooks
│   ├── autonomy/         # Self-healing + learning
│   ├── subagent/         # Sisyphus/Hephaestus/Oracle
│   ├── gateway/          # Platform adapters
│   ├── skillhub/         # Skill marketplace
│   ├── providers/        # LLM providers
│   ├── tools/            # Tool registry
│   ├── mcp/              # MCP client + server
│   ├── channels/         # Telegram/Discord/Slack/WhatsApp
│   └── webui/            # Web dashboard
├── config.yaml           # Default configuration
└── README.md
```

---

## CLI Commands

```bash
# Chat modes
./aigo chat                    # Interactive chat
./aigo "your message"          # One-shot query
./aigo start                   # Start gateway server

# Version
./aigo version                 # Show version

# Skill marketplace
./aigo skills search <query>   # Search skills
./aigo skills install <id>     # Install skill
./aigo skills list             # List installed
./aigo skills sync             # Sync online index
```

---

## Integrations

### LLM Providers

| Provider | Model Support |
|----------|---------------|
| OpenAI | GPT-4o, GPT-4o Mini, GPT-4 Turbo, o1 |
| Anthropic | Claude 3.5 Sonnet, Opus |
| OpenRouter | 100+ models |
| Local | Ollama, LM Studio |

### Gateways

| Platform | Status | Notes |
|----------|--------|-------|
| Telegram | ✅ | Bot API |
| Discord | ✅ | DiscordGO |
| Slack | ✅ | Slack API |
| WhatsApp | ✅ | Twilio |
| WebSocket | ✅ | Custom |

### MCP Server

Start MCP server on port 3100:

```bash
./aigo start
# MCP available at http://127.0.0.1:3100/mcp
```

---

## Performance

| Metric | Value |
|--------|-------|
| Binary Size | **6.3MB** (compressed) |
| Startup Time | <100ms |
| Memory Usage | ~20-30MB |
| Dependencies | Zero (pure Go) |

---

## Requirements

- **Go 1.21+**
- **SQLite** (included, pure Go via modernc.org)
- **LLM API Key** (optional for local models)

---

## License

MIT — See [LICENSE](LICENSE) for details.

---

## Acknowledgments

Inspired by and combines best features from:
- [OpenCode](https://github.com/opencode-ai/opencode)
- [Oh-My-OpenAgent](https://github.com/oh-my-openagent/oh-my-openagent)
- [Omni](https://github.com/omni/omni)
- [Hermes](https://github.com/hermes/hermes)
- [mem0](https://github.com/mem0ai/mem0)

---

<p align="center">
  <strong>Built with ❤️ by developers, for developers</strong><br>
  <a href="https://github.com/ahmad-ubaidillah/aigo">GitHub</a> •
  <a href="https://github.com/ahmad-ubaidillah/aigo/issues">Issues</a>
</p>