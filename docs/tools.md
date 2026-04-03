# Tools Reference

Aigo provides 12 built-in tools for the agent to use.

## Core Tools

### bash
Execute shell commands.
- **Params**: `command` (string, required), `timeout` (int, optional, default 30s)
- **Output**: Combined stdout/stderr, truncated to 100KB

### read
Read file contents.
- **Params**: `path` (string, required)
- **Output**: File content, max 50KB

### write
Write content to file. Creates parent directories if needed.
- **Params**: `path` (string, required), `content` (string, required)

### edit
Replace first occurrence of a string in a file. Fails if old_string not found or appears multiple times.
- **Params**: `path`, `old_string`, `new_string` (all required)

### glob
Find files matching a glob pattern.
- **Params**: `pattern` (string, required)

### grep
Search files for regex patterns. Searches .go, .ts, .js, .py, .md, .txt, .yaml, .yml, .json files. Max 100 results.
- **Params**: `pattern` (string, required), `path` (string, optional, defaults to ".")

## Agent Tools

### task
Spawn a subagent task.
- **Params**: `description` (string, required), `category` (string, optional), `session_id` (string, optional)

### delegate
Spawn a child agent with isolated context.
- **Params**: `description` (string, required), `category`, `session_id`, `max_depth` (int, default 2)

## Web Tools

### webfetch
Fetch content from a URL. Strips HTML tags. Max 50KB.
- **Params**: `url` (string, required)

### websearch
Search the web (stub implementation).
- **Params**: `query` (string, required), `num_results` (int, optional, default 5)

## Utility Tools

### todo
Manage a todo list in memory.
- **Params**: `action` (add|list|complete|cancel), `content` (for add), `index` (for complete/cancel)

## Permission System

Tools can be configured with three permission levels:
- **allow** — execute without asking
- **ask** — prompt user before execution
- **deny** — block execution

Wildcard patterns are supported (e.g., `web*` matches `webfetch` and `websearch`).
