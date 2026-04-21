package plan

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hermes-v2/aigo/internal/tools"
)

type PlanTool struct {
	mgr *Manager
}

func NewPlanTool(basePath string) (*PlanTool, error) {
	mgr, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &PlanTool{mgr: mgr}, nil
}

func (t *PlanTool) Name() string { return "plan_create" }
func (t *PlanTool) Description() string {
	return "Create a new plan with tasks"
}

func (t *PlanTool) Annotations() tools.Annotations {
	return tools.Annotations{}
}

func (t *PlanTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "plan_create",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]interface{}{"type": "string", "description": "Plan name"},
					"description": map[string]interface{}{"type": "string", "description": "Plan description"},
				},
				"required": []string{"name"},
			},
		},
	}
}

func (t *PlanTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	desc, _ := args["description"].(string)
	if name == "" {
		return "", fmt.Errorf("name is required")
	}
	plan, err := t.mgr.Create(name, desc)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Created plan: %s", plan.ID), nil
}

type PlanAddTaskTool struct {
	mgr *Manager
}

func NewPlanAddTaskTool(basePath string) (*PlanAddTaskTool, error) {
	mgr, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &PlanAddTaskTool{mgr: mgr}, nil
}

func (t *PlanAddTaskTool) Name() string { return "plan_add_task" }
func (t *PlanAddTaskTool) Description() string {
	return "Add a task to a plan"
}

func (t *PlanAddTaskTool) Annotations() tools.Annotations {
	return tools.Annotations{}
}

func (t *PlanAddTaskTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "plan_add_task",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"plan_id":     map[string]interface{}{"type": "string", "description": "Plan ID"},
					"title":      map[string]interface{}{"type": "string", "description": "Task title"},
					"description": map[string]interface{}{"type": "string", "description": "Task description"},
				},
				"required": []string{"plan_id", "title"},
			},
		},
	}
}

func (t *PlanAddTaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	planID, _ := args["plan_id"].(string)
	title, _ := args["title"].(string)
	desc, _ := args["description"].(string)
	if planID == "" || title == "" {
		return "", fmt.Errorf("plan_id and title are required")
	}
	task, err := t.mgr.AddTask(planID, title, desc)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Added task: %s", task.ID), nil
}

type PlanListTool struct {
	mgr *Manager
}

func NewPlanListTool(basePath string) (*PlanListTool, error) {
	mgr, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &PlanListTool{mgr: mgr}, nil
}

func (t *PlanListTool) Name() string { return "plan_list" }
func (t *PlanListTool) Description() string {
	return "List all plans"
}

func (t *PlanListTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *PlanListTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "plan_list",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

func (t *PlanListTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	plans := t.mgr.List()
	if len(plans) == 0 {
		return "No plans", nil
	}
	var result string
	for _, p := range plans {
		result += fmt.Sprintf("%s | %s | %s\n", p.ID, p.Name, p.Status)
	}
	return result, nil
}

type PlanShowTool struct {
	mgr *Manager
}

func NewPlanShowTool(basePath string) (*PlanShowTool, error) {
	mgr, err := New(basePath)
	if err != nil {
		return nil, err
	}
	return &PlanShowTool{mgr: mgr}, nil
}

func (t *PlanShowTool) Name() string { return "plan_show" }
func (t *PlanShowTool) Description() string {
	return "Show plan details and tasks"
}

func (t *PlanShowTool) Annotations() tools.Annotations {
	return tools.Annotations{ReadOnly: true}
}

func (t *PlanShowTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "function",
		Function: tools.ToolFunctionSchema{
			Name:        "plan_show",
			Description: t.Description(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"plan_id": map[string]interface{}{"type": "string", "description": "Plan ID"},
				},
				"required": []string{"plan_id"},
			},
		},
	}
}

func (t *PlanShowTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	planID, _ := args["plan_id"].(string)
	if planID == "" {
		return "", fmt.Errorf("plan_id is required")
	}
	plan := t.mgr.Get(planID)
	if plan == nil {
		return "", fmt.Errorf("plan not found")
	}
	data, _ := json.MarshalIndent(plan, "", "  ")
	return string(data), nil
}

func RegisterPlanTools(reg *tools.Registry, basePath string) error {
	createTool, err := NewPlanTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(createTool)

	addTaskTool, err := NewPlanAddTaskTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(addTaskTool)

	listTool, err := NewPlanListTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(listTool)

	showTool, err := NewPlanShowTool(basePath)
	if err != nil {
		return err
	}
	reg.Register(showTool)

	return nil
}