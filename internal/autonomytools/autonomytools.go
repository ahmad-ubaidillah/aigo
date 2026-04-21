// Package autonomytools implements tool functions for the autonomous agent and diary.
package autonomytools

import (
	"context"
	"fmt"
	"time"

	"github.com/hermes-v2/aigo/internal/diary"
	"github.com/hermes-v2/aigo/internal/tools"
)

// RegisterAutonomyTools registers diary and autonomy tools in the registry.
func RegisterAutonomyTools(reg *tools.Registry, d *diary.Diary, runningCheck func() bool) {
	reg.Register(&DiaryWriteTool{diary: d})
	reg.Register(&DiaryReadTool{diary: d})
	reg.Register(&AutonomyStatusTool{runningCheck: runningCheck})
}

// --- diary_write ---
type DiaryWriteTool struct {
	diary *diary.Diary
}

func (t *DiaryWriteTool) Name() string       { return "diary_write" }
func (t *DiaryWriteTool) Description() string { return "Write a diary entry manually." }
func (t *DiaryWriteTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false, SideEffects: []string{"filesystem"}}
}
func (t *DiaryWriteTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "diary_write",
			Description: "Write a manual diary entry with title, content, and optional mood.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title":   map[string]string{"type": "string", "description": "Title of the diary entry"},
					"content": map[string]string{"type": "string", "description": "Content of the diary entry"},
					"mood":    map[string]string{"type": "string", "description": "Mood tag (happy, thoughtful, frustrated, excited, calm, curious)"},
				},
				"required": []string{"title", "content"},
			},
		},
	}
}
func (t *DiaryWriteTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	mood, _ := args["mood"].(string)

	if title == "" || content == "" {
		return "", fmt.Errorf("title and content are required")
	}

	entry := diary.Entry{
		Title:   title,
		Content: content,
		Mood:    mood,
		Tags:    []string{"manual"},
	}
	if err := t.diary.Write(entry); err != nil {
		return "", fmt.Errorf("write diary: %w", err)
	}
	return fmt.Sprintf("📝 Diary entry written: %s", title), nil
}

// --- diary_read ---
type DiaryReadTool struct {
	diary *diary.Diary
}

func (t *DiaryReadTool) Name() string       { return "diary_read" }
func (t *DiaryReadTool) Description() string { return "Read recent diary entries." }
func (t *DiaryReadTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *DiaryReadTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "diary_read",
			Description: "Read recent diary entries from the last N days.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"days": map[string]interface{}{
						"type":        "integer",
						"description": "Number of recent days to read (default 3)",
					},
					"date": map[string]string{
						"type":        "string",
						"description": "Specific date to read (YYYY-MM-DD format). If set, ignores days parameter.",
					},
				},
			},
		},
	}
}
func (t *DiaryReadTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	if date, ok := args["date"].(string); ok && date != "" {
		content, err := t.diary.Read(date)
		if err != nil {
			return "", fmt.Errorf("read diary: %w", err)
		}
		return content, nil
	}

	days := 3
	if d, ok := args["days"].(float64); ok && d > 0 {
		days = int(d)
	}

	content, err := t.diary.Recent(days)
	if err != nil {
		return "", fmt.Errorf("read diary: %w", err)
	}
	if content == "" {
		return "No diary entries found.", nil
	}
	return content, nil
}

// --- autonomy_status ---
type AutonomyStatusTool struct {
	runningCheck func() bool
}

func (t *AutonomyStatusTool) Name() string       { return "autonomy_status" }
func (t *AutonomyStatusTool) Description() string { return "Check if the autonomous agent is running." }
func (t *AutonomyStatusTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *AutonomyStatusTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "autonomy_status",
			Description: "Check if the autonomous agent is currently running and active.",
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}
func (t *AutonomyStatusTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	running := t.runningCheck()
	status := "stopped"
	if running {
		status = "running"
	}
	return fmt.Sprintf(`{"status":"%s","checked":"%s"}`, status, time.Now().Format(time.RFC3339)), nil
}
