# Phase 7: Code Intelligence

**Goal:** Deep code understanding for autonomous development.
**Duration:** 3 weeks
**Dependencies:** Phase 1 (Foundation), Phase 6 (Advanced Memory)
**Status:** 📋 Planned

---

## 7.1 File Relationship Graph

### 7.1.1 Graph Package
- [ ] Create `internal/codegraph/` package structure
- [ ] Define `CodeNode` struct (file, directory, module)
- [ ] Define `CodeEdge` struct (imports, depends, contains)
- [ ] Implement graph data structure with adjacency list
- [ ] SQLite storage for code graph

### 7.1.2 Language Parsers
- [ ] Go import parser (parse `import` blocks)
- [ ] Python import parser (parse `import`, `from ... import`)
- [ ] JavaScript/TypeScript import parser (parse `import`, `require`)
- [ ] Rust module parser (parse `mod`, `use`, `extern crate`)
- [ ] Generic file dependency scanner

### 7.1.3 Graph Operations
- [ ] BuildGraph(rootDir) — scan project and build dependency tree
- [ ] DetectCircularDeps() — find circular import chains
- [ ] FindDependents(file) — find all files that depend on this file
- [ ] FindDependencies(file) — find all files this file depends on
- [ ] VisualizeDOT() — export graph in DOT format
- [ ] VisualizeText() — text-based dependency tree

### 7.1.4 Integration
- [ ] Auto-build graph on project open
- [ ] Incremental update on file changes
- [ ] Query API for agent tools
- [ ] Cache graph for fast access

---

## 7.2 LSP Integration

### 7.2.1 LSP Client
- [ ] Create `internal/lsp/` package
- [ ] Implement LSP protocol (JSON-RPC over stdio)
- [ ] Language server lifecycle (start, stop, restart)
- [ ] Connection management with timeout and retry
- [ ] Supported servers: gopls, pyright, tsserver, rust-analyzer

### 7.2.2 LSP Methods
- [ ] `textDocument/definition` — goto definition
- [ ] `textDocument/references` — find all references
- [ ] `textDocument/hover` — get hover information
- [ ] `textDocument/diagnostic` — get diagnostics/errors
- [ ] `textDocument/rename` — rename symbol across files
- [ ] `textDocument/formatting` — format document
- [ ] `textDocument/completion` — code completion
- [ ] `textDocument/signatureHelp` — function signatures

### 7.2.3 LSP Tool
- [ ] Create `lsp` tool in `internal/tools/`
- [ ] `lsp_definition(file, symbol)` — find definition location
- [ ] `lsp_references(file, symbol)` — find all usages
- [ ] `lsp_diagnostics(file)` — get errors and warnings
- [ ] `lsp_rename(file, old, new)` — rename across project
- [ ] `lsp_format(file)` — format code

### 7.2.4 Diagnostics Integration
- [ ] Auto-collect diagnostics on file save
- [ ] Feed diagnostics to agent context
- [ ] Auto-fix suggestions from diagnostics
- [ ] Track diagnostic trends over time

---

## 7.3 AST-Grep Integration

### 7.3.1 AST-Grep Wrapper
- [ ] Create `internal/astgrep/` package
- [ ] Install/verify ast-grep binary
- [ ] Implement pattern matching wrapper
- [ ] Support multiple languages (Go, Python, JS, TS, Rust)

### 7.3.2 Pattern Search
- [ ] `Search(pattern, language, path)` — find code matches
- [ ] `Replace(pattern, replacement, language, path)` — transform code
- [ ] `Extract(pattern, language, path)` — extract code snippets
- [ ] Support metavariables ($VAR, $$$ARGS)
- [ ] Return structured match results with file, line, code

### 7.3.3 AST-Grep Tool
- [ ] Create `astgrep` tool in `internal/tools/`
- [ ] `astgrep_search(pattern, lang, path)` — search code
- [ ] `astgrep_replace(pattern, replacement, lang, path)` — transform
- [ ] `astgrep_extract(pattern, lang, path)` — extract snippets
- [ ] Integration with agent loop for code analysis

---

## 7.4 Hash-Anchored Editing

### 7.4.1 Content Anchoring
- [ ] SHA-256 hash for file content blocks
- [ ] Block-level change tracking
- [ ] Zero stale-line detection (detect moved/changed lines)
- [ ] Conflict resolution for concurrent edits

### 7.4.2 Multi-File Operations
- [ ] Atomic multi-file refactoring
- [ ] Rollback on partial failure
- [ ] Pre-edit validation (syntax check, type check)
- [ ] Post-edit verification (tests pass, no new errors)

### 7.4.3 Edit Tool Enhancement
- [ ] Enhance existing `edit` tool with hash anchoring
- [ ] Add `edit_multi` tool for multi-file changes
- [ ] Add `refactor` tool for structural changes
- [ ] Add `validate` tool for post-edit checks

---

## Phase 7 Checklist

| Category | Tasks | Done | Progress |
|----------|-------|------|----------|
| File Relationship Graph | 16 | 0 | 0% |
| LSP Integration | 18 | 0 | 0% |
| AST-Grep Integration | 10 | 0 | 0% |
| Hash-Anchored Editing | 10 | 0 | 0% |
| **Total** | **54** | **0** | **0%** |
