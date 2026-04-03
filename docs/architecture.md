# Aigo Architecture

## Overview

Aigo is a modular AI agent platform built in Go. It follows a layered architecture where each layer has a single responsibility and communicates through well-defined interfaces.

```
┌─────────────────────────────────────────────────┐
│                   CLI / TUI / Web GUI            │
├─────────────────────────────────────────────────┤
│  Gateway Layer (Telegram, Discord, Slack, WA)    │
├─────────────────────────────────────────────────┤
│  Agent Layer (Autonomous Loop, Classifier)       │
├──────────────┬──────────────┬───────────────────┤
│  Tool Layer  │  Distill     │  Context Engine   │
├──────────────┼──────────────┼───────────────────┤
│  Planning    │  Execution   │  Workers          │
├──────────────┼──────────────┼───────────────────┤
│  Memory      │  Vector DB   │  Wisdom           │
├──────────────┼──────────────┼───────────────────┤
│  SQLite (Sessions, Tasks, Memories, Profiles)    │
└─────────────────────────────────────────────────┘
```

## Package Structure

| Package | Responsibility |
|---|---|
| `cmd/aigo` | Entry point, CLI commands |
| `internal/agent` | Agent loop, autonomous execution, doom loop detection |
| `internal/context` | L0/L1/L2 context engine, token budgeting, compaction |
| `internal/tools` | Tool interface, registry, 11 built-in tools, permissions |
| `internal/distill` | OMNI-style distillation: Classify → Score → Collapse → Compose |
| `internal/rewind` | SHA-256 content archive with short hash retrieval |
| `internal/planning` | Prometheus (planner), Metis (gap analysis), Momus (reviewer) |
| `internal/execution` | Atlas orchestrator: todo management, wisdom tracking |
| `internal/workers` | Sisyphus, Hephaestus, Oracle, Librarian, Explore agents |
| `internal/memory` | SQLite session DB, profiles, fact extraction |
| `internal/vectordb` | Vector store with cosine similarity search |
| `internal/wisdom` | Lesson extraction, pattern recognition |
| `internal/hooks` | Lifecycle hooks (startup, session, agent events) |
| `internal/permission` | Ruleset-based permission system (allow/deny/ask) |
| `internal/gateway` | Messaging platform adapters |
| `internal/handlers` | Native task handlers (file, system, research) |
| `internal/web` | REST API + Web GUI (HTMX + Alpine.js) |
| `internal/tui` | Terminal UI with bubbletea |
| `internal/skills` | Skill registry, marketplace, executor |
| `internal/cron` | Cron scheduler with expression parsing |
| `pkg/types` | Shared types (Session, Task, Memory, Config, etc.) |

## Key Design Decisions

### Tool System
Tools implement a unified `Tool` interface with schema validation. The `ToolRegistry` singleton manages registration and execution. Each tool returns a `ToolResult` with success status, output, and optional error.

### Distillation Pipeline
The OMNI-style pipeline reduces token usage by 70-90%:
1. **Classifier** detects content type (git diff, build output, logs, etc.)
2. **Scorer** assigns signal tiers (Critical, Important, Context, Noise)
3. **Collapse** compresses repetitive lines
4. **Composer** filters by tier threshold

### Agent Loop
The autonomous loop iterates until task completion with:
- Doom loop detection (repeated identical tool calls)
- Token budget tracking with auto-compaction
- Permission checking per tool execution
- Iteration budget (max 50 by default)

### Memory
SQLite stores sessions, tasks, memories, and profiles. FTS5 enables full-text search. The fact extractor identifies actionable facts from conversations. The wisdom store accumulates lessons across tasks.
