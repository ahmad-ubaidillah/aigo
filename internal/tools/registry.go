// Package tools implements the tool registry and built-in tools.
// Inspired by PicoClaw's ToolRegistry pattern and KrillClaw's annotations.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
)

// Tool is the interface all tools must implement.
type Tool interface {
	Name() string
	Description() string
	Annotations() Annotations
	Schema() ToolSchema
	Execute(ctx context.Context, args map[string]interface{}) (string, error)
}

// Annotations provides safety metadata (ported from KrillClaw).
type Annotations struct {
	Destructive bool     `json:"destructive"`
	ReadOnly    bool     `json:"readOnly"`
	SideEffects []string `json:"sideEffects"` // "filesystem", "network", "system"
}

// ToolSchema follows OpenAI function calling format.
type ToolSchema struct {
	Type     string              `json:"type"`
	Function ToolFunctionSchema  `json:"function"`
}

// ToolFunctionSchema defines a tool's parameters.
type ToolFunctionSchema struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// ToolEntry wraps a tool with metadata.
type ToolEntry struct {
	Tool    Tool
	Version uint64
}

// Registry is the central tool registry (copy-on-write snapshot from ClawGo).
type Registry struct {
	tools   map[string]*ToolEntry
	mu      sync.RWMutex
	version atomic.Uint64
	snapshot atomic.Value // map[string]Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]*ToolEntry)}
	r.snapshot.Store(map[string]Tool{})
	return r
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := tool.Name()
	r.tools[name] = &ToolEntry{
		Tool:    tool,
		Version: r.version.Add(1),
	}
	// Copy-on-write snapshot
	next := make(map[string]Tool, len(r.tools))
	for k, v := range r.tools {
		next[k] = v.Tool
	}
	r.snapshot.Store(next)
}

// Get returns a tool by name (lock-free via snapshot).
func (r *Registry) Get(name string) (Tool, bool) {
	snap, _ := r.snapshot.Load().(map[string]Tool)
	t, ok := snap[name]
	return t, ok
}

// Execute runs a tool by name.
func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return tool.Execute(ctx, args)
}

// Schemas returns all tool schemas for the LLM.
func (r *Registry) Schemas() []ToolSchema {
	snap, _ := r.snapshot.Load().(map[string]Tool)
	schemas := make([]ToolSchema, 0, len(snap))
	for _, tool := range snap {
		schemas = append(schemas, tool.Schema())
	}
	return schemas
}

// List returns all tool names.
func (r *Registry) List() []string {
	snap, _ := r.snapshot.Load().(map[string]Tool)
	names := make([]string, 0, len(snap))
	for name := range snap {
		names = append(names, name)
	}
	return names
}

// ListTools returns all registered tools.
func (r *Registry) ListTools() []Tool {
	snap, _ := r.snapshot.Load().(map[string]Tool)
	tools := make([]Tool, 0, len(snap))
	for _, t := range snap {
		tools = append(tools, t)
	}
	return tools
}

// Count returns the number of registered tools.
func (r *Registry) Count() int {
	snap, _ := r.snapshot.Load().(map[string]Tool)
	return len(snap)
}

// --- Built-in Tools ---

// TerminalTool executes shell commands.
type TerminalTool struct{}

func (t *TerminalTool) Name() string { return "terminal" }
func (t *TerminalTool) Description() string {
	return "Execute a shell command and return stdout/stderr."
}
func (t *TerminalTool) Annotations() Annotations {
	return Annotations{Destructive: true, ReadOnly: false, SideEffects: []string{"filesystem", "network", "system"}}
}
func (t *TerminalTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        "terminal",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]string{
						"type":        "string",
						"description": "The shell command to execute",
					},
				},
				"required": []string{"command"},
			},
		},
	}
}
func (t *TerminalTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	cmd, _ := args["command"].(string)
	if cmd == "" {
		return "", fmt.Errorf("command is required")
	}
	// Import from standard library
	output, err := execCommand(ctx, cmd)
	if err != nil {
		return fmt.Sprintf("Error: %v\n%s", err, output), nil
	}
	return output, nil
}

// ReadFileTool reads a file's contents.
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string { return "read_file" }
func (t *ReadFileTool) Description() string {
	return "Read the contents of a file at the given path."
}
func (t *ReadFileTool) Annotations() Annotations {
	return Annotations{Destructive: false, ReadOnly: true, SideEffects: []string{}}
}
func (t *ReadFileTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        "read_file",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]string{
						"type":        "string",
						"description": "Absolute path to the file to read",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}
func (t *ReadFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	return readFile(path)
}

// WriteFileTool writes content to a file.
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string { return "write_file" }
func (t *WriteFileTool) Description() string {
	return "Write content to a file, creating it if it doesn't exist."
}
func (t *WriteFileTool) Annotations() Annotations {
	return Annotations{Destructive: true, ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *WriteFileTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        "write_file",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]string{"type": "string", "description": "File path"},
					"content": map[string]string{"type": "string", "description": "Content to write"},
				},
				"required": []string{"path", "content"},
			},
		},
	}
}
func (t *WriteFileTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if path == "" || content == "" {
		return "", fmt.Errorf("path and content are required")
	}
	return writeFile(path, content)
}

// SearchFilesTool searches for text in files.
type SearchFilesTool struct{}

func (t *SearchFilesTool) Name() string { return "search_files" }
func (t *SearchFilesTool) Description() string {
	return "Search for a text pattern in files. Returns matching lines with file paths and line numbers."
}
func (t *SearchFilesTool) Annotations() Annotations {
	return Annotations{Destructive: false, ReadOnly: true, SideEffects: []string{}}
}
func (t *SearchFilesTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        "search_files",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]string{"type": "string", "description": "Text pattern to search for"},
					"path":    map[string]string{"type": "string", "description": "Directory or file to search in"},
				},
				"required": []string{"pattern"},
			},
		},
	}
}
func (t *SearchFilesTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	pattern, _ := args["pattern"].(string)
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}
	return searchFiles(pattern, path)
}

// KVTool provides key-value storage.
type KVTool struct {
	storagePath string
}

func NewKVTool(storagePath string) *KVTool {
	return &KVTool{storagePath: storagePath}
}
func (t *KVTool) Name() string { return "kv" }
func (t *KVTool) Description() string {
	return "Lightweight persistent key-value store. Actions: get, set, list, delete."
}
func (t *KVTool) Annotations() Annotations {
	return Annotations{Destructive: false, ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *KVTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        "kv",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"action": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"get", "set", "list", "delete"},
						"description": "Action to perform",
					},
					"key":   map[string]string{"type": "string", "description": "Key name"},
					"value": map[string]string{"type": "string", "description": "Value to store"},
				},
				"required": []string{"action"},
			},
		},
	}
}
func (t *KVTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)
	return kvExecute(t.storagePath, action, key, value)
}

// GetCurrentTimeTool returns the current time.
type GetCurrentTimeTool struct{}

func (t *GetCurrentTimeTool) Name() string { return "get_current_time" }
func (t *GetCurrentTimeTool) Description() string {
	return "Get the current date and time in ISO-8601 format."
}
func (t *GetCurrentTimeTool) Annotations() Annotations {
	return Annotations{Destructive: false, ReadOnly: true, SideEffects: []string{}}
}
func (t *GetCurrentTimeTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "function",
		Function: ToolFunctionSchema{
			Name:        "get_current_time",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *GetCurrentTimeTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	b, err := json.Marshal(map[string]string{"time": currentTimeISO()})
	if err != nil {
		return "", err
	}
	return string(b), nil
}
