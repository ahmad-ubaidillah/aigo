package skills

import (
	"fmt"
	"strings"
	"time"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
)

type SkillStore interface {
	ListSkills(category string) ([]types.Skill, error)
	GetSkill(name string) (*types.Skill, error)
	AddSkill(name, description, command, category, tags string) (*types.Skill, error)
	UpdateSkill(name, description, command, category, tags string, enabled bool) error
	DeleteSkill(name string) error
}

type Registry struct {
	skills    map[string]*types.Skill
	executor  *Executor
	workspace string
	db        SkillStore
}

func NewRegistry() *Registry {
	return &Registry{
		skills:   make(map[string]*types.Skill),
		executor: NewExecutor("", 60*time.Second),
	}
}

func NewRegistryWithStore(store SkillStore) *Registry {
	r := &Registry{
		skills:   make(map[string]*types.Skill),
		executor: NewExecutor("", 60*time.Second),
		db:       store,
	}
	r.loadFromDB()
	return r
}

func NewRegistryWithWorkspace(workspace string) *Registry {
	return &Registry{
		skills:    make(map[string]*types.Skill),
		executor:  NewExecutor(workspace, 60*time.Second),
		workspace: workspace,
	}
}

func (r *Registry) SetStore(store SkillStore) {
	r.db = store
	r.loadFromDB()
}

func (r *Registry) loadFromDB() {
	if r.db == nil {
		return
	}
	skills, err := r.db.ListSkills("")
	if err != nil {
		return
	}
	for i := range skills {
		s := &skills[i]
		r.skills[s.Name] = s
	}
}

func (r *Registry) SetWorkspace(workspace string) {
	r.workspace = workspace
	r.executor = NewExecutor(workspace, 60*time.Second)
}

func (r *Registry) LoadSkill(id string) (*types.Skill, error) {
	skill, ok := r.skills[id]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", id)
	}
	return skill, nil
}

func (r *Registry) List(category string) ([]types.Skill, error) {
	result := make([]types.Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		if category == "" || skill.Category == category {
			result = append(result, *skill)
		}
	}
	return result, nil
}

func (r *Registry) Register(name, description, command, category, tags string) error {
	if name == "" {
		return fmt.Errorf("skill name is required")
	}
	if command == "" {
		return fmt.Errorf("skill command is required")
	}

	now := time.Now()
	skill := &types.Skill{
		ID:          name,
		Name:        name,
		Description: description,
		Code:        command,
		Category:    category,
		Tags:        tags,
		Enabled:     true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.skills[name] = skill
	return nil
}

func (r *Registry) Execute(name, args string) (*types.SkillResult, error) {
	skill, ok := r.skills[name]
	if !ok {
		return nil, fmt.Errorf("skill not found: %s", name)
	}

	if !skill.Enabled {
		return nil, fmt.Errorf("skill disabled: %s", name)
	}

	if r.executor != nil {
		result, err := r.executor.Execute(name, args)
		if err != nil {
			return &types.SkillResult{
				Output:   result.Output,
				Metadata: result.Metadata,
				Error:    err.Error(),
			}, nil
		}
		return &types.SkillResult{
			Output:   result.Output,
			Metadata: result.Metadata,
		}, nil
	}

	return &types.SkillResult{
		Output:   fmt.Sprintf("Executed: %s %s", skill.Code, args),
		Metadata: map[string]string{"command": skill.Code, "args": args},
	}, nil
}

func (r *Registry) ExecuteRaw(command string) (*types.SkillResult, error) {
	if r.executor != nil {
		result, err := r.executor.Execute("raw", command)
		if err != nil {
			return &types.SkillResult{
				Output: result.Output,
				Error:  err.Error(),
			}, nil
		}
		return &types.SkillResult{
			Output:   result.Output,
			Metadata: result.Metadata,
		}, nil
	}
	return nil, fmt.Errorf("executor not initialized")
}

func (r *Registry) Search(query string) ([]types.Skill, error) {
	query = strings.ToLower(query)
	var result []types.Skill
	for _, skill := range r.skills {
		if strings.Contains(strings.ToLower(skill.Name), query) ||
			strings.Contains(strings.ToLower(skill.Description), query) ||
			strings.Contains(strings.ToLower(skill.Tags), query) {
			result = append(result, *skill)
		}
	}
	return result, nil
}

func (r *Registry) GetByCategory(category string) []types.Skill {
	var result []types.Skill
	for _, skill := range r.skills {
		if skill.Category == category && skill.Enabled {
			result = append(result, *skill)
		}
	}
	return result
}

func (r *Registry) SaveSkill(skill *types.Skill) error {
	if skill == nil {
		return fmt.Errorf("skill is nil")
	}
	r.skills[skill.ID] = skill
	return nil
}

func (r *Registry) DeleteSkill(id string) error {
	if _, ok := r.skills[id]; !ok {
		return fmt.Errorf("skill not found: %s", id)
	}
	delete(r.skills, id)
	return nil
}

func (r *Registry) AddBuiltIn(skills []types.Skill) {
	for i := range skills {
		skill := &skills[i]
		r.skills[skill.ID] = skill
	}
}
