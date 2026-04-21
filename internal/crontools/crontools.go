// Package crontools implements cron scheduler tools for Aigo.
package crontools

import (
	"context"
	"fmt"
	"strings"

	"github.com/hermes-v2/aigo/internal/cron"
	"github.com/hermes-v2/aigo/internal/tools"
)

func RegisterCronTools(reg *tools.Registry, scheduler *cron.Scheduler) {
	reg.Register(&CronAddTool{scheduler: scheduler})
	reg.Register(&CronListTool{scheduler: scheduler})
	reg.Register(&CronRemoveTool{scheduler: scheduler})
}

// --- cron_add ---
type CronAddTool struct{ scheduler *cron.Scheduler }

func (t *CronAddTool) Name() string        { return "cron_add" }
func (t *CronAddTool) Description() string { return "Schedule a recurring or one-shot task." }
func (t *CronAddTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: false}
}
func (t *CronAddTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "cron_add",
			Description: "Schedule a recurring or one-shot task. The task prompt will be executed by the agent at the scheduled time.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":     map[string]interface{}{"type": "string", "description": "Job name"},
					"schedule": map[string]interface{}{"type": "string", "description": "'@every 1h', '@daily', '@weekly', '@hourly', or ISO timestamp"},
					"prompt":   map[string]interface{}{"type": "string", "description": "What to do when job fires"},
				},
				"required": []string{"name", "schedule", "prompt"},
			},
		},
	}
}

func (t *CronAddTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	schedule, _ := args["schedule"].(string)
	prompt, _ := args["prompt"].(string)
	if name == "" || schedule == "" || prompt == "" {
		return "", fmt.Errorf("name, schedule, prompt required")
	}
	t.scheduler.Add(cron.Job{Name: name, Schedule: schedule, Prompt: prompt})
	return fmt.Sprintf("Scheduled: '%s' (%s)", name, schedule), nil
}

// --- cron_list ---
type CronListTool struct{ scheduler *cron.Scheduler }

func (t *CronListTool) Name() string        { return "cron_list" }
func (t *CronListTool) Description() string { return "List all scheduled jobs." }
func (t *CronListTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: false, ReadOnly: true}
}
func (t *CronListTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:            "cron_list",
			Description:     "List all scheduled jobs with status and next run time.",
			Parameters:      map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
	}
}

func (t *CronListTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	jobs := t.scheduler.List()
	if len(jobs) == 0 {
		return "No scheduled jobs.", nil
	}
	var parts []string
	for _, j := range jobs {
		next := "unknown"
		if !j.NextRun.IsZero() {
			next = j.NextRun.Format("Jan 2, 15:04")
		}
		parts = append(parts, fmt.Sprintf("- %s [%s] next: %s | runs: %d", j.Name, j.Schedule, next, j.RunCount))
	}
	return fmt.Sprintf("Jobs (%d):\n%s", len(jobs), strings.Join(parts, "\n")), nil
}

// --- cron_remove ---
type CronRemoveTool struct{ scheduler *cron.Scheduler }

func (t *CronRemoveTool) Name() string        { return "cron_remove" }
func (t *CronRemoveTool) Description() string { return "Remove a scheduled job." }
func (t *CronRemoveTool) Annotations() tools.Annotations {
	return tools.Annotations{Destructive: true, ReadOnly: false}
}
func (t *CronRemoveTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "cron_remove",
			Description: "Remove a scheduled job by name or ID.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{"type": "string", "description": "Job name to remove"},
				},
				"required": []string{"name"},
			},
		},
	}
}

func (t *CronRemoveTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	if name == "" {
		return "", fmt.Errorf("name required")
	}
	for _, j := range t.scheduler.List() {
		if j.ID == name || j.Name == name {
			t.scheduler.Remove(j.ID)
			return fmt.Sprintf("Removed: %s", j.Name), nil
		}
	}
	return fmt.Sprintf("Not found: %s", name), nil
}
