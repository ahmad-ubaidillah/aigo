# Aigo Feature Roadmap — Inspired by Reference Frameworks

## Executive Summary

Aigo adalah **OMO + OpenCode + Autonomous** yang terintegrasi penuh — tidak perlu download/setup terpisah. Setelah install, user cukup pilih mode (CLI atau Web UI), input API key atau setup local LLM, dan langsung bisa digunakan dengan Sisyphus/Prometheus mode sudah default aktif.

Vision: **Full autonomous AI agent** seperti di software house — user hanya perlu:
- Memberikan task di awal
- Verifikasi hasil akhir (UAT-style)

Aigo akan belajar dari setiap task, mengingat konteks project, dan tidak mengulangi kesalahan.

## Reference Frameworks Overview

| Framework | Key Strength | Reference Path | Applied to Aigo |
|-----------|--------------|----------------|-----------------|
| **Supermemory** | Living knowledge graph, memory profiles, vector search | `./reference/supermemory-main` | ✅ Memory graph, profiles |
| **OpenViking** | 6-category memory extraction, session lifecycle, archival | `./reference/OpenViking-main` | ✅ 6-category extraction |
| **Mem0** | Three memory layers (short/working/long), graph memory | `./reference/mem0-main` | ✅ Memory layers |
| **Omni** | Token optimization, message compression, cost savings | `./reference/omni-main` | ✅ Token optimization |
| **Hermes** | Autonomous execution, parallel agents, 48 lifecycle hooks | `./reference/hermes` | ✅ Autonomous execution |
| **OpenCode** | Tool framework, LSP integration, code analysis | `./reference/opencode-dev` | ✅ Core engine (OMO integration) |
| **OpenWork** | Architecture, design system, workspace isolation | `./reference/openwork-dev` | ✅ UI patterns |
| **Oh-My-OpenAgent** | Multi-agent orchestration, planning phase (Prometheus) | `./reference/oh-my-openagent-dev` | ✅ Sisyphus/Prometheus |

---

## Aigo Core Architecture — OMO + OpenCode Autonomous

### Architecture Principles

```
┌─────────────────────────────────────────────────────────────────┐
│                         AIGO CORE                               │
│                    (OMO + OpenCode Autonomous)                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐   │
│  │   OMO AI    │ +  │  OpenCode   │ +  │  Sisyphus Loop  │   │
│  │   Engine    │    │   Engine    │    │  (Autonomous)   │   │
│  └─────────────┘    └─────────────┘    └─────────────────┘   │
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐   │
│  │  Prometheus │    │   Memory    │    │  Token Optimizer│   │
│  │   Planner   │    │    Engine   │    │                 │   │
│  └─────────────┘    └─────────────┘    └─────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Differentiators

1. **No Separate Setup** — OMO + OpenCode langsung included dalam Aigo
2. **Default Autonomous** — Sisyphus, Prometheus planner langsung aktif
3. **Smart Memory** — Learn from conversations, remember context per project
4. **Full Autonomy** — Task dari start sampai finish, user hanya verify hasil
5. **Superior Token Efficiency** — Distillation + Toon format, 70-90% reduction, still smart
6. **Long Session Support** — Bisa dipakai berjam-jam karena hemat token

---

## User Requirements Integration

### User's Vision for Aigo

| Requirement | Description | Implementation |
|-------------|-------------|-----------------|
| **OMO + OpenCode Integration** | Tidak perlu download/setup terpisah | Core engine bundled |
| **CLI + Web UI** | User pilih mode setelah install | Dual interface support |
| **Single API Key** | Input API key atau setup local LLM | Unified model config |
| **Default Sisyphus/Prometheus** | Langsung aktif, tidak perlu setup 1-1 | Default mode |
| **Multiple Chat Channels** | CLI (TUI), Web UI, Telegram, Slack, dll | Gateway system |
| **Skill Download** | Download skill yang diinginkan | Skill marketplace |
| **MCP Setup** | Setup MCP yang mau dipakai | MCP configuration |
| **Learning Capability** | Belajar dari percakapan dan task | Memory + feedback loop |
| **Smart Memory** | Ingatan dan otak yang pintar | Memory graph |
| **Context Awareness** | Paham konteks, tahu project structure | Project-aware memory |
| **File Relationship Map** | Tau file apa, folder mana, hubungannya | Code graph |
| **Learn from Success/Failure** | Tidak mengulangi kesalahan | Error memory |
| **Full Autonomy** | Task start → finish, user verify only | UAT-style workflow |
| **Superior Token Efficiency** | Distillation + Toon format, 70-90% reduction | Token optimizer |
| **Long Session Support** | Hemat token, bisa dipakai berjam-jam | Distillation pipeline |

### UAT-Style Workflow

```
User gives task ──► Aigo executes autonomously ──► User verifies result
       │                    │                          │
   (Start)            (No intervention)           (End/UAT)
```

Ini seperti di software house — user berhubungan di awal project dan saat akan selesai untuk UAT.

---

## Feature Analysis: What to Apply to Aigo

#### Current Aigo State
- SQLite-based session storage with FTS5 full-text search
- Basic L0/L1/L2 context engine (in `internal/context/engine.go`)
- Simple memory table with category/tags

#### Features to Apply

| Feature | Source | Priority | Implementation Complexity |
|---------|--------|----------|---------------------------|
| **Memory Graph** | Supermemory, OpenViking | P0 - Critical | Medium |
| **6-Category Memory Extraction** | OpenViking | P0 - Critical | Medium |
| **Vector Search (Local)** | Supermemory, Mem0 | P0 - Critical | High |
| **Memory Profiles** | Supermemory | P1 - High | Medium |
| **Memory Archival** | OpenViking | P1 - High | Low |
| **Three Memory Layers** | Mem0 | P2 - Medium | Medium |

#### Implementation Plan

```
### Phase 1: Enhanced Memory (v0.3.0)
- [ ] Expand memory table: add relationships, version tracking
- [ ] Implement memory extraction from session messages
- [ ] Add container/tag system for memory organization
- [ ] Build memory profile generation (static + dynamic)

### Phase 2: Vector Storage (v0.4.0)
- [ ] Add local vector store (pure Go - use bfloat16 embeddings)
- [ ] Implement similarity search for memory retrieval
- [ ] Add hybrid search (FTS + vector)
- [ ] Memory graph visualization (optional)

### Phase 3: Memory Intelligence (v0.5.0)
- [ ] Automatic memory archival for cold memories
- [ ] Memory deduplication and merging
- [ ] Context-aware memory prioritization
```

---

### 2. Token Optimization (High Priority) — SUPERIOR TOKEN EFFICIENCY

#### Current Aigo State
- Basic token counting (4 chars = 1 token)
- L0 pruning with importance scoring
- Auto-compression after N turns

#### Features to Apply

| Feature | Source | Priority | Implementation Complexity |
|---------|--------|----------|---------------------------|
| **Smart Context Trimming** | Omni, Hermes | P0 - Critical | Medium |
| **Message Summarization** | Hermes | P0 - Critical | Medium |
| **Prompt Caching** | Hermes | P1 - High | Low |
| **Tool Result Compression** | OpenCode | P1 - High | Low |
| **Distillation Pipeline** | Omni | P0 - Critical | High |
| **Toon Format Integration** | toon-format | P0 - Critical | Medium |

#### Distillation Pipeline (Omni-Style)

Inspired by Omni, Aigo implements a multi-stage distillation pipeline untuk maximize token efficiency:

```
┌─────────────────────────────────────────────────────────────────┐
│                    DISTILLATION PIPELINE                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Raw Tool Output                                                │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│  │ Classifier  │───►│   Scorer    │───►│  Composer   │      │
│  │ (Content    │    │ (Context    │    │ (High-Signal│      │
│  │  Type)      │    │  Boost)     │    │  Output)    │      │
│  └─────────────┘    └─────────────┘    └─────────────┘      │
│       │                   │                   │               │
│       ▼                   ▼                   ▼               │
│  • Detect: error,        • +0.4 boost:      • Critical       │
│    success, progress,     hot files, recent  • Important      │
│    info, debug            commands, domain   • Context        │
│                           relevance           • Noise (skip)   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Token Reduction Target**: Up to 90% reduction on tool outputs while maintaining intelligence.

#### Toon Format Integration

Using [toon-format](https://github.com/toon-format/toon) untuk compact, readable output:

| Before (Raw) | After (Toon) | Token Reduction |
|--------------|--------------|-----------------|
| Full stack trace | Compact error summary | ~70% |
| Verbose logs | Key events only | ~80% |
| 1000 lines diff | Semantic diff | ~60% |
| Full JSON | Structure + sample | ~50% |

**Toon Benefits**:
- Go-native, lightweight
- Preserve semantic meaning
- Multiple output modes (compact, detailed, tree)
- Structured for LLM consumption

#### Implementation Plan

```
### Token Optimization (v0.3.0)
- [ ] Implement message summarization using LLM
- [ ] Add tool output truncation with key info preservation
- [ ] Smart prompt caching (avoid context rebuilding)
- [ ] Token budget monitoring and auto-pruning
- [ ] Context compression at intervals

### Distillation Pipeline (v0.3.5) — NEW
- [ ] Content classifier (error/success/progress/info/debug)
- [ ] Context scorer with hot files + recent commands boost
- [ ] Signal composer (Critical/Important/Context/Noise tier)
- [ ] RewindStore for raw data archival (SHA-256)
- [ ] Omni-style diff for comparison

### Toon Integration (v0.3.5) — NEW
- [ ] Integrate toon-format library
- [ ] Compact error output
- [ ] Semantic log parsing
- [ ] JSON structure extraction
- [ ] Tree view for nested data
```

#### Token Efficiency Metrics

| Metric | Target | Description |
|--------|--------|-------------|
| **Token Reduction** | ≥70% | Tool output size reduction |
| **Context Window** | 2x longer | Effective context with same tokens |
| **Cost Savings** | ≥50% | API cost reduction |
| **Intelligence** | ≥95% | Preserved semantic quality |
| **Response Time** | <100ms | Distillation latency |
### Token Optimization (v0.3.0)
- [ ] Implement message summarization using LLM
- [ ] Add tool output truncation with key info preservation
- [ ] Smart prompt caching (avoid context rebuilding)
- [ ] Token budget monitoring and auto-pruning
- [ ] Context compression at configurable intervals
```

---

### 3. Autonomous Agent Capabilities (High Priority)

#### Current Aigo State
- Single-agent loop with intent classification
- Task routing to handlers
- Error handling with consecutive error tracking

#### Features to Apply

| Feature | Source | Priority | Implementation Complexity |
|---------|--------|----------|---------------------------|
| **Multi-Agent Orchestration** | Hermes, Oh-My-OpenAgent | P0 - Critical | High |
| **Planning Phase (Prometheus)** | Oh-My-OpenAgent | P0 - Critical | High |
| **Parallel Agent Execution** | Hermes | P1 - High | High |
| **Lifecycle Hooks** | Hermes (48 hooks) | P1 - High | Medium |
| **Hash-Anchored Edits** | Oh-My-OpenAgent | P2 - Medium | Medium |

#### Implementation Plan

```
### Phase 1: Enhanced Agent (v0.3.0)
- [ ] Add subagent delegation capability
- [ ] Implement planning phase before execution
- [ ] Add lifecycle hooks (start, end, error, tool, etc.)

### Phase 2: Multi-Agent (v0.4.0)
- [ ] Agent registry and routing
- [ ] Parallel task execution
- [ ] Agent communication protocol
- [ ] Result aggregation

### Phase 3: Advanced Autonomy (v0.5.0)
- [ ] Self-correction based on error patterns
- [ ] Automatic skill selection
- [ ] Goal decomposition and planning
```

---

### 4. Coding Capabilities (High Priority)

#### Current Aigo State
- OpenCode integration for coding tasks
- Basic file operations
- Intent classification for code vs non-code

#### Features to Apply

| Feature | Source | Priority | Implementation Complexity |
|---------|--------|----------|---------------------------|
| **Tool Framework (Tool.define)** | OpenCode | P0 - Critical | Medium |
| **LSP Integration** | OpenCode, Oh-My-OpenAgent | P0 - Critical | High |
| **AST-Grep Support** | OpenCode | P1 - High | High |
| **Hash-Anchored Editing** | Oh-My-OpenAgent | P1 - High | Medium |
| **Code Generation Pipeline** | OpenCode | P2 - Medium | Medium |

#### Implementation Plan

```
### Enhanced Coding (v0.3.0)
- [ ] Expand OpenCode tool wrapper with more capabilities
- [ ] Add LSP client for goto definition, rename, diagnostics
- [ ] Implement hash-anchored edit system
- [ ] Add AST-based code search

### Advanced Tools (v0.4.0)
- [ ] Multi-file refactoring support
- [ ] Code review automation
- [ ] Test generation
- [ ] Code explanation and documentation
```

---

### 5. Local-First Architecture (Critical)

#### Principles (No External API Keys Except AI Models)

| Principle | Description |
|-----------|-------------|
| **All Data Local** | SQLite, local files - no cloud dependencies |
| **Offline-First** | Works without internet (except AI calls) |
| **User Ownership** | Data stays on user's machine |
| **Minimal Dependencies** | No optional external services |

#### Features to Apply

| Feature | Source | Priority | Implementation Complexity |
|---------|--------|----------|---------------------------|
| **Profile System** | Hermes | P1 - High | Medium |
| **Workspace Isolation** | OpenWork | P1 - High | Low |
| **Local Config Management** | Hermes, OpenWork | P0 - Critical | Low |
| **Data Export/Import** | All | P2 - Medium | Low |

---

### 6. UI/UX Improvements (Medium Priority)

#### Current Aigo State
- TUI with Bubble Tea + Lipgloss
- Web GUI with HTMX + Alpine.js
- Basic gateway support

#### Features to Apply

| Feature | Source | Priority | Implementation Complexity |
|---------|--------|----------|---------------------------|
| **Token-Driven Design System** | OpenWork | P1 - High | Medium |
| **Skin/Theme System** | Hermes | P1 - High | Medium |
| **Real-Time Updates (SSE)** | OpenWork | P1 - High | Medium |
| **Rich CLI Display** | Hermes | P2 - Medium | Low |

---

## Aigo-Specific Features (User Requirements)

### 1. Dual Interface — CLI + Web UI

| Feature | Description | Priority |
|---------|-------------|----------|
| **Setup Wizard** | Setelah install, user pilih mode CLI atau Web UI | P0 |
| **CLI with TUI** | Terminal UI untuk advanced user (Lipgloss/Bubble Tea) | P0 |
| **Web UI** | Browser-based interface untuk casual user | P0 |
| **Mode Switching** | User bisa switch antara CLI dan Web UI kapan saja | P1 |

#### Setup Flow

```
Install Aigo
      │
      ▼
┌─────────────┐
│ Setup Wizard│ ──► Pilih: CLI or Web UI
└─────────────┘
      │
      ▼
┌─────────────┐
│ API Key/    │ ──► Input API key atau setup local LLM
│ Local LLM   │
└─────────────┘
      │
      ▼
┌─────────────┐
│ Ready!      │ ──► Langsung bisa use dengan Sisyphus default
└─────────────┘
```

---

### 2. Gateway System — Multiple Chat Channels

| Channel | Description | Priority |
|---------|-------------|----------|
| **CLI (TUI)** | Interactive terminal dengan rich display | P0 |
| **Web UI** | Browser-based chat interface | P0 |
| **Telegram** | Messaging via Telegram bot | P1 |
| **Slack** | Messaging via Slack integration | P1 |
| **Discord** | Messaging via Discord bot | P2 |
| **WhatsApp** | Messaging via WhatsApp Business | P2 |

#### Architecture

```
                    ┌──────────────┐
                    │   Aigo Core   │
                    │ (OMO+OpenCode)│
                    └──────┬───────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
    ┌────▼────┐      ┌────▼────┐      ┌────▼────┐
    │  CLI    │      │ Web UI  │      │Gateway  │
    │ (TUI)   │      │ (HTMX)  │      │(Multi)  │
    └─────────┘      └─────────┘      └────┬────┘
                                           │
                    ┌──────────┬───────────┼───────────┬──────────┐
                    │          │           │           │          │
               ┌────▼───┐ ┌────▼───┐ ┌────▼───┐ ┌────▼───┐ ┌────▼───┐
               │Telegram│ │ Slack  │ │Discord │ │ WhatsApp│ │ Others │
               └────────┘ └────────┘ └────────┘ └────────┘ └────────┘
```

---

### 3. Skill Marketplace

| Feature | Description | Priority |
|---------|-------------|----------|
| **Skill Registry** | Daftar skill yang tersedia untuk didownload | P0 |
| **Skill Download** | Download skill sesuai kebutuhan | P0 |
| **Skill Management** | Enable/disable skill per project | P1 |
| **Skill Categories** | DevOps, Git, Creative, MLOps, dll | P1 |

#### Built-in Skills (Default)

| Category | Skills |
|----------|--------|
| **DevOps** | arch-dev-setup, pre-commit-checks |
| **Git** | github-pr-workflow, code-review, git-master |
| **System** | linux-scripts, docker-management |
| **Creative** | ascii-art, excalidraw |

---

### 4. MCP Configuration

| Feature | Description | Priority |
|---------|-------------|----------|
| **MCP Registry** | Daftar MCP servers yang tersedia | P0 |
| **MCP Setup** | Configure MCP yang mau dipakai | P0 |
| **MCP Enable/Disable** | Activate per-project | P1 |
| **MCP Discovery** | Auto-detect MCP dari environment | P2 |

#### MCP Integration

```
Aigo Core
    │
    ├── MCP Client
    │   ├── Exa (web search)
    │   ├── Context7 (docs)
    │   ├── GitHub (code search)
    │   └── Custom MCPs...
    │
    └── Skill System (bawa MCP sendiri)
```

---

### 5. Project-Aware Memory

| Feature | Description | Priority |
|---------|-------------|----------|
| **Project Detection** | Auto-detect project dari working directory | P0 |
| **Project Memory** | Memory spesifik per project | P0 |
| **Project Context** | Tau structure project, dependencies, dll | P0 |
| **Multi-Project** | Isolasi memory antar project | P1 |

#### Project Memory Structure

```
Aigo User
    │
    ├── Project: MyApp
    │   ├── Memory: requirements, architecture, decisions
    │   ├── Context: dependencies.json, package.json, go.mod
    │   └── Code Graph: file relationships
    │
    ├── Project: SideProject
    │   ├── Memory: requirements, progress
    │   ├── Context: tech stack, API docs
    │   └── Code Graph: module structure
    │
    └── Global Memory
        └── User preferences, patterns
```

---

### 6. Code/File Relationship Graph

| Feature | Description | Priority |
|---------|-------------|----------|
| **File Indexing** | Index semua file di project | P0 |
| **Dependency Analysis** | Parse import/require/exports | P0 |
| **Relationship Mapping** | Build graph: file → file relationships | P0 |
| **Code Search** | Semantic search dengan relationship awareness | P1 |
| **Change Impact** | Predict impact dari perubahan | P2 |

#### Graph Example

```
main.go
  ├── imports: config.go, database.go, handlers.go
  ├── called-by: none
  └── related: main_test.go

config.go
  ├── imports: env.go
  └── related: .env.example

database.go
  ├── imports: migrations.go
  └── related: schema.sql
```

---

### 7. Error Memory — Learn from Failures

| Feature | Description | Priority |
|---------|-------------|----------|
| **Error Tracking** | Record semua error yang terjadi | P0 |
| **Error Analysis** | Analisa penyebab error | P0 |
| **Error Prevention** | Jangan ulangi mistake yang sama | P0 |
| **Success Tracking** | Record juga yang berhasil | P0 |

#### Error Memory Flow

```
Task Execution
      │
      ▼
┌──────────────┐
│   Success?   │
└──────┬───────┘
       │
   ┌───┴───┐
   │  Yes  │ No ┌────────────┐
   │       │    │ Record     │
   │       ├───►│ Error +    │
   │       │    │ Context    │
   │       │    └────────────┘
   │       │
   │       ▼
   │  ┌──────────────┐
   │  │ Check Error  │
   │  │ Memory:      │
   │  │ " pernah      │
   │  │ terjadi?"    │
   │  └──────┬───────┘
   │         │
   │    ┌────┴────┐
   │    │  Found  │ Not Found ┌────────────┐
   │    │         │           │ Try Different│
   │    │ Skip    │           │ Approach    │
   │    │ approach│           └────────────┘
   │    └─────────┘
   │
   ▼
Record Success Pattern
```

---

### 8. Full Autonomy — UAT-Style Workflow

| Feature | Description | Priority |
|---------|-------------|----------|
| **Task Planning** | Prometheus planner buat execution plan | P0 |
| **Autonomous Execution** | Sisyphus eksekusi task tanpa interaksi | P0 |
| **Progress Reporting** | Report status berkala (bukan minta input) | P0 |
| **Result Verification** | User verify hasil akhir (UAT) | P0 |
| **Feedback Loop** | Simpan feedback untuk learning | P1 |

#### UAT-Style Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    FULL AUTONOMY WORKFLOW                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  USER                    AIGO                               │
│    │                      │                                 │
│    │  1. Beri task        │                                 │
│    │─────────────────────►│                                 │
│    │                      │                                 │
│    │                      │  2. Prometheus: buat plan        │
│    │                      │─────────────────────────────────►│
│    │                      │                                 │
│    │                      │  3. Sisyphus: execute            │
│    │                      │     (tanpa interaksi)           │
│    │                      │▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓│
│    │                      │                                 │
│    │                      │  4. Progress report              │
│    │                      │◄────────────────────────────────│
│    │                      │                                 │
│    │  5. Verifikasi       │                                 │
│    │◄─────────────────────│                                 │
│    │                      │                                 │
│    │  6. Feedback         │                                 │
│    │─────────────────────►│ ──► Simpan ke memory           │
│    │                      │                                 │
└─────────────────────────────────────────────────────────────┘
```

#### Key Principles

1. **No Micro-Management** — User tidak perlu guiding per step
2. **Self-Correction** — Aigo aware error dan fix sendiri
3. **Transparent Progress** — Regular updates, bukan diam
4. **UAT at End** — Baru butuh user saat akan selesai
5. **Continuous Learning** — Setiap task, Aigo tambah pintar

### P0 - Critical (v0.3.0)
1. Memory Graph and 6-Category Extraction
2. Token Optimization (compression, caching)
3. Planning Phase Before Execution
4. Local-First Configuration

### P1 - High (v0.4.0)
1. Vector Search for Memory
2. Multi-Agent Orchestration
3. Lifecycle Hooks
4. LSP Integration
5. Profile System
6. Workspace Isolation

### P2 - Medium (v0.5.0)
1. Memory Archival
2. Hash-Anchored Editing
3. AST-Grep Support
4. UI Themes/Skins
5. Advanced Autonomy

---

## Task List

### A. Core Integration (OMO + OpenCode) — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| CORE-001 | Bundle OMO engine into Aigo | None | 8 hours |
| CORE-002 | Bundle OpenCode engine into Aigo | None | 8 hours |
| CORE-003 | Integrate Sisyphus loop as default mode | CORE-001, CORE-002 | 4 hours |
| CORE-004 | Activate Prometheus planner by default | CORE-003 | 4 hours |

### B. Setup & Configuration — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| SETUP-001 | Create setup wizard (CLI mode selection) | None | 4 hours |
| SETUP-002 | Implement API key input flow | None | 2 hours |
| SETUP-003 | Local LLM setup (Ollama, LM Studio, dll) | None | 4 hours |
| SETUP-004 | Model configuration (unified) | SETUP-002, SETUP-003 | 3 hours |

### C. Dual Interface — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| UI-001 | Enhance TUI with rich display (Bubble Tea) | None | 4 hours |
| UI-002 | Build Web UI (HTMX + Alpine.js) | None | 6 hours |
| UI-003 | Mode switching (CLI ↔ Web UI) | UI-001, UI-002 | 2 hours |

### D. Gateway System — P1

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| GW-001 | Gateway core (message routing) | None | 4 hours |
| GW-002 | Telegram adapter | GW-001 | 3 hours |
| GW-003 | Slack adapter | GW-001 | 3 hours |
| GW-004 | Discord adapter | GW-001 | 3 hours |

### E. Skill System — P1

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| SKILL-001 | Skill registry and downloader | None | 4 hours |
| SKILL-002 | Skill enable/disable per project | SKILL-001 | 3 hours |
| SKILL-003 | Built-in skills bundle | SKILL-001 | 4 hours |

### F. MCP Configuration — P1

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| MCP-001 | MCP client integration | None | 4 hours |
| MCP-002 | MCP configuration UI | MCP-001 | 3 hours |
| MCP-003 | MCP auto-discovery | MCP-001 | 2 hours |

### G. Memory System — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| MEM-001 | Expand SQLite schema for memory relationships | None | 2 hours |
| MEM-002 | Implement memory extraction from session messages | None | 4 hours |
| MEM-003 | Add container/tag system for memory organization | MEM-001 | 3 hours |
| MEM-004 | Build memory profile generation (static/dynamic) | MEM-002 | 4 hours |
| MEM-005 | Add local vector store (pure Go embeddings) | MEM-001 | 8 hours |
| MEM-006 | Implement similarity search | MEM-005 | 4 hours |
| MEM-007 | Add hybrid search (FTS + vector) | MEM-005, MEM-006 | 3 hours |
| MEM-008 | Implement memory archival for cold memories | MEM-002 | 3 hours |
| MEM-009 | Memory deduplication and merging | MEM-008 | 4 hours |

### H. Project-Aware Memory — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| PROJ-001 | Auto-detect project from working directory | None | 2 hours |
| PROJ-002 | Project-specific memory isolation | MEM-001 | 4 hours |
| PROJ-003 | Project context extraction (deps, configs) | PROJ-001 | 4 hours |
| PROJ-004 | Multi-project support | PROJ-002 | 3 hours |

### I. Code/File Graph — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| GRAPH-001 | File indexing for project | PROJ-001 | 4 hours |
| GRAPH-002 | Dependency analysis (imports/exports) | GRAPH-001 | 4 hours |
| GRAPH-003 | Build relationship graph | GRAPH-002 | 4 hours |
| GRAPH-004 | Relationship-aware search | GRAPH-003 | 3 hours |

### J. Error Memory — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| ERR-001 | Error tracking and logging | None | 3 hours |
| ERR-002 | Error pattern analysis | ERR-001 | 4 hours |
| ERR-003 | Error prevention (check history) | ERR-002 | 3 hours |
| ERR-004 | Success tracking | None | 2 hours |
| ERR-005 | Learning from feedback | None | 3 hours |

### K. Token Optimization — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| TOK-001 | Implement message summarization | None | 4 hours |
| TOK-002 | Add tool output truncation | None | 2 hours |
| TOK-003 | Smart prompt caching | None | 3 hours |
| TOK-004 | Token budget monitoring dashboard | None | 2 hours |
| TOK-005 | Context compression at intervals | None | 3 hours |
| **TOK-006** | **Distillation pipeline (Classifier → Scorer → Composer)** | None | **8 hours** |
| **TOK-007** | **Content classifier (error/success/progress/info/debug)** | TOK-006 | 4 hours |
| **TOK-008** | **Context scorer (+0.4 boost for hot files, recent commands)** | TOK-007 | **4 hours** |
| **TOK-009** | **Signal composer (Critical/Important/Context/Noise tier)** | TOK-007 | **4 hours** |
| **TOK-010** | **RewindStore (SHA-256 content archive)** | TOK-008 | 3 hours |
| **TOK-011** | **Toon format integration (compact output)** | None | 4 hours |
| **TOK-012** | **Semantic log parsing with Toon** | TOK-011 | **3 hours** |
| **TOK-013** | **JSON structure extraction with Toon** | TOK-011 | **3 hours** |
| **TOK-014** | **Omni-style diff for token comparison** | TOK-010 | 2 hours |

> **Target**: 70-90% token reduction while maintaining ≥95% intelligence

### L. Agent System — P0

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| AGENT-001 | Subagent delegation capability | None | 4 hours |
| AGENT-002 | Planning phase before execution | None | 6 hours |
| AGENT-003 | Add lifecycle hooks | None | 4 hours |
| AGENT-004 | Agent registry and routing | AGENT-001 | 4 hours |
| AGENT-005 | Parallel task execution | AGENT-004 | 6 hours |
| AGENT-006 | Self-correction based on errors | ERR-003 | 4 hours |

### M. Coding Capabilities — P1

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| CODE-001 | Expand OpenCode wrapper | None | 3 hours |
| CODE-002 | Add LSP client integration | None | 6 hours |
| CODE-003 | Implement hash-anchored editing | None | 4 hours |
| CODE-004 | Add AST-based code search | CODE-003 | 4 hours |
| CODE-005 | Multi-file refactoring | CODE-004 | 6 hours |

### Token Optimization Tasks

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| TOK-001 | Implement message summarization | None | 4 hours |
| TOK-002 | Add tool output truncation | None | 2 hours |
| TOK-003 | Smart prompt caching | None | 3 hours |
| TOK-004 | Token budget monitoring dashboard | None | 2 hours |
| TOK-005 | Context compression at intervals | None | 3 hours |

### Agent System Tasks

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| AGENT-001 | Subagent delegation capability | None | 4 hours |
| AGENT-002 | Planning phase before execution | None | 6 hours |
| AGENT-003 | Add lifecycle hooks | None | 4 hours |
| AGENT-004 | Agent registry and routing | AGENT-001 | 4 hours |
| AGENT-005 | Parallel task execution | AGENT-004 | 6 hours |
| AGENT-006 | Self-correction based on errors | AGENT-003 | 4 hours |

### Coding Tasks

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| CODE-001 | Expand OpenCode wrapper | None | 3 hours |
| CODE-002 | Add LSP client integration | None | 6 hours |
| CODE-003 | Implement hash-anchored editing | None | 4 hours |
| CODE-004 | Add AST-based code search | CODE-003 | 4 hours |
| CODE-005 | Multi-file refactoring | CODE-004 | 6 hours |

### Architecture Tasks

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| ARCH-001 | Profile system implementation | None | 4 hours |
| ARCH-002 | Workspace isolation | None | 2 hours |
| ARCH-003 | Config hot-reload | ARCH-001 | 3 hours |
| ARCH-004 | Data export/import | None | 3 hours |

### UI/UX Tasks

| Task ID | Description | Dependencies | Estimated Effort |
|---------|-------------|--------------|-------------------|
| UI-001 | Token-driven design tokens | None | 3 hours |
| UI-002 | Theme system implementation | None | 4 hours |
| UI-003 | SSE for real-time updates | None | 4 hours |
| UI-004 | Rich CLI display enhancements | None | 2 hours |

---

## Roadmap Timeline

```
v0.3.0 (Q2 2025) - Foundation Month
├── Memory System (MEM-001 to MEM-004)
├── Token Optimization (TOK-001 to TOK-005)
├── Agent Planning Phase (AGENT-002)
└── Basic Lifecycle Hooks (AGENT-003)

v0.4.0 (Q3 2025) - Intelligence Month
├── Vector Search (MEM-005 to MEM-007)
├── Multi-Agent (AGENT-004 to AGENT-005)
├── LSP Integration (CODE-002)
├── Profile System (ARCH-001)
└── Theme System (UI-002)

v0.5.0 (Q4 2025) - Advanced Features
├── Memory Archival (MEM-008 to MEM-009)
├── Advanced Autonomy (AGENT-006)
├── Hash-Anchored Editing (CODE-003 to CODE-005)
├── Workspace Isolation (ARCH-002)
└── Real-Time UI (UI-003)
```

---

## Feature Flag Configuration

For gradual rollout, implement feature flags:

```yaml
features:
  memory_graph: true
  vector_search: false    # Enable after v0.4.0
  multi_agent: false       # Enable after v0.4.0
  planning_phase: true
  lifecycle_hooks: true
  lsp_integration: false   # Enable after v0.4.0
  hash_anchored_edits: false  # Enable after v0.5.0
  profiles: false         # Enable after v0.4.0
  themes: false          # Enable after v0.5.0
```

---

## Risks and Considerations

### Technical Risks
1. **Vector Search Performance** - Pure Go vector search may be slower than native implementations
2. **Memory Graph Complexity** - May impact performance with large memory sets
3. **Multi-Agent Coordination** - Complex state management

### Design Decisions Needed
1. **Vector Store Backend** - Use pure Go (bag-of-bigrams) or native extension?
2. **Memory Extraction Frequency** - Real-time or batch?
3. **Planning Phase Depth** - Simple or comprehensive?

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Memory retrieval accuracy | >90% relevant |
| Token savings | >40% reduction |
| Planning accuracy | >85% tasks completed as planned |
| Local-first uptime | 100% (no network required) |
| Profile switch time | <500ms |

---

## Conclusion

This roadmap provides a structured approach to evolving Aigo into a **powerful, local-first, autonomous AI agent** by learning from the best features of 8 reference frameworks. The key principles:

1. **Local-first** - No external dependencies except AI models
2. **Token-efficient** - Smart compression and context management  
3. **Autonomous** - Planning phase + self-correction
4. **Modular** - Feature flags for gradual rollout
5. **Extensible** - Plugin architecture for future capabilities

The phased approach allows for incremental improvements while maintaining a stable core system.

---

## Appendix: Complete Feature Analysis from References

### Supermemory — Key Features Applied
- Living Knowledge Graph architecture with semantic relationships (Updates, Extends, Derives)
- Six-stage content pipeline: Queued → Extracting → Chunking → Embedding → Indexing → Done
- Static vs Dynamic memories with versioning
- Vector indexing (internal HNSW-style, no external DB required)
- Hybrid search (semantic + keyword) with metadata filtering
- Container tags for multi-tenant isolation
- User profile generation from memories
- Multi-modal ingestion (PDFs, images, videos, URLs)

### OpenViking — Key Features Applied
- 6-category memory extraction: profile, preferences, entities, events, cases, patterns
- Two-phase session commit: Phase 1 (sync archive) → Phase 2 (async extraction)
- Memory archiver with hotness scoring for cold memory archival
- Context L0/L1/L2 levels (ABSTRACT, OVERVIEW, DETAIL)
- Multi-table persistent/volatile storage backend
- Vectorization integration with URI-based categorization

### Mem0 — Key Features Applied
- Three memory layers: Short-term (working), Long-term (persistent), User Preference
- Graph memory with relationship-based retrieval
- Vector store interfaces (pluggable backends)
- Memory operations: add, search, update, delete with versioning

### Omni — Key Features Applied
- **Learn System**: Word Prefix Frequency algorithm for automatic TOML filter generation
- **TOML Filters**: Project/user/built-in hierarchy with precedence
- **Distillation Pipeline**: Classifier → Scorer → Composer for content filtering
- **RewindStore**: SHA-256 content-addressed archive for raw data retention
- **Signal Lifecycle**: PreHook → PostHook → SessionStart → PreCompact
- **Context Boosting**: Hot files, recent commands, domain hints scoring (+0.4 boost)
- **Cost Analytics**: Token reduction tracking, ROI indicators

### Hermes (Aizen) — Key Features Applied
- Multi-agent orchestration (Aigo CEO, Atlas Architect, Cody Developer, Nova PM, Testa QA)
- 48 lifecycle hooks for fine-grained control
- Parallel agent execution (5+ concurrent agents)
- Hash-anchored edits (zero stale-line editing)
- Skill system with modular workflows
- Prompt caching preservation (no mid-conversation context rebuilding)
- Profile system for multi-instance isolation

### OpenCode — Key Features Applied
- Tool.define framework with zod schema validation
- ReadTool with line-based offset/limit, binary detection, directory listing
- Ripgrep integration for robust content search
- Cross-platform filesystem utilities with mime-type inference
- Effect-based runtime with dependency injection (FileTime service)
- Code generation pipeline for SDK and OpenAPI specs

### OpenWork — Key Features Applied
- Token-driven design system (Foundations → Semantic → Components)
- Three runtime modes: Desktop-hosted, CLI-hosted, Cloud-hosted
- Per-workspace isolation with reload-safe design
- SSE-based real-time event streaming
- OpenWork Router for messaging surface integration

### Oh-My-OpenAgent — Key Features Applied
- Sisyphus orchestrator for task delegation
- Prometheus planner for planning phase before execution
- Hephaestus autonomous deep worker
- Factory pattern for tool/agent extensibility
- Three-layer orchestration: Orchestrator → Planner → Task Router → Handlers
- Hashline (LINE#ID) for surgical, verifiable edits

### Additional Features to Consider (Future Phases)

Based on deep-dive analysis, these features are recommended for future consideration:

| Feature | Source | Rationale |
|---------|--------|-----------|
| **Auto-Learn Pattern Detection** | Omni | Automatically detect and filter repetitive noise |
| **Analytics Dashboard** | Omni, Hermes | Track token usage, cost, ROI metrics |
| **Signal Comparison (omni diff)** | Omni | Compare raw vs distilled output for transparency |
| **Context Budget Discipline** | Oh-My-OpenAgent | Skills bring on-demand MCPs to limit context |
| **Multi-Source Config with Deep Merge** | Oh-My-OpenAgent | JSONC-based multi-level config with Zod validation |

---

## Implementation Priority Matrix

### Must Have (P0) — Next 3 Months
1. **Memory Graph + 6-Category Extraction** — Core differentiator
2. **Token Compression + Caching** — Cost savings
3. **Planning Phase** — Quality assurance
4. **Local-First Config** — Core principle

### Should Have (P1) — Next 6 Months
1. **Vector Search** — Memory retrieval accuracy
2. **Multi-Agent** — Complex task handling
3. **LSP Integration** — Code quality
4. **Profile System** — Multi-tenant

### Nice to Have (P2) — Next 12 Months
1. **Hash-Anchored Edits** — Precision editing
2. **Memory Archival** — Long-running sessions
3. **Theme System** — UI customization
4. **Analytics** — Cost visibility

---

## Detailed Feature Sources

All features extracted from comprehensive analysis of the following reference repositories:

| Reference | Path | Key Strengths Analyzed |
|-----------|------|------------------------|
| Supermemory | `./reference/supermemory-main` | Memory graph, vector search, profiles |
| OpenViking | `./reference/OpenViking-main` | 6-category extraction, session lifecycle |
| Mem0 | `./reference/mem0-main` | Three memory layers, graph memory |
| Omni | `./reference/omni-main` | Token optimization, filters, distillation |
| Hermes | `./reference/hermes` | Autonomous agents, hooks, profiles |
| OpenCode | `./reference/opencode-dev` | Tool framework, LSP, code analysis |
| OpenWork | `./reference/openwork-dev` | Architecture, design system |
| Oh-My-OpenAgent | `./reference/oh-my-openagent-dev` | Orchestration, planning, hash-anchored |

---

## Document Status

- **Created**: 2026-04-02
- **Version**: 1.0
- **Status**: Draft — Ready for review
- **Next Step**: Execute P0 features or refine priorities