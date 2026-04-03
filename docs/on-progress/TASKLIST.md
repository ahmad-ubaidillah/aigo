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
- [ ] Create `internal/planning/` package
- [ ] Implement Prometheus agent:
  - [ ] Interview mode for requirements gathering
  - [ ] Strategic plan generation
  - [ ] Task breakdown with dependencies
- [ ] Implement Metis agent:
  - [ ] Gap analysis (what Prometheus missed)
  - [ ] Edge case detection
  - [ ] Risk identification
- [ ] Implement Momus agent:
  - [ ] Ruthless plan reviewer
  - [ ] Validation against quality criteria
  - [ ] Approval/rejection with feedback

### 2.2 Execution Layer
- [ ] Create `internal/execution/` package
- [ ] Implement Atlas orchestrator:
  - [ ] Todo-list management
  - [ ] Worker coordination
  - [ ] Wisdom accumulation across tasks
  - [ ] Progress tracking

### 2.3 Worker Agents
- [ ] Create `internal/workers/` package
- [ ] Implement Sisyphus (main orchestrator):
  - [ ] Multi-provider support (Claude, GPT, GLM)
  - [ ] Fallback chains
- [ ] Implement Hephaestus (deep worker):
  - [ ] Autonomous coding agent
  - [ ] Hash-anchored edits (Line#ID)
- [ ] Implement Oracle (architecture consultant):
  - [ ] Read-only analysis
  - [ ] Design recommendations
- [ ] Implement Librarian (documentation search):
  - [ ] Web search integration
  - [ ] OSS documentation lookup
- [ ] Implement Explore (codebase exploration):
  - [ ] Fast grep-based search
  - [ ] Pattern detection

### 2.4 Delegate Tool
- [ ] Create `internal/tools/task.go`
- [ ] Implement child session spawning
- [ ] Add isolated context for children
- [ ] Restrict toolset for children (no delegate, clarify, memory)
- [ ] Add depth limit (MAX_DEPTH=2)
- [ ] Implement progress callback to parent
- [ ] Add batch/parallel execution support

---

## Phase 3: Memory & Learning (3-4 weeks)

### 3.1 Vector Store Integration
- [ ] Create `internal/vectordb/` package
- [ ] Add Chroma client implementation
- [ ] Implement embedding interface:
  - [ ] OpenAI embeddings
  - [ ] Local embeddings (sentence-transformers)
- [ ] Add semantic search for memories
- [ ] Implement similarity threshold filtering

### 3.2 Fact Extraction
- [ ] Create `internal/memory/facts.go`
- [ ] Implement LLM-based fact extraction:
  - [ ] User memory extraction prompt
  - [ ] Agent memory extraction prompt
- [ ] Add memory lifecycle decisions:
  - [ ] ADD — new fact
  - [ ] UPDATE — modify existing
  - [ ] DELETE — remove outdated
  - [ ] NONE — no action needed
- [ ] Store facts in vector DB

### 3.3 Wisdom Accumulation
- [ ] Create `internal/wisdom/` package
- [ ] Implement lesson extraction per task
- [ ] Add pattern recognition for repeated issues
- [ ] Store learnings in persistent memory
- [ ] Inject relevant wisdom in future tasks

### 3.4 Enhanced Context Engine
- [ ] Extend L0/L1/L2 with vector search
- [ ] Add auto-compression with summarization
- [ ] Implement context prioritization by relevance
- [ ] Add cross-session memory retrieval

---

## Phase 4: Integration (3-4 weeks)

### 4.1 Gateway Hooks
- [ ] Create `internal/hooks/` package
- [ ] Implement hook discovery from `~/.aigo/hooks/`
- [ ] Add hook types:
  - [ ] `gateway:startup` — on gateway start
  - [ ] `session:start` — new session
  - [ ] `session:end` — session complete
  - [ ] `agent:start` — agent begins task
  - [ ] `agent:step` — after tool use
  - [ ] `agent:end` — agent completes
  - [ ] `command:*` — slash commands
- [ ] Add HOOK.yaml schema + handler.go pattern

### 4.2 Skills System Expansion
- [ ] Create `internal/skills/loader.go`
- [ ] Implement skill discovery from `~/.aigo/skills/`
- [ ] Add YAML frontmatter parsing
- [ ] Support skill structure:
  - [ ] `SKILL.md` — main instructions
  - [ ] `references/` — supporting docs
  - [ ] `templates/` — output templates
  - [ ] `scripts/` — executable scripts
- [ ] Add skill CLI commands:
  - [ ] `aigo skill list`
  - [ ] `aigo skill view <name>`
  - [ ] `aigo skill create <name>`
  - [ ] `aigo skill run <name>`

### 4.3 Cron Scheduler
- [ ] Create `internal/cron/` package
- [ ] Implement persistent job storage (`~/.aigo/cron/jobs.json`)
- [ ] Add schedule types:
  - [ ] Once (timestamp)
  - [ ] Interval (every 30m)
  - [ ] Cron expression (0 9 * * *)
- [ ] Add job delivery targets:
  - [ ] Origin chat
  - [ ] Specific platform (telegram, discord)
- [ ] Add file-based lock for concurrent prevention
- [ ] Implement skill attachment to jobs

### 4.4 Permission System
- [ ] Create `internal/permission/` package
- [ ] Define permission types:
  - [ ] bash, read, write, edit
  - [ ] task, webfetch, websearch
  - [ ] doom_loop (for loop detection)
- [ ] Implement ruleset:
  - [ ] `allow` — proceed without asking
  - [ ] `deny` — block action
  - [ ] `ask` — prompt user
- [ ] Add wildcard pattern matching
- [ ] Store permissions in config

---

## Phase 5: Polish (2-3 weeks)

### 5.1 Web GUI
- [ ] Create `web/` templates
- [ ] Implement dashboard views:
  - [ ] Session list
  - [ ] Active task progress
  - [ ] Memory browser
  - [ ] Skills library
  - [ ] Settings panel
- [ ] Add HTMX interactions
- [ ] Add Alpine.js components
- [ ] Implement WebSocket for real-time updates

### 5.2 TUI Improvements
- [ ] Add split-pane layout
- [ ] Implement task progress visualization
- [ ] Add memory browser panel
- [ ] Add skill quick-access menu
- [ ] Implement keyboard shortcuts

### 5.3 Documentation
- [ ] Update README.md with new features
- [ ] Create `docs/architecture.md`
- [ ] Create `docs/tools.md` — tool reference
- [ ] Create `docs/skills.md` — skill development guide
- [ ] Create `docs/hooks.md` — hook system guide
- [ ] Add inline code comments for public APIs

### 5.4 Testing & Benchmarks
- [ ] Add unit tests for tools (>80% coverage)
- [ ] Add integration tests for agent loop
- [ ] Add benchmark for distillation pipeline
- [ ] Add benchmark for context engine
- [ ] Test all gateway adapters
- [ ] Load test cron scheduler

---

## Quick Wins (Can do anytime)

- [ ] Add `aigo doctor` command for diagnostics
- [ ] Add `--verbose` flag for debug output
- [ ] Add colored output with lipgloss
- [ ] Add shell completion generation
- [ ] Add config migration script
- [ ] Add example `.aigo/config.yaml`
- [ ] Create GitHub Actions CI/CD workflow

---

## Blocked / Future

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

|| Phase | Tasks | Completed | Progress |
|-------|-------|-----------|----------|
|| 1. Foundation | 32 | 22 | 69% |
|| 2. Orchestration | 24 | 0 | 0% |
|| 3. Memory & Learning | 18 | 0 | 0% |
|| 4. Integration | 22 | 0 | 0% |
|| 5. Polish | 20 | 0 | 0% |
|| **Total** | **116** | **22** | **19% |

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
