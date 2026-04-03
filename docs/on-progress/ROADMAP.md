# Aigo Roadmap — From MVP to Autonomous AI Agent Platform

> Vision: Aigo = OMO (orchestration) + OpenCode (CLI power) + OMNI (token efficiency) + mem0 (memory) + Hermes (multi-agent) + Claw (autonomous) + MantisClaw (self-healing) + Acontext (skill learning) + Spec-Kitty (spec-driven) + OpenWork (workspace) + Vibe-Kanban (task board)

---

## Reference Projects Summary

| Project | Core Value | Key Features |
|---------|------------|--------------|
| **OMNI** | Token Efficiency | Signal distillation (90% reduction), RewindStore |
| **OpenCode** | CLI Power | Tool system, autonomous loop, 900+ tools |
| **mem0** | Memory Layer | Vector store, fact extraction, memory lifecycle |
| **Hermes** | Multi-Agent | Delegate tool, 100+ skills, gateway system |
| **OMO** | Orchestration | Planning layer, 11 workers, hash-anchored edits |
| **MantisClaw** | Self-Healing | Auto-repair, auto pip-install, unlimited agents |
| **Acontext** | Skill Learning | Task extraction, distillation pipeline, skill memory |
| **Spec-Kitty** | Spec-Driven | Constitutional principles, kanban workflow, SDD |
| **OpenWork** | Workspace Mgmt | Session templates, sharing, scheduled jobs |
| **Vibe-Kanban** | Task Board | Issue management, drag-drop, MCP integration |
| **Claw-Code** | Autonomous Coding | Turn-based loop, bughunter, session persistence |

---

## Current State (v0.1)

| Feature | Status | Notes |
|---------|--------|-------|
| Intent Classification | ✅ Basic | 2-tier (rule + LLM), 12 intents |
| Context Engine | ✅ Basic | L0/L1/L2 tiers, simple pruning |
| Task Router | ✅ Basic | Routes to handlers by intent |
| OpenCode Integration | ✅ Basic | Delegates coding tasks |
| Memory/Session | ✅ Basic | SQLite + FTS5 |
| Gateway Adapters | ⚠️ Stub | Telegram, Discord, Slack, WhatsApp |
| TUI Dashboard | ⚠️ Basic | bubbletea skeleton |
| Web GUI | ❌ Empty | HTMX planned, not implemented |
| Skills System | ⚠️ Basic | Registry exists, no management |
| Tool System | ❌ Missing | No proper tool abstraction |
| Autonomous Loop | ❌ Missing | No agentic loop |
| Self-Improvement | ❌ Missing | No learning/reflection |
| Token Distillation | ❌ Missing | No OMNI-style compression |
| Multi-Agent | ❌ Missing | No delegation/parallel agents |
| Planning Agents | ❌ Missing | No Prometheus/Metis/Momus |
| Permission System | ❌ Missing | No access control |
| Cron/Scheduler | ⚠️ Stub | Folder exists, not implemented |
| Hooks System | ❌ Missing | No lifecycle events |
| Self-Healing | ❌ Missing | No auto-retry/auto-fix |
| Skill Learning | ❌ Missing | No distillation → skill pipeline |
| Spec-Driven Dev | ❌ Missing | No constitutional principles |
| Workspace Mgmt | ❌ Missing | No workspace presets/profiles |
| Kanban Board | ❌ Missing | No issue tracking UI |
| Bughunter | ❌ Missing | No automated bug detection |

---

## Gap Analysis: All Reference Projects

### From OMNI (Token Efficiency)
| Feature | Description | Priority |
|---------|-------------|----------|
| Signal Distillation Pipeline | 5-stage: Classify → Score → Collapse → Compose | HIGH |
| RewindStore | Zero-info-loss archive for dropped content | HIGH |
| Auto-Learn Patterns | Detect repetitive patterns, generate filters | MEDIUM |
| Session State Tracking | Hot files, active errors, inferred task | HIGH |
| Content Type Detection | GitDiff, BuildOutput, TestOutput, etc. | MEDIUM |
| Context Boost | Hot files (+0.1), active errors (+0.25) | MEDIUM |

### From OpenCode (CLI Architecture)
| Feature | Description | Priority |
|---------|-------------|----------|
| Tool System | Schema-validated tools with Zod-like validation | HIGH |
| Agent Types | build, plan, explore, general, compaction | HIGH |
| Autonomous Loop | Streaming LLM, tool execution, doom loop protection | HIGH |
| Permission System | Allow/deny/ask rules per tool/action | MEDIUM |
| Output Truncation | Smart truncation for large outputs | MEDIUM |
| Subagent Spawning | Task tool for child sessions | HIGH |
| Tool Repair | Auto-fix lowercase tool names, invalid calls | LOW |
| Prompt Caching | Anthropic prompt caching support | MEDIUM |

### From mem0 (Memory System)
| Feature | Description | Priority |
|---------|-------------|----------|
| Fact Extraction | LLM-based extraction from conversations | MEDIUM |
| Vector Store Integration | Chroma/Qdrant for semantic search | HIGH |
| Memory Lifecycle | ADD/UPDATE/DELETE/NONE decisions | MEDIUM |
| History Tracking | Audit trail of memory changes | LOW |
| Multi-tenant Scoping | user_id, agent_id, run_id | LOW |
| Graph Memory | Neo4j entity-relationship extraction | LOW |

### From Hermes (Multi-Agent)
| Feature | Description | Priority |
|---------|-------------|----------|
| Delegate Tool | Spawn child agents with isolated context | HIGH |
| Background Agent Pool | Parallel execution with ThreadPoolExecutor | MEDIUM |
| IntentGate | Smart task classification before routing | HIGH |
| Skills System | 100+ skills with YAML frontmatter | MEDIUM |
| Cron Job System | Persistent scheduled tasks | MEDIUM |
| Gateway Hooks | Lifecycle events (startup, session, agent) | MEDIUM |
| Tool Registry | Central singleton with toolset support | HIGH |
| Mixture of Agents | Multi-LLM collaboration for complex reasoning | LOW |

### From OMO (Orchestration)
| Feature | Description | Priority |
|---------|-------------|----------|
| Planning Layer | Prometheus (planner) + Metis (gap) + Momus (reviewer) | HIGH |
| Execution Layer | Atlas (todo orchestrator) | HIGH |
| Worker Agents | 11 specialized agents with fallback chains | MEDIUM |
| Hash-Anchored Edits | Line#ID hash validation (6.7% → 68.3% success) | HIGH |
| Wisdom Accumulation | Learnings passed to subsequent tasks | MEDIUM |
| Hook System | 5 tiers, 48 hooks | MEDIUM |
| Category System | Intent-based model selection | LOW |
| Intent Gate | Classifies true intent before acting | HIGH |
| Boulder Continuation | Forces completion - agent can't stop halfway | MEDIUM |

### From MantisClaw (Self-Healing)
| Feature | Description | Priority |
|---------|-------------|----------|
| Tool-Call Loop | 25-turn autonomous execution | HIGH |
| Self-Healing Execution | Auto-detect failures, read traceback, fix code | HIGH |
| Auto Pip-Install | Detect and install missing Python packages | MEDIUM |
| Auto Skills System | Agent writes, debugs, executes skills by itself | HIGH |
| Scheduled Autonomy | Cron-based tasks run without user presence | MEDIUM |
| Unlimited Agents | Multiple agent instances, channel mapping | MEDIUM |
| Multi-Channel Control | WhatsApp, Telegram, Slack, built-in Chat | HIGH |
| Embedded Python Kernel | Built-in interpreter for code execution | MEDIUM |

### From Acontext (Skill Learning)
| Feature | Description | Priority |
|---------|-------------|----------|
| Session Buffering | Message buffer with TTL, overflow handling | HIGH |
| Task Extraction Agent | Auto-extract tasks from conversations | HIGH |
| Task Status Tracking | pending, running, success, failed | MEDIUM |
| Skill Memory Layer | Learning spaces for storing learned skills | HIGH |
| Two-Phase Learning | Distillation → Skill Agent pipeline | HIGH |
| Skill Distillation | LLM extracts what worked/failed from tasks | HIGH |
| Context Engineering | remove_tool_result, middle_out compression | MEDIUM |
| Virtual Disk System | Skills stored as files in virtual disk | LOW |

### From Spec-Kitty (Spec-Driven Development)
| Feature | Description | Priority |
|---------|-------------|----------|
| Constitutional Foundation | Immutable principles enforced via templates | MEDIUM |
| Library-First Principle | Every feature starts as standalone library | LOW |
| Test-First Imperative | NO code before tests | HIGH |
| Spec Structure | spec.md, plan.md, tasks.md, status.json | MEDIUM |
| Workflow Commands | specify → plan → tasks → implement → review → accept | MEDIUM |
| Status Model | Event-sourced kanban: planned → claimed → in_progress → done | HIGH |
| Multi-Agent Templates | 12 AI agents with slash commands | LOW |
| Quality Gates | Constitutional compliance checks | MEDIUM |

### From OpenWork (Workspace Management)
| Feature | Description | Priority |
|---------|-------------|----------|
| Workspace Management | Local folders + remote servers + cloud workers | MEDIUM |
| Workspace Presets | starter, automation, minimal configurations | LOW |
| Session Templates | Pre-configured conversation blueprints | MEDIUM |
| Skills & Automation | Markdown skills in .opencode/skills/ | HIGH |
| Scheduled Jobs | Cron-like scheduling (macOS LaunchAgents, Linux systemd) | MEDIUM |
| Sharing & Collaboration | Share bundles, template import, workspace profiles | LOW |
| Messaging Bridge | Slack/Telegram bot routes to sessions | MEDIUM |
| Approval Modes | manual (ask) vs auto permissions | MEDIUM |

### From Vibe-Kanban (Task Board)
| Feature | Description | Priority |
|---------|-------------|----------|
| Issue Management | title, description, priority, status, assignees, tags | HIGH |
| Sub-Issues | Parent/child relationships | MEDIUM |
| Issue Relationships | blocking, related, has_duplicate | LOW |
| Status Workflow | Customizable kanban columns with drag-drop | HIGH |
| Filtering & Sorting | priority, assignees, tags, hide blocked | MEDIUM |
| Views | Kanban board + list view | MEDIUM |
| MCP Integration | create_issue, list_issues, update_issue tools | HIGH |
| Real-time Sync | ElectricSQL for collaboration | LOW |
| PR Linking | Connect issues to pull requests | LOW |

### From Claw-Code (Autonomous Coding)
| Feature | Description | Priority |
|---------|-------------|----------|
| Turn-Based Loop | Configurable max_turns (default: 8) | HIGH |
| Agent/Subagent System | verificationAgent, exploreAgent, planAgent, generalAgent | HIGH |
| Fork Subagent | Spawn independent subagents | MEDIUM |
| Resume Agent | Continue interrupted sessions | HIGH |
| 900+ Tools | File, Bash, MCP, Web tools mirrored | MEDIUM |
| Permission Gating | ReadOnly, WorkspaceWrite, DangerFullAccess | MEDIUM |
| Bughunter Command | Automated bug detection and fixing | HIGH |
| Autofix-PR | Automatic PR repair | MEDIUM |
| Context Compaction | Auto-compact after N turns | HIGH |
| Session Persistence | Load/persist session state | HIGH |

---

## Target Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              AIGO PLATFORM                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │     CLI     │  │     TUI     │  │   Web GUI   │  │   Gateway   │            │
│  │   (cobra)   │  │ (bubbletea) │  │   (HTMX)    │  │ (tele/disc) │            │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘            │
│         │                │                │                │                    │
│         └────────────────┴────────────────┴────────────────┘                    │
│                                   │                                              │
│                        ┌──────────▼──────────┐                                   │
│                        │     IntentGate      │                                   │
│                        │  (Classification)   │                                   │
│                        └──────────┬──────────┘                                   │
│                                   │                                              │
│         ┌─────────────────────────┼─────────────────────────┐                   │
│         │                         │                         │                   │
│  ┌──────▼──────┐          ┌──────▼──────┐          ┌──────▼──────┐              │
│  │   PLANNING  │          │  EXECUTION  │          │  WORKERS    │              │
│  │   LAYER     │          │    LAYER    │          │   LAYER     │              │
│  ├─────────────┤          ├─────────────┤          ├─────────────┤              │
│  │ Prometheus  │          │    Atlas    │          │ Sisyphus    │              │
│  │ Metis       │─────────▶│ (orchestr.) │─────────▶│ Hephaestus  │              │
│  │ Momus       │          │             │          │ Oracle      │              │
│  └─────────────┘          └──────┬──────┘          │ Librarian   │              │
│                                  │                 │ Explore     │              │
│                                  │                 │ Bughunter   │              │
│                                  │                 └─────────────┘              │
│                                  │                                              │
│                        ┌─────────▼──────────┐                                    │
│                        │   TOOL REGISTRY    │                                    │
│                        │ ─────────────────  │                                    │
│                        │ bash, read, write  │                                    │
│                        │ edit, glob, grep   │                                    │
│                        │ task, webfetch     │                                    │
│                        │ skill, todo, ...   │                                    │
│                        │ 900+ tools total   │                                    │
│                        └─────────┬──────────┘                                    │
│                                  │                                               │
│         ┌────────────────────────┼────────────────────────┐                     │
│         │                        │                        │                     │
│  ┌──────▼──────┐          ┌──────▼──────┐          ┌──────▼──────┐              │
│  │   MEMORY    │          │  DISTILL    │          │   HOOKS     │              │
│  │   ENGINE    │          │   ENGINE    │          │   SYSTEM    │              │
│  ├─────────────┤          ├─────────────┤          ├─────────────┤              │
│  │ Vector DB   │          │ Classifier  │          │ Session     │              │
│  │ Fact Extr.  │          │ Scorer      │          │ Tool Guard  │              │
│  │ L0/L1/L2    │          │ Collapse    │          │ Transform   │              │
│  │ RewindStore │          │ Composer    │          │ Continuation│              │
│  │ Skill Mem   │          │ Auto-Learn  │          │ Bughunter   │              │
│  └─────────────┘          └─────────────┘          └─────────────┘              │
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                          SELF-HEALING LAYER                                │  │
│  │  Auto Retry │ Traceback Analysis │ Auto Pip-Install │ Auto Skill Gen      │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                          KANBAN / TASK BOARD                               │  │
│  │  Issues │ Sub-issues │ Relationships │ Status Workflow │ MCP Tools        │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                          WORKSPACE MANAGEMENT                              │  │
│  │  Presets │ Templates │ Profiles │ Scheduled Jobs │ Sharing               │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌───────────────────────────────────────────────────────────────────────────┐  │
│  │                          PERSISTENCE LAYER                                 │  │
│  │  SQLite (sessions) │ Chroma (vectors) │ RewindStore (archive)            │  │
│  │  Skill Memory (learned skills) │ Task Store (kanban)                      │  │
│  └───────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Phases

### Phase 1: Foundation (4-6 weeks)
**Goal: Core autonomous agent with tool system**

- [ ] Tool system with schema validation (from OpenCode)
- [ ] Autonomous agent loop with max_turns (from Claw-Code)
- [ ] Session state tracking (from OMNI)
- [ ] Basic distillation pipeline (from OMNI)
- [ ] Session persistence & resume (from Claw-Code)
- [ ] Permission gating (ReadOnly, WorkspaceWrite, DangerFullAccess)

### Phase 2: Orchestration (4-6 weeks)
**Goal: Multi-agent planning and execution**

- [ ] Planning agents: Prometheus, Metis, Momus (from OMO)
- [ ] Execution agent: Atlas orchestrator (from OMO)
- [ ] Worker agents: Sisyphus, Hephaestus, Oracle (from OMO)
- [ ] Delegate tool for parallel execution (from Hermes)
- [ ] Hash-anchored edits (from OMO)
- [ ] Wisdom accumulation across tasks

### Phase 3: Memory & Learning (3-4 weeks)
**Goal: Persistent memory and skill learning**

- [ ] Vector store integration - Chroma (from mem0)
- [ ] Fact extraction from conversations (from mem0)
- [ ] RewindStore implementation (from OMNI)
- [ ] Skill memory layer (from Acontext)
- [ ] Two-phase learning: Distillation → Skill Agent (from Acontext)
- [ ] Task extraction agent (from Acontext)

### Phase 4: Self-Healing (2-3 weeks)
**Goal: Autonomous error recovery**

- [ ] Self-healing execution loop (from MantisClaw)
- [ ] Auto traceback analysis and fix
- [ ] Auto pip-install for missing packages
- [ ] Auto skill generation from failures (from MantisClaw)
- [ ] Bughunter command (from Claw-Code)
- [ ] Context compaction after N turns

### Phase 5: Workspace & Kanban (3-4 weeks)
**Goal: Project and task management**

- [ ] Workspace presets and profiles (from OpenWork)
- [ ] Session templates (from OpenWork)
- [ ] Scheduled jobs/cron (from OpenWork, Hermes)
- [ ] Issue/kanban system (from Vibe-Kanban)
- [ ] MCP integration for issues (from Vibe-Kanban)
- [ ] Status workflow with drag-drop

### Phase 6: Integration (2-3 weeks)
**Goal: Gateway and hooks**

- [ ] Gateway hooks system (from Hermes)
- [ ] Messaging bridge (from OpenWork)
- [ ] Skills system expansion (from Hermes)
- [ ] Multi-channel control (from MantisClaw)
- [ ] Spec-driven workflow (from Spec-Kitty)

### Phase 7: Polish (2-3 weeks)
**Goal: Production readiness**

- [ ] Web GUI completion
- [ ] TUI improvements with RPG feel
- [ ] Documentation
- [ ] Testing & benchmarks
- [ ] Constitutional principles enforcement (from Spec-Kitty)

---

## Success Metrics

| Metric | Current | Target |
|--------|---------|--------|
| Token Efficiency | 0% (no distillation) | 70-90% reduction |
| Autonomous Task Completion | 0% (delegates to OpenCode) | 80%+ |
| Intent Classification Accuracy | ~70% (rule-based) | 95%+ |
| Memory Recall Accuracy | N/A | 90%+ |
| Multi-Agent Parallelism | 0% | 5 concurrent agents |
| Self-Improvement Cycle | N/A | Every 10 turns |
| Self-Healing Success Rate | 0% | 85%+ auto-recovery |
| Skill Learning Rate | 0% | 1 skill/10 tasks |
| Kanban Issue Tracking | 0% | Full MCP integration |

---

## Priority Matrix

```
                    HIGH IMPACT
                         │
    ┌────────────────────┼────────────────────┐
    │                    │                    │
    │  PHASE 1           │   PHASE 2          │
    │  Foundation        │   Orchestration    │
    │  • Tool System     │   • Planning Layer │
    │  • Autonomous Loop │   • Worker Agents  │
    │  • Distillation    │   • Delegate Tool  │
    │                    │                    │
LOW ┼────────────────────┼────────────────────┼ HIGH
EFFORT                   │                   EFFORT
    │                    │                    │
    │  PHASE 4           │   PHASE 3          │
    │  Self-Healing      │   Memory/Learning  │
    │  • Auto Retry      │   • Vector Store   │
    │  • Bughunter       │   • Skill Memory   │
    │  • Auto Pip-Install│   • RewindStore    │
    │                    │                    │
    └────────────────────┼────────────────────┘
                         │
                    LOW IMPACT
```

---

## Key Decisions

### Architecture Decisions
1. **Go for Core** — Performance + simplicity, stick with Go
2. **Chroma for Vectors** — Local-first, no API keys needed
3. **SQLite for Sessions** — Already implemented, keep it
4. **Bubbletea for TUI** — Already integrated, expand it
5. **HTMX for Web** — Lightweight, no heavy JS framework

### Feature Decisions
1. **Distillation First** — Token efficiency is critical for cost
2. **Self-Healing Priority** — Reduces user intervention significantly
3. **Kanban MCP** — Enables AI agents to manage their own tasks
4. **Skill Learning Pipeline** — Accumulates wisdom over time
5. **Hash-Anchored Edits** — Solves "harness problem" (6.7% → 68.3%)

### Scope Decisions
1. **No Cloud Required** — Everything runs locally
2. **No API Keys Mandatory** — Work with free models (OpenCode)
3. **Multi-Platform** — CLI, TUI, Web, Gateway all supported
4. **Extensible** — Skills, hooks, MCP servers for customization

---

## References

- [OMNI](../Reference/omni-main) — Token efficiency, distillation
- [OpenCode](../Reference/opencode-dev) — CLI architecture, tools
- [mem0](../Reference/mem0-main) — Memory layer
- [Hermes](../Reference/hermes) — Multi-agent, skills, gateway
- [OMO](../Reference/oh-my-openagent-dev) — Orchestration, planning
- [MantisClaw](../Reference/MantisClaw-main) — Self-healing, auto skills
- [Acontext](../Reference/Acontext-main) — Skill learning, distillation
- [Spec-Kitty](../Reference/spec-kitty-main) — Spec-driven development
- [OpenWork](../Reference/openwork-dev) — Workspace management
- [Vibe-Kanban](../Reference/vibe-kanban-main) — Task board, issues
- [Claw-Code](../Reference/claw-code-main) — Autonomous coding, bughunter
- [CoPaw](../Reference/CoPaw-main) — Claw variant
