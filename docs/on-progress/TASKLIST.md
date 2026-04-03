# Aigo Task List

> Generated from ROADMAP.md — Track progress toward autonomous AI agent platform

---

## Phase 1: Foundation (4-6 weeks)

### 1.1 Tool System
- [x] Create `internal/tools/` package structure
- [x] Define `Tool` interface with schema validation
- [x] Implement tool registry (singleton pattern)
- [x] Add built-in tools:
  - [x] `bash` — execute shell commands
  - [x] `read` — read file contents
  - [x] `write` — write file contents
  - [x] `edit` — patch files with fuzzy matching
  - [x] `glob` — find files by pattern
  - [x] `grep` — search file contents
  - [x] `task` — spawn subagent
  - [x] `webfetch` — fetch URLs
  - [x] `websearch` — DuckDuckGo search (HTML scraping, no API)
  - [x] `skill` — load/execute skills from ~/.hermes/skills/
  - [x] `todo` — task management
- [x] Add permission system (allow/deny/ask per tool)
- [x] Add output truncation for large outputs

### 1.2 Autonomous Agent Loop
- [x] Create `internal/agent/loop.go`
- [x] Implement streaming LLM response handling
- [x] Add tool execution with context
- [x] Implement doom loop detection (repeated identical calls)
- [x] Add iteration budget tracking
- [x] Implement finish/stop conditions
- [x] Add compaction trigger when context exceeds limit
- [x] Add LLM integration for tool selection (internal/agent/llm_prompt.go)

### 1.2.1 LLM Integration (NEW)
- [x] Create `internal/llm/client.go` with LLMClient interface
- [x] Create `internal/llm/glm.go` for GLM (z.ai) API
- [x] Create `internal/llm/local.go` for local llama.cpp server
- [x] Add provider factory (NewClient) in client.go
- [x] Add LLMConfig to pkg/types/types.go
- [x] Integrate LLMClient into AgentLoop

### 1.3 Session State Tracking
- [x] Extend `ContextEngine.SessionState`:
  - [x] `HotFiles` with access count (map[string]int)
  - [x] `ActiveErrors` (last 5 errors)
  - [x] `LastCommands` (last 20 commands)
  - [x] `InferredTask` (auto-detected from patterns)
  - [x] `InferredDomain` (project type detection)
- [x] Add context boost scoring (hot files +0.1, errors +0.25)

### 1.4 Distillation Pipeline (OMNI-style)
- [x] Create `internal/distill/` package
- [x] Implement Classifier (content type detection):
  - [x] GitDiff, BuildOutput, TestOutput
  - [x] InfraOutput, LogOutput, TabularData
  - [x] StructuredData, Unknown
- [x] Implement Scorer (signal tier assignment):
  - [x] Critical (0.9), Important (0.7), Context (0.4), Noise (0.05)
- [x] Implement Collapse (repetitive line compression)
- [x] Implement Composer (filter + RewindStore archive)
- [x] Add token savings metric calculation

### 1.5 RewindStore
- [x] Create `internal/rewind/store.go`
- [x] Implement SHA-256 hash storage (8-char short hash)
- [x] Add retrieval API (`rewind retrieve <hash>`)
- [x] Add CLI commands: `rewind list`, `rewind show`
- [x] Add SQLite persistence (`persist.go`)
- [x] Integrate with distillation pipeline

---

## Phase 2: Orchestration (4-6 weeks)

### 2.1 Planning Layer
- [x] Create `internal/planning/` package
- [x] Implement Prometheus agent:
  - [x] Interview mode for requirements gathering
  - [x] Strategic plan generation
  - [x] Task breakdown with dependencies
- [x] Implement Metis agent:
  - [x] Gap analysis (what Prometheus missed)
  - [x] Edge case detection
  - [x] Risk identification
- [x] Implement Momus agent:
  - [x] Ruthless plan reviewer
  - [x] Validation against quality criteria
  - [x] Approval/rejection with feedback

### 2.2 Execution Layer
- [x] Create `internal/execution/` package
- [x] Implement Atlas orchestrator:
  - [x] Todo-list management
  - [x] Worker coordination
  - [x] Wisdom accumulation across tasks
  - [x] Progress tracking

### 2.3 Worker Agents
- [x] Create `internal/workers/` package
- [x] Implement Sisyphus (main orchestrator):
  - [x] Multi-provider support (Claude, GPT, GLM)
  - [x] Fallback chains
- [x] Implement Hephaestus (deep worker):
  - [x] Autonomous coding agent
  - [x] Hash-anchored edits (Line#ID)
- [x] Implement Oracle (architecture consultant):
  - [x] Read-only analysis
  - [x] Design recommendations
- [x] Implement Librarian (documentation search):
  - [x] Web search integration
  - [x] OSS documentation lookup
- [x] Implement Explore (codebase exploration):
  - [x] Fast grep-based search
  - [x] Pattern detection

### 2.4 Delegate Tool
- [x] Create `internal/tools/task.go`
- [x] Implement child session spawning
- [x] Add isolated context for children
- [x] Restrict toolset for children (no delegate, clarify, memory)
- [x] Add depth limit (MAX_DEPTH=2)
- [x] Implement progress callback to parent
- [x] Add batch/parallel execution support

---

## Phase 3: Memory & Learning (3-4 weeks)

### 3.1 Vector Store Integration
- [x] Create `internal/vectordb/` package
- [x] Add Chroma client implementation
- [x] Implement embedding interface:
  - [x] OpenAI embeddings
  - [x] Local embeddings (sentence-transformers)
- [x] Add semantic search for memories
- [x] Implement similarity threshold filtering

### 3.2 Fact Extraction
- [x] Create `internal/memory/facts.go`
- [x] Implement LLM-based fact extraction:
  - [x] User memory extraction prompt
  - [x] Agent memory extraction prompt
- [x] Add memory lifecycle decisions:
  - [x] ADD — new fact
  - [x] UPDATE — modify existing
  - [x] DELETE — remove outdated
  - [x] NONE — no action needed
- [x] Store facts in vector DB

### 3.3 Wisdom Accumulation
- [x] Create `internal/wisdom/` package
- [x] Implement lesson extraction per task
- [x] Add pattern recognition for repeated issues
- [x] Store learnings in persistent memory
- [x] Inject relevant wisdom in future tasks

### 3.4 Enhanced Context Engine
- [x] Extend L0/L1/L2 with vector search
- [x] Add auto-compression with summarization
- [x] Implement context prioritization by relevance
- [x] Add cross-session memory retrieval

---

## Phase 4: Integration (3-4 weeks)

### 4.1 Gateway Hooks
- [x] Create `internal/hooks/` package
- [x] Implement hook discovery from `~/.aigo/hooks/`
- [x] Add hook types:
  - [x] `gateway:startup` — on gateway start
  - [x] `session:start` — new session
  - [x] `session:end` — session complete
  - [x] `agent:start` — agent begins task
  - [x] `agent:step` — after tool use
  - [x] `agent:end` — agent completes
  - [x] `command:*` — slash commands
- [x] Add HOOK.yaml schema + handler.go pattern

### 4.2 Skills System Expansion
- [x] Create `internal/skills/loader.go`
- [x] Implement skill discovery from `~/.aigo/skills/`
- [x] Add YAML frontmatter parsing
- [x] Support skill structure:
  - [x] `SKILL.md` — main instructions
  - [x] `references/` — supporting docs
  - [x] `templates/` — output templates
  - [x] `scripts/` — executable scripts
- [x] Add skill CLI commands:
  - [x] `aigo skill list`
  - [x] `aigo skill view <name>`
  - [x] `aigo skill create <name>`
  - [x] `aigo skill run <name>`

### 4.3 Cron Scheduler
- [x] Create `internal/cron/` package
- [x] Implement persistent job storage (`~/.aigo/cron/jobs.json`)
- [x] Add schedule types:
  - [x] Once (timestamp)
  - [x] Interval (every 30m)
  - [x] Cron expression (0 9 * * *)
- [x] Add job delivery targets:
  - [x] Origin chat
  - [x] Specific platform (telegram, discord)
- [x] Add file-based lock for concurrent prevention
- [x] Implement skill attachment to jobs

### 4.4 Permission System
- [x] Create `internal/permission/` package
- [x] Define permission types:
  - [x] bash, read, write, edit
  - [x] task, webfetch, websearch
  - [x] doom_loop (for loop detection)
- [x] Implement ruleset:
  - [x] `allow` — proceed without asking
  - [x] `deny` — block action
  - [x] `ask` — prompt user
- [x] Add wildcard pattern matching
- [x] Store permissions in config

---

## Phase 5: Polish (2-3 weeks)

### 5.1 Web GUI
- [x] Create `web/` templates
- [x] Implement dashboard views:
  - [x] Session list
  - [x] Active task progress
  - [x] Memory browser
  - [x] Skills library
  - [x] Settings panel
- [x] Add HTMX interactions
- [x] Add Alpine.js components
- [x] Implement WebSocket for real-time updates

### 5.2 TUI Improvements
- [x] Add split-pane layout
- [x] Implement task progress visualization
- [x] Add memory browser panel
- [x] Add skill quick-access menu
- [x] Implement keyboard shortcuts

### 5.3 Documentation
- [x] Update README.md with new features
- [x] Create `docs/architecture.md`
- [x] Create `docs/tools.md` — tool reference
- [x] Create `docs/skills.md` — skill development guide
- [x] Create `docs/hooks.md` — hook system guide
- [x] Add inline code comments for public APIs

### 5.4 Testing & Benchmarks
- [x] Add unit tests for tools (>80% coverage)
- [x] Add integration tests for agent loop
- [x] Add benchmark for distillation pipeline
- [x] Add benchmark for context engine
- [x] Test all gateway adapters
- [x] Load test cron scheduler

---

## Quick Wins (Can do anytime)

- [x] Add `aigo doctor` command for diagnostics
- [x] Add `--verbose` flag for debug output
- [x] Add colored output with lipgloss
- [x] Add shell completion generation
- [x] Add config migration script
- [x] Add example `.aigo/config.yaml`
- [x] Create GitHub Actions CI/CD workflow

---

## Future Enhancements (Planned)

> These are planned for future versions (v2.0+)

- [ ] Mobile responsive Web GUI
- [ ] Voice control integration
- [ ] Plugin system for external tools
- [ ] Multi-language support (i18n)
- [ ] Admin UI for user management
- [ ] Rate limiting
- [ ] Redis caching layer
- [ ] Docker images
- [ ] Homebrew tap
- [ ] Debian/RPM packages

---

## Progress Tracking

| Phase | Tasks | Completed | Progress |
|-------|-------|-----------|----------|
| 1. Foundation | 32 | 32 | 100% |
| 2. Orchestration | 24 | 24 | 100% |
| 3. Memory & Learning | 18 | 18 | 100% |
| 4. Integration | 22 | 22 | 100% |
| 5. Polish | 20 | 20 | 100% |
| Quick Wins | 7 | 7 | 100% |
| **Total** | **123** | **123** | **100%** |

### ✅ All Core Tasks Complete!

---

## How to Use This Task List

1. Mark tasks as you complete them: `- [x]` instead of `- [ ]`
2. Update progress tracking table
3. Add new tasks as discovered
4. Move blocked tasks to "Blocked / Future" section
5. Review weekly, adjust priorities

```bash
# Count remaining tasks
grep -c "^\- \[ \]" TASKLIST.md

# Count completed tasks
grep -c "^\- \[x\]" TASKLIST.md
```
