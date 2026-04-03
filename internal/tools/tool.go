// Package tools provides the core tool system for Aigo's autonomous agent loop.
package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

// Tool defines the interface for executable tools in the agent system.
type Tool interface {
	// Name returns the unique identifier for this tool.
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// Schema returns the JSON schema for the tool's parameters.
	Schema() map[string]any

	// Execute runs the tool with the given parameters and returns the result.
	Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error)
}

// ToolConfig holds configuration for tool execution.
type ToolConfig struct {
	Name          string
	Description   string
	Timeout       int // milliseconds, 0 means no timeout
	MaxOutputSize int // bytes, 0 means no limit
}

// ToolRegistry manages registered tools and provides execution capabilities.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new empty tool registry.
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry. Returns error if tool with same name exists.
func (r *ToolRegistry) Register(t Tool) error {
	if t == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := t.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %q already registered", name)
	}

	r.tools[name] = t
	return nil
}

// Get retrieves a tool by name. Returns nil if not found.
func (r *ToolRegistry) Get(name string) Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.tools[name]
}

// List returns all registered tools.
func (r *ToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// Execute runs a tool by name with the given parameters.
func (r *ToolRegistry) Execute(ctx context.Context, name string, params map[string]any) (*types.ToolResult, error) {
	r.mu.RLock()
	t, exists := r.tools[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %q not found", name)
	}

	return t.Execute(ctx, params)
}

// OutputTruncate truncates tool output if it exceeds maxSize bytes.
// Appends a truncation notice showing how many bytes were omitted.
func OutputTruncate(result *types.ToolResult, maxSize int) {
	if result == nil || maxSize <= 0 {
		return
	}

	outputLen := len(result.Output)
	if outputLen <= maxSize {
		return
	}

	omitted := outputLen - maxSize
	result.Output = result.Output[:maxSize] + fmt.Sprintf("... (truncated, %d bytes omitted)", omitted)
}
