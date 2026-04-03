package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type ChildSession struct {
	ID          string
	Description string
	Category    string
	ParentID    string
	Depth       int
	Status      string
	Result      string
	CreatedAt   time.Time
}

type DelegateTool struct {
	sessions map[string]*ChildSession
	maxDepth int
	mu       sync.RWMutex
}

func NewDelegateTool() *DelegateTool {
	return &DelegateTool{
		sessions: make(map[string]*ChildSession),
		maxDepth: 2,
	}
}

func (d *DelegateTool) Name() string        { return "delegate" }
func (d *DelegateTool) Description() string { return "Spawn a child agent with isolated context" }
func (d *DelegateTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"description": map[string]any{"type": "string", "description": "Task description"},
			"category":    map[string]any{"type": "string", "description": "Task category"},
			"session_id":  map[string]any{"type": "string", "description": "Parent session ID"},
			"max_depth":   map[string]any{"type": "integer", "description": "Max depth limit"},
		},
		"required": []string{"description"},
	}
}

func (d *DelegateTool) Execute(ctx context.Context, params map[string]any) (*types.ToolResult, error) {
	descVal, ok := params["description"]
	if !ok {
		return &types.ToolResult{Success: false, Error: "missing required parameter: description"}, nil
	}
	desc, _ := descVal.(string)
	if desc == "" {
		return &types.ToolResult{Success: false, Error: "description is required"}, nil
	}

	category, _ := params["category"].(string)
	parentID, _ := params["session_id"].(string)

	depth := 1
	if parentID != "" {
		if p, err := d.GetChild(parentID); err == nil {
			depth = p.Depth + 1
		}
	}
	if d.maxDepth > 0 && depth > d.maxDepth {
		return &types.ToolResult{Success: false, Error: fmt.Sprintf("depth %d exceeds max depth %d", depth, d.maxDepth)}, nil
	}

	id := d.SpawnChild(parentID, desc, category, depth)
	return &types.ToolResult{
		Success: true,
		Output:  "Child agent spawned: " + id,
		Metadata: map[string]string{
			"child_id": id,
			"status":   "spawned",
		},
	}, nil
}

func (d *DelegateTool) SpawnChild(parentID, description, category string, depth int) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	id := fmt.Sprintf("child-%d-%d", depth, time.Now().UnixNano())
	d.sessions[id] = &ChildSession{
		ID:          id,
		Description: description,
		Category:    category,
		ParentID:    parentID,
		Depth:       depth,
		Status:      "initialized",
		CreatedAt:   time.Now(),
	}
	return id
}

func (d *DelegateTool) GetChild(sessionID string) (*ChildSession, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if s, ok := d.sessions[sessionID]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("child not found: %s", sessionID)
}

func (d *DelegateTool) ListChildren(parentID string) []ChildSession {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out []ChildSession
	for _, s := range d.sessions {
		if s.ParentID == parentID {
			out = append(out, *s)
		}
	}
	return out
}

func (d *DelegateTool) UpdateChild(sessionID, status, result string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if s, ok := d.sessions[sessionID]; ok {
		s.Status = status
		s.Result = result
		return nil
	}
	return fmt.Errorf("child not found: %s", sessionID)
}

func (d *DelegateTool) MaxDepth() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.maxDepth
}
