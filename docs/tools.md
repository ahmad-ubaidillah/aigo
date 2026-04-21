# Tools Reference

Complete reference for all tools available in Aigo.

## Git Tools

### git_status

Get git repository status.

```json
{
  "name": "git_status",
  "description": "Get git repository status (branch, staged, modified, untracked files)",
  "parameters": {}
}
```

### git_diff

View git diff for modified files.

```json
{
  "name": "git_diff",
  "description": "Get git diff for modified files",
  "parameters": {
    "staged": "boolean - show staged diff"
  }
}
```

### git_commit

Commit staged changes.

```json
{
  "name": "git_commit",
  "description": "Commit staged changes with a message",
  "parameters": {
    "message": "string - commit message"
  }
}
```

### git_undo

Undo the last N commits.

```json
{
  "name": "git_undo",
  "description": "Undo the last N commits (soft reset)",
  "parameters": {
    "count": "number - number of commits to undo"
  }
}
```

## Plan Tools

### plan_create

Create a new plan.

```json
{
  "name": "plan_create",
  "description": "Create a new plan with tasks",
  "parameters": {
    "name": "string - plan name",
    "description": "string - plan description"
  }
}
```

### plan_add_task

Add a task to a plan.

```json
{
  "name": "plan_add_task",
  "description": "Add a task to a plan",
  "parameters": {
    "plan_id": "string - plan ID",
    "title": "string - task title",
    "description": "string - task description"
  }
}
```

### plan_list

List all plans.

```json
{
  "name": "plan_list",
  "description": "List all plans",
  "parameters": {}
}
```

### plan_show

Show plan details.

```json
{
  "name": "plan_show",
  "description": "Show plan details and tasks",
  "parameters": {
    "plan_id": "string - plan ID"
  }
}
```

## Diff Sandbox Tools

### sandbox_add

Add a change to review queue.

```json
{
  "name": "sandbox_add",
  "description": "Add a change to diff sandbox for review",
  "parameters": {
    "project": "string - project path",
    "file": "string - file path",
    "old_content": "string - original content",
    "new_content": "string - new content"
  }
}
```

### sandbox_list

List pending changes.

```json
{
  "name": "sandbox_list",
  "description": "List pending changes in sandbox",
  "parameters": {
    "project": "string - project path"
  }
}
```

### sandbox_apply

Apply a change.

```json
{
  "name": "sandbox_apply",
  "description": "Apply a change from sandbox",
  "parameters": {
    "project": "string - project path",
    "change_id": "string - change ID (omit for all)"
  }
}
```

### sandbox_reject

Reject a change.

```json
{
  "name": "sandbox_reject",
  "description": "Reject a change from sandbox",
  "parameters": {
    "project": "string - project path",
    "change_id": "string - change ID (omit for all)"
  }
}
```

## Code Tools

### codex_index

Index project symbols.

```json
{
  "name": "codex_index",
  "description": "Index project symbols and errors",
  "parameters": {
    "project_dir": "string - project directory to index"
  }
}
```

### codex_find_symbol

Find a symbol.

```json
{
  "name": "codex_find_symbol",
  "description": "Find a symbol definition in the project",
  "parameters": {
    "symbol": "string - symbol name to find"
  }
}
```

### codex_map_error

Map error to source.

```json
{
  "name": "codex_map_error",
  "description": "Map an error to its source location",
  "parameters": {
    "error": "string - error message"
  }
}
```

## Memory Tools

### project_context

Get project context.

```json
{
  "name": "project_context",
  "description": "Get project context (type, structure, last edits)",
  "parameters": {}
}
```

### project_add_fact

Add a fact to memory.

```json
{
  "name": "project_add_fact",
  "description": "Add a fact to project memory",
  "parameters": {
    "fact": "string - fact to add"
  }
}
```

## Action Log Tools

### actionlog_list

List recent actions.

```json
{
  "name": "actionlog_list",
  "description": "List recent actions",
  "parameters": {
    "project": "string - project path",
    "limit": "number - max actions"
  }
}
```

### actionlog_undo

Undo the last action.

```json
{
  "name": "actionlog_undo",
  "description": "Undo the last action",
  "parameters": {
    "project": "string - project path"
  }
}
```

## Vision Tools

### vision_encode

Encode image to base64.

```json
{
  "name": "vision_encode",
  "description": "Encode image to base64 for vision models",
  "parameters": {
    "path": "string - image file path or URL"
  }
}
```

### vision_detect_type

Detect image MIME type.

```json
{
  "name": "vision_detect_type",
  "description": "Detect image MIME type from file",
  "parameters": {
    "path": "string - image file path"
  }
}
```

## File Tools

### read

Read file contents.

```json
{
  "name": "read",
  "description": "Read a file or directory",
  "parameters": {
    "file_path": "string - file path",
    "offset": "number - line to start",
    "limit": "number - max lines"
  }
}
```

### write

Write file contents.

```json
{
  "name": "write",
  "description": "Write content to a file",
  "parameters": {
    "file_path": "string - file path",
    "content": "string - content to write"
  }
}
```

### edit

Edit a file.

```json
{
  "name": "edit",
  "description": "Edit a file",
  "parameters": {
    "file_path": "string - file path",
    "old_string": "string - text to replace",
    "new_string": "string - replacement text"
  }
}
```

### glob

Find files by pattern.

```json
{
  "name": "glob",
  "description": "Find files by name pattern",
  "parameters": {
    "pattern": "string - glob pattern"
  }
}
```

### grep

Search in files.

```json
{
  "name": "grep",
  "description": "Search for a pattern in files",
  "parameters": {
    "pattern": "string - regex pattern",
    "path": "string - directory to search",
    "output_mode": "string - content|files_with_matches|count"
  }
}
```

## Shell Tools

### terminal

Execute shell command.

```json
{
  "name": "terminal",
  "description": "Execute a shell command",
  "parameters": {
    "command": "string - command to execute",
    "description": "string - command description",
    "workdir": "string - working directory",
    "timeout": "number - timeout in ms"
  }
}
```

## Web Tools

### web_fetch

Fetch URL content.

```json
{
  "name": "web_fetch",
  "description": "Fetch content from a URL",
  "parameters": {
    "url": "string - URL to fetch",
    "format": "string - markdown|text|html"
  }
}
```

### web_search

Search the web.

```json
{
  "name": "web_search",
  "description": "Search the web",
  "parameters": {
    "query": "string - search query",
    "num_results": "number - number of results"
  }
}
```

## LSP Tools

### lsp_diagnostics

Get LSP diagnostics.

```json
{
  "name": "lsp_diagnostics",
  "description": "Get errors, warnings from language server",
  "parameters": {
    "file_path": "string - file or directory path",
    "severity": "string - error|warning|information|hint|all"
  }
}
```

### lsp_goto_definition

Jump to definition.

```json
{
  "name": "lsp_goto_definition",
  "description": "Jump to symbol definition",
  "parameters": {
    "file_path": "string - file path",
    "line": "number - line number",
    "character": "number - character position"
  }
}
```

---

For more information, see:
- [Features](./features.md)
- [Providers](./providers.md)