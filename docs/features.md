# Features

Aigo comes with a comprehensive set of features for AI-assisted coding.

## Core Features

### 1. Git Integration (Aider-like)

Full Git integration for version control:

| Tool | Description |
|------|-------------|
| `git_status` | Show branch, staged, modified files |
| `git_diff` | View changes |
| `git_commit` | Commit with message |
| `git_commit_auto` | Auto-commit all changes |
| `git_undo` | Undo last N commits |
| `git_log` | Show commit history |
| `git_branch` | Manage branches |

Example:

```
aigo> git status
aigo> git commit -m "Add feature"
aigo> git undo 2
```

### 2. Plan/Task System (Plandex-like)

Multi-step task management:

| Tool | Description |
|------|-------------|
| `plan_create` | Create new plan |
| `plan_add_task` | Add task to plan |
| `plan_list` | List all plans |
| `plan_show` | Show plan details |

Example:

```
aigo> Create a plan to build a REST API
aigo> Add task: create database schema
aigo> Add task: create HTTP handlers
```

### 3. Diff Sandbox

Queue changes for review before applying:

| Tool | Description |
|------|-------------|
| `sandbox_add` | Add change to review queue |
| `sandbox_list` | List pending changes |
| `sandbox_show` | View diff |
| `sandbox_apply` | Apply change(s) |
| `sandbox_reject` | Reject change(s) |

### 4. Code Indexing

Project-wide symbol search and error mapping:

| Tool | Description |
|------|-------------|
| `codex_index` | Index project symbols |
| `codex_find_symbol` | Find symbol definition |
| `codex_map_error` | Map error to source |

### 5. Project Memory

Per-project context memory:

| Tool | Description |
|------|-------------|
| `project_context` | Get project context |
| `project_add_fact` | Add fact to memory |

### 6. Action Log

Full audit trail with undo:

| Tool | Description |
|------|-------------|
| `actionlog_list` | List actions |
| `actionlog_undo` | Undo last action |
| `actionlog_diff` | Get diff |

### 7. Vision Pipeline

Image input for multimodal models:

| Tool | Description |
|------|-------------|
| `vision_encode` | Encode image to base64 |
| `vision_detect_type` | Detect MIME type |

### 8. Smart Routing

Route simple queries to cheap models:

Configuration:

```yaml
smart_routing:
  enabled: true
  max_simple_chars: 160
  cheap_model:
    provider: "openai"
    model: "gpt-4o-mini"
```

### 9. MCP Support

Connect to Model Context Protocol servers:

```yaml
mcp_servers:
  context7:
    url: "https://mcp.context7.com/mcp"
```

### 10. Planning System

| Component | Description |
|-----------|-------------|
| Prometheus | Plans before executing |
| Metis | Identifies ambiguities |
| Momus | Reviews for completeness |
| Resolver | Routes to right tools |

## Memory System

### Pyramid Memory

6-tier context management:
- L0: System prompt
- L1: Project context  
- L2: Session memory
- L3: Historical facts
- L4: Vector embeddings

### Session Management

Create, resume, and manage sessions:

```bash
aigo session create --name "project-x"
aigo session list
aigo session resume <id>
```

## Tools

### File Operations

- `read` - Read file contents
- `write` - Write file
- `edit` - Edit file
- `glob` - Find files by pattern
- `grep` - Search in files

### Shell Operations

- `terminal` - Execute shell commands
- `bash` - Run bash scripts

### Code Intelligence

- `lsp_diagnostics` - Check for errors
- `lsp_goto_definition` - Jump to definition
- `lsp_find_references` - Find references
- `lsp_rename` - Rename symbol

### Web Operations

- `web_fetch` - Fetch URL content
- `web_search` - Search the web

### Browser Automation

- `browser_navigate` - Navigate to URL
- `browser_click` - Click element
- `browser_type` - Type in field

## Channels

### Gateway Integration

| Platform | Status |
|----------|--------|
| CLI | ✅ Default |
| Telegram | ✅ |
| Discord | ✅ |
| Slack | ✅ |
| WhatsApp | ✅ |
| WebSocket | ✅ |

## Self-Healing

### Error Recovery

- Automatic retry on failure
- Error pattern learning
- Context-aware corrections

### Learning System

- Learns from corrections
- Improves over time
- Per-project memory

## Multi-Agent

### Sub-Agents

| Agent | Specialty |
|-------|------------|
| Sisyphus | Orchestration |
| Hephaestus | Implementation |
| Oracle | Debugging |
| Explore | Research |

---

For tool details, see [Tools Reference](./tools.md)

For providers, see [Providers](./providers.md)