# Phase 9: Multi-Agent Orchestration

**Goal:** Role-based multi-agent system built on top of OpenCode's existing task() infrastructure.
**Duration:** 2 weeks
**Dependencies:** Phase 2 (Orchestration), OpenCode (OMO)
**Status:** 📋 Planned

**Design Decision:** Aigo sudah menggunakan OpenCode (OMO) sebagai engine multi-agent. OpenCode sudah menyediakan:
- Category-based agent selection (visual-engineering, deep, ultrabrain, dll.)
- Skill injection via load_skills
- Parallel execution via run_in_background=true
- Session isolation dan continuity via session_id

Phase 9 **bukan** membangun agent dari nol, tapi membangun **role configuration layer** yang wrap OpenCode's task() function menjadi role-based agent system yang lebih tinggi levelnya.

---

## 9.1 Role Configuration Layer

### 9.1.1 Role Definitions
- [ ] Create `internal/agents/roles.go` — role registry
- [ ] Define `AgentRole` struct: Name, Category, SystemPrompt, Skills[]
- [ ] Map each role to OpenCode category + skills + system prompt
- [ ] Role validation (category exists, skills available)

### 9.1.2 Role: Aigo (CEO)
- [ ] Category: `ultrabrain`
- [ ] System prompt: decision making, coordination, task decomposition
- [ ] Skills: coordination, planning
- [ ] Final approval before user delivery

### 9.1.3 Role: Atlas (Architect)
- [ ] Category: `deep`
- [ ] System prompt: system design, architecture review, technology selection
- [ ] Skills: architecture, design-patterns
- [ ] Design pattern identification and code structure recommendations

### 9.1.4 Role: Cody (Developer)
- [ ] Category: `deep`
- [ ] System prompt: code implementation, bug fixing, code review
- [ ] Skills: code-review, testing, documentation
- [ ] Test writing and quality checks

### 9.1.5 Role: Nova (PM)
- [ ] Category: `deep`
- [ ] System prompt: requirements analysis, backlog management, estimation
- [ ] Skills: requirements, planning
- [ ] User story creation and progress tracking

### 9.1.6 Role: Testa (QA)
- [ ] Category: `deep`
- [ ] System prompt: test planning, bug identification, regression testing
- [ ] Skills: testing, quality-assurance
- [ ] Quality metrics and reporting

### 9.1.7 Role Executor
- [ ] `ExecuteRole(ctx, role, task) (*Result, error)` — wraps OpenCode task()
- [ ] Auto-inject role's system prompt + skills + category
- [ ] Return structured result with role metadata
- [ ] Fallback to generic agent if role not found

---

## 9.2 Parallel Execution

### 9.2.1 Concurrent Agents
- [ ] 5+ agents running concurrently via OpenCode run_in_background=true
- [ ] Task dependency resolution before parallel start
- [ ] Output aggregation from parallel agents
- [ ] Conflict detection and resolution

### 9.2.2 Task Orchestration
- [ ] DAG-based task execution
- [ ] Dynamic task creation based on results
- [ ] Task cancellation and rollback
- [ ] Progress reporting for parallel tasks
- [ ] Bottleneck detection and optimization

---

## 9.3 Lifecycle Hooks

### 9.3.1 Hook Expansion
- [ ] Expand `internal/hooks/` to 48 lifecycle events
- [ ] Pre-task, post-task, error, success hooks
- [ ] Agent spawn, shutdown, health hooks
- [ ] Message send, receive, timeout hooks
- [ ] File change, tool execution, result hooks

### 9.3.2 Hook Configuration
- [ ] Hook configuration via YAML file
- [ ] Enable/disable hooks per project
- [ ] Hook execution order and priority
- [ ] Hook timeout and retry settings
- [ ] Hook output capture and logging

### 9.3.3 Hook Execution
- [ ] Synchronous and asynchronous hook execution
- [ ] Hook result handling (continue, abort, retry)
- [ ] Hook error handling and fallback
- [ ] Hook performance monitoring
- [ ] Hook dependency management

---

## Phase 9 Checklist

| Category | Tasks | Done | Progress |
|----------|-------|------|----------|
| Role Configuration Layer | 18 | 0 | 0% |
| Parallel Execution | 10 | 0 | 0% |
| Lifecycle Hooks | 15 | 0 | 0% |
| **Total** | **43** | **0** | **0%** |

---

## Architecture

```
User Request
     │
     ▼
┌─────────────┐
│ Aigo (CEO)  │  ← Role: ultrabrain + coordination prompt
│  Decompose  │
└──────┬──────┘
       │
  ┌────┼────┬────────┬────────┐
  ▼    ▼    ▼        ▼        ▼
Atlas Cody Cody   Testa    Nova
(arch)(code)(code)  (QA)    (PM)
  │    │    │        │        │
  └────┴────┴────────┴────────┘
       │
       ▼
┌─────────────┐
│ Aigo (CEO)  │  ← Aggregate results, approve
│  Approve    │
└──────┬──────┘
       │
       ▼
  User Delivery
```

Each role is a thin wrapper around OpenCode's task():

```go
func ExecuteRole(ctx, role, task) {
    r := Roles[role]
    return task(category=r.Category, load_skills=r.Skills, prompt=r.Prompt + task)
}
```
