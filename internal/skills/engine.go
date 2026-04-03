package skills

import (
	"context"
	"fmt"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type SkillLoader interface {
	LoadSkill(id string) (*types.Skill, error)
	ListSkills() ([]types.Skill, error)
	SaveSkill(skill *types.Skill) error
	DeleteSkill(id string) error
}

type SkillExecutor interface {
	Execute(ctx context.Context, skill *types.Skill, input string) (*types.SkillResult, error)
}

type Engine struct {
	loader    SkillLoader
	executors map[string]SkillExecutor
}

func NewEngine(loader SkillLoader) *Engine {
	return &Engine{
		loader:    loader,
		executors: make(map[string]SkillExecutor),
	}
}

func (e *Engine) RegisterExecutor(category string, exec SkillExecutor) {
	e.executors[category] = exec
}

func (e *Engine) Execute(ctx context.Context, skillID, input string) (*types.SkillResult, error) {
	skill, err := e.loader.LoadSkill(skillID)
	if err != nil {
		return nil, fmt.Errorf("load skill %s: %w", skillID, err)
	}

	if !skill.Enabled {
		return &types.SkillResult{
			Error: fmt.Sprintf("skill %s is disabled", skillID),
		}, nil
	}

	exec, ok := e.executors[skill.Category]
	if !ok {
		return &types.SkillResult{
			Error: fmt.Sprintf("no executor for category %s", skill.Category),
		}, nil
	}

	return exec.Execute(ctx, skill, input)
}

func (e *Engine) ListSkills() ([]types.Skill, error) {
	return e.loader.ListSkills()
}

func (e *Engine) CreateSkill(name, description, code, category string) (*types.Skill, error) {
	skill := &types.Skill{
		ID:          generateID(name),
		Name:        name,
		Description: description,
		Code:        code,
		Category:    category,
		Version:     1,
		Enabled:     true,
	}

	if err := e.loader.SaveSkill(skill); err != nil {
		return nil, fmt.Errorf("save skill: %w", err)
	}

	return skill, nil
}

func generateID(name string) string {
	return fmt.Sprintf("skill_%s", name)
}
