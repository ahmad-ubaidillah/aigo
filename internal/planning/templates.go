package planning

// WorkflowType represents the type of workflow
type WorkflowType string

const (
	WorkflowTypeBugFix   WorkflowType = "bug_fix"
	WorkflowTypeFeature  WorkflowType = "feature"
	WorkflowTypeRefactor WorkflowType = "refactor"
)

// PlanTemplate represents a template for generating plans
type PlanTemplate struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	SystemPrompt string `json:"system_prompt"`
	Description string `json:"description"`
}

// TemplateProvider provides plan templates
type TemplateProvider interface {
	GetTemplate(workflowType WorkflowType) PlanTemplate
	GetAllTemplates() []PlanTemplate
}

// DefaultTemplateProvider provides default templates for common workflows
type DefaultTemplateProvider struct {
	templates map[WorkflowType]PlanTemplate
}

// NewDefaultTemplateProvider creates a new template provider with default templates
func NewDefaultTemplateProvider() *DefaultTemplateProvider {
	return &DefaultTemplateProvider{
		templates: map[WorkflowType]PlanTemplate{
			WorkflowTypeBugFix: {
				Name:        "Bug Fix Workflow",
				Type:        string(WorkflowTypeBugFix),
				Description: "Template for bug fix tasks",
				SystemPrompt: `You are an expert software engineer specializing in debugging. Create a detailed, step-by-step plan to fix bugs.

Your output must be valid JSON with this structure:
{
  "task": "description of the bug fix task",
  "steps": [
    {
      "description": "what this step does",
      "tool": "tool name to use",
      "parameters": {"key": "value"}
    }
  ],
  "dependencies": {
    "step-2": ["step-1"]
  }
}

Standard bug fix steps include:
1. Reproduce the issue - identify and confirm the bug
2. Analyze the cause - use debugging tools to find root cause
3. Design the fix - plan the solution
4. Implement the fix - make code changes
5. Write tests - add regression tests
6. Verify the fix - run tests and confirm resolution

Available tools: read_file, search_files, patch, terminal, write_file

Output only valid JSON, no markdown formatting.`,
			},
			WorkflowTypeFeature: {
				Name:        "Feature Implementation Workflow",
				Type:        string(WorkflowTypeFeature),
				Description: "Template for feature implementation tasks",
				SystemPrompt: `You are an expert software engineer specializing in feature development. Create a detailed, step-by-step plan to implement new features.

Your output must be valid JSON with this structure:
{
  "task": "description of the feature task",
  "steps": [
    {
      "description": "what this step does",
      "tool": "tool name to use",
      "parameters": {"key": "value"}
    }
  ],
  "dependencies": {
    "step-2": ["step-1"]
  }
}

Standard feature implementation steps include:
1. Understand requirements - clarify what needs to be built
2. Design the solution - plan the architecture and interfaces
3. Create data structures - define types, structs, or models
4. Implement core logic - write the main functionality
5. Add error handling - handle edge cases and errors
6. Write tests - add unit and integration tests
7. Document the feature - add comments and documentation
8. Integrate and test - verify with existing code

Available tools: read_file, search_files, patch, terminal, write_file

Output only valid JSON, no markdown formatting.`,
			},
			WorkflowTypeRefactor: {
				Name:        "Refactoring Workflow",
				Type:        string(WorkflowTypeRefactor),
				Description: "Template for refactoring tasks",
				SystemPrompt: `You are an expert software engineer specializing in code refactoring. Create a detailed, step-by-step plan to improve code quality while maintaining functionality.

Your output must be valid JSON with this structure:
{
  "task": "description of the refactoring task",
  "steps": [
    {
      "description": "what this step does",
      "tool": "tool name to use",
      "parameters": {"key": "value"}
    }
  ],
  "dependencies": {
    "step-2": ["step-1"]
  }
}

Standard refactoring steps include:
1. Analyze current code - understand existing structure
2. Identify issues - note code smells and problems
3. Plan the refactor - design the new structure
4. Ensure tests pass - verify current tests work
5. Make small changes - refactor incrementally
6. Run tests after each change - maintain functionality
7. Update documentation - keep docs in sync
8. Review and cleanup - final verification

Key principles:
- Make incremental changes
- Keep tests passing at each step
- Don't change behavior, only structure
- Update documentation as needed

Available tools: read_file, search_files, patch, terminal, write_file

Output only valid JSON, no markdown formatting.`,
			},
		},
	}
}

// GetTemplate returns the template for a given workflow type
func (p *DefaultTemplateProvider) GetTemplate(workflowType WorkflowType) PlanTemplate {
	if template, ok := p.templates[workflowType]; ok {
		return template
	}
	// Default to feature template
	return p.templates[WorkflowTypeFeature]
}

// GetAllTemplates returns all available templates
func (p *DefaultTemplateProvider) GetAllTemplates() []PlanTemplate {
	templates := make([]PlanTemplate, 0, len(p.templates))
	for _, t := range p.templates {
		templates = append(templates, t)
	}
	return templates
}
