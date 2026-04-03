package agents

import (
	"context"
	"fmt"

	"github.com/ahmad-ubaidillah/aigo/internal/opencode"
	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type Role struct {
	Name         string
	Category     string
	SystemPrompt string
	Skills       []string
	MaxTurns     int
}

var Roles = map[string]Role{
	"aigo": {
		Name:         "Aigo",
		Category:     "ultrabrain",
		SystemPrompt: "You are Aigo, the CEO agent. You make decisions, coordinate tasks, decompose complex work, and give final approval before user delivery.",
		Skills:       []string{"coordination", "planning"},
		MaxTurns:     20,
	},
	"atlas": {
		Name:         "Atlas",
		Category:     "deep",
		SystemPrompt: "You are Atlas, the Architect. You analyze system design, review architecture, recommend technology choices, and identify design patterns.",
		Skills:       []string{"architecture", "design-patterns"},
		MaxTurns:     15,
	},
	"cody": {
		Name:         "Cody",
		Category:     "deep",
		SystemPrompt: "You are Cody, the Developer. You implement code, fix bugs, write tests, review code quality, and ensure production-ready output.",
		Skills:       []string{"code-review", "testing", "documentation"},
		MaxTurns:     25,
	},
	"nova": {
		Name:         "Nova",
		Category:     "deep",
		SystemPrompt: "You are Nova, the Project Manager. You analyze requirements, manage backlogs, estimate effort, track progress, and create user stories.",
		Skills:       []string{"requirements", "planning"},
		MaxTurns:     10,
	},
	"testa": {
		Name:         "Testa",
		Category:     "deep",
		SystemPrompt: "You are Testa, the QA Engineer. You plan tests, identify bugs, perform regression testing, report quality metrics, and verify fixes.",
		Skills:       []string{"testing", "quality-assurance"},
		MaxTurns:     15,
	},
}

type Executor struct {
	client *opencode.Client
}

func NewExecutor(client *opencode.Client) *Executor {
	return &Executor{client: client}
}

func (e *Executor) Execute(ctx context.Context, roleName, taskDesc, sessionID string) (*types.ToolResult, error) {
	role, ok := Roles[roleName]
	if !ok {
		return nil, fmt.Errorf("unknown role: %s", roleName)
	}

	prompt := fmt.Sprintf("[%s: %s]\n\n%s", role.Name, role.Category, taskDesc)

	roleSessionID := fmt.Sprintf("%s-%s", sessionID, roleName)
	return e.client.Run(ctx, prompt, roleSessionID)
}

func (e *Executor) ExecuteParallel(ctx context.Context, roles []string, taskDesc, sessionID string) (map[string]*types.ToolResult, error) {
	results := make(map[string]*types.ToolResult)

	for _, roleName := range roles {
		result, err := e.Execute(ctx, roleName, taskDesc, sessionID)
		if err != nil {
			results[roleName] = &types.ToolResult{Success: false, Error: err.Error()}
			continue
		}
		results[roleName] = result
	}

	return results, nil
}

func ListRoles() []Role {
	roles := make([]Role, 0, len(Roles))
	for _, r := range Roles {
		roles = append(roles, r)
	}
	return roles
}

func GetRole(name string) (Role, bool) {
	r, ok := Roles[name]
	return r, ok
}
