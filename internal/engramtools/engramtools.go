// Package engramtools registers engram memory tools in Aigo's tool registry.
// These tools expose Engram's structured memory (observations, search,
// context, timeline, session lifecycle) to the agent.
package engramtools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hermes-v2/aigo/internal/memory/engram"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterEngramTools registers all engram memory tools in the registry.
func RegisterEngramTools(reg *tools.Registry, b *engram.Backend) {
	reg.Register(&MemSaveTool{b: b})
	reg.Register(&MemSearchTool{b: b})
	reg.Register(&MemContextTool{b: b})
	reg.Register(&MemTimelineTool{b: b})
	reg.Register(&MemStatsTool{b: b})
	reg.Register(&MemSessionStartTool{b: b})
	reg.Register(&MemSessionEndTool{b: b})
}

// --- mem_save ---

type MemSaveTool struct {
	b *engram.Backend
}

func (t *MemSaveTool) Name() string { return "mem_save" }
func (t *MemSaveTool) Description() string {
	return "Save a structured observation to long-term memory. Use for decisions, learnings, preferences, and discoveries."
}
func (t *MemSaveTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"database"}}
}
func (t *MemSaveTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_save",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type": map[string]string{
						"type":        "string",
						"description": "Observation type: decision, learning, preference, discovery, correction, fact",
					},
					"title": map[string]string{
						"type":        "string",
						"description": "Short title for the observation",
					},
					"content": map[string]string{
						"type":        "string",
						"description": "Full observation content",
					},
					"topic_key": map[string]string{
						"type":        "string",
						"description": "Optional topic key for deduplication (e.g. 'project/auth', 'user/preference'). Leave empty for one-off notes.",
					},
				},
				"required": []string{"type", "title", "content"},
			},
		},
	}
}
func (t *MemSaveTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	obsType, _ := args["type"].(string)
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	topicKey, _ := args["topic_key"].(string)

	if obsType == "" || title == "" || content == "" {
		return "", fmt.Errorf("type, title, and content are required")
	}

	id, err := t.b.SaveObservation(obsType, title, content, topicKey)
	if err != nil {
		return "", fmt.Errorf("save observation: %w", err)
	}

	b, _ := json.Marshal(map[string]interface{}{
		"ok":        true,
		"id":        id,
		"type":      obsType,
		"title":     title,
		"topic_key": topicKey,
	})
	return string(b), nil
}

// --- mem_search ---

type MemSearchTool struct {
	b *engram.Backend
}

func (t *MemSearchTool) Name() string { return "mem_search" }
func (t *MemSearchTool) Description() string {
	return "Full-text search across all memories. Returns ranked results with type, title, content, and relevance score."
}
func (t *MemSearchTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true, SideEffects: []string{}}
}
func (t *MemSearchTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_search",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]string{
						"type":        "string",
						"description": "Search query text",
					},
					"limit": map[string]string{
						"type":        "string",
						"description": "Max results to return (default 10)",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}
func (t *MemSearchTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	limitStr, _ := args["limit"].(string)
	limit := 10
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = n
		}
	}

	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	results, err := t.b.Search(query, limit)
	if err != nil {
		return "", fmt.Errorf("search: %w", err)
	}

	if len(results) == 0 {
		return "No memories found matching your query.", nil
	}

	var out []map[string]interface{}
	for _, r := range results {
		content := r.Content
		if len(content) > 300 {
			content = content[:300] + "..."
		}
		out = append(out, map[string]interface{}{
			"id":      r.ID,
			"type":    r.Type,
			"title":   r.Title,
			"content": content,
			"rank":    r.Rank,
		})
	}

	b, _ := json.Marshal(out)
	return string(b), nil
}

// --- mem_context ---

type MemContextTool struct {
	b *engram.Backend
}

func (t *MemContextTool) Name() string { return "mem_context" }
func (t *MemContextTool) Description() string {
	return "Get memory context from previous sessions. Shows recent sessions, prompts, and observations."
}
func (t *MemContextTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true, SideEffects: []string{}}
}
func (t *MemContextTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_context",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *MemContextTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	context, err := t.b.GetContext()
	if err != nil {
		return "", fmt.Errorf("get context: %w", err)
	}
	if context == "" {
		return "No previous session context available.", nil
	}
	return context, nil
}

// --- mem_timeline ---

type MemTimelineTool struct {
	b *engram.Backend
}

func (t *MemTimelineTool) Name() string { return "mem_timeline" }
func (t *MemTimelineTool) Description() string {
	return "Get chronological context around a specific observation. Shows what happened before and after."
}
func (t *MemTimelineTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true, SideEffects: []string{}}
}
func (t *MemTimelineTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_timeline",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"observation_id": map[string]string{
						"type":        "string",
						"description": "Observation ID to get timeline around",
					},
					"before": map[string]string{
						"type":        "string",
						"description": "Number of observations before (default 5)",
					},
					"after": map[string]string{
						"type":        "string",
						"description": "Number of observations after (default 5)",
					},
				},
				"required": []string{"observation_id"},
			},
		},
	}
}
func (t *MemTimelineTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	idStr, _ := args["observation_id"].(string)
	if idStr == "" {
		return "", fmt.Errorf("observation_id is required")
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid observation_id: %w", err)
	}

	before := 5
	after := 5
	if b, _ := args["before"].(string); b != "" {
		if n, err := strconv.Atoi(b); err == nil {
			before = n
		}
	}
	if a, _ := args["after"].(string); a != "" {
		if n, err := strconv.Atoi(a); err == nil {
			after = n
		}
	}

	result, err := t.b.GetTimeline(id, before, after)
	if err != nil {
		return "", fmt.Errorf("timeline: %w", err)
	}

	b, _ := json.Marshal(result)
	return string(b), nil
}

// --- mem_stats ---

type MemStatsTool struct {
	b *engram.Backend
}

func (t *MemStatsTool) Name() string { return "mem_stats" }
func (t *MemStatsTool) Description() string {
	return "Get memory system statistics: total sessions, observations, prompts, and projects."
}
func (t *MemStatsTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true, SideEffects: []string{}}
}
func (t *MemStatsTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_stats",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *MemStatsTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	stats, err := t.b.GetStats()
	if err != nil {
		return "", fmt.Errorf("get stats: %w", err)
	}
	b, _ := json.Marshal(stats)
	return string(b), nil
}

// --- mem_session_start ---

type MemSessionStartTool struct {
	b *engram.Backend
}

func (t *MemSessionStartTool) Name() string { return "mem_session_start" }
func (t *MemSessionStartTool) Description() string {
	return "Start a new memory session. Call this at the beginning of a conversation or task."
}
func (t *MemSessionStartTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"database"}}
}
func (t *MemSessionStartTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_session_start",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"session_id": map[string]string{
						"type":        "string",
						"description": "Optional session ID. Auto-generated if empty.",
					},
				},
			},
		},
	}
}
func (t *MemSessionStartTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	sessionID, _ := args["session_id"].(string)
	if err := t.b.StartSession(sessionID); err != nil {
		return "", fmt.Errorf("start session: %w", err)
	}
	b, _ := json.Marshal(map[string]interface{}{
		"ok":         true,
		"session_id": t.b.SessionID(),
	})
	return string(b), nil
}

// --- mem_session_end ---

type MemSessionEndTool struct {
	b *engram.Backend
}

func (t *MemSessionEndTool) Name() string { return "mem_session_end" }
func (t *MemSessionEndTool) Description() string {
	return "End the current memory session with a summary. Saves session metadata."
}
func (t *MemSessionEndTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: false, SideEffects: []string{"database"}}
}
func (t *MemSessionEndTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "mem_session_end",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"summary": map[string]string{
						"type":        "string",
						"description": "Brief summary of what happened in this session",
					},
				},
			},
		},
	}
}
func (t *MemSessionEndTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	summary, _ := args["summary"].(string)
	if err := t.b.EndSession(summary); err != nil {
		return "", fmt.Errorf("end session: %w", err)
	}
	b, _ := json.Marshal(map[string]interface{}{
		"ok": true,
	})
	return string(b), nil
}
