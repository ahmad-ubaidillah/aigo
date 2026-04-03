# Aigo — AI Agent Platform

<p align="center">
  <img src="docs/logo.png" alt="Aigo" width="200"/>
</p>

<p align="center">
  <strong>Execute with Zen</strong><br>
  A minimal, fast, token-efficient autonomous AI agent platform in Go
</p>

<p align="center">
  <a href="https://github.com/ahmad-ubaidillah/aigo/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"/>
  </a>
  <a href="https://github.com/ahmad-ubaidillah/aigo/actions">
    <img src="https://github.com/ahmad-ubaidillah/aigo/workflows/CI/badge.svg" alt="CI"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/ahmad-ubaidillah/aigo">
    <img src="https://goreportcard.com/badge/github.com/ahmad-ubaidillah/aigo" alt="Go Report"/>
  </a>
</p>

---

## Table of Contents

- [What is Aigo?](#what-is-aigo)
- [Why Aigo?](#why-aigo)
- [Who is Aigo for?](#who-is-aigo-for)
- [Methodology](#methodology)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Usage](#usage)
- [Features](#features)
- [Architecture](#architecture)
- [Inspirations](#inspirations)
- [Differences from Similar Apps](#differences-from-similar-apps)
- [Configuration](#configuration)
- [Development](#development)
- [License](#license)

---

## What is Aigo?

**Aigo** (pronounced "eye-go") is a local-first, autonomous AI agent platform written in Go. It combines intelligent orchestration with powerful coding capabilities to help developers automate complex tasks with minimal intervention.

Aigo is designed to be your **AI coding partner** that:
- Understands your project context
- Plans before executing
- Works autonomously until completion
- Learns from past experiences
- Respects your privacy (all data stays local)

### Core Philosophy

> **"Execute with Zen"** — Aigo handles the complexity so you can focus on what matters.

Aigo is built on three pillars:
1. **Intelligence** — Smart intent classification and context understanding
2. **Efficiency** — Token-optimized processing for long sessions
3. **Autonomy** — Minimal intervention, maximum output

---

## Why Aigo?

### Problems Aigo Solves

| Problem | Traditional Approach | With Aigo |
|---------|-------------------|----------|
| **Context Switching** | Manually explain project context every time | Aigo remembers your project |
| **Token Waste** | Pay for repetitive context | Smart distillation reduces 60%+ tokens |
| **Micro-management** | Guide AI step-by-step | Aigo plans and executes autonomously |
| **Error Repetition** | Same mistakes repeatedly | Aigo learns from failures |
| **Privacy Concerns** | Send data to external APIs | Everything runs locally |

### Key Benefits

- **🚀 Fast & Lightweight** — Single Go binary, no dependencies to manage
- **💰 Token Efficient** — 60-90% token reduction via distillation pipeline
- **🔒 Privacy First** — All data stays on your machine
- **🧠 Self-Learning** — Learns from your patterns and preferences
- **⚡ Autonomous** — Plans → Executes → Reports → Learns
- **🔌 Extensible** — Skills, hooks, and MCP integration

---

## Who is Aigo for?

Aigo is designed for:

### 1. **Software Developers**
- Automate repetitive coding tasks
- Get help with bug fixes and refactoring
- Navigate unfamiliar codebases quickly

### 2. **DevOps Engineers**
- Automate deployment scripts
- Monitor and maintain infrastructure
- Create runbooks and documentation

### 3. **Technical Leads**
- Delegate research tasks
- Maintain coding standards
- Onboard team members to projects

### 4. **Solo Entrepreneurs**
- Build MVPs faster
- Handle technical tasks without a team
- Focus on business logic, not boilerplate

### 5. **Students & Learners**
- Learn from AI explanations
- Get context-aware code examples
- Explore new technologies

---

## Methodology

Aigo's approach is based on proven AI agent patterns:

### 1. **Plan Before Execute**
```
User Task → Prometheus Planner → Execution Plan → Sisyphus Executor → Result
```

Before any action, Aigo uses the **Prometheus** planner to:
- Understand the true intent
- Break down complex tasks
- Identify dependencies
- Estimate effort

### 2. **Tiered Context Memory**

Aigo uses a three-tier context system for maximum efficiency:

| Tier | Name | Purpose | Token Cost |
|------|------|---------|------------|
| **L0** | Abstract | High-level summary | ~50 tokens |
| **L1** | Overview | Key decisions and state | ~500 tokens |
| **L2** | Detail | Full conversation history | ~5000+ tokens |

This ensures the LLM always has the right amount of context.

### 3. **Signal Distillation**

Inspired by Omni, Aigo filters noise from tool outputs:

```
Raw Output → Classifier → Scorer → Composer → Clean Output
     │           │           │           │
     ▼           ▼           ▼           ▼
  Preserve   Classify    Boost hot   Filter noise
             content     files       Skip redundant
```

### 4. **Self-Healing**

Aigo can detect and recover from errors:
- Analyze stack traces automatically
- Suggest fixes based on patterns
- Retry with learned improvements

### 5. **Multi-Agent Collaboration**

For complex tasks, Aigo can spawn specialized agents:

| Agent | Role | Category | Focus |
|-------|------|----------|-------|
| **Aigo** | CEO | ultrabrain | Decision making, coordination |
| **Atlas** | Architect | deep | System design, architecture |
| **Cody** | Developer | deep | Implementation, coding |
| **Nova** | PM | deep | Requirements, planning |
| **Testa** | QA | deep | Testing, verification |

---

## How It Works

### The Aigo Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER INTERACTION                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐   │
│   │   CLI   │     │   TUI   │     │  Web    │     │ Gateway │   │
│   │ Terminal│     │ Bubble  │     │  GUI     │     │ Telegram│   │
│   └────┬────┘     └────┬────┘     └────┬────┘     └────┬────┘   │
│        │                │                │                │        │
│        └────────────────┴────────────────┴────────────────┘        │
│                                  │                                   │
│                         ┌────────▼────────┐                        │
│                         │   IntentGate    │                        │
│                         │  (Classification)│                        │
│                         └────────┬────────┘                        │
│                                  │                                   │
│                    ┌─────────────┼─────────────┐                   │
│                    │             │             │                    │
│              ┌─────▼─────┐ ┌─────▼─────┐ ┌─────▼─────┐           │
│              │  Native   │ │ OpenCode  │ │ Planning  │           │
│              │  Handler  │ │  Handler  │ │  Layer    │           │
│              ├───────────┤ ├───────────┤ ├───────────┤           │
│              │ • websearch│ │ • read   │ │Prometheus │           │
│              │ • file ops │ │ • write  │ │  Metis    │           │
│              │ • gateway │ │ • edit   │ │  Momus    │           │
│              │ • automation│ │ • bash │ │           │           │
│              └───────────┘ └─────┬─────┘ └───────────┘           │
│                                  │                                 │
│                         ┌────────▼────────┐                      │
│                         │     Memory       │                      │
│                         │  (Session/Facts) │                      │
│                         └─────────────────┘                       │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Task Execution Lifecycle

```
1. INPUT          → User provides task
       │
2. INTENT         → IntentGate classifies intent
       │
3. PLAN           → Prometheus creates execution plan
       │
4. EXECUTE        → Sisyphus runs the plan autonomously
       │
5. TOOLS          → Tools execute (OpenCode or native)
       │
6. DISTILL        → Output filtered through distillation
       │
7. MEMORIZE       → Key facts stored in memory
       │
8. REPORT         → Progress/status sent to user
       │
9. VERIFY         → User approves (UAT-style)
       │
10. LEARN         → Success/failure patterns stored
```

### Example Session

```bash
$ aigo "Fix the authentication bug in the login flow"

# Aigo's response:
# ✓ Understood: Fix authentication bug
# ✓ Planning: Creating execution plan...
# ✓ Plan created:
#   1. Explore login code structure
#   2. Identify authentication logic
#   3. Find the bug
#   4. Implement fix
#   5. Run tests
# ✓ Executing autonomously...
# ─────────────────────────────
# [1/5] Exploring login code...
# [2/5] Analyzing auth logic... 
# [3/5] Found bug: missing token validation
# [4/5] Implementing fix...
# [5/5] Running tests...
# ─────────────────────────────
# ✓ Fix complete! 2 tests passed.
# 
# Summary: Fixed missing token validation in auth.go:142
# Confidence: High
# Learnings saved to memory.
```

---

## Installation

### Prerequisites

- **Go 1.21+** — Required for building
- **Git** — For cloning
- **LLM API Key** (optional) — OpenAI, Anthropic, or local model

### Install from Source

```bash
# Clone the repository
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo

# Build the binary
make build

# Run setup wizard
make setup

# Or run directly
./dist/aigo
```

### Install with Make

```bash
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo
make install  # Builds and installs to ~/.local/bin/aigo
```

### Docker (Coming Soon)

```bash
docker pull ghcr.io/ahmad-ubaidillah/aigo:latest
docker run -it ghcr.io/ahmad-ubaidillah/aigo:latest
```

### Dependencies

Aigo uses pure Go dependencies where possible:

| Package | Purpose |
|---------|---------|
| `modernc.org/sqlite` | Local database (pure Go) |
| `cobra` | CLI framework |
| `bubbletea` | Terminal UI |
| `lipgloss` | TUI styling |
| `yaml.v3` | Configuration |
| `go-openai` | LLM integration |

---

## Usage

### Interactive Mode (TUI)

```bash
aigo
```

Opens an interactive terminal UI with:
- Session history
- Command palette
- Progress indicators
- Memory browser

### Single Task Mode

```bash
# Run a single task and exit
aigo run "create a REST API for user management"
aigo run "fix the null pointer exception in main.go"
```

### CLI Commands

```bash
# Session management
aigo session create --name "my-project"
aigo session list
aigo session resume <session-id>
aigo session delete <session-id>

# Memory management
aigo memory add "Project uses PostgreSQL"
aigo memory search "database"
aigo memory list --category "tech_stack"
aigo memory delete <memory-id>

# Tool execution
aigo tool list
aigo tool run <tool-name> --args <json>

# Diagnostics
aigo doctor          # Check system health
aigo config show     # Show current config
aigo config edit     # Edit configuration

# Gateway
aigo gateway start   # Start gateway server
aigo gateway status  # Check gateway status
```

### Configuration

Create `~/.aigo/config.yaml`:

```yaml
# LLM Configuration
llm:
  provider: "openai"  # openai, anthropic, local, openrouter
  model: "gpt-4o"
  api_key: "${OPENAI_API_KEY}"  # Use env var

# Context settings
context:
  max_tokens: 100000
  distillation_enabled: true
  
# Memory settings
memory:
  storage: "sqlite"  # sqlite, memory
  auto_extract: true
  
# Gateway settings
gateway:
  enabled: true
  platforms:
    - telegram
    - discord
```

### Environment Variables

```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export AIGO_CONFIG_PATH="/path/to/config.yaml"
```

---

## Features

### 1. Intent Classification

Two-tier classification system:

| Tier | Method | Speed | Accuracy |
|------|--------|-------|----------|
| **Fast Path** | Rule-based regex patterns | <1ms | ~70% |
| **Fallback** | LLM classification | ~100ms | ~95% |

### 2. Context Engine

Tiered memory for efficient context management:

- **L0 (Abstract)** — High-level summary
- **L1 (Overview)** — Key decisions and current state
- **L2 (Detail)** — Full conversation history

### 3. Task Router

Routes tasks to specialized handlers:

```
Intent → Handler Mapping
─────────────────────────
code_edit    → OpenCode Handler
code_read    → OpenCode Handler
web_search   → Native Handler
file_ops     → Native Handler
planning     → Planning Layer
automation   → Native Handler
```

### 4. Gateway System

Multi-platform messaging support:

| Platform | Status | Features |
|----------|--------|----------|
| **Telegram** | ✅ | Commands, inline queries |
| **Discord** | ✅ | Slash commands, buttons |
| **Slack** | ✅ | App mentions, shortcuts |
| **WhatsApp** | ⚠️ | Basic messaging |
| **CLI** | ✅ | Full features |
| **Web GUI** | ✅ | Browser interface |

### 5. Memory System

Persistent, searchable memory:

```bash
# Store facts
aigo memory add "User prefers TypeScript over JavaScript" --category preferences

# Search memories
aigo memory search "TypeScript"  # Returns relevant memories

# Auto-extraction
# Aigo automatically extracts and stores:
# - User preferences
# - Project facts
# - Error patterns
# - Success stories
```

### 6. Token Optimization

Distillation pipeline reduces token usage by 60-90%:

```
Distillation Stages:
1. Classifier  → Detects content type (error, success, info, etc.)
2. Scorer      → Assigns relevance scores (+0.4 for hot files)
3. Collapse    → Compresses repetitive patterns
4. Composer    → Outputs high-signal, minimal noise
```

### 7. Self-Healing

Automatic error recovery:

- Stack trace analysis
- Pattern-based fix suggestions
- Automatic retry with learned improvements

### 8. Multi-Agent System

Spawn specialized agents for complex tasks:

```go
// Example: Run multiple agents in parallel
results, _ := executor.ExecuteParallel(ctx, 
    []string{"atlas", "cody", "testa"},
    "Design and implement a user authentication system",
    sessionID,
)
```

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                           AIGO CORE                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │   Intent    │  │   Context   │  │       Planning          │ │
│  │   Gate      │  │   Engine    │  │  Prometheus/Metis/Momus │ │
│  └──────┬──────┘  └──────┬──────┘  └───────────┬─────────────┘ │
│         │                │                      │               │
│         └────────────────┼──────────────────────┘               │
│                          │                                      │
│                    ┌─────▼─────┐                               │
│                    │   Task    │                               │
│                    │   Router  │                               │
│                    └─────┬─────┘                               │
│                          │                                      │
│    ┌─────────────────────┼─────────────────────┐                │
│    │                     │                     │                 │
│┌───▼───┐          ┌──────▼─────┐        ┌─────▼─────┐        │
│ Native │          │  OpenCode   │        │  Memory   │        │
│Handler │          │   Handler  │        │  Engine   │        │
│        │          │            │        │           │        │
│ • bash │          │ • read    │        │ • SQLite  │        │
│ • search│          │ • write   │        │ • Vector  │        │
│ • file │          │ • edit    │        │ • Facts   │        │
│ • gateway│         │ • grep    │        │ • Session │        │
│ └───────┘          │ • ast_grep│        └───────────┘        │
│                    └───────────┘                                │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                      TOOL REGISTRY                          │ │
│  │  100+ built-in tools • Skill system • MCP integration      │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
aigo/
├── cmd/aigo/              # Entry point, CLI commands
│   ├── main.go
│   └── commands.go
│
├── internal/
│   ├── agent/             # Core agent loop
│   ├── agents/            # Multi-agent roles (Aigo, Atlas, Cody, etc.)
│   ├── budget/            # Token budget management
│   ├── context/           # L0/L1/L2 context engine
│   ├── distill/           # Token distillation pipeline
│   ├── execution/         # Atlas orchestrator
│   ├── gateway/           # Multi-platform adapters
│   ├── handlers/          # Native task handlers
│   ├── healing/           # Self-healing loop
│   ├── hooks/             # Lifecycle hooks (48 events)
│   ├── intent/            # Intent classification
│   ├── llm/               # LLM providers (OpenAI, Anthropic, etc.)
│   ├── memory/            # SQLite session + memory
│   ├── nodes/             # Agent nodes/workers
│   ├── opencode/          # OpenCode delegation client
│   ├── orchestration/     # Multi-agent orchestration
│   ├── permission/        # Permission ruleset
│   ├── planning/          # Prometheus/Metis/Momus planners
│   ├── selfimprove/       # Self-improvement system
│   ├── skills/            # Skill system
│   ├── templates/         # Prompt templates
│   ├── token/             # Token optimization (Toon format)
│   ├── tools/             # Tool registry
│   ├── tui/               # Terminal UI (bubbletea)
│   ├── vector/            # Vector store
│   ├── vectordb/          # Vector database client
│   ├── web/               # Web GUI server
│   ├── wisdom/            # Wisdom accumulation
│   └── workers/           # Worker agents
│
├── pkg/types/             # Shared types
├── web/                   # Web assets
├── configs/               # Default configurations
└── docs/                  # Documentation
```

---

## Inspirations

Aigo is inspired by and combines the best features from these open-source projects:

| Project | Key Inspiration |
|---------|----------------|
| **[OpenCode](https://github.com/opencode-ai/opencode)** | Tool framework, autonomous loop, LSP integration |
| **[Oh-My-OpenAgent](https://github.com/oh-my-openagent/oh-my-openagent)** | Prometheus planner, Sisyphus orchestrator, hash-anchored edits |
| **[Omni](https://github.com/omni/omni)** | Token distillation, signal filtering, RewindStore |
| **[Hermes](https://github.com/hermes/hermes)** | Multi-agent system, 48 lifecycle hooks, skill marketplace |
| **[mem0](https://github.com/mem0ai/mem0)** | Memory layers, fact extraction, vector search |
| **[MantisClaw](https://github.com/mantis/mantisclaw)** | Self-healing, auto-retry, auto pip-install |
| **[OpenWork](https://github.com/openwork/openwork)** | Workspace isolation, hot reload, design tokens |
| **[Vibe-Kanban](https://github.com/vibe/vibe-kanban)** | Task board, issue management, MCP integration |

### Feature Map

```
Aigo Feature          ← Inspired By
─────────────────────────────────────────────────────
Sisyphus Loop         ← OpenCode, Oh-My-OpenAgent
Prometheus Planner    ← Oh-My-OpenAgent
Token Distillation    ← Omni
Memory Graph          ← mem0, Supermemory
Multi-Agent           ← Hermes
Self-Healing          ← MantisClaw
Hash-Anchored Edits   ← Oh-My-OpenAgent
48 Lifecycle Hooks     ← Hermes
Skill System          ← Hermes
Workspace Isolation   ← OpenWork
Issue Tracking        ← Vibe-Kanban
```

---

## Differences from Similar Apps

### Aigo vs OpenCode

| Aspect | OpenCode | Aigo |
|--------|----------|------|
| **Language** | TypeScript | Go |
| **Focus** | Coding-centric | Full-stack agent |
| **Memory** | Session only | Persistent memory |
| **Planning** | Basic | Prometheus planner |
| **Multi-Agent** | Subagents only | Full multi-agent |
| **Token Efficiency** | Standard | 60-90% reduction |

### Aigo vs Claude Code

| Aspect | Claude Code | Aigo |
|--------|------------|------|
| **Platform** | API-only | Local + Gateway |
| **Memory** | Ephemeral | Persistent |
| **Cost** | Per-token | One-time build |
| **Privacy** | Cloud-based | Local-first |
| **Extensibility** | Limited | Full MCP support |

### Aigo vs Cursor

| Aspect | Cursor | Aigo |
|--------|--------|------|
| **Interface** | IDE plugin | CLI + TUI + Web |
| **Autonomy** | Human-in-loop | Fully autonomous |
| **Privacy** | Cloud | Local |
| **Multi-Platform** | IDE only | Any messaging platform |

### What Makes Aigo Unique

1. **Go-Native** — Single binary, no Node.js dependency
2. **Local-First** — All data stays on your machine
3. **Token Efficiency** — 60-90% reduction via distillation
4. **Multi-Platform** — CLI, TUI, Web, Telegram, Discord, Slack
5. **Self-Learning** — Learns from your patterns
6. **Prometheus Planning** — Plans before executing
7. **48 Lifecycle Hooks** — Fine-grained control

---

## Configuration

### Config File Locations

Aigo searches for config in this order:

1. `--config` flag: `aigo --config /path/to/config.yaml`
2. Project level: `./.aigo/config.yaml`
3. User level: `~/.aigo/config.yaml`
4. System level: `/etc/aigo/config.yaml`

### Full Configuration Reference

```yaml
# ~/.aigo/config.yaml

app:
  name: "aigo"
  version: "1.0.0"
  log_level: "info"  # debug, info, warn, error

# LLM Configuration
llm:
  provider: "openai"  # openai, anthropic, local, openrouter, glm
  model: "gpt-4o"
  api_key: "${OPENAI_API_KEY}"
  base_url: ""  # For proxies/custom endpoints
  timeout: 60s

# Context settings
context:
  max_tokens: 100000
  l0_tokens: 50
  l1_tokens: 500
  distillation_enabled: true
  auto_compress: true

# Memory settings
memory:
  storage: "sqlite"  # sqlite, memory
  db_path: "~/.aigo/memory.db"
  auto_extract: true
  extraction_interval: 5m

# Gateway settings
gateway:
  enabled: true
  host: "0.0.0.0"
  port: 8080
  platforms:
    telegram:
      enabled: true
      bot_token: "${TELEGRAM_BOT_TOKEN}"
    discord:
      enabled: false
      bot_token: "${DISCORD_BOT_TOKEN}"
    slack:
      enabled: false
      bot_token: "${SLACK_BOT_TOKEN}"

# Tools configuration
tools:
  allowed:
    - bash
    - read
    - write
    - edit
    - glob
    - grep
    - websearch
    - webfetch
  denied: []
  timeout: 30s

# Self-healing
healing:
  enabled: true
  max_retries: 3
  auto_fix: false  # Ask before fixing

# Skills
skills:
  directory: "~/.aigo/skills"
  auto_update: false

# Hooks
hooks:
  directory: "~/.aigo/hooks"
  enabled: true
```

---

## Development

### Building from Source

```bash
# Clone and build
git clone https://github.com/ahmad-ubaidillah/aigo.git
cd aigo
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Lint code
make lint

# Full CI pipeline
make ci
```

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Code Style

This project follows Go best practices:
- Run `make fmt` before committing
- Run `make lint` to check for issues
- Write tests for new features
- Update documentation

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

## Acknowledgments

Aigo stands on the shoulders of giants. Special thanks to:

- The **OpenCode** team for the tool framework
- The **Oh-My-OpenAgent** community for orchestration patterns
- The **Omni** team for token optimization insights
- The **Hermes** project for multi-agent inspiration
- All contributors to the open-source AI agent ecosystem

---

<p align="center">
  <strong>Built with ❤️ by developers, for developers</strong><br>
  <a href="https://github.com/ahmad-ubaidillah/aigo">GitHub</a> •
  <a href="https://github.com/ahmad-ubaidillah/aigo/issues">Issues</a> •
  <a href="https://github.com/ahmad-ubaidillah/aigo/discussions">Discussions</a>
</p>
