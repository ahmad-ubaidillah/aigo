# Aigo — Production Ready Checklist

**Goal:** Aigo v1.0 — Production Ready
**Current:** v0.2.0 (architecture complete, many components stub/inactive)
**Build:** `go build -o aigo ./cmd/aigo/`

---

## Priority 1 — Must Fix Before Production

### 1.1 Config Provider is Broken
**Problem:** `config.yaml` uses `kilocode` provider with empty `api_key`. Default model `kilo-auto/free` is non-standard.

**Fix:**
- Set working provider in `~/.aigo/config.yaml` (e.g. `openai`, `openrouter`, or `kilocode` with valid key)
- Or configure `OPENAI_API_KEY` / `ANTHROPIC_API_KEY` env var
- Remove `api_key: ""` if using env vars

### 1.2 Planner LLM Provider Not Wired
**Problem:** `Planner.llmEnabled: false` by default. Rule-based planning is too simple for real tasks.

**Fix in `cmd/aigo/main.go`:**
```go
// After creating agent a:
planner := planning.NewPlanner()
planner.SetLLMProvider(pm) // Wire LLM to planner
a.SetPlanner(planner)
metis := planning.NewMetis()
a.SetMetis(metis)
momus := planning.NewMomus()
a.SetMomus(momus)
```

### 1.3 Autonomy Error Analyzer is Hardcoded
**Problem:** `ErrorAnalyzer` patterns in `autonomy_extra.go` are manually written strings, not learned.

**Fix:** Implement pattern learning:
- Store error → fix mappings in SQLite after successful resolution
- On similar error, suggest previously successful fix first
- Track success rate per pattern

---

## Priority 2 — Core Feature Gaps

### 2.1 Vector Search Not Working
**Problem:** `vectortools` and `vector.New()` are called in main.go but likely fail silently. COMPARISON.md lists this as a red gap.

**Fix:** Either:
- Integrate `sqlite-vec` extension properly, OR
- Use lightweight alternative: `github.com/mmcloughlin/spooky` SimHash + in-memory map
- Add `vector_search(query, top_k)` tool that actually returns results

### 2.2 Semantic Router is Placeholder
**Problem:** `router.New()` is created but `routertools` likely returns stub responses.

**Fix:**
- Implement TF-IDF based intent classification (no ML needed)
- 10-15 intent categories: coding, research, web_search, memory, planning, chat
- <100 LOC, no external dependencies

### 2.3 Sub-Agent Delegation is Stub
**Problem:** `subagent.NewOrchestrator()` exists but `Sisyphus/Hephaestus/Oracle/Explore` agents likely don't work end-to-end.

**Fix:**
- Test `subagenttools` end-to-end
- Each sub-agent type needs its own system prompt
- Add `delegate_task(task_description)` tool that returns sub-agent result

### 2.4 Skill Hub Search/Index Likely Broken
**Problem:** `skillhub.NewOnlineHub("")` may silently fail. `sources.json` references `smithery.ai`, `anthropic/skills`, and `aigo-registry` — last sync Apr 18 but skill system is not used by agent.

**Fix:** Either:
- Make skill hub work: fix `Search()`, `Install()`, `ListInstalled()`
- Or remove skill hub entirely from startup to avoid log noise
- Remove `sources.json` if skill hub is deferred

---

## Priority 3 — Missing Features

### 3.1 WhatsApp Channel
- Port `bailey` patterns to Go OR use `github.com/yesore/gomba`
- Add to `internal/channels/whatsapp/`

### 3.2 Streaming SSE in WebUI
- `RunStream()` exists in agent.go but WebUI may not use it
- Verify: WebUI shows real-time token streaming, not just "loading then result"

### 3.3 MCP Server Mode
- Expose Aigo as MCP server so other agents (Hermes, etc.) can call Aigo tools
- Implement `internal/mcp/server.go` using `github.com/modelcontextprotocol/server`

### 3.4 Prompt Caching
- Add support for `cacheControl` fields when provider supports it (Nous, Anthropic)
- Reduces token cost on long conversations

---

## Priority 4 — Polish

### 4.1 Test Coverage
- Run `go test ./...` — fix failing tests
- Aim for 60%+ coverage on core packages: `agent`, `planning`, `memory`

### 4.2 Version Bump
- Update `const version = "0.2.0"` to `"0.3.0"` or `"1.0.0"` when production-ready

### 4.3 Update COMPARISON.md
- Mark vector search, semantic routing, sub-agents as ✅ when implemented
- Update "Gap" column to reflect reality

### 4.4 Binary Size Optimization
- Currently ~12MB static binary
- Run `go build -ldflags="-s -w"` to strip debug info
- Target: <10MB

---

## Cleanup Log

Files deleted (stale/outdated):
- `aigo-v0.3.4` — old binary
- `ROADMAP.md` — replaced by TASK.md
- `sources.json` — skill sources config (skill hub not functional)
- `TASK.md` (old) — referenced `/mnt/projects/Aigo/` path
- `docs/` — all stale planning docs and old upgrade guides
- `mnt/` — empty playground directory
- `.sisyphus/` — unused planning system artifacts

---

## Quick Verification Commands

```bash
# Build
go build -o aigo ./cmd/aigo/

# Run tests
go test ./...

# Lint
go vet ./...

# Check for dead code
find . -name "*.go" -exec grep -l "TODO\|FIXME\|STUB" {} \;

# Run aigo
./aigo chat
./aigo version

# Provider check (must set API key first)
export OPENAI_API_KEY="sk-..."
./aigo "Hello, what version are you?"
```

---

## Architecture (Current)

```
aigo/
├── cmd/aigo/main.go          # Entry point (chat/start/skills)
├── internal/
│   ├── agent/                # ReAct loop + loop detector + compressor
│   ├── planning/             # Prometheus → Metis → Momus (LLM provider NOT wired)
│   ├── memory/               # FTS5, Pyramid, Engram, Vector (vector search broken)
│   ├── autonomy/             # Self-healing (error patterns hardcoded), autonomous agent
│   ├── multiagent/           # Roundtable (stub)
│   ├── subagent/             # Orchestrator (Sisyphus/Hephaestus/Oracle/Explore stub)
│   ├── skillhub/             # Online hub (likely broken)
│   ├── channels/             # Telegram, Discord, Slack, WebSocket ✅
│   ├── providers/            # 35+ LLM providers ✅
│   ├── tools/                # Registry + annotations ✅
│   ├── webui/                # Web dashboard ✅
│   └── hooks/                # 48 lifecycle hooks ✅
└── config.yaml               # Provider misconfigured
```
